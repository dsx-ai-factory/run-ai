package write

import (
	"context"
	"errors"

	openapi "github.com/run-ai/runai/control-plane/audit-service/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
)

type PubsubWriter struct{}

func NewPubSubWriter() (*PubsubWriter, error) {
	runaipubsub.SetLogger(logging.Actual())
	_, err := runaipubsub.InitAsync(nil)
	if err != nil {
		return nil, err
	}

	return &PubsubWriter{}, nil
}

func (psw *PubsubWriter) Write(log openapi.AuditLog) error {
	if !runaipubsub.IsInitialized() {
		return errors.New("pubsub is not initialized")
	}
	return runaipubsub.Publish(context.Background(), &log)
}
