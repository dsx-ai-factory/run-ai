/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"

	events "github.com/run-ai/runai/runai-application-packages/async-events"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"

	org_unit "github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/clients/org-unit"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"

	"k8s.io/utils/ptr"

	authorizationClient "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/auth"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/common"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

const (
	StatusNoIssuesFound = "No issues found"
	StatusIssuesFound   = "Issues found"
)

func (ps *PvcsService) CreatePvc(ctx context.Context, clusterId string, scopeType authorizationClient.ScopeType, scopeId string,
	pvcName string, request api.PvcCreationRequest,
) (*api.Pvc, *api.Error) {
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

	record := models.Pvc{
		PvcFilters: models.PvcFilters{
			TenantId:  tenantId,
			ClusterId: &clusterId,
			ScopeQuery: &models.ScopeQueryParams{
				ScopeType: string(scopeType),
				ScopeId:   scopeId,
			},
			PvcName: &pvcName,
			Origin:  dao.OriginPtrToStringPtr(request.Origin),
		},
		StorageClassName:  request.StorageClassName,
		RequestedCapacity: request.RequestedCapacity,
		ActualCapacity:    request.ActualCapacity,
		DataSharing:       request.DataSharing,
	}

	if request.Phase != nil {
		record.Phase = convutils.Ref(string(*request.Phase))
	}

	pvc, isNewPvc, err := dao.CreatePvc(ctx, &record)
	if err != nil {
		return nil, err
	}

	issuesChanges, err := HandleResourceIssues(ctx, tenantId, clusterId, pvc.ScopeQuery.ScopeType, pvc.ScopeQuery.ScopeId, pvc.ID.String(), request.Issues, string(api.CLUSTERRESOURCETYPE_PVC))
	if err != nil {
		return nil, err
	}

	status, shouldPublish := checkStatusChange(issuesChanges, isNewPvc != nil && *isNewPvc)
	if shouldPublish {
		publishPvcEvents(ctx, &[]events.PvcEvent{*events.NewPvcEvent(clusterId, pvc.ScopeQuery.ScopeType, pvc.ScopeQuery.ScopeId, pvcName, status)})
	}

	return &api.Pvc{
		Name:   pvcName,
		Origin: request.Origin,
	}, nil
}

func (ps *PvcsService) ListPvcs(ctx context.Context, scopeTypeStr string, scopeId string, optionalPvcName string, clusterId string,
	optionalOrigin string, includeAncestors bool, includeDescendants bool,
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

	filters := models.PvcFilters{
		TenantId: tenantId,
	}
	var ancestorsResolver *org_unit.AncestorsResolver

	if clusterId != "" {
		ancestorsResolver = org_unit.NewAncestorsResolverOfCluster(clusterId)
	} else {
		ancestorsResolver = org_unit.NewAncestorsResolver(orgUnitApi.ScopeType(scopeTypeStr), scopeId)
	}

	filters.ScopeQuery = models.NewScopeQueryParams(ancestorsResolver)

	authErr := auth.CheckReadPermission(ctx, authorizationClient.RESOURCETYPE_PVC_ASSETS, ancestorsResolver)
	if authErr != nil {
		return NewErrorResponse(authErr), nil
	}

	if optionalPvcName != "" {
		filters.PvcName = &optionalPvcName
	}

	filters.Origin = dao.OriginFilterToStringPtr(optionalOrigin)

	result, err := dao.ListPvcs(ctx, &filters, &models.ListOptions{
		IncludeAncestors:   includeAncestors,
		IncludeDescendants: includeDescendants,
	})
	if err != nil {
		return NewErrorResponse(err), nil
	}

	payload := api.PvcListResponse{
		Pvcs: make([]api.PvcListInfo, 0, len(result)),
	}
	for _, pvc := range result {
		origin := api.Origin(ptr.Deref(pvc.Origin, ""))

		pvcListInfo := api.PvcListInfo{
			ClusterId: ptr.Deref(pvc.ClusterId, ""),
			Name:      ptr.Deref(pvc.PvcName, ""),
			Origin:    &origin,

			ActualCapacity:    pvc.ActualCapacity,
			RequestedCapacity: pvc.RequestedCapacity,
			StorageClassName:  pvc.StorageClassName,
			DataSharing:       pvc.DataSharing,
		}

		if pvc.Phase != nil {
			pvcListInfo.Phase = convutils.Ref(api.PvcPhaseOptional(*pvc.Phase))
		}

		if pvc.ScopeQuery != nil {
			pvcListInfo.ScopeType = pvc.ScopeQuery.ScopeType
			pvcListInfo.ScopeId = pvc.ScopeQuery.ScopeId
		}

		payload.Pvcs = append(payload.Pvcs, pvcListInfo)
	}

	return api.Response(http.StatusOK, payload), nil
}

