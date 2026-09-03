package userinfo

import "github.com/dsx-ai-factory/run-ai/runai-common-packages/auth/providers/oidc"

type Provider interface {
	Load(userIdentifier string) (*oidc.UserInfo, error)
}
