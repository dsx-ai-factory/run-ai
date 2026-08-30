package services_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/services"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/database"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/database/dao/mocks"
	"github.com/run-ai/runai/control-plane/metrics-service/internal/database/models"
	services_mocks "github.com/run-ai/runai/control-plane/metrics-service/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-application-packages/queryapi"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

var ctx = context.WithValue(context.WithValue(context.Background(), runaictx.CfTenantId, "1"), runaictx.CfSubjectID, "user")

var processingStatusFilter = queryapi.Filter{
	Field:    "status",
	Operator: "=",
	Values:   []string{string(api.REPORTSTATUS_PROCESSING), string(api.REPORTSTATUS_DELETING)},
}
var request = api.ReportRequestFields{
	Name: "testing",
	Range: &api.Range{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now(),
	},
	Metrics: []api.MetricInformationForReport{
		{
			Name:        "metric_a",
			DisplayName: "Metric A",
			Filters: map[string][]string{
				"project": {"a", "b"},
			},
		},
	},
	Labels: []api.Label{
		{
			Name:        "labelA",
			DisplayName: "Label A",
		},
	},
	AdditionalFields: nil,
}

var queries = map[string]string{
	"Metric A": "metric_a",
}

var (
	report   models.Report
	reportId = uuid.New()
)

func TestMain(m *testing.M) {
	r := database.NewReportFromApi(request, queries)
	r.TenantId = 1
	r.ID = reportId
	report = r
	os.Exit(m.Run())
}

func initTestSubject(t *testing.T) (metricQueryBuilder *services_mocks.MetricQueryBuilder, dao *mocks.ReportsDao, reportService api.ReportsAPIServicer) {
	metricQueryBuilder = services_mocks.NewMetricQueryBuilder(t)
	dao = mocks.NewReportsDao(t)
	reportService = services.NewReportsApiService(metricQueryBuilder, dao)
	return
}

