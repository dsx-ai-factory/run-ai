package runaipubsub

import (
	"context"

	config_models "github.com/run-ai/runai/runai-common-packages/runaipubsub/config"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/logger"
)

type PubSuber interface {
	Connect(url, user, password string) error
	Publish(topic string, payload []byte) error
	Subscribe(ctx context.Context, topic string, callback func(event []byte) error) error
	SubscribeAsync(ctx context.Context, topic string, callback func(event []byte) error, errChan chan error)
	Close() error
}

type PubSubFactoryer interface {
	Create(logger logger.Loggerer, psType config_models.PubSubType) (PubSuber, error)
}

type PubsubConnectorer interface {
	Connect() error
	ConnectAsync(ctx context.Context) chan error
}
