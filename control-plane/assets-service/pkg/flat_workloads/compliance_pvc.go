/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flat_workloads

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	k8sTrackerClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/k8s-objects-tracker"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	k8sTrackerApi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	workloadsApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

// patchPvcStorageClassCompliance patch StorageClassCompliance for pvc assets.
// no error returned, to allow smooth experience
// in case there is a compliance issue, later on there is another check and the workload will fail
func (cip *ComplianceInfoPatcher) patchPvcStorageClassCompliance(ctx context.Context) {
	// Create a map of pvcs by cluster id and storage class
	// cluster id --> set of storage classes --> list of pvc ids
	pvcs := make(map[string]map[string][]string)
	for assetId := range cip.assetIdsToPatch {
		ptrs := cip.patchingInfo[assetId]
		if ptrs.Meta.Kind != api.ASSETKIND_PVC {
			continue
		}
		if _, exists := pvcs[utils.Deref(ptrs.Meta.ClusterId)]; !exists {
			pvcs[utils.Deref(ptrs.Meta.ClusterId)] = make(map[string][]string)
		}
		pvcSpec, ok := castPatchSpecToPvcSpec(ptrs.Spec)
		if !ok {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to cast spec to PVC spec")
			continue
		}
		if pvcSpec.ClaimInfo == nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("missing claim info for pvc id %v", assetId)
			continue
		}
		if pvcSpec.ClaimInfo.StorageClass != nil {
			pvcs[utils.Deref(ptrs.Meta.ClusterId)][*pvcSpec.ClaimInfo.StorageClass] = append(pvcs[utils.Deref(ptrs.Meta.ClusterId)][*pvcSpec.ClaimInfo.StorageClass], assetId)
		}
	}
	// if no pvc was found, nothing to do here
	if len(pvcs) == 0 {
		return
	}

	// the first layer of pvcs is cluster, iterate over the cluster
	for clusterId := range pvcs {

		// build a list of storage class used in the cluster
		var storageClasses []string
		for storageClass := range pvcs[clusterId] {
			storageClasses = append(storageClasses, storageClass)
		}
		// get the configuration of storageClasses in the cluster
		configs, err := k8sTrackerClient.Get().GetStorageClassesConfigInt(ctx, clusterId, storageClasses)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to get storage class configuration. cluser id %s, storage classes [%v], err [%+v]", clusterId, storageClasses, err)
			continue
		}

		for _, storageClassConfig := range configs {
			for _, pvcId := range pvcs[clusterId][storageClassConfig.Name] {
				// storage class is not allowed for running workload
				ptrs := cip.patchingInfo[pvcId]

				if !*storageClassConfig.Permissions.Get().AllowedForWorkloads.Get() {
					setStorageClassIssue(&ptrs, "The underlying storage class of the PVC is forbidden for running workloads")
					continue
				}

				pvcSpec, castOk := castPatchSpecToPvcSpec(ptrs.Spec)
				if castOk {
					configErr := ValidatePvcSpecVsScCustomization(ptrs.Meta.AutoDelete, pvcSpec, &storageClassConfig)
					if configErr != nil {
						setStorageClassIssue(&ptrs, configErr.Message)
						continue
					}
				}
			}
		}
	}
	cip.updateRemainingAssets()
}

