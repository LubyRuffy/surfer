package surfer

import (
	"io"
	"net/http"
	"net/url"
	"time"
)

// Downloader represents an core of HTTP web browser for crawler.
type Surfer interface {
	// SetUserAgent sets the user agent.
	SetUserAgent(ua string)

	// SetAtt sets a browser instruction attribute.
	SetAttr(a Attr, v bool)

	// SetAttributes is used to set all the browser attributes.
	SetAttrs(a AttrMap)

	// SetCookieJar is used to set the cookie jar the browser uses.
	SetCookieJar(cj http.CookieJar)

	// SetProxy sets a download ProxyHost.
	SetProxy(proxy string)

	// SetTryTimes sets the tryTimes of download.
	SetTryTimes(tryTimes int)

	// SetPaseTime sets the pase time of retry.
	SetPaseTime(paseTime time.Duration)

	// Get requests the given URL using the GET method.
	Get(u string, header http.Header, cookies []*http.Cookie) (*http.Response, error)

	// Open requests the given URL using the HEAD method.
	Head(u string, header http.Header, cookies []*http.Cookie) (*http.Response, error)

	// Post requests the given URL using the POST method.
	Post(u string, ref string, contentType string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error)

	// PostForm requests the given URL using the POST method with the given data.
	PostForm(u string, ref string, data url.Values, header http.Header, cookies []*http.Cookie) (*http.Response, error)

	// PostMultipart requests the given URL using the POST method with the given data using multipart/form-data format.
	PostMultipart(u string, ref string, data url.Values, header http.Header, cookies []*http.Cookie) (*http.Response, error)

	Download(method string, u string, ref string, data url.Values, header http.Header, cookies []*http.Cookie) (resp *http.Response, err error)
}
