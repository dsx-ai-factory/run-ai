package client

import (
	"net/http"
)

type ResponseStatusCodeCapture struct {
	http.ResponseWriter
	StatusCode int
}

func NewResponseWriterCapture(rw http.ResponseWriter) *ResponseStatusCodeCapture {
	return &ResponseStatusCodeCapture{
		ResponseWriter: rw,
		StatusCode:     http.StatusOK,
	}
}

func (r *ResponseStatusCodeCapture) Write(b []byte) (int, error) {
	return r.ResponseWriter.Write(b)
}

func (r *ResponseStatusCodeCapture) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
