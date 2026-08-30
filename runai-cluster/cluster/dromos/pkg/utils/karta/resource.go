package karta

import (
	"context"
	"errors"
	"fmt"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	kartaresource "github.com/run-ai/karta/pkg/resource"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Factory struct {
	client client.Reader
}

func NewFactory(client client.Reader) *Factory {
	return &Factory{
		client: client,
	}
}

func (ru *Factory) GetKartaForGVK(ctx context.Context, gvk schema.GroupVersionKind) (*kartav1alpha1.Karta, error) {
	labelSelectors := GetKartaPerGvkLabelSelectors(gvk)
	kartaList := &kartav1alpha1.KartaList{}

	if err := ru.client.List(ctx, kartaList, labelSelectors); err != nil {
		return nil, fmt.Errorf("failed to list Kartas: %w", err)
	}

	for i := range kartaList.Items {
		karta := &kartaList.Items[i]

		kartaGVK, err := GetRootComponentGVK(karta)
		if err != nil {
			return nil, err
		}

		if *kartaGVK == gvk {
			return karta, nil
		}
	}

	return nil, apierrors.NewNotFound(
		kartav1alpha1.Resource("Karta"),
		fmt.Sprintf("for gvk %s", gvk.String()),
	)
}

func GetRootComponentGVK(karta *kartav1alpha1.Karta) (*schema.GroupVersionKind, error) {
	rootKind := karta.Spec.StructureDefinition.RootComponent.Kind
	if rootKind == nil {
		return nil, errors.New("root component kind is nil")
	}

	return &schema.GroupVersionKind{
		Group:   rootKind.Group,
		Version: rootKind.Version,
		Kind:    rootKind.Kind,
	}, nil
}

func IsReady(karta *kartav1alpha1.Karta) bool {
	ready := meta.FindStatusCondition(karta.Status.Conditions, string(Ready))
	return ready != nil && ready.Status == metav1.ConditionTrue
}

// GetResourcePhase extracts the phase from a resource using the Karta status utils.
func (ru *Factory) GetResourcePhase(ctx context.Context, karta *kartav1alpha1.Karta, resource *unstructured.Unstructured) (dromosv1alpha1.WorkloadRunnerPhase, error) {
	factory := kartaresource.NewComponentFactoryFromObject(karta, resource)

	rootComponent, err := factory.GetRootComponent()
	if err != nil {
		return dromosv1alpha1.WorkloadRunnerPhaseUnknown, fmt.Errorf("failed to get root component: %w", err)
	}

	status, err := rootComponent.GetStatus(ctx)
	if err != nil {
		return dromosv1alpha1.WorkloadRunnerPhaseUnknown, fmt.Errorf("failed to get status from root component: %w", err)
	}

	if status == nil {
		return dromosv1alpha1.WorkloadRunnerPhaseUnknown, nil
	}

	return mapResourceStatusToWorkloadRunnerPhase(status.MatchedStatuses), nil
}

// mapResourceStatusToWorkloadRunnerPhase maps Karta ResourceStatus values to WorkloadRunnerPhase.
// Priority order (first match wins):
//
// Failed ranks above Suspending/Suspended so that a pod crash during suspension is surfaced immediately.
func mapResourceStatusToWorkloadRunnerPhase(statuses []kartav1alpha1.ResourceStatus) dromosv1alpha1.WorkloadRunnerPhase {
	statusesSet := make(map[kartav1alpha1.ResourceStatus]bool)
	for _, status := range statuses {
		statusesSet[status] = true
	}

	if statusesSet[kartav1alpha1.FailedStatus] {
		return dromosv1alpha1.WorkloadRunnerPhaseFailed
	}
	if statusesSet[kartav1alpha1.CompletedStatus] {
		return dromosv1alpha1.WorkloadRunnerPhaseCompleted
	}
	if statusesSet[kartav1alpha1.ResumingStatus] {
		return dromosv1alpha1.WorkloadRunnerPhaseResuming
	}
	if statusesSet[kartav1alpha1.SuspendingStatus] {
		return dromosv1alpha1.WorkloadRunnerPhaseSuspending
	}
	if statusesSet[kartav1alpha1.SuspendedStatus] {
		return dromosv1alpha1.WorkloadRunnerPhaseSuspended
	}
	if statusesSet[kartav1alpha1.RunningStatus] {
		return dromosv1alpha1.WorkloadRunnerPhaseRunning
	}
	if statusesSet[kartav1alpha1.InitializingStatus] {
		return dromosv1alpha1.WorkloadRunnerPhaseInitializing
	}

	return dromosv1alpha1.WorkloadRunnerPhaseUnknown
}
