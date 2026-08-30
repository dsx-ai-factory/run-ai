package install_v2_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/installconfig"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// skipPrometheus disables the Prometheus prerequisite on a config built from
// connectedConfig (which enables it). Installing Prometheus is rejected on
// OpenShift — Run:ai uses the built-in openshift-monitoring stack — so
// OpenShift tests must opt out of it.
func skipPrometheus(cfg *installconfig.Config) {
	cfg.Prerequisites.Prometheus = &installconfig.Dependency{Mode: "skip"}
}

func TestConnectedInstall_OpenShift_NoCA(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	seedOpenShiftPrereqs(k8s) // OpenShift installs the GPU + network operators via OperatorHub (OLM)
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Install.KubernetesDistribution = "openshift"
	cfg.ExternalConnection = nil // OpenShift uses Routes; externalConnection must be unset
	skipPrometheus(cfg)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	cpCalls := helm.CallsByRelease("runai-backend")
	require.Len(t, cpCalls, 1)
	assert.Contains(t, cpCalls[0].Values, "global.config.kubernetesDistribution=openshift")
	assert.NotContains(t, cpCalls[0].Values, "global.customCA.enabled=true", "no custom CA without a CA chain")

	clCalls := helm.CallsByRelease("runai-cluster")
	require.Len(t, clCalls, 1)
	assert.NotContains(t, clCalls[0].Values, "global.config.kubernetesDistribution=openshift", "distribution flag is control-plane only")

	applied := appliedNames(k8s)
	assert.False(t, applied["runai-ca-cert"], "no CA secret is created without a CA chain")

	assert.True(t, apiClient.ClusterCreated, "cluster registered via control-plane API")
	assert.True(t, hasPipelineEvent(readEvents(t, eventsPath), "pipeline_succeeded", "runai"))
}

func TestAirgapInstall_OpenShift(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := airgapConfig(t, harness.CreateTestTarball(t), eventsPath)
	cfg.Install.KubernetesDistribution = "openshift"
	cfg.ExternalConnection = nil // OpenShift uses Routes; externalConnection must be unset
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	cpCalls := helm.CallsByRelease("runai-backend")
	require.Len(t, cpCalls, 1)
	assert.Contains(t, cpCalls[0].Values, "global.config.kubernetesDistribution=openshift")

	clCalls := helm.CallsByRelease("runai-cluster")
	require.Len(t, clCalls, 1)
	assert.NotContains(t, clCalls[0].Values, "global.config.kubernetesDistribution=openshift", "distribution flag is control-plane only")

	assert.True(t, apiClient.ClusterCreated, "cluster registered via control-plane API")
	assert.True(t, hasPipelineEvent(readEvents(t, eventsPath), "pipeline_succeeded", "runai"))
}

func TestConnectedInstall_OpenShift_GatewayAPI(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}
	seedOpenShiftPrereqs(k8s) // GPU + network operators install via OperatorHub on OpenShift

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Install.KubernetesDistribution = "openshift"
	cfg.ExternalConnection = &installconfig.ExternalConnection{
		Type:       "gatewayApi",
		GatewayAPI: &installconfig.GatewayAPIConnection{Name: "runai-gw", Namespace: "gw-system"},
	}
	skipPrometheus(cfg)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	cpCalls := helm.CallsByRelease("runai-backend")
	require.Len(t, cpCalls, 1)
	assert.Contains(t, cpCalls[0].Values, "global.config.kubernetesDistribution=openshift")
	assert.Contains(t, cpCalls[0].Values, "global.gatewayAPI.enabled=true")
	assert.Contains(t, cpCalls[0].Values, "global.gatewayAPI.parentRefs[0].name=runai-gw")
	assert.Contains(t, cpCalls[0].Values, "global.gatewayAPI.parentRefs[0].namespace=gw-system")

	clCalls := helm.CallsByRelease("runai-cluster")
	require.Len(t, clCalls, 1)
	assert.Contains(t, clCalls[0].Values, "clusterConfig.global.gatewayAPI.enabled=true")
	assert.Contains(t, clCalls[0].Values, "clusterConfig.global.gatewayAPI.name=runai-gw")
	assert.Contains(t, clCalls[0].Values, "clusterConfig.global.gatewayAPI.namespace=gw-system")

	assert.True(t, apiClient.ClusterCreated, "cluster registered via control-plane API")
	assert.True(t, hasPipelineEvent(readEvents(t, eventsPath), "pipeline_succeeded", "runai"))
}

func TestConnectedInstall_OpenShift_FakeGPUSetsOpenShiftEnv(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	labeler := &harness.FakeNodeLabeler{Labeled: 2}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), NodeLabeler: labeler}
	seedOpenShiftNetworkOperator(k8s) // network operator still installs via OperatorHub even with a fake GPU

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Install.KubernetesDistribution = "openshift"
	cfg.ExternalConnection = nil // OpenShift uses Routes; externalConnection must be unset
	cfg.Prerequisites.GPUOperator.Type = "fake"
	skipPrometheus(cfg)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	gpu := helm.CallsByRelease("gpu-operator")
	require.Len(t, gpu, 1)
	assert.Contains(t, gpu[0].Values, "environment.openshift=true", "fake GPU operator binds the privileged SCC on OpenShift")
	assert.Nil(t, applyCallByKindName(k8s, "Subscription", "gpu-operator-certified"), "the fake GPU operator takes the Helm route, never OperatorHub/OLM")
}

func TestConnectedInstall_FakeGPU_NoOpenShiftEnvWhenVanilla(t *testing.T) {
	for _, dist := range []string{"", "vanilla"} {
		t.Run("dist="+dist, func(t *testing.T) {
			helm := harness.NewFakeHelmWithReleases()
			k8s := &harness.FakeK8sExecutor{}
			labeler := &harness.FakeNodeLabeler{Labeled: 2}
			_, provider := harness.NewFakeRunAIClientProvider(helm)
			deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), NodeLabeler: labeler}

			eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
			cfg, opts := connectedConfig(t, eventsPath, false)
			cfg.Install.KubernetesDistribution = dist
			cfg.Prerequisites.GPUOperator.Type = "fake"
			term, _, _ := newTestTerminal()

			require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

			gpu := helm.CallsByRelease("gpu-operator")
			require.Len(t, gpu, 1)
			assert.NotContains(t, gpu[0].Values, "environment.openshift=true", "vanilla/empty must not enable the OpenShift env on the fake GPU operator")
		})
	}
}

func TestConnectedInstall_VanillaDistribution_NoFlag(t *testing.T) {
	for _, dist := range []string{"", "vanilla"} {
		t.Run("dist="+dist, func(t *testing.T) {
			helm := harness.NewFakeHelmWithReleases()
			k8s := &harness.FakeK8sExecutor{}
			_, provider := harness.NewFakeRunAIClientProvider(helm)
			deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

			eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
			cfg, opts := connectedConfig(t, eventsPath, false)
			cfg.Install.KubernetesDistribution = dist
			term, _, _ := newTestTerminal()

			require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

			cpCalls := helm.CallsByRelease("runai-backend")
			require.Len(t, cpCalls, 1)
			for _, v := range cpCalls[0].Values {
				assert.NotContains(t, v, "kubernetesDistribution", "vanilla/empty must not emit a distribution flag")
			}
		})
	}
}
