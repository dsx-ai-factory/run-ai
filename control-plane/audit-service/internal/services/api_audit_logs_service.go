package services

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/run-ai/runai/control-plane/audit-service/internal/database"
	"github.com/run-ai/runai/control-plane/audit-service/internal/database/models"
	"github.com/run-ai/runai/control-plane/audit-service/internal/generated/pkg/api"
	auth_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	auth_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"k8s.io/utils/ptr"
)

type AuditReader interface {
	Read(ctx context.Context, tenantId int, qp database.QueryParams) ([]*models.AuditLog, int, error)
}

type ApiAuditLogsService struct {
	dao             AuditReader
	authClient      auth_client.IAuthorizationClient
	deleteFilesChan chan string
}

func NewApiAuditLogsService(dao AuditReader, authClient auth_client.IAuthorizationClient) *ApiAuditLogsService {
	als := ApiAuditLogsService{dao: dao, authClient: authClient, deleteFilesChan: make(chan string, math.MaxInt16)}
	go als.deleteDownloadedFiles()
	return &als
}

func (als *ApiAuditLogsService) DownloadAuditLogs(ctx context.Context, start time.Time, end time.Time, fileType string, offset int32, numberOfSamples int32, filterBy []string) (api.ImplResponse, error) {
	if ok, statusCode, err := als.isAuditLogActionPermittedByUser(ctx); !ok {
		logging.Errorf("user is not permitted to read audit logs %v", err)
		return api.ImplResponse{Code: statusCode}, err
	}

	apiLogs, _, httpErr, err := als.getLogsFromDB(ctx, start, end, offset, numberOfSamples, "", "", filterBy, false)
	if err != nil {
		logging.Errorf("failed to fetch audit logs on download %v", err)
		return api.ImplResponse{Code: httpErr}, fmt.Errorf("failed to fetch audit logs %v", err)
	}

	file, err := os.CreateTemp("", fmt.Sprintf("audit_logs_*.%s", fileType))
	if err != nil {
		logging.Errorf("failed to create audit file %v", err)
		return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to create audit file")
	}
	defer file.Close()

	als.deleteFilesChan <- file.Name()

	err = als.writeLogsToFile(fileType, apiLogs, file)
	if err != nil {
		logging.Errorf("failed to write audit file %v", err)
		return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to write audit file %v", err)
	}

	return api.ImplResponse{Code: http.StatusOK, Body: file.Name()}, nil
}

func (als *ApiAuditLogsService) writeLogsToFile(fileType string, apiLogs []api.AuditLog, file *os.File) error {
	switch strings.ToLower(fileType) {
	case "json":
		return writeJSONFile(apiLogs, file)
	case "csv":
		return writeCSVFile(apiLogs, file)
	default:
		return fmt.Errorf("format %s not supported", fileType)
	}
}

func (als *ApiAuditLogsService) GetAuditLogs(ctx context.Context, start, end time.Time, offset, numberOfSamples int32, sortOrder, sortBy string, filterBy []string) (api.ImplResponse, error) {
	if ok, statusCode, err := als.isAuditLogActionPermittedByUser(ctx); !ok {
		logging.Errorf("user is not permitted to read audit logs %v", err)
		return api.ImplResponse{Code: statusCode}, err
	}

	apiLogs, totalCount, httpErr, err := als.getLogsFromDB(ctx, start, end, offset, numberOfSamples, sortOrder, sortBy, filterBy, true)
	if err != nil {
		logging.Errorf("failed to fetch audit logs %v", err)
		return api.ImplResponse{Code: httpErr}, fmt.Errorf("failed to fetch audit logs %v", err)
	}

	return api.ImplResponse{Code: http.StatusOK, Body: api.GetAuditLogs200Response{
		Next:      offset + int32(len(apiLogs)),
		Total:     int32(totalCount),
		AuditLogs: apiLogs,
	}}, nil
}

func (als *ApiAuditLogsService) getLogsFromDB(ctx context.Context, start time.Time, end time.Time, offset int32,
	numberOfSamples int32, sortOrder string, sortBy string, filterBy []string, withTotalCount bool,
) ([]api.AuditLog, int, int, error) {
	tenantId, err := parseTenantId(ctx)
	if err != nil {
		logging.Errorf("failed to parse tenant id: %v", err)
		return nil, 0, http.StatusBadRequest, fmt.Errorf("failed to parse tenant id: %v", err)
	}

	filterBy = unescapeSourceIpFilters(filterBy)
	filters, filterErr := models.QueryApiParser.Filters(filterBy)
	if filterErr != nil {
		logging.Errorf("error in filters %v", filterErr)
		return nil, 0, http.StatusBadRequest, fmt.Errorf("error in filters %v", filterErr)
	}

	sortByColumn, sortErr := models.QueryApiParser.GetMapping(sortBy)
	if sortErr != nil {
		logging.Errorf("error in sort by %v", sortErr)
		return nil, 0, http.StatusBadRequest, fmt.Errorf("error in sort by %v", sortErr)
	}

	qp := database.QueryParams{
		Offset:         int(offset),
		Samples:        int(numberOfSamples),
		Order:          sortOrder,
		SortBy:         sortByColumn,
		Start:          &start,
		End:            &end,
		Filters:        filters,
		WithTotalCount: withTotalCount,
	}
	logs, totalCount, err := als.dao.Read(ctx, tenantId, qp)
	if err != nil {
		logging.Errorf("failed to fetch audit logs from DB: %v", err)
		return nil, 0, http.StatusInternalServerError, fmt.Errorf("failed to fetch audit logs from DB")
	}

	return toApi(logs), totalCount, -1, nil
}

func (als *ApiAuditLogsService) deleteDownloadedFiles() {
	for filePath := range als.deleteFilesChan {
		go time.AfterFunc(2*time.Minute, func() {
			err := os.Remove(filePath)
			if err != nil {
				logging.Errorf("failed to delete audit file %s \n error: %v", filePath, err)
			}
		})
	}
}

func (als *ApiAuditLogsService) isAuditLogActionPermittedByUser(ctx context.Context) (bool, int, error) {
	tenantId, err := parseTenantId(ctx)
	if err != nil {
		return false, http.StatusBadRequest, err
	}

	permittedScopes, _, err := als.authClient.GetPermittedScopes(ctx, auth_api.RESOURCETYPE_EVENTS_HISTORY, ptr.To(auth_api.ACTION_READ))
	if err != nil {
		return false, http.StatusInternalServerError, errors.Wrap(err, "error while getting permitted scopes")
	}
	if permittedScopes == nil {
		return false, http.StatusInternalServerError, fmt.Errorf("error while getting permitted scopes")
	}

	if permittedScopes.Read.System != nil && *permittedScopes.Read.System {
		return true, http.StatusOK, nil
	}

	if contains := slices.Contains(permittedScopes.Read.Tenants, strconv.Itoa(tenantId)); contains {
		return true, http.StatusOK, nil
	}

	return false, http.StatusForbidden, fmt.Errorf("user is not permitted to read audit logs")
}

// Problem: source_ip values can contain commas (e.g. "1.1.1.1, 2.2.2.2"), but
// open API generated code separates filterBy values with commas to separate clauses - so the value would
// be separated as a result of the split and cause the filters validation to fail.
// Fix: the frontend escapes those commas with ';' and we swap them back here.
func unescapeSourceIpFilters(filterBy []string) []string {
	for i, f := range filterBy {
		if strings.HasPrefix(f, "source_ip") {
			filterBy[i] = strings.ReplaceAll(f, ";", ",")
		}
	}
	return filterBy
}
