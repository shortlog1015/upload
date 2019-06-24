package main

import (
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	http.Handle("/", http.StripPrefix("", http.FileServer(http.Dir("./public"))))
	http.HandleFunc("/sub", getUploadFile)
	http.HandleFunc("/trans", getTransFile)
	http.ListenAndServe(":12700", nil)
}

func getUploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "wrong method", http.StatusNotAcceptable)
		return
	}
	if err := r.ParseMultipartForm(8 << 24); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fileHeaders, err := getMultiPartFiles(r, "files")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, fh := range fileHeaders {
		srcFile, err := fh.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusResetContent)
			return
		}
		defer srcFile.Close()
		dstFile, err := os.Create("./files/" + fh.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusResetContent)
			return
		}
		defer dstFile.Close()
		if _, err := io.Copy(dstFile, srcFile); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("files ok"))
	}
}

func getMultiPartFiles(r *http.Request, key string) ([]*multipart.FileHeader, error) {
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		if fhs := r.MultipartForm.File[key]; len(fhs) > 0 {
			return fhs, nil
		}
	}
	return nil, http.ErrMissingFile
}

func getTransFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "wrong method", http.StatusNotAcceptable)
		return
	}
	if err := r.ParseMultipartForm(8 << 24); err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	query := r.URL.Query()
	name := query.Get("name")
	off, _ := strconv.Atoi(query.Get("off"))
	log.Println(name, off)
	if name == "" {
		http.Error(w, "no name", http.StatusBadRequest)
		return
	}
	fileHeaders, err := getMultiPartFiles(r, "content")
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, fh := range fileHeaders {
		content, err := getSrcContent(fh)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusResetContent)
			return
		}

		dstFile, err := os.OpenFile("./files/"+name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusResetContent)
			return
		}
		defer dstFile.Close()

		_, err = dstFile.WriteAt(content, int64(off))
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("trans file ok"))
	}
}

func exist(name string) bool {
	if _, err := os.Stat("./files/" + name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func getSrcContent(fh *multipart.FileHeader) (content []byte, err error) {
	srcFile, err := fh.Open()
	if err != nil {
		return
	}
	defer srcFile.Close()
	content = make([]byte, fh.Size)
	_, err = srcFile.Read(content)
	return
}
