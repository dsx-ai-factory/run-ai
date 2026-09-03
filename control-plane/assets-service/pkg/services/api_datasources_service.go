/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/flat_workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/projects"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/asset_cluster_status"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/status_info"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/usage_info"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
)

// DatasourcesApiService is a service that implements the logic for the Data-sources service
type DatasourcesApiService struct {
	configMapDBAccessor ConfigMapsDBAccessor
	secretDBAccessor    SecretVolumeDBAccessor
	dataVolumesAccessor DataVolumesAccessor
}

// NewDatasourcesApiService creates a default api service
func NewDatasourcesApiService(configMapDBAccessor ConfigMapsDBAccessor, secretDBAccessor SecretVolumeDBAccessor, dataVolumesAccessor DataVolumesAccessor) *DatasourcesApiService {
	return &DatasourcesApiService{
		configMapDBAccessor: configMapDBAccessor,
		secretDBAccessor:    secretDBAccessor,
		dataVolumesAccessor: dataVolumesAccessor,
	}
}

func (s *DatasourcesApiService) ListDatasourceAssets(ctx context.Context, nameParam string, scopeParam string,
	projectIdParam int32, departmentId string, clusterId string, usageInfo bool,
	complianceProject int32, complianceWorkloadType string, statusInfo bool, assetIdsParam string, complianceReplicaType string, includeDataVolumes bool, includeDescendants bool, typeParam string,
) (api.ImplResponse, error) {

	if includeDataVolumes && typeParam != "" && typeParam != "dataVolume" {
		return NewErrorResponse(api.NewError(
			http.StatusBadRequest,
			"The parameters includeDataVolumes and type are mutually exclusive",
			nil,
		)), nil
	}

	queryParams, queryError := query_types.NewQueryParams(assetIdsParam, nameParam, scopeParam, includeDescendants, projectIdParam, departmentId, clusterId, nil)
	if queryError != nil {
		return NewErrorResponse(queryError), nil
	}

	permittedScopes, authError := query_types.ApplyRbacRules(ctx, queryParams)
	if authError != nil {
		return NewErrorResponse(authError), nil
	}

	result := api.DatasourceListResponse{
		Entries: make([]api.DatasourceListResponseEntry, 0),
	}

	type handler struct {
		dsType   string
		addFunc  func(ctx context.Context, result *api.DatasourceListResponse, qp *query_types.QueryParams) *api.Error
		included bool
	}

	handlers := []handler{
		{"hostPath", s.addHostVolumes, true},
		{"nfs", s.addNfsVolumes, true},
		{"pvc", func(ctx context.Context, r *api.DatasourceListResponse, p *query_types.QueryParams) *api.Error {
			return s.addPvcs(ctx, r, p, statusInfo)
		}, true},
		{"git", s.addGitVolumes, true},
		{"s3", s.addS3Volumes, true},
		{"configMap", s.addConfigMapVolumes, true},
		{"secretVolume", s.addSecretVolume, true},
		{"dataVolume", s.addDataVolumes, includeDataVolumes},
	}

	var actionError *api.Error

	for _, h := range handlers {
		if h.included && (typeParam == "" || typeParam == h.dsType) {
			actionError = h.addFunc(ctx, &result, queryParams)
			if actionError != nil {
				return NewErrorResponse(actionError), nil
			}
		}
	}
	//
	// NOTE!!! -> when adding new datasources you must update the mapping tables in:
	//                  workloads-and-policies/workload_composer/assets_kind_mapping.go
	//            please add yours addWhatEver call BEFORE this note, so next time someone adds datasource they will still see it!
	//

	patchingPointers := response_patching.LoadResponsePatchingPtrs(result.Entries)

	projects.PatchProjectNames(ctx, patchingPointers)

	if usageInfo {
		patchErr := usage_info.PatchUsageInfoMultipleAssets(ctx, patchingPointers, queryParams.AllowedOrgUnits.AllowedProjects)
		if patchErr != nil {
			return NewErrorResponse(patchErr), nil
		}
	}

	if (complianceProject != 0 && complianceWorkloadType != "") || statusInfo {
		assetClusterStatus := asset_cluster_status.NewAssetClusterStatus(patchingPointers, permittedScopes)

		if complianceProject != 0 && complianceWorkloadType != "" {
			patcher, patchErr := flat_workloads.NewComplianceInfoPatcher(ctx, complianceProject, complianceWorkloadType, complianceReplicaType, patchingPointers, permittedScopes, assetClusterStatus)
			if patchErr != nil {
				return NewErrorResponse(patchErr), nil
			}

			patchErr = patcher.PatchComplianceInfo(ctx)
			if patchErr != nil {
				return NewErrorResponse(patchErr), nil
			}

			patchErr = usage_info.PatchUsageTimes(ctx, complianceProject, complianceWorkloadType, patchingPointers)
			if patchErr != nil {
				return NewErrorResponse(patchErr), nil
			}
		}

		if statusInfo {
			patchErr := status_info.PatchStatusInfoMultipleAssets(ctx, assetClusterStatus, patchingPointers)
			if patchErr != nil {
				return NewErrorResponse(patchErr), nil
			}
		}
	}

	return api.Response(http.StatusOK, result), nil
}

