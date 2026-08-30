package async_events

import (
	"context"

	"github.com/run-ai/runai/runai-common-packages/logging"

	event "github.com/run-ai/runai/runai-application-packages/async-events"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
)

//go:generate mockery
type TenantPublishServicer interface {
	PublishTenantDeleteEvent(ctx context.Context, event *event.TenantDelete) error
}

type TenantPublishService struct{}

func NewTenantPublishService() TenantPublishServicer {
	return &TenantPublishService{}
}

func (pubsub *TenantPublishService) PublishTenantDeleteEvent(ctx context.Context, e *event.TenantDelete) error {
	logging.Infof("publishing delete tenant event for tenant: %s", e.TenantId)
	err := runaipubsub.Publish[event.TenantDelete](ctx, e)
	if err != nil {
		return err
	}
	return nil
}
