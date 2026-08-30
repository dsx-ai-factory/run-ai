package pubsubpackage

import (
	"github.com/run-ai/runai/control-plane/datavolumes-service/pkg/services"
	events "github.com/run-ai/runai/runai-application-packages/async-events"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
)

// Register registers event handlers to the pubsub
func Register() {
	runaipubsub.SubscribeAsync[events.ProjectDeleted](services.HandleProjectDeleteEvent)
	runaipubsub.SubscribeAsync[events.ClusterDelete](services.HandleClusterDeleteEvent)
	runaipubsub.SubscribeAsync[events.TenantDelete](services.HandleTenantDeleteEvent)
}
