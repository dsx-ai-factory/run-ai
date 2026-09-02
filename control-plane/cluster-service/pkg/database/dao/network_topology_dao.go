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
	"time"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/adapters"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	querydbenrich "github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi/dbenrich"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

const (
	NetworkTopologiesRunAiSource   = "Run.ai"
	NetworkTopologiesClusterSource = "Cluster"
)

func SyncNetworkTopologies(ctx context.Context, tenantId int32, subjectId string, clusterUuid uuid.UUID, request api.NetworkTopologySyncRequest) *api.Error {
	if request.FullSync {
		return fullSyncNetworkTopologies(ctx, tenantId, subjectId, clusterUuid, request)
	}

	return syncActionNetworkTopologies(ctx, tenantId, subjectId, clusterUuid, request)
}

func DeleteNetworkTopology(ctx context.Context, tenantId int32, clusterUuid uuid.UUID, networkTopologyId string) *api.Error {
	tx := runaidb.DB().WithContext(ctx)
	return deleteNetworkTopology(ctx, tx, tenantId, clusterUuid, &networkTopologyId, nil)
}

func CreateNetworkTopology(ctx context.Context, networkTopology models.NetworkTopology) (*models.NetworkTopology, *api.Error) {
	_, err := runaidb.DB().WithContext(ctx).
		Table(models.NetworkTopologyTable).
		Create(&networkTopology)
	if err != nil {
		errorMessagePrefix := fmt.Sprintf("Failed to create network topology <%s> in cluster <%s>", networkTopology.Name, networkTopology.ClusterID)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errorMessagePrefix, err)
		apiError := api.NewError(http.StatusInternalServerError, errorMessagePrefix, err)
		if apiError.Code == http.StatusConflict {
			apiError.Message = fmt.Sprintf("%s. The network topology name must be unique.", errorMessagePrefix)
		}
		return nil, apiError
	}

	return &networkTopology, nil
}

func GetNetworkTopologiesView(ctx context.Context, tenantId int32, permittedScopes *openapi.PermittedScopes, filterBy []queryapi.Filter) ([]models.NetworkTopologyView, *api.Error) {
	query := runaidb.DB().WithContext(ctx).
		Model(&models.NetworkTopologyView{}).
		Where("tenant_id = ?", tenantId)

	if permittedScopes != nil {
		query = query.
			Where(GetPermittedScopesFilter(*permittedScopes, "cluster_id"))
	}

	query = query.WithEnricher(querydbenrich.Filters(filterBy))
	query = query.Order("id DESC")

	var networkTopologiesView []models.NetworkTopologyView
	if err := query.Find(&networkTopologiesView); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get network topologies view for tenant <%d>, error: %v", tenantId, err)
		return nil, api.NewError(http.StatusInternalServerError, "Failed to get network topologies view", err)
	}
	return networkTopologiesView, nil
}

func GetNetworkTopologyById(ctx context.Context, tenantId int32, networkTopologyId string) (*models.NetworkTopology, *api.Error) {
	var networkTopology models.NetworkTopology
	err := runaidb.DB().WithContext(ctx).
		Where("tenant_id = ?", tenantId).
		Where("id = ?", networkTopologyId).
		Where("deleted_at IS NULL").
		First(&networkTopology)
	if err != nil {
		if !errors.Is(err, orm.ErrRecordNotFound) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get network topology <%s> for tenant <%d>, error: %v", networkTopologyId, tenantId, err)
		}
		return nil, api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Network Topology with id=<%s> not found", networkTopologyId), err)
	}

	return &networkTopology, nil
}

