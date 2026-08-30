package resources

import "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"

const (
	name = operands.RunaiClusterApiName

	serviceInPort     = 4190
	serviceTargetPort = 4180

	servicePortName = "cluster-api-service"
	metricPortName  = "metrics"
)

var (
	appStringMap = map[string]string{
		"app": name,
	}
)
