/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/Nerzal/gocloak/v13"
	gocloakInterface "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/gocloak"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils"
	ptr "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils/pointer"
)

const (
	MasterRealmName                    = "master"
	MasterRealmAdminClient             = "admin-cli"
	ServiceClientID                    = "runai-admin-app"
	DefaultServiceAccountTokenLifespan = 10 * time.Minute
	AdminRole                          = "admin"
	DefaultMasterRole                  = "default-roles-master"
)

const (
	ErrAdminLoginFailed = "authentication failed for admin user"
)

type UserAuthentication struct {
	username           string
	password           string
	userToken          *gocloak.JWT
	userTokenExpiresAt time.Time
}

type ServiceAccountAuthentication struct {
	appToken          *gocloak.JWT
	appTokenExpiresAt time.Time
	appSecret         *gocloak.CredentialRepresentation
}

type KeycloakAuthenticator struct {
	KcClient  gocloakInterface.GoCloak
	host      string
	loginType LoginType
	userAuth  *UserAuthentication
	appAuth   *ServiceAccountAuthentication
}

func NewKeycloakAuthenticator(config Configuration) *KeycloakAuthenticator {
	return &KeycloakAuthenticator{
		KcClient: buildKcClient(config),
		host:     config.Host,
		userAuth: &UserAuthentication{
			username: config.Username,
			password: config.Password,
		},
		loginType: LoginTypeUser,
		appAuth:   &ServiceAccountAuthentication{},
	}
}

func buildKcClient(config Configuration) *gocloak.GoCloak {
	client := gocloak.NewClient(config.Host)
	restyClient := client.RestyClient()
	restyClient.BaseURL = config.Host
	return client
}

func (k *KeycloakAuthenticator) GetHost() string {
	return k.host
}

func (k *KeycloakAuthenticator) SetLoginType(ctx context.Context, loginType LoginType) error {
	err := k.userAdminLogin(ctx)
	if err != nil {
		return fmt.Errorf("%s. error: %w", ErrAdminLoginFailed, err)
	}
	if loginType == LoginTypeServiceAccount {
		err = k.setAppLogin(ctx)
		if err != nil {
			return err
		}
		k.loginType = loginType
	}
	return nil
}

func (k *KeycloakAuthenticator) GetAdminToken(ctx context.Context) (*string, error) {
	if k.loginType == LoginTypeServiceAccount {
		return k.getAppAdminAccessToken(ctx)
	} else {
		return k.getUserAdminAccessToken(ctx)
	}
}

func (k *KeycloakAuthenticator) getAppAdminAccessToken(ctx context.Context) (*string, error) {
	if k.appAuth.isTokenValid() {
		return k.appAuth.getAccessToken(), nil
	}
	err := k.setAppAdminToken(ctx)
	if err != nil {
		return nil, err
	}
	return k.appAuth.getAccessToken(), nil
}

func (k *KeycloakAuthenticator) getUserAdminAccessToken(ctx context.Context) (*string, error) {
	if k.userAuth.isTokenValid() {
		return k.userAuth.getAccessToken(), nil
	}
	if err := k.refreshToken(ctx); err != nil {
		err := k.userAdminLogin(ctx)
		if err != nil {
			return nil, err
		}
	}
	return k.userAuth.getAccessToken(), nil
}

func (k *KeycloakAuthenticator) setAppLogin(ctx context.Context) error {
	idOfClient, err := k.getAppAdminClientID(ctx)
	if err != nil {
		return err
	}
	err = k.setAppAdminSecret(ctx, *idOfClient)
	if err != nil {
		return err
	}
	err = k.appAdminLogin(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (k *KeycloakAuthenticator) getAppAdminClientID(ctx context.Context) (*string, error) {
	client, _ := k.getAppAdminClient(ctx)
	if client != nil {
		return client.ID, nil
	}
	clientID, err := k.createNewAppAdminClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s. error: %w", "failed to create admin application", err)
	}
	return &clientID, nil
}

func (k *KeycloakAuthenticator) refreshToken(ctx context.Context) error {
	if k.userAuth.userToken == nil || k.userAuth.userToken.RefreshToken == "" {
		return k.userAdminLogin(ctx)
	}
	newToken, err := k.KcClient.RefreshToken(ctx, k.userAuth.userToken.RefreshToken, MasterRealmAdminClient, "", MasterRealmName)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}
	k.userAuth.setToken(newToken)
	return nil
}

