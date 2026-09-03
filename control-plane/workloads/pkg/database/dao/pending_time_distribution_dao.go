/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
	telemetry_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/telemetry"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const (
	bucketGroupKey = "Bucket"
)

// DbGetPendingTimeDistribution returns distribution of pending workloads across 4 time buckets - calculated based on cluster-level P95 percentile pending time
func (w *WorkloadsDao) DbGetPendingTimeDistribution(
	ctx context.Context,
	tenantId uint,
	clusterId string,
	nodepoolName string,
	departmentId string,
	projectId string,
	permittedScopes *openapi.PermittedScopes,
	groupBy [][]string,
) (*[]telemetry_utils.TelemetryGroupValue, *api.Error) {
	dbErr := api.NewError(http.StatusInternalServerError, "Failed to get pending time distribution", nil)

	p95PendingTime, err := w.getClusterP95PendingTime(ctx, tenantId, clusterId, nodepoolName)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get P95 pending time, err: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	buckets := services_utils.CalculatePendingTimeBuckets(p95PendingTime)

	telemetryValues, err := w.getPendingWorkloadCountsByBucket(ctx, tenantId, clusterId, nodepoolName, departmentId, projectId, permittedScopes, groupBy, buckets)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to get all bucket counts, err: %v", err)
		return nil, dbErr.WithDetails(err)
	}

	return telemetryValues, nil
}

// getClusterP95PendingTime calculates the 95th percentile of pending time at cluster level.
func (w *WorkloadsDao) getClusterP95PendingTime(
	ctx context.Context,
	tenantId uint,
	clusterId string,
	nodepoolName string,
) (int64, error) {
	var p95PendingTime float64

	phases := []api.Phase{api.PHASE_PENDING}
	query := w.buildTelemetryQuery(ctx, tenantId, clusterId, nodepoolName, "", "", nil, [][]string{}, phases).
		Select("COALESCE(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY EXTRACT(EPOCH FROM (NOW() - phase_updated_at))::BIGINT), 0) as p95_pending")

	err := query.Row().Scan(&p95PendingTime)
	return int64(p95PendingTime), err
}

// getBucketColumnName returns the SQL column name for a bucket at the given index.
// Bucket indices are 0-based, but column names are 1-based (bucket_1, bucket_2, etc.)
func getBucketColumnName(bucketIndex int) string {
	return fmt.Sprintf("bucket_%d", bucketIndex+1)
}

func (w *WorkloadsDao) getPendingWorkloadCountsByBucket(
	ctx context.Context,
	tenantId uint,
	clusterId string,
	nodepoolName string,
	departmentId string,
	projectId string,
	permittedScopes *openapi.PermittedScopes,
	groupBy [][]string,
	buckets []services_utils.Bucket,
) (*[]telemetry_utils.TelemetryGroupValue, error) {
	// Build bucket CASE statements for SELECT
	selectFields := []string{}
	for i, bucket := range buckets {
		var caseStmt string
		if bucket.MaxSeconds < 0 {
			// Last bucket (open-ended)
			caseStmt = fmt.Sprintf(
				"SUM(CASE WHEN EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) >= %d THEN 1 ELSE 0 END) as %s",
				bucket.MinSeconds, getBucketColumnName(i))
		} else {
			caseStmt = fmt.Sprintf(
				"SUM(CASE WHEN EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) >= %d AND EXTRACT(EPOCH FROM (NOW() - phase_updated_at)) < %d THEN 1 ELSE 0 END) as %s",
				bucket.MinSeconds, bucket.MaxSeconds, getBucketColumnName(i))
		}
		selectFields = append(selectFields, caseStmt)
	}

	for _, groups := range groupBy {
		selectFields = append(selectFields, groups[0])
		if len(groups) > 1 {
			selectFields = append(selectFields, groups[1])
		}
	}

	phases := []api.Phase{api.PHASE_PENDING}
	query := w.buildTelemetryQuery(ctx, tenantId, clusterId, nodepoolName, departmentId, projectId, permittedScopes, groupBy, phases).
		Select(strings.Join(selectFields, ", "))

	var results []map[string]interface{}
	err := query.Find(&results).Error
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return createEmptyPendingTimeBuckets(buckets), nil
	}

	telemetryValues := transformPendingTimeResultsToTelemetry(results, buckets, groupBy)
	return &telemetryValues, nil
}

