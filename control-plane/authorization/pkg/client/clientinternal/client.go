/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/structs"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
	"k8s.io/utils/ptr"
)

const GroupsCacheHashField = "groups"

type AuthorizationClient struct {
	authorization_api.InternalAPI
	authorization_api.AccessRulesAPI
	authorization_api.RolesAPI
	authorization_api.PermissionsAPI
	CachedClient *redis.Client
}

func (a *AuthorizationClient) GetPermittedScopes(ctx context.Context, resourceType authorization_api.ResourceType, action *authorization_api.Action) (*authorization_api.PermittedScopesActions, int, error) {
	var actionStr authorization_api.NullableString
	if action != nil {
		actionStr = *authorization_api.NewNullableString(ptr.To(string(*action)))
	}

	resData, resp, err := a.PermissionsAPI.GetPermittedScopes(ctx).GetPermittedScopesRequest(authorization_api.GetPermittedScopesRequest{
		ResourceType: resourceType,
		Action:       actionStr,
	}).Execute()

	if err != nil || (resp != nil && resp.StatusCode != http.StatusOK) {
		statusCode := http.StatusServiceUnavailable
		if resp != nil {
			statusCode = resp.StatusCode
		}
		return nil, statusCode, fmt.Errorf("failed to get permitted scopes from authz service for resource %s. Response: %+v, Error: %w", resourceType, resp, err)
	}
	return resData, http.StatusOK, nil
}

func (a *AuthorizationClient) AuthorizeMiddleware(next http.Handler, resourceType string, requestAction string, _ bool, resourceIdParam string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		params := authorization_api.NewBackendEvaluationRequest()
		if resourceType != "" {
			params.SetResourceType(authorization_api.ResourceType(resourceType))

			if resourceIdParam != "" {
				resourceId := chi.URLParam(r, resourceIdParam)

				params.SetResourceId(resourceId)
			}
		}

		if requestAction != "" {
			params.SetRequestAction(authorization_api.Action(strings.ToLower(requestAction)))
		}

		permittedScopes, response, err := a.EvaluatePermission(ctx).BackendEvaluationRequest(*params).Execute()
		if response != nil {
			if response.StatusCode > 299 {
				defer response.Body.Close()

				body, err := io.ReadAll(response.Body)
				if err != nil {
					http.Error(w, response.Status, response.StatusCode)
				}

				http.Error(w, string(body), response.StatusCode)
				return
			}
			if response.StatusCode == http.StatusNoContent {
				next.ServeHTTP(w, r)
				return
			}
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}

		ctx = context.WithValue(ctx, structs.PermittedScopesKey, permittedScopes)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func (a *AuthorizationClient) getGroupsForTenantCached(ctx context.Context, tenantId string) ([]string, error) {
	groupsBytes, err := a.CachedClient.HGet(ctx, fmt.Sprintf("tenant:%s", tenantId), GroupsCacheHashField).Bytes()
	if err == nil {
		var groups []string
		err = json.Unmarshal(groupsBytes, &groups)

		if err == nil {
			return groups, nil
		}
	}

	return nil, err
}

func (a *AuthorizationClient) GetGroupsForTenant(ctx context.Context, tenantId string) ([]string, error) {
	if a.CachedClient != nil {
		groups, err := a.getGroupsForTenantCached(ctx, tenantId)
		if err == nil {
			return groups, nil
		}
	}

	ctxForRequest := context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	resData, resp, err := a.InternalAPI.GetGroups(ctxForRequest).Execute()

	if err != nil || (resp != nil && resp.StatusCode != http.StatusOK) {
		return nil, fmt.Errorf("failed to get groups from authz for tenant id %s. Response: %+v, Error: %w", tenantId, resp, err)
	}
	return resData.GetNames(), nil
}

func (a *AuthorizationClient) SetServiceAccountPermissions(ctx context.Context, tenantId, clientId string, scope authorization_api.InternalCreateAccessRuleRequestScope, role authorization_api.AllowedInternalCreateRoles) error {
	ctxForRequest := context.WithValue(ctx, runaictx.CfTenantId, tenantId)
	request := a.CreateInternalAccessRule(ctxForRequest)
	reqBody := authorization_api.NewInternalCreateAccessRuleRequest(clientId, authorization_api.SUBJECTTYPE_SERVICE_ACCOUNT, role)
	reqBody.SetScope(scope)

	_, resp, err := request.InternalCreateAccessRuleRequest(*reqBody).Execute()
	if err != nil && (resp == nil || resp.StatusCode != http.StatusConflict) {
		return fmt.Errorf("failed to set runai installer permissions for tenant id %s. Response: %+v, Error: %w", tenantId, resp, err)
	}
	return nil
}

// Deprecated: use SetServiceAccountPermissions. Kept until consumers move to the new method; it also creates a
// service-account rule.
func (a *AuthorizationClient) SetAppPermissions(ctx context.Context, tenantId, clientId string, scope authorization_api.InternalCreateAccessRuleRequestScope, role authorization_api.AllowedInternalCreateRoles) error {
	return a.SetServiceAccountPermissions(ctx, tenantId, clientId, scope, role)
}

func (a *AuthorizationClient) CalculatePermittedScopes(ctx context.Context, request authorization_api.GetPermittedScopesRequest) (*authorization_api.PermittedScopesActions, error) {
	permittedScopes, response, err := a.PermissionsAPI.GetPermittedScopes(ctx).GetPermittedScopesRequest(request).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get permitted scopes. Response: %+v, Error: %w", response, err)
	}
	return permittedScopes, nil
}

