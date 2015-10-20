package surfer

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// 基于Phantomjs的下载器实现，作为surfer的补充
// 效率较surfer会慢很多，但是因为模拟浏览器，破防性更好

type Phantom struct {
	FullPhantomjsName string //Phantomjs完整文件名
	FullTempJsFile    string //js临时文件存放完整文件名
}

func NewPhantom(fullPhantomjsName, fullTempJsFile string) Surfer {
	phantom := &Phantom{
		FullPhantomjsName: fullPhantomjsName,
		FullTempJsFile:    fullTempJsFile,
	}
	phantom.creatFile()
	return phantom
}

// 实现surfer下载器接口
func (self *Phantom) Download(req Request) (resp *http.Response, err error) {
	resp = new(http.Response)
	ct := strings.ToLower(req.GetHeader().Get("Content-Type"))
	idx := strings.Index(ct, "charset=")
	if idx != -1 {
		ct = strings.Trim(string(ct[idx+8:]), ";")
		ct = strings.Trim(ct, " ")
	} else {
		ct = "utf-8"
	}

	ua := strings.ToLower(req.GetHeader().Get("User-Agent"))
	if ua == "" {
		resp.Body, err = self.download(req.GetUrl(), ct)
	} else {
		resp.Body, err = self.download(req.GetUrl(), ct, ua)
	}
	return
}

//url 为请求地址
//encoding 为页面编码
//userAgent 为客户端代理请求设备，默认为百度爬虫
func (self *Phantom) download(url string, encoding string, userAgent ...string) (stdout io.ReadCloser, err error) {
	args := []string{self.FullTempJsFile, url, encoding}
	if len(userAgent) > 0 {
		args = append(args, userAgent...)
	}
	cmd := exec.Command(self.FullPhantomjsName, args...)
	stdout, err = cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	return stdout, err
}

func (self *Phantom) creatFile() error {
	if !filepath.IsAbs(self.FullTempJsFile) {
		self.FullTempJsFile, _ = filepath.Abs(self.FullTempJsFile)
	}
	if !filepath.IsAbs(self.FullPhantomjsName) {
		self.FullPhantomjsName, _ = filepath.Abs(self.FullPhantomjsName)
	}

	// 创建/打开目录
	p, _ := filepath.Split(self.FullTempJsFile)
	d, err := os.Stat(p)
	if err != nil || !d.IsDir() {
		if err := os.MkdirAll(p, 0777); err != nil {
			return err
		}
	}

	// 创建并写入文件
	f, _ := os.Create(self.FullTempJsFile)
	f.Write([]byte(JS_CODE))
	f.Close()
	return nil
}

const (
	JS_CODE = `//system 用于
	var system = require('system');
	var page = require('webpage').create();
	// console.log(system.args[0],system.args[1],system.args[2])
	page.settings.userAgent = 'Mozilla/5.0+(compatible;+Baiduspider/2.0;++http://www.baidu.com/search/spider.html)';
	if(system.args.length ==1){
		phantom.exit();
	}else{
		var url = system.args[1];
		var encode = system.args[2];

		if(encode != undefined){
			//设置编码
			phantom.outputEncoding=encode;
		}
		if(system.args[3] != undefined){
			
			//设置客户端代理设备
			page.settings.userAgent = system.args[3]
		}
		
		page.open(url, function (status) {
		    if (status !== 'success') {
		        console.log('Unable to access network');
		    } else {
		        // var ua = page.evaluate(function () {
		        //     return page.content;
		        // });
		        console.log(page.content);
		    }
		    phantom.exit();
		});
	}`
)
