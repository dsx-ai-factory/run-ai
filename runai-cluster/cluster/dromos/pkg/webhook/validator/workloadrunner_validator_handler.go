package validator

import (
	"context"
	"encoding/json"
	controllervalidator "github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/utils/validator"
	"slices"

	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	admissionv1 "k8s.io/api/admission/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type WorkloadRunnerValidatorHandler struct {
	controllerValidator *controllervalidator.ControllerValidator
	framework           *framework.Framework
}

func NewWorkloadRunnerValidatorHandler(fw *framework.Framework) *WorkloadRunnerValidatorHandler {
	return &WorkloadRunnerValidatorHandler{
		controllerValidator: controllervalidator.NewControllerValidator(),
		framework:           fw,
	}
}

func (h *WorkloadRunnerValidatorHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	logger := log.FromContext(ctx)

	switch req.Operation {
	case admissionv1.Create, admissionv1.Update:
		return h.handleValidation(ctx, req)
	case admissionv1.Delete:
		return admission.Allowed("deletion allowed")
	default:
		logger.Info("Unsupported operation", "operation", req.Operation)
		return admission.Allowed("operation not validated")
	}
}

func (h *WorkloadRunnerValidatorHandler) handleValidation(
	ctx context.Context,
	req admission.Request,
) admission.Response {
	logger := log.FromContext(ctx)

	// expose the admission request (requester identity) to validator plugins
	ctx = admission.NewContextWithRequest(ctx, req)

	var desiredRunner dromosv1alpha1.WorkloadRunner
	if err := json.Unmarshal(req.Object.Raw, &desiredRunner); err != nil {
		logger.Error(err, "Failed to decode WorkloadRunner")
		return responseErr(apierrors.NewBadRequest(err.Error()))
	}

	logger.Info("Validating WorkloadRunner", "namespace", desiredRunner.Namespace, "name", desiredRunner.Name)

	var existingRunner *dromosv1alpha1.WorkloadRunner
	if req.Operation == admissionv1.Update {
		var existingObj dromosv1alpha1.WorkloadRunner
		if err := json.Unmarshal(req.OldObject.Raw, &existingObj); err != nil {
			logger.Error(err, "Failed to decode existing WorkloadRunner")
			return responseErr(apierrors.NewBadRequest(err.Error()))
		}
		existingRunner = &existingObj
	}

	allWarnings, errs := h.controllerValidator.Run(ctx, existingRunner, &desiredRunner)
	if len(errs) == 0 {
		warnings, pluginErrs := h.framework.RunValidators(ctx, existingRunner, &desiredRunner)
		allWarnings = append(allWarnings, warnings...)
		errs = pluginErrs
	}

	if len(errs) > 0 {
		return responseErr(statusErrorFor(desiredRunner.Name, errs)).WithWarnings(allWarnings...)
	}

	logger.V(2).Info( //nolint:mnd // log verbosity level
		"WorkloadRunner validation passed",
		"namespace", desiredRunner.Namespace,
		"name", desiredRunner.Name,
		"warnings", len(allWarnings),
	)

	return admission.Allowed("validation passed").WithWarnings(allWarnings...)
}

func statusErrorFor(name string, errs field.ErrorList) *apierrors.StatusError {
	isInternal := slices.ContainsFunc(errs, func(e *field.Error) bool { return e.Type == field.ErrorTypeInternal })
	if isInternal {
		return apierrors.NewInternalError(errs.ToAggregate())
	}
	gk := dromosv1alpha1.GroupVersion.WithKind(dromosv1alpha1.WorkloadRunnerKind).GroupKind()
	return apierrors.NewInvalid(gk, name, errs)
}

func responseErr(err *apierrors.StatusError) admission.Response {
	return admission.Response{
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed: false,
			Result:  &err.ErrStatus,
		},
	}
}
