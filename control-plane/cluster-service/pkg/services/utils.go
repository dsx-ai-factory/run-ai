/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/run-ai/runai/control-plane/cluster-service/pkg/services/metrics"

	ext_api "github.com/run-ai/runai/control-plane/cluster-service/pkg/api"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/utils"
	tmClient "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/cluster-service/pkg/database/models"
)

const (
	CsvAccept = "text/csv"
)

func NoContentResponse() api.ImplResponse {
	return api.ImplResponse{
		Code: http.StatusNoContent,
	}
}

func AcceptedResponse() api.ImplResponse {
	return api.ImplResponse{
		Code: http.StatusAccepted,
	}
}

func NewErrorResponse(apiError *api.Error) api.ImplResponse {
	apiError.Details = ""
	return api.ImplResponse{
		Code: int(apiError.Code),
		Body: *apiError,
	}
}

func extractTenantIdFromCtxAndClusterId(ctx context.Context, clusterId string) (tenantId int32, apiErr *api.Error) {
	tenantId, err := ParseTenantId(ctx)
	if !runaictx.IsManagementTenant(ctx) {
		return tenantId, err
	}
	cluster, err := dao.GetClusterByIdAndTenantId(ctx, clusterId, tenantId, ext_api.VerbosityFull)
	if err != nil {
		return 0, err
	}
	return cluster.TenantId, nil
}

func extractTenantIdFromCtxAndValidateClusterExistence(ctx context.Context, clusterId string) (tenantId int32, apiErr *api.Error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return 0, api.NewError(http.StatusBadRequest, "Failed to get parse tenant for cluster with id: "+clusterId, nil)
	}

	cluster, err := dao.GetClusterByIdAndTenantId(ctx, clusterId, tenantId, ext_api.VerbosityFull)
	if err != nil {
		return 0, api.NewError(http.StatusBadRequest, "Failed to get cluster with id: "+clusterId, nil)
	}
	if !runaictx.IsManagementTenant(ctx) {
		return tenantId, err
	}
	return cluster.TenantId, nil
}

func ParseTenantId(ctx context.Context) (int32, *api.Error) {
	tenantId, err := utils.ParseTenantId(ctx)
	if err != nil {
		return 0, api.NewError(http.StatusBadRequest, err.Error(), err)
	}
	return tenantId, nil
}

func ParseAuthData(ctx context.Context) (int32, string, *api.Error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return 0, "", err
	}

	var ok bool
	var subjectId string
	if subjectId, ok = ctx.Value(runaictx.CfSubjectID).(string); !ok {
		return 0, "", api.NewError(http.StatusBadRequest, "No subject id provided", nil)
	}

	return tenantId, subjectId, nil
}

func isValidTenantId(ctx context.Context, tmClient *tmClient.APIClient, tenantID int32) (bool, *api.Error) {
	tenantReq := tmClient.TenantsAPI.GetTenantByIdInternal(ctx, tenantID)
	_, resp, err := tmClient.TenantsAPI.GetTenantByIdInternalExecute(tenantReq)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return false, api.NewError(http.StatusBadRequest, "invalid tenant id", err)
	}
	if err != nil {
		return false, api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to get tenant %d: %s", tenantID, err.Error()), err)
	}
	return true, nil
}

func TelemetryGroupsToApi(telemetryType string, timestamp time.Time, groups []models.TelemetryGroupValue, groupBy []string) *api.TelemetryResponse {
	var telemetryGroups []api.TelemetryResponseValuesInner
	for _, group := range groups {
		telemetryGroupApi := TelemetryGroupToApi(group, groupBy)
		telemetryGroups = append(telemetryGroups, *telemetryGroupApi)
	}
	telemetryResponse := api.TelemetryResponse{
		Type:      telemetryType,
		Timestamp: timestamp,
		Values:    telemetryGroups,
	}
	return &telemetryResponse
}

