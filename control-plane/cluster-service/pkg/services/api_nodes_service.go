/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	labels_cache "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/clients/cache"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/versions"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"golang.org/x/exp/slices"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/clients"
	backend_clients "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/clients/backend"
	workloads_clients "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/clients/workloads"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/services/metrics"
	telemetry_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/telemetry"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"
)

const (
	SupportedDbTelemetryVersion = "2.15"
)

var (
	metricsOnlyTelemetryTypes = map[api.NodeTelemetryType]bool{
		api.NODETELEMETRYTYPE_IDLE_ALLOCATED_GPUS: true,
		api.NODETELEMETRYTYPE_FREE_GPUS:           true,
		// api.NODETELEMETRYTYPE_USED_GPUS:             true,
		api.NODETELEMETRYTYPE_ALLOCATED_GPUS:         true,
		api.NODETELEMETRYTYPE_TOTAL_CPU_CORES:        true,
		api.NODETELEMETRYTYPE_USED_CPU_CORES:         true,
		api.NODETELEMETRYTYPE_ALLOCATED_CPU_CORES:    true,
		api.NODETELEMETRYTYPE_TOTAL_GPU_MEMORY_BYTES: true,
		api.NODETELEMETRYTYPE_USED_GPU_MEMORY_BYTES:  true,
		// api.NODETELEMETRYTYPE_ALLOCATED_GPU_MEMORY: true,
		api.NODETELEMETRYTYPE_TOTAL_CPU_MEMORY_BYTES:     true,
		api.NODETELEMETRYTYPE_USED_CPU_MEMORY_BYTES:      true,
		api.NODETELEMETRYTYPE_ALLOCATED_CPU_MEMORY_BYTES: true,
		api.NODETELEMETRYTYPE_FULLY_FREE_GPU_NODES:       true,
	}

	// those telemetry types can only be fetched using the db, metrics not supported
	dbOnlyTelemetryTypes = map[api.NodeTelemetryType]bool{
		api.NODETELEMETRYTYPE_READY_GPU_NODES: true,
		api.NODETELEMETRYTYPE_READY_GPUS:      true,
	}

	allSupportedTypesMap = map[api.NodeTelemetryType]bool{
		api.NODETELEMETRYTYPE_IDLE_ALLOCATED_GPUS: true,
		api.NODETELEMETRYTYPE_READY_GPU_NODES:     true,
		api.NODETELEMETRYTYPE_READY_GPUS:          true,
		api.NODETELEMETRYTYPE_TOTAL_GPU_NODES:     true,
		api.NODETELEMETRYTYPE_TOTAL_GPUS:          true,
		api.NODETELEMETRYTYPE_FREE_GPUS:           true,
		// api.NODETELEMETRYTYPE_USED_GPUS:             true,
		api.NODETELEMETRYTYPE_ALLOCATED_GPUS:         true,
		api.NODETELEMETRYTYPE_TOTAL_CPU_CORES:        true,
		api.NODETELEMETRYTYPE_USED_CPU_CORES:         true,
		api.NODETELEMETRYTYPE_ALLOCATED_CPU_CORES:    true,
		api.NODETELEMETRYTYPE_TOTAL_GPU_MEMORY_BYTES: true,
		api.NODETELEMETRYTYPE_USED_GPU_MEMORY_BYTES:  true,
		// api.NODETELEMETRYTYPE_ALLOCATED_GPU_MEMORY: true,
		api.NODETELEMETRYTYPE_TOTAL_CPU_MEMORY_BYTES:     true,
		api.NODETELEMETRYTYPE_USED_CPU_MEMORY_BYTES:      true,
		api.NODETELEMETRYTYPE_ALLOCATED_CPU_MEMORY_BYTES: true,
		api.NODETELEMETRYTYPE_FULLY_FREE_GPU_NODES:       true,
	}
	allSupportedTypes = []string{
		string(api.NODETELEMETRYTYPE_IDLE_ALLOCATED_GPUS),
		string(api.NODETELEMETRYTYPE_READY_GPU_NODES),
		string(api.NODETELEMETRYTYPE_READY_GPUS),
		string(api.NODETELEMETRYTYPE_TOTAL_GPU_NODES),
		string(api.NODETELEMETRYTYPE_TOTAL_GPUS),
		string(api.NODETELEMETRYTYPE_FREE_GPUS),
		// string(api.NODETELEMETRYTYPE_USED_GPUS),
		string(api.NODETELEMETRYTYPE_ALLOCATED_GPUS),
		string(api.NODETELEMETRYTYPE_TOTAL_CPU_CORES),
		string(api.NODETELEMETRYTYPE_USED_CPU_CORES),
		string(api.NODETELEMETRYTYPE_ALLOCATED_CPU_CORES),
		string(api.NODETELEMETRYTYPE_TOTAL_GPU_MEMORY_BYTES),
		string(api.NODETELEMETRYTYPE_USED_GPU_MEMORY_BYTES),
		// string(api.NODETELEMETRYTYPE_ALLOCATED_GPU_MEMORY),
		string(api.NODETELEMETRYTYPE_TOTAL_CPU_MEMORY_BYTES),
		string(api.NODETELEMETRYTYPE_USED_CPU_MEMORY_BYTES),
		string(api.NODETELEMETRYTYPE_ALLOCATED_CPU_MEMORY_BYTES),
		string(api.NODETELEMETRYTYPE_FULLY_FREE_GPU_NODES),
	}
)

