package dao

import (
	"context"
	"fmt"
	"sort"
	"time"

	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/runai-application-packages/queryapi"
	querydbenrich "github.com/run-ai/runai/runai-application-packages/queryapi/dbenrich"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm/clause"
	"gorm.io/gorm"
)

const (
	CpuQuotaSystemSettingFlag = "project.quota.cpu_resources"
	WeightsSystemSettingFlag  = "project.quota.manual_over_quota"
)

var nodepoolMinimalFields = []string{
	"id", models.NameColumn, models.ClusterIdColumn, models.PhaseColumn,
	models.StatusGpuNetworkAccelerationDetectedColumn, models.AdjustableWeightsColumn,
	models.RanksColumn, models.QuotaColumn, models.CpuQuotaColumn,
}

// GetNodepoolMinimalFields returns the list of basic metadata fields for nodepools
func GetNodepoolMinimalFields() []string {
	return nodepoolMinimalFields
}

func CreateNodepool(ctx context.Context, nodepool models.Nodepool) (models.Nodepool, error) {
	if _, err := runaidb.DB().WithContext(ctx).Create(&nodepool); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create nodepool <%s> in cluster <%s>, error: %v",
			nodepool.Name, nodepool.ClusterId, err)
		return models.Nodepool{}, err
	}

	return nodepool, nil
}

// CreateNodepools Used by the cluster-migrator for bulk creations
func CreateNodepools(ctx context.Context, nodepools []models.Nodepool) error {
	if len(nodepools) == 0 {
		return nil
	}

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	_, err := tx.
		Clauses(clause.OnConflict{
			UpdateAll: true,
		}).
		CreateInBatches(nodepools, 20)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed create nodepools: %v", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed commit nodepools creation: %v", err)
		return err
	}
	committed = true
	return nil
}

func UpdateNodepool(ctx context.Context, nodepool models.Nodepool) (models.Nodepool, error) {
	_, err := runaidb.DB().WithContext(ctx).
		Where("tenant_id = ? AND id = ?", nodepool.TenantId, nodepool.ID).
		Save(&nodepool)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update nodepool <%s> in cluster <%s>, error: %v",
			nodepool.Name, nodepool.ClusterId, err)
		return models.Nodepool{}, err
	}

	return nodepool, nil
}

