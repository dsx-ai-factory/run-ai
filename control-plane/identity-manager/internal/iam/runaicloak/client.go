/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaicloak

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-resty/resty/v2"

	iamError "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/errors"
)

type RunAICloakClient interface {
	ExchangeExternalIdpToken(ctx context.Context, adminToken, realm, token, clientId, clientSecret, idpAlias string) (*Tokens, error)
	GetClientScopes(ctx context.Context, realm, accessToken string) ([]map[string]interface{}, error)
	DeleteDefaultClientScopeById(ctx context.Context, realm, accessToken, scopeID string) error
	AddOptionalClientScopeById(ctx context.Context, realm, accessToken, scopeID string) error
	DeleteAuthFlowExecutionConfig(ctx context.Context, realm, accessToken, configID string) error
	UpdateAuthFlowExecutionConfig(ctx context.Context, realm, accessToken, executionID, alias string) error
	UpdateIDPExecutionConfig(ctx context.Context, realm, accessToken, executionId, alias string) error
	ToggleManagementPermissionsForIdp(ctx context.Context, realm, accessToken, idpAlias string, enabled bool) error
	DoesIdpHaveExchangePermissions(ctx context.Context, realm, accessToken, idpAlias string) bool
	GetPolicyNameByScope(ctx context.Context, realm, accessToken, mngIdOfClient, scopePermissionId string) (string, error)
	GetPolicyIdForTokenExchange(ctx context.Context, realm, accessToken, mngIdOfClient string) (string, error)
	GetPermissionsResourceIds(ctx context.Context, realm, accessToken, idpAlias string) (string, string, error)
	CreatePolicyExchangeRequest(ctx context.Context, realm, accessToken, targetClientID, sourceIdOfClient, exchangePolicyName string) error
	SetExchangePolicy(ctx context.Context, realm, accessToken, mngIdOfClient, permissionresourceID,
		exchangeScopePermissionId, exchangePolicyScopeId, exchangePolicyId, exchangePolicyName string) error
	ToggleUserAttribute(ctx context.Context, realm, accessToken string, attributes map[string]interface{}, attributeKey string, enabled bool) error
	GetRealmUserAttributes(ctx context.Context, realm, accessToken string) (map[string]interface{}, error)
	FetchKeycloakConfiguration(ctx context.Context, realm string) (map[string]interface{}, error)
	FetchOAuthAuthorizationServerMetadata(ctx context.Context, realm string) (map[string]interface{}, error)
}

const (
	ExchangePolicyName = "token-exchange"
	ScopeName          = "token-exchange"
)

type Client struct {
	restyClient *resty.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		restyClient: resty.New().SetBaseURL(baseURL),
	}
}

func (c *Client) ExchangeExternalIdpToken(ctx context.Context, adminToken, realm, token, clientId, clientSecret, idpAlias string) (*Tokens, error) {
	resp, err := c.restyClient.R().SetAuthToken(adminToken).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(map[string]string{
			"grant_type":     "urn:ietf:params:oauth:grant-type:token-exchange",
			"client_id":      clientId,
			"client_secret":  clientSecret,
			"subject_token":  token,
			"subject_issuer": idpAlias,
		}).
		Post(fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", c.restyClient.BaseURL, realm))
	if err != nil {
		return nil, fmt.Errorf("error on token exchange. err: %s %w", err.Error(), iamError.ErrInternal)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("error on token exchange. res statusCode: %d. err:%v body: %s. %w", resp.StatusCode(), resp.Error(), string(resp.Body()), iamError.ErrInternal)
	}

	var tokenResp tokenExchangeResponse
	if err := json.Unmarshal(resp.Body(), &tokenResp); err != nil {
		return nil, fmt.Errorf("error on token exchange. err: %s %w", err.Error(), iamError.ErrInternal)
	}

	return &Tokens{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
	}, nil
}

func (c *Client) GetClientScopes(ctx context.Context, realm, accessToken string) ([]map[string]interface{}, error) {
	resp, err := c.restyClient.R().
		SetAuthToken(accessToken).
		Get(fmt.Sprintf("admin/realms/%s/client-scopes", realm))
	if err != nil {
		return nil, fmt.Errorf("failed to get client scopes: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get client scopes: %s, %s", resp.Status(), string(resp.Body()))
	}

	var clientScopes []map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &clientScopes); err != nil {
		return nil, fmt.Errorf("failed to decode client scopes response: %w", err)
	}

	return clientScopes, nil
}

func (c *Client) DeleteDefaultClientScopeById(ctx context.Context, realm, accessToken, scopeID string) error {
	resp, err := c.restyClient.R().
		SetAuthToken(accessToken).
		Delete(fmt.Sprintf("admin/realms/%s/default-default-client-scopes/%s", realm, scopeID))
	if err != nil {
		return fmt.Errorf("failed to delete client scope: %w", err)
	}

	if resp.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("failed to delete client scope: %s, %s", resp.Status(), string(resp.Body()))
	}
	return nil
}