func createEmptyPendingTimeBuckets(buckets []services_utils.Bucket) *[]telemetry_utils.TelemetryGroupValue {
	telemetryValues := []telemetry_utils.TelemetryGroupValue{}
	for _, bucket := range buckets {
		telemetryValue := createPendingTimeBucketTelemetry("0", bucket)
		telemetryValues = append(telemetryValues, telemetryValue)
	}
	return &telemetryValues
}

// transformPendingTimeResultsToTelemetry transforms database query results into telemetry group values.
// results: Each map contains bucket counts (bucket_1, bucket_2, etc.) and optionally entity grouping fields.
//
//	For example: {"bucket_1": 5, "bucket_2": 3, "cluster_id": "uuid", "cluster_name": "cluster1"}
//
// buckets: The 4 bucket definitions used for grouping pending times
// groupBy: Optional entity grouping fields (max 1 item). Each item is [valueField] or [valueField, nameField]
func transformPendingTimeResultsToTelemetry(
	results []map[string]interface{},
	buckets []services_utils.Bucket,
	groupBy [][]string,
) []telemetry_utils.TelemetryGroupValue {
	telemetryValues := []telemetry_utils.TelemetryGroupValue{}

	for _, result := range results {
		for i, bucket := range buckets {
			bucketCol := getBucketColumnName(i)
			countValue := result[bucketCol]

			// Handle missing or nil bucket count - default to "0"
			count := "0"
			if countValue != nil {
				countStr := fmt.Sprintf("%v", countValue)
				if countStr != "" && countStr != "<nil>" {
					count = countStr
				}
			}

			telemetryValue := createPendingTimeBucketTelemetry(count, bucket)

			// Add entity grouping (cluster/department/project) if requested
			if len(groupBy) > 0 {
				addPendingTimeEntityGrouping(&telemetryValue, result, groupBy[0])
			}

			telemetryValues = append(telemetryValues, telemetryValue)
		}
	}

	return telemetryValues
}

func createPendingTimeBucketTelemetry(count string, bucket services_utils.Bucket) telemetry_utils.TelemetryGroupValue {
	return telemetry_utils.TelemetryGroupValue{
		Data:                 count,
		Group1MainName:       bucketGroupKey,
		Group1MainValue:      bucket.Name,
		Group1SecondaryName:  bucketGroupKey,
		Group1SecondaryValue: bucket.Name,
	}
}

// addPendingTimeEntityGrouping adds entity grouping fields (cluster/department/project) to Group2.
// groupFields: Either [valueField] or [valueField, nameField] format
// Note: Group2MainName and Group2SecondaryName are not used - the API key names come from groupBy []string in TelemetryGroupToApi
func addPendingTimeEntityGrouping(
	telemetryValue *telemetry_utils.TelemetryGroupValue,
	result map[string]interface{},
	groupFields []string,
) {
	if len(groupFields) >= 2 {
		// Two-field grouping: value and name (e.g., cluster_id, cluster_name)
		valueField := groupFields[0]
		nameField := groupFields[1]

		var valueStr, nameStr string
		if result[valueField] != nil {
			valueStr = fmt.Sprintf("%v", result[valueField])
		}
		if result[nameField] != nil {
			nameStr = fmt.Sprintf("%v", result[nameField])
		}

		telemetryValue.Group2MainValue = valueStr
		telemetryValue.Group2SecondaryValue = nameStr
	} else if len(groupFields) == 1 {
		// Single-field grouping: value only (e.g., cluster_id)
		valueField := groupFields[0]
		fieldValue := result[valueField]

		var valueStr string
		if fieldValue != nil {
			valueStr = fmt.Sprintf("%v", fieldValue)
		}

		telemetryValue.Group2MainValue = valueStr
		telemetryValue.Group2SecondaryValue = valueStr
	}
}