// NodeService is a service that implements the logic for the api.NodesApiServicer
type NodeService struct {
	ModelFieldToDbFieldMap          map[string][]string
	ModelFieldToMetricLabelMap      map[string]string
	tenantSettingsClient            backend_clients.TenantSettingsClient
	workloadsClient                 workloads_clients.WorkloadsClient
	metricsService                  QueryMetricsService
	clusterCapabilityCheckerFactory clients.ClusterCapabilityCheckerFactory
}

func NewNodeService(tenantSettingsClient backend_clients.TenantSettingsClient, workloadsClient workloads_clients.WorkloadsClient) *NodeService {
	err := cache.RegisterStore[metrics.AdvancedGpuMetricsSettingCacheValue](options.Ttl(10 * time.Minute))
	if err != nil {
		logging.Warnf("Failed to register advanced GPU metrics settings cache store: %v", err)
	}

	return &NodeService{
		ModelFieldToDbFieldMap:          createFieldsMap(),
		ModelFieldToMetricLabelMap:      createMetricLabelMap(),
		tenantSettingsClient:            tenantSettingsClient,
		workloadsClient:                 workloadsClient,
		metricsService:                  NewDefaultMetricsService(),
		clusterCapabilityCheckerFactory: clients.LocalClusterCapabilityCheckerFactory,
	}
}

func createFieldsMap() map[string][]string {
	return map[string][]string{
		metrics.ClusterId: {"cluster_uuid"},
		metrics.Nodepool:  {"node_pool"},
		metrics.Node:      {"name"},
	}
}

func createMetricLabelMap() map[string]string {
	return map[string]string{
		metrics.ClusterId: "clusterId",
		metrics.Nodepool:  "nodepool",
		metrics.Node:      "node",
	}
}

func (s *NodeService) DeleteNode(ctx context.Context, clusterUuid, nodeName string) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	_, err = ValidatePermittedScope(ctx, tenantId, clusterUuid)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	err = dao.DeleteNode(ctx, tenantId, clusterUuid, nodeName)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	tlCache := labels_cache.NewNetworkTopologyLabelValuesCache()
	tlCache.Delete(ctx, clusterUuid, nodeName)

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Deleted node <%s> from cluster <%s>", nodeName, clusterUuid)
	return NoContentResponse(), nil
}

func (s *NodeService) GetTenantNodes(ctx context.Context, filterBy []string, sortBy api.NodeFilterSortFields, sortOrder string, offset int32, limit int32, search string) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	permittedScopes, _ := authorization_utils.GetPermittedScopesFromCtx(ctx)

	filters, filterErr := models.NodeQueryParser.Filters(filterBy)
	if filterErr != nil {
		apiErr := api.NewError(http.StatusBadRequest, filterErr.Error(), filterErr)
		return NewErrorResponse(apiErr), nil
	}

	sortColumn, sortError := models.NodeQueryParser.GetMapping(string(sortBy))
	if sortError != nil {
		apiErr := api.NewError(http.StatusBadRequest, sortError.Error(), sortError)
		return NewErrorResponse(apiErr), nil
	}

	nodes, err := dao.GetNodes(ctx, tenantId, permittedScopes, filters, sortColumn, sortOrder, offset, limit, search)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	nodesAdapter := models.NewNodesAdapter(nodes)
	response := nodesAdapter.ToApiModels()

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Got nodes, number of nodes: %d", len(response.Nodes))
	return api.Response(http.StatusOK, response), nil
}

