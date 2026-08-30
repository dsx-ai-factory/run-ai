package wellknown

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc/v3/oidc"
	roidc "github.com/run-ai/runai/runai-common-packages/auth/providers/oidc"
	"github.com/run-ai/runai/runai-common-packages/auth/providers/oidc/configuration"
	"golang.org/x/oauth2"
)

type ConfigurationProvider struct {
	oidcConfigProvider *oidc.Provider
}

func New(issuerURL string) (configuration.Provider, error) {
	provider, err := oidc.NewProvider(context.Background(), issuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed fetching oidc configuration: %v", err)
	}

	return &ConfigurationProvider{
		oidcConfigProvider: provider,
	}, nil
}

func (cp *ConfigurationProvider) Configuration() (*roidc.Configuration, error) {
	oidcConfig := &roidc.Configuration{}
	err := cp.oidcConfigProvider.Claims(oidcConfig)
	if err != nil {
		return nil, fmt.Errorf("failed marshalling oidc configuration request: %v", err)
	}

	return oidcConfig, nil
}

func (cp *ConfigurationProvider) UserInfo(userAccessToken string) (*roidc.UserInfo, error) {
	userInfo, err := cp.oidcConfigProvider.UserInfo(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: userAccessToken,
	}))
	if err != nil {
		return nil, err
	}

	unmarshalledUserInfo := &roidc.UserInfo{}
	err = userInfo.Claims(unmarshalledUserInfo)
	if err != nil {
		return nil, err
	}

	return unmarshalledUserInfo, nil
}
