package surfer

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Param struct {
	method        string
	url           *url.URL
	referer       string
	contentType   string
	body          io.Reader
	header        http.Header
	cookies       []*http.Cookie
	enableCookie  bool
	dialTimeout   time.Duration
	connTimeout   time.Duration
	tryTimes      int
	retryPause    time.Duration
	redirectTimes int
	proxy         string
	client        *http.Client
}

// checkRedirect is used as the value to http.Client.CheckRedirect
// when redirectTimes equal 0, redirect times is ∞
// when redirectTimes less than 0, not allow redirects
func (self *Param) checkRedirect(req *http.Request, via []*http.Request) error {
	if self.redirectTimes == 0 {
		return nil
	}
	if len(via) >= self.redirectTimes {
		if self.redirectTimes < 0 {
			return fmt.Errorf("not allow redirects.")
		}
		return fmt.Errorf("stopped after %v redirects.", self.redirectTimes)
	}
	return nil
}
