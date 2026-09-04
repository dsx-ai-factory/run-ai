/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package internal

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	middlewares "github.com/dsx-ai-factory/run-ai/runai-application-packages/old-auth-token-middleware"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type IdmClient struct {
	ApiClient *openapi.APIClient
}

func (client *IdmClient) GetInstallerServiceAccountForTenantId(ctx context.Context, tenantId string) (*openapi.ServiceAccountPostResponse, error) {
	jwt, ok := ctx.Value(middlewares.ContextJwtToken).(string)
	ctx = context.WithValue(ctx, openapi.ContextAccessToken, jwt)
	req := client.ApiClient.ServiceAccountsAPI.GetInstallerServiceAccount(ctx)
	req = req.TenantId(tenantId)
	installerServiceAccount, resp, err := client.ApiClient.ServiceAccountsAPI.GetInstallerServiceAccountExecute(req)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusUnauthorized && !ok {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(
				"Unauthorized response while attempting to get installer service account for tenantId '%s', note that the %s in ctx is missing. Response: %+v",
				tenantId, middlewares.ContextJwtToken, resp,
			)
		}
		return nil, errors.Wrap(err, "failed getting installer service account")
	}
	return installerServiceAccount, nil
}

func (client *IdmClient) GetInstallerServiceAccount(ctx context.Context) (*openapi.ServiceAccountPostResponse, error) {
	return client.GetInstallerServiceAccountForTenantId(ctx, "")
}

func (client *IdmClient) MintInstallerBootstrapToken(ctx context.Context, tenantId string, clusterId string) (string, error) {
	jwt, _ := ctx.Value(middlewares.ContextJwtToken).(string)
	ctx = context.WithValue(ctx, openapi.ContextAccessToken, jwt)
	req := client.ApiClient.ServiceAccountsAPI.MintInstallerBootstrapToken(ctx)
	if tenantId != "" {
		req = req.TenantId(tenantId)
	}
	if clusterId != "" {
		req = req.ClusterId(clusterId)
	}
	resp, _, err := client.ApiClient.ServiceAccountsAPI.MintInstallerBootstrapTokenExecute(req)
	if err != nil {
		return "", fmt.Errorf("failed to mint installer bootstrap token: %w", err)
	}
	return resp.Token, nil
}

func (client *IdmClient) GetInternalClientByFilter(ctx context.Context, filter string) ([]openapi.Client, error) {
	req := client.ApiClient.ClientsAPI.GetAllClients(ctx)
	req = req.Filter(filter)
	internalClients, _, err := client.ApiClient.ClientsAPI.GetAllClientsExecute(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting clients by filter")
	}

	return internalClients, nil
}

func (client *IdmClient) AppendRedirectUriToInternalClient(ctx context.Context, internalClient openapi.Client, uri string) error {
	if !strings.HasSuffix(uri, "/") {
		uri += "/"
	}
	uri += "*"
	uriExistInList := isUriContained(internalClient.RedirectUris, uri)
	if uriExistInList {
		return nil
	}
	redirectUris := append(internalClient.RedirectUris, uri)
	updateRequest := client.ApiClient.ClientsAPI.UpdateRedirectUris(ctx, internalClient.ClientId)
	updateRedirectUris := openapi.NewClientRedirectUrisUpdateRequest(redirectUris)
	updateRequest = updateRequest.ClientRedirectUrisUpdateRequest(*updateRedirectUris)
	_, err := client.ApiClient.ClientsAPI.UpdateRedirectUrisExecute(updateRequest)
	if err != nil {
		return errors.Wrap(err, "failed updating client")
	}
	return nil
}

func (client *IdmClient) DeleteRealm(ctx context.Context, realmName string, isHardDelete bool) (*http.Response, error) {
	httpResponse, err := client.ApiClient.RealmAPI.DeleteRealm(ctx, realmName).IsHardDelete(isHardDelete).Execute()
	if err != nil || httpResponse == nil || httpResponse.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("failed to delete auth realm with name %s. isHardDelete %t. StatusCode: %v, err: %v", realmName, isHardDelete, httpResponse, err)
		return httpResponse, errors.Wrap(err, msg)
	}
	return httpResponse, nil
}