func GetNetworkTopologyViewById(ctx context.Context, tenantId int32, networkTopologyId string) (*models.NetworkTopologyView, *api.Error) {
	var networkTopologyView models.NetworkTopologyView
	err := runaidb.DB().WithContext(ctx).
		Model(&models.NetworkTopologyView{}).
		Where("tenant_id = ?", tenantId).
		Where("id = ?", networkTopologyId).
		First(&networkTopologyView)
	if err != nil {
		if !errors.Is(err, orm.ErrRecordNotFound) {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get network topology view <%s> for tenant <%d>, error: %v", networkTopologyId, tenantId, err)
		}
		return nil, api.NewError(http.StatusInternalServerError,
			fmt.Sprintf("Network Topology with id=<%s> not found", networkTopologyId), err)
	}
	return &networkTopologyView, nil
}

func CountNetworkTopologies(ctx context.Context, tenantId int32, permittedScopes openapi.PermittedScopes, filterBy []queryapi.Filter) (int64, *api.Error) {
	query := runaidb.DB().WithContext(ctx).
		Model(&models.NetworkTopology{}).
		Where("tenant_id = ?", tenantId).
		Where("deleted_at IS NULL").
		Where(GetPermittedScopesFilter(permittedScopes, "cluster_id"))

	query = query.WithEnricher(querydbenrich.Filters(filterBy))

	var count int64
	if err := query.Count(&count); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count network topologies for tenant <%d>, error: %v", tenantId, err)
		return 0, api.NewError(http.StatusInternalServerError, "Failed to count network topologies", err)
	}

	return count, nil
}

// findTopologiesToDelete finds NetworkTopology records with Source "Cluster" that are not present in the provided networkTopologiesFields array
func findTopologiesToDelete(ctx context.Context, tx orm.DB, tenantId int32, clusterUuid uuid.UUID, networkTopologiesFields []api.NetworkTopologySyncFields) ([]models.NetworkTopology, *api.Error) {
	providedNames := datastructs.Set[string]{}
	for _, fields := range networkTopologiesFields {
		providedNames.Add(fields.Name)
	}

	namesList := providedNames.Keys()

	// Query database directly for Cluster topologies that are NOT in the provided names list
	var topologiesToDelete []models.NetworkTopology
	query := tx.
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterUuid.String()).
		Where("source = ?", NetworkTopologiesClusterSource).
		Where("deleted_at IS NULL")

	if len(namesList) > 0 {
		query = query.Where("name NOT IN ?", namesList)
	}

	if err := query.Find(&topologiesToDelete); err != nil {
		errorMsg := "Failed to query Cluster topologies not in provided names"
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errorMsg, err)
		return nil, api.NewError(http.StatusInternalServerError, errorMsg, err)
	}

	return topologiesToDelete, nil
}

// bulkDeleteNetworkTopologies performs a bulk soft delete of network topologies by their IDs
func bulkDeleteNetworkTopologies(ctx context.Context, tx orm.DB, tenantId int32, clusterId uuid.UUID, topologyIDs []uuid.UUID) *api.Error {
	if len(topologyIDs) == 0 {
		return nil
	}
	now := time.Now()

	rowsAffected, err := tx.Model(&models.NetworkTopology{}).
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterId.String()).
		Where("id IN ?", topologyIDs).
		Updates(map[string]any{
			"deleted_at": &now,
			"phase":      models.NetworkTopologyPhaseDeleted,
		})
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to bulk delete network topologies, error: %v", err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errorMsg)
		return api.NewError(http.StatusInternalServerError, errorMsg, err)
	}

	if rowsAffected > 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Successfully deleted %d network topologies", rowsAffected)
	}

	// Clear network topology fields from nodepools for each deleted topology
	for _, topologyUuid := range topologyIDs {
		topologyIdStr := topologyUuid.String()
		if clearErr := clearNetworkTopologyFromNodepools(ctx, tx, tenantId, clusterId, &topologyIdStr, nil); clearErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to clear network topology <%s> from nodepools: %v", topologyIdStr, clearErr)
			return clearErr
		}
	}

	return nil
}

