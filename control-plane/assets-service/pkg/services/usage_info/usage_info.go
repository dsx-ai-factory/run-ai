/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package usage_info

import (
	"context"

	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/workloads"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
)

// PatchUsageInfoMultipleAssets performs PatchUsageInfo on set of assets containing mixed asset kinds,
// so it first needs to sort out the patching pointers based on asset kind, then calls PatchUsageInfo per kind
func PatchUsageInfoMultipleAssets(ctx context.Context, patchingInfo map[string]response_patching.ResponsePatchingPtrs, allowedProjectIds []string) *api.Error {
	patchingInfoPerKind := map[api.AssetKind]map[string]response_patching.ResponsePatchingPtrs{}

	for assetId, assetInfo := range patchingInfo {
		assetKind := assetInfo.Meta.Kind

		if patchingInfoPerKind[assetKind] == nil {
			patchingInfoPerKind[assetKind] = map[string]response_patching.ResponsePatchingPtrs{
				assetId: assetInfo,
			}
		} else {
			patchingInfoPerKind[assetKind][assetId] = assetInfo
		}
	}

	for assetKind, patchingInfoOfThisKind := range patchingInfoPerKind {
		err := PatchUsageInfo(ctx, assetKind, patchingInfoOfThisKind, allowedProjectIds)
		if err != nil {
			return err
		}
	}

	return nil
}

func PatchUsageInfo(ctx context.Context, assetKind api.AssetKind, patchingInfo map[string]response_patching.ResponsePatchingPtrs, allowedProjectIds []string) *api.Error {
	if len(patchingInfo) == 0 {
		return nil
	}

	var singleId *string
	if len(patchingInfo) == 1 {
		for id := range patchingInfo {
			singleId = &id
		}
	}

	switch assetKind {
	case api.ASSETKIND_PASSWORD:
		dbError := findAssetsUsingGivenPasswords(ctx, assetKind, patchingInfo, singleId)
		if dbError != nil {
			return dbError
		}
	case api.ASSETKIND_ACCESS_KEY:
		dbError := findAssetsUsingGivenAccessKeys(ctx, assetKind, patchingInfo, singleId)
		if dbError != nil {
			return dbError
		}
	case api.ASSETKIND_GENERIC_SECRET:
		dbError := findAssetsUsingGivenGenericSecret(ctx, assetKind, patchingInfo, singleId)
		if dbError != nil {
			return dbError
		}
	case api.ASSETKIND_NGC_API_KEY:
		// currently, NGC API keys are not used by any workloads or templates,
		dbError := findAssetsUsingGivenNgcApiKey(ctx, assetKind, patchingInfo, singleId)
		if dbError != nil {
			return dbError
		}
	default:
		for _, workloadType := range []string{auth.WorkloadTypeWorkspace, auth.WorkloadTypeTraining, auth.WorkloadTypeInference, auth.WorkloadTypeDistributed} {
			dbError := findWorkloadsWhichUseGivenAssets(ctx, workloadType, assetKind, patchingInfo, singleId, &allowedProjectIds)
			if dbError != nil {
				return dbError
			}
		}

		dbError := findTemplatesUsingGivenAssets(ctx, assetKind, patchingInfo, singleId)
		if dbError != nil {
			return dbError
		}
	}

	return nil
}

func findWorkloadsWhichUseGivenAssets(ctx context.Context, workloadType string, assetKind api.AssetKind,
	patchingInfo map[string]response_patching.ResponsePatchingPtrs, singleId *string, allowedProjectIds *[]string,
) *api.Error {
	_, workloads, dbError := workloads.SearchWorkloads(ctx, workloadType, singleId, &assetKind, allowedProjectIds)
	if dbError != nil {
		return dbError
	}

	uniqueWorkloads := make(map[string]dbmodel.CombinedWorkloadsRec)
	for _, workload := range workloads {
		_, found := uniqueWorkloads[workload.WorkloadsAssetsRec.WorkloadId]
		if !found {
			uniqueWorkloads[workload.WorkloadsAssetsRec.WorkloadId] = workload
		} else {
			// When an asset is being used by distributed workload, it can be used by the master and workers
			// So we don't need to add the same workload twice
			continue
		}
		curPatch, ok := patchingInfo[workload.WorkloadsAssetsRec.AssetId]
		if ok {
			curWorkloadInfo := api.WorkloadRefAndStatus{
				Id:   workload.WorkloadsAssetsRec.WorkloadId,
				Name: workload.WorkloadRec.Name,
			}
			if workload.WorkloadRec.CurrentState != "" {
				curWorkloadInfo.Status = string(workload.WorkloadRec.CurrentState)
			}

			if *curPatch.UsedBy == nil {
				*curPatch.UsedBy = NewAssetUsageInfo()
			}

			switch workloadType {
			case auth.WorkloadTypeWorkspace:
				(*curPatch.UsedBy).Workspaces = append((*curPatch.UsedBy).Workspaces, curWorkloadInfo)
			case auth.WorkloadTypeTraining:
				(*curPatch.UsedBy).Trainings = append((*curPatch.UsedBy).Trainings, curWorkloadInfo)
			case auth.WorkloadTypeDistributed:
				(*curPatch.UsedBy).Distributed = append((*curPatch.UsedBy).Distributed, curWorkloadInfo)
			default:
				(*curPatch.UsedBy).Inferences = append((*curPatch.UsedBy).Inferences, curWorkloadInfo)
			}
		}
	}
	return nil
}

