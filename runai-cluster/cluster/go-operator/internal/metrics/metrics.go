package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	OperandStatusConditionReady = "ready"
)

var (
	OperandStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "runai_operator_operand_condition_status",
			Help: "Status of operand resources managed by the operator",
		},
		[]string{"name", "condition"},
	)

	ClusterInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "runai_operator_cluster_info",
			Help: "Cluster Information",
		},
		[]string{"tenant_id", "cluster_id", "cluster_name", "version", "platform_type"},
	)
)

func init() {
	metrics.Registry.MustRegister(OperandStatus)
	metrics.Registry.MustRegister(ClusterInfo)
}