func (s *NodeService) GetNode(ctx context.Context, nodeId string) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	permittedScopes, ok := authorization_utils.GetPermittedScopesFromCtx(ctx)
	if !ok || authorization_utils.IsPermittedScopesEmpty(permittedScopes) {
		return NewErrorResponse(insufficientPermissionsError), nil
	}

	node, dbErr := dao.GetNodeById(ctx, tenantId, nodeId, *permittedScopes)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	numaNodes, numaErr := dao.GetNumaNodes(ctx, node.TenantId, node.ID)
	if numaErr != nil {
		return NewErrorResponse(numaErr), nil
	}

	return api.Response(http.StatusOK, node.ToExtendedApiModel(numaNodes)), nil
}

func (s *NodeService) CountTenantNodes(ctx context.Context, filterBy []string, search string) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	permittedScopes, _ := authorization_utils.GetPermittedScopesFromCtx(ctx)

	filters, filterErr := models.NodeQueryParser.Filters(filterBy)
	if filterErr != nil {
		apiErr := api.NewError(http.StatusBadRequest, filterErr.Error(), filterErr)
		return NewErrorResponse(apiErr), nil
	}

	count, dbErr := dao.CountNodes(ctx, tenantId, permittedScopes, filters, search)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Count tenant nodes request failed. error: %v", dbErr)
		return NewErrorResponse(dbErr), nil
	}

	response := api.CountNodes200Response{
		Count: count,
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Counted tenant nodes, count: %d", count)
	return api.Response(http.StatusOK, response), nil
}

func (s *NodeService) CountNodes(ctx context.Context, clusterUuid string, filterBy []string, search string) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	permittedScopes, err := ValidatePermittedScope(ctx, tenantId, clusterUuid)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	filters, filterErr := models.NodeQueryParser.Filters(filterBy)
	if filterErr != nil {
		apiErr := api.NewError(http.StatusBadRequest, filterErr.Error(), filterErr)
		return NewErrorResponse(apiErr), nil
	}

	count, dbErr := dao.CountNodesByClusterUuid(ctx, tenantId, clusterUuid, *permittedScopes, filters, search)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Count nodes request failed. error: %v", dbErr)
		return NewErrorResponse(dbErr), nil
	}

	response := api.CountNodes200Response{
		Count: count,
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *NodeService) SyncNodes(ctx context.Context, clusterUuid string, nodesSyncRequest api.NodesSyncRequest) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	_, err = ValidatePermittedScope(ctx, tenantId, clusterUuid)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	nodes := []models.Node{}
	for _, nodeForSync := range nodesSyncRequest.NodesForSync {
		if nodeForSync.NodeInfo.NodePool == ExcludedNodepoolName {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf(
				"Node <%s> in cluster <%s> should be excluded and not persisted", nodeForSync.Name, clusterUuid)
			continue
		}
		node := dao.NodeInfoToNodeDao(tenantId, clusterUuid, nodeForSync.Name, nodeForSync.NodeInfo)
		nodes = append(nodes, *node)
	}

	upsertResults, err := dao.UpsertNodes(ctx, &nodes, clusterUuid)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	s.updatePodsForChangedNodes(ctx, clusterUuid, upsertResults)

	err = dao.DeleteUnreferencedNodes(ctx, &nodes, tenantId, clusterUuid, nodesSyncRequest.From, nodesSyncRequest.To)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Finished sync nodes in cluster <%s>, number of nodes: %d",
		clusterUuid, len(nodesSyncRequest.NodesForSync))
	return NoContentResponse(), nil
}

