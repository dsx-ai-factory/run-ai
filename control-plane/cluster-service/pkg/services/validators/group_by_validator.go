/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validators

import (
	"context"
	"errors"
	"fmt"
	"slices"

	clients_local "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/clients"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
)

var groupByByMetrics = map[string][]api.MetricsType{
	string(client.GroupByCluster):    {api.METRICSTYPE_WORKLOADS_COUNT},
	string(client.GroupByNodePool):   {api.METRICSTYPE_WORKLOADS_COUNT},
	string(client.GroupByProject):    {api.METRICSTYPE_WORKLOADS_COUNT},
	string(client.GroupByDepartment): {api.METRICSTYPE_WORKLOADS_COUNT},
	string(client.GroupByCategory):   {api.METRICSTYPE_WORKLOADS_COUNT},
}

func validateGroupByForMetric(metricsType api.MetricsType, groupBy string) error {
	groupByMetrics, ok := groupByByMetrics[groupBy]

	if !ok || groupByMetrics == nil || !slices.Contains(groupByMetrics, metricsType) {
		return fmt.Errorf("invalid groupBy %s for metrics type: %s", groupBy, metricsType)
	}
	return nil
}

func validateClusterSupportForGroupBy(ctx context.Context, capChecker clients_local.ClusterCapabilityChecker, groupBy string) error {
	if groupBy != string(client.GroupByCategory) {
		return nil
	}

	supported, err := capChecker.ValidateClusterCapability(ctx, clients_local.LocalMinVersionOfGroupByWorkloadCategory)
	if err != nil {
		return err
	}

	if !supported {
		return fmt.Errorf("group by <%s> not supported for cluster", groupBy)
	}
	return nil
}

func ValidateGroupByForMetrics(ctx context.Context, capChecker clients_local.ClusterCapabilityChecker, metricsTypes []api.MetricsType, groupBy string) error {
	if groupBy == "" {
		return nil
	}
	var errs error
	err := validateClusterSupportForGroupBy(ctx, capChecker, groupBy)
	if err != nil {
		errs = errors.Join(errs, err)
	}
	for _, metricType := range metricsTypes {
		err = validateGroupByForMetric(metricType, groupBy)
		if err != nil {
			errs = errors.Join(errs, err)
		}
	}
	return errs
}
