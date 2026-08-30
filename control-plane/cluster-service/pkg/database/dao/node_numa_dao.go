package dao

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm/clause"
)

type NodeNumaSnapshot struct {
	NodeName  string
	NumaNodes []models.NodeNuma
}

func GetNumaNodes(ctx context.Context, tenantId int32, nodeId uuid.UUID) ([]models.NodeNuma, *api.Error) {
	numaNodes := []models.NodeNuma{}
	err := runaidb.DB().WithContext(ctx).
		Where("tenant_id = ?", tenantId).
		Where("node_id = ?", nodeId).
		Order("name ASC").
		Find(&numaNodes)
	if err != nil {
		return nil, api.NewError(
			http.StatusInternalServerError,
			fmt.Sprintf("Failed to get NUMA nodes for node <%s>", nodeId),
			nil,
		).WithDetails(err)
	}
	return numaNodes, nil
}

func SyncNumaNodes(
	ctx context.Context,
	snapshots []NodeNumaSnapshot,
	tenantId int32,
	clusterId uuid.UUID,
	deleteMissing bool,
	from string,
	to string,
) ([]string, []string, *api.Error) {
	dbErr := api.NewError(
		http.StatusInternalServerError,
		fmt.Sprintf("Failed to sync NUMA nodes in cluster <%s>", clusterId),
		nil,
	)

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	nodesByName, err := getNodesByNameForNumaSync(tx, tenantId, clusterId.String(), snapshots)
	if err != nil {
		return nil, nil, dbErr.WithDetails(err)
	}

	const savepoint = "numa_nodes_sync"
	tx.SavePoint(savepoint)

	syncedNodes := []string{}
	failedNodes := []string{}
	requestedNodeNames := make([]string, 0, len(snapshots))
	for _, snapshot := range snapshots {
		requestedNodeNames = append(requestedNodeNames, snapshot.NodeName)
		node, found := nodesByName[snapshot.NodeName]
		if !found {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(
				"Failed to sync NUMA nodes for node <%s> in cluster <%s>: node does not exist",
				snapshot.NodeName, clusterId)
			failedNodes = append(failedNodes, snapshot.NodeName)
			continue
		}

		if err = upsertNumaNodesSnapshot(tx, tenantId, clusterId, node.ID, snapshot.NumaNodes); err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(
				"Failed to sync NUMA nodes for node <%s> in cluster <%s>: %v",
				snapshot.NodeName, clusterId, err)
			tx.RollbackTo(savepoint)
			failedNodes = append(failedNodes, snapshot.NodeName)
			continue
		}

		tx.SavePoint(savepoint)
		syncedNodes = append(syncedNodes, snapshot.NodeName)
	}

	if deleteMissing {
		if err = deleteMissingNumaNodes(
			tx, tenantId, clusterId, requestedNodeNames, from, to,
		); err != nil {
			return nil, nil, dbErr.WithDetails(err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, dbErr.WithDetails(err)
	}
	committed = true
	return syncedNodes, failedNodes, nil
}

func getNodesByNameForNumaSync(
	tx orm.DB,
	tenantId int32,
	clusterUuid string,
	snapshots []NodeNumaSnapshot,
) (map[string]models.Node, error) {
	nodeNames := make([]string, 0, len(snapshots))
	for _, snapshot := range snapshots {
		nodeNames = append(nodeNames, snapshot.NodeName)
	}

	nodes := []models.Node{}
	if len(nodeNames) > 0 {
		if err := tx.
			Select("id", "name").
			Where("tenant_id = ?", tenantId).
			Where("cluster_uuid = ?", clusterUuid).
			Where("name IN ?", nodeNames).
			Find(&nodes); err != nil {
			return nil, fmt.Errorf("failed to resolve nodes for NUMA sync: %w", err)
		}
	}

	nodesByName := make(map[string]models.Node, len(nodes))
	for _, node := range nodes {
		nodesByName[node.Name] = node
	}
	return nodesByName, nil
}

func upsertNumaNodesSnapshot(
	tx orm.DB,
	tenantId int32,
	clusterId uuid.UUID,
	nodeId uuid.UUID,
	numaNodes []models.NodeNuma,
) error {
	numaNodeNames := make([]string, 0, len(numaNodes))
	for i := range numaNodes {
		numaNodes[i].TenantId = tenantId
		numaNodes[i].ClusterId = clusterId
		numaNodes[i].NodeId = nodeId
		numaNodeNames = append(numaNodeNames, numaNodes[i].Name)
	}

	if len(numaNodes) > 0 {
		if _, err := tx.
			Clauses(clause.OnConflict{
				Columns: []clause.Column{
					{Name: "node_id"},
					{Name: "name"},
				},
				DoUpdates: clause.AssignmentColumns([]string{
					"total_cpu_cores",
					"allocated_cpu_cores",
					"total_cpu_memory_bytes",
					"allocated_cpu_memory_bytes",
					"total_gpus",
					"allocated_gpus",
					"updated_at",
				}),
			}).
			Create(&numaNodes); err != nil {
			return err
		}
	}

	deleteQuery := tx.
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterId).
		Where("node_id = ?", nodeId)
	if len(numaNodeNames) > 0 {
		deleteQuery = deleteQuery.Where("name NOT IN ?", numaNodeNames)
	}
	_, err := deleteQuery.Delete(&models.NodeNuma{})
	return err
}

func deleteMissingNumaNodes(
	tx orm.DB,
	tenantId int32,
	clusterId uuid.UUID,
	requestedNodeNames []string,
	from string,
	to string,
) error {
	nodeIdsToDelete := tx.Model(&models.Node{}).
		Select("id").
		Where("tenant_id = ?", tenantId).
		Where("cluster_uuid = ?", clusterId.String())
	if from != "" {
		nodeIdsToDelete = nodeIdsToDelete.Where("name >= ?", from)
	}
	if to != "" {
		nodeIdsToDelete = nodeIdsToDelete.Where("name <= ?", to)
	}
	if len(requestedNodeNames) > 0 {
		nodeIdsToDelete = nodeIdsToDelete.Where("name NOT IN ?", requestedNodeNames)
	}

	if _, err := tx.
		Where("tenant_id = ?", tenantId).
		Where("cluster_id = ?", clusterId).
		Where("node_id IN (?)", nodeIdsToDelete).
		Delete(&models.NodeNuma{}); err != nil {
		return fmt.Errorf("failed to delete missing NUMA details in cluster <%s>: %w", clusterId, err)
	}
	return nil
}
