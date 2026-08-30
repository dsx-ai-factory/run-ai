package dao

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
	"github.com/run-ai/runai/runai-application-packages/queryapi"
	querydbenrich "github.com/run-ai/runai/runai-application-packages/queryapi/dbenrich"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"github.com/run-ai/runai/runai-common-packages/runaidb/dbenrich"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm"
)

const ReadyStatus = "Ready"

// NodeFieldsSubset holds a subset of node fields for comparison (can be extended with more fields)
type NodeFieldsSubset struct {
	NetworkTopologyLabelValues *pq.StringArray `gorm:"column:network_topology_label_values;type:text[]"`
}

// UpsertNodeResult contains the result of an upsert operation
type UpsertNodeResult struct {
	Name           string
	NewValues      NodeFieldsSubset
	PreviousValues NodeFieldsSubset `gorm:"embedded;embeddedPrefix:previous_"`
}

// prepareNodeForUpsert sets ID/UpdatedAt and returns INSERT params for a node
func prepareNodeForUpsert(node *models.Node) []any {
	node.ID = uuid.New()
	node.UpdatedAt = time.Now()

	// Prepare JSON values
	conditionsJSON, _ := node.Conditions.Value()
	taintsJSON, _ := node.Taints.Value()
	var networkTopologyLabelsJSON any
	if node.NetworkTopologyLabels != nil {
		networkTopologyLabelsJSON, _ = node.NetworkTopologyLabels.Value()
	}

	return []any{
		node.ID, node.TenantId, node.ClusterUuid, node.Name,
		node.Status, conditionsJSON, taintsJSON, node.NodePool,
		node.GpuType, node.GpuCount, node.CreatedAt, node.UpdatedAt,
		node.NvLinkDomainUid, node.NvLinkCliqueId,
		node.NetworkTopologyLabelValues, networkTopologyLabelsJSON,
		node.TopologyManagerPolicy,
	}
}

func UpsertNode(ctx context.Context, node *models.Node) (*UpsertNodeResult, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError,
		fmt.Sprintf("Failed to upsert node <%s> in cluster <%s>", node.Name, node.ClusterUuid), nil)

	db := runaidb.DB().WithContext(ctx)

	nodeParams := prepareNodeForUpsert(node)

	var scanResult struct {
		PreviousNetworkTopologyLabelValues *pq.StringArray `gorm:"column:previous_network_topology_label_values;type:text[]"`
	}

	// Single query: CTE captures old value, then upsert, then return old value
	query := fmt.Sprintf(`
		WITH old_values AS (
			SELECT network_topology_label_values
			FROM node
			WHERE cluster_uuid = ? AND name = ?
		),
		upserted AS (
			INSERT INTO node (%s)
			VALUES (%s)
			ON CONFLICT (name, cluster_uuid) DO UPDATE SET %s
			RETURNING id
		)
		SELECT o.network_topology_label_values AS previous_network_topology_label_values
		FROM upserted u
		LEFT JOIN old_values o ON true`,
		models.Node{}.InsertColumnsSQL(),
		models.Node{}.InsertPlaceholdersSQL(),
		models.Node{}.UpdateSetSQL(),
	)

	// Build params: CTE params + INSERT params
	params := []any{node.ClusterUuid, node.Name}
	params = append(params, nodeParams...)

	if _, err := db.Raw(query, params...).Scan(&scanResult); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to upsert node <%s> in cluster <%s>, error: %v",
			node.Name, node.ClusterUuid, err)
		return nil, dbErr.WithDetails(err)
	}

	return &UpsertNodeResult{
		Name: node.Name,
		NewValues: NodeFieldsSubset{
			NetworkTopologyLabelValues: node.NetworkTopologyLabelValues,
		},
		PreviousValues: NodeFieldsSubset{
			NetworkTopologyLabelValues: scanResult.PreviousNetworkTopologyLabelValues,
		},
	}, nil
}

