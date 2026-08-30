package datasources

import (
	"context"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/connections"
)

type ConnectionGetter interface {
	GetConnection(ctx context.Context, tenantId int32) (*connections.Connection, error)
}
