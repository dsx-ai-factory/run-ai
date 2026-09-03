/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"github.com/google/uuid"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	ext_api "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/api"
	clusters_metadata_cache "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao/cache/clusters/metadata"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	querydbenrich "github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi/dbenrich"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

func GetClusterByIdAndTenantId(ctx context.Context, clusterId string, tenantId int32, verbosity ext_api.Verbosity) (*models.Cluster, *api.Error) {
	var cluster *models.Cluster
	var err error

	isMgmtTenant := runaictx.IsManagementTenant(ctx)

	getClusterFromDbFn := func() (*models.Cluster, error) {
		cluster := &models.Cluster{}
		tx := runaidb.DB().WithContext(ctx)
		if verbosity == ext_api.VerbosityMetadata {
			tx = tx.Select(models.BasicVerbosityClusterFields)
		}
		query := tx.
			Where("id = ?", clusterId)

		if !isMgmtTenant {
			query = query.Where("tenant_id = ?", tenantId)
		}

		if err := query.First(cluster); err != nil {
			return nil, err
		}
		return cluster, nil
	}

	switch verbosity {
	case ext_api.VerbosityMetadata:
		cluster, err = clusters_metadata_cache.GetOrSetClusterMetadata(ctx, tenantId, clusterId, func() (*models.Cluster, error) {
			return getClusterFromDbFn()
		})
	case ext_api.VerbosityFull:
		cluster, err = getClusterFromDbFn()
	}

	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get cluster <%s> for tenant <%d>", clusterId, tenantId), nil).WithDetails(err)
	}
	return cluster, nil
}

func GetClusterDomainById(ctx context.Context, clusterId string) (string, *api.Error) {
	var cluster models.Cluster
	tx := runaidb.DB().WithContext(ctx)

	if err := tx.Select("domain").Where("id = ?", clusterId).First(&cluster); err != nil {
		return "", api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get cluster domain for cluster <%s>", clusterId), nil).WithDetails(err)
	}

	return cluster.Domain, nil
}

// GetClusterCountGroupBySubdomain returns the count of clusters grouped by subdomain
func GetClusterCountGroupBySubdomain(ctx context.Context, tenantId int32, permittedScopes openapi.PermittedScopes) (*models.ClustersCountSubdomainStatus, *api.Error) {
	var clustersCount models.ClustersCountSubdomainStatus

	_, err := runaidb.DB().WithContext(ctx).Model(&models.ClustersCountSubdomainStatus{}).
		Where("tenant_id = ?", tenantId).
		Where(GetPermittedScopesFilter(permittedScopes, "id")).
		Where("delete_requested_at IS NULL").
		Select(
			"COUNT(*) FILTER (WHERE status->'config'->>'subdomainEnabled' = 'true') AS enabled_count",
			"COUNT(*) FILTER (WHERE status->'config'->>'subdomainEnabled' = 'false') AS disabled_count",
			"COUNT(*) FILTER (WHERE status IS NULL OR NOT (status ? 'config') OR NOT (status->'config' ? 'subdomainEnabled')) AS missing_count",
		).
		Scan(&clustersCount)
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get clusters count for tenant <%d>", tenantId), nil).WithDetails(err)
	}

	return &clustersCount, nil
}