func UpsertNodes(ctx context.Context, nodes *[]models.Node, clusterUuid string) ([]UpsertNodeResult, *api.Error) {
	if len(*nodes) == 0 {
		return []UpsertNodeResult{}, nil
	}

	dbErr := api.NewError(http.StatusInternalServerError,
		fmt.Sprintf("Failed to upsert nodes in cluster <%s>", clusterUuid), nil)

	db := runaidb.DB().WithContext(ctx)

	// Build map of node name -> new topology values for result building
	nodeNewNetworkTopologyLabelValues := make(map[string]*pq.StringArray)
	nodeNames := make([]string, len(*nodes))
	for i, n := range *nodes {
		nodeNames[i] = n.Name
		nodeNewNetworkTopologyLabelValues[n.Name] = n.NetworkTopologyLabelValues
	}

	// Build VALUES clause and params
	var values []string
	params := []any{clusterUuid, pq.Array(nodeNames)}

	for i := range *nodes {
		node := &(*nodes)[i]
		nodeParams := prepareNodeForUpsert(node)

		values = append(values, fmt.Sprintf("(%s)", models.Node{}.InsertPlaceholdersSQL()))
		params = append(params, nodeParams...)
	}
	valuesClause := strings.Join(values, ", ")

	query := fmt.Sprintf(`
		WITH old_values AS (
			SELECT name, network_topology_label_values
			FROM node
			WHERE cluster_uuid = ? AND name = ANY(?)
		),
		upserted AS (
			INSERT INTO node (%s)
			VALUES %s
			ON CONFLICT (name, cluster_uuid) DO UPDATE SET %s
			RETURNING name
		)
		SELECT u.name, o.network_topology_label_values AS previous_network_topology_label_values
		FROM upserted u
		LEFT JOIN old_values o ON u.name = o.name`,
		models.Node{}.InsertColumnsSQL(),
		valuesClause,
		models.Node{}.UpdateSetSQL(),
	)

	var scanResults []struct {
		Name                               string
		PreviousNetworkTopologyLabelValues *pq.StringArray `gorm:"column:previous_network_topology_label_values;type:text[]"`
	}
	if _, err := db.Raw(query, params...).Scan(&scanResults); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to upsert nodes in cluster <%s>, error: %v",
			clusterUuid, err)
		return nil, dbErr.WithDetails(err)
	}

	// Build results with both new and previous values
	results := make([]UpsertNodeResult, len(scanResults))
	for i, sr := range scanResults {
		results[i] = UpsertNodeResult{
			Name: sr.Name,
			NewValues: NodeFieldsSubset{
				NetworkTopologyLabelValues: nodeNewNetworkTopologyLabelValues[sr.Name],
			},
			PreviousValues: NodeFieldsSubset{
				NetworkTopologyLabelValues: sr.PreviousNetworkTopologyLabelValues,
			},
		}
	}

	return results, nil
}

func GetNodeById(ctx context.Context, tenantId int32, nodeId string, permittedScopes openapi.PermittedScopes) (*models.Node, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError,
		fmt.Sprintf("Failed to get node with id <%s>", nodeId), nil)

	node := &models.Node{}
	query := runaidb.DB().WithContext(ctx)

	if !runaictx.IsManagementTenant(ctx) {
		// TODO remove the OR once all tenants are in CP version 2.18 or higher
		query = query.Where("tenant_id = ? OR tenant_id IS NULL", tenantId)
	}

	query = query.Where(GetPermittedScopesFilter(permittedScopes, "cluster_uuid"))
	query = query.
		Where("id = ?", nodeId)

	if err := query.First(&node); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed get node with id <%s>, error: %v",
			nodeId, err)
		return nil, dbErr.WithDetails(err)
	}
	return node, nil
}

