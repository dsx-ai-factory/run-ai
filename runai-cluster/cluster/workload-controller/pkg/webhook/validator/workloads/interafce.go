package workloads

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
)

type WorkloadValidator interface {
	Validate(_ context.Context, job *jobs.JobWorkload) error
}
