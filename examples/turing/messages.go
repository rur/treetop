package turing

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/json"
	"io"
)

// Messages is used to serialize the state of the current chat conversation
type Messages struct {
	Hello  string
	Number int
	Data   []float64
}

// infoTypeDictionary is the DEFLATE dictionary used for compressing and
// decompressing JSON marshalled from Messages type
var infoTypeDictionary = []byte(`{"Hello":"","Number":,"Data":[`)

func (i *Messages) MarshalBase64() (string, error) {
	jsonD, err := json.Marshal(i)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	// Compress the data using the specially crafted dictionary.
	zw, err := flate.NewWriterDict(
		&b, flate.DefaultCompression,
		infoTypeDictionary,
	)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(zw, bytes.NewReader(jsonD)); err != nil {
		return "", err
	}
	if err := zw.Close(); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

func (i *Messages) UnmarshalBase64(in string) error {
	return nil
}
