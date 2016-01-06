package main

import (
	"github.com/henrylee2cn/surfer"
	"io/ioutil"
	"log"
	"time"
)

func main() {
	// 默认使用surf内核下载
	log.Println("********************************************* surf内核下载测试开始 *********************************************")
	resp, err := surfer.Download(&surfer.DefaultRequest{
		Url: "http://github.com/henrylee2cn/surfer",
	})
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	log.Println(string(b), err)

	log.Println("********************************************* surf内核下载测试完毕 *********************************************")

	log.Println("********************************************* phantomjs内核下载测试开始 *********************************************")

	// 指定使用phantomjs内核下载
	resp, err = surfer.Download(&surfer.DefaultRequest{
		Url:          "http://github.com/henrylee2cn",
		DownloaderID: 1,
	})
	if err != nil {
		log.Fatal(err)
	}
	b, err = ioutil.ReadAll(resp.Body)
	log.Println(string(b), err)

	log.Println("********************************************* phantomjs内核下载测试完毕 *********************************************")

	resp.Body.Close()

	surfer.DestroyJsFiles()

	time.Sleep(600e9)
}
