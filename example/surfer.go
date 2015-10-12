package example

import (
	"github.com/henrylee2cn/surfer"
	"net/http"
)

var SurferDownloader = surfer.New()

func Download(req *Request) (*http.Response, error) {
	return SurferDownloader.Download(req)
}
