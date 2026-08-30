package ingest

import (
	"github.com/run-ai/runai/control-plane/audit-service/internal/database/models"
)

type DbSinkHandler struct {
	batcher Batcherer[models.AuditLog]
}

func NewDbSinkHandler(batcher Batcherer[models.AuditLog]) *DbSinkHandler {
	return &DbSinkHandler{
		batcher: batcher,
	}
}

func (s *DbSinkHandler) Stop() error {
	s.batcher.WaitDone()
	return nil
}

func (s *DbSinkHandler) Handle(log *models.AuditLog) error {
	return s.batcher.Add(log)
}

func (s *DbSinkHandler) Next() Handler {
	return nil // this is final stop
}
