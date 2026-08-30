package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/step"
)

func installChart(ctx context.Context, inst step.HelmInstaller, repoURL, version, releaseName, chartName, namespace string, isOCI bool, values []string, auth step.ChartAuth) error {
	if authInst, ok := inst.(step.AuthenticatedHelmInstaller); ok {
		return authInst.InstallOrUpgradeReleaseWithAuth(ctx, repoURL, version, releaseName, chartName, namespace, isOCI, values, auth)
	}
	return inst.InstallOrUpgradeRelease(ctx, repoURL, version, releaseName, chartName, namespace, isOCI, values)
}

// mergeValueStrings overlays user overrides onto base --set values. When an
// override shares a key with a base value it replaces it in place (so only one
// entry per key is emitted); otherwise it is appended in order. Helm honors the
// last --set, but emitting both is confusing.
func mergeValueStrings(base, overrides []string) []string {
	out := make([]string, len(base))
	copy(out, base)
	idx := make(map[string]int, len(base))
	for i, b := range base {
		if k := valueKey(b); k != "" {
			idx[k] = i
		}
	}
	for _, o := range overrides {
		k := valueKey(o)
		if i, ok := idx[k]; ok && k != "" {
			out[i] = o
			continue
		}
		if k != "" {
			idx[k] = len(out)
		}
		out = append(out, o)
	}
	return out
}

func valueKey(s string) string {
	if i := strings.IndexByte(s, '='); i >= 0 {
		return s[:i]
	}
	return ""
}

// describeHelmCommand renders a helm command with each value override (--set)
// and value file (--values) on its own continued line, so dry-run output is
// readable in the terminal. When there are no values it returns base unchanged.
func describeHelmCommand(base string, values, valueFiles []string) string {
	lines := make([]string, 0, 1+len(values)+len(valueFiles))
	lines = append(lines, base)
	for _, v := range values {
		lines = append(lines, "--set "+v)
	}
	for _, f := range valueFiles {
		lines = append(lines, "--values "+f)
	}
	return strings.Join(lines, " \\\n  ")
}

type HelmInstall struct {
	RepoURL     string
	Version     string
	ReleaseName string
	ChartName   string
	Namespace   string
	IsOCI       bool
	Values      []string
	Auth        step.ChartAuth
	// KubeVersion, when set (e.g. "1.28.15"), overrides the Kubernetes version
	// the chart is rendered against. Reserved for charts whose capability checks
	// reject the live cluster version; no chart currently sets it.
	KubeVersion string
}

// kubeVersionInstaller is the optional capability a HelmInstaller exposes to
// render a chart against a pinned Kubernetes version. HelmInstall type-asserts
// it only when KubeVersion is set, so the shared step.HelmInstaller interface
// stays unchanged.
type kubeVersionInstaller interface {
	InstallOrUpgradeReleaseWithKubeVersion(ctx context.Context, repoURL, version, releaseName, chartName, namespace string, isOCI bool, values []string, kubeVersion string) error
}

func (a HelmInstall) Kind() string { return "helm.install" }

func (a HelmInstall) Release() (name, namespace string) { return a.ReleaseName, a.Namespace }

func (a HelmInstall) Args() map[string]any {
	return map[string]any{
		"repo":        a.RepoURL,
		"version":     a.Version,
		"release":     a.ReleaseName,
		"chart":       a.ChartName,
		"namespace":   a.Namespace,
		"oci":         a.IsOCI,
		"values":      a.Values,
		"kubeVersion": a.KubeVersion,
	}
}

func (a HelmInstall) Describe() string {
	base := fmt.Sprintf("helm upgrade --install %s %s --version %s --namespace %s", a.ReleaseName, a.ChartName, a.Version, a.Namespace)
	return describeHelmCommand(base, a.Values, nil)
}

func (a HelmInstall) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	if a.KubeVersion != "" {
		kv, ok := deps.HelmInstaller.(kubeVersionInstaller)
		if !ok {
			return orchestrator2.ActionResult{}, fmt.Errorf("helm installer does not support a kube version override (release %q)", a.ReleaseName)
		}
		return orchestrator2.ActionResult{}, kv.InstallOrUpgradeReleaseWithKubeVersion(ctx, a.RepoURL, a.Version, a.ReleaseName, a.ChartName, a.Namespace, a.IsOCI, a.Values, a.KubeVersion)
	}
	return orchestrator2.ActionResult{}, installChart(ctx, deps.HelmInstaller, a.RepoURL, a.Version, a.ReleaseName, a.ChartName, a.Namespace, a.IsOCI, a.Values, a.Auth)
}
