/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reports

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"

	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	auth_swagger "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	metrics_swagger "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	org_unit_config "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaihttp"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/s3"
	"github.com/minio/minio-go/v7"
	"golang.org/x/exp/slices"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
)

const (
	groupByKey  = "groupBy"
	filterByKey = "filterBy"
)

type ReportsApiService struct {
	MetricsClient metrics_client.MetricsClientContract
	S3Client      *minio.Client
}

func NewReportsApiService() api.ReportsAPIServicer {
	var s3ClientPtr *minio.Client
	s3Client, err := s3.Client()
	if err != nil {
		logging.Warnf("Failed to fetch s3 client. error: %v", err)
	} else {
		s3ClientPtr = &s3Client
	}

	return &ReportsApiService{
		MetricsClient: metrics_client.GetClientInstance(),
		S3Client:      s3ClientPtr,
	}
}

func (r ReportsApiService) AreReportsAvailable(_ context.Context) (api.ImplResponse, error) {
	return api.ImplResponse{Code: http.StatusOK, Body: api.AreReportsAvailable200Response{IsAvailable: r.S3Client != nil}}, nil
}

func (r ReportsApiService) DeleteReportById(ctx context.Context, id string) (api.ImplResponse, error) {
	_, errResponse := r.getReportAndValidatePermissions(ctx, id, string(auth_swagger.ACTION_DELETE))
	if errResponse != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error response for report %s. error: %+v", id, *errResponse)
		return *errResponse, nil
	}

	statusCode, err := r.MetricsClient.DeleteReport(ctx, id)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to delete report with id %s. error: %+v", id, err)
		return services.HandleError(err, statusCode, fmt.Sprintf("Failed to delete report with id %s", id)), nil
	}

	return api.ImplResponse{Code: http.StatusNoContent, Body: nil}, nil
}

func (r ReportsApiService) GetReportById(ctx context.Context, id string) (api.ImplResponse, error) {
	report, errResponse := r.getReportAndValidatePermissions(ctx, id, string(auth_swagger.ACTION_READ))
	if errResponse != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error response for report %s. error: %+v", id, *errResponse)
		return *errResponse, nil
	}

	return api.ImplResponse{
		Code: http.StatusOK,
		Body: metricsReportToApi(report),
	}, nil
}

func (r ReportsApiService) ListReports(ctx context.Context, filterBy []string, sortBy api.ReportFilterAndSortFields, sortOrder string, offset int32, limit int32, search string) (api.ImplResponse, error) {
	filterBy, errResponse := setFilterByBasedOnPermission(ctx, filterBy)
	if errResponse != nil {
		return *errResponse, nil
	}

	reports, code, err := r.MetricsClient.ListReports(ctx, filterBy, metrics_swagger.ReportFilterAndSortFields(sortBy), sortOrder, offset, limit, search)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to list reports, got an error from metrics service. Code: %d Err: %v", code, err)
		if code >= http.StatusBadRequest && code < http.StatusInternalServerError {
			return api.ImplResponse{Code: code, Body: err.Error()}, nil
		}

		return services.HandleError(err, http.StatusInternalServerError, "Failed to list reports"), nil
	}

	apiReports := []api.Report{}
	for _, report := range reports {
		apiReports = append(apiReports, metricsReportToApi(&report))
	}

	return api.ImplResponse{
		Code: http.StatusOK,
		Body: api.ListReports200Response{Reports: apiReports},
	}, nil
}

func (r ReportsApiService) CountReports(ctx context.Context, filterBy []string, search string) (api.ImplResponse, error) {
	filterBy, errResponse := setFilterByBasedOnPermission(ctx, filterBy)
	if errResponse != nil {
		return *errResponse, nil
	}

	count, code, err := r.MetricsClient.CountReports(ctx, filterBy, search)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to count reports, got an error from metrics service. Code: %d Err: %v", code, err)
		if code >= http.StatusBadRequest && code < http.StatusInternalServerError {
			return api.ImplResponse{Code: code, Body: err.Error()}, nil
		}

		return services.HandleError(err, http.StatusInternalServerError, "Failed to count reports"), nil
	}

	return api.ImplResponse{
		Code: http.StatusOK,
		Body: api.CountResponse{Count: count},
	}, nil
}