func (c *Client) AddOptionalClientScopeById(ctx context.Context, realm, accessToken, scopeID string) error {
	resp, err := c.restyClient.R().SetAuthToken(accessToken).
		Put(fmt.Sprintf("admin/realms/%s/default-optional-client-scopes/%s", realm, scopeID))
	if err != nil {
		return fmt.Errorf("failed to add optional client scope: %w", err)
	}

	if resp.StatusCode() != http.StatusNoContent {
		return fmt.Errorf("failed to add optional client scope: %s, %s", resp.Status(), string(resp.Body()))
	}
	return nil
}

func (c *Client) DeleteAuthFlowExecutionConfig(ctx context.Context, realm, accessToken, configID string) error {
	url := fmt.Sprintf("%s/admin/realms/%s/authentication/config/%s", c.restyClient.BaseURL, realm, configID)
	response, err := c.restyClient.R().SetAuthToken(accessToken).Delete(url)
	if err != nil {
		return fmt.Errorf("failed to delete auth flow execution config: %w", err)
	}
	if response.IsError() {
		return fmt.Errorf("failed to delete auth flow execution config: %v", response)
	}
	return nil
}

func (c *Client) UpdateAuthFlowExecutionConfig(ctx context.Context, realm, accessToken, executionID, alias string) error {
	url := fmt.Sprintf("%s/admin/realms/%s/authentication/executions/%s/config", c.restyClient.BaseURL, realm, executionID)
	response, err := c.restyClient.R().SetAuthToken(accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(executionConfigRequest{
			Alias:  alias,
			Config: executionConfig{DefaultProvider: alias},
		}).
		Post(url)
	if err != nil {
		return err
	}
	if response.IsError() {
		return fmt.Errorf("failed to update auth flow execution config: %v", response)
	}
	return nil
}

func (c *Client) UpdateIDPExecutionConfig(ctx context.Context, realm, accessToken, executionId, alias string) error {
	const errMessage = "could not update IDP execution config"
	response, err := c.restyClient.R().SetAuthToken(accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(executionConfigRequest{
			Alias:  alias,
			Config: executionConfig{DefaultProvider: alias},
		}).
		Post(fmt.Sprintf("%s/auth/admin/realms/%s/authentication/executions/%s/config", c.restyClient.BaseURL, realm, executionId))
	if err != nil {
		return err
	}
	if response.IsError() {
		return fmt.Errorf("%s - %v", errMessage, response)
	}
	return nil
}

func (c *Client) ToggleManagementPermissionsForIdp(ctx context.Context, realm, accessToken, idpAlias string, enabled bool) error {
	type request struct {
		Enabled bool `json:"enabled,omitempty"`
	}

	response, err := c.restyClient.R().SetAuthToken(accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(request{enabled}).
		Put(fmt.Sprintf("%s/admin/realms/%s/identity-provider/instances/%s/management/permissions", c.restyClient.BaseURL, realm, idpAlias))
	if err != nil {
		return fmt.Errorf("failed to enable permissions for idp: %s. err: %s. %w", idpAlias, err.Error(), iamError.ErrInternal)
	}
	if response.IsError() {
		return fmt.Errorf("failed to enable permissions for idp: %s. res: %v. %w", idpAlias, response, iamError.ErrInternal)
	}
	return nil
}

func (c *Client) DoesIdpHaveExchangePermissions(ctx context.Context, realm, accessToken, idpAlias string) bool {
	type responseObj struct {
		Enabled          bool              `json:"enabled,omitempty"`
		Resource         string            `json:"resource,omitempty"`
		ScopePermissions map[string]string `json:"scopePermissions,omitempty"`
	}
	response, err := c.restyClient.R().SetAuthToken(accessToken).
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("%s/admin/realms/%s/identity-provider/instances/%s/management/permissions", c.restyClient.BaseURL, realm, idpAlias))
	if err != nil {
		return false
	}

	if response.IsError() {
		return false
	}

	var resp responseObj
	unmarshalErr := json.Unmarshal(response.Body(), &resp)
	if unmarshalErr != nil || !resp.Enabled {
		return false
	}

	for scope := range resp.ScopePermissions {
		if scope == ScopeName {
			return true
		}
	}
	return false
}

func (c *Client) GetPolicyNameByScope(ctx context.Context, realm, accessToken, mngIdOfClient, scopePermissionId string) (string, error) {
	type responseObj struct {
		Id               string      `json:"id,omitempty"`
		Name             string      `json:"name,omitempty"`
		Type             string      `json:"type,omitempty"`
		Logic            string      `json:"logic,omitempty"`
		DecisionStrategy string      `json:"decisionStrategy,omitempty"`
		Config           interface{} `json:"config,omitempty"`
	}

	response, err := c.restyClient.R().SetAuthToken(accessToken).
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("%s/admin/realms/%s/clients/%s/authz/resource-server/policy/%s", c.restyClient.BaseURL, realm, mngIdOfClient, scopePermissionId))
	if err != nil {
		return "", fmt.Errorf("failed to get policies for scope id: %s, in realm: %s. err: %s. %w", scopePermissionId, realm, err.Error(), iamError.ErrInternal)
	}
	if response.IsError() {
		return "", fmt.Errorf("failed to get policies for scope id: %s, in realm: %s. res: %v. %w", scopePermissionId, realm, response, iamError.ErrInternal)
	}

	var res responseObj
	unmarshalErr := json.Unmarshal(response.Body(), &res)
	if unmarshalErr != nil {
		return "", fmt.Errorf("getPolicyNameByScope unmarshalErr. %x. %w", response.Body(), iamError.ErrInternal)
	}
	return res.Name, nil
}

func (c *Client) GetPolicyIdForTokenExchange(ctx context.Context, realm, accessToken, mngIdOfClient string) (string, error) {
	response, err := c.restyClient.R().SetAuthToken(accessToken).
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("%s/admin/realms/%s/clients/%s/authz/resource-server/scope?first=0&max=11&deep=false", c.restyClient.BaseURL, realm, mngIdOfClient))
	if err != nil {
		return "", fmt.Errorf("failed to get policies for id of client: %s:, in realm: %s. err: %s. %w", mngIdOfClient, realm, err.Error(), iamError.ErrInternal)
	}
	if response.IsError() {
		return "", fmt.Errorf("failed to get policies for id of client: %s:, in realm: %s. res: %v %w", mngIdOfClient, realm, response, iamError.ErrInternal)
	}
	type Policy struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	var policies []Policy
	unmarshalErr := json.Unmarshal(response.Body(), &policies)
	if unmarshalErr != nil {
		return "", fmt.Errorf("getPolicyIdForTokenExchange unmarshalErr. %x. %w", response.Body(), iamError.ErrInternal)
	}

	for _, policy := range policies {
		if policy.Name == ExchangePolicyName {
			return policy.ID, nil
		}
	}
	return "", fmt.Errorf("missing policy for token exchange scope %w", iamError.ErrInternal)
}

