package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"time"
)

const URL = "http://139.224.22.239/upload/trans"
const TestURL = "http://localhost:12700/trans?name=%s&off=%d"

func main() {
	file, err := os.Open("test.txt")
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()
	client := initClient()
	var endErr error
	for i := 0; endErr == nil; i++ {
		content := make([]byte, 8<<24)
		offset := (8 << 24) * int64(i)
		n, err := file.ReadAt(content, offset)
		if err != nil {
			if err == io.EOF {
				endErr = err
			} else {
				log.Println(err)
				return
			}
		}
		log.Println(n)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		fw, err := writer.CreateFormFile("content", file.Name())
		if err != nil {
			log.Println(err)
			return
		}
		_, err = io.Copy(fw, bytes.NewReader(content))
		if err != nil {
			log.Println(err)
			return
		}
		writer.Close()
		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf(TestURL, file.Name(), offset), body)
		if err != nil {
			log.Println(err)
			return
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			return
		}
		defer resp.Body.Close()
		res, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			continue
		}
		log.Println("resp:", string(res))
	}
}

func initClient() *http.Client {
	dialer := net.Dialer{Timeout: 5 * time.Second, KeepAlive: 10 * time.Second}
	return &http.Client{Transport: &http.Transport{DialContext: dialer.DialContext}, Timeout: 7 * time.Second}
}
