package pubsub

import (
	"github.com/run-ai/runai/control-plane/workloads-manager/pkg/services"
	events "github.com/run-ai/runai/runai-application-packages/async-events"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
)

// Register registers event handlers to the pubsub
func Register() {
	runaipubsub.SubscribeAsync[events.TenantDelete](services.HandleTenantDeleteEvent)
	runaipubsub.SubscribeAsync[events.ClusterDelete](services.HandleClusterDeleteEvent)
	runaipubsub.SubscribeAsync[events.ProjectDeleted](services.HandleProjectDeletedEvent)
}