func createClusterNetworkTopology(ctx context.Context, tx orm.DB, tenantId int32, subjectId string, clusterUuid uuid.UUID, networkTopologyFields api.NetworkTopologySyncFields) *api.Error {
	if networkTopologyFields.Name == "" {
		return api.NewError(http.StatusBadRequest, "Network Topology Name is required", nil)
	}
	levels := pq.StringArray(networkTopologyFields.Levels)

	source := networkTopologyFields.Source
	if source == "" {
		source = NetworkTopologiesClusterSource
	}

	// Create new network topology
	newTopology := models.NetworkTopology{
		Name:      networkTopologyFields.Name,
		Levels:    levels,
		Source:    source,
		ClusterID: clusterUuid,
		TenantID:  tenantId,
		Phase:     models.NetworkTopologyPhaseReady,
		CreatedBy: &subjectId,
		Grove:     adapters.NewNetworkTopologySyncGroveAdapter(networkTopologyFields.Grove).ToDBModel(),
	}

	if _, createErr := tx.Create(&newTopology); createErr != nil {
		errorMsg := fmt.Sprintf("Failed to create network topology with name <%s>",
			networkTopologyFields.Name)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errorMsg, createErr)
		return api.NewError(http.StatusInternalServerError, errorMsg, createErr)
	}

	return nil
}

func updateClusterNetworkTopology(ctx context.Context, tx orm.DB, tenantId int32, clusterUuid uuid.UUID, networkTopologyFields api.NetworkTopologySyncFields) *api.Error {
	if networkTopologyFields.Name == "" {
		return api.NewError(http.StatusBadRequest, "Network Topology Name is required", nil)
	}

	grove := adapters.NewNetworkTopologySyncGroveAdapter(networkTopologyFields.Grove).ToDBModel()
	if grove == nil {
		grove = &models.Grove{}
	}
	update := models.NetworkTopology{Grove: grove}

	rowsAffected, err := tx.WithContext(ctx).
		Model(&models.NetworkTopology{}).
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterUuid.String()).
		Where("name = ?", networkTopologyFields.Name).
		Where("source = ?", NetworkTopologiesClusterSource).
		Where("deleted_at IS NULL").
		Select("grove_name", "grove_domains", "grove_conditions").
		Updates(&update)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to update network topology with name <%s>", networkTopologyFields.Name)
		return api.NewError(http.StatusInternalServerError, errorMsg, err)
	}

	if rowsAffected == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Network topology <%s> not found for update in cluster <%s>", networkTopologyFields.Name, clusterUuid.String())
	}

	return nil
}

func setRunAiTopologiesPhaseToReady(ctx context.Context, tx orm.DB, tenantId int32, clusterUuid uuid.UUID, request api.NetworkTopologySyncRequest) *api.Error {
	runAiNames := adapters.NewNetworkTopologiesSyncAdapter(request.NetworkTopologies).
		BySources(NetworkTopologiesRunAiSource).ToNames()

	if len(runAiNames) == 0 {
		return nil
	}

	// Update only network topologies whose names are in runAiNetworkTopologies
	_, err := tx.Model(&models.NetworkTopology{}).
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterUuid.String()).
		Where("source = ?", NetworkTopologiesRunAiSource).
		Where("name IN ?", runAiNames).
		Where("deleted_at IS NULL").
		Updates(map[string]any{
			"phase": models.NetworkTopologyPhaseReady,
		})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update runAi topologies phase to ready: %v", err)
		return api.NewError(http.StatusInternalServerError, "Failed to update runAi topologies phase", err)
	}

	return nil
}

