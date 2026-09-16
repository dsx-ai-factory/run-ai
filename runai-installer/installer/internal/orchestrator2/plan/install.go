/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package plan

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/airgapvalues"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/config"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/steps"
	preflightpkg "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/preflight"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/airgap"
)

const (
	cpRegistrySecret = "runai-reg-creds"
	cpTLSSecret      = "runai-backend-tls"
	cpCASecret       = "runai-ca-cert"
	clusterTLSSecret = "runai-cluster-domain-tls-secret"
	clusterCASecret  = "runai-ca-cert"
)

var clusterCASecretLabels = map[string]string{
	"run.ai/cluster-wide": "true",
	"run.ai/name":         "runai-ca-cert",
}

// buildInstall returns the preflight plan and the install plan separately. The
// preflight pipeline is always read-only, so it executes for real even under
// --dry-run (DryRun stays false); only the install pipelines honor s.DryRun.
func buildInstall(s *spec.Spec, deps orchestrator2.Deps) (preflight *orchestrator2.Plan, install *orchestrator2.Plan) {
	preflightPipeline := &orchestrator2.Pipeline{Name: "preflight", Title: "Preflight checks", DryRun: false}
	if flow := preflightpkg.BuildFlow(s, deps); flow != nil {
		preflightPipeline.Flows = append(preflightPipeline.Flows, flow)
	}
	if s.UsesAirgapLoad() && !s.Airgap.SkipPrepare {
		if len(preflightPipeline.Flows) == 0 {
			preflightPipeline.Flows = append(preflightPipeline.Flows, &orchestrator2.Flow{
				Name:  "preflight-checks",
				Title: "Running checks",
			})
		}
		preflightPipeline.Flows[0].Steps = append(preflightPipeline.Flows[0].Steps, airgapLoadPreflightStep(s.AirgapLoad))
	}

	// Prerequisites (the infra Run:ai depends on) install before Run:ai;
	// features (capabilities layered on top, some of which depend on the Run:ai
	// cluster's scheduler/CRDs) install after Run:ai.
	prereqFlows, prereqComps := buildPrereqFlows(s, deps)
	featureFlows, featureComps := buildFeatureFlows(s, deps)

	prereq := &orchestrator2.Pipeline{Name: "prereq", Title: "Prerequisites", DryRun: s.DryRun, Flows: prereqFlows}
	runai := &orchestrator2.Pipeline{Name: "runai", Title: "NVIDIA Run:ai", DryRun: s.DryRun}

	if s.IsAirGapped() && !s.Airgap.SkipPrepare {
		if s.UsesAirgapLoad() {
			runai.Flows = append(runai.Flows, airgapLoadFlow(s.AirgapLoad))
		} else {
			runai.Flows = append(runai.Flows, buildAirgapPrepFlow(s))
		}
	}

	var cpFlow, clusterFlow *orchestrator2.Flow
	if s.ControlPlane != nil {
		cpFlow = buildControlPlaneFlow(s, deps)
	}
	if s.Cluster != nil {
		clusterFlow = buildClusterFlow(s, deps)
	}
	// The system-node labeling is one cluster-wide mutation shared by both
	// sides. It must run before any pod with the system affinity is scheduled,
	// so it is prepended to the control-plane flow when the control plane is
	// installed (which runs before the cluster flow), otherwise to the cluster
	// flow. This keeps full / CP-only / cluster-only installs correct.
	if ls := systemNodeLabelStep(s); ls != nil {
		switch {
		case cpFlow != nil:
			cpFlow.Steps = append([]orchestrator2.Step{ls}, cpFlow.Steps...)
		case clusterFlow != nil:
			clusterFlow.Steps = append([]orchestrator2.Step{ls}, clusterFlow.Steps...)
		}
	}
	if cpFlow != nil {
		runai.Flows = append(runai.Flows, cpFlow)
	}
	if clusterFlow != nil {
		runai.Flows = append(runai.Flows, clusterFlow)
	}

	pipelines := []*orchestrator2.Pipeline{prereq, runai}
	if len(featureFlows) > 0 {
		pipelines = append(pipelines, &orchestrator2.Pipeline{Name: "features", Title: "Optional features", DryRun: s.DryRun, Flows: featureFlows})
	}

	installPlan := &orchestrator2.Plan{Pipelines: pipelines}
	if fallback := compatFallback(s, append(prereqComps, featureComps...)); fallback != nil {
		installPlan.PreEvents = append(installPlan.PreEvents, orchestrator2.Event{
			Kind:   orchestrator2.EventKindCompatFallback,
			Compat: fallback,
		})
	}

	return &orchestrator2.Plan{Pipelines: []*orchestrator2.Pipeline{preflightPipeline}}, installPlan
}