func (c *Client) GetPermissionsResourceIds(ctx context.Context, realm, accessToken, idpAlias string) (string, string, error) {
	type responseObj struct {
		Enabled          bool              `json:"enabled,omitempty"`
		Resource         string            `json:"resource,omitempty"`
		ScopePermissions map[string]string `json:"scopePermissions,omitempty"`
	}
	response, err := c.restyClient.R().SetAuthToken(accessToken).
		SetHeader("Content-Type", "application/json").
		Get(fmt.Sprintf("%s/admin/realms/%s/identity-provider/instances/%s/management/permissions", c.restyClient.BaseURL, realm, idpAlias))
	if err != nil {
		return "", "", fmt.Errorf("failed to get idp permissions for : %s, in realm: %s. err: %s. %w", idpAlias, realm, err.Error(), iamError.ErrInternal)
	}

	if response.IsError() {
		return "", "", fmt.Errorf("failed to get idp permissions for : %s, in realm: %s. res: %v. %w", idpAlias, realm, response, iamError.ErrInternal)
	}

	var resp responseObj
	unmarshalErr := json.Unmarshal(response.Body(), &resp)
	if unmarshalErr != nil {
		return "", "", fmt.Errorf("getPermissionsResourceIds unmarshalErr. %x. %w", response.Body(), iamError.ErrInternal)
	}

	for scope, id := range resp.ScopePermissions {
		if scope == ScopeName {
			return resp.Resource, id, nil
		}
	}
	return "", "", fmt.Errorf("missing idp token-exchange scope permissions for idp. %w", iamError.ErrInternal)
}

