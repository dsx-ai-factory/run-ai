package configuration

import "github.com/dsx-ai-factory/run-ai/runai-common-packages/auth/providers/oidc"

type Provider interface {
	Configuration() (*oidc.Configuration, error)
	UserInfo(userIdentifier string) (*oidc.UserInfo, error)
}
