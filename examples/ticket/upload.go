package ticket

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/rur/treetop"
)

type FileInfo struct {
	SHA1    string
	Name    string
	Size    string
	Encoded string
}

func (fi *FileInfo) MarshalBase64() ([]byte, error) {
	vals := url.Values{}
	vals.Set("sha1", fi.SHA1)
	vals.Set("name", fi.Name)
	vals.Set("size", fi.Size)
	return []byte(base64.StdEncoding.EncodeToString([]byte(vals.Encode()))), nil
}

func (fi *FileInfo) UnmarshalBase64(in []byte) error {
	query := make([]byte, len(in))
	count, err := base64.StdEncoding.Decode(query, in)
	if err != nil {
		return err
	}
	query = query[:count]
	vals, err := url.ParseQuery(string(query))
	if err != nil {
		return err
	}
	fi.SHA1 = vals.Get("sha1")
	fi.Name = vals.Get("name")
	fi.Size = vals.Get("size")
	fi.Encoded = string(in)
	return nil
}

// Method: POST
// Doc: Load a list of uploaded files, save to storage and return metadata
func uploadedFilesHandler(rsp treetop.Response, req *http.Request) interface{} {
	data := struct {
		Files []*FileInfo
	}{}

	if err := req.ParseMultipartForm(1024 * 1024 * 16 /*16 MiB*/); err != nil {
		log.Println(err)
		http.Error(rsp, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return nil
	}

	for _, fh := range req.MultipartForm.File["file-upload"] {
		info := FileInfo{
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
