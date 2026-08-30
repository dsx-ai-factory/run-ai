package common

import (
	"context"

	common "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	job "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type PvcBuilder struct {
	builder *workloadbuilders.PvcBuilder
}

func NewPvcBuilder(k8sClient client.Client) ResourceBuilder {
	return &PvcBuilder{
		builder: workloadbuilders.NewPvcBuilder(k8sClient),
	}
}

func (self *PvcBuilder) Build(ctx context.Context, workload *job.JobWorkload, options ResourceBuilderOptions) ([]client.Object, error) {
	var result []client.Object

	pvcList, err := self.GetPvcs(ctx, workload, &workload.ObjectMeta, options)
	if err != nil {
		return nil, err
	}
	for _, pvc := range pvcList {
		result = append(result, &pvc)
	}
	return result, nil
}

func (self *PvcBuilder) GetPreview(ctx context.Context, workload *job.JobWorkload) ([]common.Resource, error) {
	var result []common.Resource

	if workload == nil || workload.Spec.Pvcs == nil {
		return result, nil
	}

	isDistributed := isDistributionWorkload(workload)

	resultWorker, err := self.builder.GetPreview(ctx, &workload.ObjectMeta, workload.GetChildName(), workload, workload.Spec.Pvcs, GetPvcNameSuffix(isDistributed, false))
	if err != nil {
		return nil, err
	}

	resultWorkerCommon := common.CommonResourceToResource(resultWorker)
	result = append(result, resultWorkerCommon...)

	if workload.MasterSpec != nil {
		resultMaster, err := self.builder.GetPreview(ctx, &workload.ObjectMeta, workload.GetChildName(), workload, workload.MasterSpec.Pvcs, GetPvcNameSuffix(isDistributed, true))
		if err != nil {
			return nil, err
		}
		resultMasterCommon := common.CommonResourceToResource(resultMaster)
		result = append(result, resultMasterCommon...)
	}
	return result, nil
}

func (self *PvcBuilder) GetPvcs(ctx context.Context, workload *job.JobWorkload, topmostResource *metav1.ObjectMeta,
	options ResourceBuilderOptions) ([]corev1.PersistentVolumeClaim, error) {
	isDistributed := isDistributionWorkload(workload)

	jobType := GetJobType(workload)

	result, err := self.builder.Build(ctx, jobType, topmostResource, workload.GetChildName(), options.Owner, workload.Spec.Pvcs, GetPvcNameSuffix(isDistributed, false))
	if err != nil {
		return nil, err
	}
	if workload.MasterSpec != nil {
		resultMaster, err := self.builder.Build(ctx, jobType, topmostResource, workload.GetChildName(), options.Owner, workload.MasterSpec.Pvcs, GetPvcNameSuffix(isDistributed, true))
		if err != nil {
			return nil, err
		}
		result = append(result, resultMaster...)
	}

	if len(result) != 0 {
		return result, nil
	} else {
		return nil, nil
	}
}

func GetPvcNameSuffix(isDistributed bool, isMaster bool) string {
	if isDistributed {
		if isMaster {
			return "1"
		} else {
			return "0"
		}
	} else {
		return ""
	}
}

func isDistributionWorkload(workload *job.JobWorkload) bool {
	return workload.Kind == constants.KindDistributedWorkload
}