func fullSyncNetworkTopologies(ctx context.Context, tenantId int32, subjectId string, clusterUuid uuid.UUID, request api.NetworkTopologySyncRequest) *api.Error {
	fullSync := request.FullSync
	if !fullSync {
		return nil
	}
	clusterNetworkTopologiesFields := adapters.NewNetworkTopologiesSyncAdapter(request.NetworkTopologies).
		BySources(NetworkTopologiesClusterSource, "").ToFields()
	// Use a transaction for consistency
	tx := runaidb.DB().WithContext(ctx).Begin()

	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
	tx.SavePoint("sp")

	// Get list of idsToDelete and toAddOrUpdate
	var toDelete []models.NetworkTopology
	var toAddOrUpdate []api.NetworkTopologySyncFields

	var apiErr *api.Error
	toDelete, apiErr = findTopologiesToDelete(ctx, tx, tenantId, clusterUuid, clusterNetworkTopologiesFields)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to find topologies to delete: %v", apiErr)
		return apiErr
	}

	toDeleteAdapter := models.NewNetworkTopologyArrayAdapter(toDelete)
	namesToDelete := toDeleteAdapter.ToNames()
	for _, ntf := range clusterNetworkTopologiesFields {
		if !namesToDelete.Has(ntf.Name) {
			toAddOrUpdate = append(toAddOrUpdate, ntf)
		}
	}

	// Actually execute the operations
	// Execute delete operations
	idsToDelete := toDeleteAdapter.ToIds()
	err := bulkDeleteNetworkTopologies(ctx, tx, tenantId, clusterUuid, idsToDelete)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete topologies: %v", err)
		tx.RollbackTo("sp")
	} else {
		tx.SavePoint("sp")
	}

	// Execute create operations
	for _, fields := range toAddOrUpdate {
		err := createClusterNetworkTopology(ctx, tx, tenantId, subjectId, clusterUuid, fields)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to sync <%s> in cluster <%s>, error: %v",
				fields.Name, clusterUuid.String(), err)
			tx.RollbackTo("sp")
		} else {
			tx.SavePoint("sp")
		}
	}

	// Set cluster topologies phase to ready after successful sync
	if err := setRunAiTopologiesPhaseToReady(ctx, tx, tenantId, clusterUuid, request); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to set Cluster Netowrk Topologies Phase to Ready in cluster <%s>, error: %v",
			clusterUuid.String(), err)
		tx.RollbackTo("sp")
	} else {
		tx.SavePoint("sp")
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit network topologies sync transaction: %v", err)
		return api.NewError(http.StatusInternalServerError, "Failed to commit transaction", err)
	}

	committed = true

	return nil
}

func syncActionNetworkTopologies(ctx context.Context, tenantId int32, subjectId string, clusterUuid uuid.UUID, request api.NetworkTopologySyncRequest) *api.Error {
	fullSync := request.FullSync
	if fullSync || len(request.NetworkTopologies) == 0 {
		return nil
	}

	clusterNetworkTopologiesFields := adapters.NewNetworkTopologiesSyncAdapter(request.NetworkTopologies).
		BySources(NetworkTopologiesClusterSource, "").ToFields()

	// Use a transaction for consistency
	tx := runaidb.DB().WithContext(ctx).Begin()

	committed := false
	defer orm.FinalizeTransaction(tx, &committed)
	tx.SavePoint("sp")

	// Execute create operations
	for _, fields := range clusterNetworkTopologiesFields {
		switch fields.Action {
		case api.NETWORKTOPOLOGYSYNCACTION_ADD:
			err := createClusterNetworkTopology(ctx, tx, tenantId, subjectId, clusterUuid, fields)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create <%s> in cluster <%s>, error: %v",
					fields.Name, clusterUuid.String(), err)
				tx.RollbackTo("sp")
			} else {
				tx.SavePoint("sp")
			}
		case api.NETWORKTOPOLOGYSYNCACTION_DELETE:
			err := deleteNetworkTopology(ctx, tx, tenantId, clusterUuid, nil, &fields.Name)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete <%s> in cluster <%s>, error: %v",
					fields.Name, clusterUuid.String(), err)
				tx.RollbackTo("sp")
			} else {
				tx.SavePoint("sp")
			}
		case api.NETWORKTOPOLOGYSYNCACTION_UPDATE:
			err := updateClusterNetworkTopology(ctx, tx, tenantId, clusterUuid, fields)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to update <%s> in cluster <%s>, error: %v",
					fields.Name, clusterUuid.String(), err)
				tx.RollbackTo("sp")
			} else {
				tx.SavePoint("sp")
			}
		}
	}

	// Set cluster topologies phase to ready after successful sync
	if apiErr := setRunAiTopologiesPhaseToReady(ctx, tx, tenantId, clusterUuid, request); apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to set cluster topologies phase to ready: %v", apiErr)
		return apiErr
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit network topologies sync transaction: %v", err)
		return api.NewError(http.StatusInternalServerError, "Failed to commit transaction", err)
	}

	committed = true

	return nil
}

