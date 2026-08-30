package auth

import (
	"context"
	"net/http"
)

type Provider interface {
	Token(ctx context.Context) (string, error)
	http.RoundTripper
	HttpClient() *http.Client
}
