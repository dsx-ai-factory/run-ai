package pubsub

import (
	"fmt"
	"sync"

	"github.com/run-ai/runai/control-plane/cluster-service/pkg/services/async"
	clusterdelete "github.com/run-ai/runai/control-plane/cluster-service/pkg/services/cluster-delete"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub"
)

var once sync.Once

func Init() error {
	errorCallback := func(err error) {
		logging.Errorf("Error in pubsub: %v", err)
	}
	runaipubsub.SetLogger(logging.Actual())
	_, err := runaipubsub.InitAsync(errorCallback)
	if err != nil {
		return fmt.Errorf("failed to initialize pubsub: %v", err)
	}
	return nil
}

func Register(clusterDeleter clusterdelete.ClusterDeleter) {
	once.Do(func() {
		// Subscribe to events - all return the same global error channel
		clusterDeleteSubscriber := async.NewClusterDeleteSubscriber(clusterDeleter)
		clusterDeleteSubscriber.SubscribeClusterDelete()

		tenantSubscriber := async.NewTenantDeleteSubscriber(clusterDeleter)
		tenantSubscriber.SubscribeTenantDelete()
	})
}
