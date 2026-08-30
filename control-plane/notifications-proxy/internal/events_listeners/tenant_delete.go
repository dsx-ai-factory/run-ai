package events_listeners

import (
	"context"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/tenant_delete"

	events "github.com/run-ai/runai/runai-application-packages/async-events"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
)

type TenantDeleteClienter interface {
	Delete(ctx context.Context, tenantId string) error
}

type TenantDeleteListener struct {
	client TenantDeleteClienter
}

func NewTenantDeleteListener() *TenantDeleteListener {
	return &TenantDeleteListener{
		client: tenant_delete.NewClient(),
	}
}

func (el *TenantDeleteListener) Listen(_ context.Context) {
	runaipubsub.SubscribeAsync[events.TenantDelete](el.tenantDeleteCallback)
}

func (el *TenantDeleteListener) tenantDeleteCallback(ctx context.Context, event *events.TenantDelete) error {
	logging.Infof("Handling tenant delete request for tenant %s: %+v", event.TenantId, event)
	if err := el.client.Delete(ctx, event.TenantId); err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("client failed to delete tenant subscriptions and channel, tenant id: %s, error: %v", event.TenantId, err)
		return err
	}

	logging.Infof("Successfully deleted subscriptions on tenant delete event. tenant id: %s", event.TenantId)
	return nil
}
