package common

import (
	"context"
	"fmt"

	kubeflowcommon "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/mnnvlutils"
	dp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	workloadCommon "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"k8s.io/utils/ptr"
)

func (db *DistributedBuilder) BuildMasterReplicaSpec(ctx context.Context, masterSpec *kubeflowcommon.ReplicaSpec, masterReplicaName workloadCommon.ReplicaType) error {

	// We reach this point only if there's a master (NoMaster is not set),
	if db.MergedWorkload.MasterSpec == nil {
		return fmt.Errorf("master spec is nil, but NoMaster is not set")
	}

	// Therefore, the number of master replicas must be 1 (kubeflow defines bigger values as invalid)
	masterSpec.Replicas = ptr.To(int32(1))

	err := BuildAnnotationsAndLabels(ctx, db.K8sClient, &masterSpec.Template.ObjectMeta, db.MergedWorkload,
		db.MergedWorkload.MasterSpec.Gpu, db.MergedWorkload.MasterSpec.GpuMemory, db.MergedWorkload.MasterSpec.GpuDevices,
		db.PodLabels, db.PodAnnotations, db.MergedWorkload.MasterSpec.RunaiSchedulerFields)
	if err != nil {
		return fmt.Errorf("failed to build master  annotations and labels: %v", err)
	}

	workloadbuilders.AddUserDefinedAnnotationsAndLabels(&masterSpec.Template.ObjectMeta, &db.MergedWorkload.MasterSpec.ContainerFields)

	// initializing a struct of type DistributedSpec with all the shared
	// fields, to use the shared function buildPodSpec
	distMasterSpec := dp.DistributedSpec{
		WorkloadFields:           db.MergedWorkload.MasterSpec.WorkloadFields,
		ResourceAllocationFields: db.MergedWorkload.MasterSpec.ResourceAllocationFields,
		RunaiSchedulerFields:     db.MergedWorkload.MasterSpec.RunaiSchedulerFields,
		AccessControlFields:      db.MergedWorkload.MasterSpec.AccessControlFields,
		SchedulingFields:         db.MergedWorkload.Spec.SchedulingFields, // use the scheduling fields from the workload spec
	}
	mnnvlStatus := mnnvlutils.FromContext(ctx)
	newMasterPodSpec, err := db.buildPodSpec(ctx, &masterSpec.Template.Spec,
		&distMasterSpec, masterReplicaName, mnnvlStatus.MasterMode)

	if err != nil {
		return err
	}
	if newMasterPodSpec == nil {
		return fmt.Errorf("failed to build master pod spec is nil")
	}
	masterSpec.Template.Spec = *newMasterPodSpec

	return nil
}
