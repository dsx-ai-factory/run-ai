package pubsub

import (
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/services"
	events "github.com/run-ai/runai/runai-application-packages/async-events"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
)

// Register registers event handlers to the pubsub
func Register() {
	runaipubsub.SubscribeAsync[events.ProjectDeleted](services.HandleProjectDeletedEvent)
	runaipubsub.SubscribeAsync[events.ClusterDelete](services.HandleClusterDeleteEvent)
	runaipubsub.SubscribeAsync[events.TenantDelete](services.HandleTenantDeleteEvent)
}
