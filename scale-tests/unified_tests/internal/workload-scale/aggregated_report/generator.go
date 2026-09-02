/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package aggregated_report

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/scale"
	workload_scale "github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/workload-scale"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/workload-scale/aggregated_report/aggregate"

	"github.com/dsx-ai-factory/run-ai/scale-tests/sdk/reports"
)

type WorkloadAnalysisAggregator interface {
	Aggregate(analyzedWorkloads []workload_scale.WorkloadFinalRawDataResponse) (map[string]reports.PerformanceMetric, error)
}

type WorkloadReportGenerator struct {
	Aggregator WorkloadAnalysisAggregator
}

func NewWorkloadReportGenerator() *WorkloadReportGenerator {
	return &WorkloadReportGenerator{
		Aggregator: aggregate.NewWorkloadAnalisysAggregator(),
	}
}

func (g *WorkloadReportGenerator) parseResultsToType(testResults *scale.SamplesReport) ([]workload_scale.WorkloadFinalRawDataResponse, error) {
	jsonData, _ := json.Marshal(testResults.Results)
	var converted []workload_scale.WorkloadFinalRawDataResponse
	err := json.Unmarshal(jsonData, &converted)
	if err != nil {
		return nil, fmt.Errorf("failed to convert results to type: %v", err)
	}
	return converted, nil
}

func (g *WorkloadReportGenerator) GenerateReport(testResults *scale.SamplesReport) (*reports.SuiteReport, error) {
	// Analyze the workloads
	converted, err := g.parseResultsToType(testResults)
	if err != nil {
		return nil, fmt.Errorf("failed to convert raw test results to expected type: %v", err)
	}

	// Aggregate the results
	performanceMetrics, err := g.Aggregator.Aggregate(converted)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate the results: %v", err)
	}
	for metricType, performanceMetric := range testResults.Metrics {
		performanceMetrics[string(metricType)] = performanceMetric
	}

	// Parse infrastructure details from metadata
	infraDetails := parseInfrastructureDetails(testResults.MetaData)

	suiteReport := &reports.SuiteReport{
		ID:                 testResults.RunIdentifier,
		CreatedAt:          time.Now(),
		PerformanceMetrics: performanceMetrics,
		MetaData:           testResults.MetaData,
		Environment: reports.Environment{
			OperatorVersion:       testResults.OperatorVersion,
			ControlPlaneVersion:   testResults.CpVersion,
			InfrastructureDetails: infraDetails,
		},
	}
	return suiteReport, nil
}

func parseInfrastructureDetails(metadata map[string]string) *reports.InfrastructureDetails {
	if metadata == nil {
		return nil
	}

	infraDetails := &reports.InfrastructureDetails{
		KubernetesVersion:      metadata["kubernetes_version"],
		KubernetesDistribution: metadata["kubernetes_distribution"],
		ResourceTier:           metadata["resource_tier"],
	}

	parseIntField(metadata, "total_nodes", &infraDetails.TotalNodes)
	parseIntField(metadata, "real_node_count", &infraDetails.RealNodeCount)
	parseIntField(metadata, "fake_node_count", &infraDetails.FakeNodeCount)
	parseIntField(metadata, "nodepool_count", &infraDetails.NodepoolCount)
	parseIntField(metadata, "nodes_per_pool", &infraDetails.NodesPerPool)
	parseBoolField(metadata, "kwok_enabled", &infraDetails.KwokEnabled)

	if gpuProducts := metadata["gpu_products"]; gpuProducts != "" {
		infraDetails.GpuProducts = strings.Split(gpuProducts, ",")
	}

	infraDetails.RunaiResources = parseRunaiResources(metadata)

	return infraDetails
}

func parseIntField(metadata map[string]string, key string, target *int) {
	if val, err := strconv.Atoi(metadata[key]); err == nil {
		*target = val
	}
}

func parseBoolField(metadata map[string]string, key string, target *bool) {
	if val, err := strconv.ParseBool(metadata[key]); err == nil {
		*target = val
	}
}

func parseRunaiResources(metadata map[string]string) *reports.RunaiResources {
	runaiResources := &reports.RunaiResources{}
	hasResources := false

	if schedulingSvc := parseComponentResources(metadata, "scheduling_services"); schedulingSvc != nil {
		runaiResources.SchedulingServices = schedulingSvc
		hasResources = true
	}

	if syncSvc := parseComponentResources(metadata, "sync_services"); syncSvc != nil {
		runaiResources.SyncServices = syncSvc
		hasResources = true
	}

	if workloadSvc := parseComponentResources(metadata, "workload_services"); workloadSvc != nil {
		runaiResources.WorkloadServices = workloadSvc
		hasResources = true
	}

	if exporter := parseComponentResources(metadata, "workload_exporter"); exporter != nil {
		runaiResources.WorkloadExporter = exporter
		hasResources = true
	}

	if !hasResources {
		return nil
	}
	return runaiResources
}

func parseComponentResources(metadata map[string]string, prefix string) *reports.ComponentResources {
	cpuReq := metadata[prefix+"_cpu_request"]
	cpuLimit := metadata[prefix+"_cpu_limit"]
	memReq := metadata[prefix+"_memory_request"]
	memLimit := metadata[prefix+"_memory_limit"]

	if cpuReq == "" && cpuLimit == "" && memReq == "" && memLimit == "" {
		return nil
	}

	return &reports.ComponentResources{
		CPURequest:    cpuReq,
		CPULimit:      cpuLimit,
		MemoryRequest: memReq,
		MemoryLimit:   memLimit,
	}
}
