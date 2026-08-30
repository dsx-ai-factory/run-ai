package write

import (
	"context"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	client_utils "github.com/run-ai/runai/runai-application-packages/client-utils"
	"github.com/run-ai/runai/control-plane/audit-service/pkg/client/api"
	auditconfig "github.com/run-ai/runai/control-plane/audit-service/pkg/client/config"
	openapi "github.com/run-ai/runai/control-plane/audit-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

const (
	maxWriteRetries = 3
	writeRetryDelay = 2 * time.Second
)

type AuditSaver interface {
	SaveAuditLog(ctx context.Context, auditLog openapi.AuditLog) error
}

type APIWriter struct {
	client          AuditSaver
	maxWriteRetries uint
	writeRetryDelay time.Duration
}

func NewAPIWriter() *APIWriter {
	client := initAuditClient()
	return &APIWriter{client: client, maxWriteRetries: maxWriteRetries, writeRetryDelay: writeRetryDelay}
}

func (w *APIWriter) Write(log openapi.AuditLog) error {
	go func() {
		err := retry.Retry(
			func(u uint) error {
				return w.client.SaveAuditLog(context.Background(), log)
			},
			strategy.Wait(w.writeRetryDelay),
			strategy.Limit(w.maxWriteRetries),
		)
		if err != nil {
			logging.Warnf("Failed writing audit log using API writer: %v", err)
		}
	}()
	return nil
}

func initAuditClient() AuditSaver {
	auditServiceUrl := config.Get[*auditconfig.Config]().AuditServiceUrl
	config := openapi.NewConfiguration()
	config.Servers = openapi.ServerConfigurations{{URL: client_utils.GetInternalUrl(auditServiceUrl)}}
	return &api.AuditClient{ApiClient: openapi.NewAPIClient(config)}
}