func GetClusters(ctx context.Context, tenantId int32, verbosity ext_api.Verbosity, permittedScopes openapi.PermittedScopes, tenantIdsFilter []string, includeRequestedForDelete bool, includeDeleted bool,
	offset int32, limit int32, filters []queryapi.Filter,
) (*[]models.Cluster, *api.Error) {
	var err error
	isMgmtTenant := runaictx.IsManagementTenant(ctx)

	getClustersFromDbFn := func() (*[]models.Cluster, error) {
		clusters := &[]models.Cluster{}
		query := runaidb.DB().WithContext(ctx)
		if includeDeleted {
			query = query.Unscoped()
		}

		if !isMgmtTenant {
			query = query.Where("tenant_id = ?", tenantId)
		}

		if limit > 0 {
			applyPaginationParam(query, offset, limit)
		}

		query = query.Where(GetPermittedScopesFilter(permittedScopes, "id"))

		if !includeRequestedForDelete {
			query = query.Where("delete_requested_at IS NULL")
		}

		if len(tenantIdsFilter) > 0 {
			query = query.Where("tenant_id IN ?", tenantIdsFilter)
		}

		query = query.WithEnricher(querydbenrich.Filters(filters))

		if verbosity == ext_api.VerbosityMetadata {
			query = query.Select(models.BasicVerbosityClusterFields)
		}

		if err := query.Find(clusters); err != nil {
			return nil, err
		}
		return clusters, nil
	}

	var clusters *[]models.Cluster
	switch verbosity {
	case ext_api.VerbosityMetadata:
		if includeDeleted {
			return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get clusters for tenant <%d> cannot use includeDeleted with metadata verbosity", tenantId), nil).WithDetails(err)
		}
		clusters, err = clusters_metadata_cache.GetOrSetAllClustersMetadata(ctx, tenantId, func() (*[]models.Cluster, error) {
			return getClustersFromDbFn()
		})
	case ext_api.VerbosityFull:
		clusters, err = getClustersFromDbFn()
	}

	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get clusters for tenant <%d>", tenantId), nil).WithDetails(err)
	}
	return clusters, nil
}

func CreateCluster(ctx context.Context, cluster *models.Cluster) (*models.Cluster, *api.Error) {
	apiErr := checkClusterNameUniqueness(ctx, cluster)
	if apiErr != nil {
		return nil, apiErr
	}
	defer invalidateAllClustersMetadataAndLog(ctx, cluster.TenantId)

	dbErr := api.NewError(http.StatusInternalServerError,
		fmt.Sprintf("Failed to create cluster <%s>", cluster.ID), nil)

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	// Cluster and its config are created in a single transaction
	if _, err := tx.Create(cluster); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create cluster <%s>, error: %v", cluster.ID, err)
		return nil, dbErr.WithDetails(err)
	}
	newClusterConfig := models.NewClusterConfig(cluster.TenantId, cluster.ID.String())
	if _, err := tx.Create(newClusterConfig); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create cluster config for cluster <%s>, error: %v", cluster.ID, err)
		return nil, dbErr.WithDetails(err)
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit create cluster <%s>, error: %v",
			cluster.ID, err)
		return nil, dbErr.WithDetails(err)
	}
	committed = true
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Succeed to commit create cluster %s", cluster.ID)
	return cluster, nil
}

func checkClusterNameUniqueness(ctx context.Context, cluster *models.Cluster) *api.Error {
	count := int64(0)

	err := runaidb.DB().WithContext(ctx).Model(cluster).
		Where("tenant_id = ?", cluster.TenantId).
		Where("name = ?", cluster.Name).
		Where("delete_requested_at IS NULL").
		Count(&count)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Error while validating cluster details, error: %v", err)
		return api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to check if cluster creation valid for tenant <%d>", cluster.TenantId), err)
	}
	if count > 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Cluster with the given name: <%s> already exist for tenant: <%d>", cluster.Name, cluster.TenantId)
		msg := fmt.Sprintf("Cluster with the given name: <%s> already exist for tenant: <%d>", cluster.Name, cluster.TenantId)
		return api.NewError(http.StatusConflict, msg, errors.New(msg))
	}
	return nil
}

func DeleteCluster(ctx context.Context, clusterId string, tenantId int32) *api.Error {
	defer invalidateAllClustersMetadataAndLog(ctx, tenantId)

	dbErr := api.NewError(http.StatusInternalServerError,
		fmt.Sprintf("Failed to delete cluster <%s> of tenant <%d>", clusterId, tenantId), nil)

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	rowsAffected, err := tx.
		Where("id = ?", clusterId).
		Where("tenant_id = ?", tenantId).
		Delete(&models.Cluster{})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete cluster <%s> of tenant <%d>, error: %v", clusterId, tenantId, err)
		return dbErr.WithDetailsFromDbResult(rowsAffected, err)
	}
	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit delete cluster <%s> of tenant <%d>, error: %v",
			clusterId, tenantId, err)
		return dbErr.WithDetails(err)
	}
	committed = true
	return nil
}

