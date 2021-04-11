package inputs

import (
	"encoding/base64"
	"net/url"
)

// FileInfo encapsulates information about an uploaded file
type FileInfo struct {
	SHA1    string
	Name    string
	Size    string
	Encoded string
}

// MarshalBase64 will encode the fields of this object
// as a base 64 encoded url key value string
func (fi *FileInfo) MarshalBase64() ([]byte, error) {
	vls := url.Values{}
	vls.Set("sha1", fi.SHA1)
	vls.Set("name", fi.Name)
	vls.Set("size", fi.Size)
	return []byte(base64.StdEncoding.EncodeToString([]byte(vls.Encode()))), nil
}

// UnmarshalBase64 populate file info fields from a b64 encoded url key value string
func (fi *FileInfo) UnmarshalBase64(in []byte) error {
	query := make([]byte, len(in))
	count, err := base64.StdEncoding.Decode(query, in)
	if err != nil {
		return err
	}
	query = query[:count]
	vls, err := url.ParseQuery(string(query))
	if err != nil {
		return err
	}
	fi.SHA1 = vls.Get("sha1")
	fi.Name = vls.Get("name")
	fi.Size = vls.Get("size")
	fi.Encoded = string(in)
	return nil
}
