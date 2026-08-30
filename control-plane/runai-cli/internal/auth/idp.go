package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/url"

	"golang.org/x/oauth2"
)

const oauthStateBytes = 16

type LoginUrlGetter interface {
	GetAuthUrl(ctx context.Context) (*url.URL, error)
}
type TokenExchanger interface {
	Exchange(ctx context.Context, code string, authURL *url.URL) (*oauth2.Token, error)
}

type AuthorizationCodeHandler interface {
	Handle(ctx context.Context, auth *url.URL) (string, error)
}

type IDPAuthenticator struct {
	urlGetter            LoginUrlGetter
	exchangeCodeForToken TokenExchanger
	browserHandler       AuthorizationCodeHandler
}

func NewIDPAuthenticator(urlGetter LoginUrlGetter,
	exchangeCodeForToken TokenExchanger, browserHandler AuthorizationCodeHandler) *IDPAuthenticator {
	return &IDPAuthenticator{
		urlGetter:            urlGetter,
		exchangeCodeForToken: exchangeCodeForToken,
		browserHandler:       browserHandler,
	}
}

func (idp *IDPAuthenticator) Authenticate(ctx context.Context) (*oauth2.Token, error) {
	authUrl, err := idp.urlGetter.GetAuthUrl(ctx)
	if err != nil {
		return nil, err
	}

	if authUrl.Scheme == "" || authUrl.Host == "" {
		return nil, fmt.Errorf("authentication URL is invalid, missing scheme or host: '%s'", authUrl.String())
	}

	state, err := generateOAuthState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OAuth state: %w", err)
	}
	q := authUrl.Query()
	q.Set("state", state)
	authUrl.RawQuery = q.Encode()

	authCode, err := idp.browserHandler.Handle(ctx, authUrl)
	if err != nil {
		return nil, err
	}

	code, err := idp.exchangeCodeForToken.Exchange(ctx, authCode, authUrl)
	if err != nil {
		return nil, err
	}

	return code, nil
}

func generateOAuthState() (string, error) {
	b := make([]byte, oauthStateBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
