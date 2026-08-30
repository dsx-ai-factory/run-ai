package actions

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
)

// openShiftMonitoringNamespace is the namespace the OpenShift cluster-monitoring
// stack runs in. Its presence is the signal that monitoring is installed.
const openShiftMonitoringNamespace = "openshift-monitoring"

// CheckOpenShiftMonitoring is a read-only preflight check that confirms the
// OpenShift monitoring stack is present, by verifying its namespace exists.
// Run:ai relies on OpenShift's built-in monitoring on OpenShift (the Prometheus
// prerequisite is rejected there), so a missing stack is a failure, not a
// warning. It only runs when the target distribution is OpenShift.
type CheckOpenShiftMonitoring struct{}

func (a CheckOpenShiftMonitoring) Kind() string { return "preflight.openshift_monitoring" }

func (a CheckOpenShiftMonitoring) Args() map[string]any { return map[string]any{} }

func (a CheckOpenShiftMonitoring) Describe() string {
	return "verify the OpenShift monitoring stack (openshift-monitoring) is installed"
}

func (a CheckOpenShiftMonitoring) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	exists, err := deps.ClusterProbe.NamespaceExists(ctx, openShiftMonitoringNamespace)
	if err != nil {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusWarn,
			Reason:      fmt.Sprintf("could not verify the OpenShift monitoring stack: %v", err),
			Remediation: "grant the installer permission to get namespaces so the openshift-monitoring stack can be validated",
		}, nil
	}
	if !exists {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusFail,
			Reason:      "the OpenShift monitoring stack was not found (namespace openshift-monitoring is missing)",
			Remediation: "enable the built-in OpenShift monitoring stack; Run:ai depends on it for metrics on OpenShift",
		}, nil
	}
	return orchestrator2.ActionResult{}, nil
}