func (s *NodeService) SyncNumaNodes(
	ctx context.Context,
	clusterUuid string,
	request api.NumaNodesSyncRequest,
) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	_, err = ValidatePermittedScope(ctx, tenantId, clusterUuid)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	clusterId, parseErr := uuid.Parse(clusterUuid)
	if parseErr != nil {
		return NewErrorResponse(api.NewError(
			http.StatusBadRequest,
			fmt.Sprintf("Invalid cluster UUID <%s>", clusterUuid),
			parseErr,
		)), nil
	}

	if request.DeleteMissing && request.From != "" && request.To != "" && request.From > request.To {
		return NewErrorResponse(api.NewError(
			http.StatusBadRequest,
			fmt.Sprintf(
				"Invalid NUMA sync range: from <%s> must not be greater than to <%s>",
				request.From,
				request.To,
			),
			nil,
		)), nil
	}

	_, failedNodes, syncErr := dao.SyncNumaNodes(
		ctx,
		NumaNodesSyncRequestToSnapshots(request),
		tenantId,
		clusterId,
		request.DeleteMissing,
		request.From,
		request.To,
	)
	if syncErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(
			"Failed to sync NUMA nodes in cluster <%s>: %v", clusterUuid, syncErr)
		return NewErrorResponse(syncErr), nil
	}

	if len(failedNodes) > 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf(
			"Failed to sync NUMA nodes <%s> in cluster <%s>",
			strings.Join(failedNodes, ", "), clusterUuid)
		return NewErrorResponse(api.NewError(
			http.StatusInternalServerError,
			fmt.Sprintf("Failed to sync NUMA nodes in cluster <%s>", clusterUuid),
			nil,
		)), nil
	}

	return NoContentResponse(), nil
}

func NumaNodesSyncRequestToSnapshots(request api.NumaNodesSyncRequest) []dao.NodeNumaSnapshot {
	snapshots := make([]dao.NodeNumaSnapshot, 0, len(request.Nodes))
	for _, node := range request.Nodes {
		numaNodes := make([]models.NodeNuma, 0, len(node.NumaNodes))
		for _, numaNode := range node.NumaNodes {
			numaNodes = append(numaNodes, models.NodeNuma{
				Name:                    numaNode.Name,
				TotalCpuCores:           numaNode.TotalCpuCores,
				AllocatedCpuCores:       numaNode.AllocatedCpuCores,
				TotalCpuMemoryBytes:     numaNode.TotalCpuMemoryBytes,
				AllocatedCpuMemoryBytes: numaNode.AllocatedCpuMemoryBytes,
				TotalGpus:               numaNode.TotalGpus,
				AllocatedGpus:           numaNode.AllocatedGpus,
			})
		}
		snapshots = append(snapshots, dao.NodeNumaSnapshot{
			NodeName:  node.Name,
			NumaNodes: numaNodes,
		})
	}
	return snapshots
}

func (s *NodeService) UpsertNode(ctx context.Context, clusterUuid, nodeName string, nodeInfo api.NodeInfo) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	_, err = ValidatePermittedScope(ctx, tenantId, clusterUuid)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	node := dao.NodeInfoToNodeDao(tenantId, clusterUuid, nodeName, nodeInfo)

	upsertNodeResult, err := dao.UpsertNode(ctx, node)
	if err != nil {
		return NewErrorResponse(err), nil
	}

	s.updatePodsForChangedNodes(ctx, clusterUuid, []dao.UpsertNodeResult{*upsertNodeResult})

	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Upserted node <%s> in cluster <%s>, status: %s", nodeName, clusterUuid, nodeInfo.Status)
	return NoContentResponse(), nil
}

