/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"time"

	metrics2 "github.com/run-ai/runai/control-plane/cluster-service/pkg/services/metrics"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/run-ai/runai/control-plane/cluster-service/pkg/runaimetrics"
	metricsClient "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
	openapi "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"

	org_unit_service_client "github.com/run-ai/runai/runai-clients/org-unit-service-client"
)

// QueryMetricsService defines the interface for querying metrics
type QueryMetricsService interface {
	QueryClusterMetricType(ctx context.Context, orgUnitClient org_unit_service_client.IOrgUnitClient,
		clusterId string, start time.Time, end time.Time, metricType api.MetricsType, groupBy metricsClient.MetricGroupBy, numberOfSamples int32, nodePoolName string) *openapi.GetMetrics200Response

	QueryNodepoolMetricType(ctx context.Context, orgUnitClient org_unit_service_client.IOrgUnitClient,
		clusterId string, nodepoolName string, start time.Time, end time.Time, metricType api.MetricsType, numberOfSamples int32) *openapi.GetMetrics200Response

	QueryNodeMetricType(ctx context.Context, clusterId string, nodeName string, start time.Time, end time.Time,
		metricType api.NodeMetricType, numberOfSamples int32) *openapi.GetMetrics200Response

	QueryNodeMetric(ctx context.Context, clusterIds []string, nodePoolName string, nodeTelemetryType api.NodeTelemetryType,
		groupBy []string) (*openapi.GetMetrics200Response, error)
}

// DefaultMetricsService is the default implementation of QueryMetricsService
type DefaultMetricsService struct{}

// NewDefaultMetricsService creates a new instance of DefaultMetricsService
func NewDefaultMetricsService() *DefaultMetricsService {
	return &DefaultMetricsService{}
}

func (s *DefaultMetricsService) QueryClusterMetricType(ctx context.Context, orgUnitClient org_unit_service_client.IOrgUnitClient,
	clusterId string, start time.Time, end time.Time, metricType api.MetricsType, groupBy metricsClient.MetricGroupBy, numberOfSamples int32, nodePoolName string,
) *openapi.GetMetrics200Response {
	return QueryClusterMetricType(ctx, orgUnitClient, clusterId, start, end, metricType, groupBy, numberOfSamples, nodePoolName)
}

func (s *DefaultMetricsService) QueryNodepoolMetricType(ctx context.Context, orgUnitClient org_unit_service_client.IOrgUnitClient,
	clusterId string, nodepoolName string, start time.Time, end time.Time, metricType api.MetricsType, numberOfSamples int32,
) *openapi.GetMetrics200Response {
	return QueryNodepoolMetricType(ctx, orgUnitClient, clusterId, nodepoolName, start, end, metricType, numberOfSamples)
}

func (s *DefaultMetricsService) QueryNodeMetricType(ctx context.Context, clusterId string, nodeName string, start time.Time, end time.Time,
	metricType api.NodeMetricType, numberOfSamples int32,
) *openapi.GetMetrics200Response {
	return QueryNodeMetricType(ctx, clusterId, nodeName, start, end, metricType, numberOfSamples)
}

func (s *DefaultMetricsService) QueryNodeMetric(ctx context.Context, clusterIds []string, nodePoolName string, nodeTelemetryType api.NodeTelemetryType,
	groupBy []string,
) (*openapi.GetMetrics200Response, error) {
	return QueryNodeMetric(ctx, clusterIds, nodePoolName, nodeTelemetryType, groupBy)
}

