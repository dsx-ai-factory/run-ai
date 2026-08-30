package ingest

import (
	"context"

	"github.com/run-ai/runai/control-plane/audit-service/internal/database/models"
)

type Dao interface {
	Store(ctx context.Context, batch []*models.AuditLog) error
}