func (s *NodeService) GetNodeMetrics(ctx context.Context, nodeId string, metricTypes []api.NodeMetricType, start time.Time, end time.Time, numberOfSamples int32) (api.ImplResponse, error) {
	apiErr := metrics.ValidateRequestedMetrics(len(metricTypes))
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	tenantId, apiErr := ParseTenantId(ctx)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	apiErr = metrics.ValidateRequestedAdvancedMetrics(ctx, metricTypes)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	permittedScopes, ok := authorization_utils.GetPermittedScopesFromCtx(ctx)
	if !ok || authorization_utils.IsPermittedScopesEmpty(permittedScopes) {
		return NewErrorResponse(insufficientPermissionsError), nil
	}

	node, dbErr := dao.GetNodeById(ctx, tenantId, nodeId, *permittedScopes)
	if dbErr != nil {
		return NewErrorResponse(dbErr), nil
	}

	_, apiErr = ValidatePermittedClusterScope(ctx, node.ClusterUuid)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	measurements := []api.MeasurementResponse{}
	responseCode := http.StatusOK

	for _, metricType := range metricTypes {
		result := s.metricsService.QueryNodeMetricType(ctx, node.ClusterUuid, node.Name, start, end, metricType, numberOfSamples)
		if result == nil {
			responseCode = http.StatusMultiStatus
			measurements = append(measurements, api.MeasurementResponse{
				Type: string(metricType),
			})

			continue
		}

		measurement := metrics.ParseNodeMetricResultToMeasurements(ctx, string(metricType), *result)
		measurements = append(measurements, measurement...)
	}

	responseStruct := api.MetricsResponse{Measurements: measurements}
	accept := runaictx.GetString(ctx, runaictx.CfAccept)
	if accept != nil && *accept == CsvAccept {
		data := GetMetricsCsvFromMetricsResponse(&responseStruct, metricTypes)
		response := runaihttp.CsvFile{Data: data, FileName: "node-metrics.csv"}
		return api.Response(http.StatusOK, response), nil
	}

	return api.Response(responseCode, responseStruct), nil
}

// GetNodeTelemetry
// FREE_GPUS - this telemetry is only fetched from metrics, for all cluster versions.
// TOTAL_GPU_NODES, TOTAL_GPUS - is fetched from the db for clusters with version >= 2.15, otherwise from metrics.
// READY_GPU_NODES, READY_GPUS - is fetched from the db for clusters with version >= 2.15,
// otherwise - not implemented! this data doesn't exist even in metrics for older clusters.
func (s *NodeService) GetNodeTelemetry(ctx context.Context, telemetryType api.NodeTelemetryType, clusterId string, nodepoolName string, groupBy []string, tenantIds []string) (api.ImplResponse, error) {
	err := validateNodepoolFilterGroups(clusterId, nodepoolName, groupBy, metrics.ClusterId, metrics.Nodepool)
	if err != nil {
		apiErr := api.NewError(http.StatusBadRequest, err.Error(), err)
		return NewErrorResponse(apiErr), nil
	}
	if _, found := allSupportedTypesMap[telemetryType]; !found {
		apiErr := api.NewError(http.StatusBadRequest,
			fmt.Sprintf("telemetry type %s is invalid. supported types are: %s",
				string(telemetryType), strings.Join(allSupportedTypes, ", ")), nil)
		return NewErrorResponse(apiErr), nil
	}

	if runaictx.IsManagementTenant(ctx) {
		return s.GetNodeTelemetryMgmt(ctx, telemetryType, clusterId, nodepoolName, groupBy, tenantIds)
	}

	permittedClusters, scopeErr := GetPermittedClusterScopes(ctx)
	if scopeErr != nil {
		return NewErrorResponse(scopeErr), nil
	}
	if permittedClusters[0] == "all" {
		permittedClusters = nil
	}

	if _, found := metricsOnlyTelemetryTypes[telemetryType]; found {
		return s.getNodeTelemetryFromMetrics(ctx, telemetryType, clusterId, nodepoolName, groupBy, permittedClusters)
	}

	dbTelemetrySupported, apiErr := s.isDbTelemetrySupportedForScopes(ctx, clusterId, permittedClusters)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}

	if dbTelemetrySupported {
		permittedScopes, _ := authorization_utils.GetPermittedScopesFromCtx(ctx)
		return s.getNodeTelemetryFromDb(ctx, telemetryType, []string{}, clusterId, nodepoolName, groupBy, permittedScopes)
	}

	if _, found := dbOnlyTelemetryTypes[telemetryType]; found {
		apiErr = api.NewError(http.StatusBadRequest,
			fmt.Sprintf("telemetry type %s is not supported for the given cluster versions", string(telemetryType)), nil)
		return NewErrorResponse(apiErr), nil
	}

	return s.getNodeTelemetryFromMetrics(ctx, telemetryType, clusterId, nodepoolName, groupBy, permittedClusters)
}

