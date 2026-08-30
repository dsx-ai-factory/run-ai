package services

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/run-ai/runai/control-plane/audit-service/internal/database/models"
	"github.com/run-ai/runai/control-plane/audit-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/audit-service/internal/ingest"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

type Batcher interface {
	Add(item *models.AuditLog) error
}

type InternalApiAuditLogsService struct {
	logsBatcher Batcher
}

func NewInternalApiAuditLogsService(logsBatcher Batcher) *InternalApiAuditLogsService {
	return &InternalApiAuditLogsService{logsBatcher}
}

func (ils *InternalApiAuditLogsService) SaveAuditLog(_ context.Context, log api.AuditLog) (api.ImplResponse, error) {
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	parsedLog, err := ingest.ParseToModelLog(&log)
	if err != nil {
		logging.Errorf("failed to parse log: %v", err)
		return api.ImplResponse{Code: http.StatusBadRequest}, fmt.Errorf("failed to parse log")
	}

	if err = ils.logsBatcher.Add(parsedLog); err != nil {
		logging.Errorf("failed to batch audit log, tenant [%s]: %v", log.TenantId, err)
		return api.ImplResponse{Code: http.StatusInternalServerError}, fmt.Errorf("failed to batch audit log for tenant [%s]", log.TenantId)
	}

	return api.ImplResponse{Code: http.StatusOK}, err
}
