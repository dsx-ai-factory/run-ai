package userinfo

import "github.com/run-ai/runai/runai-common-packages/auth/providers/oidc"

type Provider interface {
	Load(userIdentifier string) (*oidc.UserInfo, error)
}