func QueryClusterMetricType(ctx context.Context, orgUnitClient org_unit_service_client.IOrgUnitClient,
	clusterId string, start time.Time, end time.Time, metricType api.MetricsType, groupBy metricsClient.MetricGroupBy, numberOfSamples int32, nodePoolName string,
) *openapi.GetMetrics200Response {
	var metric metricsClient.Metric
	clusterFilter := []uuid.UUID{uuid.MustParse(clusterId)}
	switch metricType {
	case api.METRICSTYPE_GPU_UTILIZATION:
		metric = runaimetrics.RunaiGpuUtilizationPerClusterMetric{
			Filters: runaimetrics.RunaiGpuUtilizationPerClusterFilters{
				ClusterId: clusterFilter,
			},
		}
	case api.METRICSTYPE_GPU_MEMORY_UTILIZATION:
		metric = runaimetrics.RunaiGpuMemoryUtilizationPerClusterMetric{
			Filters: runaimetrics.RunaiGpuMemoryUtilizationPerClusterFilters{
				ClusterId: clusterFilter,
			},
		}
	case api.METRICSTYPE_CPU_UTILIZATION:
		metric = runaimetrics.RunaiCpuUtilizationPerClusterMetric{
			Filters: runaimetrics.RunaiCpuUtilizationPerClusterFilters{
				ClusterId: clusterFilter,
			},
		}
	case api.METRICSTYPE_CPU_MEMORY_UTILIZATION:
		metric = runaimetrics.RunaiMemoryUtilizationPerClusterMetric{
			Filters: runaimetrics.RunaiMemoryUtilizationPerClusterFilters{
				ClusterId: clusterFilter,
			},
		}
	case api.METRICSTYPE_ALLOCATED_GPU:
		metric = runaimetrics.RunaiAllocatedGpuCountPerClusterMetric{
			Filters: runaimetrics.RunaiAllocatedGpuCountPerClusterFilters{
				ClusterId: clusterFilter,
			},
		}
	case api.METRICSTYPE_AVG_WORKLOAD_WAIT_TIME:
		metric = runaimetrics.RunaiAvgWorkloadWaitTimeSecondsPerClusterMetric{
			Filters: runaimetrics.RunaiAvgWorkloadWaitTimeSecondsPerClusterFilters{
				ClusterId: clusterFilter,
			},
		}
	case api.METRICSTYPE_TOTAL_GPU:
		metric = runaimetrics.RunaiGpuCountPerClusterMetric{
			Filters: runaimetrics.RunaiGpuCountPerClusterFilters{
				ClusterId: clusterFilter,
			},
		}
	case api.METRICSTYPE_GPU_QUOTA:
		if useDepartmentsMetrics(ctx, orgUnitClient, clusterId) {
			metric = runaimetrics.RunaiDepartmentsGpuQuotaPerClusterMetric{
				Filters: runaimetrics.RunaiDepartmentsGpuQuotaPerClusterFilters{
					ClusterId: clusterFilter,
				},
			}
		} else {
			metric = runaimetrics.RunaiProjectsGpuQuotaPerClusterMetric{
				Filters: runaimetrics.RunaiProjectsGpuQuotaPerClusterFilters{
					ClusterId: clusterFilter,
				},
			}
		}
	case api.METRICSTYPE_WORKLOADS_COUNT:
		metric = metrics2.GetWorkloadsCountMetric(clusterFilter, groupBy, nodePoolName)

	default:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Unknown metric type %s.", metricType)
		return nil
	}

	return submitMetric(ctx, metric, string(metricType), start, end, numberOfSamples)
}

