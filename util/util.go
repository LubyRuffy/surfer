// Package util contains some utility methods used by other packages.
package util

import (
	"net/url"
	"os"
)

// FileExists returns if the given file exists.
func FileExists(file string) bool {
	if _, err := os.Stat(file); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}

	return true
}

// 返回编码后的url.URL指针、及解析错误
func UrlEncode(urlStr string) (*url.URL, error) {
	urlObj, err := url.Parse(urlStr)
	urlObj.RawQuery = urlObj.Query().Encode()
	return urlObj, err
}
