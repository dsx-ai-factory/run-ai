/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package plan

import (
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/actions"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2/spec"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
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
		"install-cluster-registry-secret",
		"install-cluster-tls",
		"install-cluster-ca",
		"register-cluster",
		"install-cluster-chart",
	}, stepNames(runai.Flows[1]))

	// Connected installs use the remote-repo helm actions.
	cpAction := actionForStep(t, runai.Flows[0], "install-control-plane-chart")
	require.IsType(t, actions.HelmInstall{}, cpAction)
	assert.Contains(t, cpAction.(actions.HelmInstall).Values, "global.imagePullSecrets[0].name=runai-reg-creds")

	clusterAction := actionForStep(t, runai.Flows[1], "install-cluster-chart")
	require.IsType(t, actions.InstallClusterChart{}, clusterAction)
	assert.Contains(t, clusterAction.(actions.InstallClusterChart).ExtraValues, "global.additionalImagePullSecrets[0]=runai-reg-creds")
}

func TestResolve_Install_ExternalControlPlane(t *testing.T) {
	s := &spec.Spec{
		Command:      spec.CommandInstall,
		InstallMode:  spec.InstallModeClusterOnly,
		Connectivity: spec.ConnectivityConnected,
		Registry:     &spec.Registry{Type: spec.RegistryTypeNGC},
		Cluster: &spec.Cluster{
			Name:            "cluster-a",
			Domain:          "cluster.example.com",
			Version:         "2.26.0",
			ControlPlaneURL: "https://cp.example.com",
			UID:             "cluster-uid",
			BootstrapToken:  "bootstrap-token",
		},
	}

	_, inst, err := Resolve(s, orchestrator2.Deps{})
	require.NoError(t, err)
	runai := inst.Pipelines[1]
	require.Len(t, runai.Flows, 1)
	clusterFlow := runai.Flows[0]
	assert.NotContains(t, stepNames(clusterFlow), "register-cluster")

	action := actionForStep(t, clusterFlow, "install-cluster-chart")
	require.IsType(t, actions.InstallClusterChart{}, action)
	clusterAction := action.(actions.InstallClusterChart)
	assert.Equal(t, "https://cp.example.com", clusterAction.ControlPlaneURL)
	assert.Equal(t, "cluster.example.com", clusterAction.Domain)
	assert.Equal(t, "cluster-uid", clusterAction.ClusterUID)
	assert.Equal(t, "bootstrap-token", clusterAction.BootstrapToken)
}

func TestResolve_Install_ExternalControlPlaneCredentials(t *testing.T) {
	s := &spec.Spec{
		Command:      spec.CommandInstall,
		InstallMode:  spec.InstallModeClusterOnly,
		Connectivity: spec.ConnectivityConnected,
		Registry:     &spec.Registry{Type: spec.RegistryTypeNGC},
		Cluster: &spec.Cluster{
			Name:                    "cluster-a",
			Domain:                  "cluster.example.com",
			Version:                 "2.26.0",
			TLS:                     &spec.TLS{CAChainPath: "/ca.pem"},
			ControlPlaneURL:         "https://cp.example.com",
			ControlPlaneCredentials: &spec.ControlPlaneCredentials{Username: "admin@example.com", Password: "pw"},
		},
	}
	deps := orchestrator2.Deps{RunAIClientWithCredentials: func(string, string, step.ControlPlaneCredentials) (step.RunAIAPIClient, error) { return nil, nil }}

	pre, inst, err := Resolve(s, deps)
	require.NoError(t, err)

	require.Len(t, pre.Pipelines, 1)
	require.Len(t, pre.Pipelines[0].Flows, 1)
	assert.Contains(t, stepNames(pre.Pipelines[0].Flows[0]), "preflight-control-plane-access")
	access := actionForStep(t, pre.Pipelines[0].Flows[0], "preflight-control-plane-access").(actions.CheckControlPlaneAccess)
	assert.Equal(t, "cluster-a", access.ClusterName)
	assert.Equal(t, "https://cp.example.com", access.ControlPlane.URL)
	assert.Equal(t, "/ca.pem", access.ControlPlane.CACertPath)
	require.NotNil(t, access.ControlPlane.Credentials)
	assert.Equal(t, "admin@example.com", access.ControlPlane.Credentials.Username)

	runai := inst.Pipelines[1]
	require.Len(t, runai.Flows, 1)
	clusterFlow := runai.Flows[0]
	assert.Contains(t, stepNames(clusterFlow), "register-cluster", "the installer registers the cluster itself")

	register := actionForStep(t, clusterFlow, "register-cluster").(actions.EnsureCluster)
	assert.Equal(t, "cluster-a", register.ClusterName)
	assert.Equal(t, "https://cp.example.com", register.ControlPlane.URL)
	assert.Equal(t, "cluster.example.com", register.ControlPlane.Domain)
	assert.Equal(t, "/ca.pem", register.ControlPlane.CACertPath)
	require.NotNil(t, register.ControlPlane.Credentials)
	assert.Equal(t, "pw", register.ControlPlane.Credentials.Password)

	clusterAction := actionForStep(t, clusterFlow, "install-cluster-chart").(actions.InstallClusterChart)
	assert.Equal(t, "https://cp.example.com", clusterAction.ControlPlaneURL)
	assert.Equal(t, "cluster.example.com", clusterAction.Domain)
	assert.Empty(t, clusterAction.ClusterUID)
	assert.Empty(t, clusterAction.BootstrapToken)
	assert.False(t, clusterAction.UsesSuppliedCredentials())
	require.NotNil(t, clusterAction.Credentials)
	assert.Equal(t, "admin@example.com", clusterAction.Credentials.Username)
}

func TestResolve_Install_ExternalControlPlane_NoAccessCheckWithSuppliedIdentity(t *testing.T) {
	s := &spec.Spec{
		Command:      spec.CommandInstall,
		InstallMode:  spec.InstallModeClusterOnly,
		Connectivity: spec.ConnectivityConnected,
		Registry:     &spec.Registry{Type: spec.RegistryTypeNGC},
		Cluster: &spec.Cluster{
			Name:            "cluster-a",
			Domain:          "cluster.example.com",
			ControlPlaneURL: "https://cp.example.com",
			UID:             "cluster-uid",
			BootstrapToken:  "bootstrap-token",
		},
	}
	deps := orchestrator2.Deps{RunAIClientWithCredentials: func(string, string, step.ControlPlaneCredentials) (step.RunAIAPIClient, error) { return nil, nil }}

	pre, _, err := Resolve(s, deps)
	require.NoError(t, err)
	for _, f := range pre.Pipelines[0].Flows {
		assert.NotContains(t, stepNames(f), "preflight-control-plane-access")
	}
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

func TestUsesManagedRegistrySecret(t *testing.T) {
	registry := &spec.Registry{Type: spec.RegistryTypeJFrog, Token: "tok"}

	assert.True(t, usesManagedRegistrySecret(&spec.Spec{Registry: registry}))
	assert.True(t, usesManagedRegistrySecret(&spec.Spec{
		Connectivity: spec.ConnectivityAirGapped,
		DevMode:      true,
		ControlPlane: &spec.ControlPlane{Version: "3.0.0"},
		Registry:     registry,
	}))
	assert.False(t, usesManagedRegistrySecret(&spec.Spec{
		Connectivity: spec.ConnectivityAirGapped,
		ControlPlane: &spec.ControlPlane{Version: "2.26.0"},
		Registry:     registry,
	}))
	assert.False(t, usesManagedRegistrySecret(&spec.Spec{}))
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
