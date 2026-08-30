package install_v2_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/installconfig"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/run-ai/runai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// populatedAirgapWorkingDir mirrors an already-extracted bundle (charts/ +
// custom-env.yaml) so the SkipPrepare path can install local charts without
// running extraction or the image push.
func populatedAirgapWorkingDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "charts"), 0755))
	files := map[string]string{
		filepath.Join("charts", "control-plane-2.20.0.tgz"): "fake-cp-chart",
		filepath.Join("charts", "runai-cluster-2.20.0.tgz"): "fake-cluster-chart",
		"custom-env.yaml": "registry: test\n",
	}
	for name, content := range files {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(content), 0644))
	}
	return dir
}

// airgapConfig builds a full air-gapped install (control plane + cluster) from
// the given bundle and the matching prod-env run options writing events to path.
func airgapConfig(t *testing.T, bundlePath, eventsPath string) (*installconfig.Config, installconfig.RunOptions) {
	t.Helper()
	cfg := newInstallRunConfig()
	cfg.Install.Connectivity = string(spec.ConnectivityAirGapped)
	cfg.Airgap = &installconfig.Airgap{BundlePath: bundlePath, WorkingDir: t.TempDir()}
	cfg.ControlPlane.Version = "2.26.0"
	cfg.Cluster.Version = "2.26.0"

	opts := baseRunOptions()
	opts.Env = spec.EnvProd
	opts.EventsFile = eventsPath
	opts.Registry = &spec.Registry{Host: "registry.local:5000", Username: "admin", Password: "Harbor12345"}
	return cfg, opts
}

// TestInputValidation_AirgapForbidsDependencies asserts that enabling a
// prerequisite or feature on an air-gapped install is rejected by offline input
// validation before any work: the offline bundle carries only the Run:ai
// charts, so third-party prerequisites/features cannot be installed without
// registry access.
func TestInputValidation_AirgapForbidsDependencies(t *testing.T) {
	cases := []struct {
		name    string
		setup   func(cfg *installconfig.Config)
		wantSub string
	}{
		{
			name: "gpu operator prerequisite enabled",
			setup: func(cfg *installconfig.Config) {
				cfg.Prerequisites = &installconfig.Prerequisites{
					GPUOperator: &installconfig.Dependency{Mode: "install", Type: "fake"},
				}
			},
			wantSub: "does not support installing prerequisites in air-gapped mode",
		},
		{
			name: "prometheus prerequisite enabled",
			setup: func(cfg *installconfig.Config) {
				cfg.Prerequisites = &installconfig.Prerequisites{
					Prometheus: &installconfig.Dependency{Mode: "install"},
				}
			},
			wantSub: "does not support installing prerequisites in air-gapped mode",
		},
		{
			name: "prerequisite enabled by default (mode omitted)",
			setup: func(cfg *installconfig.Config) {
				cfg.Prerequisites = &installconfig.Prerequisites{
					Prometheus: &installconfig.Dependency{},
				}
			},
			wantSub: "does not support installing prerequisites in air-gapped mode",
		},
		{
			name: "feature enabled",
			setup: func(cfg *installconfig.Config) {
				cfg.Features = &installconfig.Features{
					LWS: &installconfig.Dependency{Mode: "install"},
				}
			},
			wantSub: "does not support installing features in air-gapped mode",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
			cfg, opts := airgapConfig(t, harness.CreateTestTarball(t), eventsPath)
			tc.setup(cfg)
			assertValidationFails(t, cfg, opts, tc.wantSub)
		})
	}
}

