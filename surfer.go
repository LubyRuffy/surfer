// Package surfer ensembles other packages into a usable browser.
package surfer

import (
	"github.com/henrylee2cn/surfer/agent"
	"github.com/henrylee2cn/surfer/jar"
	"time"
)

var (
	// DefaultUserAgent is the global user agent value.
	DefaultUserAgent = agent.Create()

	// DefaultSendReferer is the global value for the AttributeSendReferer attribute.
	DefaultSendReferer = true

	// DefaultFollowRedirects is the global value for the AttributeFollowRedirects attribute.
	DefaultFollowRedirects = true
)

func NewDownload(tryTimes int, paseTime time.Duration, proxy string) *Download {
	dl := &Download{}
	dl.SetTryTimes(tryTimes)
	dl.SetPaseTime(paseTime)
	dl.SetProxy(proxy)
	dl.SetUserAgent(DefaultUserAgent)
	dl.SetCookieJar(jar.NewMemoryCookies())
	dl.SetAttrs(AttrMap{
		SendReferer:     DefaultSendReferer,
		FollowRedirects: DefaultFollowRedirects,
	})
	return dl
}

func NewSurfer(tryTimes int, paseTime time.Duration, proxy string) Surfer {
	return NewDownload(tryTimes, paseTime, proxy)
}
