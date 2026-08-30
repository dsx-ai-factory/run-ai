package dao

import (
	"context"
	"fmt"
	"net/http"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/clients/org_unit"

	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"

	"github.com/run-ai/runai/runai-application-packages/datastructs"

	"github.com/iancoleman/strcase"
	"github.com/run-ai/runai/control-plane/assets-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

// ListWorkloadAssets receives a list of asset ids contained in workload, and returns
// - an array of asset specs of the assets directly contained in the workload
// - a map from asset id to credential assets, holding the specs of all credentials assets referenced by the workload assets
func ListWorkloadAssets(assetIds []string, credentialsFromRunParams []string, db orm.DB) ([]dbmodel.AssetsContainedInWorkload, map[string]dbmodel.CredentialsAssetSpecs, *api.Error) {
	assets := make([]dbmodel.AssetsContainedInWorkload, 0, len(assetIds))

	model := dbmodel.AssetMetaRec{}

	result := db.
		Model(model).
		Select("*").
		Order(&model.Id).
		Where("id IN ?", assetIds)

	for _, tableName := range dbtables.TablesOfWorkloadAssets {
		result.Preload(strcase.ToCamel(tableName))
	}

	if findErr := result.Find(&assets); findErr != nil {
		return nil, nil, database.Error(database.LIST, "assets contained in workloads").WithDetails(findErr)
	}

	credentialIds := IdentifyCredentialAssets(assets)
	credentialIds = append(credentialIds, credentialsFromRunParams...)

	credentialsAssets, err := ListCredentialsAssets(credentialIds, db)
	if err != nil {
		return nil, nil, err
	}

	return assets, credentialsAssets, nil
}

// ValidateContainedAssets receives a set of asset ids from which the caller wants to create a template or a workload.
// it validates that the assets are compatible with the target project/department of the workload/template.
// project scoped workload/asset can contain assets from the project, its parent department and tenant.
// department scoped template can contain assets from the department and tenant.
// tenant scoped template can contain assets from the tenant and system.
// system scoped template can contain assets from system only.
func ValidateContainedAssets(ctx context.Context, workloadTypePtr *string, targetScope api.Scope,
	targetClusterId *string, targetProjectId *int32, targetDepartmentId *string, assetIds *api.AssetsIds,
) ([]string, *api.Error) {
	containedAssetIds, correlateAssetKindToIds, workloadVolumeAssets := assetIdsContainedInWorkload(assetIds)
	if len(containedAssetIds) == 0 {
		return nil, api.NewValidationError("workload must contain at least one asset")
	}

	assetsMetaById, dbError := ListContainedAssetsByIds(ctx, containedAssetIds, nil)
	if dbError != nil {
		return nil, dbError
	}

	//
	//   when the workload/template scope is project, we allow assets from the parent department,
	//   so we need to interrogate the backend for the parent department of the project
	//
	if targetProjectId != nil && targetDepartmentId == nil {
		projectIdStr := fmt.Sprintf("%d", *targetProjectId)
		ancestors, apiError := org_unit.Get().GetProjectAncestors(ctx, projectIdStr)
		if apiError != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("get ancestors failed: %s", apiError.Message)
			return nil, api.NewError(http.StatusServiceUnavailable, "failed to interrogate ancestors of the project", nil)
		}
		targetDepartmentId = convutils.Ref(ancestors.GetDepartmentId())
	}

	targetScopeLevel := levelOfScope(targetScope)

	for _, meta := range assetsMetaById.Sorted() {
		// system template can contain system scope assets
		// tenant template can contain system+tenant scope assets
		// dept template can contain system+tenant+dept assets
		// project template can contain assets from all levels
		assetScopeLevel := levelOfScope(meta.Scope)
		if assetScopeLevel > targetScopeLevel {
			return nil, api.NewError(http.StatusBadRequest, fmt.Sprintf("%s scope template cannot contain asset %s of scope %s", targetScope, meta.Name, meta.Scope), nil)
		}

		if meta.ProjectId != nil && targetProjectId != nil {
			if *meta.ProjectId != *targetProjectId {
				return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("asset %s is not associated with the requested %s or its ancestors", meta.Id, targetScope), nil)
			}
		}
		if meta.DepartmentId != nil && targetDepartmentId != nil {
			if *meta.DepartmentId != *targetDepartmentId {
				return nil, api.NewError(http.StatusNotFound, fmt.Sprintf("asset %s is not associated with the requested %s or its ancestors", meta.Id, targetScope), nil)
			}
		}
	}

	//
	//   validate that the kind provided for datasources is correct. for example caller can call
	//   create workload specifying pvc asset 123, while 123 is not a pvc, which should be rejected
	//
	for kind, idsToValidate := range correlateAssetKindToIds {
		validationError := assetsMetaById.ValidateKindOfAssets(kind, idsToValidate, workloadVolumeAssets, targetScope)
		if validationError != nil {
			return nil, validationError
		}
	}

	//
	// inference workload requires some special handling
	//
	if convutils.Deref(workloadTypePtr, "") == auth.WorkloadTypeInference {
		err := validateContainedAssetsForInferenceWorkload(ctx, targetClusterId, containedAssetIds, workloadVolumeAssets, &assetsMetaById)
		if err != nil {
			return nil, err
		}
	}

	return containedAssetIds, nil
}