func SyncNodepools(ctx context.Context, nodepools []models.Nodepool, tenantId int32, clusterId string, deleteMissing bool) ([]string, []string, error) {
	sort.Slice(nodepools, func(i, j int) bool {
		return nodepools[i].Name < nodepools[j].Name
	})
	const savingPoint = "sp"

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
	tx.SavePoint(savingPoint)

	// allow to actually delete (update phase to DELETED) if and only if the nodepool is already in DELETING phase.
	// AND: don't allow to update to DELETING - unless it was already in DELETING phase.
	// ((requestedPhase != Deleted && originalPhase != Deleting) OR (requestedPhase == Deleted && originalPhase == Deleting))
	// AND (requestedPhase != Deleting || originalPhase == Deleting)

	syncedNodepools := []string{}
	failedNodepools := []string{}
	for _, nodepool := range nodepools {
		_, err := tx.
			Table(models.NodepoolTable).
			Where("tenant_id = ? AND cluster_id = ? AND name = ? AND deleted_at IS NULL", tenantId, clusterId, nodepool.Name).
			Where("((? AND phase <> ?) OR (? AND phase = ?)) AND (? OR phase = ?)",
				nodepool.Phase != api.NODEPOOLPHASE_DELETED, api.NODEPOOLPHASE_DELETING,
				nodepool.Phase == api.NODEPOOLPHASE_DELETED, api.NODEPOOLPHASE_DELETING,
				nodepool.Phase != api.NODEPOOLPHASE_DELETING, api.NODEPOOLPHASE_DELETING).
			UpdateColumns(
				map[string]interface{}{
					models.PhaseColumn:                            nodepool.Phase,
					models.PhaseMsgColumn:                         nodepool.PhaseMessage,
					models.StatusLabelKeyColumn:                   nodepool.NodepoolStatus.LabelKey,
					models.StatusLabelValueColumn:                 nodepool.NodepoolStatus.LabelValue,
					models.StatusGpuPlacementStrategyColumn:       nodepool.NodepoolStatus.GpuPlacementStrategy,
					models.StatusGpuDevicePlacementStrategyColumn: nodepool.NodepoolStatus.GpuDevicePlacementStrategy,
					models.StatusCpuPlacementStrategyColumn:       nodepool.NodepoolStatus.CpuPlacementStrategy,
					models.StatusNodesColumn:                      nodepool.Nodes,
					models.StatusGPUNetworkAccelerationDetected:   nodepool.GPUNetworkAccelerationDetected,
					models.StatusMinGuaranteedRuntimeColumn:       nodepool.NodepoolStatus.MinGuaranteedRuntime,
					models.StatusConditions:                       nodepool.Conditions,
					models.DeletedAtColumn: gorm.DeletedAt{
						Time:  time.Now(),
						Valid: nodepool.Phase == api.NODEPOOLPHASE_DELETED,
					},
				})

		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync nodepool %s in cluster %s. error: %v", nodepool.Name, clusterId, err)
			tx.RollbackTo(savingPoint) // Rollback to last saved point
			failedNodepools = append(failedNodepools, nodepool.Name)
		} else {
			tx.SavePoint(savingPoint)
			syncedNodepools = append(syncedNodepools, nodepool.Name)
		}
	}

	if deleteMissing && (len(syncedNodepools) != 0 || len(failedNodepools) != 0) {
		deleteQuery := tx.Table(models.NodepoolTable).
			Where("tenant_id = ? AND cluster_id = ?", tenantId, clusterId)
		_, err := deleteQuery.
			Where("is_default = false").
			Where("phase = ?", api.NODEPOOLPHASE_DELETING).
			Not("name IN ?", append(syncedNodepools, failedNodepools...)).
			UpdateColumns(
				map[string]interface{}{
					models.PhaseColumn:    api.NODEPOOLPHASE_DELETED,
					models.PhaseMsgColumn: "",
					models.DeletedAtColumn: gorm.DeletedAt{
						Time:  time.Now(),
						Valid: true,
					},
				})
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete missing nodepools in sync request of cluster %s. error: %v", clusterId, err)
			return nil, nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit the nodepools sync: %v", err)
		return nil, nil, err
	}
	committed = true

	return syncedNodepools, failedNodepools, nil
}

func GetNodepoolById(ctx context.Context, tenantId int32, nodepoolId string) (models.Nodepool, error) {
	nodepool := models.Nodepool{}
	query := runaidb.DB().WithContext(ctx)

	query = query.
		Where("tenant_id = ?", tenantId).
		Where("id = ?", nodepoolId)

	if err := query.First(&nodepool); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed get nodepool with id <%s>, error: %v",
			nodepoolId, err)
		return models.Nodepool{}, err
	}
	return nodepool, nil
}

func GetNodepools(ctx context.Context, tenantId int32, permittedScopes openapi.PermittedScopes, offset int32, limit int32, sortOrder string, sortBy string, filterBy []queryapi.Filter) ([]models.Nodepool, error) {
	tx := runaidb.DB().WithContext(ctx)

	nodepools := []models.Nodepool{}

	query := tx.
		Where("tenant_id = ?", tenantId).
		Where("phase <> ?", api.NODEPOOLPHASE_DELETED).
		Where(GetPermittedScopesFilter(permittedScopes, "cluster_id"))

	query = query.WithEnricher(querydbenrich.Filters(filterBy))
	query = applySortingParam(query, sortBy, sortOrder)
	query = applyPaginationParam(query, offset, limit)

	if err := query.Find(&nodepools); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list nodepools, err: %v", err)
		return []models.Nodepool{}, err
	}

	return nodepools, nil
}