func findTemplatesUsingGivenAssets(ctx context.Context,
	_ api.AssetKind, patchingInfo map[string]response_patching.ResponsePatchingPtrs, singleId *string,
) *api.Error {
	templates, actionError := dao.ListTemplatesUsingAssets(ctx, singleId)
	if actionError != nil {
		return actionError
	}

	for _, template := range templates {
		curPatch, ok := patchingInfo[template.ContainedAssetId]
		if ok {
			curTemplateInfo := api.AssetRef{
				Id:   template.AssetMeta.Id,
				Name: template.AssetMeta.Name,
			}

			if *curPatch.UsedBy == nil {
				*curPatch.UsedBy = NewAssetUsageInfo()
			}

			(*curPatch.UsedBy).Templates = append((*curPatch.UsedBy).Templates, curTemplateInfo)
		}
	}

	return nil
}

func findAssetsUsingGivenPasswords(ctx context.Context,
	_ api.AssetKind, patchingInfo map[string]response_patching.ResponsePatchingPtrs, singleId *string,
) *api.Error {
	gitAssets, dbError := dao.ListGitUsingPassword(ctx, singleId)
	if dbError != nil {
		return dbError
	}

	for _, gitAsset := range gitAssets.Entries {
		curPatch, ok := patchingInfo[*dbmodel.GetGitCredentialId(gitAsset.Spec)]
		if ok {
			curTemplateInfo := api.AssetDatasourceRef{
				Id:   gitAsset.Meta.Id,
				Name: gitAsset.Meta.Name,
				Kind: gitAsset.Meta.Kind,
			}

			if *curPatch.UsedBy == nil {
				*curPatch.UsedBy = NewAssetUsageInfo()
			}

			(*curPatch.UsedBy).Assets.Datasources = append((*curPatch.UsedBy).Assets.Datasources, curTemplateInfo)
		}
	}

	secretVolumeAssets, err := dao.ListSecretVolumeUsingCredentialId(ctx, singleId)
	if err != nil {
		return err
	}
	for _, secretVolumeAsset := range secretVolumeAssets.Entries {
		curPatch, ok := patchingInfo[*secretVolumeAsset.Spec.CredentialAssetId]
		if ok {
			curTemplateInfo := api.AssetDatasourceRef{
				Id:   secretVolumeAsset.Meta.Id,
				Name: secretVolumeAsset.Meta.Name,
				Kind: secretVolumeAsset.Meta.Kind,
			}

			if *curPatch.UsedBy == nil {
				*curPatch.UsedBy = NewAssetUsageInfo()
			}

			(*curPatch.UsedBy).Assets.Datasources = append((*curPatch.UsedBy).Assets.Datasources, curTemplateInfo)
		}
	}

	// Find environment assets using the given password credential
	return findEnvironmentAssesUsingCredentials(ctx, patchingInfo, singleId)
}

func findAssetsUsingGivenAccessKeys(ctx context.Context,
	_ api.AssetKind, patchingInfo map[string]response_patching.ResponsePatchingPtrs, singleId *string,
) *api.Error {
	s3Assets, dbError := dao.ListS3UsingAccessKey(ctx, singleId)
	if dbError != nil {
		return dbError
	}

	for _, s3Asset := range s3Assets.Entries {
		curPatch, ok := patchingInfo[*s3Asset.Spec.AccessKeyAssetId]
		if ok {
			curTemplateInfo := api.AssetDatasourceRef{
				Id:   s3Asset.Meta.Id,
				Name: s3Asset.Meta.Name,
				Kind: s3Asset.Meta.Kind,
			}

			if *curPatch.UsedBy == nil {
				*curPatch.UsedBy = NewAssetUsageInfo()
			}

			(*curPatch.UsedBy).Assets.Datasources = append((*curPatch.UsedBy).Assets.Datasources, curTemplateInfo)
		}
	}

	secretVolumeAssets, err := dao.ListSecretVolumeUsingCredentialId(ctx, singleId)
	if err != nil {
		return err
	}
	for _, secretVolumeAsset := range secretVolumeAssets.Entries {
		curPatch, ok := patchingInfo[*secretVolumeAsset.Spec.CredentialAssetId]
		if ok {
			curTemplateInfo := api.AssetDatasourceRef{
				Id:   secretVolumeAsset.Meta.Id,
				Name: secretVolumeAsset.Meta.Name,
				Kind: secretVolumeAsset.Meta.Kind,
			}

			if *curPatch.UsedBy == nil {
				*curPatch.UsedBy = NewAssetUsageInfo()
			}

			(*curPatch.UsedBy).Assets.Datasources = append((*curPatch.UsedBy).Assets.Datasources, curTemplateInfo)
		}
	}

	// Find environment assets using the given access key
	return findEnvironmentAssesUsingCredentials(ctx, patchingInfo, singleId)
}

