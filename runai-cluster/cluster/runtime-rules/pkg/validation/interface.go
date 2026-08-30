package validation

import "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"

type Validator interface {
	Validate(workload *jobs.JobWorkload) error
}