func QueryNodepoolMetricType(
	ctx context.Context,
	orgUnitClient org_unit_service_client.IOrgUnitClient,
	clusterId string,
	nodepoolName string,
	start time.Time,
	end time.Time,
	metricType api.MetricsType,
	numberOfSamples int32,
) *openapi.GetMetrics200Response {
	var metric metricsClient.Metric
	clusterFilter := []uuid.UUID{uuid.MustParse(clusterId)}
	NodepoolFilter := []string{nodepoolName}
	switch metricType {
	case api.METRICSTYPE_TOTAL_GPU_NODES:
		metric = runaimetrics.RunaiGpuNodesCountPerNodepoolMetric{
			Filters: runaimetrics.RunaiGpuNodesCountPerNodepoolFilters{
				ClusterId: clusterFilter,
				Nodepool:  NodepoolFilter,
			},
		}
	case api.METRICSTYPE_GPU_UTILIZATION:
		metric = runaimetrics.RunaiGpuUtilizationPerNodepoolMetric{
			Filters: runaimetrics.RunaiGpuUtilizationPerNodepoolFilters{
				ClusterId: clusterFilter,
				Nodepool:  NodepoolFilter,
			},
		}
	case api.METRICSTYPE_GPU_UTILIZATION_DISTRIBUTION:
		metric = runaimetrics.RunaiGpuUtilizationDistributionPerNodepoolMetric{
			Filters: runaimetrics.RunaiGpuUtilizationDistributionPerNodepoolFilters{
				ClusterId: clusterFilter,
				Nodepool:  NodepoolFilter,
			},
		}
	case api.METRICSTYPE_GPU_MEMORY_UTILIZATION:
		metric = runaimetrics.RunaiGpuMemoryUtilizationPerNodepoolMetric{
			Filters: runaimetrics.RunaiGpuMemoryUtilizationPerNodepoolFilters{
				ClusterId: clusterFilter,
				Nodepool:  NodepoolFilter,
			},
		}
	case api.METRICSTYPE_CPU_UTILIZATION:
		metric = runaimetrics.RunaiCpuUtilizationPerNodepoolMetric{
			Filters: runaimetrics.RunaiCpuUtilizationPerNodepoolFilters{
				ClusterId: clusterFilter,
				Nodepool:  NodepoolFilter,
			},
		}
	case api.METRICSTYPE_CPU_MEMORY_UTILIZATION:
		metric = runaimetrics.RunaiMemoryUtilizationPerNodepoolMetric{
			Filters: runaimetrics.RunaiMemoryUtilizationPerNodepoolFilters{
				ClusterId: clusterFilter,
				Nodepool:  NodepoolFilter,
			},
		}
	case api.METRICSTYPE_ALLOCATED_GPU:
		metric = runaimetrics.RunaiAllocatedGpuCountPerNodepoolMetric{
			Filters: runaimetrics.RunaiAllocatedGpuCountPerNodepoolFilters{
				ClusterId: clusterFilter,
				Nodepool:  NodepoolFilter,
			},
		}
	case api.METRICSTYPE_UNALLOCATED_GPU:
		metric = runaimetrics.RunaiUnallocatedGpuCountPerNodepoolMetric{
			Filters: runaimetrics.RunaiUnallocatedGpuCountPerNodepoolFilters{
				ClusterId: clusterFilter,
				Nodepool:  NodepoolFilter,
			},
		}
	case api.METRICSTYPE_TOTAL_GPU:
		metric = runaimetrics.RunaiGpuCountPerNodepoolMetric{
			Filters: runaimetrics.RunaiGpuCountPerNodepoolFilters{
				ClusterId: clusterFilter,
				Nodepool:  NodepoolFilter,
			},
		}
	case api.METRICSTYPE_GPU_QUOTA:
		if useDepartmentsMetrics(ctx, orgUnitClient, clusterId) {
			metric = runaimetrics.RunaiDepartmentsGpuQuotaPerNodepoolMetric{
				Filters: runaimetrics.RunaiDepartmentsGpuQuotaPerNodepoolFilters{
					ClusterId: clusterFilter,
					Nodepool:  NodepoolFilter,
				},
			}
		} else {
			metric = runaimetrics.RunaiProjectsGpuQuotaPerNodepoolMetric{
				Filters: runaimetrics.RunaiProjectsGpuQuotaPerNodepoolFilters{
					ClusterId: clusterFilter,
					Nodepool:  NodepoolFilter,
				},
			}
		}
	case api.METRICSTYPE_AVG_WORKLOAD_WAIT_TIME:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Not implemented metric type %s.", metricType)
		return nil
	default:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Unknown metric type %s.", metricType)
		return nil
	}

	return submitMetric(ctx, metric, string(metricType), start, end, numberOfSamples)
}

