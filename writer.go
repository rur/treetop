package idea

import "net/http"

type dataWriter struct {
	// TODO
	writer        http.ResponseWriter
	responseToken string
}

// Implement http.ResponseWriter interface by delegating to embedded instance
func (dw *dataWriter) Header() http.Header {
	return dw.writer.Header()
}

// Implement http.ResponseWriter interface by delegating to embedded instance
func (dw *dataWriter) Write(b []byte) (int, error) {
	return dw.writer.Write(b)
}

// Implement http.ResponseWriter interface by delegating to embedded instance
func (dw *dataWriter) WriteHeader(statusCode int) {
	dw.writer.WriteHeader(statusCode)
}

// Handler pass down data for template execution
func (dw *dataWriter) Data(d interface{}) {
	// TODO: Implement
}

//
func (dw *dataWriter) Status(status int) {
	// TODO: Implement
}

func (dw *dataWriter) BlockData(name string, req *http.Request) (interface{}, bool) {
	// TODO: Implement
	return nil, false
}

func (dw *dataWriter) ResponseToken() string {
	// TODO: Implement
	return dw.responseToken
}
