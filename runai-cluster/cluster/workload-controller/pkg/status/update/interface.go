package update

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatusUpdater interface {
	UpdateStatus(ctx context.Context, modifiedWorkload *jobs.JobWorkload) error
}

type IsStatusChanged func(ctx context.Context, modifiedWorkload *jobs.JobWorkload, originalWorkload *jobs.JobWorkload) bool

type PreUpdateStatusFunc func(ctx context.Context, modifiedWorkload *jobs.JobWorkload) error

type StatusUpdaterFactory interface {
	CreateStatusUpdater(client client.Client) StatusUpdater
}

func GetStatusUpdaterFactory(workloadType constants.WorkloadType) (StatusUpdaterFactory, error) {
	switch workloadType {
	case constants.InteractiveWorkload:
		return &InteractiveWorkloadStatusUpdaterFactory{}, nil
	case constants.TrainingWorkload:
		return &TrainingWorkloadStatusUpdaterFactory{}, nil
	case constants.DistributedWorkload:
		return &DistributedWorkloadStatusUpdaterFactory{}, nil
	default:
		return nil, fmt.Errorf("unknown workload type: %s", workloadType)
	}
}
