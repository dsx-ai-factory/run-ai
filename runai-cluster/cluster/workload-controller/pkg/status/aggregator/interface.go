package aggregator

import (
	"context"
	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common/resource"
	jobInterface "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/jobs"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WorkloadStatusCalculater interface {
	WithWorkload(workload *jobs.JobWorkload)
	WithResources(resource []client.Object)

	WithServices(services []*v1.Service)
	WithJob(job jobInterface.JobAdapter)
	WithComputeDomain(computeDomain *nvapi.ComputeDomain)
	WithSecrets(secrets []*v1.Secret)
	WithExpectedResources(resources []common.Resource)
	// Clear clears the status, and the resources added to struct before,
	//aka by using WithResources, WithServices, WithJob, WithExpectedResources
	Clear()
	CalculateAll(ctx context.Context, aggr *resource.Aggregator) jobs.JobStatus
	CalculatePhaseAndConditions(ctx context.Context, aggr *resource.Aggregator) jobs.JobStatus
}