// buildAirgapPrepFlow extracts the offline bundle and pushes its images to the
// mirror registry before the Run:ai charts install. Both steps locate the
// bundle through the deterministic Airgap.WorkingDir, so no inter-step state is
// needed.
func buildAirgapPrepFlow(s *spec.Spec) *orchestrator2.Flow {
	workingDir := s.Airgap.WorkingDir
	var registryURL, username, password string
	if s.Registry != nil {
		registryURL = s.Registry.Host
		username = s.Registry.Username
		password = s.Registry.Password
	}
	return &orchestrator2.Flow{
		Name:  "airgap-preparation",
		Title: "Air-gapped bundle preparation",
		Steps: []orchestrator2.Step{
			steps.NewActionStepTitled("extract-airgap-package", "Extracting air-gapped bundle", actions.ExtractAirgap{
				BundlePath: s.Airgap.BundlePath,
				WorkingDir: workingDir,
			}),
			steps.WithExtraTimeout(
				steps.NewActionStepTitled("push-airgap-images", "Pushing images to the mirror registry (this can take a while)", actions.PushAirgapImages{
					WorkingDir:  workingDir,
					RegistryURL: registryURL,
					Username:    username,
					Password:    password,
				}),
				// Pushing/loading the full image set is very long and dominated by
				// external factors (network throughput, disk speed), so it far
				// exceeds the per-step norm.
				120*time.Minute,
			),
		},
	}
}

func buildControlPlaneFlow(s *spec.Spec, deps orchestrator2.Deps) *orchestrator2.Flow {
	tls := s.ControlPlaneTLS()
	flowSteps := []orchestrator2.Step{
		steps.NewNamespace("create-control-plane-namespace", step.CPNamespace, deps),
	}

	if usesManagedRegistrySecret(s) {
		if dockerConfig, ok := s.Registry.ImagePullConfig(); ok {
			flowSteps = append(flowSteps, steps.NewRegistrySecret("install-registry-secret", cpRegistrySecret, step.CPNamespace, dockerConfig, deps))
		}
	}

	if hasServingCertificate(tls) {
		flowSteps = append(flowSteps, steps.NewTLSSecret("install-control-plane-tls", cpTLSSecret, step.CPNamespace, tls.CertPath, tls.KeyPath, deps))
	}

	caPath := caChainPath(tls)
	if caPath != "" {
		flowSteps = append(flowSteps, steps.NewCASecret("install-control-plane-ca", cpCASecret, step.CPNamespace, step.CPReleaseName, caPath, nil, deps))
	}

	flowSteps = append(flowSteps, newControlPlaneChartStep(s, caPath, deps))

	return &orchestrator2.Flow{Name: "control-plane", Title: "Control plane", Steps: flowSteps}
}