func findAssetsUsingGivenNgcApiKey(ctx context.Context,
	_ api.AssetKind, patchingInfo map[string]response_patching.ResponsePatchingPtrs, singleId *string,
) *api.Error {
	// Currently, NGC API keys are not used by any workloads or templates,
	return nil
}

func findAssetsUsingGivenGenericSecret(ctx context.Context,
	_ api.AssetKind, patchingInfo map[string]response_patching.ResponsePatchingPtrs, singleId *string,
) *api.Error {
	var err *api.Error
	err = findSecretVolumeUsingCredentialId(ctx, patchingInfo, singleId)
	if err != nil {
		return err
	}
	err = findEnvironmentAssesUsingCredentials(ctx, patchingInfo, singleId)
	if err != nil {
		return err
	}
	err = findGitAssetsUsingCredentialId(ctx, patchingInfo, singleId)
	if err != nil {
		return err
	}

	return nil
}

func findSecretVolumeUsingCredentialId(ctx context.Context, patchingInfo map[string]response_patching.ResponsePatchingPtrs, singleId *string) *api.Error {
	secretVolumeAssets, err := dao.ListSecretVolumeUsingCredentialId(ctx, singleId)
	if err != nil {
		return err
	}
	for _, secretVolumeAsset := range secretVolumeAssets.Entries {
		curPatch, ok := patchingInfo[*secretVolumeAsset.Spec.CredentialAssetId]
		if ok {
			curTemplateInfo := api.AssetDatasourceRef{
				Id:   secretVolumeAsset.Meta.Id,
				Name: secretVolumeAsset.Meta.Name,
				Kind: secretVolumeAsset.Meta.Kind,
			}

			if *curPatch.UsedBy == nil {
				*curPatch.UsedBy = NewAssetUsageInfo()
			}

			(*curPatch.UsedBy).Assets.Datasources = append((*curPatch.UsedBy).Assets.Datasources, curTemplateInfo)
		}
	}

	return nil
}

func findEnvironmentAssesUsingCredentials(ctx context.Context, patchingInfo map[string]response_patching.ResponsePatchingPtrs, singleId *string) *api.Error {
	environmentAssets, err := dao.ListEnvironmentAssetsUsingCredentialAsEnvVar(ctx, singleId)
	if err != nil {
		return err
	}
	if environmentAssets == nil {
		return nil
	}
	for _, environmentAsset := range environmentAssets.Entries {
		if environmentAsset.Spec.EnvironmentVariables == nil {
			continue
		}
		for _, envVar := range *environmentAsset.Spec.EnvironmentVariables {
			if envVar.Credential == nil {
				continue
			}
			curPatch, ok := patchingInfo[*envVar.Credential.AssetId]
			if ok {
				curTemplateInfo := api.AssetRef{
					Id:   environmentAsset.Meta.Id,
					Name: environmentAsset.Meta.Name,
				}

				if *curPatch.UsedBy == nil {
					*curPatch.UsedBy = NewAssetUsageInfo()
				}

				(*curPatch.UsedBy).Assets.Environments = append((*curPatch.UsedBy).Assets.Environments, curTemplateInfo)
			}
		}
	}

	return nil
}

func findGitAssetsUsingCredentialId(ctx context.Context, patchingInfo map[string]response_patching.ResponsePatchingPtrs, singleId *string) *api.Error {
	gitAssets, err := dao.ListGitUsingGenericSecret(ctx, singleId)
	if err != nil {
		return err
	}
	for _, gitAsset := range gitAssets.Entries {
		curPatch, ok := patchingInfo[*gitAsset.Spec.CredentialId]
		if ok {
			curTemplateInfo := api.AssetDatasourceRef{
				Id:   gitAsset.Meta.Id,
				Name: gitAsset.Meta.Name,
				Kind: gitAsset.Meta.Kind,
			}

			if *curPatch.UsedBy == nil {
				*curPatch.UsedBy = NewAssetUsageInfo()
			}

			(*curPatch.UsedBy).Assets.Datasources = append((*curPatch.UsedBy).Assets.Datasources, curTemplateInfo)
		}
	}

	return nil
}

func NewAssetUsageInfo() *api.AssetUsageInfo {
	return &api.AssetUsageInfo{
		Workspaces: make([]api.WorkloadRefAndStatus, 0),
		Trainings:  make([]api.WorkloadRefAndStatus, 0),
		Inferences: make([]api.WorkloadRefAndStatus, 0),
		Templates:  make([]api.AssetRef, 0),
	}
}
