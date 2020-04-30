package inputs

import (
	"encoding/base64"
	"net/url"
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