func TestCreateReport_Ok(t *testing.T) {
	metricQueryBuilder, dao, reportService := initTestSubject(t)
	metricQueryBuilder.On("Build", mock.Anything, mock.Anything).Return("metric_query", nil)
	dao.On("CreateReport", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(report, nil)

	response, err := reportService.CreateReport(ctx, request)
	assert.Nil(t, err)
	assert.Equal(t, response.Code, http.StatusAccepted)

	report, ok := response.Body.(api.Report)
	assert.True(t, ok)
	assert.Equal(t, request.Name, report.Name)
	assert.Equal(t, reportId.String(), report.Id)
	assert.Equal(t, report.TenantId, int32(1))
}

func TestCreateReport_FailedToBuildQuery(t *testing.T) {
	metricQueryBuilder, _, reportService := initTestSubject(t)
	metricQueryBuilder.On("Build", mock.Anything, mock.Anything).Return("", fmt.Errorf("failed to build query"), nil)

	response, err := reportService.CreateReport(ctx, request)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateReport_FailedConflict(t *testing.T) {
	metricQueryBuilder, dao, reportService := initTestSubject(t)
	metricQueryBuilder.On("Build", mock.Anything, mock.Anything).Return("metric_query", nil)
	dao.On("CreateReport", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(report, &api.Error{
		Code: http.StatusConflict,
	})

	response, err := reportService.CreateReport(ctx, request)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusConflict, response.Code)
}

func TestGetReport_Ok(t *testing.T) {
	_, dao, reportService := initTestSubject(t)
	dao.On("GetReport", mock.Anything, mock.Anything, report.ID.String()).Return(report, nil)

	response, err := reportService.GetReportById(ctx, report.ID.String())
	assert.Nil(t, err)
	assert.Equal(t, response.Code, http.StatusOK)

	report, ok := response.Body.(api.Report)
	assert.True(t, ok)
	assert.Equal(t, request.Name, report.Name)
	assert.Equal(t, reportId.String(), report.Id)
	assert.Equal(t, report.TenantId, int32(1))
}

func TestGetReport_Failed(t *testing.T) {
	_, dao, reportService := initTestSubject(t)
	dao.On("GetReport", mock.Anything, mock.Anything, report.ID.String()).Return(report, &api.Error{Code: http.StatusBadRequest, Message: "Failed to get report"})

	response, err := reportService.GetReportById(ctx, report.ID.String())
	assert.NotNil(t, err)
	assert.Equal(t, response.Code, http.StatusBadRequest)
}

func TestListReports_Ok(t *testing.T) {
	_, dao, reportService := initTestSubject(t)
	dao.On("GetReports", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]models.Report{report}, nil)

	response, err := reportService.ListReports(ctx, []string{}, "", "", 0, 0, "")
	assert.Nil(t, err)
	assert.Equal(t, response.Code, http.StatusOK)

	responseObj, ok := response.Body.(api.ListReports200Response)
	assert.True(t, ok)
	assert.NotEmpty(t, responseObj.Reports)
	assert.Equal(t, len(responseObj.Reports), 1)
	assert.Equal(t, request.Name, responseObj.Reports[0].Name)
	assert.Equal(t, reportId.String(), responseObj.Reports[0].Id)
	assert.Equal(t, responseObj.Reports[0].TenantId, int32(1))
}

func TestListReports_Failed(t *testing.T) {
	_, dao, reportService := initTestSubject(t)
	dao.On("GetReports", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, &api.Error{Code: http.StatusBadRequest, Message: "Failed to list reports"})

	response, err := reportService.ListReports(ctx, []string{}, "", "", 0, 0, "")
	assert.NotNil(t, err)
	assert.Equal(t, response.Code, http.StatusBadRequest)
}

func TestDeleteReport_Ok(t *testing.T) {
	_, dao, reportService := initTestSubject(t)
	dao.On("DeleteReport", mock.Anything, mock.Anything, report.ID.String()).Return(nil)

	response, err := reportService.DeleteReportById(ctx, report.ID.String())
	assert.Nil(t, err)
	assert.Equal(t, response.Code, http.StatusNoContent)
}

func TestDeleteReport_Failed(t *testing.T) {
	_, dao, reportService := initTestSubject(t)
	dao.On("DeleteReport", mock.Anything, mock.Anything, report.ID.String()).Return(&api.Error{Code: http.StatusTooManyRequests, Message: "Failed to delete report"})

	response, err := reportService.DeleteReportById(ctx, report.ID.String())
	assert.NotNil(t, err)
	assert.Equal(t, response.Code, http.StatusTooManyRequests)
}

func TestBatchDeleteReports_AcceptedAndNotFound(t *testing.T) {
	_, dao, reportService := initTestSubject(t)
	existingId := report.ID.String()
	missingId := uuid.NewString()

	dao.On("UpdatePendingDelete", mock.Anything, int32(1), []string{existingId, missingId}).
		Return([]string{existingId}, nil).Once()
	dao.On("GetReportsByIds", mock.Anything, int32(1), []string{existingId, missingId}, []queryapi.Filter{processingStatusFilter}).
		Return([]models.Report{}, nil).Once()

	response, err := reportService.BatchDeleteReports(ctx, api.ReportsIdsBatchRequest{Ids: []string{existingId, missingId}})
	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, response.Code)

	responseObj, ok := response.Body.(api.BatchResponseAsync)
	assert.True(t, ok)
	assert.NotEmpty(t, responseObj.Id)
	assert.Equal(t, []string{existingId}, responseObj.Accepted)
	assert.Len(t, responseObj.Failed, 1)
	assert.Equal(t, missingId, responseObj.Failed[0].Id)
	assert.Equal(t, int32(http.StatusNotFound), responseObj.Failed[0].Code)
}

func TestBatchDeleteReports_Empty(t *testing.T) {
	_, dao, reportService := initTestSubject(t)

	dao.On("UpdatePendingDelete", mock.Anything, int32(1), []string{}).
		Return([]string{}, nil).Once()
	dao.On("GetReportsByIds", mock.Anything, int32(1), []string{}, []queryapi.Filter{processingStatusFilter}).
		Return([]models.Report{}, nil).Once()

	response, err := reportService.BatchDeleteReports(ctx, api.ReportsIdsBatchRequest{Ids: []string{}})
	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, response.Code)

	responseObj, ok := response.Body.(api.BatchResponseAsync)
	assert.True(t, ok)
	assert.Empty(t, responseObj.Accepted)
	assert.Empty(t, responseObj.Failed)
}

