package authn

import (
	idmclient "github.com/run-ai/runai/control-plane/identity-manager/pkg/client"
)

type AuthnManager struct {
	idmClient idmclient.IIdm
}

func NewAuthnManager(idmClient idmclient.IIdm) *AuthnManager {
	return &AuthnManager{
		idmClient: idmClient,
	}
}
