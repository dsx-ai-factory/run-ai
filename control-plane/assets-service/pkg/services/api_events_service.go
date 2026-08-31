/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"
	"strconv"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
)

// EventsAPIService is a service that implements the logic for the EventsAPIServicer
// This service should implement the business logic for every endpoint for the EventsAPI API.
// Include any external packages or services that will be required by this service.
type EventsAPIService struct {
}

type ObjectAssetIdAndScope struct {
	AssetId   string
	ScopeType string
	ScopeId   string
}

// NewEventsAPIService creates a default api service
func NewEventsAPIService() *EventsAPIService {
	return &EventsAPIService{}
}

// SyncEvents - Sync events.
func (s *EventsAPIService) SyncEvents(ctx context.Context, eventsSyncRequest api.EventsSyncRequest) (api.ImplResponse, error) {
	tenantId := auth.TenantIdOf(ctx)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Received %d events to sync", len(eventsSyncRequest.Events))

	eventsModels := make([]*dbmodel.Event, 0, len(eventsSyncRequest.Events))

	resourceUidToAssetIdMap := make(map[string]ObjectAssetIdAndScope)
	err := patchAssetIdsAndScopeOfObjects(ctx, eventsSyncRequest.Events, &resourceUidToAssetIdMap)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	for _, event := range eventsSyncRequest.Events {
		assetIdAndScope, ok := resourceUidToAssetIdMap[event.InvolvedObject.Uid]
		if !ok {
			logging.Warnf("Asset id and scope not found for event %s", event.Id)
			continue
		}
		model := dbmodel.NewEventModelFromSyncApi(tenantId, &event, assetIdAndScope.AssetId, assetIdAndScope.ScopeId, assetIdAndScope.ScopeType)
		eventsModels = append(eventsModels, &model)
	}

	if len(eventsModels) != 0 {
		dbErr := dao.UpsertEvents(ctx, eventsModels)
		if dbErr != nil {
			return NewErrorResponse(dbErr), nil
		}
	}

	return api.Response(http.StatusOK, nil), nil
}

func patchAssetIdsAndScopeOfObjects(ctx context.Context, events []api.EventSyncFields, resourceUidToAssetIdMap *map[string]ObjectAssetIdAndScope) *api.Error {
	for _, event := range events {
		if _, ok := (*resourceUidToAssetIdMap)[event.InvolvedObject.Uid]; ok {
			continue
		}

		if event.InvolvedObject.Kind != dbmodel.PVC_OBJECT_KIND {
			continue
		}

		assetMeta, err := dao.FindAssetMetaOfPVCEvent(ctx, &event)
		if err != nil {
			logging.Warnf("failed to find asset meta for event %s, %v", event.Id, err)
			continue
		}

		scopeId := ""
		switch assetMeta.Scope {
		case api.SCOPE_CLUSTER:
			scopeId = *assetMeta.ClusterId
		case api.SCOPE_DEPARTMENT:
			scopeId = *assetMeta.DepartmentId
		case api.SCOPE_PROJECT:
			scopeId = strconv.Itoa(int(*assetMeta.ProjectId))
		}
		(*resourceUidToAssetIdMap)[event.InvolvedObject.Uid] = ObjectAssetIdAndScope{
			AssetId:   assetMeta.Id,
			ScopeType: string(assetMeta.Scope),
			ScopeId:   scopeId,
		}

	}
	return nil
}