func HardDeleteAllTenantClusters(ctx context.Context, tenantId int32) error {
	permittedScopes := openapi.PermittedScopes{
		Tenants: []string{strconv.Itoa(int(tenantId))},
	}
	clusters, err := GetClusters(ctx, tenantId, ext_api.VerbosityMetadata, permittedScopes, nil, true, false, 0, 0, nil)
	if err != nil {
		return fmt.Errorf("failed to load tenant clusters: %v", err)
	}
	defer invalidateAllTenantClustersMetaDataAndLog(ctx, tenantId, clusters)

	tx := runaidb.DB().WithContext(ctx).Begin().Unscoped()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	if _, dbErr := tx.Where("tenant_id = ?", tenantId).Delete(&models.Cluster{}); dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete all tenant clusters, error: %v", dbErr)
		return dbErr
	}
	if dbErr := tx.Commit(); dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit delete all tenant clusters, error: %v", dbErr)
		return dbErr
	}
	committed = true
	return nil
}

func UpdateClusterName(ctx context.Context, cluster *models.Cluster) *api.Error {
	apiErr := checkClusterNameUniqueness(ctx, cluster)
	if apiErr != nil {
		return apiErr
	}
	defer invalidateAllClustersMetadataAndLog(ctx, cluster.TenantId)
	defer invalidateClusterMetadataAndLog(ctx, cluster.TenantId, cluster.ID.String())

	dbErr := api.NewError(http.StatusInternalServerError,
		fmt.Sprintf("Failed to update cluster <%s> of tenant <%d>", cluster.ID, cluster.TenantId), nil)

	rowsAffected, err := runaidb.DB().WithContext(ctx).
		Table("clusters").
		Where("tenant_id = ?", cluster.TenantId).
		Where("id = ?", cluster.ID).
		UpdateColumn("name", cluster.Name)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit update name for cluster <%s> of tenant <%d>, error: %v",
			cluster.ID, cluster.TenantId, err)
		return dbErr.WithDetailsFromDbResult(rowsAffected, err)
	}
	return nil
}

func UpdateClusterLiveness(ctx context.Context, clusterId string, tenantId int32, liveness time.Time) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError,
		fmt.Sprintf("Failed to update liveness of cluster <%s> of tenant <%d>", clusterId, tenantId), nil)

	rowsAffected, err := runaidb.DB().WithContext(ctx).
		Table("clusters").
		Where("tenant_id = ?", tenantId).
		Where("id = ?", clusterId).
		UpdateColumn("last_liveness", liveness)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit update last_liveness for cluster <%s> of tenant <%d>, error: %v",
			clusterId, tenantId, err)
		return dbErr.WithDetailsFromDbResult(rowsAffected, err)
	}
	return nil
}

func UpdateClusterInfo(ctx context.Context, c *models.Cluster) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError,
		fmt.Sprintf("Failed to update cluster-info for cluster <%s> of tenant <%d>", c.ID, c.TenantId), nil)

	emptyGetter := func() (*models.Cluster, error) {
		return nil, nil
	}
	data, _ := clusters_metadata_cache.GetOrSetClusterMetadata(ctx, c.TenantId, c.ID.String(), emptyGetter)
	if data != nil {
		// Invalidate cache if version or status changed
		versionChanged := data.Version != c.Version
		statusChanged := !reflect.DeepEqual(data.Status, c.Status)

		if versionChanged || statusChanged {
			defer invalidateAllClustersMetadataAndLog(ctx, c.TenantId)
			defer invalidateClusterMetadataAndLog(ctx, c.TenantId, c.ID.String())
		}
	}

	rowsAffected, updateErr := runaidb.DB().WithContext(ctx).
		Table("clusters").
		Where("tenant_id = ?", c.TenantId).
		Where("id = ?", c.ID).
		UpdateColumns(map[string]interface{}{
			"domain":  c.Domain,
			"version": c.Version,
			"status":  c.Status,
		})

	err := InsertClusterHistory(ctx, c)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create cluster history %s, error: %v",
			c.ID, err)
	}

	if updateErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit update cluster-info for cluster <%s> of tenant <%d>, error: %v", c.ID, c.TenantId, updateErr)
		return dbErr.WithDetailsFromDbResult(rowsAffected, updateErr)
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Committed update cluster-info for cluster <%s> of tenant <%d>", c.ID, c.TenantId)
	return nil
}

