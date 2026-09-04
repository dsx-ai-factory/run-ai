package clusters

import (
	"context"
	"net/http"

	clusterApi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
)

type SimpleMock struct {
	clusters map[string]*clusterApi.DisplayedCluster
}

func NewSimpleMock() *SimpleMock {
	return &SimpleMock{
		clusters: make(map[string]*clusterApi.DisplayedCluster),
	}
}

func (sm *SimpleMock) WithCluster(id string, info *clusterApi.DisplayedCluster) *SimpleMock {
	sm.clusters[id] = info
	return sm
}

func (sm *SimpleMock) GetClusterByUuid(ctx context.Context, id string) (*clusterApi.DisplayedCluster, *clusterApi.Error) {
	info := sm.clusters[id]
	if info == nil {
		return nil, clusterApi.NewError(http.StatusNotFound, "cluster not found")
	}

	return info, nil
}
