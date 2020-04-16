package handlers

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/rur/treetop"
	"github.com/rur/treetop/examples/ticket/inputs"
)

// Method: POST
// Doc: Load a list of uploaded files, save to storage and return metadata
func UploadedFilesHandler(rsp treetop.Response, req *http.Request) interface{} {
	data := struct {
		Files []*inputs.FileInfo
	}{}

	if err := req.ParseMultipartForm(1024 * 1024 * 16 /*16 MiB*/); err != nil {
		log.Println(err)
		http.Error(rsp, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return nil
	}

	for _, fh := range req.MultipartForm.File["file-upload"] {
		info := inputs.FileInfo{
			Name: fh.Filename,
		}
		if fh.Size < 2048 {
			info.Size = fmt.Sprintf("%dB", fh.Size)
		} else {
			info.Size = fmt.Sprintf("%.0fKB", float64(fh.Size)/1024)
		}
		f, err := fh.Open()
		if err != nil {
			log.Println(err)
			http.Error(rsp, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		}
		defer f.Close()

		h := sha1.New()
		if _, err := io.Copy(h, f); err != nil {
			log.Println(err)
			http.Error(rsp, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		}
		info.SHA1 = fmt.Sprintf("%x", h.Sum(nil))
		if b64, err := info.MarshalBase64(); err != nil {
			log.Println(err)
			http.Error(rsp, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return nil
		} else {
			info.Encoded = string(b64)
		}
		data.Files = append(data.Files, &info)
	}
	return data
}

// AttachmentFileListHandler decoded URL query encoded attachment file info
// Extends: attachment-file-list
// Method: GET
// Doc: Default software attachment file list template handler,
//      parse file info from query string
func AttachmentFileListHandler(rsp treetop.Response, req *http.Request) interface{} {
	// load file info from query
	query := req.URL.Query()
	data := struct {
		Files []*inputs.FileInfo
	}{}

	for _, enc := range query["attachment"] {
		info := &inputs.FileInfo{}
		if err := info.UnmarshalBase64([]byte(enc)); err != nil {
			// skip it
			log.Println(err)
		} else {
			data.Files = append(data.Files, info)
		}
	}
	return data
}