func ValidatePvcSpecVsScCustomization(autoDeleteAsset *bool, pvc *api.PvcAssetSpec, config *k8sTrackerApi.StorageClassConfiguration) *api.Error {
	if convutils.BoolIsTrue(autoDeleteAsset) {
		return nil // WAIT_FOR_OFER AutoDelete assets (volumes) not yet supporting attributes
	}

	if config.Customization.Get() == nil {
		return MakeSureNoAddedAttrs(pvc.ClaimInfo)
	}

	scName := *pvc.ClaimInfo.StorageClass
	if scName == "" {
		scName = composer.NoneStorageClass
	}

	// convert claimInfo to workloadsApi
	claimInfo, convErr := convertClaimInfoToWorkloadsApi(pvc.ClaimInfo)
	if convErr != nil {
		return convErr
	}

	configErr := composer.ApplyStorageClassConfig(scName, &claimInfo, config)
	if configErr != nil {
		return api.NewError(int(configErr.Code), configErr.Message, nil)
	}

	//  update the claimInfo with the updated set of attributes
	pvc.ClaimInfo.AddedAttrValues = make([]api.PvcAddedAttrValue, 0, len(claimInfo.AddedAttrValues))

	for _, attr := range claimInfo.AddedAttrValues {
		pvc.ClaimInfo.AddedAttrValues = append(pvc.ClaimInfo.AddedAttrValues, api.PvcAddedAttrValue{
			Key:   attr.Key,
			Value: convutils.Deref(attr.Value, ""),
		})
	}

	return nil
}

func convertClaimInfoToWorkloadsApi(source *api.ClaimInfo) (*workloadsApi.ClaimInfo, *api.Error) {
	txt, jsonErr := json.Marshal(source)
	if jsonErr != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to parse the claim info", jsonErr)
	}

	var result workloadsApi.ClaimInfo
	jsonErr = json.Unmarshal(txt, &result)
	if jsonErr != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to compose the claim info", jsonErr)
	}

	return &result, nil
}

func MakeSureNoAddedAttrs(claimInfo *api.ClaimInfo) *api.Error {
	if claimInfo == nil {
		return nil
	}

	if len(claimInfo.AddedAttrValues) != 0 {
		return api.NewValidationError("addedAttrValues are not permitted for PVC of this storage class")
	}

	return nil
}

func setStorageClassIssue(ptrs *response_patching.ResponsePatchingPtrs, message string) {
	if *ptrs.ComplianceInfo == nil {
		*ptrs.ComplianceInfo = &api.ComplianceInfo{}
	}
	(*ptrs.ComplianceInfo).Compliance = false
	(*ptrs.ComplianceInfo).Reason = []api.ComplianceInfoReason{
		{
			Details: message,
		},
	}
}

// castPatchSpecToPvcSpec casts a "general spec" to pvc asset spec
func castPatchSpecToPvcSpec(spec interface{}) (*api.PvcAssetSpec, bool) {
	switch assetSpec := spec.(type) {
	case *api.PvcAssetSpec:
		return assetSpec, true
	case *api.DatasourceListResponseAssetSpec:
		// if Pvc isn't nil than the content of the response is PVC
		return assetSpec.Pvc, assetSpec.Pvc != nil
	}
	return nil, false
}

func (cip *ComplianceInfoPatcher) patchDataVolumeCompliance(ctx context.Context) {
	for assetId := range cip.assetIdsToPatch {
		ptr := cip.patchingInfo[assetId]
		if ptr.Meta.Kind != api.ASSETKIND_DATA_VOLUME {
			continue
		}

		if ptr.Status == nil {
			continue
		}

		if *ptr.ComplianceInfo == nil {
			*ptr.ComplianceInfo = &api.ComplianceInfo{}
		}

		(*ptr.ComplianceInfo).Compliance = dataVolumeStatusToCompliance((*ptr.Status).Status)
		if !(*ptr.ComplianceInfo).Compliance {
			(*ptr.ComplianceInfo).Reason = []api.ComplianceInfoReason{
				{
					Details: fmt.Sprintf("The data volume status is %s", (*ptr.Status).Status),
				},
			}
		}
	}

	cip.updateRemainingAssets()
}

func dataVolumeStatusToCompliance(status api.AssetSyncStatusEnum) bool {
	switch status {
	case api.ASSETSYNCSTATUSENUM_NO_ISSUES_FOUND, api.ASSETSYNCSTATUSENUM_NO_STATUS:
		return true
	case api.ASSETSYNCSTATUSENUM_FAILED, api.ASSETSYNCSTATUSENUM_DELETING, api.ASSETSYNCSTATUSENUM_CREATING:
		return false
	default:
		return true
	}
}