// TestInputValidation_AirgapDependencies_OneFindingPerGroup asserts that with
// many prerequisites and features enabled, the air-gap check reports a single
// finding per group (prerequisites, features) listing the enabled members —
// not one identical finding per member.
func TestInputValidation_AirgapDependencies_OneFindingPerGroup(t *testing.T) {
	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := airgapConfig(t, harness.CreateTestTarball(t), eventsPath)
	cfg.Prerequisites = &installconfig.Prerequisites{
		Prometheus:      &installconfig.Dependency{Mode: "install"},
		GPUOperator:     &installconfig.Dependency{Mode: "install", Type: "fake"},
		NetworkOperator: &installconfig.Dependency{Mode: "install"},
	}
	cfg.Features = &installconfig.Features{
		Knative:     &installconfig.Dependency{Mode: "install"},
		Kubeflow:    &installconfig.Dependency{Mode: "install"},
		MPIOperator: &installconfig.Dependency{Mode: "install"},
		LWS:         &installconfig.Dependency{Mode: "install"},
		Grove:       &installconfig.Dependency{Mode: "install"},
	}

	assertValidationFails(t, cfg, opts, "does not support installing prerequisites in air-gapped mode")

	evs := readEvents(t, eventsPath)
	ev := findValidationEvent(evs)
	require.NotNil(t, ev)
	findings, _ := ev["findings"].([]any)

	var airgapByField []string
	for _, f := range findings {
		m := f.(map[string]any)
		if strings.Contains(m["message"].(string), "air-gapped mode") {
			airgapByField = append(airgapByField, m["field"].(string))
		}
	}
	assert.ElementsMatch(t, []string{"prerequisites", "features"}, airgapByField,
		"exactly one air-gap finding per group, regardless of how many members are enabled")
}

// TestInputValidation_AirgapAllowsSkippedDependencies asserts that a group set
// to skip short-circuits the air-gap check even when a dependency inside it is
// marked install, so the air-gapped install proceeds to completion.
func TestInputValidation_AirgapAllowsSkippedDependencies(t *testing.T) {
	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := airgapConfig(t, harness.CreateTestTarball(t), eventsPath)
	cfg.Prerequisites = &installconfig.Prerequisites{
		Skip:        true,
		GPUOperator: &installconfig.Dependency{Mode: "install", Type: "fake"},
	}
	cfg.Features = &installconfig.Features{
		Skip: true,
		LWS:  &installconfig.Dependency{Mode: "install"},
	}
	assertInstallSucceeds(t, cfg, opts, eventsPath)
}

func TestAirgapInstall_FullFlow(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := airgapConfig(t, harness.CreateTestTarball(t), eventsPath)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	cpCalls := helm.CallsByRelease("runai-backend")
	require.Len(t, cpCalls, 1)
	// The chart path must resolve to the extracted charts/ dir, not the bare
	// working-dir fallback. The fallback would mean the path was resolved at
	// plan-build time (before extraction) instead of lazily at execute time.
	assert.Contains(t, cpCalls[0].ChartName, filepath.Join(cfg.Airgap.WorkingDir, "charts"), "control plane installs from the extracted charts/ dir")
	assert.True(t, strings.HasSuffix(cpCalls[0].ChartName, ".tgz"))
	assert.Contains(t, cpCalls[0].ValueFiles, filepath.Join(cfg.Airgap.WorkingDir, "custom-env.yaml"))
	assert.Subset(t, cpCalls[0].Values, []string{
		"global.domain=test.runai.local",
	})
	assert.NotContains(t, cpCalls[0].Values, "global.customCA.enabled=true")

	clCalls := helm.CallsByRelease("runai-cluster")
	require.Len(t, clCalls, 1)
	assert.Contains(t, clCalls[0].ChartName, filepath.Join(cfg.Airgap.WorkingDir, "charts"), "cluster installs from the extracted charts/ dir")
	assert.True(t, strings.HasSuffix(clCalls[0].ChartName, ".tgz"))
	assert.Contains(t, clCalls[0].ValueFiles, filepath.Join(cfg.Airgap.WorkingDir, "custom-env.yaml"))
	assert.Subset(t, clCalls[0].Values, []string{
		"controlPlane.url=test.runai.local",
		"controlPlane.clientSecret=s3cr3t-airgap",
		"cluster.uid=uuid-airgap-test",
		"clusterConfig.global.subdomainSupport=true",
	})
	assert.NotContains(t, clCalls[0].Values, "global.customCA.enabled=true")

	applied := appliedNames(k8s)
	assert.True(t, applied["runai-backend"], "CP namespace applied")
	assert.True(t, applied["runai"], "cluster namespace applied")
	assert.False(t, applied["runai-reg-creds"], "air-gapped install must not create the JFrog registry secret")
	assert.False(t, applied["runai-ca-cert"], "no CA secret without a CA chain")

	assert.True(t, apiClient.ClusterCreated, "cluster registered via control-plane API")
	require.NotEmpty(t, apiClient.ProviderCalls)
	for _, call := range apiClient.ProviderCalls {
		assert.Empty(t, call.CACertPath)
	}

	// The push step actually ran setup.sh with the configured mirror registry
	// URL and credentials (passed via the environment).
	pushed, err := os.ReadFile(filepath.Join(cfg.Airgap.WorkingDir, harness.PushRegistryMarker))
	require.NoError(t, err)
	assert.Equal(t, "registry.local:5000 admin Harbor12345\n", string(pushed), "setup.sh received the configured mirror registry URL and credentials")

	evs := readEvents(t, eventsPath)
	require.NotNil(t, findEvent(evs, "step_succeeded", "extract-airgap-package"))
	require.NotNil(t, findEvent(evs, "step_succeeded", "push-airgap-images"))
	assert.Nil(t, findEvent(evs, "step_started", "install-registry-secret"), "air-gapped install must not run the registry-secret step")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	assert.False(t, hasPipelineEvent(evs, "pipeline_failed", ""))
}

