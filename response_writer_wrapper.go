package healthcheck

import "net/http"

type ResponseWriterWrapper struct {
	http.ResponseWriter

	StatusCode int
}

func NewResponseWriterWrapper(w http.ResponseWriter) *ResponseWriterWrapper {
	return &ResponseWriterWrapper{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}
}

func (w *ResponseWriterWrapper) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
	w.StatusCode = code
}