// TODO - remove since nodepools are in the service
func validateNodepoolFilterGroups(clusterFilter string, nodepoolFilter string, groups []string, clusterGroupName string, nodepoolGroupName string) error {
	if clusterFilter == "" {
		if nodepoolFilter != "" {
			return fmt.Errorf("filter of nodepools %s without cluster filter is not allowed", nodepoolFilter)
		}
		firstGroup := ""
		if len(groups) > 0 {
			firstGroup = groups[0]
		}
		secondGroup := ""
		if len(groups) > 1 {
			secondGroup = groups[1]
		}
		if (firstGroup == nodepoolGroupName && secondGroup != clusterGroupName) ||
			(secondGroup == nodepoolGroupName && firstGroup != clusterGroupName) {
			return fmt.Errorf("group by nodepools %s without cluster filter or cluster groups is not allowed", nodepoolGroupName)
		}

	}
	return nil
}

func (s *NodeService) GetNodeTelemetryMgmt(ctx context.Context, telemetryType api.NodeTelemetryType, clusterId string, nodepoolName string, groupBy []string, tenantIds []string) (api.ImplResponse, error) {
	permittedScopes, ok := authorization_utils.GetPermittedScopesFromCtx(ctx)
	if !ok || authorization_utils.IsPermittedScopesEmpty(permittedScopes) {
		return NewErrorResponse(insufficientPermissionsError), nil
	}

	if _, found := metricsOnlyTelemetryTypes[telemetryType]; found {
		return NewErrorResponse(api.NewError(http.StatusNotImplemented, fmt.Sprintf("telemetry trype %s is not suppoted for management tenant", telemetryType), nil)), nil
	}

	if clusterId != "" {
		dbTelemetrySupported, apiErr := s.isDbTelemetrySupportedForClusters(ctx, []string{clusterId})
		if apiErr != nil {
			return NewErrorResponse(apiErr), nil
		}

		if !dbTelemetrySupported {
			apiErr = api.NewError(http.StatusBadRequest,
				fmt.Sprintf("telemetry type %s is not supported for the given cluster %s version", telemetryType, clusterId), nil)
			return NewErrorResponse(apiErr), nil
		}
	}

	return s.getNodeTelemetryFromDb(ctx, telemetryType, tenantIds, clusterId, nodepoolName, groupBy, permittedScopes)
}

func (s *NodeService) isDbTelemetrySupportedForScopes(ctx context.Context, clusterId string, permittedClusters []string) (bool, *api.Error) {
	var clusters []string
	if clusterId != "" {
		clusters = []string{clusterId}
	} else if permittedClusters != nil {
		clusters = permittedClusters
	}

	return s.isDbTelemetrySupportedForClusters(ctx, clusters)
}

func (s *NodeService) isDbTelemetrySupportedForClusters(ctx context.Context, clusters []string) (bool, *api.Error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return false, err
	}
	clusterVersions, err := dao.GetClusterVersions(ctx, tenantId, clusters)
	if err != nil {
		return false, err
	}
	if len(clusterVersions) == 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("no clusters returned for tenant %d , clusters %s", tenantId, clusters)
		return false, nil
	}

	// in case one of the clusters is not connected and doesn't have a version - ignore it;
	// it will not have info in the db anyway.
	nonEmptyClusterVersions := []string{}
	for _, v := range clusterVersions {
		if v != "" {
			nonEmptyClusterVersions = append(nonEmptyClusterVersions, v)
		}
	}

	result, innerErr := versions.AreAllVersionsSupported(nonEmptyClusterVersions, SupportedDbTelemetryVersion)
	if innerErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to check if telemetry is supported for cluster versions, error: %v", innerErr)
		return false, nil
	}

	return result, nil
}

func (s *NodeService) getNodeTelemetryFromDb(ctx context.Context, telemetryType api.NodeTelemetryType, tenantIds []string, clusterId string, nodepoolName string, groupBy []string,
	permittedScopes *authorization_api.PermittedScopes,
) (api.ImplResponse, error) {
	telemetryGroupsApi, apiErr := s.getNodeTelemetryFromDbInner(ctx, telemetryType, tenantIds, clusterId, nodepoolName, groupBy, permittedScopes)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	accept := runaictx.GetString(ctx, runaictx.CfAccept)
	if accept != nil && *accept == CsvAccept {
		data := GetTelemetryCsv(telemetryGroupsApi)
		response := runaihttp.CsvFile{Data: data}
		loc, err := time.LoadLocation("UTC")
		if err == nil {
			response.FileName = fmt.Sprintf("%s-%s.csv", string(telemetryType), telemetryGroupsApi.Timestamp.In(loc).Format("2006-01-02_15:04:05"))
		} else {
			response.FileName = fmt.Sprintf("%s.csv", string(telemetryType))
		}
		return api.Response(http.StatusOK, response), nil
	}
	return api.Response(http.StatusOK, telemetryGroupsApi), nil
}

