package handlers

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/common/dra"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/handlers/commonparsers"
	kubeflowv1beta1 "github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/thirdpartyapi/kubeflownotebook/v1beta1"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	notebookConditionRunning    = "Running"
	notebookConditionReady      = "Ready"
	notebookConditionWaiting    = "Waiting"
	notebookConditionTerminated = "Terminated"
)

type KubeflowNotebookTopOwnerHandler struct {
	cachedClient client.Client
	draParser    *dra.DRAParser
}

func NewKubeflowNotebookTopOwnerHandler(cachedClient client.Client, draParser *dra.DRAParser) *KubeflowNotebookTopOwnerHandler {
	return &KubeflowNotebookTopOwnerHandler{
		cachedClient: cachedClient,
		draParser:    draParser,
	}
}

func (h *KubeflowNotebookTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &kubeflowv1beta1.Notebook{} })
}

func (h *KubeflowNotebookTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	notebook, ok := object.(*kubeflowv1beta1.Notebook)
	if !ok {
		return fmt.Errorf("failed converting object to kubeflowv1beta1.Notebook")
	}

	podTemplate := &corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       notebook.Spec.Template.Spec,
	}
	commonparsers.FillExternalWorkloadDataFromPodTemplateSpec(ctx, notebook.ObjectMeta, notebook.GroupVersionKind().GroupKind(), podTemplate,
		nil, h.cachedClient, h.draParser, externalWorkload)
	externalWorkload.Spec.RequestedPods.Replicas = ptr.To(int32(1))

	observedConditions := []v1alpha1.ObservedCondition{}
	for _, condition := range notebook.Status.Conditions {
		observedConditions = append(observedConditions, v1alpha1.ObservedCondition{
			Type:               condition.Type,
			Status:             corev1.ConditionStatus(condition.Status),
			Reason:             condition.Reason,
			Message:            condition.Message,
			LastTransitionTime: condition.LastTransitionTime,
		})

		if (condition.Type == notebookConditionRunning || condition.Type == notebookConditionReady) &&
			condition.Status == string(corev1.ConditionTrue) {
			externalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWRunningPhase
			externalWorkload.Status.ObservedStatus.K8sStatus = notebookConditionRunning
			externalWorkload.Status.ObservedStatus.Message = condition.Message
			continue
		}

		if condition.Type == notebookConditionWaiting && condition.Status == string(corev1.ConditionTrue) {
			externalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWPendingPhase
			externalWorkload.Status.ObservedStatus.K8sStatus = notebookConditionWaiting
			externalWorkload.Status.ObservedStatus.Message = condition.Message
			continue
		}

		if condition.Type == notebookConditionTerminated && condition.Status == string(corev1.ConditionTrue) {
			externalWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWCompletedPhase
			externalWorkload.Status.ObservedStatus.K8sStatus = notebookConditionTerminated
			externalWorkload.Status.ObservedStatus.Message = condition.Message
			continue
		}
	}

	externalWorkload.Status.ObservedStatus.Conditions = observedConditions

	return nil
}

func (h *KubeflowNotebookTopOwnerHandler) Filter(object client.Object) bool {
	notebook, ok := object.(*kubeflowv1beta1.Notebook)
	if !ok {
		return false
	}

	return notebook.Spec.Template.Spec.SchedulerName == constants.RunaiSchedulerName
}
