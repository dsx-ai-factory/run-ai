package token

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	api "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/config"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/httputils"
)

var (
	ErrResponseIsNil = errors.New("response is nil")
	ErrTokenNotFound = errors.New("your session has expired")
)

type LogoutAPI interface {
	api.LogoutAPI
}

type ControlPlaneTokenRevoker struct {
	configuration *api.Configuration
	client        api.LogoutAPI
}

func NewControlPlaneTokenRevoker(client *api.APIClient) *ControlPlaneTokenRevoker {
	return &ControlPlaneTokenRevoker{configuration: client.GetConfig(), client: client.LogoutAPI}
}

func (ctr *ControlPlaneTokenRevoker) Revoke(ctx context.Context, token string) error {
	ctr.configuration.AddDefaultHeader(config.AuthorizationHeader, fmt.Sprintf(config.BearerTokenStr, token))
	response, err := ctr.client.Logout(ctx).Execute()

	if response == nil {
		if err == nil {
			return ErrResponseIsNil
		}
		return err
	}

	defer func() {
		if response.Body != nil {
			_ = response.Body.Close()
		}
	}()

	if response.StatusCode == http.StatusNotFound {
		return ErrTokenNotFound
	}

	if !httputils.CheckResponseStatusCodeOK(response) {
		return fmt.Errorf("failed to revoke token: %s", response.Status)
	}

	return err
}
