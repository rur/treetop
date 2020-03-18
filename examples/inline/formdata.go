package inline

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
)

// FormData is used to serialize the state of the inline example form
type FormData struct {
	FirstName   string
	LastName    string
	Email       string
	Country     string
	Description string
}

// formDataTypeDict is the DEFLATE dictionary used for compressing and
// decompressing JSON marshalled from FormData type
var formDataTypeDict = []byte(`{"FirstName":,"LastName": ,"Email": ,"Country": ,"Description":`)

func (i *FormData) MarshalBase64() ([]byte, error) {
	jsonD, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	var b bytes.Buffer
	// Compress the data using the specially crafted dictionary.
	zw, err := flate.NewWriterDict(
		&b, flate.DefaultCompression,
		formDataTypeDict,
	)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(zw, bytes.NewReader(jsonD)); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}

	return []byte(base64.StdEncoding.EncodeToString(b.Bytes())), nil
}

func (i *FormData) UnmarshalBase64(in []byte) error {
	compressed := make([]byte, len(in))
	count, err := base64.StdEncoding.Decode(compressed, in)
	if err != nil {
		return err
	}
	compressed = compressed[:count]

	zr := flate.NewReaderDict(
		bytes.NewReader(compressed),
		formDataTypeDict,
	)
	defer zr.Close()

	raw, err := ioutil.ReadAll(zr)
	if err != nil {
		return err
	}

	return json.Unmarshal(raw, i)
}

func getDefaultFormData() *FormData {
	return &FormData{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john.doe@gmail.com",
		Country:     "Canada",
		Description: "test",
	}
}
