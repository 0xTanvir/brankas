package main

import (
	"crypto/sha1"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var tpl *template.Template

// iamgeTable will be used for comparing image file type
var iamgeTable = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
}

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
}

func main() {
	http.Handle("/", http.HandlerFunc(index))
	http.Handle("/upload", http.HandlerFunc(upload))
	http.ListenAndServe(":8080", nil)
}

// index endpoint
func index(res http.ResponseWriter, req *http.Request) {
	tpl.ExecuteTemplate(res, "index.tpl", nil)
}

// upload endpoint
func upload(res http.ResponseWriter, req *http.Request) {

	// reading the auth from field
	auth := req.PostFormValue("auth")
	if auth != os.Getenv("BRANKAS_AUTH") {
		res.WriteHeader(http.StatusForbidden)
		return
	}

	// reading the file from field
	mf, fh, err := req.FormFile("upload")
	if err != nil {
		fmt.Println(err)
	}
	defer mf.Close()

	// comparing if file size is less than 8MB
	if fh.Size > 8000000 {
		res.WriteHeader(http.StatusForbidden)
		return
	}

	// detecting the content type of a file
	mimeType, err := getContentType(mf)
	if err != nil {
		fmt.Println(err)
	}
	if ok := iamgeTable[mimeType]; !ok {
		res.WriteHeader(http.StatusForbidden)
		return
	}

	// create sha for file name
	ext := strings.Split(fh.Filename, ".")[1]
	h := sha1.New()
	io.Copy(h, mf)
	fname := fmt.Sprintf("%x", h.Sum(nil)) + "." + ext
	// create new file
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	path := filepath.Join(wd, "public", "pics", fname)
	nf, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
	}
	defer nf.Close()
	// copy
	mf.Seek(0, 0)
	io.Copy(nf, mf)

	// save image info into database
	saveImageInfo(fname, mimeType, fh.Size)

}

func getContentType(seeker io.ReadSeeker) (string, error) {
	// At most the first 512 bytes of data are used:
	// https://golang.org/src/net/http/sniff.go?s=646:688#L11
	buff := make([]byte, 512)

	_, err := seeker.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	bytesRead, err := seeker.Read(buff)
	if err != nil && err != io.EOF {
		return "", err
	}

	// Slice to remove fill-up zero values which cause a wrong content type detection in the next step
	buff = buff[:bytesRead]

	return http.DetectContentType(buff), nil
}

// saveImageInfo saves image info into database
func saveImageInfo(fileName, contentType string, size int64) error {
	db, err := sql.Open("mysql", "root:tanvir@tcp(127.0.0.1:3306)/fakedb")
	if err != nil {
		return err
	}
	insert, err := db.Query("INSERT INTO img (file_name,content_type,size) VALUES (?,?,?)", fileName, contentType, size)
	if err != nil {
		return err
	}
	defer insert.Close()
	return nil
}
