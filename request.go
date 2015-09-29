package surfer

import (
	"net/http"
	"net/url"
	"time"
)

type Request interface {
	GetUrl() string
	GetMethod() string
	GetReferer() string
	GetPostData() url.Values
	GetHeader() http.Header
	GetCookies() []*http.Cookie
	GetDeadline() time.Duration
	GetPauseTime() time.Duration
}
