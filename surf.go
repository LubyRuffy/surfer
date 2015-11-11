package surfer

import (
	"bytes"
	"crypto/tls"
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

// Default is the default Download implementation.
type Surf struct {
	// userAgent is the User-Agent header value sent with requests.
	userAgents map[string][]string

	// cookies stores cookies for every site visited by the browser.
	cookieJar http.CookieJar
}

func New() Surfer {
	surf := &Surf{
		userAgents: agent.UserAgents,
		cookieJar:  jar.NewCookiesMemory(),
	}
	l := len(surf.userAgents["common"])
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := r.Intn(l)
	surf.userAgents["common"][0], surf.userAgents["common"][idx] = surf.userAgents["common"][idx], surf.userAgents["common"][0]

	return surf
}

func (self *Surf) Download(req Request) (resp *http.Response, err error) {
	var param = new(Param)

	param.url, err = util.UrlEncode(req.GetUrl())
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

	param.header = req.GetHeader()
	param.enableCookie = req.GetEnableCookie()

	param.dialTimeout = req.GetDialTimeout()
	if param.dialTimeout < 0 {
		param.dialTimeout = 0
	}

	param.connTimeout = req.GetConnTimeout()
	param.tryTimes = req.GetTryTimes()
	param.retryPause = req.GetRetryPause()
	param.redirectTimes = req.GetRedirectTimes()
	param.proxy = req.GetProxy()

	param.client = self.buildClient(param)

	return self.httpRequest(param)
}

// buildClient creates, configures, and returns a *http.Client type.
func (self *Surf) buildClient(param *Param) *http.Client {
	client := &http.Client{
		CheckRedirect: param.checkRedirect,
	}

	if param.enableCookie {
		client.Jar = self.cookieJar
	}

	transport := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			c, err := net.DialTimeout(network, addr, param.dialTimeout)
			if err != nil {
				return nil, err
			}
			if param.connTimeout > 0 {
				c.SetDeadline(time.Now().Add(param.connTimeout))
			}
			return c, nil
		},
	}

	if param.proxy != "" {
		if px, err := url.Parse(param.proxy); err == nil {
			transport.Proxy = http.ProxyURL(px)
		}
	}

	if strings.ToLower(param.url.Scheme) == "https" {
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

	if param.contentType != "" {
		req.Header.Set("Content-Type", param.contentType)
	}

	for k, v := range param.header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	if len(req.Header.Get("User-Agent")) == 0 {
		if param.enableCookie {
			req.Header.Set("User-Agent", self.userAgents["common"][0])
		} else {
			l := len(self.userAgents["common"])
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			req.Header.Set("User-Agent", self.userAgents["common"][r.Intn(l)])
		}
	}

	if param.tryTimes <= 0 {
		for {
			resp, err = param.client.Do(req)
			if err != nil {
				time.Sleep(param.retryPause)
				continue
			}
			break
		}
	} else {
		for i := 0; i < param.tryTimes; i++ {
			resp, err = param.client.Do(req)
			if err != nil {
				time.Sleep(param.retryPause)
				continue
			}
			break
		}
	}

	return resp, err
}
