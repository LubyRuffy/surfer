// surfer是一款Go语言编写的高并发爬虫下载器，支持 GET/POST/HEAD 方法及 http/https 协议，同时支持固定UserAgent自动保存cookie与随机大量UserAgent禁用cookie两种模式，高度模拟浏览器行为，可实现模拟登录等功能。
package surfer

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/henrylee2cn/surfer/agent"
	"github.com/henrylee2cn/surfer/jar"
	"github.com/henrylee2cn/surfer/util"
)

const (
	TryTimes    = 3
	Deadline    = 30 * time.Second
	DialTimeout = 10 * time.Second
	PauseTime   = 1 * time.Second
)

// Downloader represents an core of HTTP web browser for crawler.
type Surfer interface {
	// static UserAgent/can cookie or dynamic UserAgent/disable cookie
	SetUseCookie(use bool) Surfer

	// SetProxy sets a download ProxyHost.
	SetProxy(proxy string) Surfer

	// SetTryTimes sets the tryTimes of download.
	SetTryTimes(tryTimes int) Surfer

	// SetDeadline sets the default deadline of connect.
	SetDeadline(t time.Duration) Surfer

	// SetDialTimeout sets the default  timeout of dial.
	SetDialTimeout(t time.Duration) Surfer

	// SetPauseTime sets the pase time of retry.
	SetPauseTime(t time.Duration) Surfer

	// GET @param url string, header http.Header, cookies []*http.Cookie
	// HEAD @param url string, header http.Header, cookies []*http.Cookie
	// POST PostForm @param url, referer string, values url.Values, header http.Header, cookies []*http.Cookie
	// POST PostMultipart @param url, referer string, values url.Values, header http.Header, cookies []*http.Cookie
	Download(Request) (resp *http.Response, err error)
}

// Default is the default Download implementation.
type Surf struct {
	// userAgent is the User-Agent header value sent with requests.
	userAgents map[string][]string

	// "true": static UserAgent/can cookie or "false": dynamic UserAgent/disable cookie
	useCookie bool

	// cookies stores cookies for every site visited by the browser.
	cookieJar http.CookieJar

	// can sends referer
	sendReferer bool

	// can follows redirects
	followRedirect bool

	//the time of trying to download
	tryTimes int

	// how long pase when retry
	pauseTime time.Duration

	deadline    time.Duration
	dialTimeout time.Duration

	// proxy host
	proxy string
}

type Param struct {
	method      string
	url         *url.URL
	referer     string
	contentType string
	body        io.Reader
	header      http.Header
	cookies     []*http.Cookie
	client      *http.Client
	pauseTime   time.Duration
	deadline    time.Duration
}

func New() Surfer {
	return &Surf{
		userAgents:     agent.UserAgents,
		useCookie:      true,
		cookieJar:      jar.NewCookiesMemory(),
		sendReferer:    true,
		followRedirect: true,
		tryTimes:       TryTimes,
		deadline:       Deadline,
		dialTimeout:    DialTimeout,
		pauseTime:      PauseTime,
		proxy:          "",
	}
}

// "true": static UserAgent/can cookie or "false": dynamic UserAgent/disable cookie
func (self *Surf) SetUseCookie(use bool) Surfer {
	self.useCookie = use
	if use {
		self.cookieJar = jar.NewCookiesMemory()
		l := len(self.userAgents["common"])
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		idx := r.Intn(l)
		self.userAgents["common"][0], self.userAgents["common"][idx] = self.userAgents["common"][idx], self.userAgents["common"][0]
	} else {
		self.cookieJar = nil
	}
	return self
}

// SetTryTimes sets the tryTimes of download.
func (self *Surf) SetTryTimes(tryTimes int) Surfer {
	self.tryTimes = tryTimes
	return self
}

// SetDeadline sets the default deadline of connect.
func (self *Surf) SetDeadline(t time.Duration) Surfer {
	if t == 0 {
		return self
	}
	self.deadline = t
	return self
}

// SetDialTimeout sets the default  timeout of dial.
func (self *Surf) SetDialTimeout(t time.Duration) Surfer {
	if t == 0 {
		return self
	}
	self.dialTimeout = t
	return self
}