func TelemetryGroupToApi(group models.TelemetryGroupValue, groupBy []string) *api.TelemetryResponseValuesInner {
	var apiGroups []api.TelemetryResponseValuesInnerGroupsInner
	for i, fieldName := range groupBy {
		groupValue := group.Group1Value
		if i == 1 {
			groupValue = group.Group2Value
		}
		apiGroups = append(apiGroups, api.TelemetryResponseValuesInnerGroupsInner{
			Key:   fieldName,
			Value: groupValue,
		})
	}
	telemetryGroup := api.TelemetryResponseValuesInner{
		Value:  group.Data,
		Groups: apiGroups,
	}
	return &telemetryGroup
}

func GetTelemetryCsv(telemetryGroupsApi *api.TelemetryResponse) [][]string {
	var csvData [][]string
	numberOfValues := len(telemetryGroupsApi.Values)
	if numberOfValues == 0 {
		return csvData
	}
	titles := []string{"value"}
	for _, group := range telemetryGroupsApi.Values[0].Groups {
		groupKey := group.Key
		titles = append(titles,
			fmt.Sprintf("%s.value", groupKey),
			fmt.Sprintf("%s.name", groupKey))
	}
	csvData = append(csvData, titles)
	for _, telemetryValue := range telemetryGroupsApi.Values {
		values := []string{telemetryValue.Value}
		for _, group := range telemetryValue.Groups {
			values = append(values, group.Value, group.Name)
		}
		csvData = append(csvData, values)
	}
	return csvData
}

func getTimeMeasurementsMap(response metrics.MetricsResponseInterface) map[time.Time]map[string]string {
	timestampMap := make(map[time.Time]map[string]string)

	for _, m := range response.GetMeasurements() {
		if m.GetValues() != nil {
			for _, v := range *m.GetValues() {
				if _, ok := timestampMap[*v.Timestamp]; !ok {
					timestampMap[*v.Timestamp] = make(map[string]string)
				}
				timestampMap[*v.Timestamp][m.GetType()] = v.Value
			}
		}
	}

	return timestampMap
}

func getOrderedTimestamps(timeToMeasurements map[time.Time]map[string]string) []time.Time {
	var orderedTimestamps []time.Time
	for k := range timeToMeasurements {
		orderedTimestamps = append(orderedTimestamps, k)
	}

	sort.Slice(orderedTimestamps, func(i, j int) bool {
		return orderedTimestamps[i].Before(orderedTimestamps[j])
	})

	return orderedTimestamps
}

type StringLike interface {
	~string
}

// GetMetricsCsvFromInterface is a generic function that works with any MetricsResponseInterface
func GetMetricsCsvFromInterface[T StringLike](response metrics.MetricsResponseInterface, metricTypes []T) [][]string {
	var csvData [][]string

	timeToMeasurements := getTimeMeasurementsMap(response)
	orderedTimestamps := getOrderedTimestamps(timeToMeasurements)

	titles := []string{"Timestamp"}
	for _, metricType := range metricTypes {
		titles = append(titles, string(metricType))
	}

	csvData = append(csvData, titles)

	for _, timestamp := range orderedTimestamps {
		row := []string{}
		row = append(row, timestamp.String())
		timestampVal := timeToMeasurements[timestamp]
		for _, metricType := range metricTypes {
			metricValue := timestampVal[string(metricType)]
			row = append(row, metricValue)
		}
		csvData = append(csvData, row)
	}

	return csvData
}

func GetMetricsCsv[T StringLike](response *api.GroupedMetricsResponse, metricTypes []T) [][]string {
	wrapper := &metrics.GroupedMetricsResponseWrapper{GroupedMetricsResponse: response}
	return GetMetricsCsvFromInterface(wrapper, metricTypes)
}

// GetMetricsCsvFromMetricsResponse works with api.MetricsResponse
func GetMetricsCsvFromMetricsResponse[T StringLike](response *api.MetricsResponse, metricTypes []T) [][]string {
	wrapper := &metrics.MetricsResponseWrapper{MetricsResponse: response}
	return GetMetricsCsvFromInterface(wrapper, metricTypes)
}

func TranslateDaoError(err error, defaultMsg string) *api.Error {
	if err == nil {
		return nil
	}
	apiErr := &api.Error{
		Code:    http.StatusInternalServerError,
		Message: defaultMsg,
	}
	return apiErr.WithDetails(err)
}
