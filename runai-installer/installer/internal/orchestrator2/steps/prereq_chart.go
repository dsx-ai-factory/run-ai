package steps

import (
	"context"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// prereqStep installs a third-party prerequisite via one or more actions.
// Unlike the Run:ai chart steps, Detect inspects a sentinel cluster resource (a
// known Deployment) rather than a Helm release, so a prerequisite a user
// installed by other means is also recognized as already present and skipped.
type prereqStep struct {
	*orchestrator2.BaseStep
	deps     orchestrator2.Deps
	actions  []orchestrator2.PerformAction
	sentinel client.Object
}

func newPrereqStep(name string, sentinel client.Object, deps orchestrator2.Deps, acts ...orchestrator2.PerformAction) *prereqStep {
	return &prereqStep{BaseStep: orchestrator2.NewBaseStep(name), deps: deps, actions: acts, sentinel: sentinel}
}

func (s *prereqStep) Detect(ctx context.Context, _ *spec.Spec) (orchestrator2.DetectResult, error) {
	err := s.deps.K8sExecutor.Get(ctx, client.ObjectKeyFromObject(s.sentinel), s.sentinel)
	if err == nil {
		return orchestrator2.DetectResult{Phase: orchestrator2.DetectInstalled}, nil
	}
	if apierrors.IsNotFound(err) {
		return orchestrator2.DetectResult{Phase: orchestrator2.DetectNotInstalled}, nil
	}
	return orchestrator2.DetectResult{}, err
}

func (s *prereqStep) BuildActions(*spec.Spec) ([]orchestrator2.PerformAction, error) {
	return s.actions, nil
}

// deploymentSentinel builds the read-only object Detect looks up to decide
// whether a prerequisite is already present.
func deploymentSentinel(name, namespace string) *appsv1.Deployment {
	return &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{Kind: "Deployment", APIVersion: "apps/v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}
}

const (
	prometheusRelease   = "kube-prometheus-stack"
	prometheusChart     = "kube-prometheus-stack"
	prometheusNamespace = "monitoring"
	prometheusRepo      = "oci://ghcr.io/prometheus-community/charts/kube-prometheus-stack"
	prometheusSentinel  = "kube-prometheus-stack-operator"
)

// NewPrometheusChart builds the Prometheus (kube-prometheus-stack) install step.
func NewPrometheusChart(version string, customValues map[string]string, deps orchestrator2.Deps) orchestrator2.Step {
	values := actions.MergeValues([]string{
		"grafana.enabled=false",
		"prometheus.prometheusSpec.scrapeConfigs[0].jobLabel=runai-cluster",
	}, customValues)
	action := actions.HelmInstall{
		RepoURL:     prometheusRepo,
		Version:     version,
		ReleaseName: prometheusRelease,
		ChartName:   prometheusChart,
		Namespace:   prometheusNamespace,
		IsOCI:       true,
		Values:      values,
	}
	return newPrereqStep("install-prometheus", deploymentSentinel(prometheusSentinel, prometheusNamespace), deps, action)
}
