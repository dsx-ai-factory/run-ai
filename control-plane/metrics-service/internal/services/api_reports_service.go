/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
)

type ReportsApiService struct {
	metricQueryBuilder MetricQueryBuilder
	reportsDao         ReportsDao
}

func NewReportsApiService(metricQueryBuilder MetricQueryBuilder, reportsDao ReportsDao) api.ReportsAPIServicer {
	return &ReportsApiService{
		metricQueryBuilder: metricQueryBuilder,
		reportsDao:         reportsDao,
	}
}

func (s *ReportsApiService) ListReports(ctx context.Context, filterBy []string, sortBy api.ReportFilterAndSortFields, sortOrder string, offset int32, limit int32, search string) (api.ImplResponse, error) {
	tenantId, err := GetTenantIdFromCtx(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get tenant id: %v", err)
		return api.Response(http.StatusInternalServerError, nil), fmt.Errorf("failed to get tenant id: %v", err)
	}
	parser := queryapi.NewParser((&models.Report{}).SortAndFilterFieldsToColumn())
	filters, filterErr := parser.Filters(filterBy)
	if filterErr != nil {
		return api.Response(http.StatusBadRequest, fmt.Sprintf("filters %s invalid", filterBy)), filterErr
	}
	sortColumn, sortError := parser.GetMapping(string(sortBy))
	if sortError != nil {
		return api.Response(http.StatusBadRequest, fmt.Sprintf("sort %s invalid", sortBy)), sortError
	}

	modelReports, dbErr := s.reportsDao.GetReports(ctx, tenantId, filters, sortColumn, sortOrder, offset, limit, search)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get reports record: %v", dbErr)
		return NewErrorResponse(dbErr), fmt.Errorf("failed to get reports record: %v", dbErr)
	}

	response := api.ListReports200Response{Reports: []api.Report{}}
	for _, report := range modelReports {
		response.Reports = append(response.Reports, modelReportToApi(report))
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *ReportsApiService) CountReports(ctx context.Context, filterBy []string, search string) (api.ImplResponse, error) {
	tenantId, err := GetTenantIdFromCtx(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get tenant id: %v", err)
		return api.Response(http.StatusInternalServerError, nil), fmt.Errorf("failed to get tenant id: %v", err)
	}
	parser := queryapi.NewParser((&models.Report{}).SortAndFilterFieldsToColumn())
	filters, filterErr := parser.Filters(filterBy)
	if filterErr != nil {
		return api.Response(http.StatusBadRequest, fmt.Sprintf("filters %s invalid", filterBy)), filterErr
	}

	count, dbErr := s.reportsDao.CountReports(ctx, tenantId, filters, search)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to count reports record: %v", dbErr)
		return NewErrorResponse(dbErr), fmt.Errorf("failed to count reports record: %v", dbErr)
	}

	return api.Response(http.StatusOK, api.CountReports200Response{Count: count}), nil
}

// CreateReport - Create a new report.
func (s *ReportsApiService) CreateReport(ctx context.Context, reportRequestFields api.ReportRequestFields) (api.ImplResponse, error) {
	tenantId, err := GetTenantIdFromCtx(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get tenant id: %v", err)
		return api.Response(http.StatusInternalServerError, nil), fmt.Errorf("failed to get tenant id: %v", err)
	}
	metricInformations := buildMetricInformationsFromReportRequestFields(reportRequestFields)
	metricQueries, err := s.buildMetricQueries(metricInformations)
	if err != nil {
		logging.Errorf("Failed to build queries for report, tenant_id [%d], report_name: [%s]. err: %v", tenantId, reportRequestFields.Name, err)
		return api.Response(http.StatusInternalServerError, nil), fmt.Errorf("failed to build queries for report, tenant_id [%d], report_name: [%s]. err: %v", tenantId, reportRequestFields.Name, err)
	}
	report := database.NewReportFromApi(reportRequestFields, metricQueries)
	modelReport, dbErr := s.reportsDao.CreateReport(ctx, tenantId, report)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to create report record %v: %v", reportRequestFields, dbErr)
		return NewErrorResponse(dbErr), fmt.Errorf("failed to create report record %v: %v", reportRequestFields, dbErr)
	}

	return api.Response(http.StatusAccepted, modelReportToApi(modelReport)), nil
}