// GetNodepoolsWithFields retrieves nodepools with specified fields.
func GetNodepoolsWithFields(
	ctx context.Context,
	tenantId int32,
	selectFields []string,
	offset int32,
	limit int32,
	sortOrder string,
	sortBy string,
	filterBy []queryapi.Filter,
	permittedScopes openapi.PermittedScopes,
) ([]models.Nodepool, error) {
	tx := runaidb.DB().WithContext(ctx)

	nodepools := []models.Nodepool{}

	// Select only the specified fields to optimize query performance
	query := tx.
		Select(selectFields).
		Where("tenant_id = ?", tenantId).
		Where("phase <> ?", api.NODEPOOLPHASE_DELETED).
		Where(GetPermittedScopesFilter(permittedScopes, "cluster_id"))

	query = query.WithEnricher(querydbenrich.Filters(filterBy))
	query = applySortingParam(query, sortBy, sortOrder)
	query = applyPaginationParam(query, offset, limit)

	if err := query.Find(&nodepools); err != nil {
		return []models.Nodepool{}, err
	}

	return nodepools, nil
}

// GetNodepoolMinimalById retrieves a nodepool by ID with specified fields.
func GetNodepoolMinimalById(ctx context.Context, tenantId int32, selectFields []string, nodepoolId string, permittedScopes openapi.PermittedScopes) (models.Nodepool, error) {
	nodepool := models.Nodepool{}
	query := runaidb.DB().WithContext(ctx)

	// Select only the specified fields to optimize query performance
	query = query.
		Select(selectFields).
		Where("tenant_id = ?", tenantId).
		Where("id = ?", nodepoolId).
		Where("phase <> ?", api.NODEPOOLPHASE_DELETED).
		Where(GetPermittedScopesFilter(permittedScopes, "cluster_id"))

	if err := query.First(&nodepool); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get nodepool metadata with id <%s>, error: %v",
			nodepoolId, err)
		return models.Nodepool{}, err
	}
	return nodepool, nil
}

func GetNodepoolsForTenants(ctx context.Context, tenantIds []string, permittedScopes *openapi.PermittedScopes) ([]models.Nodepool, error) {
	db := runaidb.DB().WithContext(ctx)

	nodepools := []models.Nodepool{}

	query := db.
		Where("tenant_id in ?", tenantIds).
		Where("phase <> ?", api.NODEPOOLPHASE_DELETED)

	if permittedScopes != nil {
		query = query.
			Where(GetPermittedScopesFilter(*permittedScopes, "cluster_id"))
	}

	if err := query.Order("id").Find(&nodepools); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list nodepools, err: %v", err)
		return []models.Nodepool{}, err
	}

	return nodepools, nil
}

func GetNodepoolsForCluster(ctx context.Context, tenantId string, clusterId string, permittedScopes *openapi.PermittedScopes) ([]models.Nodepool, error) {
	db := runaidb.DB().WithContext(ctx)

	nodepools := []models.Nodepool{}

	query := db.
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterId).
		Where("phase <> ?", api.NODEPOOLPHASE_DELETED)

	if permittedScopes != nil {
		query = query.
			Where(GetPermittedScopesFilter(*permittedScopes, "cluster_id"))
	}

	if err := query.Order("id").Find(&nodepools); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list nodepools, err: %v", err)
		return []models.Nodepool{}, err
	}

	return nodepools, nil
}

func CountNodepools(ctx context.Context, tenantId int32, permittedScopes openapi.PermittedScopes, filterBy []queryapi.Filter) (int64, error) {
	tx := runaidb.DB().WithContext(ctx)

	query := tx.
		Table(models.NodepoolTable).
		Where("tenant_id = ?", tenantId).
		Where("phase <> ?", api.NODEPOOLPHASE_DELETED).
		Where(GetPermittedScopesFilter(permittedScopes, "cluster_id"))

	query = query.WithEnricher(querydbenrich.Filters(filterBy))

	var count int64

	if err := query.Count(&count); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count nodepools: %v", err)
		return 0, err
	}

	return count, nil
}

