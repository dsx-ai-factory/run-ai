/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	org_unit "github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/clients/org-unit"
	orgUnitApi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/common"

	authorizationClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"
)

func (ss *SecretsService) CreateSecret(ctx context.Context, clusterId string, scopeType authorizationClient.ScopeType,
	scopeId string, secretName string, request api.SecretCreationRequest,
) (*api.Secret, *api.Error) {
	if scopeType == authorizationClient.SCOPETYPE_PROJECT {
		var scopeError *api.Error
		scopeId, scopeError = namespaceToProjectId(ctx, clusterId, scopeId)
		if scopeError != nil {
			return nil, scopeError
		}
	}

	tenantId, err := common.TenantIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	record := models.Secret{
		SecretFilters: models.SecretFilters{
			TenantId:  tenantId,
			ClusterId: &clusterId,
			ScopeQuery: &models.ScopeQueryParams{
				ScopeType: string(scopeType),
				ScopeId:   scopeId,
			},
			SecretName: &secretName,
			SecretType: ptr.To(string(request.Type)),
			Origin:     dao.OriginPtrToStringPtr(request.Origin),
		},
		Private: request.Private,
	}

	result, err := dao.CreateSecret(ctx, &record)
	if err != nil {
		return nil, err
	}

	// We do not handle status changes for secrets
	_, err = HandleResourceIssues(ctx, tenantId, clusterId, result.ScopeQuery.ScopeType, result.ScopeQuery.ScopeId, result.ID.String(), request.Issues, string(api.CLUSTERRESOURCETYPE_SECRET))
	if err != nil {
		return nil, err
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("created secret %+v", *result)

	return &api.Secret{
		Name:   secretName,
		Type:   request.Type,
		Origin: request.Origin,
	}, nil
}

func (ss *SecretsService) ListSecrets(ctx context.Context, scopeTypeStr string,
	scopeId string, optionalName string, optionalType string, clusterId string, optionalOrigin string, includeAncestors bool, includePrivate bool,
) (api.ImplResponse, error) {
	if scopeTypeStr == "" && clusterId == "" {
		return NewErrorResponse(api.NewValidationError("either scopeType or clusterId must be provided")), nil
	}

	if scopeTypeStr != "" && scopeId == "" {
		return NewErrorResponse(api.NewValidationError("scopeId is mandatory when scopeType is provided")), nil
	}

	if err := common.ValidateScopeType(scopeTypeStr); err != nil {
		return NewErrorResponse(err), nil
	}

	tenantId, err := common.TenantIdFromContext(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	var ancestorsResolver *org_unit.AncestorsResolver
	filters := models.SecretFilters{
		TenantId: tenantId,
	}

	if clusterId != "" {
		ancestorsResolver = org_unit.NewAncestorsResolverOfCluster(clusterId)
	} else {
		ancestorsResolver = org_unit.NewAncestorsResolver(orgUnitApi.ScopeType(scopeTypeStr), scopeId)
	}

	filters.ScopeQuery = models.NewScopeQueryParams(ancestorsResolver)

	authErr := auth.CheckReadPermission(ctx, authorizationClient.RESOURCETYPE_CREDENTIALS, ancestorsResolver)
	if authErr != nil {
		return NewErrorResponse(authErr), nil
	}

	if optionalName != "" {
		filters.SecretName = &optionalName
	}

	if optionalType != "" {
		filters.SecretType = &optionalType
	}

	filters.Origin = dao.OriginFilterToStringPtr(optionalOrigin)

	result, err := dao.ListSecrets(ctx, &filters, &models.ListOptions{
		IncludeAncestors: includeAncestors,
		IncludePrivate:   includePrivate,
	})
	if err != nil {
		return NewErrorResponse(err), nil
	}

	payload := api.SecretListResponse{
		Secrets: make([]api.SecretListInfo, 0, len(result)),
	}
	for _, secret := range result {
		origin := api.Origin(ptr.Deref(secret.Origin, ""))

		secretListInfo := api.SecretListInfo{
			ClusterId: ptr.Deref(secret.ClusterId, ""),
			Name:      ptr.Deref(secret.SecretName, ""),
			Type:      api.SecretType(ptr.Deref(secret.SecretType, "")),
			Origin:    &origin,
			Private:   secret.Private,
		}

		if secret.ScopeQuery != nil {
			secretListInfo.ScopeType = secret.ScopeQuery.ScopeType
			secretListInfo.ScopeId = secret.ScopeQuery.ScopeId
		}

		payload.Secrets = append(payload.Secrets, secretListInfo)
	}

	return api.Response(http.StatusOK, payload), nil
}

func (ss *SecretsService) SyncSecrets(ctx context.Context, clusterId string, scopeType authorizationClient.ScopeType,
	scopeId string, secrets api.Secrets,
) (*api.Secrets, *api.Error) {
	for _, secret := range secrets.Secrets {

		if secret.Name == "" {
			return nil, api.NewValidationError("mandatory secret name is not provided")
		}
		if secret.Origin == nil || *secret.Origin == "" {
			secret.Origin = &originExisting
		}
	}

	if scopeType == authorizationClient.SCOPETYPE_PROJECT {
		var scopeError *api.Error
		scopeId, scopeError = namespaceToProjectId(ctx, clusterId, scopeId)
		if scopeError != nil {
			return nil, scopeError
		}
	}

	tenantId, err := common.TenantIdFromContext(ctx)
	if err != nil {
		return nil, err
	}

	filters := models.SecretFilters{
		TenantId:  tenantId,
		ClusterId: &clusterId,
		ScopeQuery: &models.ScopeQueryParams{
			ScopeType: string(scopeType),
			ScopeId:   scopeId,
		},
	}

	secretIdMap, err := dao.ApplySecrets(ctx, &filters, secrets.Secrets)
	if err != nil {
		return nil, err
	}

	for id, secret := range secretIdMap {
		// We do not handle status changes for config-maps
		_, err = HandleResourceIssues(ctx, tenantId, clusterId, string(scopeType), scopeId, id, secret.Issues, string(api.CLUSTERRESOURCETYPE_SECRET))
		if err != nil {
			return nil, err
		}
	}

	secretIds := make([]string, 0, len(secretIdMap))
	for id := range secretIdMap {
		secretIds = append(secretIds, id)
	}
	err = dao.DeleteSecretsNotInCluster(ctx, &filters, secretIds)
	if err != nil {
		return nil, err
	}

	return &secrets, nil
}

func (ss *SecretsService) DeleteSecret(ctx context.Context, clusterId string, scopeType authorizationClient.ScopeType,
	scopeId string, name string,
) *api.Error {
	if scopeType == authorizationClient.SCOPETYPE_PROJECT {
		var scopeError *api.Error
		scopeId, scopeError = namespaceToProjectId(ctx, clusterId, scopeId)
		if scopeError != nil {
			return scopeError
		}
	}

	tenantId, err := common.TenantIdFromContext(ctx)
	if err != nil {
		return err
	}

	filters := models.SecretFilters{
		TenantId:  tenantId,
		ClusterId: &clusterId,
		ScopeQuery: &models.ScopeQueryParams{
			ScopeType: string(scopeType),
			ScopeId:   scopeId,
		},
		SecretName: &name,
	}

	err = dao.DeleteSecret(ctx, &filters)
	if err != nil {
		return err
	}

	return nil
}
