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

const URL = "http://139.224.22.239/upload/trans"
const TestURL = "http://localhost:12700/trans?%s"

func main() {
	f := flag.String("f", "", "translate file")
	size := flag.Int64("s", 8<<24, "part size")
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
	// init client
	client := initClient()
	// start part translate
	for i := 0; ; i++ {
		// read part file return offset
		content, offset, end, err := readPartFile(file, i, *size)
		// set mulitpart file form
		name := path.Base(file.Name())
		body, contentType, err := writeMulitPart(content, name)
		if err != nil {
			log.Println(err)
			return
		}
		vals := url.Values{
			"name": {name},
			"off":  {strconv.Itoa(int(offset))},
		}
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(TestURL, vals.Encode()), body)
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
			continue
		}
		log.Printf("translate %d - %d ok", offset, offset+int64(len(content)))
		if end {
			break
		}
	}
}

func initClient() *http.Client {
	dialer := net.Dialer{Timeout: 5 * time.Second, KeepAlive: 10 * time.Second}
	return &http.Client{Transport: &http.Transport{DialContext: dialer.DialContext}, Timeout: 7 * time.Second}
}

func readPartFile(file *os.File, i int, size int64) (content []byte, offset int64, end bool, err error) {
	content = make([]byte, size)
	offset = size * int64(i)
	n, err := file.ReadAt(content, offset)
	if err != nil {
		end = err == io.EOF
		if !end {
			return
		}
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