func newControlPlaneChartStep(s *spec.Spec, caPath string, deps orchestrator2.Deps) orchestrator2.Step {
	cp := s.ControlPlane
	const title = "Installing control plane"
	if s.IsAirGapped() {
		workingDir := s.Airgap.WorkingDir
		if s.UsesAirgapLoad() {
			return steps.Titled(steps.NewControlPlaneChart("install-control-plane-chart", actions.HelmInstallLocal{
				WorkingDir:  workingDir,
				LockChart:   step.CPChartName,
				ReleaseName: step.CPReleaseName,
				Namespace:   step.CPNamespace,
				Values:      controlPlaneValuesAirgap(s, caPath),
				ValueFiles:  []string{filepath.Join(s.Airgap.ValuesDir, airgapvalues.ControlPlaneValuesFile)},
			}, caPath, deps), title)
		}
		return steps.Titled(steps.NewControlPlaneChart("install-control-plane-chart", actions.HelmInstallLocal{
			WorkingDir:  workingDir,
			ChartPrefix: step.CPChartName,
			ReleaseName: step.CPReleaseName,
			Namespace:   step.CPNamespace,
			Values:      controlPlaneValuesAirgap(s, caPath),
			ValueFiles:  []string{filepath.Join(workingDir, airgap.CustomEnvFile)},
		}, caPath, deps), title)
	}
	repo, isOCI := cpChartSource(s)
	return steps.Titled(steps.NewControlPlaneChart("install-control-plane-chart", actions.HelmInstall{
		RepoURL:     repo,
		IsOCI:       isOCI,
		Auth:        chartAuth(s.Registry),
		Version:     cp.Version,
		ReleaseName: step.CPReleaseName,
		ChartName:   step.CPChartName,
		Namespace:   step.CPNamespace,
		Values:      controlPlaneValues(s, caPath),
	}, caPath, deps), title)
}

func buildClusterFlow(s *spec.Spec, deps orchestrator2.Deps) *orchestrator2.Flow {
	c := s.Cluster
	tls := s.ClusterTLS()
	domain := c.Domain
	if domain == "" && s.ControlPlane != nil {
		domain = s.ControlPlane.Domain
	}
	version := c.Version
	if version == "" && s.ControlPlane != nil {
		version = s.ControlPlane.Version
	}
	caPath := caChainPath(tls)

	flowSteps := []orchestrator2.Step{
		steps.NewNamespace("create-cluster-namespace", step.ClusterNamespace, deps),
	}
	if usesManagedRegistrySecret(s) {
		if dockerConfig, ok := s.Registry.ImagePullConfig(); ok {
			flowSteps = append(flowSteps, steps.NewRegistrySecret("install-cluster-registry-secret", cpRegistrySecret, step.ClusterNamespace, dockerConfig, deps))
		}
	}

	if hasServingCertificate(tls) {
		flowSteps = append(flowSteps, steps.NewTLSSecret("install-cluster-tls", clusterTLSSecret, step.ClusterNamespace, tls.CertPath, tls.KeyPath, deps))
	}

	if caPath != "" {
		flowSteps = append(flowSteps, steps.NewCASecret("install-cluster-ca", clusterCASecret, step.ClusterNamespace, step.ClusterReleaseName, caPath, clusterCASecretLabels, deps))
	}

	if !c.UsesSuppliedClusterIdentity() {
		flowSteps = append(flowSteps, steps.NewEnsureCluster("register-cluster", actions.EnsureCluster{
			ClusterName:  c.Name,
			ControlPlane: controlPlaneRef(c, domain, caPath),
		}, deps))
	}

	flowSteps = append(flowSteps, newClusterChartStep(s, domain, version, caPath, deps))

	return &orchestrator2.Flow{Name: "cluster", Title: "Run:ai cluster", Steps: flowSteps}
}

