package factory

import (
	"context"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type JobMergeCreator interface {
	CreateMergeStack(ctx context.Context, workloadObject client.Object) (map[api.SubresourceComponent]*enforcer.MergeStack, error)
}