// CreateDatasourceAsset is a placeholder. the code generation assumes create must exist in the swagger
// and if you do not provide it, build fails
func (s *DatasourcesApiService) CreateDatasourceAsset(context.Context, api.AssetMeta) (api.ImplResponse, error) {
	return api.Response(http.StatusNotImplemented, nil), nil
}

func (s *DatasourcesApiService) addConfigMapVolumes(ctx context.Context, result *api.DatasourceListResponse, queryParams *query_types.QueryParams) *api.Error {
	volumes, err := s.configMapDBAccessor.ListConfigMapAssets(ctx, queryParams)
	if err != nil {
		return err
	}

	entries := make([]api.DatasourceListResponseEntry, 0, len(volumes.Entries))

	for _, ds := range volumes.Entries {
		entries = append(entries, api.DatasourceListResponseEntry{
			Meta: ds.Meta,
			Spec: api.DatasourceListResponseAssetSpec{ConfigMap: &ds.Spec},
		})
	}

	result.Entries = append(result.Entries, entries...)

	return nil
}

func (s *DatasourcesApiService) addDataVolumes(ctx context.Context, result *api.DatasourceListResponse, queryParams *query_types.QueryParams) *api.Error {
	if queryParams.ProjectId == nil {
		return api.NewError(http.StatusBadRequest, "when listing data volumes project id must be provided", fmt.Errorf("when listing data volumes project id must be provided"))
	}
	if queryParams.ClusterId != nil {
		return api.NewError(http.StatusBadRequest, "should not provided cluster id when listing data volumes", fmt.Errorf("should not provided cluster id when listing data volumes"))
	}
	if queryParams.DepartmentId != nil {
		return api.NewError(http.StatusBadRequest, "should not provided department id when listing data volumes", fmt.Errorf("should not provided department id when listing data volumes"))
	}

	dataVolumes, err := s.dataVolumesAccessor.ListDataVolumes(ctx, queryParams)
	if err != nil {
		return api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list data volumes assets", err)
	}

	result.Entries = append(result.Entries, dataVolumes...)
	return nil
}

func (s *DatasourcesApiService) addSecretVolume(ctx context.Context, result *api.DatasourceListResponse, queryParams *query_types.QueryParams) *api.Error {
	volumes, err := s.secretDBAccessor.ListSecretVolumeAssets(ctx, queryParams)
	if err != nil {
		return err
	}

	entries := make([]api.DatasourceListResponseEntry, 0, len(volumes.Entries))

	for _, ds := range volumes.Entries {
		entries = append(entries, api.DatasourceListResponseEntry{
			Meta: ds.Meta,
			Spec: api.DatasourceListResponseAssetSpec{Secret: &ds.Spec},
		})
	}

	result.Entries = append(result.Entries, entries...)

	return nil
}

