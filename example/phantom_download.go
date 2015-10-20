package example

import (
	"github.com/henrylee2cn/surfer"
	"net/http"
)

var PhantomDownloader = surfer.NewPhantom("../phantomjs/phantomjs", "./")

func PhantomDownload(req *Request) (*http.Response, error) {
	return PhantomDownloader.Download(req)
}
