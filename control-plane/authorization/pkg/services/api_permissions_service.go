/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/database/models"
	"github.com/google/uuid"
)

// PermissionsApiService is a service that implements the logic for the PermissionsApiServicer
// This service should implement the business logic for every endpoint for the Permissions API.
// Include any external packages or services that will be required by this service.
type PermissionsApiService struct{}

func NewPermissionsApiService() api.PermissionsAPIServicer {
	return &PermissionsApiService{}
}

func (p PermissionsApiService) GetPermissions(ctx context.Context) (api.ImplResponse, error) {
	tenantIdUint, subjectId, groups, err := parseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	permissionSets, err := dao.GetPermissionSets(ctx, tenantIdUint, subjectId, groups)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	resultPermissions := RolePermissionsDaoToApi(permissionSets)

	return api.Response(http.StatusOK, resultPermissions), nil
}

// EvaluatePermission - Evaluate user permission to the requested resource. The returned permitted scopes are the scopes that the action on the resource is permitted, with OR between the multiple scopes.
func (s PermissionsApiService) EvaluatePermission(ctx context.Context, backendEvaluationRequest api.BackendEvaluationRequest) (api.ImplResponse, error) {
	tenantId, subjectId, groups, err := parseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	permittedScopes, err := dao.EvaluatePermission(ctx, tenantId, subjectId, groups, backendEvaluationRequest, false)
	if err != nil {
		return NewErrorResponse(err), nil
	}
	if permittedScopes == nil {
		return api.Response(http.StatusNoContent, nil), nil
	}

	return api.Response(http.StatusOK, permittedScopes), nil
}

func (p PermissionsApiService) GetPermittedScopes(ctx context.Context, request api.GetPermittedScopesRequest) (api.ImplResponse, error) {
	tenantIdUint, subjectId, groups, err := parseAuthData(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	var requestedActions []api.Action
	if request.Action != nil {
		requestedActions = []api.Action{api.Action(*request.Action)}
	} else {
		requestedActions = []api.Action{api.ACTION_CREATE, api.ACTION_READ, api.ACTION_UPDATE, api.ACTION_DELETE}
	}

	result := &api.PermittedScopesActions{
		Create: emptyPermittedScopes(),
		Read:   emptyPermittedScopes(),
		Update: emptyPermittedScopes(),
		Delete: emptyPermittedScopes(),
	}

	for _, action := range requestedActions {
		err = getPermittedScopesForAction(ctx, result, request.ResourceType, action, tenantIdUint, subjectId, groups)
		if err != nil {
			return NewErrorResponse(err), nil
		}
	}

	return api.Response(http.StatusOK, result), nil
}

func (p PermissionsApiService) GetPermissionSets(ctx context.Context) (api.ImplResponse, error) {
	permissionSets, apiErr := dao.DbGetPermissionSets(ctx, false)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	apiPermissionSets := permissionSetsToApi(permissionSets)

	return api.Response(http.StatusOK, api.GetPermissionSets200Response{PermissionSets: apiPermissionSets}), nil
}

func (p PermissionsApiService) GetPermissionSet(ctx context.Context, permissionSetId string) (api.ImplResponse, error) {
	psUuid, err := uuid.Parse(permissionSetId)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusInternalServerError, "Failed to get permission set", nil)), nil
	}
	permissionSet, apiErr := dao.DbGetPermissionSetById(ctx, psUuid)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	apiPermissionSet := permissionSetToApi(*permissionSet)

	return api.Response(http.StatusOK, apiPermissionSet), nil
}

func getPermittedScopesForAction(ctx context.Context, result *api.PermittedScopesActions,
	resourceType api.ResourceType, action api.Action,
	tenantId uint, subjectId string, groups []string,
) *api.Error {
	permittedScopes, err := dao.EvaluatePermission(ctx, tenantId, subjectId, groups, api.BackendEvaluationRequest{
		ResourceType:  &resourceType,
		RequestAction: &action,
	}, true)
	if err != nil {
		if err.Code != http.StatusForbidden {
			return err
		}
		return nil
	}

	switch action {
	case api.ACTION_CREATE:
		result.Create = *permittedScopes
	case api.ACTION_READ:
		result.Read = *permittedScopes
	case api.ACTION_UPDATE:
		result.Update = *permittedScopes
	case api.ACTION_DELETE:
		result.Delete = *permittedScopes
	}

	return nil
}

func permissionSetsToApi(permissionSets *[]models.PermissionSet) []api.PermissionSet {
	apiPermissionSets := []api.PermissionSet{}
	for _, ps := range *permissionSets {
		apiPermissionSets = append(apiPermissionSets, permissionSetToApi(ps))
	}
	return apiPermissionSets
}

func permissionSetToApi(permissionSet models.PermissionSet) api.PermissionSet {
	apiPermissions := []api.RolePermission{}
	for _, permission := range permissionSet.Permissions {
		actions := []api.PermissionAction{}
		if permission.Read {
			actions = append(actions, api.PERMISSIONACTION_READ)
		}
		if permission.Create {
			actions = append(actions, api.PERMISSIONACTION_CREATE)
		}
		if permission.Update {
			actions = append(actions, api.PERMISSIONACTION_UPDATE)
		}
		if permission.Delete {
			actions = append(actions, api.PERMISSIONACTION_DELETE)
		}
		apiPermissions = append(apiPermissions, api.RolePermission{
			ResourceType: permission.ResourceType,
			Actions:      actions,
		})
	}
	return api.PermissionSet{
		Id:          permissionSet.ID.String(),
		Name:        permissionSet.Name,
		Description: permissionSet.Description,
		Permissions: apiPermissions,
	}
}

func emptyPermittedScopes() api.PermittedScopes {
	return api.PermittedScopes{
		Clusters:    []string{},
		Departments: []string{},
		Projects:    []string{},
	}
}
