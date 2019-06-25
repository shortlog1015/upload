package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"
)

const URL = "http://139.224.22.239/upload/trans?%s"
const TestURL = "http://localhost:12700/trans?%s"

var client *http.Client

func main() {
	f := flag.String("f", "", "translate file")
	size := flag.Int64("s", 1<<19, "part size")
	flag.Parse()
	if *f == "" {
		log.Println("set f, translate file")
		return
	}
	// open file
	file, err := os.Open(*f)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	// get file size
	fi, _ := os.Stat(*f)
	all := fi.Size()
	// init client
	client = initClient()
	// create queue
	queue := newQueue()
	queue.start(10, dealTranslate)
	var count int
	if all%*size == 0 {
		count = int(all / *size)
	} else {
		count = int(all / *size) + 1
	}
	var s int64
	for i := 0; i < count; i++ {
		args := &Args{file: file, index: i, size: *size, n: make(chan int64)}
		queue.submit(args)
		s += <-args.n
		log.Printf("translate %d%%\n", int(float64(s)/float64(all)*100))
	}
	queue.stop()
}

func dealTranslate(args *Args) {
	// read part file return offset
	content, offset, err := readPartFile(args.file, args.index, args.size)
	defer func() { args.n <- int64(len(content)) }()
	if err != nil {
		log.Println(err)
		return
	}
	// set mulitpart file form
	name := path.Base(args.file.Name())
	body, contentType, err := writeMulitPart(content, name)
	if err != nil {
		log.Println(err)
		return
	}
	vals := url.Values{
		"name": {name},
		"off":  {strconv.Itoa(int(offset))},
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(URL, vals.Encode()), body)
	if err != nil {
		log.Println(err)
		return
	}
	// must set this head
	req.Header.Set("Content-Type", contentType)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		res, _ := ioutil.ReadAll(resp.Body)
		log.Printf("translate %d - %d failed: %s", offset, offset+int64(len(content)), string(res))
	}
	return
}

func initClient() *http.Client {
	dialer := net.Dialer{Timeout: 10 * time.Second, KeepAlive: 30 * time.Second}
	return &http.Client{Transport: &http.Transport{DialContext: dialer.DialContext}, Timeout: 10 * time.Second}
}

func readPartFile(file *os.File, i int, size int64) (content []byte, offset int64, err error) {
	content = make([]byte, size)
	offset = size * int64(i)
	n, err := file.ReadAt(content, offset)
	if err != nil && err != io.EOF {
		return
	} else {
		err = nil
	}
	if n < len(content) {
		content = content[:n]
	}
	return
}

func writeMulitPart(content []byte, name string) (body *bytes.Buffer, contentType string, err error) {
	body = &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormFile("content", name)
	if err != nil {
		return
	}
	_, err = io.Copy(fw, bytes.NewReader(content))
	if err != nil {
		return
	}
	contentType = writer.FormDataContentType()
	writer.Close()
	return
}
