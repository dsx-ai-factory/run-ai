package update

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/workloads/convertors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

type StatusPatcher interface {
	PatchBytes(ctx context.Context, modifiedWorkload *jobs.JobWorkload) ([]byte, error)
}

type WorkloadUpdater struct {
	client               client.Client
	WorkloadType         constants.WorkloadType
	StatusPatcher        StatusPatcher
	JobWorkloadConvertor convertors.Convertor
}

func NewWorkloadUpdater(client client.Client, workloadType constants.WorkloadType, convertor convertors.Convertor,
	patcher StatusPatcher) *WorkloadUpdater {
	return &WorkloadUpdater{
		client:               client,
		WorkloadType:         workloadType,
		JobWorkloadConvertor: convertor,
		StatusPatcher:        patcher,
	}
}

// UpdateStatus updates the status of the workload
func (wu *WorkloadUpdater) UpdateStatus(ctx context.Context, modifiedWorkload *jobs.JobWorkload) error {
	if modifiedWorkload == nil {
		return fmt.Errorf("workload canot be nil")
	}
	log := logger.FromContext(ctx).WithValues("name", modifiedWorkload.Name, "namespace", modifiedWorkload.Namespace)
	log.V(4).Info(fmt.Sprintf("Updating %s workload status", wu.WorkloadType))

	patchBytes, err := wu.StatusPatcher.PatchBytes(ctx, modifiedWorkload)
	if err != nil {
		log.Error(err, "Failed to json.Marshal patch for update workload status",
			"name", modifiedWorkload.Name, "namespace", modifiedWorkload.Namespace, "status", modifiedWorkload.Status)
		return fmt.Errorf("failed to update interactive workload status: %v", err)
	}
	workload := wu.JobWorkloadConvertor.Convert(modifiedWorkload)

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = wu.client.Status().Patch(ctx, workload, patch)
	if err != nil {

		log.Error(err, fmt.Sprintf("Failed to patch %s workload status", wu.WorkloadType),
			"name", workload.GetName(), "namespace", workload.GetNamespace(), "status", modifiedWorkload.Status)
		return fmt.Errorf("failed to update %s workload status: %v", wu.WorkloadType, err)
	}
	return nil
}

type InteractiveWorkloadStatusUpdaterFactory struct {
}

func (self *InteractiveWorkloadStatusUpdaterFactory) CreateStatusUpdater(client client.Client) StatusUpdater {
	return NewWorkloadUpdater(client, constants.InteractiveWorkload, convertors.NewInteractiveWorkloadConvertor(), NewRunaiWorkloadPatcher())
}

type TrainingWorkloadStatusUpdaterFactory struct {
}

func (self *TrainingWorkloadStatusUpdaterFactory) CreateStatusUpdater(client client.Client) StatusUpdater {
	return NewWorkloadUpdater(client, constants.TrainingWorkload, convertors.NewTrainingWorkloadConvertor(), NewRunaiWorkloadPatcher())
}

type DistributedWorkloadStatusUpdaterFactory struct {
}

func (self *DistributedWorkloadStatusUpdaterFactory) CreateStatusUpdater(client client.Client) StatusUpdater {
	return NewWorkloadUpdater(client, constants.DistributedWorkload, convertors.NewDistributedWorkloadConvertor(), NewDistributedWorkloadPatcher())
}
