/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"strings"

	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/org_unit"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	clusterClient "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/clients/clusters-service"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
)

func CreateEnvironmentAsset(ctx context.Context, request api.EnvironmentCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.EnvironmentAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, err := createEnvironmentAssetInTransaction(ctx, tx, request, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_ENVIRONMENT).WithDetails(err)
	}
	committedTx = true

	return result, nil
}

func createEnvironmentAssetInTransaction(ctx context.Context, tx orm.DB, request api.EnvironmentCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.EnvironmentAsset, *api.Error) {
	id := uuid.New().String()

	assetMeta, err := createAssetMeta(ctx, tx, id, api.ASSETKIND_ENVIRONMENT, request.Meta, nameValidationQuery)
	if err != nil {
		if strings.Contains(err.Error(), alreadyExistsError) {
			return nil, database.ErrorWithCode(database.CREATE, api.ASSETKIND_ENVIRONMENT, http.StatusConflict).WithDetails(err)
		}
		return nil, database.Error(database.CREATE, api.ASSETKIND_ENVIRONMENT).WithDetails(err)
	}

	environmentAssetRec := dbmodel.EnvironmentAssetRec{
		AssetId:              id,
		EnvironmentAssetSpec: request.Spec,
	}
	if request.Spec.Connections == nil {
		environmentAssetRec.Connections = []api.Connection{}
	}

	_, err = tx.Create(environmentAssetRec)
	if err != nil {
		return nil, database.Error(database.CREATE, api.ASSETKIND_ENVIRONMENT).WithDetails(err)
	}

	return &api.EnvironmentAsset{
		Meta: *assetMeta,
		Spec: request.Spec,
	}, nil
}

func GetEnvironmentAssetById(ctx context.Context, id string) (*api.EnvironmentAsset, *api.Error) {
	var combinedRec dbmodel.EnvironmentCombinedRec
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_ENVIRONMENT,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, dbtables.TableNameAssetsEnvironment)
	if err != nil {
		return nil, err
	}

	scanRows, scanErr := result.Scan(&combinedRec)
	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_ENVIRONMENT).WithDetailsFromDbResult(scanRows, scanErr)
	}

	response := api.EnvironmentAsset{
		Meta: MergedAssetMetaOf(combinedRec),
		Spec: combinedRec.EnvironmentAssetSpec,
	}

	return &response, nil
}

// withClusterSubdomain returns db object with a filter environment that support the cluster routing method
func withClusterSubdomain(ctx context.Context, db orm.DB, queryParams *query_types.QueryParams) (orm.DB, error) {
	// filtering by subdomain, depends on the clusters routing method (host(subdomain) or path)
	// if query param target a specific cluster, the routing method is taken from the cluster
	// o.w. listing assets for all clusters, the routing method is taken from the common routing methods

	clusterId, err := clusterIdFromQueryParams(ctx, queryParams)
	if err != nil {
		return db, err
	}
	var subdomainRouting string
	if clusterId == nil {
		// there is no target cluster, get the common routing method
		subdomainRouting, err = getRoutingMethodByCommonClusters(ctx)
	} else {
		// there is a target cluster, get the cluster routing method
		subdomainRouting, err = getClusterRoutingMethod(ctx, *clusterId)
	}
	if err != nil {
		return db, err
	}

	// if there wasn't a common routing method, don't filter by routing method
	if subdomainRouting == "" {
		return db, nil
	}

	// add filter by routing method, empty string === agnostic
	return db.Where(fmt.Sprintf("%s IN (?, '', ?)", dbtables.EnvironmentColumnRoutingMethod),
		subdomainRouting,
		dbtables.EnvironmentColumnRoutingMethodAgnostic,
	), nil
}

func clusterIdFromQueryParams(ctx context.Context, queryParams *query_types.QueryParams) (*string, error) {
	// no queryParams no target cluster
	if queryParams == nil {
		return nil, nil
	}

	if queryParams.ClusterId != nil {
		return queryParams.ClusterId, nil
	}

	// When filtering by department or project, check the ancestor cluster
	if queryParams.DepartmentId != nil {
		resp, err := org_unit.Get().GetDepartmentAncestors(ctx, fmt.Sprintf("%v", *queryParams.DepartmentId))
		if err != nil {
			return nil, fmt.Errorf("fail to query department ancestors %v", err)
		}
		return resp.ClusterUuid.Get(), nil
	}
	if queryParams.ProjectId != nil {
		resp, err := org_unit.Get().GetProjectAncestors(ctx, fmt.Sprintf("%v", *queryParams.ProjectId))
		if err != nil {
			return nil, fmt.Errorf("fail to query department ancestors %v", err)
		}
		return resp.ClusterUuid.Get(), nil
	}
	return nil, nil
}