func (c *Client) CreatePolicyExchangeRequest(ctx context.Context, realm, accessToken, targetClientID, sourceIdOfClient, exchangePolicyName string) error {
	type request struct {
		Name        string   `json:"name,omitempty"`
		Description string   `json:"description,omitempty"`
		Clients     []string `json:"clients,omitempty"`
		Logic       string   `json:"logic,omitempty"`
	}
	response, err := c.restyClient.R().SetAuthToken(accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(request{
			Name:        exchangePolicyName,
			Description: "",
			Clients:     []string{targetClientID},
			Logic:       "POSITIVE",
		}).
		Post(fmt.Sprintf("%s/admin/realms/%s/clients/%s/authz/resource-server/policy/client", c.restyClient.BaseURL, realm, sourceIdOfClient))
	if err != nil {
		return fmt.Errorf("failed to create policy for exchange client: %s. res: %v. %w", targetClientID, response, iamError.ErrInternal)
	}
	if response.IsError() {
		if response.StatusCode() == http.StatusConflict || strings.Contains(strings.ToLower(response.String()), "already exists") {
			return nil
		}
		return fmt.Errorf("failed to create policy for exchange client: %s. res: %v. %w", targetClientID, response, iamError.ErrInternal)
	}
	return nil
}

func (c *Client) SetExchangePolicy(ctx context.Context, realm, accessToken, mngIdOfClient, permissionresourceID,
	exchangeScopePermissionId, exchangePolicyScopeId, exchangePolicyId, exchangePolicyName string,
) error {
	type request struct {
		Id               string   `json:"id,omitempty"`
		Name             string   `json:"name,omitempty"`
		Type             string   `json:"type,omitempty"`
		Logic            string   `json:"logic,omitempty"`
		DecisionStrategy string   `json:"decisionStrategy,omitempty"`
		Resources        []string `json:"resources,omitempty"`
		Policies         []string `json:"policies,omitempty"`
		Scopes           []string `json:"scopes,omitempty"`
		Description      string   `json:"description,omitempty"`
	}
	response, err := c.restyClient.R().SetAuthToken(accessToken).
		SetHeader("Content-Type", "application/json").
		SetBody(request{
			Id:               exchangeScopePermissionId,
			Name:             exchangePolicyName,
			Type:             "scope",
			Logic:            "POSITIVE",
			DecisionStrategy: "UNANIMOUS",
			Resources:        []string{permissionresourceID},
			Policies:         []string{exchangePolicyId},
			Scopes:           []string{exchangePolicyScopeId},
			Description:      "",
		}).Put(fmt.Sprintf("%s/admin/realms/%s/clients/%s/authz/resource-server/permission/scope/%s", c.restyClient.BaseURL, realm, mngIdOfClient, exchangeScopePermissionId))
	if err != nil {
		return fmt.Errorf("failed to set exchange policy for client: %s. err: %s. %w", mngIdOfClient, err.Error(), iamError.ErrInternal)
	}
	if response.IsError() {
		return fmt.Errorf("failed to set exchange policy for client: %s. res: %v. %w", mngIdOfClient, response, iamError.ErrInternal)
	}
	return nil
}

func (c *Client) ToggleUserAttribute(ctx context.Context, realm, accessToken string, attributes map[string]interface{}, attributeKey string, enabled bool) error {
	enabledValues := map[bool]string{
		true: "ENABLED",
	}
	if !enabled {
		delete(attributes, attributeKey)
	} else {
		attributes[attributeKey] = enabledValues[enabled]
	}
	url := fmt.Sprintf("%s/admin/realms/%s/users/profile", c.restyClient.BaseURL, realm)
	putResponse, err := c.restyClient.R().SetAuthToken(accessToken).
		SetBody(attributes).Put(url)
	if err != nil {
		return fmt.Errorf("failed to update user attribute %s. err: %s", attributeKey, err.Error())
	}
	if putResponse.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to update user attribute %s. status code: %d", attributeKey, putResponse.StatusCode())
	}
	return nil
}

func (c *Client) GetRealmUserAttributes(ctx context.Context, realm, accessToken string) (map[string]interface{}, error) {
	var attributesJson map[string]interface{}
	url := fmt.Sprintf("%s/admin/realms/%s/users/profile", c.restyClient.BaseURL, realm)
	AttrResponse, err := c.restyClient.R().SetAuthToken(accessToken).Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get users attributes for realm %s. err: %s", realm, err.Error())
	}
	if AttrResponse.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to get users attributes for realm %s. status code: %d", realm, AttrResponse.StatusCode())
	}
	err = json.Unmarshal(AttrResponse.Body(), &attributesJson)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal realm users attributes. err: %s", err.Error())
	}
	return attributesJson, nil
}

func (c *Client) fetchWellKnown(ctx context.Context, url string) (map[string]interface{}, error) {
	resp, err := c.restyClient.R().
		SetContext(ctx).
		Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch configuration from Keycloak: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("keycloak returned status %d: %s", resp.StatusCode(), string(resp.Body()))
	}

	var config map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &config); err != nil {
		return nil, fmt.Errorf("failed to decode configuration: %w", err)
	}

	return config, nil
}

func (c *Client) FetchKeycloakConfiguration(ctx context.Context, realm string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/realms/%s/.well-known/openid-configuration", c.restyClient.BaseURL, realm)
	return c.fetchWellKnown(ctx, url)
}

func (c *Client) FetchOAuthAuthorizationServerMetadata(ctx context.Context, realm string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/.well-known/oauth-authorization-server/realms/%s", c.restyClient.BaseURL, realm)
	return c.fetchWellKnown(ctx, url)
}