func buildMetricInformationsFromReportRequestFields(reportRequestFields api.ReportRequestFields) map[string]*MetricInformation {
	res := map[string]*MetricInformation{}
	for _, metric := range reportRequestFields.Metrics {
		res[metric.DisplayName] = &MetricInformation{
			Name:           metric.Name,
			Filters:        metric.Filters,
			ResponseLabels: []string{},
			Range: &MetricInformationRange{
				Start:       reportRequestFields.Range.Start,
				End:         reportRequestFields.Range.End,
				Aggregation: true,
			},
			Quantile: metric.Quantile,
		}
	}

	return res
}

// DeleteReportById - Delete report
func (s *ReportsApiService) DeleteReportById(ctx context.Context, reportId string) (api.ImplResponse, error) {
	tenantId, err := GetTenantIdFromCtx(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get tenant id: %v", err)
		return api.Response(http.StatusInternalServerError, nil), fmt.Errorf("failed to get tenant id: %v", err)
	}

	dbErr := s.reportsDao.DeleteReport(ctx, tenantId, reportId)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to delete report %s record: %v", reportId, dbErr)
		return NewErrorResponse(dbErr), fmt.Errorf("failed to delete report %s record: %v", reportId, dbErr)
	}

	return api.Response(http.StatusNoContent, nil), nil
}

func (s *ReportsApiService) BatchDeleteReports(ctx context.Context, request api.ReportsIdsBatchRequest) (api.ImplResponse, error) {
	tenantId, err := GetTenantIdFromCtx(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get tenant id: %v", err)
		return api.Response(http.StatusInternalServerError, nil), fmt.Errorf("failed to get tenant id: %v", err)
	}

	accepted, dbErr := s.reportsDao.UpdatePendingDelete(
		ctx,
		tenantId,
		request.Ids,
	)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to mark reports for deletion: %v", dbErr)
		return NewErrorResponse(dbErr), fmt.Errorf("failed to mark reports for deletion: %v", dbErr)
	}

	acceptedMap := toMap(accepted)
	response := api.BatchResponseAsync{
		Id:       uuid.NewString(),
		Accepted: accepted,
		Failed:   make([]api.BatchResponseFailureItem, 0, len(request.Ids)-len(acceptedMap)),
	}

	inProgress, dbErr := s.reportsDao.GetReportsByIds(ctx, tenantId, request.Ids, queryapi.Filter{
		Field:    "status",
		Operator: "=",
		Values:   []string{string(api.REPORTSTATUS_PROCESSING), string(api.REPORTSTATUS_DELETING)},
	})
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get reports by ids: %v", dbErr)
		return NewErrorResponse(dbErr), fmt.Errorf("failed to get reports by ids: %v", dbErr)
	}

	inProgressMap := mapByReportId(inProgress)
	for _, reportId := range request.Ids {
		if _, ok := acceptedMap[reportId]; ok {
			continue
		}

		if _, ok := inProgressMap[reportId]; ok {
			if inProgressMap[reportId].Status == string(api.REPORTSTATUS_DELETING) {
				response.Accepted = append(response.Accepted, reportId)
				continue
			}

			response.Failed = append(response.Failed, api.BatchResponseFailureItem{
				Id:      reportId,
				Code:    http.StatusProcessing,
				Message: fmt.Sprintf("report with id '%s' is processing", reportId),
			})
			continue
		}

		response.Failed = append(response.Failed, api.BatchResponseFailureItem{
			Id:      reportId,
			Code:    http.StatusNotFound,
			Message: fmt.Sprintf("report with id '%s' not found", reportId),
		})
	}

	return api.Response(http.StatusAccepted, response), nil
}

