package surfer

import (
	"net/http"
	"net/url"
	"time"
)

type Request interface {
	GetUrl() string
	// GET POST POST-M HEAD
	GetMethod() string
	GetReferer() string
	// POST values
	GetPostData() url.Values
	// http header
	GetHeader() http.Header
	// http cookies, when Cookies equal nil, the UserAgent auto changes
	GetCookies() []*http.Cookie
	// enable http cookies
	GetEnableCookie() bool
	// timeout of dial
	GetDialTimeout() time.Duration
	// deadline of connect
	GetDeadline() time.Duration
	// the max times of download
	GetTryTimes() int
	// the pause time of retry
	GetRetryPause() time.Duration
	// the download ProxyHost
	GetProxy() string
	// max redirect times
	GetRedirectTimes() int
}