// getClusterRoutingMethod return the routing method of the cluster
func getClusterRoutingMethod(ctx context.Context, clusterId string) (string, error) {
	subdomain, err := isClusterSubdomainRouting(ctx, clusterId)
	if err != nil {
		return "", err
	}
	if subdomain {
		return dbtables.EnvironmentColumnRoutingMethodHostBase, nil
	}
	return dbtables.EnvironmentColumnRoutingMethodPathBase, nil
}

func isClusterSubdomainRouting(ctx context.Context, clusterId string) (bool, error) {
	clusterConfiguration, err := checker.NewClusterCapabilitiesChecker(clusterId).ClusterConfiguration(ctx)
	if err != nil {
		return false, err
	}
	return clusterConfiguration.SubdomainRouting, nil
}

// getRoutingMethodByCommonClusters return the routing method using the common clusters routing method
func getRoutingMethodByCommonClusters(ctx context.Context) (string, error) {

	client := clusterClient.GetClient()
	clustersCount, apiErr := client.GetClustersCount(ctx, "subdomain")
	if apiErr != nil {
		return "", fmt.Errorf("internal error, fail to reslove clusters routing method err: %v", apiErr)
	}
	pathRoutingFlag := false
	hostRoutingFlag := false
	unknownFlag := false
	for _, groupCount := range clustersCount {
		if groupCount.Key == "true" && groupCount.Count > 0 {
			hostRoutingFlag = true
			continue
		}
		if groupCount.Key == "false" && groupCount.Count > 0 {
			pathRoutingFlag = true
			continue
		}
		if groupCount.Key == "missing" && groupCount.Count > 0 {
			unknownFlag = true
			continue
		}
	}

	// if there are clusters with unknown routing method or there are mix of path and host routing clusters
	if (unknownFlag) || (pathRoutingFlag && hostRoutingFlag) {
		return "", nil
	}

	if pathRoutingFlag {
		return dbtables.EnvironmentColumnRoutingMethodPathBase, nil
	}
	if hostRoutingFlag {
		return dbtables.EnvironmentColumnRoutingMethodHostBase, nil
	}

	return "", fmt.Errorf("internal error, fail to reslove clusters routing method")
}

func ListEnvironmentAssets(ctx context.Context, queryParams *query_types.QueryParams) (*api.EnvironmentListResponse, *api.Error) {
	combinedList := make([]dbmodel.EnvironmentCombinedRec, 0)
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_ENVIRONMENT,
		},
	}

	clonePresetAssetsToTenant(ctx, auth.TenantIdOf(ctx))

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, queryParams, dbtables.TableNameAssetsEnvironment)
	if err != nil {
		return nil, err
	}

	var errSubdomain error
	result, errSubdomain = withClusterSubdomain(ctx, result, queryParams)
	if errSubdomain != nil {
		// Since the environment/cluster routing method has a small impact on filtering the assets, silence the failure to smooth the experience
		logMsg := fmt.Sprintf("failed to resolve the cluster subdomain (failure is silence). error: %v", errSubdomain)
		if queryParams != nil {
			logMsg = logMsg + fmt.Sprintf(" query params: %+v", *queryParams)
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("%s", logMsg)
	}

	_, scanErr := result.
		Order(&query.AssetMeta.Id).
		Scan(&combinedList)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list environment assets", scanErr)
	}

	response := api.EnvironmentListResponse{
		Entries: make([]api.EnvironmentAsset, 0),
	}

	for _, row := range combinedList {
		response.Entries = append(response.Entries, api.EnvironmentAsset{
			Meta: MergedAssetMetaOf(row),
			Spec: row.EnvironmentAssetSpec,
		})
	}

	return &response, nil
}

func UpdateEnvironmentAsset(ctx context.Context, id string, request api.EnvironmentUpdateRequest, nameValidationQuery *database.ParametrizedQuery) (*api.EnvironmentAsset, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, err := updateEnvironmentAssetInTransaction(ctx, tx, id, request, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_ENVIRONMENT).WithDetails(err)
	}
	committedTx = true

	NotifyAssetModified(ctx, id, result.Meta.Kind, ScopeInfoOfAssetMeta(&result.Meta))

	return result, nil
}