// GetReportById - Get report
func (s *ReportsApiService) GetReportById(ctx context.Context, reportId string) (api.ImplResponse, error) {
	tenantId, err := GetTenantIdFromCtx(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get tenant id: %v", err)
		return api.Response(http.StatusInternalServerError, nil), fmt.Errorf("failed to get tenant id: %v", err)
	}

	modelReport, dbErr := s.reportsDao.GetReport(ctx, tenantId, reportId)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get report %s record: %v", reportId, dbErr)
		return NewErrorResponse(dbErr), fmt.Errorf("failed to get report %s record: %v", reportId, dbErr)
	}

	return api.Response(http.StatusOK, modelReportToApi(modelReport)), nil
}

func (s *ReportsApiService) BatchGetReportsByIds(ctx context.Context, request api.ReportsIdsBatchRequest) (api.ImplResponse, error) {
	if len(request.Ids) == 0 {
		return api.Response(http.StatusOK, newBatchGetReportsByIdsResponse()), nil
	}

	tenantId, err := GetTenantIdFromCtx(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get tenant id: %v", err)
		return api.Response(http.StatusInternalServerError, nil), fmt.Errorf("failed to get tenant id: %v", err)
	}

	response := newBatchGetReportsByIdsResponse()

	modelReports, dbErr := s.reportsDao.GetReportsByIds(ctx, tenantId, request.Ids)
	if dbErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to get batchreports by ids: %v", dbErr)
		return api.Response(http.StatusInternalServerError, nil), fmt.Errorf("failed to get batchreports by ids")
	}

	reportsById := mapByReportId(modelReports)
	for _, id := range request.Ids {
		if modelReport, ok := reportsById[id]; ok {
			response.Succeeded = append(response.Succeeded, modelReportToApi(modelReport))
		} else {
			response.Failed = append(response.Failed, reportFailure(id, http.StatusNotFound, "requested resource does not exist"))
		}
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *ReportsApiService) buildMetricQueries(mi map[string]*MetricInformation) (map[string]string, error) {
	res := map[string]string{}
	for metricDisplayName, metric := range mi {
		query, err := s.metricQueryBuilder.Build(context.Background(), metric)
		if err != nil {
			return nil, fmt.Errorf("failed to build metric %s query, err %v", metricDisplayName, err)
		}

		res[metricDisplayName] = query
	}

	return res, nil
}

func modelReportToApi(modelReport models.Report) api.Report {
	return api.Report{
		Name:        modelReport.Name,
		Description: modelReport.Description,
		Labels:      convutils.ConvertSliceTypes(modelReport.Labels, modelLabelToApi),
		Range: &api.Range{
			Start: modelReport.Start.Truncate(time.Millisecond),
			End:   modelReport.End.Truncate(time.Millisecond),
		},
		AdditionalFields: modelReport.AdditionalFields,
		Id:               modelReport.ID.String(),
		CreatedAt:        modelReport.CreatedAt.Truncate(time.Millisecond),
		CreatedBy:        modelReport.CreatedBy,
		TenantId:         modelReport.TenantId,
		Status:           api.ReportStatus(modelReport.Status),
		StatusUpdatedAt:  modelReport.StatusUpdatedAt,
		StatusMessage:    modelReport.StatusMessage,
	}
}

func modelLabelToApi(label models.Label) api.Label {
	return api.Label{
		Name:        label.Name,
		DisplayName: label.DisplayName,
	}
}

func NewErrorResponse(apiError *api.Error) api.ImplResponse {
	apiError.Details = ""
	return api.ImplResponse{
		Code: int(apiError.Code),
		Body: *apiError,
	}
}

func toMap(ids []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		m[id] = struct{}{}
	}
	return m
}

func newBatchGetReportsByIdsResponse() api.ReportsBatchGetResponse {
	return api.ReportsBatchGetResponse{
		Id:        uuid.New().String(),
		Succeeded: []api.Report{},
		Failed:    []api.BatchResponseFailureItem{},
	}
}

func mapByReportId(modelReports []models.Report) map[string]models.Report {
	reportsById := make(map[string]models.Report, len(modelReports))
	for _, modelReport := range modelReports {
		reportsById[modelReport.ID.String()] = modelReport
	}
	return reportsById
}

func reportFailure(id string, code int, message string) api.BatchResponseFailureItem {
	return api.BatchResponseFailureItem{
		Id:      id,
		Code:    int32(code),
		Message: message,
	}
}
