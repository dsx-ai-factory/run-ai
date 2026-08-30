package builders

import (
	"context"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	corev1 "k8s.io/api/core/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetPriorityClassName(ctx context.Context, k8sClient client.Client,
	priorityClassNameField *flds.StringField, podSpec *corev1.PodSpec) error {
	priorityClassName := priorityClassNameField.Deref()
	if podSpec == nil || priorityClassName == "" {
		return nil
	}

	err := validatePriorityClassExists(ctx, k8sClient, priorityClassName)
	if err != nil {
		return fmt.Errorf("failed to set priority class '%s' in pod spec, error: %s",
			priorityClassName, err.Error())
	}

	podSpec.PriorityClassName = priorityClassName
	return nil
}

func SetPriorityClassLabel(ctx context.Context, k8sClient client.Client,
	calculatedPriorityClassNameLabel string, objectMeta *metav1.ObjectMeta) error {
	if objectMeta.Labels == nil {
		objectMeta.Labels = make(map[string]string)
	}

	return SetPriorityClassLabelOnLabels(ctx, k8sClient, calculatedPriorityClassNameLabel, objectMeta.Labels)
}

func SetPreemptibilityLabel(calculatedPreemptibilityNameLabel string, objectMeta *metav1.ObjectMeta) {
	if calculatedPreemptibilityNameLabel == "" {
		return
	}
	if objectMeta.Labels == nil {
		objectMeta.Labels = make(map[string]string)
	}
	objectMeta.Labels[constants.LabelPreemptibilityName] = calculatedPreemptibilityNameLabel
}

func validatePriorityClassExists(ctx context.Context, k8sClient client.Client, priorityClassName string) error {
	if priorityClassName == "" {
		return nil
	}

	priorityClass := &schedulingv1.PriorityClass{}
	err := k8sClient.Get(ctx, client.ObjectKey{Name: priorityClassName}, priorityClass)
	if err != nil {
		return err
	}

	return nil
}
