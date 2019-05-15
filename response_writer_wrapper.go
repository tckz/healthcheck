package healthcheck

import "net/http"

// ResponseWriterWrapper StatusCodeを後から参照できるようにする
type ResponseWriterWrapper struct {
	http.ResponseWriter

	StatusCode int

	orgWriter http.ResponseWriter
}

func NewResponseWriterWrapper(w http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{
		StatusCode: http.StatusOK,
		orgWriter:  w,
	}
}

func (w *ResponseWriterWrapper) Header() http.Header {
	return w.orgWriter.Header()
}

func (w *ResponseWriterWrapper) Write(b []byte) (int, error) {
	return w.orgWriter.Write(b)
}

func (w *ResponseWriterWrapper) WriteHeader(code int) {
	w.orgWriter.WriteHeader(code)
	w.StatusCode = code
}