func newClusterChartStep(s *spec.Spec, domain, version, caPath string, deps orchestrator2.Deps) orchestrator2.Step {
	c := s.Cluster
	const title = "Installing Run:ai cluster"
	if s.IsAirGapped() {
		workingDir := s.Airgap.WorkingDir
		if s.UsesAirgapLoad() {
			return steps.Titled(steps.NewClusterChart("install-cluster-chart", actions.InstallClusterChartLocal{
				ClusterName:     c.Name,
				Domain:          domain,
				ControlPlaneURL: c.ControlPlaneURL,
				ClusterUID:      c.UID,
				BootstrapToken:  c.BootstrapToken,
				ClientSecret:    c.ClientSecret,
				Credentials:     controlPlaneCredentials(c),
				CACertPath:      caPath,
				WorkingDir:      workingDir,
				LockChart:       step.ClusterChartName,
				Version:         version,
				ReleaseName:     step.ClusterReleaseName,
				Namespace:       step.ClusterNamespace,
				HostRouting:     c.HostRouting(),
				ValueFiles:      []string{filepath.Join(s.Airgap.ValuesDir, airgapvalues.ClusterValuesFile)},
				ExtraValues:     clusterExtraValues(s, c),
			}, deps), title)
		}
		return steps.Titled(steps.NewClusterChart("install-cluster-chart", actions.InstallClusterChartLocal{
			ClusterName:     c.Name,
			Domain:          domain,
			ControlPlaneURL: c.ControlPlaneURL,
			ClusterUID:      c.UID,
			BootstrapToken:  c.BootstrapToken,
			ClientSecret:    c.ClientSecret,
			Credentials:     controlPlaneCredentials(c),
			CACertPath:      caPath,
			WorkingDir:      workingDir,
			ChartPrefix:     step.ClusterChartName,
			Version:         version,
			ReleaseName:     step.ClusterReleaseName,
			Namespace:       step.ClusterNamespace,
			HostRouting:     c.HostRouting(),
			ValueFiles:      []string{filepath.Join(workingDir, airgap.CustomEnvFile)},
			ExtraValues:     clusterExtraValues(s, c),
		}, deps), title)
	}
	repo, isOCI := clusterChartSource(s)
	return steps.Titled(steps.NewClusterChart("install-cluster-chart", actions.InstallClusterChart{
		ClusterName:     c.Name,
		Domain:          domain,
		ControlPlaneURL: c.ControlPlaneURL,
		ClusterUID:      c.UID,
		BootstrapToken:  c.BootstrapToken,
		ClientSecret:    c.ClientSecret,
		Credentials:     controlPlaneCredentials(c),
		CACertPath:      caPath,
		Repo:            repo,
		IsOCI:           isOCI,
		Auth:            chartAuth(s.Registry),
		Version:         version,
		ReleaseName:     step.ClusterReleaseName,
		ChartName:       step.ClusterChartName,
		Namespace:       step.ClusterNamespace,
		HostRouting:     c.HostRouting(),
		ExtraValues:     clusterExtraValues(s, c),
	}, deps), title)
}

func controlPlaneRef(c *spec.Cluster, domain, caPath string) actions.ControlPlaneRef {
	return actions.ControlPlaneRef{
		URL:         c.ControlPlaneURL,
		Domain:      domain,
		CACertPath:  caPath,
		Credentials: controlPlaneCredentials(c),
	}
}

func controlPlaneCredentials(c *spec.Cluster) *step.ControlPlaneCredentials {
	if !c.UsesControlPlaneCredentials() {
		return nil
	}
	return &step.ControlPlaneCredentials{
		Username: c.ControlPlaneCredentials.Username,
		Password: c.ControlPlaneCredentials.Password,
		Token:    c.ControlPlaneCredentials.Token,
	}
}

func controlPlaneValues(s *spec.Spec, caPath string) []string {
	values := []string{"global.domain=" + s.ControlPlane.Domain}
	if usesManagedRegistrySecret(s) {
		values = append(values, "global.imagePullSecrets[0].name="+cpRegistrySecret)
	}
	if caPath != "" {
		values = append(values, "global.customCA.enabled=true")
	}
	if s.KubernetesDistribution != "" {
		values = append(values, "global.config.kubernetesDistribution="+s.KubernetesDistribution)
	}
	values = append(values, externalConnectionCPValues(s.ExternalConnection)...)
	values = append(values, systemAffinity(s, "global.affinity")...)
	return actions.MergeValues(values, s.ControlPlane.CustomHelmValues)
}