func (k *KeycloakAuthenticator) setAppAdminToken(ctx context.Context) error {
	err := k.setAppSecret(ctx)
	if err != nil {
		return err
	}
	err = k.appAdminLogin(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (k *KeycloakAuthenticator) setAppSecret(ctx context.Context) error {
	if secret := k.appAuth.getClientSecret(); secret != nil {
		return nil
	}
	err := k.setNewAppSecret(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (k *KeycloakAuthenticator) setNewAppSecret(ctx context.Context) error {
	err := k.userAdminLogin(ctx)
	if err != nil {
		return fmt.Errorf("%s. error: %w", ErrAdminLoginFailed, err)
	}
	client, err := k.getAppAdminClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get admin client. error: %w", err)
	}
	if client == nil {
		return fmt.Errorf("error: %s", "admin client not found")
	}
	err = k.setAppAdminSecret(ctx, *client.ID)
	if err != nil {
		return err
	}
	return nil
}

func (k *KeycloakAuthenticator) setAppAdminSecret(ctx context.Context, IdOfClient string) error {
	secret, err := k.KcClient.GetClientSecret(ctx, k.userAuth.userToken.AccessToken, MasterRealmName, IdOfClient)
	if err != nil || secret.Value == nil {
		return fmt.Errorf("%s. error: %w", "failed to get admin application secret", err)
	}
	k.appAuth.appSecret = secret
	return nil
}

func (k *KeycloakAuthenticator) userAdminLogin(ctx context.Context) error {
	token, err := k.KcClient.LoginAdmin(ctx, k.userAuth.username, k.userAuth.password, MasterRealmName)
	if err != nil {
		return fmt.Errorf("failed to login admin user. error: %w", err)
	}
	k.userAuth.setToken(token)
	return nil
}

func (k *KeycloakAuthenticator) appAdminLogin(ctx context.Context) error {
	token, err := k.KcClient.LoginClient(ctx, ServiceClientID, *k.appAuth.appSecret.Value, MasterRealmName)
	if err != nil {
		return fmt.Errorf("failed to login admin app. error: %w", err)
	}
	k.appAuth.setToken(token)
	return nil
}

func (k *KeycloakAuthenticator) createNewAppAdminClient(ctx context.Context) (string, error) {
	client := gocloak.Client{
		ClientID:                  gocloak.StringP(ServiceClientID),
		Name:                      gocloak.StringP(ServiceClientID),
		Enabled:                   gocloak.BoolP(true),
		ClientAuthenticatorType:   gocloak.StringP(utils.AuthTypeClientSecret),
		ServiceAccountsEnabled:    gocloak.BoolP(true),
		StandardFlowEnabled:       gocloak.BoolP(false),
		DirectAccessGrantsEnabled: gocloak.BoolP(false),
		PublicClient:              gocloak.BoolP(false),
		Protocol:                  gocloak.StringP(utils.AuthProtocolOidc),
		Attributes: &map[string]string{
			utils.LightWeightToken:       "true",
			utils.TokenLifespanAttribute: fmt.Sprintf("%d", int64(DefaultServiceAccountTokenLifespan.Seconds())),
		},
	}
	adminToken := k.userAuth.userToken.AccessToken
	clientID, err := k.KcClient.CreateClient(ctx, adminToken, MasterRealmName, client)
	if err != nil {
		return "", fmt.Errorf("failed to create admin app client. error: %w", err)
	}

	serviceAccountUser, err := k.KcClient.GetClientServiceAccount(ctx, adminToken, MasterRealmName, clientID)
	if err != nil {
		return "", fmt.Errorf("failed to get service account user. error: %w", err)
	}

	if err := k.assignAdminRolesToAdminApp(ctx, adminToken, serviceAccountUser); err != nil {
		return "", fmt.Errorf("failed to get assign admin roles for admin app. error: %w", err)
	}
	return clientID, nil
}

func (k *KeycloakAuthenticator) getAppAdminClient(ctx context.Context) (*gocloak.Client, error) {
	existingClients, err := k.KcClient.GetClients(ctx, k.userAuth.userToken.AccessToken, MasterRealmName, gocloak.GetClientsParams{
		ClientID: ptr.ToString(ServiceClientID),
	})
	if err != nil || len(existingClients) == 0 {
		return nil, fmt.Errorf("failed to check existing clients. error: %w", err)
	}
	return existingClients[0], nil
}

func (k *KeycloakAuthenticator) assignAdminRolesToAdminApp(ctx context.Context, adminToken string, serviceAccountUser *gocloak.User) error {
	if serviceAccountUser.ID == nil {
		return fmt.Errorf("service account user ID is nil")
	}

	var rolesToAssign []gocloak.Role

	if adminRole, err := k.KcClient.GetRealmRole(ctx, adminToken, MasterRealmName, AdminRole); err == nil && adminRole != nil {
		rolesToAssign = append(rolesToAssign, *adminRole)
	}

	if defaultRole, err := k.KcClient.GetRealmRole(ctx, adminToken, MasterRealmName, DefaultMasterRole); err == nil && defaultRole != nil {
		rolesToAssign = append(rolesToAssign, *defaultRole)
	}

	if len(rolesToAssign) > 0 {
		if err := k.KcClient.AddRealmRoleToUser(ctx, adminToken, MasterRealmName, *serviceAccountUser.ID, rolesToAssign); err != nil {
			return fmt.Errorf("failed to assign admin roles to admin app. error: %w", err)
		}
	}

	return nil
}

func (ua *UserAuthentication) getAccessToken() *string {
	if ua.userToken == nil || ua.userToken.AccessToken == "" {
		return nil
	}
	return &ua.userToken.AccessToken
}

func (ua *UserAuthentication) isTokenValid() bool {
	if ua.userToken == nil || ua.userToken.AccessToken == "" {
		return false
	}
	return time.Until(ua.userTokenExpiresAt) > 10*time.Second
}

func (ua *UserAuthentication) setToken(token *gocloak.JWT) {
	ua.userToken = token
	ua.userTokenExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
}

func (aa *ServiceAccountAuthentication) getAccessToken() *string {
	if aa.appToken == nil || aa.appToken.AccessToken == "" {
		return nil
	}
	return &aa.appToken.AccessToken
}

func (aa *ServiceAccountAuthentication) isTokenValid() bool {
	if aa.appToken == nil || aa.appToken.AccessToken == "" {
		return false
	}
	return time.Until(aa.appTokenExpiresAt) > 20*time.Second
}

func (aa *ServiceAccountAuthentication) setToken(token *gocloak.JWT) {
	aa.appToken = token
	aa.appTokenExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
}

func (aa *ServiceAccountAuthentication) getClientSecret() *string {
	if aa.appSecret == nil || aa.appSecret.Value == nil {
		return nil
	}
	return aa.appSecret.Value
}
