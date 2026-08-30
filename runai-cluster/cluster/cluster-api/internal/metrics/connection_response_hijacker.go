package metrics

import (
	"bufio"
	"net"
	"net/http"
)

type ResponseHijackerWrapper struct {
	http.ResponseWriter
	StatusCode int
}

func NewResponseHijackerWrapper(w http.ResponseWriter) *ResponseHijackerWrapper {
	return &ResponseHijackerWrapper{
		ResponseWriter: w,
		StatusCode:     http.StatusOK,
	}
}

func (r *ResponseHijackerWrapper) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *ResponseHijackerWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return r.ResponseWriter.(http.Hijacker).Hijack()
}

func (r *ResponseHijackerWrapper) Flush() {
	r.ResponseWriter.(http.Flusher).Flush()
}
