package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"

	"log"
)

const (
	fileUrl  = "https://vd3.bdstatic.com/mda-kgkaebxu40y4d7th/v2-hknm/sc/mda-kgkaebxu40y4d7th.mp4?v_from_s=hkapp-haokan-hnb&amp;auth_key=1663056622-0-0-bf87e34d6342e8519b3f5c611ab5a7e5&amp;bcevod_channel=searchbox_feed&amp;cd=0&amp;pd=1&amp;pt=3&amp;logid=2421828632&amp;vid=5381429901807271381&amp;abtest=104378_2&amp;klogid=2421828632"
	htmlFile = "https://www.baidu.com/"
)

var (
	logger *log.Logger
)

func init() {
	logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
}

func downloadEx() {
	url := fileUrl
	headers := make(map[string]string)
	res, err := Request(http.MethodGet, url, nil, headers)
	if err != nil {
		logger.Println(`Request err=`, err)
		return
	}
	defer res.Body.Close() // nolint
	filepath := "ex.mp4"
	out, err := os.Create(filepath)
	if err != nil {
		logger.Println(`Create err=`, err)
		return
	}
	defer out.Close()
	written, copyErr := io.Copy(out, res.Body)
	logger.Printf(`written=%d,copyErr=%+v`, written, copyErr)
}

func downloadSimple() {
	// 文件url需要修改成目标地址
	url := htmlFile
	logger.Println(`start download...`)
	err := DownloadFile("22.mp4", url)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Downloaded: " + url)
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	written, copyErr := io.Copy(out, resp.Body)
	logger.Printf(`written=%d,copyErr=%+v`, written, copyErr)
	return err
}

// FakeHeadersVideo fake http headers
var FakeHeadersVideo = map[string]string{
	"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	"Accept-Charset":  "UTF-8,*;q=0.5",
	"Accept-Encoding": "gzip,deflate,sdch",
	"Accept-Language": "en-US,en;q=0.8",
	"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.81 Safari/537.36",
}

// FakeHeaders fake http headers html文件也能下载
var FakeHeaders = map[string]string{
	"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
	"Connection":      "keep-alive",
	"Accept-Encoding": "*",
	"Accept-Charset":  "UTF-8,*;q=0.5",
	"Accept-Language": "en-US,en;q=0.8",
	"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.81 Safari/537.36",
}

const retryTimes = 3

func Request(method, url string, body io.Reader, headers map[string]string) (*http.Response, error) {
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		DisableCompression:  true,
		TLSHandshakeTimeout: 10 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   15 * time.Minute,
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for k, v := range FakeHeaders {
		req.Header.Set(k, v)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	if _, ok := headers["Referer"]; !ok {
		req.Header.Set("Referer", url)
	}

	var (
		res          *http.Response
		requestError error
	)
	for i := 0; ; i++ {
		res, requestError = client.Do(req)
		if requestError == nil && res.StatusCode < 400 {
			break
		} else if i+1 >= retryTimes {
			var err error
			if requestError != nil {
				err = errors.Errorf("request error: %v", requestError)
			} else {
				err = errors.Errorf("%s request error: HTTP %d", url, res.StatusCode)
			}
			return nil, errors.WithStack(err)
		}
		time.Sleep(1 * time.Second)
	}
	return res, nil
}

func ServeHTTPLihuanying(w http.ResponseWriter, r *http.Request) {
	video, err := os.Open("./22.mp4")
	if err != nil {
		log.Fatal(err)
	}
	defer video.Close()

	http.ServeContent(w, r, "lihuanying.mp4", time.Now(), video)
}

func main() {
	// 视频播放服务器
	http.HandleFunc("/22", ServeHTTPLihuanying)
	http.ListenAndServe(":8080", nil)
	return
}
