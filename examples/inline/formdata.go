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

var CountryOptions = []string{
	"Canada",
	"Mexico",
	"USA",
}

// formDataTypeDict is the DEFLATE dictionary used for compressing and
// decompressing JSON marshalled from FormData type
var formDataTypeDict = []byte(`{"FirstName":","LastName":","Email":","Country":"USACanadaMexico","Description":"`)

// MarshalBase64 encodes struct value for use as cookie value with limited characterset and available space.
// The struct value is encoded to a JSON byte array, compress and return base 64 encoding.
func (fd *FormData) MarshalBase64() ([]byte, error) {
	jsonD, err := json.Marshal(fd)
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

// UnmarshalBase64 assigns fields from a base64 encoded and DEFLATE compressed JSON string
func (fd *FormData) UnmarshalBase64(in []byte) error {
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

	return json.Unmarshal(raw, fd)
}

// getDefaultFormData is used to initialize the demo
func getDefaultFormData() *FormData {
	return &FormData{
		FirstName:   "John",
		LastName:    "Doe",
		Email:       "john.doe@gmail.com",
		Country:     "Canada",
		Description: "test",
	}
}

// SetField update struct with form field name
func (fd *FormData) SetField(field, value string) {
	switch field {
	case "firstName":
		fd.FirstName = value
	case "surname":
		fd.LastName = value
	case "email":
		fd.Email = value
	case "country":
		fd.Country = value
	case "description":
		fd.Description = value
	}
}