func setFilterByBasedOnPermission(ctx context.Context, filterBy []string) ([]string, *api.ImplResponse) {
	tenantId, subjectId, _, err := runaictx.ParseAuthData(ctx)
	if err != nil {
		return nil, ptr.To(services.HandleError(err, http.StatusBadRequest, err.Error()))
	}

	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	if permittedScopes == nil || (!permittedScopes.GetSystem() && !slices.Contains(permittedScopes.Tenants, tenantId)) {
		filterBy = append(filterBy, fmt.Sprintf("%s==%s", metrics_swagger.REPORTFILTERANDSORTFIELDS_NAME, subjectId))
	}
	return filterBy, nil
}

func (r ReportsApiService) DownloadReportById(ctx context.Context, id string) (api.ImplResponse, error) {
	if r.S3Client == nil {
		return api.ImplResponse{Code: http.StatusBadRequest, Body: "Reports storage is not configured"}, nil
	}

	report, errResponse := r.getReportAndValidatePermissions(ctx, id, string(auth_swagger.ACTION_READ))
	if errResponse != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got an error response for report %s. error: %+v", id, *errResponse)
		return *errResponse, nil
	}

	if report.Status != metrics_swagger.REPORTSTATUS_READY {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("report %s is not ready yet, current status: %s", id, report.Status)
		return api.ImplResponse{Code: http.StatusBadRequest, Body: fmt.Sprintf("report is not ready yet, current status: %s", report.Status)}, nil
	}

	bucketName := config.Get[*org_unit_config.Reports]().BucketName
	obj, err := r.S3Client.GetObject(ctx, bucketName, id, minio.GetObjectOptions{})
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("download report %s from bucket %s failed. error: %v", id, bucketName, err)
		return services.HandleError(err, http.StatusInternalServerError, fmt.Sprintf("download report %s failed", id)), nil
	}
	defer obj.Close()

	csvReader := csv.NewReader(obj)
	records, err := csvReader.ReadAll()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("parsing CSV report %s failed. error: %v", id, err)
		return services.HandleError(err, http.StatusInternalServerError, fmt.Sprintf("failed to parse CSV file or report %s", id)), nil
	}

	var csvBuffer bytes.Buffer
	csvWriter := csv.NewWriter(&csvBuffer)
	err = csvWriter.WriteAll(records)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("writing CSV report %s failed. error: %v", id, err)
		return services.HandleError(err, http.StatusInternalServerError, fmt.Sprintf("failed to write CSV file or report %s", id)), nil
	}
	csvWriter.Flush()

	return api.ImplResponse{Code: http.StatusOK, Body: runaihttp.CsvFile{Data: records, FileName: fmt.Sprintf("%s.csv", strings.ReplaceAll(strings.TrimSpace(report.Name), " ", "_"))}}, nil
}

