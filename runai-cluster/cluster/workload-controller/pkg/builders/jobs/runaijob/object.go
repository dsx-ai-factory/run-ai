package runaijob

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	workloadbuilders "github.com/run-ai/runai/runai-cluster/cluster/workload-common/builders"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	topowner "github.com/run-ai/runai/runai-cluster/common/top-owner"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetSpecObjectMeta(ctx context.Context, k8sClient client.Client, mergedWorkload *jobs.JobWorkload, podTemplateSpec *corev1.PodTemplateSpec,
	podLabels map[string]string, podAnnotations map[string]string) error {
	var memoryQuantity resource.Quantity

	som := &podTemplateSpec.ObjectMeta

	if som.Annotations == nil {
		som.Annotations = map[string]string{}
	}
	if som.Labels == nil {
		som.Labels = map[string]string{}
	}

	for k, v := range podLabels {
		som.Labels[k] = v
	}

	for k, v := range podAnnotations {
		som.Annotations[k] = v
	}

	workloadbuilders.AddUserDefinedAnnotationsAndLabels(som, &mergedWorkload.Spec.ContainerFields)
	som.Annotations[constants.AnnotationUserName] = mergedWorkload.Spec.UserName.Deref()

	priorityClassNameLabel := common.CalculatePriorityClassNameLabel(mergedWorkload.Spec.PriorityClassName.Deref(), mergedWorkload.Spec.Preemptibility, mergedWorkload.Kind)
	err := workloadbuilders.SetPriorityClassLabel(ctx, k8sClient, priorityClassNameLabel, som)
	if err != nil {
		return err
	}

	calculatedPreemptibilityNameLabel := common.CalculatePreemptibilityNameLabel(mergedWorkload.Spec.Preemptibility, mergedWorkload.Kind)
	workloadbuilders.SetPreemptibilityLabel(calculatedPreemptibilityNameLabel, som)

	if mergedWorkload.Spec.GpuMemory.IsValueInitialized() {
		memoryQuantity, err = mergedWorkload.Spec.GpuMemory.DerefQuantity(0)
		if err != nil {
			return fmt.Errorf("invalid gpuMemory: %s", err.Error())
		}
	}

	if workloadbuilders.IsFractionalGpuOrMemroy(ctx, mergedWorkload.Spec.Gpu.Deref(), memoryQuantity) {
		workloadbuilders.AddPartialGpuAnnotations(som.Annotations, mergedWorkload.Spec.Gpu, memoryQuantity, mergedWorkload.Spec.GpuDevices.Deref())
	}

	err = workloadbuilders.SetObjectMetaLabels(ctx, mergedWorkload.Kind, constants.KindRunaiJob, mergedWorkload.Name, som)
	if err != nil {
		return err
	}

	// the pod should be created with the top-owner label
	som.Labels[commonconstants.TopOwnerUidLabel] = topowner.GetRunaiWorkloadId(mergedWorkload.UID, mergedWorkload.Annotations)

	// set the used nodepools in an annotations
	nodepoolsValue := field_groups.GetNodepoolsValue(mergedWorkload.Spec.RunaiSchedulerFields)
	if nodepoolsValue != "" {
		som.Annotations[commonconstants.AnnotationNodepools] = nodepoolsValue
	}

	return nil
}
