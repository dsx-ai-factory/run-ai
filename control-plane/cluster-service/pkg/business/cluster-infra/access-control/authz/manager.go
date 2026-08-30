package authz

import (
	authorizationclient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
)

type AuthzManager struct {
	authorizationClient authorizationclient.IAuthorizationClient
}

func NewAuthzManager(authorizationClient authorizationclient.IAuthorizationClient) *AuthzManager {
	return &AuthzManager{
		authorizationClient: authorizationClient,
	}
}