func QueryNodeMetricType(ctx context.Context, clusterId string, nodeName string, start time.Time, end time.Time, metricType api.NodeMetricType, numberOfSamples int32) *openapi.GetMetrics200Response {
	var metric metricsClient.Metric
	clusterFilter := []uuid.UUID{uuid.MustParse(clusterId)}
	nodeFilter := []string{nodeName}
	switch metricType {
	case api.NODEMETRICTYPE_GPU_UTILIZATION_PER_GPU:
		metric = runaimetrics.RunaiGpuUtilizationPerGpuMetric{
			Filters: runaimetrics.RunaiGpuUtilizationPerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_UTILIZATION:
		metric = runaimetrics.RunaiGpuUtilizationPerNodeMetric{
			Filters: runaimetrics.RunaiGpuUtilizationPerNodeFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_MEMORY_UTILIZATION_PER_GPU:
		metric = runaimetrics.RunaiGpuMemoryUtilizationPerGpuMetric{
			Filters: runaimetrics.RunaiGpuMemoryUtilizationPerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_MEMORY_UTILIZATION:
		metric = runaimetrics.RunaiGpuMemoryUtilizationPerNodeMetric{
			Filters: runaimetrics.RunaiGpuMemoryUtilizationPerNodeFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_MEMORY_USAGE_BYTES_PER_GPU:
		metric = runaimetrics.RunaiGpuMemoryUsedBytesPerGpuMetric{
			Filters: runaimetrics.RunaiGpuMemoryUsedBytesPerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_MEMORY_USAGE_BYTES:
		metric = runaimetrics.RunaiUsedGpuMemoryBytesPerNodeMetric{
			Filters: runaimetrics.RunaiUsedGpuMemoryBytesPerNodeFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_ALLOCATION_PER_GPU:
		metric = runaimetrics.RunaiAllocatedGpuCountPerGpuMetric{
			Filters: runaimetrics.RunaiAllocatedGpuCountPerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_MEMORY_ALLOCATION_BYTES_PER_GPU:
		metric = runaimetrics.RunaiAllocatedGpuMemoryBytesPerGpuMetric{
			Filters: runaimetrics.RunaiAllocatedGpuMemoryBytesPerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_IDLE_TIME_SECONDS_PER_GPU:
		metric = runaimetrics.RunaiIdleSecondsPerGpuMetric{
			Filters: runaimetrics.RunaiIdleSecondsPerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_CPU_USAGE_CORES:
		metric = runaimetrics.RunaiUsedCpuCoresPerNodeMetric{
			Filters: runaimetrics.RunaiUsedCpuCoresPerNodeFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_CPU_UTILIZATION:
		metric = runaimetrics.RunaiCpuUtilizationPerNodeMetric{
			Filters: runaimetrics.RunaiCpuUtilizationPerNodeFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_CPU_MEMORY_USAGE_BYTES:
		metric = runaimetrics.RunaiUsedMemoryBytesPerNodeMetric{
			Filters: runaimetrics.RunaiUsedMemoryBytesPerNodeFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_CPU_MEMORY_UTILIZATION:
		metric = runaimetrics.RunaiMemoryUtilizationPerNodeMetric{
			Filters: runaimetrics.RunaiMemoryUtilizationPerNodeFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_OOMKILL_BURST_COUNT_PER_GPU:
		metric = runaimetrics.RunaiGpuOomkillBurstCountPerGpuMetric{
			Filters: runaimetrics.RunaiGpuOomkillBurstCountPerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_OOMKILL_IDLE_COUNT_PER_GPU:
		metric = runaimetrics.RunaiGpuOomkillIdleCountPerGpuMetric{
			Filters: runaimetrics.RunaiGpuOomkillIdleCountPerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_OOMKILL_SWAP_OUT_OF_RAM_COUNT_PER_GPU:
		metric = runaimetrics.RunaiGpuOomkillSwapOutOfRamCountPerGpuMetric{
			Filters: runaimetrics.RunaiGpuOomkillSwapOutOfRamCountPerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_GRAPHICS_ENGINE_ACTIVITY_PER_GPU:
		metric = runaimetrics.RunaiGpuGREngineActivePerGpuMetric{
			Filters: runaimetrics.RunaiGpuGREngineActivePerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_SM_ACTIVITY_PER_GPU:
		metric = runaimetrics.RunaiGpuSmActivePerGpuMetric{
			Filters: runaimetrics.RunaiGpuSmActivePerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_SM_OCCUPANCY_PER_GPU:
		metric = runaimetrics.RunaiGpuSmOccupancyPerGpuMetric{
			Filters: runaimetrics.RunaiGpuSmOccupancyPerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_TENSOR_ACTIVITY_PER_GPU:
		metric = runaimetrics.RunaiGpuPipeTensorActivePerGpuMetric{
			Filters: runaimetrics.RunaiGpuPipeTensorActivePerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_FP64_ENGINE_ACTIVITY_PER_GPU:
		metric = runaimetrics.RunaiGpuPipeFp64ActivePerGpuMetric{
			Filters: runaimetrics.RunaiGpuPipeFp64ActivePerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_FP32_ENGINE_ACTIVITY_PER_GPU:
		metric = runaimetrics.RunaiGpuPipeFp32ActivePerGpuMetric{
			Filters: runaimetrics.RunaiGpuPipeFp32ActivePerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_FP16_ENGINE_ACTIVITY_PER_GPU:
		metric = runaimetrics.RunaiGpuPipeFp16ActivePerGpuMetric{
			Filters: runaimetrics.RunaiGpuPipeFp16ActivePerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_MEMORY_BANDWIDTH_UTILIZATION_PER_GPU:
		metric = runaimetrics.RunaiGpuDramActivePerGpuMetric{
			Filters: runaimetrics.RunaiGpuDramActivePerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_NVLINK_TRANSMITTED_BANDWIDTH_PER_GPU:
		metric = runaimetrics.RunaiGpuNvlinkTxBytesPerGpuMetric{
			Filters: runaimetrics.RunaiGpuNvlinkTxBytesPerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_NVLINK_RECEIVED_BANDWIDTH_PER_GPU:
		metric = runaimetrics.RunaiGpuNvlinkRxBytesPerGpuMetric{
			Filters: runaimetrics.RunaiGpuNvlinkRxBytesPerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_PCIE_TRANSMITTED_BANDWIDTH_PER_GPU:
		metric = runaimetrics.RunaiGpuPcieTxBytesPerGpuMetric{
			Filters: runaimetrics.RunaiGpuPcieTxBytesPerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_GPU_PCIE_RECEIVED_BANDWIDTH_PER_GPU:
		metric = runaimetrics.RunaiGpuPcieRxBytesPerGpuMetric{
			Filters: runaimetrics.RunaiGpuPcieRxBytesPerGpuFilters{
				ClusterId: clusterFilter,
				Node:      nodeFilter,
			},
		}
	case api.NODEMETRICTYPE_NVLINK_BANDWIDTH_TOTAL:
		metric = metrics2.GetDcgmNvlinkBandwidthTotalBytesMetric(clusterFilter, "", nodeFilter)
	case api.NODEMETRICTYPE_NVLINK_BANDWIDTH_TOTAL_PER_GPU:
		metric = metrics2.GetDcgmNvlinkBandwidthTotalBytesMetric(clusterFilter, metricsClient.GroupByGpu, nodeFilter)
	default:
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Unknown metric type %s.", metricType)
		return nil
	}

	return submitMetric(ctx, metric, string(metricType), start, end, numberOfSamples)
}

func submitMetric(ctx context.Context, metric metricsClient.Metric, metricType string, start time.Time, end time.Time, numberOfSamples int32) *openapi.GetMetrics200Response {
	q, err := metricsClient.NewMetricQuery(metric)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to construct query for metric type %s. err: %v", metricType, err)
		return nil
	}

	q.SetRange(start, end)
	q.SetDefaultNumOfSamples(numberOfSamples)
	res, err := q.Query(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to query metric type %s. err: %v", metricType, err)
		return nil
	}
	return res
}

func QueryNodeMetric(ctx context.Context, clusterIds []string, nodePoolName string, nodeTelemetryType api.NodeTelemetryType, groupBy []string) (*openapi.GetMetrics200Response, error) {
	clusterIdsFilter := []uuid.UUID{}
	for _, clusterId := range clusterIds {
		clusterIdsFilter = append(clusterIdsFilter, uuid.MustParse(clusterId))
	}
	nodePoolFilter := []string{}
	if nodePoolName != "" {
		nodePoolFilter = []string{nodePoolName}
	}

	// we will use the metric per the "lowest" scope
	metricGroupBy := getMetricGroupBy(groupBy, len(nodePoolFilter) > 0)

	var metric metricsClient.Metric
	switch nodeTelemetryType {
	case api.NODETELEMETRYTYPE_IDLE_ALLOCATED_GPUS:
		metric = metrics2.GetIdleAllocatedGpusMetric(clusterIdsFilter, nodePoolFilter, metricGroupBy)
	case api.NODETELEMETRYTYPE_FREE_GPUS:
		metric = metrics2.GetFreeGpusMetric(clusterIdsFilter, nodePoolFilter, metricGroupBy)
	case api.NODETELEMETRYTYPE_TOTAL_GPU_NODES:
		metric = metrics2.GetTotalGpuNodesMetric(clusterIdsFilter, nodePoolFilter, metricGroupBy)
	case api.NODETELEMETRYTYPE_TOTAL_GPUS:
		metric = metrics2.GetTotalGpusMetric(clusterIdsFilter, nodePoolFilter, metricGroupBy)

	case api.NODETELEMETRYTYPE_ALLOCATED_GPUS:
		metric = metrics2.GetAllocatedGpusMetric(clusterIdsFilter, nodePoolFilter, metricGroupBy)
	case api.NODETELEMETRYTYPE_TOTAL_CPU_CORES:
		metric = metrics2.GetTotalCpusMetric(clusterIdsFilter, nodePoolFilter, metricGroupBy)
	case api.NODETELEMETRYTYPE_USED_CPU_CORES:
		metric = metrics2.GetUsedCpusMetric(clusterIdsFilter, nodePoolFilter, metricGroupBy)
	case api.NODETELEMETRYTYPE_ALLOCATED_CPU_CORES:
		metric = metrics2.GetAllocatedCpusMetric(clusterIdsFilter, nodePoolFilter, metricGroupBy)
	case api.NODETELEMETRYTYPE_TOTAL_GPU_MEMORY_BYTES:
		metric = metrics2.GetTotalGpuMemoryMetric(clusterIdsFilter, nodePoolFilter, metricGroupBy)
	case api.NODETELEMETRYTYPE_USED_GPU_MEMORY_BYTES:
		metric = metrics2.GetUsedGpuMemoryMetric(clusterIdsFilter, nodePoolFilter, metricGroupBy)
	case api.NODETELEMETRYTYPE_TOTAL_CPU_MEMORY_BYTES:
		metric = metrics2.GetTotalCpuMemoryMetric(clusterIdsFilter, nodePoolFilter, metricGroupBy)
	case api.NODETELEMETRYTYPE_USED_CPU_MEMORY_BYTES:
		metric = metrics2.GetUsedCpuMemoryMetric(clusterIdsFilter, nodePoolFilter, metricGroupBy)
	case api.NODETELEMETRYTYPE_ALLOCATED_CPU_MEMORY_BYTES:
		metric = metrics2.GetAllocatedCpuMemoryMetric(clusterIdsFilter, nodePoolFilter, metricGroupBy)
	case api.NODETELEMETRYTYPE_FULLY_FREE_GPU_NODES:
		metric = metrics2.GetFullyFreeGpuNodesMetric(clusterIdsFilter, nodePoolFilter, metricGroupBy)
	default:
		// shouldn't get here - validated before
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Unknown node telemetry metric type %v.", nodeTelemetryType)
		return nil, fmt.Errorf("unknown node telemetry metric type %s", nodeTelemetryType)
	}

	q, err := metricsClient.NewMetricQuery(metric)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to construct query for node telemetry metric type %s. err: %v", nodeTelemetryType, err)
		return nil, err
	}

	res, err := q.Query(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to query node telemetry metric type %s. err: %v", nodeTelemetryType, err)
		return nil, err
	}

	return res, nil
}

func useDepartmentsMetrics(ctx context.Context, orgUnitClient org_unit_service_client.IOrgUnitClient, clusterId string) bool {
	departments, err := orgUnitClient.CountClusterDepartments(ctx, clusterId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get departments, assumng no more than one. err: %v", err)
		return false
	}
	return departments > 1
}

// we will use the metric per the "lowest" scope -
// if requested groupBy node, we will use the per_node metric (and it doesn't matter which other groupBy we have).
// else - if requested groupBy nodepool, or nodepool filter, we will use the per_nodepool metric.
// else - if requested groupBy cluster, we will use the per_cluster metric.
// else - we will use the "no group by" metric.
func getMetricGroupBy(requestedGroupBy []string, filterNodepool bool) metrics2.MetricGroupBy {
	if len(requestedGroupBy) == 0 && !filterNodepool {
		return metrics2.GroupByTenant
	}

	clusterGroupByFound := false
	nodepoolGroupByFound := false
	nodeGroupByFound := false
	for _, group := range requestedGroupBy {
		switch group {
		case metrics2.ClusterId:
			clusterGroupByFound = true
		case metrics2.Nodepool:
			nodepoolGroupByFound = true
		case metrics2.Node:
			nodeGroupByFound = true
		}
	}

	if nodeGroupByFound {
		return metrics2.GroupByNode
	}
	if nodepoolGroupByFound || filterNodepool {
		return metrics2.GroupByNodePool
	}
	if clusterGroupByFound {
		return metrics2.GroupByCluster
	}
	return metrics2.GroupByTenant
}
