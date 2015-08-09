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
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/henrylee2cn/surfer/agent"
	"github.com/henrylee2cn/surfer/jar"
	"github.com/henrylee2cn/surfer/util"
)

// Downloader represents an core of HTTP web browser for crawler.
type Surfer interface {
	// static UserAgent/can cookie or dynamic UserAgent/disable cookie
	SetUseCookie(use bool) Surfer

	// SetProxy sets a download ProxyHost.
	SetProxy(proxy string) Surfer

	// SetTryTimes sets the tryTimes of download.
	SetTryTimes(tryTimes int) Surfer

	// SetPaseTime sets the pase time of retry.
	SetPaseTime(paseTime time.Duration) Surfer

	// Get requests the given URL using the GET method.
	Get(u string, header http.Header, cookies []*http.Cookie) (*http.Response, error)

	// Open requests the given URL using the HEAD method.
	Head(u string, header http.Header, cookies []*http.Cookie) (*http.Response, error)

	// Post requests the given URL using the POST method.
	Post(u, ref string, contentType string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error)

	// PostForm requests the given URL using the POST method with the given data.
	PostForm(u, ref string, data url.Values, header http.Header, cookies []*http.Cookie) (*http.Response, error)

	// PostMultipart requests the given URL using the POST method with the given data using multipart/form-data format.
	PostMultipart(u, ref string, data url.Values, header http.Header, cookies []*http.Cookie) (*http.Response, error)

	Download(method, u, ref string, data url.Values, header http.Header, cookies []*http.Cookie) (resp *http.Response, err error)
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
	paseTime time.Duration

	// proxy host
	proxy string
}

func New() Surfer {
	return &Surf{
		userAgents:     agent.UserAgents,
		useCookie:      true,
		cookieJar:      jar.NewCookiesMemory(),
		sendReferer:    true,
		followRedirect: true,
		tryTimes:       3,
		paseTime:       0,
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

// SetPaseTime sets the pase time of retry.
func (self *Surf) SetPaseTime(paseTime time.Duration) Surfer {
	self.paseTime = paseTime
	return self
}

// SetProxy sets a download ProxyHost.
func (self *Surf) SetProxy(proxy string) Surfer {
	self.proxy = proxy
	return self
}

func (self *Surf) Download(method, u, ref string, data url.Values, header http.Header, cookies []*http.Cookie) (resp *http.Response, err error) {
	switch strings.ToUpper(method) {
	case "GET":
		resp, err = self.Get(u, header, cookies)
	case "HEAD":
		resp, err = self.Head(u, header, cookies)
	case "POST":
		resp, err = self.PostForm(u, ref, data, header, cookies)
	case "POST-M":
		resp, err = self.PostMultipart(u, ref, data, header, cookies)
	}

	return resp, err
}

// Get requests the given URL using the GET method.
func (self *Surf) Get(u string, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	urlObj, err := util.UrlEncode(u)
	if err != nil {
		return nil, err
	}
	client := self.buildClient(urlObj.Scheme, self.proxy)
	return self.httpGET(urlObj, "", header, cookies, client)
}

// Open requests the given URL using the HEAD method.
func (self *Surf) Head(u string, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	urlObj, err := util.UrlEncode(u)
	if err != nil {
		return nil, err
	}
	client := self.buildClient(urlObj.Scheme, self.proxy)
	return self.httpHEAD(urlObj, "", header, cookies, client)
}

// PostForm requests the given URL using the POST method with the given data.
func (self *Surf) PostForm(u, ref string, data url.Values, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	return self.Post(u, ref, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()), header, cookies)
}

// PostMultipart requests the given URL using the POST method with the given data using multipart/form-data format.
func (self *Surf) PostMultipart(u, ref string, data url.Values, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for k, vs := range data {
		for _, v := range vs {
			writer.WriteField(k, v)
		}
	}
	err := writer.Close()
	if err != nil {
		return nil, err
	}
	return self.Post(u, ref, writer.FormDataContentType(), body, header, cookies)
}

// Post requests the given URL using the POST method.
func (self *Surf) Post(u, ref, contentType string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	urlObj, err := util.UrlEncode(u)
	if err != nil {
		return nil, err
	}
	client := self.buildClient(urlObj.Scheme, self.proxy)
	return self.httpPOST(urlObj, ref, contentType, body, header, cookies, client)
}

// -- Unexported methods --

// httpGET makes an HTTP GET request for the given URL.
// When via is not nil, and sendReferer is true, the Referer header will
// be set to ref.
func (self *Surf) httpGET(u *url.URL, ref string, header http.Header, cookies []*http.Cookie, client *http.Client) (*http.Response, error) {
	req, err := self.buildRequest("GET", u.String(), ref, nil, header, cookies)
	if err != nil {
		return nil, err
	}
	return self.httpRequest(req, client)
}

// httpHEAD makes an HTTP HEAD request for the given URL.
// When via is not nil, and sendReferer is true, the Referer header will
// be set to ref.
func (self *Surf) httpHEAD(u *url.URL, ref string, header http.Header, cookies []*http.Cookie, client *http.Client) (*http.Response, error) {
	req, err := self.buildRequest("HEAD", u.String(), ref, nil, header, cookies)
	if err != nil {
		return nil, err
	}
	return self.httpRequest(req, client)
}

// httpPOST makes an HTTP POST request for the given URL.
// When via is not nil, and sendReferer is true, the Referer header will
// be set to ref.
func (self *Surf) httpPOST(u *url.URL, ref string, contentType string, body io.Reader, header http.Header, cookies []*http.Cookie, client *http.Client) (*http.Response, error) {
	req, err := self.buildRequest("POST", u.String(), ref, body, header, cookies)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)

	return self.httpRequest(req, client)
}

// send uses the given *http.Request to make an HTTP request.
func (self *Surf) httpRequest(req *http.Request, client *http.Client) (resp *http.Response, err error) {
	for i := 0; i < self.tryTimes; i++ {
		resp, err = client.Do(req)
		if err != nil {
			time.Sleep(self.paseTime)
			continue
		}
		break
	}
	return
}

// buildClient creates, configures, and returns a *http.Client type.
func (self *Surf) buildClient(scheme string, proxy string) *http.Client {
	client := &http.Client{}

	client.Jar = self.cookieJar

	client.CheckRedirect = self.checkRedirect

	transport := &http.Transport{}

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

// buildRequest creates and returns a *http.Request type.
// Sets any headers that need to be sent with the request.
func (self *Surf) buildRequest(method, url string, ref string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for k, v := range header {
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
		req.Header.Set("Referer", ref)
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	return req, nil
}

// checkRedirect is used as the value to http.Client.CheckRedirect.
func (self *Surf) checkRedirect(req *http.Request, _ []*http.Request) error {
	if self.followRedirect {
		return nil
	}
	return errors.New(fmt.Sprintf("Redirect are disabled. Cannot follow '%s'.", req.URL.String()))
}