func TestAirgapInstall_FullFlowWithBootstrapToken(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	apiClient.BootstrapToken = "bootstrap-token"
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := airgapConfig(t, harness.CreateTestTarball(t), eventsPath)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	require.Len(t, helm.CallsByRelease("runai-backend"), 1)
	clusterCalls := helm.CallsByRelease("runai-cluster")
	require.Len(t, clusterCalls, 1)
	assert.Contains(t, clusterCalls[0].Values, "controlPlane.bootstrapToken=bootstrap-token")
	assert.NotContains(t, clusterCalls[0].Values, "controlPlane.clientSecret=s3cr3t-airgap")
	assert.True(t, apiClient.ClusterCreated)

	events := readEvents(t, eventsPath)
	assert.True(t, hasPipelineEvent(events, "pipeline_succeeded", "runai"))
	assert.False(t, hasPipelineEvent(events, "pipeline_failed", ""))
}

func TestAirgapInstall_CustomCA(t *testing.T) {
	caPath := filepath.Join(t.TempDir(), "ca.crt")
	require.NoError(t, os.WriteFile(caPath, []byte("test-ca"), 0o600))
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}
	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := airgapConfig(t, "", eventsPath)
	cfg.Airgap.SkipPrepare = true
	cfg.Airgap.WorkingDir = populatedAirgapWorkingDir(t)
	cfg.TLS = &installconfig.TLS{CAChainPath: caPath}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	cpCalls := helm.CallsByRelease("runai-backend")
	require.Len(t, cpCalls, 1)
	assert.Contains(t, cpCalls[0].Values, "global.customCA.enabled=true")
	clusterCalls := helm.CallsByRelease("runai-cluster")
	require.Len(t, clusterCalls, 1)
	assert.Contains(t, clusterCalls[0].Values, "global.customCA.enabled=true")
	requireCASecret(t, k8s, "runai-backend", caPath)
	requireCASecret(t, k8s, "runai", caPath)
	require.NotEmpty(t, apiClient.ProviderCalls)
	for _, call := range apiClient.ProviderCalls {
		assert.Equal(t, caPath, call.CACertPath)
	}
}

func TestAirgapInstall_BadTarball_Fails(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	// The bundle exists (so it passes the input-validation file-exists check) but
	// is not a valid gzip stream, so extraction fails at the orchestrator step.
	badPath := filepath.Join(t.TempDir(), "corrupt.tar.gz")
	require.NoError(t, os.WriteFile(badPath, []byte("not a gzip stream"), 0o600))
	cfg, opts := airgapConfig(t, badPath, eventsPath)
	term, _, _ := newTestTerminal()

	err := specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts())
	require.Error(t, err)
	assert.Empty(t, helm.Calls, "no charts must be installed when bundle extraction fails")

	evs := readEvents(t, eventsPath)
	require.NotNil(t, findEvent(evs, "step_failed", "extract-airgap-package"))
}