func updateEnvironmentAssetInTransaction(ctx context.Context, tx orm.DB, id string, request api.EnvironmentUpdateRequest, nameValidationQuery *database.ParametrizedQuery) (*api.EnvironmentAsset, *api.Error) {
	assetMeta, rowsAffected, apiErr, dbErr := updateAssetMeta(ctx, tx, id, request.Meta, nameValidationQuery, nil)
	if apiErr != nil {
		return nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_ENVIRONMENT).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	workloadSupportedTypes, wstErr := WorkloadSupportedTypesOf(tx, id)
	if wstErr != nil {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_ENVIRONMENT).WithDetails(wstErr)
	}

	recId := dbmodel.EnvironmentAssetRec{AssetId: id}
	environmentAssetUpdate := dbmodel.EnvironmentAssetUpdate{}

	if err := tx.Model(recId).Where(recId).First(&environmentAssetUpdate); err != nil {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_ENVIRONMENT).WithDetails(err)
	}

	environmentAssetUpdate.EnvironmentAssetSpec = request.Spec
	if environmentAssetUpdate.Connections == nil {
		environmentAssetUpdate.Connections = []api.Connection{}
	}

	saveRows, saveErr := tx.Where(recId).Save(&environmentAssetUpdate)
	if saveErr != nil {
		return nil, database.Error(database.UPDATE, api.ASSETKIND_ENVIRONMENT).WithDetailsFromDbResult(saveRows, saveErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, assetMeta.Id); policyErr != nil {
		return nil, policyErr
	}

	return &api.EnvironmentAsset{
		Meta: MergedAssetMetaAndWST(assetMeta, workloadSupportedTypes),
		Spec: request.Spec,
	}, nil
}

func DeleteEnvironmentAsset(ctx context.Context, assetMeta *dbmodel.AssetMeta, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	err := deleteEnvironmentAssetInTransaction(ctx, tx, assetMeta.Id, deletionType)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_ENVIRONMENT).WithDetails(err)
	}
	committedTx = true

	NotifyAssetDeleted(ctx, assetMeta.Id, assetMeta.Kind, ScopeInfoOfDbMeta(assetMeta))

	return nil
}

func deleteEnvironmentAssetInTransaction(ctx context.Context, tx orm.DB, id string, deletionType DeletionType) *api.Error {
	recId := dbmodel.EnvironmentAssetRec{AssetId: id}

	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, id, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_ENVIRONMENT).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == SoftDelete {
		return nil
	}

	delRows, delErr := tx.Where(&recId).Delete(&recId)
	if delErr != nil || delRows == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_ENVIRONMENT).WithDetailsFromDbResult(delRows, delErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, id); policyErr != nil {
		return policyErr
	}

	if err := deleteWorkloadAssetsByAssetId(ctx, tx, id); err != nil {
		return err
	}

	return nil
}

// ListEnvironmentAssetsUsingCredentialAsEnvVar returns a list of environment assets that are using a credential in the environment variables
func ListEnvironmentAssetsUsingCredentialAsEnvVar(ctx context.Context, credentialId *string) (*api.EnvironmentListResponse, *api.Error) {
	result, err := credentialInEnvarQuery(ctx, credentialId)
	if err != nil {
		return nil, err
	}

	combinedList := make([]dbmodel.EnvironmentCombinedRec, 0)
	kindFilter := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_ENVIRONMENT,
		},
	}

	_, scanErr := result.
		Order(&kindFilter.AssetMeta.Id).
		Scan(&combinedList)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list environment assets", scanErr)
	}

	response := api.EnvironmentListResponse{
		Entries: make([]api.EnvironmentAsset, 0),
	}

	for _, row := range combinedList {
		response.Entries = append(response.Entries, api.EnvironmentAsset{
			Meta: MergedAssetMetaOf(row),
			Spec: row.EnvironmentAssetSpec,
		})
	}

	return &response, nil
}

func countEnvironmentAssetsUsingCredentialAsEnvVar(ctx context.Context, credentialId string) (int64, *api.Error) {
	result, err := credentialInEnvarQuery(ctx, &credentialId)
	if err != nil {
		return 0, err
	}

	var count int64
	if countErr := result.Count(&count); countErr != nil {
		return 0, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list environment assets", countErr)
	}

	return count, nil
}

func credentialInEnvarQuery(ctx context.Context, credentialId *string) (orm.DB, *api.Error) {
	kindFilter := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_ENVIRONMENT,
		},
	}

	clonePresetAssetsToTenant(ctx, auth.TenantIdOf(ctx))

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), kindFilter, nil, dbtables.TableNameAssetsEnvironment)
	if err != nil {
		return nil, err
	}

	if credentialId == nil {
		result.Where(fmt.Sprintf("EXISTS (SELECT 1 FROM jsonb_array_elements(%s) AS elem WHERE elem->'credential'->>'assetId' IS NOT NULL)",
			dbtables.EnvironmentColumnEnvironmentVariables))
	} else {
		result.Where(fmt.Sprintf("EXISTS (SELECT 1 FROM jsonb_array_elements(%s) AS elem WHERE elem->'credential'->>'assetId' = ?)",
			dbtables.EnvironmentColumnEnvironmentVariables), credentialId)
	}

	return result, nil
}