func (client *IdmClient) CreateRealm(ctx context.Context, realmName string, tenantId string) (*openapi.RealmCreationResponse, *http.Response, error) {
	req := openapi.RealmCreationRequest{
		RealmName: realmName,
		TenantId:  tenantId,
	}
	return client.ApiClient.RealmAPI.CreateRealm(ctx).RealmCreationRequest(req).Execute()
}

func (client *IdmClient) GetOrCreateRealm(ctx context.Context, realmName string, tenantId string) (*openapi.RealmCreationResponse, *http.Response, error) {
	req := openapi.RealmCreationRequest{
		RealmName: realmName,
		TenantId:  tenantId,
	}
	return client.ApiClient.RealmAPI.GetOrCreateRealm(ctx).RealmCreationRequest(req).Execute()
}

func (client *IdmClient) CreateInternalUser(ctx context.Context, req openapi.InternalUserCreationRequest) (*openapi.UserPostResponse, *http.Response, error) {
	return client.ApiClient.UsersAPI.CreateIntenralUser(ctx).InternalUserCreationRequest(req).Execute()
}

func (client *IdmClient) DeleteInternalUserByEmail(ctx context.Context, realm string, email string) (*http.Response, error) {
	return client.ApiClient.UsersAPI.DeleteInternalUserByEmail(ctx, realm, email).Execute()
}

func (client *IdmClient) CreateInternalClient(ctx context.Context, req openapi.ClientCreationRequest) (*openapi.CreateInternalClient201Response, *http.Response, error) {
	return client.ApiClient.ClientsAPI.CreateInternalClient(ctx).ClientCreationRequest(req).Execute()
}

func (client *IdmClient) DeleteClient(ctx context.Context, clientId string) (*http.Response, error) {
	return client.ApiClient.ClientsAPI.DeleteClient(ctx, clientId).Execute()
}

func (client *IdmClient) CreateClient(ctx context.Context, name string, clientType string, tenantId *string) (*string, error) {
	newClient := openapi.ClientCreationRequest{
		Type:     openapi.ClientType(clientType),
		ClientId: name,
		TenantId: tenantId,
	}
	req := client.ApiClient.ClientsAPI.CreateInternalClient(ctx)
	req = req.ClientCreationRequest(newClient)
	res, httpResponse, err := client.ApiClient.ClientsAPI.CreateInternalClientExecute(req)
	if err != nil || httpResponse == nil || httpResponse.StatusCode != http.StatusOK {
		return nil, errors.Wrap(err, fmt.Sprintf("failed to create client %s with type: %s for tenant %s. StatusCode: %v, err: %v", name, clientType, *tenantId, httpResponse, err))
	}
	return res.Id, err
}

func (client *IdmClient) ClientCredentialsLogin(ctx context.Context, clientId string, clientSecret string) (*openapi.TokenResponse, error) {
	return clientCredentialsLogin(ctx, client.ApiClient, clientId, clientSecret)
}

func (client *IdmClient) ClientCredentialsLoginV2(ctx context.Context, clientId string, clientSecret string) (*openapi.OAuth2TokenResponse, error) {
	return clientCredentialsLoginV2(ctx, client.ApiClient, clientId, clientSecret)
}

func (client *IdmClient) TokenExchange(ctx context.Context, exchangeCode string, clientId string, redirectUri string, codeVerifier string) (*openapi.TokenResponse, error) {
	return client.exchangeCodeForToken(ctx, exchangeCode, clientId, redirectUri, codeVerifier)
}

func (client *IdmClient) RefreshToken(ctx context.Context, refreshToken string, clientId string) (*openapi.TokenResponse, error) {
	return client.refreshToken(ctx, refreshToken, clientId)
}

func (client *IdmClient) UserLogin(ctx context.Context, username string, password string, clientId string) (*openapi.TokenResponse, error) {
	return client.userLogin(ctx, username, password, clientId)
}

func (client *IdmClient) CreateServiceAccountInternally(ctx context.Context, name string, realmName string) (*openapi.ServiceAccountPostResponse, error) {
	req := client.ApiClient.ServiceAccountsAPI.CreateServiceAccountInternally(ctx)
	req = req.InternalServiceAccountCreationRequest(openapi.InternalServiceAccountCreationRequest{Name: name, Realm: realmName})
	serviceAccount, _, err := client.ApiClient.ServiceAccountsAPI.CreateServiceAccountInternallyExecute(req)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("failed creating service account: %s. for realm: %s", name, realmName))
	}
	return serviceAccount, nil
}