// systemAffinity renders the system-node affinity helm values under prefix, or
// nil when the feature is off.
func systemAffinity(s *spec.Spec, prefix string) []string {
	if s.SystemNodes == nil {
		return nil
	}
	return actions.SystemNodeAffinityValues(prefix, s.SystemNodes.Label, s.SystemNodes.Mode == spec.SystemNodeModeRequired)
}

// systemNodeLabelStep builds the step that labels the selected system nodes, or
// nil when the feature is off or the installer is not managing labels.
func systemNodeLabelStep(s *spec.Spec) orchestrator2.Step {
	if s.SystemNodes == nil || !s.SystemNodes.ManageLabels {
		return nil
	}
	return steps.NewActionStep("label-system-nodes", actions.LabelSystemNodes{
		Names:    s.SystemNodes.Nodes,
		Selector: s.SystemNodes.Selector,
		Label:    s.SystemNodes.Label,
	})
}

const defaultIngressClass = "nginx"

// knativeIngressClass is the class for the Knative fronting Ingress, taken from
// the external-connection config (ingress mode) and defaulting to nginx.
func knativeIngressClass(s *spec.Spec) string {
	if ec := s.ExternalConnection; ec != nil && ec.Type == spec.ExternalConnectionIngress && ec.Ingress != nil && ec.Ingress.ClassName != "" {
		return ec.Ingress.ClassName
	}
	return defaultIngressClass
}

// knativeGateway returns the existing Gateway the Knative front-door attaches to
// under Gateway API routing, or nil when routing is Ingress-based (the default).
func knativeGateway(s *spec.Spec) *spec.GatewayAPIConnection {
	if ec := s.ExternalConnection; ec != nil && ec.Type == spec.ExternalConnectionGatewayAPI {
		return ec.GatewayAPI
	}
	return nil
}

// externalConnectionCPValues renders the control-plane chart helm values for the
// external routing config (direct global.* keys). It is empty when the config is
// unset (OpenShift, where Routes are used automatically).
func externalConnectionCPValues(ec *spec.ExternalConnection) []string {
	if ec == nil {
		return nil
	}
	switch ec.Type {
	case spec.ExternalConnectionIngress:
		if ec.Ingress != nil && ec.Ingress.ClassName != "" {
			return []string{"global.ingress.ingressClass=" + ec.Ingress.ClassName}
		}
	case spec.ExternalConnectionGatewayAPI:
		if ec.GatewayAPI != nil {
			return []string{
				"global.gatewayAPI.enabled=true",
				"global.gatewayAPI.parentRefs[0].name=" + ec.GatewayAPI.Name,
				"global.gatewayAPI.parentRefs[0].namespace=" + ec.GatewayAPI.Namespace,
			}
		}
	}
	return nil
}

// externalConnectionClusterValues renders the cluster chart helm values. The
// cluster chart sets RunaiConfig fields via the clusterConfig merge-patch, so
// the keys differ from the control-plane chart (and Gateway API uses a single
// name/namespace rather than the CP's parentRefs list).
func externalConnectionClusterValues(ec *spec.ExternalConnection) []string {
	if ec == nil {
		return nil
	}
	switch ec.Type {
	case spec.ExternalConnectionIngress:
		if ec.Ingress != nil && ec.Ingress.ClassName != "" {
			return []string{"clusterConfig.global.ingress.ingressClass=" + ec.Ingress.ClassName}
		}
	case spec.ExternalConnectionGatewayAPI:
		if ec.GatewayAPI != nil {
			return []string{
				"clusterConfig.global.gatewayAPI.enabled=true",
				"clusterConfig.global.gatewayAPI.name=" + ec.GatewayAPI.Name,
				"clusterConfig.global.gatewayAPI.namespace=" + ec.GatewayAPI.Namespace,
			}
		}
	}
	return nil
}

