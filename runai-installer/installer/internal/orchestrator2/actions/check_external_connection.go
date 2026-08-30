package actions

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
)

// CheckExternalConnection is a read-only preflight check that confirms the
// external routing the spec selects is present and ready on the cluster: the
// configured IngressClass exists (ingress mode), or the Gateway API is installed
// and the referenced Gateway is programmed (Gateway API mode). The installer
// does not install the ingress controller or the Gateway, so these are
// preconditions — a missing/unready front-door is a failure, not a warning.
//
// Exactly one mode is populated by the builder: IngressClassName for ingress, or
// GatewayName+GatewayNamespace for Gateway API.
type CheckExternalConnection struct {
	IngressClassName string
	GatewayName      string
	GatewayNamespace string
}

func (a CheckExternalConnection) Kind() string { return "preflight.external_connection" }

func (a CheckExternalConnection) Args() map[string]any { return map[string]any{} }

func (a CheckExternalConnection) Describe() string {
	return "verify the configured external connection (ingress class or Gateway API gateway) is installed and ready"
}

func (a CheckExternalConnection) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	if a.IngressClassName != "" {
		return a.checkIngress(ctx, deps)
	}
	if a.GatewayName != "" {
		return a.checkGateway(ctx, deps)
	}
	return orchestrator2.ActionResult{}, nil
}

func (a CheckExternalConnection) checkIngress(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	exists, err := deps.ClusterProbe.IngressClassExists(ctx, a.IngressClassName)
	if err != nil {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusWarn,
			Reason:      fmt.Sprintf("could not verify ingress class %q: %v", a.IngressClassName, err),
			Remediation: "grant the installer permission to get networking.k8s.io/ingressclasses so the ingress controller can be validated",
		}, nil
	}
	if !exists {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusFail,
			Reason:      fmt.Sprintf("ingress class %q was not found", a.IngressClassName),
			Remediation: fmt.Sprintf("install and configure an ingress controller that provides the %q IngressClass before installing", a.IngressClassName),
		}, nil
	}
	return orchestrator2.ActionResult{}, nil
}

func (a CheckExternalConnection) checkGateway(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	installed, err := deps.ClusterProbe.GatewayAPIInstalled(ctx)
	if err != nil {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusWarn,
			Reason:      fmt.Sprintf("could not verify the Gateway API is installed: %v", err),
			Remediation: "grant the installer discovery access so the Gateway API can be validated",
		}, nil
	}
	if !installed {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusFail,
			Reason:      "the Gateway API is not installed (gateway.networking.k8s.io/v1 is not served)",
			Remediation: "install the Gateway API CRDs and a Gateway controller before installing",
		}, nil
	}

	gw := fmt.Sprintf("%s/%s", a.GatewayNamespace, a.GatewayName)
	status, err := deps.ClusterProbe.GatewayStatus(ctx, a.GatewayNamespace, a.GatewayName)
	if err != nil {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusWarn,
			Reason:      fmt.Sprintf("could not verify gateway %s: %v", gw, err),
			Remediation: "grant the installer permission to get gateway.networking.k8s.io/gateways so the gateway can be validated",
		}, nil
	}
	if !status.Found {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusFail,
			Reason:      fmt.Sprintf("gateway %s was not found", gw),
			Remediation: "create the Gateway, or point externalConnection.gatewayApi at an existing Gateway",
		}, nil
	}
	if !status.Accepted || !status.Programmed {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusFail,
			Reason:      fmt.Sprintf("gateway %s is not ready (accepted=%t, programmed=%t)", gw, status.Accepted, status.Programmed),
			Remediation: "wait for the Gateway controller to accept and program the gateway (its listeners and address must be provisioned)",
		}, nil
	}
	return orchestrator2.ActionResult{}, nil
}
