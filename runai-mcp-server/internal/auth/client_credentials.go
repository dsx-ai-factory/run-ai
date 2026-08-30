package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	idmclient "github.com/run-ai/runai/control-plane/identity-manager/pkg/client"
)

type ClientCredentials struct {
	mgr *idmclient.AuthnManager
}

func NewClientCredentials(ctx context.Context, baseURL, clientID, clientSecret string) (*ClientCredentials, error) {
	baseURL = strings.TrimRight(baseURL, "/")
	idmClient := idmclient.NewIdmExternalClient(baseURL)
	mgr, err := idmclient.NewAuthnManager(ctx, idmclient.AuthnManagerConfig{
		ClientId:     clientID,
		ClientSecret: clientSecret,
		IdmClient:    idmClient,
	})
	if err != nil {
		return nil, fmt.Errorf("client credentials: %w", err)
	}
	return &ClientCredentials{mgr: mgr}, nil
}

func (c *ClientCredentials) Token(ctx context.Context) (string, error) {
	return c.mgr.Token(ctx)
}

func (c *ClientCredentials) RoundTrip(req *http.Request) (*http.Response, error) {
	return c.mgr.RoundTrip(req)
}

func (c *ClientCredentials) HttpClient() *http.Client {
	return &http.Client{Transport: c}
}