// SetPauseTime sets the pase time of retry.
func (self *Surf) SetPauseTime(t time.Duration) Surfer {
	if t == 0 {
		return self
	}
	self.pauseTime = t
	return self
}

// SetProxy sets a download ProxyHost.
func (self *Surf) SetProxy(proxy string) Surfer {
	self.proxy = proxy
	return self
}

func (self *Surf) Download(req Request) (resp *http.Response, err error) {
	param, err := self.param(req)
	if err != nil {
		return nil, err
	}

	switch method := strings.ToUpper(req.GetMethod()); method {
	case "GET", "HEAD":
		param.method = method

	case "POST":
		param.method = method
		param.contentType = "application/x-www-form-urlencoded"
		param.body = strings.NewReader(req.GetPostData().Encode())

	case "POST-M":
		param.method = "POST"
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		for k, vs := range req.GetPostData() {
			for _, v := range vs {
				writer.WriteField(k, v)
			}
		}
		err := writer.Close()
		if err != nil {
			return nil, err
		}
		param.contentType = writer.FormDataContentType()
		param.body = body

	default:
		param.method = "GET"
	}

	return self.httpRequest(param)
}

// -- Unexported methods --

func (self *Surf) param(req Request) (param *Param, err error) {
	param = new(Param)

	param.url, err = util.UrlEncode(req.GetUrl())
	if err != nil {
		return nil, err
	}

	param.deadline = req.GetDeadline()
	if param.deadline == 0 {
		param.deadline = self.deadline
	}
	param.client = self.buildClient(param.url.Scheme, self.proxy, param.deadline)

	param.referer = req.GetReferer()
	param.header = req.GetHeader()
	param.cookies = req.GetCookies()

	param.pauseTime = req.GetPauseTime()
	if param.pauseTime == 0 {
		param.pauseTime = self.pauseTime
	}

	return param, err
}

// buildClient creates, configures, and returns a *http.Client type.
func (self *Surf) buildClient(scheme string, proxy string, deadline time.Duration) *http.Client {
	client := &http.Client{
		Jar:           self.cookieJar,
		CheckRedirect: self.checkRedirect,
	}

	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			c, err := net.DialTimeout(network, addr, self.dialTimeout)
			if err != nil {
				return nil, err
			}
			c.SetDeadline(time.Now().Add(deadline))
			return c, nil
		},
	}

	if proxy != "" {
		if px, err := url.Parse(proxy); err == nil {
			transport.Proxy = http.ProxyURL(px)
		}
	}

	if strings.ToLower(scheme) == "https" {
		transport.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true}
		transport.DisableCompression = true
	}
	client.Transport = transport
	return client
}

// send uses the given *http.Request to make an HTTP request.
func (self *Surf) httpRequest(param *Param) (resp *http.Response, err error) {
	req, err := http.NewRequest(param.method, param.url.String(), param.body)
	if err != nil {
		return nil, err
	}

	for k, v := range param.header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	// if user can't sets User-Agent
	if req.UserAgent() == "" {
		if self.useCookie {
			req.Header.Set("User-Agent", self.userAgents["common"][0])
		} else {
			l := len(self.userAgents["common"])
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			req.Header.Set("User-Agent", self.userAgents["common"][r.Intn(l)])
		}
	}

	if self.sendReferer {
		req.Header.Set("Referer", param.referer)
	}

	for _, cookie := range param.cookies {
		req.AddCookie(cookie)
	}

	if param.contentType != "" {
		req.Header.Add("Content-Type", param.contentType)
	}

	for i := 0; i < self.tryTimes; i++ {
		resp, err = param.client.Do(req)
		if err != nil {
			time.Sleep(param.pauseTime)
			continue
		}
		break
	}

	return resp, err
}

// checkRedirect is used as the value to http.Client.CheckRedirect.
func (self *Surf) checkRedirect(req *http.Request, _ []*http.Request) error {
	if self.followRedirect {
		return nil
	}
	return errors.New(fmt.Sprintf("Redirect are disabled. Cannot follow '%s'.", req.URL.String()))
}
