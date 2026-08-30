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
	TaskStartedStatus  = "Started"
	TaskPendingStatus  = "Pending"
	TaskRunningStatus  = "Running"
	TaskCanceledStatus = "TaskRunCancelled"
)

type TaskRunTopOwnerHandler struct {
	cachedClient client.Client
}

func NewTaskRunTopOwnerHandler(cachedClient client.Client) *TaskRunTopOwnerHandler {
	return &TaskRunTopOwnerHandler{
		cachedClient: cachedClient,
	}
}

func (h *TaskRunTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &v1.TaskRun{} })
}

func (h *TaskRunTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	taskRun, ok := object.(*v1.TaskRun)
	if !ok {
		return fmt.Errorf("failed converting object to v1.TaskRun")
	}

	podSpecPriorityClassName := ""
	if taskRun.Spec.PodTemplate != nil && taskRun.Spec.PodTemplate.PriorityClassName != nil &&
		*taskRun.Spec.PodTemplate.PriorityClassName != "" {
		podSpecPriorityClassName = *taskRun.Spec.PodTemplate.PriorityClassName
	}
	externalWorkload.Spec.PriorityClassName, externalWorkload.Spec.Preemptibility = commonparsers.CalculatePriorityClassNameAndPreemptibility(ctx, h.cachedClient, taskRun.Labels,
		taskRun.GetGroupVersionKind().GroupKind(), podSpecPriorityClassName)

	phase, k8sStatus, message := calculateTaskRunStatus(taskRun)

	externalWorkload.Status.ObservedStatus.Phase = phase
	externalWorkload.Status.ObservedStatus.K8sStatus = k8sStatus
	externalWorkload.Status.ObservedStatus.Message = message

	observedConditions := []v1alpha1.ObservedCondition{}
	for _, condition := range taskRun.Status.Conditions {
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

// calculatePipelineRunStatus get the task status + reasons according to
// https://tekton.dev/docs/pipelines/taskruns/#monitoring-execution-status
func calculateTaskRunStatus(taskRun *v1.TaskRun) (v1alpha1.ExternalWorkloadPhase, string, string) {
	var phase v1alpha1.ExternalWorkloadPhase
	var k8sStatus string
	var message string
	for _, condition := range taskRun.Status.Conditions {
		if condition.Status == corev1.ConditionUnknown {
			switch condition.Reason {
			case TaskStartedStatus:
				phase = v1alpha1.ExWCreatingPhase
			case TaskPendingStatus:
				phase = v1alpha1.ExWPendingPhase
			case TaskRunningStatus:
				phase = v1alpha1.ExWRunningPhase
			case TaskCanceledStatus:
				phase = v1alpha1.ExWStoppedPhase
			}
			k8sStatus = condition.Reason
			message = condition.Reason
			break
		} else if condition.Status == corev1.ConditionFalse {
			phase = v1alpha1.ExWFailedPhase
			if condition.Reason == TaskCanceledStatus {
				phase = v1alpha1.ExWStoppedPhase
			}
			k8sStatus = string(phase)
			message = fmt.Sprintf("%s. %s", condition.Reason, condition.Message)
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
