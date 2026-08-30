package ingest

import (
	"github.com/run-ai/runai/control-plane/audit-service/internal/database/models"
)

type Handler interface {
	Stop() error
	Handle(log *models.AuditLog) error
	Next() Handler
}
