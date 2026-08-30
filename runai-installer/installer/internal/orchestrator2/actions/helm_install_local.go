package actions

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/artifacts"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/utils/airgap"
)

type HelmInstallLocal struct {
	ChartPath   string
	WorkingDir  string
	ChartPrefix string
	LockChart   string
	ReleaseName string
	Namespace   string
	Values      []string
	ValueFiles  []string
}

func (a HelmInstallLocal) chartPath() string {
	if a.LockChart != "" {
		return artifacts.LockedChartPreviewPath(a.WorkingDir, a.LockChart)
	}
	if a.ChartPrefix != "" {
		return airgap.FindChartFile(a.WorkingDir, a.ChartPrefix)
	}
	return a.ChartPath
}

func (a HelmInstallLocal) resolvedChartPath() (string, error) {
	if a.LockChart != "" {
		return artifacts.LockedChartPath(a.WorkingDir, a.LockChart)
	}
	return a.chartPath(), nil
}

func (a HelmInstallLocal) Kind() string { return "helm.install_local" }

func (a HelmInstallLocal) Release() (name, namespace string) { return a.ReleaseName, a.Namespace }

func (a HelmInstallLocal) Args() map[string]any {
	return map[string]any{
		"chartPath":  a.chartPath(),
		"release":    a.ReleaseName,
		"namespace":  a.Namespace,
		"values":     a.Values,
		"valueFiles": a.ValueFiles,
	}
}

func (a HelmInstallLocal) Describe() string {
	base := fmt.Sprintf("helm upgrade --install %s %s --namespace %s", a.ReleaseName, a.chartPath(), a.Namespace)
	return describeHelmCommand(base, a.Values, a.ValueFiles)
}

func (a HelmInstallLocal) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	chartPath, err := a.resolvedChartPath()
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	return orchestrator2.ActionResult{}, deps.HelmInstaller.InstallOrUpgradeFromLocalChart(ctx, chartPath, a.ReleaseName, a.Namespace, a.Values, a.ValueFiles)
}