func deleteWorkloadAssetsByAssetId(ctx context.Context, tx orm.DB, assetId string) *api.Error {
	_, err := tx.Table(dbtables.TableNameWorkloadsAssets).Where(fmt.Sprintf("%s = ?", dbtables.WorkloadsAssetsColumnAssetId), assetId).Delete(nil)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("failed to delete workload assets: %s", err.Error())
		return database.Error(database.DELETE, "workload assets").WithDetails(err)
	}

	return nil
}

// map scopes to their level in the org tree.
func levelOfScope(scope api.Scope) int {
	switch scope {
	case api.SCOPE_SYSTEM:
		return 0
	case api.SCOPE_TENANT:
		return 1
	case api.SCOPE_CLUSTER:
		return 2
	case api.SCOPE_DEPARTMENT:
		return 3
	default:
		return 4
	}
}

// validateContainedAssetsForInferenceWorkload validates that the assets contained in an inference workload are
// compatible with the inference workload requirements:
//   - workloadVolume assets are only supported for inference from a certain cluster version
//     and are not allowed to be ephemeral
//   - s3 is not supported, as our implementation requires privileged permissions that are not allowed by knative
func validateContainedAssetsForInferenceWorkload(ctx context.Context, clusterId *string, containedAssetIds []string, workloadVolumeAssets datastructs.Set[string], assetMetaById *dbmodel.AssetIdToMetaMap) *api.Error {
	// validate workload volume assets
	if len(workloadVolumeAssets) > 0 {
		// check if the cluster supports workload volumes
		workloadVolumesSupported := false
		if clusterId != nil { // if no clusterId is given, assume no support
			var err error
			clusterCapabilityChecker := checker.NewClusterCapabilitiesChecker(*clusterId)
			workloadVolumesSupported, err = clusterCapabilityChecker.ValidateClusterCapability(context.Background(), checker.MinVersionOfWorkloadVolumesForInference)
			if err != nil {
				return api.NewError(http.StatusInternalServerError, err.Error(), nil)
			}
		}

		if !workloadVolumesSupported {
			return api.NewError(http.StatusBadRequest, "inference workload cannot use workload volumes in this cluster version", nil)
		}

		// verify that the workload volumes are not ephemeral
		assetIds := make([]string, 0, len(workloadVolumeAssets))
		for assetId := range workloadVolumeAssets {
			assetIds = append(assetIds, assetId)
		}

		found, err := FoundEphemeralPvcsByIds(ctx, assetIds)
		if err != nil {
			return err
		}

		if found {
			return api.NewError(http.StatusBadRequest, "inference workload cannot use ephemeral workload volumes", nil)
		}
	}

	// verify no nfs, hostPath or s3 assets
	for _, assetId := range containedAssetIds {
		assetMeta := (*assetMetaById)[assetId]
		if assetMeta.Kind == api.ASSETKIND_S3 {
			return api.NewError(http.StatusBadRequest, fmt.Sprintf("inference workload cannot use data-source asset of type %s", assetMeta.Kind), nil)
		}
	}

	return nil
}
