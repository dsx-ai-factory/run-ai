/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	workloads_api "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

type OrgUnitWithWorkloadCounts interface {
	GetId() string
	SetRunningWorkloadsCount(count *int32)
	SetPendingWorkloadsCount(count *int32)
}

type ProjectRankingWrapper struct {
	*api.ProjectGetRanking
}

func (p *ProjectRankingWrapper) GetId() string {
	return p.Id
}

func (p *ProjectRankingWrapper) SetRunningWorkloadsCount(count *int32) {
	p.RunningWorkloadsCount = count
}

func (p *ProjectRankingWrapper) SetPendingWorkloadsCount(count *int32) {
	p.PendingWorkloadsCount = count
}

type DepartmentRankingWrapper struct {
	*api.DepartmentGetRanking
}

func (d *DepartmentRankingWrapper) GetId() string {
	return d.Id
}

func (d *DepartmentRankingWrapper) SetRunningWorkloadsCount(count *int32) {
	d.RunningWorkloadsCount = count
}

func (d *DepartmentRankingWrapper) SetPendingWorkloadsCount(count *int32) {
	d.PendingWorkloadsCount = count
}

func EnrichOrgUnitsWithWorkloadCounts[T OrgUnitWithWorkloadCounts](
	ctx context.Context,
	orgUnits []T,
	orgUnitType string,
	clusterId string,
	nodePoolName string,
	workloadsClient interface {
		InternalEnrichOrgUnitRanking(ctx context.Context, request workloads_api.InternalEnrichOrgUnitRankingRequest) (*workloads_api.InternalEnrichOrgUnitRankingResponse, error)
	},
) ([]T, error) {
	if len(orgUnits) == 0 || workloadsClient == nil {
		return orgUnits, nil
	}

	// Extract IDs from org units
	ids := make([]string, len(orgUnits))
	for i, orgUnit := range orgUnits {
		ids[i] = orgUnit.GetId()
	}

	// Build request with global fields
	request := workloads_api.InternalEnrichOrgUnitRankingRequest{
		ClusterId:   clusterId,
		OrgUnitType: orgUnitType,
		Ids:         ids,
	}

	// Add optional node pool filter
	if nodePoolName != "" {
		request.NodePoolName = *workloads_api.NewNullableString(&nodePoolName)
	}

	response, err := workloadsClient.InternalEnrichOrgUnitRanking(ctx, request)
	if err != nil {
		return orgUnits, fmt.Errorf("workloads enrichment failed: %w", err)
	}

	// Build counts map from response
	countsMap := make(map[string]struct {
		running *int32
		pending *int32
	})
	for _, item := range response.Items {
		runningCount := item.RunningWorkloadsCount.Get()
		pendingCount := item.PendingWorkloadsCount.Get()
		countsMap[item.Id] = struct {
			running *int32
			pending *int32
		}{
			running: runningCount,
			pending: pendingCount,
		}
	}

	// Apply counts to org units
	for i := range orgUnits {
		if counts, exists := countsMap[orgUnits[i].GetId()]; exists {
			orgUnits[i].SetRunningWorkloadsCount(counts.running)
			orgUnits[i].SetPendingWorkloadsCount(counts.pending)
		}
	}

	return orgUnits, nil
}

func ValidateGetOrgUnitRankingParams(
	ctx context.Context,
	metricType api.RankingMetricType,
	orderType api.RankingOrderType,
	clusterId string,
	clusterClient cluster_service_client.IClusterExternalClient,
) error {
	err := validateGetOrgUnitRankingParamsInner(metricType, orderType, clusterId)
	if err != nil {
		return err
	}

	clusterUuid, err := uuid.Parse(clusterId)
	if err != nil {
		return fmt.Errorf("failed to parse cluster id %s: %w", clusterId, err)
	}

	err = ValidateMinClusterVersionForRankingMetric(ctx, clusterUuid, metricType, clusterClient)
	if err != nil {
		return err
	}

	return nil
}

func validateGetOrgUnitRankingParamsInner(metricType api.RankingMetricType, orderType api.RankingOrderType, clusterId string) error {
	if clusterId == "" {
		return fmt.Errorf("invalid clusterId")
	}

	validMetricTypes := map[api.RankingMetricType]bool{
		api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION24H:         true,
		api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION24H:        true,
		api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION24H: true,
		api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION7D:          true,
		api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION7D:         true,
		api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION7D:  true,
		api.RANKINGMETRICTYPE_AVG_GPU_ALLOCATION30D:         true,
		api.RANKINGMETRICTYPE_AVG_GPU_UTILIZATION30D:        true,
		api.RANKINGMETRICTYPE_AVG_GPU_MEMORY_UTILIZATION30D: true,
		api.RANKINGMETRICTYPE_GPU_ALLOCATION_RATIO:          true,
	}

	if !validMetricTypes[metricType] {
		return fmt.Errorf("invalid metric type %s", metricType)
	}

	if orderType != api.RANKINGORDERTYPE_TOP && orderType != api.RANKINGORDERTYPE_BOTTOM {
		return fmt.Errorf("invalid order type %s", orderType)
	}

	return nil
}
