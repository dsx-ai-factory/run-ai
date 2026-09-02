/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package iam

import (
	"context"
	"fmt"

	iamAuth "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/repositories/auth"
	authflow "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/repositories/auth_flow"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/repositories/clients"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/repositories/idps"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/repositories/realms"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/repositories/users"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/runaicloak"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/services"
)

type IamServiceContainer struct {
	UserService      services.UserService
	RealmService     services.RealmService
	AuthService      services.AuthService
	IdpService       services.IdpService
	ClientService    services.ClientService
	AuthFlowService  services.AuthFlowService
	authenticator    *iamAuth.KeycloakAuthenticator
	runaicloakClient runaicloak.RunAICloakClient
}

type ReposList struct {
	UserRepo     users.UserRepository
	RealmRepo    realms.RealmRepository
	AuthRepo     auth.AuthRepository
	IdpRepo      idps.IdpRepository
	ClientRepo   clients.ClientRepository
	AuthFlowRepo authflow.AuthFlowRepository
}

func NewIamServiceContainer(repos ReposList) *IamServiceContainer {
	return &IamServiceContainer{
		UserService:     services.NewUserService(repos.UserRepo),
		RealmService:    services.NewRealmService(repos.RealmRepo),
		AuthService:     services.NewAuthService(repos.AuthRepo),
		IdpService:      services.NewIdpService(repos.IdpRepo),
		ClientService:   services.NewClientService(repos.ClientRepo),
		AuthFlowService: services.NewAuthFlowService(repos.AuthFlowRepo),
	}
}

func NewIamServices(url, username, password string) *IamServiceContainer {
	keycloakHost := fmt.Sprintf("%s/auth", url)
	kcAuth := iamAuth.NewKeycloakAuthenticator(iamAuth.Configuration{
		Host:     keycloakHost,
		Username: username,
		Password: password,
	})

	runaicloakClient := runaicloak.NewClient(keycloakHost)

	return &IamServiceContainer{
		UserService:      services.NewUserService(users.NewKeycloakUserRepository(kcAuth, runaicloakClient)),
		RealmService:     services.NewRealmService(realms.NewKeycloakRealmRepository(kcAuth, runaicloakClient)),
		AuthService:      services.NewAuthService(auth.NewKeycloakAuthRepository(kcAuth, runaicloakClient)),
		IdpService:       services.NewIdpService(idps.NewKeycloakIdpRepository(kcAuth, runaicloakClient)),
		ClientService:    services.NewClientService(clients.NewKeycloakClientRepository(kcAuth, runaicloakClient)),
		AuthFlowService:  services.NewAuthFlowService(authflow.NewKeycloakAuthRepository(kcAuth)),
		authenticator:    kcAuth,
		runaicloakClient: runaicloakClient,
	}
}

func (i *IamServiceContainer) SetLoginType(ctx context.Context, loginType iamAuth.LoginType) error {
	return i.authenticator.SetLoginType(ctx, loginType)
}

func (i *IamServiceContainer) FetchKeycloakConfiguration(ctx context.Context, realm string) (map[string]interface{}, error) {
	return i.runaicloakClient.FetchKeycloakConfiguration(ctx, realm)
}

func (i *IamServiceContainer) FetchOAuthAuthorizationServerMetadata(ctx context.Context, realm string) (map[string]interface{}, error) {
	return i.runaicloakClient.FetchOAuthAuthorizationServerMetadata(ctx, realm)
}

func NewTestIamServiceContainer(runaicloakClient runaicloak.RunAICloakClient) *IamServiceContainer {
	return &IamServiceContainer{
		runaicloakClient: runaicloakClient,
	}
}
