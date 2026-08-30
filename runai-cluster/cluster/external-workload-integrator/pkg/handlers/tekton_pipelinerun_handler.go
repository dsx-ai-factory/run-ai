package handlers

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PipelineStartedStatus  = "Started"
	PipelinePendingStatus  = "Pending"
	PipelineRunningStatus  = "Running"
	PipelineCanceledStatus = "Cancelled"
)

type PipelineRunTopOwnerHandler struct {
	cachedClient client.Client
}

func NewPipelineRunTopOwnerHandler(cachedClient client.Client) *PipelineRunTopOwnerHandler {
	return &PipelineRunTopOwnerHandler{
		cachedClient: cachedClient,
	}
}

func (h *PipelineRunTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &v1.PipelineRun{} })
}

func (h *PipelineRunTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	pipelineRun, ok := object.(*v1.PipelineRun)
	if !ok {
		return fmt.Errorf("failed converting object to v1.PipelineRun")
	}

	podSpecPriorityClassName := ""
	for _, taskRunSpec := range pipelineRun.Spec.TaskRunSpecs {
		if podSpecPriorityClassName == "" && taskRunSpec.PodTemplate != nil &&
			taskRunSpec.PodTemplate.PriorityClassName != nil && *taskRunSpec.PodTemplate.PriorityClassName != "" {
			podSpecPriorityClassName = *taskRunSpec.PodTemplate.PriorityClassName
		}
	}
	externalWorkload.Spec.PriorityClassName, externalWorkload.Spec.Preemptibility = commonparsers.CalculatePriorityClassNameAndPreemptibility(ctx, h.cachedClient, pipelineRun.Labels,
		pipelineRun.GetGroupVersionKind().GroupKind(), podSpecPriorityClassName)

	phase, k8sStatus, message := calculatePipelineRunStatus(pipelineRun)

	externalWorkload.Status.ObservedStatus.Phase = phase
	externalWorkload.Status.ObservedStatus.K8sStatus = k8sStatus
	externalWorkload.Status.ObservedStatus.Message = message

	observedConditions := []v1alpha1.ObservedCondition{}
	for _, condition := range pipelineRun.Status.Conditions {
		observedConditions = append(observedConditions, v1alpha1.ObservedCondition{
			Type:               string(condition.Type),
			Status:             condition.Status,
			Reason:             condition.Reason,
			Message:            condition.Message,
			LastTransitionTime: condition.LastTransitionTime.Inner,
		})
	}
	externalWorkload.Status.ObservedStatus.Conditions = observedConditions

	return nil
}

// calculatePipelineRunStatus get the pipeline status + reasons according to
// https://tekton.dev/docs/pipelines/pipelineruns/#monitoring-execution-status
func calculatePipelineRunStatus(taskRun *v1.PipelineRun) (v1alpha1.ExternalWorkloadPhase, string, string) {
	var phase v1alpha1.ExternalWorkloadPhase
	var k8sStatus string
	var message string
	for _, condition := range taskRun.Status.Conditions {
		if condition.Status == corev1.ConditionUnknown {
			switch condition.Reason {
			case PipelineStartedStatus:
				phase = v1alpha1.ExWCreatingPhase
			case PipelineRunningStatus:
				phase = v1alpha1.ExWRunningPhase
			case PipelineCanceledStatus:
				phase = v1alpha1.ExWStoppedPhase
			}
			k8sStatus = condition.Reason
			message = condition.Reason
			break
		} else if condition.Status == corev1.ConditionFalse {
			phase = v1alpha1.ExWFailedPhase
			if condition.Reason == PipelineCanceledStatus {
				phase = v1alpha1.ExWStoppedPhase
			}
			k8sStatus = string(v1alpha1.ExWFailedPhase)
			message = condition.Reason
			break
		} else if condition.Status == corev1.ConditionTrue {
			phase = v1alpha1.ExWCompletedPhase
			k8sStatus = condition.Reason
			message = condition.Reason
			break
		}
	}
	return phase, k8sStatus, message
}