func GetClusterVersions(ctx context.Context, tenantId int32, clusterIds []string) ([]string, *api.Error) {
	versions := []string{}
	isMgmtTenant := runaictx.IsManagementTenant(ctx)

	query := runaidb.DB().WithContext(ctx).
		Table(models.ClusterTable).
		Select("version").
		Where("deleted_at IS NULL")

	if !isMgmtTenant {
		query = query.Where("tenant_id = ?", tenantId)
	}

	if clusterIds != nil {
		query = query.
			Where("id in ?", clusterIds)
	}

	_, err := query.Scan(&versions)
	if err != nil {
		dbErr := api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to get cluster versions for tenant <%d> and clusters %v. Error: %v", tenantId, clusterIds, err), nil)
		return versions, dbErr
	}
	return versions, nil
}

// ErrClusterNotFound is returned by GetClusterVersion when no live cluster
// row matches the requested id+tenant. Detect with errors.Is.
var ErrClusterNotFound = errors.New("cluster not found")

func GetClusterVersion(ctx context.Context, tenantId int32, clusterId uuid.UUID) (string, error) {
	vers, apiErr := GetClusterVersions(ctx, tenantId, []string{clusterId.String()})
	if apiErr != nil {
		return "", errors.New(apiErr.Message)
	}
	if len(vers) < 1 {
		return "", ErrClusterNotFound
	}
	return vers[0], nil
}

func UpdateClusterDeleteRequestTimestamp(ctx context.Context, clusterId string, tenantId int32, deleteRequestTimestamp time.Time) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError,
		fmt.Sprintf("Failed to update delete request timestamp of cluster <%s> of tenant <%d>", clusterId, tenantId), nil)
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
	rowsAffected, err := tx.Table("clusters").
		Where("tenant_id = ?", tenantId).
		Where("id = ?", clusterId).
		UpdateColumn("delete_requested_at", deleteRequestTimestamp)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit update delete_requested_at for cluster <%s> of tenant <%d>, error: %v",
			clusterId, tenantId, err)
		return dbErr.WithDetailsFromDbResult(rowsAffected, err)
	}
	clusterConfigEntity, clusterConfigErr := GetClusterConfigByClusterId(ctx, tenantId, clusterId)
	if clusterConfigErr != nil {
		return api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to get cluster config for tenantId: [%d], cluster uuid: [%s]", tenantId, clusterId), clusterConfigErr)
	}
	clusterConfig := &clusterConfigEntity.Config
	clusterConfig.SystemConfig.Deactivated = true
	_, clusterConfigErr = SetClusterConfig(ctx, tenantId, clusterId, clusterConfig)
	if clusterConfigErr != nil {
		return api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Failed to update cluster config, with cluster deactivated flag=true for tenantId: [%d], cluster uuid: [%s]", tenantId, clusterId), clusterConfigErr)
	}
	if commitErr := tx.Commit(); commitErr != nil {
		return dbErr.WithDetailsFromDbResult(rowsAffected, commitErr)
	}
	committed = true
	return nil
}

func GetPendingDeletionClusters(ctx context.Context) ([]models.Cluster, *api.Error) {
	var clusters []models.Cluster

	err := runaidb.DB().WithContext(ctx).
		Select("id", "tenant_id", "delete_requested_at").
		Where("deleted_at IS NULL").
		Where("delete_requested_at IS NOT NULL").
		Find(&clusters)
	if err != nil {
		apiErr := api.Error{}
		return nil, apiErr.WithDetails(err)
	}
	return clusters, nil
}