func (a *AuthorizationClient) EvaluatePermission(ctx context.Context) authorization_api.ApiEvaluatePermissionRequest {
	return a.PermissionsAPI.EvaluatePermission(ctx)
}

func (a *AuthorizationClient) EvaluatePermissionExecute(r authorization_api.ApiEvaluatePermissionRequest) (*authorization_api.PermittedScopes, *http.Response, error) {
	return a.PermissionsAPI.EvaluatePermissionExecute(r)
}

func (a *AuthorizationClient) GetSyncAccessRules(ctx context.Context, limit int32, offset int32, lastUpdated string, clusterId string) (*authorization_api.GetAccessRulesSync200Response, *http.Response, error) {
	request := a.AccessRulesAPI.GetAccessRulesSync(ctx).Limit(limit).Offset(offset).LastUpdated(lastUpdated).ClusterId(clusterId)
	return a.AccessRulesAPI.GetAccessRulesSyncExecute(request)
}

func (a *AuthorizationClient) UpdateStatusAccessRule(ctx context.Context, accessRuleId int32, clusterId string, accessRulesStatusRequest authorization_api.AccessRuleStatusRequest) (*http.Response, error) {
	request := a.AccessRulesAPI.UpdateAccessRuleStatus(
		ctx,
		accessRuleId,
		clusterId)
	request = request.AccessRuleStatusRequest(accessRulesStatusRequest)

	return request.Execute()
}

func (a *AuthorizationClient) AccessRulesSync(ctx context.Context, clusterId string, accessRuleForSync []authorization_api.AccessRuleForSync) (*http.Response, error) {
	request := a.AccessRulesAPI.SyncAccessRules(ctx, clusterId)
	request = request.AccessRuleForSync(accessRuleForSync)

	return request.Execute()
}

func (a *AuthorizationClient) ValidatePrivilegeEscalation(ctx context.Context, subjectId *string, groupIds *[]string) error {
	request := a.PrivilegeEscalationEvaluate(ctx)
	if subjectId != nil {
		request = request.SubjectId(*subjectId)
	}
	if groupIds != nil {
		request = request.GroupIds(*groupIds)
	}
	response, err := a.PrivilegeEscalationEvaluateExecute(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed validating Priveledge escalation: %s, status code: %v", response.Status, response.StatusCode)
	}
	return nil
}