func DeleteNodepool(ctx context.Context, tenantId int32, subjectId string, nodepoolId string) error {
	updateMap := map[string]interface{}{
		models.PhaseColumn:     api.NODEPOOLPHASE_DELETED,
		models.PhaseMsgColumn:  "",
		models.DeletedByColumn: subjectId,
		models.DeletedAtColumn: gorm.DeletedAt{
			Time:  time.Now(),
			Valid: true,
		},
	}

	_, err := runaidb.DB().WithContext(ctx).
		Table(models.NodepoolTable).
		Where("tenant_id = ?", tenantId).
		Where("id = ?", nodepoolId).
		UpdateColumns(updateMap)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete nodepool <%s>, error: %v",
			nodepoolId, err)
		return err
	}

	return nil
}

func SetNodepoolPhase(ctx context.Context, tenantId int32, subjectId string, nodepoolId string, phase api.NodepoolPhase) error {
	updateMap := map[string]interface{}{
		models.PhaseColumn:     phase,
		models.UpdatedByColumn: subjectId,
	}

	if phase == api.NODEPOOLPHASE_DELETING {
		updateMap[models.DeletedByColumn] = subjectId
	}

	if phase == api.NODEPOOLPHASE_DELETED {
		updateMap[models.DeletedAtColumn] = gorm.DeletedAt{
			Time:  time.Now(),
			Valid: true,
		}
	}

	_, err := runaidb.DB().WithContext(ctx).
		Table(models.NodepoolTable).
		Where("tenant_id = ?", tenantId).
		Where("id = ?", nodepoolId).
		UpdateColumns(updateMap)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to set nodepool <%s> phase to %s. err: %v", nodepoolId, phase, err)
		return err
	}

	return nil
}

func DeleteClusterNodepools(ctx context.Context, tenantId int32, subjectId string, clusterId string) error {
	updateMap := map[string]interface{}{
		models.PhaseColumn:     api.NODEPOOLPHASE_DELETED,
		models.PhaseMsgColumn:  "",
		models.DeletedByColumn: subjectId,
		models.DeletedAtColumn: gorm.DeletedAt{
			Time:  time.Now(),
			Valid: true,
		},
	}

	_, err := runaidb.DB().WithContext(ctx).
		Table(models.NodepoolTable).
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterId).
		Where("deleted_at IS NULL").
		UpdateColumns(updateMap)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete cluster <%s> nodepools , error: %v",
			clusterId, err)
		return err
	}

	return nil
}

// UpdateSystemSettingsOnTenantNodepools
// only supports CpuQuotaSystemSettingFlag and WeightsSystemSettingFlag
func UpdateSystemSettingsOnTenantNodepools(ctx context.Context, tenantId int32, settingsToSet map[string]bool) error {
	updateMap := map[string]interface{}{
		"updated_at": time.Now(),
	}

	for settingName, enabled := range settingsToSet {
		switch settingName {
		case CpuQuotaSystemSettingFlag:
			if enabled {
				// allow_cpu_quota=true is only allowed if allow_quota=true
				updateMap[models.CpuQuotaColumn] = gorm.Expr(fmt.Sprintf("CASE WHEN %s = true THEN true ELSE false END", models.QuotaColumn))
			} else {
				updateMap[models.CpuQuotaColumn] = false
			}
		case WeightsSystemSettingFlag:
			updateMap[models.AdjustableWeightsColumn] = enabled
		default:
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Unknown setting enum value <%s>", settingName)
			return nil
		}
	}

	_, err := runaidb.DB().WithContext(ctx).
		Table(models.NodepoolTable).
		Where("tenant_id = ?", tenantId).
		Where("deleted_at IS NULL").
		UpdateColumns(updateMap)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update system settings <%v> on tenant <%d> nodepools, error: %v",
			settingsToSet, tenantId, err)
		return err
	}

	return nil
}
