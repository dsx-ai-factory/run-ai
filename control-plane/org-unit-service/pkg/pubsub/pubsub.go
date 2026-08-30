package pubsub

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/apis"

	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	async_events "github.com/run-ai/runai/runai-application-packages/async-events"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
)

var once sync.Once

type pubSubHandler[T any] func(ctx context.Context, event *T) error

func Init() {
	runaipubsub.SetLogger(logging.Actual())
	_, err := runaipubsub.InitAsync(nil)
	if err != nil {
		logging.Fatalf("Failed to initialize pubsub: %v", err)
	}
}

func Register(notificationsService api.NotificationsAPIServicer, hardDeletionService apis.HardTenantDeletionAPIServicer) {
	once.Do(func() {
		runaipubsub.SubscribeAsync(getClusterDeletionHandler(notificationsService))
		runaipubsub.SubscribeAsync(getTenantDeletionHandler(hardDeletionService))
	})
}

func getClusterDeletionHandler(notificationsService api.NotificationsAPIServicer) pubSubHandler[async_events.ClusterDelete] {
	return func(ctx context.Context, event *async_events.ClusterDelete) error {
		response, err := notificationsService.ClusterDeleted(ctx, event.ClusterId)
		if err != nil {
			return err
		}

		if response.Code >= http.StatusBadRequest {
			return fmt.Errorf("failed to perform async cluster deletion. status code: %d error: %v", response.Code, response.Body)
		}

		return nil
	}
}

func getTenantDeletionHandler(hardDeletionService apis.HardTenantDeletionAPIServicer) pubSubHandler[async_events.TenantDelete] {
	return func(ctx context.Context, event *async_events.TenantDelete) error {
		return hardDeletionService.HardDelete(ctx, event.TenantId)
	}
}
