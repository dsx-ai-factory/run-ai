package configuration

import "github.com/run-ai/runai/runai-common-packages/auth/providers/oidc"

type Provider interface {
	Configuration() (*oidc.Configuration, error)
	UserInfo(userIdentifier string) (*oidc.UserInfo, error)
}
