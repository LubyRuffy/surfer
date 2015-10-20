package example

import (
	"github.com/henrylee2cn/surfer"
	"net/http"
)

var SurfDownloader = surfer.New()

func SurfDownload(req *Request) (*http.Response, error) {
	return SurfDownloader.Download(req)
}
