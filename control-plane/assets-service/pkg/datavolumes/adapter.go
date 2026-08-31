/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datavolumes

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	dvclient "github.com/run-ai/runai/runai-clients/datavolumes-service-client"
	dvapi "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	// MaxDvResultCount is a hard coded number of dv returned by the adapter
	MaxDvResultCount = 1000
	// LimitDvRequestSize is a limit
	LimitDvRequestSize = 500
	// DvFilterName is the filter name for the name property of the dv
	DvFilterName = "name"
)

// Adapter is a bridge between the assets service and the data volumes service
type Adapter struct {
	client dvclient.IDatavolumesClient
}

func NewAdapter(client dvclient.IDatavolumesClient) *Adapter {
	return &Adapter{
		client: client,
	}
}

// ListDatavolumes lists data volumes
func (a *Adapter) ListDataVolumes(ctx context.Context, queryParams *query_types.QueryParams) ([]api.DatasourceListResponseEntry, error) {
	filters := buildFiltersFromQueryParams(queryParams)
	projectId := fmt.Sprintf("%d", convutils.Deref(queryParams.ProjectId, 0))

	datavolumes, err := a.aggregateDataVolumes(ctx, projectId, filters, MaxDvResultCount, LimitDvRequestSize)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to get data volumes from data colume service. err [%v]", err)
		return nil, err
	}

	assets := make([]api.DatasourceListResponseEntry, 0, len(datavolumes))
	for _, dv := range datavolumes {
		assets = append(assets, castDataVolumeToAsset(dv, convutils.Deref(queryParams.ProjectId, 0)))
	}

	return assets, nil
}

// aggregateDataVolumes calls get dv API multipal times, until it gets all dv to reach total limit
func (a *Adapter) aggregateDataVolumes(ctx context.Context, projectId string, filters []string, totalLimit int, requsetLimit int) ([]dvapi.DatavolumeNoSharedScopes, error) {
	var result []dvapi.DatavolumeNoSharedScopes
	var offset int32 = 0
	done := false
	for !done {
		resp, err := a.client.GetDatavolumes(ctx, dvapi.DATAVOLUMEREQUESTTYPE_USABLE_IN_PROJECT, projectId, int32(requsetLimit), offset, filters, "", "")
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to get data volume from data volume service: %v", err)
			return nil, err
		}
		result = append(result, resp.Datavolumes...)

		done = len(result) >= totalLimit || resp.Next == nil
		offset = convutils.Deref(resp.Next, 0) // safe access even if next is nil
	}
	return result, nil
}

// buildFilterFromQueryParams build filters for the data volume client using GET data source query params
func buildFiltersFromQueryParams(queryParams *query_types.QueryParams) []string {
	var filters []string

	if queryParams.Name != nil {
		filters = append(filters, fmt.Sprintf("%s=%s", DvFilterName, convutils.Deref(queryParams.Name, "")))
	}

	return filters
}

// castDataVolumeToAsset casts the data volume into "data volume" asset form
func castDataVolumeToAsset(dv dvapi.DatavolumeNoSharedScopes, projectId int32) api.DatasourceListResponseEntry {
	return api.DatasourceListResponseEntry{
		Meta: api.AssetMeta{
			Name:        dv.Name,
			Description: dv.Description,
			Scope:       api.SCOPE_PROJECT,
			ProjectId:   convutils.Ref(projectId),
			ClusterId:   dv.ClusterId,
			Id:          dv.Id,
			Kind:        api.ASSETKIND_DATA_VOLUME,
			CreatedBy:   convutils.Deref(dv.CreatedBy, ""),
			CreatedAt:   convutils.Deref(dv.CreatedAt, time.Time{}),
			UpdatedBy:   convutils.Deref(dv.UpdatedBy, ""),
			UpdatedAt:   convutils.Deref(dv.UpdatedAt, time.Time{}),
		},
		Spec: api.DatasourceListResponseAssetSpec{
			DataVolume: &api.DataVolumeAssetSpec{
				Id:            convutils.Ref(dv.Id),
				OriginPvcName: convutils.Ref(dv.OriginPvcName),
			},
		},
		Status: &api.AssetClusterStatusInfo{
			Status:  dataVolumePhaseToAssetStatus(dv.Status.Phase),
			Message: convutils.Deref(dv.Status.PhaseMessage, ""),
		},
	}
}

// dataVolumePhaseToAssetStatus tx the data volume phase into an assets status
func dataVolumePhaseToAssetStatus(phase dvapi.DatavolumePhase) api.AssetSyncStatusEnum {
	switch phase {
	case dvapi.DATAVOLUMEPHASE_READY:
		return api.ASSETSYNCSTATUSENUM_NO_ISSUES_FOUND
	case dvapi.DATAVOLUMEPHASE_NOT_READY:
		return api.ASSETSYNCSTATUSENUM_FAILED
	case dvapi.DATAVOLUMEPHASE_CREATING:
		return api.ASSETSYNCSTATUSENUM_CREATING
	case dvapi.DATAVOLUMEPHASE_DELETING:
		return api.ASSETSYNCSTATUSENUM_DELETING
	case dvapi.DATAVOLUMEPHASE_UNKNOWN:
		return api.ASSETSYNCSTATUSENUM_NO_STATUS
	}
	return api.ASSETSYNCSTATUSENUM_NO_STATUS
}