func TestBatchDeleteReports_TenantIsolation(t *testing.T) {
	_, dao, reportService := initTestSubject(t)
	existingId := report.ID.String()

	dao.On("UpdatePendingDelete", mock.Anything, int32(1), []string{existingId}).
		Return([]string{existingId}, nil).Once()
	dao.On("GetReportsByIds", mock.Anything, int32(1), []string{existingId}, []queryapi.Filter{processingStatusFilter}).
		Return([]models.Report{}, nil).Once()

	response, err := reportService.BatchDeleteReports(ctx, api.ReportsIdsBatchRequest{Ids: []string{existingId}})
	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, response.Code)
}

func TestBatchDeleteReports_DbError(t *testing.T) {
	_, dao, reportService := initTestSubject(t)
	existingId := report.ID.String()

	dao.On("UpdatePendingDelete", mock.Anything, int32(1), []string{existingId}).
		Return(nil, &api.Error{Code: http.StatusInternalServerError, Message: "failed"}).Once()

	response, err := reportService.BatchDeleteReports(ctx, api.ReportsIdsBatchRequest{Ids: []string{existingId}})
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestBatchDeleteReports_ProcessingFails(t *testing.T) {
	_, dao, reportService := initTestSubject(t)
	readyId := uuid.NewString()
	processingId := uuid.NewString()
	processingReport := models.Report{ID: uuid.MustParse(processingId), Status: string(api.REPORTSTATUS_PROCESSING)}

	dao.On("UpdatePendingDelete", mock.Anything, int32(1), []string{readyId, processingId}).
		Return([]string{readyId}, nil).Once()
	dao.On("GetReportsByIds", mock.Anything, int32(1), []string{readyId, processingId}, []queryapi.Filter{processingStatusFilter}).
		Return([]models.Report{processingReport}, nil).Once()

	response, err := reportService.BatchDeleteReports(ctx, api.ReportsIdsBatchRequest{Ids: []string{readyId, processingId}})
	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, response.Code)

	responseObj, ok := response.Body.(api.BatchResponseAsync)
	assert.True(t, ok)
	assert.Equal(t, []string{readyId}, responseObj.Accepted)
	assert.Len(t, responseObj.Failed, 1)
	assert.Equal(t, processingId, responseObj.Failed[0].Id)
	assert.Equal(t, int32(http.StatusProcessing), responseObj.Failed[0].Code)
}

func TestBatchDeleteReports_MarksNonProcessingAsPendingDelete(t *testing.T) {
	_, dao, reportService := initTestSubject(t)
	pendingId := uuid.NewString()
	readyId := uuid.NewString()

	dao.On("UpdatePendingDelete", mock.Anything, int32(1), []string{pendingId, readyId}).
		Return([]string{pendingId, readyId}, nil).Once()
	dao.On("GetReportsByIds", mock.Anything, int32(1), []string{pendingId, readyId}, []queryapi.Filter{processingStatusFilter}).
		Return([]models.Report{}, nil).Once()

	response, err := reportService.BatchDeleteReports(ctx, api.ReportsIdsBatchRequest{Ids: []string{pendingId, readyId}})
	assert.Nil(t, err)
	assert.Equal(t, http.StatusAccepted, response.Code)

	responseObj, ok := response.Body.(api.BatchResponseAsync)
	assert.True(t, ok)
	assert.ElementsMatch(t, []string{pendingId, readyId}, responseObj.Accepted)
	assert.Empty(t, responseObj.Failed)
}

func TestBatchGetReportsByIds_AllFound(t *testing.T) {
	_, dao, reportService := initTestSubject(t)
	dao.On("GetReportsByIds", mock.Anything, int32(1), []string{reportId.String()}).Return([]models.Report{report}, nil)

	response, err := reportService.BatchGetReportsByIds(ctx, api.ReportsIdsBatchRequest{Ids: []string{reportId.String()}})
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	responseObj, ok := response.Body.(api.ReportsBatchGetResponse)
	assert.True(t, ok)
	assert.Len(t, responseObj.Succeeded, 1)
	assert.Equal(t, reportId.String(), responseObj.Succeeded[0].Id)
	assert.Empty(t, responseObj.Failed)
	assert.NotEmpty(t, responseObj.Id)
}

