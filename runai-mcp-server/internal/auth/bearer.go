package auth

import (
	"context"
	"fmt"
	"net/http"
)

const (
	authHeaderKey = "Authorization"
	bearerPrefix  = "Bearer "
)

type tokenCtxKey struct{}

func WithRequestToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenCtxKey{}, token)
}

func requestToken(ctx context.Context) string {
	v, _ := ctx.Value(tokenCtxKey{}).(string)
	return v
}

type BearerPassthrough struct{}

func NewBearerPassthrough() *BearerPassthrough {
	return &BearerPassthrough{}
}

func (b *BearerPassthrough) Token(ctx context.Context) (string, error) {
	tok := requestToken(ctx)
	if tok == "" {
		return "", fmt.Errorf("missing bearer token on request context")
	}
	return tok, nil
}

func (b *BearerPassthrough) RoundTrip(req *http.Request) (*http.Response, error) {
	tok, err := b.Token(req.Context())
	if err != nil {
		return nil, err
	}
	// net/http forbids mutating an inbound request.
	clonedReq := req.Clone(req.Context())
	clonedReq.Header.Set(authHeaderKey, bearerPrefix+tok)
	return http.DefaultTransport.RoundTrip(clonedReq)
}

func (b *BearerPassthrough) HttpClient() *http.Client {
	return &http.Client{Transport: b}
}
