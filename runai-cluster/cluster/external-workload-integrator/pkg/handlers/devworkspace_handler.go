package handlers

import (
	"context"
	"fmt"
	devv1alpha2 "github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var devWorkspacePhaseToExternalWorkloadPhase = map[devv1alpha2.DevWorkspacePhase]v1alpha1.ExternalWorkloadPhase{
	devv1alpha2.DevWorkspaceStatusStarting: v1alpha1.ExWInitializingPhase,
	devv1alpha2.DevWorkspaceStatusRunning:  v1alpha1.ExWRunningPhase,
	devv1alpha2.DevWorkspaceStatusStopped:  v1alpha1.ExWStoppedPhase,
	devv1alpha2.DevWorkspaceStatusStopping: v1alpha1.ExWStoppingPhase,
	devv1alpha2.DevWorkspaceStatusFailed:   v1alpha1.ExWFailedPhase,
	devv1alpha2.DevWorkspaceStatusError:    v1alpha1.ExWDegradedPhase,
}

type DevWorkspaceTopOwnerHandler struct {
	cachedClient client.Client
}

func NewDevWorkspaceTopOwnerHandler(cachedClient client.Client) *DevWorkspaceTopOwnerHandler {
	return &DevWorkspaceTopOwnerHandler{
		cachedClient: cachedClient,
	}
}

func (h *DevWorkspaceTopOwnerHandler) GetObject(ctx context.Context, name, namespace string) (client.Object, error) {
	return utils.GetObject(ctx, h.cachedClient, name, namespace, func() client.Object { return &devv1alpha2.DevWorkspace{} })
}

func (h *DevWorkspaceTopOwnerHandler) Parse(ctx context.Context, object client.Object,
	externalWorkload *v1alpha1.ExternalWorkload) error {
	devWorkspace, ok := object.(*devv1alpha2.DevWorkspace)
	if !ok {
		return fmt.Errorf("failed converting object to devv1alpha2.DevWorkspace")
	}

	// Note: currently supporting only status, since the spec is a little more complicated to parse
	// (no PodSpec field).
	// could be handled in the future if needed.

	// PriorityClassName is not set in DevWorkspace spec, so we don't set it in ExternalWorkload.

	externalWorkload.Status.ObservedStatus.Phase = convertDevWorkspacePhase(devWorkspace.Status.Phase)
	externalWorkload.Status.ObservedStatus.K8sStatus = string(devWorkspace.Status.Phase)
	externalWorkload.Status.ObservedStatus.Message = devWorkspace.Status.Message

	observedConditions := []v1alpha1.ObservedCondition{}
	for _, condition := range devWorkspace.Status.Conditions {
		observedConditions = append(observedConditions, v1alpha1.ObservedCondition{
			Type:               string(condition.Type),
			Status:             condition.Status,
			Reason:             condition.Reason,
			Message:            condition.Message,
			LastTransitionTime: condition.LastTransitionTime,
		})
	}
	externalWorkload.Status.ObservedStatus.Conditions = observedConditions

	return nil
}

func convertDevWorkspacePhase(devWorkspacePhase devv1alpha2.DevWorkspacePhase) v1alpha1.ExternalWorkloadPhase {
	phase, found := devWorkspacePhaseToExternalWorkloadPhase[devWorkspacePhase]
	if found {
		return phase
	}
	return v1alpha1.ExWUnknownPhase
}