func TestBatchGetReportsByIds_SomeNotFound(t *testing.T) {
	_, dao, reportService := initTestSubject(t)
	missingId := uuid.New().String()
	dao.On("GetReportsByIds", mock.Anything, int32(1), mock.MatchedBy(func(ids []string) bool {
		return len(ids) == 2
	})).Return([]models.Report{report}, nil)

	response, err := reportService.BatchGetReportsByIds(ctx, api.ReportsIdsBatchRequest{Ids: []string{
		reportId.String(),
		missingId,
	}})
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	responseObj, ok := response.Body.(api.ReportsBatchGetResponse)
	assert.True(t, ok)
	assert.Len(t, responseObj.Succeeded, 1)
	assert.Equal(t, reportId.String(), responseObj.Succeeded[0].Id)
	assert.Len(t, responseObj.Failed, 1)
	assert.Equal(t, missingId, responseObj.Failed[0].Id)
	assert.Equal(t, int32(http.StatusNotFound), responseObj.Failed[0].Code)
	assert.Equal(t, "requested resource does not exist", responseObj.Failed[0].Message)
}

func TestBatchGetReportsByIds_MixedFoundAndNotFound(t *testing.T) {
	_, dao, reportService := initTestSubject(t)
	missingId := uuid.New().String()
	anotherMissingId := uuid.New().String()
	dao.On("GetReportsByIds", mock.Anything, int32(1), []string{reportId.String(), missingId, anotherMissingId}).Return([]models.Report{report}, nil)

	response, err := reportService.BatchGetReportsByIds(ctx, api.ReportsIdsBatchRequest{Ids: []string{
		reportId.String(),
		missingId,
		anotherMissingId,
	}})
	assert.Nil(t, err)

	responseObj, ok := response.Body.(api.ReportsBatchGetResponse)
	assert.True(t, ok)
	assert.Len(t, responseObj.Succeeded, 1)
	assert.Len(t, responseObj.Failed, 2)
}

func TestBatchGetReportsByIds_TenantIsolation(t *testing.T) {
	_, dao, reportService := initTestSubject(t)
	dao.On("GetReportsByIds", mock.Anything, int32(1), []string{reportId.String()}).Return([]models.Report{}, nil)

	response, err := reportService.BatchGetReportsByIds(ctx, api.ReportsIdsBatchRequest{Ids: []string{reportId.String()}})
	assert.Nil(t, err)

	responseObj, ok := response.Body.(api.ReportsBatchGetResponse)
	assert.True(t, ok)
	assert.Empty(t, responseObj.Succeeded)
	assert.Len(t, responseObj.Failed, 1)
	assert.Equal(t, int32(http.StatusNotFound), responseObj.Failed[0].Code)
}

func TestBatchGetReportsByIds_EmptyInput(t *testing.T) {
	_, dao, reportService := initTestSubject(t)

	response, err := reportService.BatchGetReportsByIds(ctx, api.ReportsIdsBatchRequest{Ids: []string{}})
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	responseObj, ok := response.Body.(api.ReportsBatchGetResponse)
	assert.True(t, ok)
	assert.Empty(t, responseObj.Succeeded)
	assert.Empty(t, responseObj.Failed)
	dao.AssertNotCalled(t, "GetReportsByIds", mock.Anything, mock.Anything, mock.Anything)
}

func TestBatchGetReportsByIds_PassesAllIdsToDao(t *testing.T) {
	_, dao, reportService := initTestSubject(t)

	ids := make([]string, 101)
	for i := range ids {
		ids[i] = uuid.New().String()
	}
	dao.On("GetReportsByIds", mock.Anything, int32(1), ids).Return([]models.Report{}, nil)

	response, err := reportService.BatchGetReportsByIds(ctx, api.ReportsIdsBatchRequest{Ids: ids})
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	responseObj, ok := response.Body.(api.ReportsBatchGetResponse)
	assert.True(t, ok)
	assert.Empty(t, responseObj.Succeeded)
	assert.Len(t, responseObj.Failed, 101)
}

func TestBatchGetReportsByIds_DbError(t *testing.T) {
	_, dao, reportService := initTestSubject(t)
	dao.On("GetReportsByIds", mock.Anything, int32(1), []string{reportId.String()}).Return(nil, &api.Error{
		Code:    http.StatusInternalServerError,
		Message: "Failed to get reports",
	})

	response, err := reportService.BatchGetReportsByIds(ctx, api.ReportsIdsBatchRequest{Ids: []string{reportId.String()}})
	assert.Error(t, err)
	assert.Equal(t, "failed to get batchreports by ids", err.Error())
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}
