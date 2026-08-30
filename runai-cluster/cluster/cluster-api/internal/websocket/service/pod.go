package service

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/workload"
	v1 "k8s.io/api/core/v1"
)

type PodsGetter interface {
	GetPod(ctx context.Context, workload *workload.Workload, preferredPod string) (*v1.Pod, error)
}

type SelectorFactoryFN func(workload *workload.Workload) workload.Selector