func (s *NodeService) getNodeTelemetryFromDbInner(ctx context.Context, telemetryType api.NodeTelemetryType,
	tenantIds []string, clusterId string, nodepoolName string, groupBy []string, permittedScopes *authorization_api.PermittedScopes,
) (*api.TelemetryResponse, *api.Error) {
	tenantId, parsingErr := ParseTenantId(ctx)
	if parsingErr != nil {
		return nil, parsingErr
	}

	modelGroupBy, err := telemetry_utils.GetModelGroupBy(groupBy, s.ModelFieldToDbFieldMap)
	if err != nil {
		return nil, api.NewError(http.StatusBadRequest, err.Error(), err)
	}

	var telemetryGroups *[]models.TelemetryGroupValue
	var dbErr *api.Error

	timestamp := time.Now()
	switch telemetryType {
	case api.NODETELEMETRYTYPE_READY_GPU_NODES:
		telemetryGroups, dbErr = dao.GetGpuNodesCount(ctx, tenantId, tenantIds, clusterId, nodepoolName, permittedScopes, modelGroupBy, true)
	case api.NODETELEMETRYTYPE_READY_GPUS:
		telemetryGroups, dbErr = dao.GetGpuCount(ctx, tenantId, tenantIds, clusterId, nodepoolName, permittedScopes, modelGroupBy, true)
	case api.NODETELEMETRYTYPE_TOTAL_GPU_NODES:
		telemetryGroups, dbErr = dao.GetGpuNodesCount(ctx, tenantId, tenantIds, clusterId, nodepoolName, permittedScopes, modelGroupBy, false)
	case api.NODETELEMETRYTYPE_TOTAL_GPUS:
		telemetryGroups, dbErr = dao.GetGpuCount(ctx, tenantId, tenantIds, clusterId, nodepoolName, permittedScopes, modelGroupBy, false)
	default:
		// shouldn't get here - validated before
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Unknown node telemetry metric type %v.", telemetryType)
		return nil, nil
	}
	if dbErr != nil {
		return nil, dbErr
	}

	telemetryGroupsApi := TelemetryGroupsToApi(string(telemetryType), timestamp, *telemetryGroups, groupBy)
	return telemetryGroupsApi, nil
}

func (s *NodeService) getNodeTelemetryFromMetrics(ctx context.Context, telemetryType api.NodeTelemetryType, clusterId string, nodepoolName string, groupBy []string, permittedClusters []string) (api.ImplResponse, error) {
	if permittedClusters != nil && clusterId != "" && !slices.Contains(permittedClusters, clusterId) {
		// user is not permitted to see results from this cluster, return empty response
		return api.Response(http.StatusOK, &api.TelemetryResponse{
			Type:      string(telemetryType),
			Timestamp: time.Now(),
			Values:    []api.TelemetryResponseValuesInner{},
		}), nil
	}
	_, err := telemetry_utils.GetModelGroupBy(groupBy, s.ModelFieldToDbFieldMap)
	if err != nil {
		return NewErrorResponse(api.NewError(http.StatusBadRequest, err.Error(), err)), nil
	}

	telemetryGroupsApi, apiErr := s.getNodeTelemetryFromMetricsInner(ctx, telemetryType, clusterId, nodepoolName, groupBy, permittedClusters)
	if apiErr != nil {
		return NewErrorResponse(apiErr), nil
	}
	accept := runaictx.GetString(ctx, runaictx.CfAccept)
	if accept != nil && *accept == CsvAccept {
		data := GetTelemetryCsv(telemetryGroupsApi)
		response := runaihttp.CsvFile{Data: data}
		loc, err := time.LoadLocation("UTC")
		if err == nil {
			response.FileName = fmt.Sprintf("%s-%s.csv", string(telemetryType), telemetryGroupsApi.Timestamp.In(loc).Format("2006-01-02_15:04:05"))
		} else {
			response.FileName = fmt.Sprintf("%s.csv", string(telemetryType))
		}
		return api.Response(http.StatusOK, response), nil
	}
	return api.Response(http.StatusOK, telemetryGroupsApi), nil
}

