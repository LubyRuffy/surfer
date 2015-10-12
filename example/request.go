package example

import (
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultDialTimeout = 2 * time.Minute // 默认请求服务器超时
	DefaultConnTimeout = 2 * time.Minute // 默认下载超时
	DefaultTryTimes    = 3               // 默认最大下载次数
	DefaultRetryPause  = 2 * time.Second // 默认重新下载前停顿时长
)

// Request represents object waiting for being crawled.
type Request struct {
	Url     string
	Referer string
	// GET POST POST-M HEAD
	Method string
	// http header
	Header http.Header
	// enable http cookies
	EnableCookie bool
	// http cookies, when Cookies equal nil, the UserAgent auto changes
	Cookies []*http.Cookie
	// POST values
	PostData url.Values
	// dial tcp: i/o timeout
	DialTimeout time.Duration
	// WSARecv tcp: i/o timeout
	ConnTimeout time.Duration
	// the max times of download
	TryTimes int
	// how long pause when retry
	RetryPause time.Duration
	// max redirect times
	// when RedirectTimes equal 0, redirect times is ∞
	// when RedirectTimes less than 0, redirect times is 0
	RedirectTimes int
	// the download ProxyHost
	Proxy string
}

func (self *Request) GetUrl() string {
	return self.Url
}

func (self *Request) GetMethod() string {
	return strings.ToUpper(self.Method)
}

func (self *Request) GetReferer() string {
	return self.Referer
}

func (self *Request) GetPostData() url.Values {
	return self.PostData
}

func (self *Request) GetHeader() http.Header {
	return self.Header
}

func (self *Request) GetEnableCookie() bool {
	return self.EnableCookie
}

func (self *Request) GetCookies() []*http.Cookie {
	return self.Cookies
}

func (self *Request) GetDialTimeout() time.Duration {
	return self.DialTimeout
}

func (self *Request) GetConnTimeout() time.Duration {
	return self.ConnTimeout
}

func (self *Request) GetTryTimes() int {
	return self.TryTimes
}

func (self *Request) GetRetryPause() time.Duration {
	return self.RetryPause
}

func (self *Request) GetProxy() string {
	return self.Proxy
}

func (self *Request) GetRedirectTimes() int {
	return self.RedirectTimes
}
