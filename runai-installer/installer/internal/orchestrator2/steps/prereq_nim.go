package steps

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
)

const (
	nimOperatorRelease   = "nim-operator"
	nimOperatorChart     = "k8s-nim-operator"
	nimOperatorNamespace = "nim-operator"
	// nimOperatorSentinel is the operator's controller-manager Deployment. The
	// chart names it with the Helm fullname "<release>-<chart>", i.e.
	// "nim-operator-k8s-nim-operator" (not a "-controller-manager" suffix).
	nimOperatorSentinel = "nim-operator-k8s-nim-operator"
)

// NewNIMOperator builds the NVIDIA NIM Operator install step (NGC). It manages
// the lifecycle of NVIDIA NIM and NeMo inference microservices.
func NewNIMOperator(version string, customValues map[string]string, deps orchestrator2.Deps) orchestrator2.Step {
	action := actions.HelmInstall{
		RepoURL:     ngcHelmRepo,
		Version:     version,
		ReleaseName: nimOperatorRelease,
		ChartName:   nimOperatorChart,
		Namespace:   nimOperatorNamespace,
		Values:      actions.MapToValues(customValues),
	}
	return newPrereqStep("install-nim-operator", deploymentSentinel(nimOperatorSentinel, nimOperatorNamespace), deps, action)
}