func (ps *PvcsService) SyncPvcs(ctx context.Context, clusterId string, scopeType authorizationClient.ScopeType, scopeId string,
	pvcs api.Pvcs,
) (*api.Pvcs, *api.Error) {
	for _, pvc := range pvcs.Pvcs {
		if pvc.Name == "" {
			return nil, api.NewValidationError("mandatory PVC name is not provided")
		}
		if pvc.Origin == nil || *pvc.Origin == "" {
			pvc.Origin = &originExisting
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

	filters := models.PvcFilters{
		TenantId:  tenantId,
		ClusterId: &clusterId,
		ScopeQuery: &models.ScopeQueryParams{
			ScopeType: string(scopeType),
			ScopeId:   scopeId,
		},
	}

	tx := runaidb.DB().Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	pvcIdMap, isNewPvcMap, err := dao.ApplyPvcs(ctx, &filters, pvcs.Pvcs)
	if err != nil {
		return nil, err
	}

	eventsToPublish := make([]events.PvcEvent, 0, len(pvcIdMap))
	for id, pvc := range pvcIdMap {
		issuesChanges, err := HandleResourceIssues(ctx, tenantId, clusterId, string(scopeType), scopeId, id, pvc.Issues, string(api.CLUSTERRESOURCETYPE_PVC))
		if err != nil {
			return nil, err
		}

		status, shouldPublish := checkStatusChange(issuesChanges, isNewPvcMap[id] == dao.NewlyCreatedPvc)
		if shouldPublish {
			eventsToPublish = append(eventsToPublish, *events.NewPvcEvent(
				clusterId,
				string(scopeType),
				scopeId,
				pvc.Name,
				status,
			))
		}
	}

	if len(eventsToPublish) > 0 {
		publishPvcEvents(ctx, &eventsToPublish)
	}

	pvcIds := make([]string, 0, len(pvcIdMap))
	for id := range pvcIdMap {
		pvcIds = append(pvcIds, id)
	}
	err = dao.DeletePvcsNotInCluster(ctx, &filters, pvcIds)
	if err != nil {
		return nil, err
	}

	if commitErr := tx.Commit(); commitErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "Failed to commit transaction", commitErr)
	}
	committed = true

	return &pvcs, nil
}

func (ps *PvcsService) DeletePvc(ctx context.Context, clusterId string, scopeType authorizationClient.ScopeType, scopeId string,
	name string,
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

	filters := models.PvcFilters{
		TenantId:  tenantId,
		ClusterId: &clusterId,
		ScopeQuery: &models.ScopeQueryParams{
			ScopeType: string(scopeType),
			ScopeId:   scopeId,
		},
		PvcName: &name,
	}

	err = dao.DeletePvc(ctx, &filters)
	if err != nil {
		return err
	}

	return nil
}

func publishPvcEvents(ctx context.Context, pvcEvents *[]events.PvcEvent) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Publishing event")
	pubSubErr := runaipubsub.Publish[events.PvcEvents](ctx, events.NewPvcEvents(*pvcEvents))
	if pubSubErr != nil {
		// We only log error since we don't want to block the request if the pubsub fails
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to publish pvc events, err: %v", pubSubErr)
	}
}

func checkStatusChange(issuesChanges *dao.IssuesChanges, isNewPvc bool) (string, bool) {
	// For new PVCs, always publish event with initial status
	if isNewPvc {
		return StatusNoIssuesFound, true
	}

	// For existing PVCs, only publish if status changed
	if issuesChanges.Removed == 0 && issuesChanges.Added != 0 {
		return StatusIssuesFound, true
	}

	if issuesChanges.Removed != 0 && issuesChanges.Added == 0 {
		return StatusNoIssuesFound, true
	}

	return "", false
}
