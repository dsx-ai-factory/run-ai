package steps

import (
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
)

const (
	lwsRelease   = "lws"
	lwsChart     = "lws"
	lwsNamespace = "lws-system"
	lwsRepo      = "oci://registry.k8s.io/lws/charts/lws"
	lwsSentinel  = "lws-controller-manager"
)

// NewLWS builds the LeaderWorkerSet (distributed inference) install step.
func NewLWS(version string, customValues map[string]string, deps orchestrator2.Deps) orchestrator2.Step {
	action := actions.HelmInstall{
		RepoURL:     lwsRepo,
		Version:     version,
		ReleaseName: lwsRelease,
		ChartName:   lwsChart,
		Namespace:   lwsNamespace,
		IsOCI:       true,
		Values:      actions.MapToValues(customValues),
	}
	return newPrereqStep("install-lws", deploymentSentinel(lwsSentinel, lwsNamespace), deps, action)
}
