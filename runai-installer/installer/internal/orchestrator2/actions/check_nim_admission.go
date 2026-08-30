package actions

import (
	"context"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
)

// CheckNIMAdmissionWebhook is a read-only preflight check that warns when the
// NIM Operator is installed with its admission controller left enabled. That
// webhook needs cert-manager for its TLS certificates; on a cluster without
// cert-manager the operator's webhook fails to come up. It is only assembled
// (by the preflight builder) when NIM is enabled and the admission controller
// is not explicitly disabled, so its Execute always reports the warning.
type CheckNIMAdmissionWebhook struct{}

func (a CheckNIMAdmissionWebhook) Kind() string { return "preflight.nim_admission_webhook" }

func (a CheckNIMAdmissionWebhook) Args() map[string]any { return map[string]any{} }

func (a CheckNIMAdmissionWebhook) Describe() string {
	return "verify the NIM Operator admission webhook can obtain TLS certificates (warning only)"
}

func (a CheckNIMAdmissionWebhook) Execute(_ context.Context, _ orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	return orchestrator2.ActionResult{
		Status:      orchestrator2.StatusWarn,
		Reason:      "NIM Operator admission controller is enabled; its webhook requires cert-manager for TLS certificates",
		Remediation: "install cert-manager, or disable the admission controller via features.nimOperator.customHelmValues with operator.admissionController.enabled=false",
	}, nil
}
