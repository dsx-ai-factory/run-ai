/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package authorization_client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/clientinternal"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/structs"
	client_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/redis/go-redis/v9"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
)

type IAuthorizationExternalClient interface {
	authorization_api.AccessRulesAPI
	authorization_api.RolesAPI
	EvaluatePermission(ctx context.Context) authorization_api.ApiEvaluatePermissionRequest
	EvaluatePermissionExecute(r authorization_api.ApiEvaluatePermissionRequest) (*authorization_api.PermittedScopes, *http.Response, error)
	GetSyncAccessRules(ctx context.Context, limit int32, offset int32, lastUpdated string, clusterId string) (*authorization_api.GetAccessRulesSync200Response, *http.Response, error)
	UpdateStatusAccessRule(ctx context.Context, accessRuleId int32, clusterId string, accessRulesStatusRequest authorization_api.AccessRuleStatusRequest) (*http.Response, error)
	AccessRulesSync(ctx context.Context, clusterId string, accessRuleForSync []authorization_api.AccessRuleForSync) (*http.Response, error)
}

type IAuthorizationClient interface {
	IAuthorizationExternalClient
	authorization_api.InternalAPI
	AuthorizeMiddleware(next http.Handler, resourceType string, requestAction string, hasResourceId bool, resourceIdParam string) http.HandlerFunc
	SetAppPermissions(ctx context.Context, tenantId, clientId string, scope authorization_api.InternalCreateAccessRuleRequestScope, role authorization_api.AllowedInternalCreateRoles) error
	GetGroupsForTenant(ctx context.Context, tenantId string) ([]string, error)
	CalculatePermittedScopes(ctx context.Context, request authorization_api.GetPermittedScopesRequest) (*authorization_api.PermittedScopesActions, error)
	GetPermittedScopes(ctx context.Context, ResourceType authorization_api.ResourceType, Action *authorization_api.Action) (*authorization_api.PermittedScopesActions, int, error)
	ValidatePrivilegeEscalation(ctx context.Context, subjectId *string, groupIds *[]string) error
}

func NewAuthorizationExternalClient() IAuthorizationExternalClient {
	authConfig := config.Get[*Config]()
	authConfig.IsExternal = true
	config.Set(authConfig)
	return NewAuthorizationClient()
}

func NewAuthorizationClientWithConfig(config *authorization_api.Configuration) IAuthorizationExternalClient {
	return &internal.AuthorizationClient{
		InternalAPI:    authorization_api.NewAPIClient(config).InternalAPI,
		AccessRulesAPI: authorization_api.NewAPIClient(config).AccessRulesAPI,
		RolesAPI:       authorization_api.NewAPIClient(config).RolesAPI,
		PermissionsAPI: authorization_api.NewAPIClient(config).PermissionsAPI,
	}
}

func NewAuthorizationExternalClientWithConfig(config *authorization_api.Configuration) IAuthorizationExternalClient {
	return NewAuthorizationClientWithConfig(config)
}

func NewAuthorizationClient() IAuthorizationClient {
	return NewAuthorizationCacheClient(nil)
}

func NewAuthorizationCacheClient(cachedClient *redis.Client) IAuthorizationClient {
	if mockClient != nil {
		return *mockClient
	}
	authConfig := config.Get[*Config]()
	if authConfig == nil {
		return nil
	}

	url := authConfig.Url
	if authConfig.IsExternal {
		url = client_utils.GetExternalUrl(url)
	} else {
		url = client_utils.GetInternalUrl(url)
	}

	clientConfig := authorization_api.NewConfiguration()
	clientConfig.Servers = authorization_api.ServerConfigurations{{URL: url}}
	client := authorization_api.NewAPIClient(clientConfig)
	return &internal.AuthorizationClient{
		InternalAPI:    client.InternalAPI,
		AccessRulesAPI: client.AccessRulesAPI,
		RolesAPI:       client.RolesAPI,
		PermissionsAPI: client.PermissionsAPI,
		CachedClient:   cachedClient,
	}
}

func ParsePermittedScopes(ctx context.Context) (*authorization_api.PermittedScopes, error) {
	var ok bool
	var permittedScopes *authorization_api.PermittedScopes
	if permittedScopes, ok = ctx.Value(structs.PermittedScopesKey).(*authorization_api.PermittedScopes); !ok {
		return nil, fmt.Errorf("no permitted scopes in ctx")
	}

	return permittedScopes, nil
}

func NewAccessRulesClient(baseURL string) IAuthorizationExternalClient {
	clientConfig := authorization_api.NewConfiguration()
	clientConfig.Servers = authorization_api.ServerConfigurations{{URL: client_utils.GetInternalUrl(baseURL)}}
	return NewAccessRulesClientWithConfig(clientConfig)
}

func NewAccessRulesClientWithConfig(config *authorization_api.Configuration) IAuthorizationExternalClient {
	client := authorization_api.NewAPIClient(config)
	return &internal.AuthorizationClient{
		InternalAPI:    client.InternalAPI,
		AccessRulesAPI: client.AccessRulesAPI,
		RolesAPI:       client.RolesAPI,
		PermissionsAPI: client.PermissionsAPI,
		CachedClient:   nil,
	}
}