func TestAirgapInstall_PushImagesFails_Aborts(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := airgapConfig(t, harness.CreateTarballFailingSetup(t), eventsPath)
	term, _, _ := newTestTerminal()

	err := specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts())
	require.Error(t, err)
	assert.Empty(t, helm.Calls, "no charts must be installed when image push fails")

	evs := readEvents(t, eventsPath)
	require.NotNil(t, findEvent(evs, "step_succeeded", "extract-airgap-package"), "extraction succeeds before the push fails")
	require.NotNil(t, findEvent(evs, "step_failed", "push-airgap-images"))
	assert.False(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"), "runai pipeline must not succeed after the push fails")
}

// TestAirgapInstall_SkipPrepare_MispopulatedWorkingDir simulates a user setting
// skipPrepare while pointing at a working dir that was never populated. Nothing
// extracts, so the resolved local chart path does not exist on disk.
func TestAirgapInstall_SkipPrepare_MispopulatedWorkingDir(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	emptyDir := t.TempDir()

	cfg := newInstallRunConfig()
	cfg.Install.Connectivity = string(spec.ConnectivityAirGapped)
	cfg.Airgap = &installconfig.Airgap{WorkingDir: emptyDir, SkipPrepare: true}
	cfg.ControlPlane.Version = "2.26.0"
	cfg.Cluster.Version = "2.26.0"

	opts := baseRunOptions()
	opts.Env = spec.EnvProd
	opts.EventsFile = eventsPath
	opts.Registry = &spec.Registry{Host: "registry.local:5000"}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	cpCalls := helm.CallsByRelease("runai-backend")
	require.Len(t, cpCalls, 1)
	assert.Contains(t, cpCalls[0].ChartName, emptyDir, "chart path resolves under the (empty) working dir")
	_, err := os.Stat(cpCalls[0].ChartName)
	assert.True(t, os.IsNotExist(err), "the resolved local chart does not exist; a real helm install would fail here")
}

func TestAirgapInstall_SkipPrepare(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	workingDir := populatedAirgapWorkingDir(t)

	cfg := newInstallRunConfig()
	cfg.Install.Connectivity = string(spec.ConnectivityAirGapped)
	cfg.Airgap = &installconfig.Airgap{WorkingDir: workingDir, SkipPrepare: true}
	cfg.ControlPlane.Version = "2.26.0"
	cfg.Cluster.Version = "2.26.0"

	opts := baseRunOptions()
	opts.Env = spec.EnvProd
	opts.EventsFile = eventsPath
	opts.Registry = &spec.Registry{Host: "registry.local:5000"}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	cpCalls := helm.CallsByRelease("runai-backend")
	require.Len(t, cpCalls, 1)
	assert.True(t, strings.HasSuffix(cpCalls[0].ChartName, "control-plane-2.20.0.tgz"), "installs the already-extracted local chart")
	assert.Contains(t, cpCalls[0].ValueFiles, filepath.Join(workingDir, "custom-env.yaml"))
	assert.Subset(t, cpCalls[0].Values, []string{
		"global.domain=test.runai.local",
	})
	assert.NotContains(t, cpCalls[0].Values, "global.customCA.enabled=true")

	clCalls := helm.CallsByRelease("runai-cluster")
	require.Len(t, clCalls, 1)
	assert.True(t, strings.HasSuffix(clCalls[0].ChartName, "runai-cluster-2.20.0.tgz"), "installs the already-extracted local chart")
	assert.Contains(t, clCalls[0].ValueFiles, filepath.Join(workingDir, "custom-env.yaml"))
	assert.Subset(t, clCalls[0].Values, []string{
		"controlPlane.url=test.runai.local",
		"controlPlane.clientSecret=s3cr3t-airgap",
		"cluster.uid=uuid-airgap-test",
	})
	assert.NotContains(t, clCalls[0].Values, "global.customCA.enabled=true")

	assert.True(t, apiClient.ClusterCreated, "cluster registered via control-plane API")

	evs := readEvents(t, eventsPath)
	assert.Nil(t, findEvent(evs, "step_started", "extract-airgap-package"), "extraction is skipped")
	assert.Nil(t, findEvent(evs, "step_started", "push-airgap-images"), "image push is skipped")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
}
