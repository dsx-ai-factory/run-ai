package finalizers

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/workloads/convertors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	log2 "sigs.k8s.io/controller-runtime/pkg/log"
)

type WorkloadRemoveFinalizer struct {
	client               client.Client
	workloadType         constants.WorkloadType
	JobWorkloadConvertor convertors.Convertor
}

func NewWorkloadRemoveFinalizer(client client.Client, workloadType constants.WorkloadType, JobWorkloadConvertor convertors.Convertor) *WorkloadRemoveFinalizer {
	return &WorkloadRemoveFinalizer{
		client:               client,
		workloadType:         workloadType,
		JobWorkloadConvertor: JobWorkloadConvertor,
	}
}

func (wrf *WorkloadRemoveFinalizer) RemoveFinalizer(ctx context.Context, job *jobs.JobWorkload) error {
	workload := wrf.JobWorkloadConvertor.Convert(job)
	log := log2.FromContext(ctx)
	log.V(4).Info("Removing finalizer from workload", "name", workload.GetName(), "namespace", workload.GetNamespace(), "workloadType", wrf.workloadType)
	controllerutil.RemoveFinalizer(workload, constants.FinalizerName)
	err := wrf.client.Update(ctx, workload)
	if err != nil {
		log.Error(err, "Failed to remove finalizer from workload", "name", workload.GetName(), "namespace", workload.GetNamespace(), "workloadType", wrf.workloadType)
		return fmt.Errorf("failed to remove finalizer from workload: %v", err)
	}
	return nil
}

type InteractiveWorkloadRemoveFinalizerFactory struct {
}

func (self *InteractiveWorkloadRemoveFinalizerFactory) CreateRemoveFinalizer(client client.Client) RemoveFinalizer {
	return NewWorkloadRemoveFinalizer(client, constants.InteractiveWorkload, convertors.NewInteractiveWorkloadConvertor())
}

type TrainingWorkloadRemoveFinalizerFactory struct {
}

func (self *TrainingWorkloadRemoveFinalizerFactory) CreateRemoveFinalizer(client client.Client) RemoveFinalizer {
	return NewWorkloadRemoveFinalizer(client, constants.TrainingWorkload, convertors.NewTrainingWorkloadConvertor())
}

type DistributedWorkloadRemoveFinalizerFactory struct {
}

func (self *DistributedWorkloadRemoveFinalizerFactory) CreateRemoveFinalizer(client client.Client) RemoveFinalizer {
	return NewWorkloadRemoveFinalizer(client, constants.DistributedWorkload, convertors.NewDistributedWorkloadConvertor())
}
