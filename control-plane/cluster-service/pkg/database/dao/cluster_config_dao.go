package dao

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm/clause"
)

func GetClusterConfigByClusterId(ctx context.Context, tenantId int32, clusterIdStr string) (*models.ClusterConfig, error) {
	tx := runaidb.DB().WithContext(ctx)

	clusterId, err := uuid.Parse(clusterIdStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse clusterId: %w", err)
	}

	cc := &models.ClusterConfig{}
	err = tx.Where(&models.ClusterConfig{
		ClusterID: clusterId,
		TenantID:  tenantId,
	}).Take(cc)

	// supporting the situation where during rolling update the cluster is created on the older version of the code,
	// after moving to the new version of the code the cluster config is not found

	if err != nil && errors.Is(err, orm.ErrRecordNotFound) {
		cc = models.NewClusterConfig(tenantId, clusterIdStr)
		_, err = tx.Create(cc)
	}
	return cc, err
}

func SetClusterConfig(ctx context.Context, tenantId int32, clusterId string, clusterConfig *models.ClusterConfigData) (*models.ClusterConfig, error) {
	config := models.NewClusterConfig(tenantId, clusterId)
	tx := runaidb.DB().WithContext(ctx).Begin().Clauses(clause.Locking{Strength: "SHARE"})
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
	_, err := tx.Where(&models.ClusterConfig{
		ClusterID: uuid.MustParse(clusterId),
		TenantID:  tenantId,
	}).FirstOrCreate(config)
	if err != nil {
		return nil, err
	}
	config.Config.SystemConfig = clusterConfig.SystemConfig
	config.Config.UserConfig = clusterConfig.UserConfig
	_, err = tx.Save(config)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	return config, nil
}

func SetClusterConfigStatus(ctx context.Context, tenantId int32, clusterId string, clusterConfigStatus *models.ClusterConfigStatus) (*models.ClusterConfig, error) {
	// Invalidate cache when config status changes (dependency status like LWS/Knative)
	defer invalidateClusterMetadataAndLog(ctx, tenantId, clusterId)

	config := models.NewClusterConfig(tenantId, clusterId)
	tx := runaidb.DB().WithContext(ctx).Begin().Clauses(clause.Locking{Strength: "SHARE"})
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
	_, err := tx.Where(&models.ClusterConfig{
		ClusterID: uuid.MustParse(clusterId),
		TenantID:  tenantId,
	}).FirstOrCreate(config)
	if err != nil {
		return nil, err
	}
	_, err = tx.Model(config).UpdateColumn("status", clusterConfigStatus)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	return config, nil
}

func SetClusterUserConfig(ctx context.Context, tenantId int32, clusterId string, clusterUserConfig *models.ClusterUserConfigData) (*models.ClusterConfig, error) {
	existingConfigModel := models.NewClusterConfig(tenantId, clusterId)
	tx := runaidb.DB().WithContext(ctx).Begin().Clauses(clause.Locking{Strength: "SHARE"})
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
	_, err := tx.Where(&models.ClusterConfig{
		ClusterID: uuid.MustParse(clusterId),
		TenantID:  tenantId,
	}).FirstOrCreate(existingConfigModel)
	if err != nil {
		return nil, err
	}
	existingConfigModel.Config.UserConfig = *clusterUserConfig
	if _, saveErr := tx.Save(existingConfigModel); saveErr != nil {
		return nil, saveErr
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true
	return existingConfigModel, nil
}

func DeleteClusterConfig(ctx context.Context, tenantId int32, clusterId string) error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	_, err := tx.Where(&models.ClusterConfig{
		ClusterID: uuid.MustParse(clusterId),
		TenantID:  tenantId,
	}).Delete(&models.ClusterConfig{})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete cluster config for cluster <%s> of tenant <%d>, error: %v", clusterId, tenantId, err)
		return err
	}
	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit delete cluster config for cluster <%s> of tenant <%d>, error: %v",
			clusterId, tenantId, err)
		return err
	}
	committed = true
	return nil
}