func (r ReportsApiService) CreateReport(ctx context.Context, fields api.ReportRequestFields) (api.ImplResponse, error) {
	if r.S3Client == nil {
		return api.ImplResponse{Code: http.StatusBadRequest, Body: "Reports storage is not configured"}, nil
	}

	tenantId, tenantErr := services.ParseTenantId(ctx)
	if tenantErr != nil {
		return services.HandleError(tenantErr, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	filtersMap, err := reportValidations(fields)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("report %s request failed validations. error: %+v", fields.Name, err)
		return api.ImplResponse{Code: http.StatusBadRequest, Body: err.Error()}, nil
	}

	err = SetPermissionFilters(ctx, *tenantId, &filtersMap)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to set permission filter on report %s. error: %+v", fields.Name, err)
		return api.ImplResponse{Code: http.StatusInternalServerError, Body: err.Error()}, nil
	}

	groupBy := api.REPORTGROUPBYOPTIONS_NODEPOOL
	if fields.GroupBy != nil {
		groupBy = *fields.GroupBy
	}

	reportCreator := metrics_client.NewReportCreator(fields.Name, &fields.Start, &fields.End, fields.Description)
	metrics := getReportMetrics(ctx, groupBy, filtersMap)

	reportCreator.
		SetMetrics(metrics).
		SetLabels(getReportLabels(groupBy)).
		AddAdditionalField(groupByKey, fields.GroupBy).
		AddAdditionalField(filterByKey, fields.FilterBy)

	report, code, err := reportCreator.Create(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to create report %s, got an error from metrics service. Code: %d Err: %v", fields.Name, code, err)
		if code >= http.StatusBadRequest && code < http.StatusInternalServerError {
			return api.ImplResponse{Code: code, Body: err.Error()}, nil
		}

		return api.ImplResponse{Code: http.StatusInternalServerError, Body: "Report creation request failed unexpectedly. Please try again later"}, nil
	}

	responseFilterBy, responseGroupBy := parseAdditionalFieldsToReport(report.AdditionalFields)

	return api.ImplResponse{Code: http.StatusAccepted, Body: api.Report{
		Name:            report.Name,
		Description:     convutils.Deref(report.Description, ""),
		Start:           report.Range.Get().Start,
		End:             report.Range.Get().End,
		GroupBy:         responseGroupBy,
		FilterBy:        responseFilterBy,
		Id:              report.Id,
		CreatedAt:       report.CreatedAt,
		CreatedBy:       report.CreatedBy,
		TenantId:        report.TenantId,
		Status:          api.ReportStatus(report.Status),
		StatusUpdatedAt: report.StatusUpdatedAt,
		StatusMessage:   convutils.Deref(report.StatusMessage, ""),
	}}, nil
}

func (r ReportsApiService) getReportAndValidatePermissions(ctx context.Context, reportId string, action string) (*metrics_swagger.Report, *api.ImplResponse) {
	tenantId, err := services.ParseTenantId(ctx)
	if err != nil {
		return nil, ptr.To(services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"))
	}

	subjectId := ""
	subjectIdPtr, err := services.ParseSubjectId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to parse subject id: %v", err)
	} else {
		subjectId = *subjectIdPtr
	}

	report, statusCode, err := r.MetricsClient.GetReport(ctx, reportId)
	if err != nil {
		return nil, ptr.To(services.HandleError(err, statusCode, fmt.Sprintf("Failed to get report with id %s", reportId)))
	}

	if report.CreatedBy != subjectId {
		permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
		if err != nil {
			permittedScopes = nil
		}

		if permittedScopes == nil || (!permittedScopes.GetSystem() && !slices.Contains(permittedScopes.Tenants, *tenantId)) {
			return nil, ptr.To(services.HandleError(err, http.StatusForbidden, fmt.Sprintf("User %s is not the report creator nor has tenant level %s project permissions", subjectId, action)))
		}
	}

	return report, nil
}

func parseAdditionalFieldsToReport(additionalFields map[string]interface{}) (filterBy []string, groupBy *api.ReportGroupByOptions) {
	for key, value := range additionalFields {
		switch key {
		case filterByKey:
			if v, ok := value.([]string); ok {
				filterBy = v
			} else if v, ok := value.([]interface{}); ok {
				for _, item := range v {
					if str, ok := item.(string); ok {
						filterBy = append(filterBy, str)
					} else {
						logging.Warnf("Non-string value in %s array: %v", filterByKey, item)
					}
				}
			} else {
				logging.Warnf("Unexpected type for %s: %T", filterByKey, value)
			}
		case groupByKey:
			if v, ok := value.(string); ok {
				groupByOpt := api.ReportGroupByOptions(v)
				groupBy = &groupByOpt
			} else {
				logging.Warnf("Unexpected type for %s: %T", groupByKey, value)
			}
		default:
		}
	}

	return
}

func (r ReportsApiService) BatchDeleteReports(_ context.Context, _ api.BatchDeleteReportsRequest) (api.ImplResponse, error) {
	return api.Response(http.StatusNotImplemented, nil), nil
}
