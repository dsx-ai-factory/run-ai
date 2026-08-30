package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/run-ai/runai/runai-cluster/cluster/nodepool-controller/pkg/config"
)

// nodeNodePool is constructed lazily on first access because its Subsystem
// depends on config.Get().MetricsNamespace, which is populated by flag
// parsing — i.e., it is not available at package-init time.
var (
	nodeNodePool     *prometheus.GaugeVec
	nodeNodePoolOnce sync.Once
)

func getNodeNodePool() *prometheus.GaugeVec {
	nodeNodePoolOnce.Do(func() {
		nodeNodePool = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: config.Get().MetricsNamespace,
				Name:      "node_nodepool",
				Help:      "the node's nodepool",
			}, []string{"node", "nodepool"},
		)
	})
	return nodeNodePool
}

func SetNodeNodePool(nodeName string, nodePoolName string) {
	getNodeNodePool().WithLabelValues(nodeName, nodePoolName).Set(1)
}

func RemoveNodeNodePool(nodeName string, nodePoolName string) {
	getNodeNodePool().WithLabelValues(nodeName, nodePoolName).Set(0)
}

func GetNodePoolMetric() *prometheus.GaugeVec {
	return getNodeNodePool()
}

func Init() {
	// Register custom metrics with the global prometheus registry.
	metrics.Registry.MustRegister(getNodeNodePool())
}