func (s *DatasourcesApiService) addHostVolumes(ctx context.Context, result *api.DatasourceListResponse, queryParams *query_types.QueryParams) *api.Error {
	volumes, actionError := dao.ListHostPathAssets(ctx, queryParams)
	if actionError != nil {
		return actionError
	}

	entries := make([]api.DatasourceListResponseEntry, 0)

	for _, ds := range volumes.Entries {
		entries = append(entries, api.DatasourceListResponseEntry{
			Meta: ds.Meta,
			Spec: api.DatasourceListResponseAssetSpec{HostPath: &ds.Spec},
		})
	}

	result.Entries = append(result.Entries, entries...)

	return nil
}

func (s *DatasourcesApiService) addNfsVolumes(ctx context.Context, result *api.DatasourceListResponse, queryParams *query_types.QueryParams) *api.Error {
	volumes, actionError := dao.ListNFSAssets(ctx, queryParams)
	if actionError != nil {
		return actionError
	}

	entries := make([]api.DatasourceListResponseEntry, 0)

	for _, ds := range volumes.Entries {
		entries = append(entries, api.DatasourceListResponseEntry{
			Meta: ds.Meta,
			Spec: api.DatasourceListResponseAssetSpec{Nfs: &ds.Spec},
		})
	}

	result.Entries = append(result.Entries, entries...)

	return nil
}

func (s *DatasourcesApiService) addPvcs(ctx context.Context, result *api.DatasourceListResponse,
	queryParams *query_types.QueryParams, statusInfo bool,
) *api.Error {
	options := dao.DatasourcesOptions{
		IncludeDeleted: statusInfo, // if we include status info we can include deleted assets and show their deletion status
	}

	volumes, actionError := dao.ListPvcAssets(ctx, queryParams, options)
	if actionError != nil {
		return actionError
	}

	entries := make([]api.DatasourceListResponseEntry, 0)
	for _, ds := range volumes.Entries {
		entries = append(entries, api.DatasourceListResponseEntry{
			Meta:        ds.Meta,
			Spec:        api.DatasourceListResponseAssetSpec{Pvc: &ds.Spec},
			ClusterInfo: ds.ClusterInfo,
		})
	}

	result.Entries = append(result.Entries, entries...)

	return nil
}

func (s *DatasourcesApiService) addGitVolumes(ctx context.Context, result *api.DatasourceListResponse, queryParams *query_types.QueryParams) *api.Error {
	volumes, actionError := dao.ListGitAssets(ctx, queryParams)
	if actionError != nil {
		return actionError
	}

	entries := make([]api.DatasourceListResponseEntry, 0)

	for _, ds := range volumes.Entries {
		entries = append(entries, api.DatasourceListResponseEntry{
			Meta: ds.Meta,
			Spec: api.DatasourceListResponseAssetSpec{Git: &ds.Spec},
		})
	}

	result.Entries = append(result.Entries, entries...)

	return nil
}

func (s *DatasourcesApiService) addS3Volumes(ctx context.Context, result *api.DatasourceListResponse, queryParams *query_types.QueryParams) *api.Error {
	volumes, actionError := dao.ListS3Assets(ctx, queryParams)
	if actionError != nil {
		return actionError
	}

	entries := make([]api.DatasourceListResponseEntry, 0)

	for _, ds := range volumes.Entries {
		entries = append(entries, api.DatasourceListResponseEntry{
			Meta: ds.Meta,
			Spec: api.DatasourceListResponseAssetSpec{S3: &ds.Spec},
		})
	}

	result.Entries = append(result.Entries, entries...)

	return nil
}

func (s *SyncApiService) DatasourceSyncUpdates(ctx context.Context, clusterId string, after int64, clusterSyncReport api.ClusterSyncReport) (api.ImplResponse, error) {
	unifiedResult, err := s.handleAgentSyncUpdates(ctx, clusterId, after, &clusterSyncReport, dbmodel.DatasourceAssetsToSync, false /*no need to decrypt*/)
	if err != nil {
		return api.Response(int(err.Code), err), nil
	}

	myResult := unifiedResult.ToDatasource()
	return api.Response(http.StatusOK, myResult), nil
}