// clusterExtraValues combines the external-connection cluster values with the
// user's custom helm values for the cluster chart.
func clusterExtraValues(s *spec.Spec, c *spec.Cluster) []string {
	values := externalConnectionClusterValues(s.ExternalConnection)
	if usesManagedRegistrySecret(s) {
		values = append(values, "global.additionalImagePullSecrets[0]="+cpRegistrySecret)
	}
	// The cluster needs the affinity in two places: global.affinity covers the
	// operator + hook pods, while runai-operator.config.global.affinity flows
	// into RunaiConfig.spec.global.affinity and covers the operand-managed
	// system Deployments (which have no built-in runai-system default).
	values = append(values, systemAffinity(s, "global.affinity")...)
	values = append(values, systemAffinity(s, "runai-operator.config.global.affinity")...)
	return append(values, actions.MapToValues(c.CustomHelmValues)...)
}

// controlPlaneValuesAirgap renders the local chart values for air-gapped installs.
func controlPlaneValuesAirgap(s *spec.Spec, caPath string) []string {
	values := []string{"global.domain=" + s.ControlPlane.Domain}
	if usesManagedRegistrySecret(s) {
		values = append(values, "global.imagePullSecrets[0].name="+cpRegistrySecret)
	}
	if caPath != "" {
		values = append(values, "global.customCA.enabled=true")
	}
	if s.KubernetesDistribution != "" {
		values = append(values, "global.config.kubernetesDistribution="+s.KubernetesDistribution)
	}
	values = append(values, externalConnectionCPValues(s.ExternalConnection)...)
	values = append(values, systemAffinity(s, "global.affinity")...)
	return actions.MergeValues(values, s.ControlPlane.CustomHelmValues)
}

func hasImagePullConfig(registry *spec.Registry) bool {
	_, ok := registry.ImagePullConfig()
	return ok
}

func usesManagedRegistrySecret(s *spec.Spec) bool {
	return (!s.IsAirGapped() || s.UsesAirgapLoad()) && hasImagePullConfig(s.Registry)
}

func caChainPath(tls *spec.TLS) string {
	if tls == nil {
		return ""
	}
	return tls.CAChainPath
}

func cpChartSource(s *spec.Spec) (repo string, isOCI bool) {
	if isNGC(s.Registry) {
		return ngcChartRepo(s.Registry, toConfigEnv(s.Env)), false
	}
	return step.CPChartRepo(toConfigEnv(s.Env)), false
}

func clusterChartSource(s *spec.Spec) (repo string, isOCI bool) {
	if isNGC(s.Registry) {
		return ngcChartRepo(s.Registry, toConfigEnv(s.Env)), false
	}
	return step.ClusterChartRepo(toConfigEnv(s.Env)), false
}

// ngcChartRepo honors the configured override.
func ngcChartRepo(r *spec.Registry, env config.RunaiEnvType) string {
	if r.ChartRepository != "" {
		return strings.TrimSuffix(r.ChartRepository, "/")
	}
	return step.NGCChartRepo(env)
}

func isNGC(r *spec.Registry) bool {
	return r != nil && r.Type == spec.RegistryTypeNGC
}

func chartAuth(r *spec.Registry) step.ChartAuth {
	username, password := r.ChartCredentials()
	return step.ChartAuth{Username: username, Password: password}
}

func hasServingCertificate(tls *spec.TLS) bool {
	return tls != nil && tls.CertPath != "" && tls.KeyPath != ""
}

func toConfigEnv(e spec.Env) config.RunaiEnvType {
	switch e {
	case spec.EnvProd:
		return config.EnvProd
	case spec.EnvStaging:
		return config.EnvStaging
	case spec.EnvLab:
		return config.EnvLab
	default:
		return config.EnvTest
	}
}