func (s *NodeService) getNodeTelemetryFromMetricsInner(ctx context.Context, telemetryType api.NodeTelemetryType, clusterId string,
	nodepoolName string, groupBy []string, permittedClusters []string,
) (*api.TelemetryResponse, *api.Error) {
	clusterIds := []string{}
	if clusterId != "" {
		clusterIds = []string{clusterId}
	} else if len(permittedClusters) > 0 {
		clusterIds = permittedClusters
	}

	timestamp := time.Now()

	result, err := s.metricsService.QueryNodeMetric(ctx, clusterIds, nodepoolName, telemetryType, groupBy)
	if err != nil {
		if strings.Contains(err.Error(), "400 Bad Request") {
			return nil, api.NewError(http.StatusBadRequest, "Requested metric is not supported", err)
		}
		return nil, api.NewError(http.StatusInternalServerError, "failed to get metric result", err)
	}

	telemetryGroupsApi := metrics.ParseNodeMetricResultToTelemetryResponse(ctx, telemetryType, *result, timestamp, groupBy, s.ModelFieldToMetricLabelMap)
	return telemetryGroupsApi, nil
}

// updatePodsForChangedNodes calls workloads service to update pods for nodes
// whose NetworkTopologyLabelValues have changed.
func (s *NodeService) updatePodsForChangedNodes(ctx context.Context, clusterUuid string, upsertResults []dao.UpsertNodeResult) {
	if s.workloadsClient == nil {
		return
	}

	clusterCapabilityChecker := s.clusterCapabilityCheckerFactory(clusterUuid)
	networkTopologyLabelsSupported, _ := clusterCapabilityChecker.ValidateClusterCapability(ctx, checker.MinVersionTopologyAddress)
	if !networkTopologyLabelsSupported {
		return
	}

	const maxConcurrency = 10
	sem := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup

	for _, result := range upsertResults {
		if !networkTopologyLabelValuesChanged(result.PreviousValues.NetworkTopologyLabelValues, result.NewValues.NetworkTopologyLabelValues) {
			continue
		}
		wg.Add(1)
		sem <- struct{}{} // acquire semaphore
		go func(result dao.UpsertNodeResult) {
			defer wg.Done()
			defer func() { <-sem }() // release semaphore
			s.updateNodePodsNetworkTopologyLabelValues(ctx, clusterUuid, result)
		}(result)
	}

	wg.Wait()
}

// updateNodePodsNetworkTopologyLabelValues updates pods for a single node if its topology labels changed.
func (s *NodeService) updateNodePodsNetworkTopologyLabelValues(ctx context.Context, clusterUuid string, result dao.UpsertNodeResult) {
	var newValues []string
	nodeName := result.Name
	if result.NewValues.NetworkTopologyLabelValues != nil {
		newValues = *result.NewValues.NetworkTopologyLabelValues
	}

	if err := s.workloadsClient.InternalUpdateNodePods(ctx, clusterUuid, nodeName, newValues); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to update pods for node <%s> in cluster <%s>: %v",
			nodeName, clusterUuid, err)
	}

	tlCache := labels_cache.NewNetworkTopologyLabelValuesCache()
	tlCache.Set(ctx, clusterUuid, nodeName, newValues)
}

// networkTopologyLabelValuesChanged compares two pq.StringArray pointers and returns true if they differ.
// both prev and new arrays are expected to be sorted according to the topology level position
func networkTopologyLabelValuesChanged(prev, new *pq.StringArray) bool {
	if prev == nil && new == nil {
		return false
	}
	if (prev != nil && new == nil) || prev == nil {
		return true
	}
	prevSlice := []string(*prev)
	newSlice := []string(*new)
	if len(prevSlice) != len(newSlice) {
		return true
	}
	for i := range prevSlice {
		if prevSlice[i] != newSlice[i] {
			return true
		}
	}
	return false
}
