package auth

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	api "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth/pkce"
)

const (
	LoginLocationKeyHeader = "Absolute-Location"
	RemoteBrowserFlow      = "external-browser"
)

// LoginApi is an interface that wraps the LoginsApi interface to allow for easier testing
type LoginApi interface {
	api.LoginsAPI
}

type ControlPlaneUrlGetter struct {
	client        LoginApi
	userAgent     string
	redirectUrl   string
	remoteBrowser bool
	pkceParams    pkce.Params
}

func (c *ControlPlaneUrlGetter) GetAuthUrl(ctx context.Context) (*url.URL, error) {
	request := c.buildRequest(ctx)
	response, err := request.Execute() //nolint:bodyclose // close is handle in the handler
	if response == nil && err != nil {
		return nil, err
	}
	return handleAuthURlResponse(response)
}

func (c *ControlPlaneUrlGetter) buildRequest(ctx context.Context) api.ApiGetLoginScreenURLRequest {
	request := c.client.GetLoginScreenURL(ctx).UserAgent(c.userAgent)
	if c.remoteBrowser {
		request = request.AuthenticationFlow(RemoteBrowserFlow)
	} else if c.redirectUrl != "" {
		request = request.RedirectUrl(c.redirectUrl)
	}
	if !c.pkceParams.IsZero() {
		request = request.CodeChallenge(c.pkceParams.CodeChallenge).
			CodeChallengeMethod(api.CodeChallengeMethod(c.pkceParams.CodeChallengeMethod))
	}
	return request
}

func handleAuthURlResponse(response *http.Response) (*url.URL, error) {
	if response.StatusCode != http.StatusTemporaryRedirect {
		return nil, errors.New("unexpected status code")
	}
	absoluteLocation := response.Header.Get(LoginLocationKeyHeader)
	if absoluteLocation == "" {
		return nil, errors.New("no location header in response")
	}
	return url.Parse(absoluteLocation)
}

func NewControlPlaneUrlGetter(client LoginApi, userAgent string, redirectUrl string, remoteBrowser bool, pkceParams pkce.Params) *ControlPlaneUrlGetter {
	return &ControlPlaneUrlGetter{
		client:        client,
		userAgent:     userAgent,
		redirectUrl:   redirectUrl,
		remoteBrowser: remoteBrowser,
		pkceParams:    pkceParams,
	}
}
