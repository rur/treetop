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
	"United States",
}

// formDataTypeDict is the DEFLATE dictionary used for compressing and
// decompressing JSON marshalled from FormData type
var formDataTypeDict = []byte(`{"FirstName":","LastName":","Email":","Country":"United StatesCanadaMexico","Description":"`)

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

var defaultFormData = FormData{
	FirstName: "Theodore H.",
	LastName:  "Fakeman",
	Email:     "test@example.com",
	Country:   "United States",
	Description: `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed ` +
		`do eiusmod tempor incididunt ut labore et dolore magna aliqua.` + "\n\n" +
		`Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ` +
		`iquip ex ea commodo consequat.` + "\n\n" +
		`Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore ` +
		`eu fugiat nulla pariatur. Except eur sint occaecat cupidatat non proident, ` +
		`sunt in culpa qui officia de serunt mollit anim id est laborum.`,
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
