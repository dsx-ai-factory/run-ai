package mutator

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/mutator"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	admissionv1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type WorkloadRunnerMutatorHandler struct {
	controllerMutator *mutator.ControllerMutator
	framework         *framework.Framework
}

func NewWorkloadRunnerMutatorHandler(fw *framework.Framework) *WorkloadRunnerMutatorHandler {
	return &WorkloadRunnerMutatorHandler{
		controllerMutator: mutator.NewControllerMutator(),
		framework:         fw,
	}
}

func (h *WorkloadRunnerMutatorHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	logger := log.FromContext(ctx)

	switch req.Operation {
	case admissionv1.Create, admissionv1.Update:
		return h.handleMutation(ctx, req)
	case admissionv1.Delete:
		return admission.Allowed("deletion allowed")
	default:
		logger.Info("Unsupported operation", "operation", req.Operation)
		return admission.Allowed("operation not mutated")
	}
}

func (h *WorkloadRunnerMutatorHandler) handleMutation(ctx context.Context, req admission.Request) admission.Response {
	logger := log.FromContext(ctx)

	var runner dromosv1alpha1.WorkloadRunner
	if err := json.Unmarshal(req.Object.Raw, &runner); err != nil {
		logger.Error(err, "Failed to decode WorkloadRunner")
		return admission.Errored(http.StatusBadRequest, err)
	}

	logger.Info("Mutating WorkloadRunner", "namespace", runner.Namespace, "name", runner.Name)

	allWarnings, err := h.controllerMutator.Run(ctx, &runner)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err).WithWarnings(allWarnings...)
	}
	allWarnings, err = h.framework.RunMutators(ctx, &runner)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err).WithWarnings(allWarnings...)
	}

	mutatedJSON, err := json.Marshal(runner)
	if err != nil {
		logger.Error(err, "Failed to marshal mutated WorkloadRunner")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	logger.Info(
		"WorkloadRunner mutation completed",
		"namespace", runner.Namespace,
		"name", runner.Name,
		"warnings", len(allWarnings),
	)

	return admission.PatchResponseFromRaw(req.Object.Raw, mutatedJSON).WithWarnings(allWarnings...)
}
