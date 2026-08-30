package topology_controller

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	crdPollingInterval = 60 * time.Second
)

var startOnce sync.Once

// StartWhenCRDReady polls for the grove ClusterTopologyBinding CRD and calls onReady
// once the CRD is installed. onReady is guaranteed to be called at most once,
// even if StartWhenCRDReady is invoked multiple times.
// This should be called after the manager has started (e.g., after leader election).
func StartWhenCRDReady(ctx context.Context, c client.Reader, onReady func()) {
	go func() {
		log.Info().Msgf("Waiting for grove ClusterTopologyBinding CRD <%s> to be installed", groveCRDName)

		ticker := time.NewTicker(crdPollingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Context cancelled, stopping grove CRD watcher")
				return
			case <-ticker.C:
				if isGroveCRDInstalled(ctx, c) {
					log.Info().Msgf("Grove ClusterTopologyBinding CRD <%s> detected", groveCRDName)
					startOnce.Do(onReady)
					return
				}
			}
		}
	}()
}

func isGroveCRDInstalled(ctx context.Context, c client.Reader) bool {
	return utils.IsCRDInstalled(ctx, c, groveCRDName)
}
