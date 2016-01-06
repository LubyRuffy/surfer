# surfer    [![GoDoc](https://godoc.org/github.com/tsuna/gohbase?status.png)](https://godoc.org/github.com/henrylee2cn/surfer) [![GitHub release](https://img.shields.io/github/release/henrylee2cn/surfer.svg)](https://github.com/henrylee2cn/surfer/releases)

High Level Concurrency Downloader
</br>
surfer是一款Go语言编写的高并发爬虫下载器，拥有surf与phantom两种下载内核。
</br>
支持固定UserAgent自动保存cookie与随机大量UserAgent禁用cookie两种模式，高度模拟浏览器行为，可实现模拟登录等功能。
</br>
它是高并发爬虫[Pholcus](https://github.com/henrylee2cn/pholcus)的专用下载器。
</br>
（官方QQ群：Go大数据 42731170，欢迎加入我们的讨论）
</br>

```
package main

import (
    "github.com/henrylee2cn/surfer"
    "io/ioutil"
    "log"
    "time"
)

func main() {
    // 默认使用surf内核下载
    log.Println("************************************ surf内核下载测试开始 ************************************")
    resp, err := surfer.Download(&surfer.DefaultRequest{
        Url: "http://github.com/henrylee2cn/surfer",
    })
    if err != nil {
        log.Fatal(err)
    }
    b, err := ioutil.ReadAll(resp.Body)
    log.Println(string(b), err)

    log.Println("************************************ surf内核下载测试完毕 ************************************")

    log.Println("************************************ phantomjs内核下载测试开始 ************************************")

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

    log.Println("************************************ phantomjs内核下载测试完毕 ************************************")

    resp.Body.Close()

    surfer.DestroyJsFiles()

    time.Sleep(600e9)
}
```