// GetNodeInternal retrieves minimal node fields by cluster UUID and name.
// This is optimized for internal services that only need specific fields like topology labels.
func GetNodeInternal(ctx context.Context, tenantId int32, clusterId, nodeName string, selectFields []string) (*models.Node, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError,
		fmt.Sprintf("Failed to get node <%s> in cluster <%s> tenantId <%d>", nodeName, clusterId, tenantId), nil)

	node := &models.Node{}
	query := runaidb.DB().WithContext(ctx).
		Select(selectFields).
		Where("tenant_id = ?", tenantId).
		Where("cluster_uuid = ?", clusterId).
		Where("name = ?", nodeName)

	if err := query.First(&node); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get node <%s> in cluster <%s>, error: %v",
			nodeName, clusterId, err)
		return nil, dbErr.WithDetails(err)
	}
	return node, nil
}

func GetNodesByClusterUuid(
	ctx context.Context,
	tenantId int32,
	clusterUuid string,
	permittedScopes openapi.PermittedScopes,
	filterBy []queryapi.Filter,
	sortBy string,
	sortOrder string,
	offset int32,
	limit int32,
	search string,
) (*[]models.Node, *api.Error) {
	query := runaidb.DB().WithContext(ctx).
		Where("cluster_uuid = ?", clusterUuid)
	return getNodesInternal(ctx, query, tenantId, &permittedScopes, filterBy, sortBy, sortOrder, offset, limit, search)
}

func getNodesInternal(
	ctx context.Context,
	query orm.DB,
	tenantId int32,
	permittedScopes *openapi.PermittedScopes,
	filterBy []queryapi.Filter,
	sortBy string,
	sortOrder string,
	offset int32,
	limit int32,
	search string,
) (*[]models.Node, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get nodes", nil)
	nodes := &[]models.Node{}

	if !runaictx.IsManagementTenant(ctx) {
		// TODO remove the OR once all tenants are in CP version 2.18 or higher
		query = query.Where("tenant_id = ? OR tenant_id IS NULL", tenantId)
	}

	if permittedScopes != nil {
		query = query.
			Where(GetPermittedScopesFilter(*permittedScopes, "cluster_uuid"))
	}
	query = query.WithEnricher(querydbenrich.Filters(filterBy))
	query = applySortingParam(query, sortBy, sortOrder)

	if offset != 0 {
		query = query.Offset(int(offset))
	}
	if limit != 0 {
		query = query.Limit(int(limit))
	}

	if search != "" {
		query = query.WithEnricher(dbenrich.TSSearch(search, "search_vector"))
	}

	if err := query.Find(nodes); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed get nodes>, error: %v", err)
		return nil, dbErr.WithDetails(err)
	}
	return nodes, nil
}

func GetNodes(ctx context.Context, tenantId int32, permittedScopes *openapi.PermittedScopes, filterBy []queryapi.Filter,
	sortBy string, sortOrder string, offset int32, limit int32, search string,
) (*[]models.Node, *api.Error) {
	query := runaidb.DB().WithContext(ctx)
	return getNodesInternal(ctx, query, tenantId, permittedScopes, filterBy, sortBy, sortOrder, offset, limit, search)
}

func CountNodesByClusterUuid(ctx context.Context, tenantId int32, clusterUuid string, permittedScopes openapi.PermittedScopes, filterBy []queryapi.Filter, search string) (int64, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError,
		fmt.Sprintf("Failed to count nodes in cluster <%s>", clusterUuid), nil)

	query := runaidb.DB().WithContext(ctx).Table(models.NodeTable)

	if !runaictx.IsManagementTenant(ctx) {
		// TODO remove the OR once all tenants are in CP version 2.18 or higher
		query = query.Where("tenant_id = ? OR tenant_id IS NULL", tenantId)
	}

	query = query.
		Where(GetPermittedScopesFilter(permittedScopes, "cluster_uuid")).
		Where("cluster_uuid = ?", clusterUuid)

	query = query.WithEnricher(querydbenrich.Filters(filterBy))

	if search != "" {
		query = query.WithEnricher(dbenrich.TSSearch(search, "search_vector"))
	}

	var count int64
	if err := query.Count(&count); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count nodes, err: %v", err)
		return 0, dbErr.WithDetails(err)
	}

	return count, nil
}

