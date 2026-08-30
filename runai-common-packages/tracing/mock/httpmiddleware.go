package mocks

import (
	"net/http"
)

type mockHTTPMiddleware struct{}

func (m *mockHTTPMiddleware) Middleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	})
}

func NewMockHTTPMiddleware() *mockHTTPMiddleware {
	return &mockHTTPMiddleware{}
}
