/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package plan

import (
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stepNames(f *orchestrator2.Flow) []string {
	names := make([]string, 0, len(f.Steps))
	for _, s := range f.Steps {
		names = append(names, s.Name())
	}
	return names
}

func stepForName(t *testing.T, f *orchestrator2.Flow, stepName string) orchestrator2.Step {
	t.Helper()
	for _, s := range f.Steps {
		if s.Name() == stepName {
			return s
		}
	}
	t.Fatalf("step %q not found", stepName)
	return nil
}

// actionForStep returns the single perform action a named step builds, so tests
// can assert the concrete action type (e.g. local vs remote helm) rather than
// only the step's name.
func actionForStep(t *testing.T, f *orchestrator2.Flow, stepName string) orchestrator2.PerformAction {
	t.Helper()
	for _, s := range f.Steps {
		if s.Name() == stepName {
			acts, err := s.BuildActions(nil)
			require.NoError(t, err)
			require.Len(t, acts, 1, "step %q should build exactly one action", stepName)
			return acts[0]
		}
	}
	t.Fatalf("step %q not found in flow %q", stepName, f.Name)
	return nil
}

func fullInstallSpec() *spec.Spec {
	return &spec.Spec{
		Command:  spec.CommandInstall,
		Registry: &spec.Registry{Type: spec.RegistryTypeJFrog, Token: "tok"},
		ControlPlane: &spec.ControlPlane{
			Domain:  "test.runai.local",
			Version: "2.26.0",
			TLS:     &spec.TLS{CertPath: "/cp.crt", KeyPath: "/cp.key", CAChainPath: "/ca.crt"},
		},
		Cluster: &spec.Cluster{
			Name: "test-cluster",
			TLS:  &spec.TLS{CertPath: "/cl.crt", KeyPath: "/cl.key", CAChainPath: "/ca.crt"},
		},
	}
}

func TestResolve_Install_ControlPlaneAndCluster(t *testing.T) {
	pf, inst, err := Resolve(fullInstallSpec(), orchestrator2.Deps{})
	require.NoError(t, err)
	require.Len(t, pf.Pipelines, 1)
	assert.Equal(t, "preflight", pf.Pipelines[0].Name)
	require.Len(t, inst.Pipelines, 2)
	assert.Equal(t, "prereq", inst.Pipelines[0].Name)

	runai := inst.Pipelines[1]
	assert.Equal(t, "runai", runai.Name)
	require.Len(t, runai.Flows, 2)

	assert.Equal(t, "control-plane", runai.Flows[0].Name)
	assert.Equal(t, []string{
		"create-control-plane-namespace",
		"install-registry-secret",
		"install-control-plane-tls",
		"install-control-plane-ca",
		"install-control-plane-chart",
	}, stepNames(runai.Flows[0]))

	assert.Equal(t, "cluster", runai.Flows[1].Name)
	assert.Equal(t, []string{
		"create-cluster-namespace",
		"install-cluster-tls",
		"install-cluster-ca",
		"register-cluster",
		"install-cluster-chart",
	}, stepNames(runai.Flows[1]))

	// Connected installs use the remote-repo helm actions.
	assert.IsType(t, actions.HelmInstall{}, actionForStep(t, runai.Flows[0], "install-control-plane-chart"))
	assert.IsType(t, actions.InstallClusterChart{}, actionForStep(t, runai.Flows[1], "install-cluster-chart"))
}

func TestResolve_Install_ControlPlaneOnly_NoOptionalSecrets(t *testing.T) {
	s := &spec.Spec{
		Command:      spec.CommandInstall,
		ControlPlane: &spec.ControlPlane{Domain: "test.runai.local", Version: "2.26.0"},
	}

	_, inst, err := Resolve(s, orchestrator2.Deps{})
	require.NoError(t, err)
	runai := inst.Pipelines[1]
	require.Len(t, runai.Flows, 1)
	assert.Equal(t, "control-plane", runai.Flows[0].Name)
	assert.Equal(t, []string{
		"create-control-plane-namespace",
		"install-control-plane-chart",
	}, stepNames(runai.Flows[0]))
}

func TestControlPlaneValues_CustomCA(t *testing.T) {
	s := &spec.Spec{ControlPlane: &spec.ControlPlane{Domain: "runai.example.com"}}

	assert.Contains(t, controlPlaneValues(s, "/ca.crt"), "global.customCA.enabled=true")
	assert.NotContains(t, controlPlaneValues(s, ""), "global.customCA.enabled=true")
	assert.Contains(t, controlPlaneValuesAirgap(s, "/ca.crt"), "global.customCA.enabled=true")
	assert.NotContains(t, controlPlaneValuesAirgap(s, ""), "global.customCA.enabled=true")
}

func TestHasServingCertificate(t *testing.T) {
	assert.False(t, hasServingCertificate(nil))
	assert.False(t, hasServingCertificate(&spec.TLS{}))
	assert.False(t, hasServingCertificate(&spec.TLS{CAChainPath: "/ca.crt"}))
	assert.False(t, hasServingCertificate(&spec.TLS{CertPath: "/tls.crt"}))
	assert.False(t, hasServingCertificate(&spec.TLS{KeyPath: "/tls.key"}))
	assert.True(t, hasServingCertificate(&spec.TLS{CertPath: "/tls.crt", KeyPath: "/tls.key"}))
}

func TestResolve_Install_Airgap(t *testing.T) {
	s := fullInstallSpec()
	s.Connectivity = spec.ConnectivityAirGapped
	s.Airgap = &spec.Airgap{BundlePath: "/tmp/bundle.tar.gz", WorkingDir: "/tmp/runai-airgap"}
	s.Registry = &spec.Registry{Host: "mirror.local:5000", Username: "admin", Password: "Harbor12345"}

	_, inst, err := Resolve(s, orchestrator2.Deps{})
	require.NoError(t, err)
	runai := inst.Pipelines[1]
	require.Len(t, runai.Flows, 3)

	assert.Equal(t, "airgap-preparation", runai.Flows[0].Name)
	assert.Equal(t, []string{
		"extract-airgap-package",
		"push-airgap-images",
	}, stepNames(runai.Flows[0]))

	// Air-gapped control plane skips the JFrog registry secret.
	assert.Equal(t, []string{
		"create-control-plane-namespace",
		"install-control-plane-tls",
		"install-control-plane-ca",
		"install-control-plane-chart",
	}, stepNames(runai.Flows[1]))

	assert.Equal(t, []string{
		"create-cluster-namespace",
		"install-cluster-tls",
		"install-cluster-ca",
		"register-cluster",
		"install-cluster-chart",
	}, stepNames(runai.Flows[2]))

	// Air-gapped installs use the local-chart helm actions, not the remote ones.
	assert.IsType(t, actions.HelmInstallLocal{}, actionForStep(t, runai.Flows[1], "install-control-plane-chart"))
	assert.IsType(t, actions.InstallClusterChartLocal{}, actionForStep(t, runai.Flows[2], "install-cluster-chart"))

	// The configured mirror registry is wired into the image-push action.
	push := actionForStep(t, runai.Flows[0], "push-airgap-images").(actions.PushAirgapImages)
	assert.Equal(t, "mirror.local:5000", push.RegistryURL)
	assert.Equal(t, "admin", push.Username)
	assert.Equal(t, "Harbor12345", push.Password)

	// The image push gets extra time on top of the configured step timeout.
	assert.Equal(t, 120*time.Minute, stepForName(t, runai.Flows[0], "push-airgap-images").ExtraTimeout())
}

func TestResolve_Install_AirgapSkipPrepare(t *testing.T) {
	s := fullInstallSpec()
	s.Connectivity = spec.ConnectivityAirGapped
	s.Airgap = &spec.Airgap{WorkingDir: "/tmp/runai-airgap", SkipPrepare: true}

	_, inst, err := Resolve(s, orchestrator2.Deps{})
	require.NoError(t, err)
	runai := inst.Pipelines[1]
	require.Len(t, runai.Flows, 2, "airgap-preparation flow must be omitted when SkipPrepare is set")
	assert.Equal(t, "control-plane", runai.Flows[0].Name)
	assert.Equal(t, "cluster", runai.Flows[1].Name)
}

func TestResolve_Install_DryRunPropagates(t *testing.T) {
	s := fullInstallSpec()
	s.DryRun = true

	pf, inst, err := Resolve(s, orchestrator2.Deps{})
	require.NoError(t, err)
	for _, pipeline := range pf.Pipelines {
		assert.False(t, pipeline.DryRun, "preflight is read-only and runs for real even under dry-run")
	}
	for _, pipeline := range inst.Pipelines {
		assert.True(t, pipeline.DryRun, "install pipeline %s should be dry-run", pipeline.Name)
	}
}

func TestResolve_Upgrade_Empty(t *testing.T) {
	_, inst, err := Resolve(&spec.Spec{Command: spec.CommandUpgrade}, orchestrator2.Deps{})
	require.NoError(t, err)
	assert.Empty(t, inst.Pipelines)
}

func TestResolve_UnknownCommand(t *testing.T) {
	_, _, err := Resolve(&spec.Spec{Command: spec.Command("bogus")}, orchestrator2.Deps{})
	require.ErrorContains(t, err, "unsupported command")
}