func CountNodes(ctx context.Context, tenantId int32, permittedScopes *openapi.PermittedScopes, filterBy []queryapi.Filter, search string) (int64, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to count nodes", nil)

	query := runaidb.DB().WithContext(ctx).Table(models.NodeTable)

	if !runaictx.IsManagementTenant(ctx) {
		// TODO remove the OR once all tenants are in CP version 2.18 or higher
		query = query.Where("tenant_id = ? OR tenant_id IS NULL", tenantId)
	}

	if permittedScopes != nil {
		query = query.Where(GetPermittedScopesFilter(*permittedScopes, "cluster_uuid"))
	}

	query = query.WithEnricher(querydbenrich.Filters(filterBy))

	if search != "" {
		query = query.WithEnricher(dbenrich.TSSearch(search, "search_vector"))
	}

	var count int64
	if err := query.Count(&count); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count nodes, err: %v", err)
		return 0, dbErr.WithDetails(err)
	}

	return count, nil
}

func DeleteNode(ctx context.Context, tenantId int32, clusterUuid, nodeName string) *api.Error {
	dbErr := api.NewError(http.StatusInternalServerError,
		fmt.Sprintf("Failed to delete node <%s> in cluster <%s>", nodeName, clusterUuid), nil)

	rowsAffected, err := runaidb.DB().WithContext(ctx).
		// TODO remove the OR once all tenants are in CP version 2.18 or higher
		Where("tenant_id = ? OR tenant_id IS NULL", tenantId).
		Where("cluster_uuid = ?", clusterUuid).
		Where("name = ?", nodeName).
		Delete(&models.Node{})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete node <%s> in cluster <%s>, error: %v",
			nodeName, clusterUuid, err)
		return dbErr.WithDetailsFromDbResult(rowsAffected, err)
	}

	return nil
}

func DeleteUnreferencedNodes(ctx context.Context, nodes *[]models.Node, tenantId int32, clusterUuid, from, to string) *api.Error {
	errorStr := fmt.Sprintf("Failed to delete unreferenced nodes in cluster <%s>", clusterUuid)
	if from != "" {
		errorStr = fmt.Sprintf("%s; from node <%s>", errorStr, from)
	}
	if to != "" {
		errorStr = fmt.Sprintf("%s; to node <%s>", errorStr, to)
	}
	dbErr := api.NewError(http.StatusInternalServerError,
		errorStr, nil)

	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	query := tx.
		// TODO remove the OR once all tenants are in CP version 2.18 or higher
		Where("tenant_id = ? OR tenant_id IS NULL", tenantId).
		Where("cluster_uuid = ?", clusterUuid)
	if from != "" {
		query = query.Where("name >= ?", from)
	}
	if to != "" {
		query = query.Where("name <= ?", to)
	}
	if len(*nodes) > 0 {
		query = query.Where("name NOT IN ?", listNodeNamesForQuery(nodes))
	}
	rowsAffected, err := query.Delete(&[]models.Node{})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete unreferenced nodes in cluster <%s>, error: %v",
			clusterUuid, err)
		return dbErr.WithDetailsFromDbResult(rowsAffected, err)
	}
	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit delete unreferenced nodes in cluster <%s>, error: %v",
			clusterUuid, err)
		return dbErr.WithDetails(err)
	}
	committed = true
	return nil
}

func listNodeNamesForQuery(nodes *[]models.Node) []string {
	nodeNames := []string{}
	for _, node := range *nodes {
		nodeNames = append(nodeNames, node.Name)
	}

	return nodeNames
}

