package write

import (
	openapi "github.com/run-ai/runai/control-plane/audit-service/pkg/client/generated"
)

type EmptyWriter struct{}

func (d *EmptyWriter) Write(log openapi.AuditLog) error {
	return nil
}
