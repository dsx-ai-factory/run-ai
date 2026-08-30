package convertors

import (
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Convertor interface {
	Convert(jobWorkload *jobs.JobWorkload) client.Object
}