func deleteNetworkTopology(ctx context.Context, tx orm.DB, tenantId int32, clusterUuid uuid.UUID, networkTopologyId, networkTopologyName *string) *api.Error {
	if networkTopologyName == nil && networkTopologyId == nil {
		err := errors.New("failed to delete network topology, either name or networkTopologyId must be specified")
		errMessage := fmt.Sprintf("Failed to delete network topology for clusterId <%s> and tenant <%d>", clusterUuid.String(), tenantId)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s, error: %v", errMessage, err)
		return api.NewError(http.StatusBadRequest, errMessage, err)
	}

	now := time.Now()
	query := tx.WithContext(ctx).
		Model(&models.NetworkTopology{}).
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterUuid.String())

	if networkTopologyName != nil {
		query = query.
			Where("name = ?", *networkTopologyName)
	}
	if networkTopologyId != nil {
		query = query.
			Where("id = ?", *networkTopologyId)
	}

	rowsAffected, err := query.
		Updates(map[string]any{
			"deleted_at": &now,
			"phase":      models.NetworkTopologyPhaseDeleted,
		})

	networkTopologyIdStr := ""
	if networkTopologyId != nil {
		networkTopologyIdStr = *networkTopologyId
	}

	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete network topology <%s> for tenant <%d>, error: %v", networkTopologyIdStr, tenantId, err)
		return api.NewError(http.StatusInternalServerError, "Failed to delete network topology", err)
	}

	if rowsAffected == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Network topology <%s> not found for tenant <%d>", networkTopologyIdStr, tenantId)
		return nil
	}

	// Clear network topology fields from all nodepools using this network topology
	if clearErr := clearNetworkTopologyFromNodepools(ctx, tx, tenantId, clusterUuid, networkTopologyId, networkTopologyName); clearErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to clear network topology from nodepools: %v", clearErr)
		return clearErr
	}

	return nil
}

// clearNetworkTopologyFromNodepools sets network_topology_name and network_topology_id to nil for all nodepools using the specified network topology
func clearNetworkTopologyFromNodepools(ctx context.Context, tx orm.DB, tenantId int32, clusterId uuid.UUID, networkTopologyId *string, networkTopologyName *string) *api.Error {
	if networkTopologyName == nil && networkTopologyId == nil {
		return nil // Nothing to clear
	}

	updateMap := map[string]any{
		"network_topology_name": nil,
		"network_topology_id":   nil,
	}

	query := tx.WithContext(ctx).
		Model(&models.Nodepool{}).
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterId.String()).
		Where("deleted_at IS NULL")

	if networkTopologyName != nil {
		query = query.Where("network_topology_name = ?", *networkTopologyName)
	}
	if networkTopologyId != nil {
		query = query.Where("network_topology_id = ?", *networkTopologyId)
	}

	rowsAffected, err := query.UpdateColumns(updateMap)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to clear network topology from nodepools for tenant <%d>, cluster <%s>, error: %v", tenantId, clusterId.String(), err)
		return api.NewError(http.StatusInternalServerError, "Failed to clear network topology from nodepools", err)
	}

	if rowsAffected > 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Cleared network topology from %d nodepools for tenant <%d>, cluster <%s>", rowsAffected, tenantId, clusterId.String())
	}

	return nil
}
