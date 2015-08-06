package surfer

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/henrylee2cn/surfer/errors"
	"github.com/henrylee2cn/surfer/util"
)

// Attr represents a Download capability.
type Attr int

// AttrMap represents a map of Attr values.
type AttrMap map[Attr]bool

const (
	// SendRefererAttr instructs a Download to send the Referer header.
	SendReferer Attr = iota

	// FollowRedirectsAttr instructs a Download to follow Location headers.
	FollowRedirects
)

// Default is the default Download implementation.
type Download struct {
	// userAgent is the User-Agent header value sent with requests.
	userAgent string

	// attrs is the set download attrs.
	attrs AttrMap

	// cookies stores cookies for every site visited by the browser.
	cookies http.CookieJar

	//the time of trying to download
	tryTimes int

	paseTime time.Duration

	proxy string
}

// SetTryTimes sets the tryTimes of download.
func (self *Download) SetTryTimes(tryTimes int) {
	self.tryTimes = tryTimes
}

// SetPaseTime sets the pase time of retry.
func (self *Download) SetPaseTime(paseTime time.Duration) {
	self.paseTime = paseTime
}

// SetCookieJar is used to set the cookie jar the download uses.
func (self *Download) SetCookieJar(cj http.CookieJar) {
	self.cookies = cj
}

// SetUserAgent sets the user agent.
func (self *Download) SetUserAgent(userAgent string) {
	self.userAgent = userAgent
}

// SetProxy sets a download ProxyHost.
func (self *Download) SetProxy(proxy string) {
	self.proxy = proxy
}

// SetAttr sets a download instruction attribute.
func (self *Download) SetAttr(a Attr, v bool) {
	self.attrs[a] = v
}

// SetAttrs is used to set all the download attrs.
func (self *Download) SetAttrs(a AttrMap) {
	self.attrs = a
}

func (self *Download) Download(method string, u string, ref string, data url.Values, header http.Header, cookies []*http.Cookie) (resp *http.Response, err error) {
	switch method {
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
func (self *Download) Get(u string, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	urlObj, err := util.UrlEncode(u)
	if err != nil {
		return nil, err
	}
	client := self.buildClient(urlObj.Scheme, self.proxy)
	return self.httpGET(urlObj, "", header, cookies, client)
}

// Open requests the given URL using the HEAD method.
func (self *Download) Head(u string, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	urlObj, err := util.UrlEncode(u)
	if err != nil {
		return nil, err
	}
	client := self.buildClient(urlObj.Scheme, self.proxy)
	return self.httpHEAD(urlObj, "", header, cookies, client)
}

// PostForm requests the given URL using the POST method with the given data.
func (self *Download) PostForm(u string, ref string, data url.Values, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	return self.Post(u, ref, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()), header, cookies)
}

// PostMultipart requests the given URL using the POST method with the given data using multipart/form-data format.
func (self *Download) PostMultipart(u string, ref string, data url.Values, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
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
func (self *Download) Post(u string, ref string, contentType string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	urlObj, err := util.UrlEncode(u)
	if err != nil {
		return nil, err
	}
	client := self.buildClient(urlObj.Scheme, self.proxy)
	return self.httpPOST(urlObj, ref, contentType, body, header, cookies, client)
}

// -- Unexported methods --

// httpGET makes an HTTP GET request for the given URL.
// When via is not nil, and AttributeSendReferer is true, the Referer header will
// be set to ref.
func (self *Download) httpGET(u *url.URL, ref string, header http.Header, cookies []*http.Cookie, client *http.Client) (*http.Response, error) {
	req, err := self.buildRequest("GET", u.String(), ref, nil, header, cookies)
	if err != nil {
		return nil, err
	}
	return self.httpRequest(req, client)
}

// httpHEAD makes an HTTP HEAD request for the given URL.
// When via is not nil, and AttributeSendReferer is true, the Referer header will
// be set to ref.
func (self *Download) httpHEAD(u *url.URL, ref string, header http.Header, cookies []*http.Cookie, client *http.Client) (*http.Response, error) {
	req, err := self.buildRequest("HEAD", u.String(), ref, nil, header, cookies)
	if err != nil {
		return nil, err
	}
	return self.httpRequest(req, client)
}

// httpPOST makes an HTTP POST request for the given URL.
// When via is not nil, and AttributeSendReferer is true, the Referer header will
// be set to ref.
func (self *Download) httpPOST(u *url.URL, ref string, contentType string, body io.Reader, header http.Header, cookies []*http.Cookie, client *http.Client) (*http.Response, error) {
	req, err := self.buildRequest("POST", u.String(), ref, body, header, cookies)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", contentType)

	return self.httpRequest(req, client)
}

// send uses the given *http.Request to make an HTTP request.
func (self *Download) httpRequest(req *http.Request, client *http.Client) (resp *http.Response, err error) {
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
func (self *Download) buildClient(scheme string, proxy string) *http.Client {
	client := &http.Client{}

	client.Jar = self.cookies

	client.CheckRedirect = self.shouldRedirect

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
func (self *Download) buildRequest(method, url string, ref string, body io.Reader, header http.Header, cookies []*http.Cookie) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for k, v := range header {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	req.Header.Set("User-Agent", self.userAgent)

	if self.attrs[SendReferer] {
		req.Header.Set("Referer", ref)
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	if os.Getenv("SURF_DEBUG_HEADERS") != "" {
		d, _ := httputil.DumpRequest(req, false)
		fmt.Fprintln(os.Stderr, "===== [DUMP] =====\n", string(d))
	}

	return req, nil
}

// shouldRedirect is used as the value to http.Client.CheckRedirect.
func (self *Download) shouldRedirect(req *http.Request, _ []*http.Request) error {
	if self.attrs[FollowRedirects] {
		return nil
	}
	return errors.NewLocation(
		"Redirects are disabled. Cannot follow '%s'.", req.URL.String())
}