func getGpuNodeTelemetry(ctx context.Context, tenantId int32, tenantIds []string, clusterId string,

	nodepoolName string, permittedScopes *openapi.PermittedScopes, groupBy [][]string, dataField string, onlyReadyNodes bool,
) (*[]models.TelemetryGroupValue, error) {
	tx := runaidb.DB().WithContext(ctx)

	telemetryGroupValues := []models.TelemetryGroupValue{}

	selectPhrase := fmt.Sprintf("%s as data", dataField)
	for i, group := range groupBy {
		selectPhrase = fmt.Sprintf("%s, %s as group%d_value", selectPhrase, group[0], i+1)
	}

	isMgmtTenant := runaictx.IsManagementTenant(ctx)

	query := tx.
		Table(models.NodeTable).
		Select(selectPhrase).
		Where("gpu_count > 0")

	if !isMgmtTenant {
		// TODO remove the OR once all tenants are in CP version 2.18 or higher
		query = query.Where("tenant_id = ? OR tenant_id IS NULL", tenantId)
	}

	query = query.Where(GetPermittedScopesFilter(*permittedScopes, "cluster_uuid"))

	if len(tenantIds) > 0 {
		query = query.Where("tenant_id IN ?", tenantIds)
	}
	if onlyReadyNodes {
		query = query.Where("status = ?", ReadyStatus)
	}
	if clusterId != "" {
		query = query.Where("cluster_uuid = ?", clusterId)
	}
	if nodepoolName != "" {
		query = query.Where("node_pool = ?", nodepoolName)
	}
	for _, group := range groupBy {
		query = query.Group(group[0])
	}

	if _, err := query.Scan(&telemetryGroupValues); err != nil {
		return &telemetryGroupValues, err
	}

	return &telemetryGroupValues, nil
}

func GetGpuNodesCount(ctx context.Context, tenantId int32, tenantIds []string, clusterId string, nodepoolName string,
	permittedScopes *openapi.PermittedScopes, groupBy [][]string, onlyReadyNodes bool,
) (*[]models.TelemetryGroupValue, *api.Error) {
	readyStr := ""
	if onlyReadyNodes {
		readyStr = "ready "
	}

	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get %sgpu nodes count", readyStr), nil)

	workloadTelemetry, err := getGpuNodeTelemetry(ctx, tenantId, tenantIds, clusterId, nodepoolName, permittedScopes, groupBy, "COUNT(id)", onlyReadyNodes)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get %sgpu nodes count for cluster %s and nodepool %s, err: %v",
			readyStr, clusterId, nodepoolName, err)
		return workloadTelemetry, dbErr.WithDetails(err)
	}

	return workloadTelemetry, nil
}

func GetGpuCount(ctx context.Context, tenantId int32, tenantIds []string, clusterId string, nodepoolName string,
	permittedScopes *openapi.PermittedScopes, groupBy [][]string, onlyReadyNodes bool,
) (*[]models.TelemetryGroupValue, *api.Error) {
	readyStr := ""
	if onlyReadyNodes {
		readyStr = "ready "
	}

	dbErr := api.NewError(http.StatusInternalServerError, fmt.Sprintf("Failed to get %sgpu count", readyStr), nil)

	workloadTelemetry, err := getGpuNodeTelemetry(ctx, tenantId, tenantIds, clusterId, nodepoolName, permittedScopes, groupBy, "COALESCE(SUM(gpu_count), 0)", onlyReadyNodes)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get %sgpu count for cluster %s and nodepool %s, err: %v",
			readyStr, clusterId, nodepoolName, err)
		return workloadTelemetry, dbErr.WithDetails(err)
	}

	return workloadTelemetry, nil
}

func HardDeleteAllTenantNodes(ctx context.Context, tenantId int32) error {
	tx := runaidb.DB().WithContext(ctx).Begin().Unscoped()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	if _, err := tx.Where("tenant_id = ?", tenantId).Delete(&models.Node{}); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete all tenant nodes, error: %v", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to commit delete all tenant nodes, error: %v", err)
		return err
	}
	committed = true
	return nil
}

func DeleteClusterNodes(ctx context.Context, tenantId int32, clusterUuid string) error {
	_, err := runaidb.DB().WithContext(ctx).
		Where("tenant_id = ?", tenantId).
		Where("cluster_uuid = ?", clusterUuid).
		Delete(&models.Node{})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete cluster %s nodes, error: %v", clusterUuid, err)
		return err
	}

	return nil
}
