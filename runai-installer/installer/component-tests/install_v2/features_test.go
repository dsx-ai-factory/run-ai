/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_v2_test

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/installconfig"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// runFeaturesInstall executes a connected install whose Features block is set by
// mutate, returning the fake helm installer and decoded events for assertions.
func runFeaturesInstall(t *testing.T, mutate func(*installconfig.Config)) (*harness.FakeHelmInstaller, *harness.FakeK8sExecutor, []map[string]any) {
	t.Helper()
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	mutate(cfg)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))
	return helm, k8s, readEvents(t, eventsPath)
}

// TestConnectedInstall_NIMOperator verifies the NIM Operator feature installs
// the NGC chart with the matrix-resolved version into the nim-operator
// namespace, in the features pipeline that runs after Run:ai.
func TestConnectedInstall_NIMOperator(t *testing.T) {
	helm, _, evs := runFeaturesInstall(t, func(c *installconfig.Config) {
		c.Features = &installconfig.Features{NIMOperator: &installconfig.Dependency{Mode: "install"}}
	})

	nim := helm.CallsByRelease("nim-operator")
	require.Len(t, nim, 1)
	assert.Equal(t, "k8s-nim-operator", nim[0].ChartName)
	assert.Equal(t, "nim-operator", nim[0].Namespace)
	assert.Equal(t, "3.1.1", nim[0].Version, "version resolved from the 2.26 matrix")
	assert.Contains(t, nim[0].RepoUrl, "helm.ngc.nvidia.com")
	assert.False(t, nim[0].IsOci, "NIM operator installs from the classic NGC helm repo")

	require.NotNil(t, findEvent(evs, "step_succeeded", "install-nim-operator"))
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "features"))

	// Ordering: features install on top of Run:ai, so the feature step must
	// start only after the Run:ai pipeline has completed.
	runaiDone := pipelineEventIndex(evs, "pipeline_succeeded", "runai")
	nimStart := eventIndex(evs, "step_started", "install-nim-operator")
	require.NotEqual(t, -1, runaiDone)
	require.NotEqual(t, -1, nimStart)
	assert.Less(t, runaiDone, nimStart, "features must install after Run:ai")
}

// TestConnectedInstall_Grove verifies the Grove feature installs the OCI chart
// with the matrix-resolved version into the grove-system namespace.
func TestConnectedInstall_Grove(t *testing.T) {
	helm, _, evs := runFeaturesInstall(t, func(c *installconfig.Config) {
		c.Features = &installconfig.Features{Grove: &installconfig.Dependency{Mode: "install"}}
	})

	grove := helm.CallsByRelease("grove")
	require.Len(t, grove, 1)
	assert.Equal(t, "grove-charts", grove[0].ChartName)
	assert.Equal(t, "grove-system", grove[0].Namespace)
	assert.Equal(t, "v0.1.0-alpha.10-rc1", grove[0].Version, "version resolved from the 2.26 matrix")
	assert.Contains(t, grove[0].RepoUrl, "oci://ghcr.io/ai-dynamo/grove/grove-charts")
	assert.True(t, grove[0].IsOci, "Grove installs from an OCI registry")

	require.NotNil(t, findEvent(evs, "step_succeeded", "install-grove"))
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
}

// TestConnectedInstall_Dynamo verifies the Dynamo feature installs the platform
// OCI chart into dynamo-system with the matrix-resolved version, integrating
// (not bundling) KAI and Grove. Grove is enabled alongside it because Dynamo
// requires it.
func TestConnectedInstall_Dynamo(t *testing.T) {
	helm, _, evs := runFeaturesInstall(t, func(c *installconfig.Config) {
		c.Features = &installconfig.Features{
			Grove:  &installconfig.Dependency{Mode: "install"},
			Dynamo: &installconfig.Dependency{Mode: "install"},
		}
	})

	dynamo := helm.CallsByRelease("dynamo-platform")
	require.Len(t, dynamo, 1)
	assert.Equal(t, "dynamo-platform", dynamo[0].ChartName)
	assert.Equal(t, "dynamo-system", dynamo[0].Namespace)
	assert.Equal(t, "1.2.1", dynamo[0].Version, "version resolved from the 2.26 matrix")
	assert.Contains(t, dynamo[0].RepoUrl, "https://helm.ngc.nvidia.com/nvidia/ai-dynamo")
	assert.False(t, dynamo[0].IsOci, "Dynamo installs from NGC's anonymous classic Helm repo, not OCI")
	// KAI is integrated, never bundled (the Run:ai cluster provides it); Grove
	// is integrated (installed by its own feature step).
	assert.Contains(t, dynamo[0].Values, "global.kai-scheduler.install=false")
	assert.Contains(t, dynamo[0].Values, "global.kai-scheduler.enabled=true")
	assert.Contains(t, dynamo[0].Values, "global.grove.install=false")
	assert.Contains(t, dynamo[0].Values, "global.grove.enabled=true")

	require.NotNil(t, findEvent(evs, "step_succeeded", "install-dynamo"))
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
}

// TestConnectedInstall_NIMAndGrove verifies both features install together and
// are counted in the features summary, while the mandatory prerequisites are
// counted separately in the prereq summary.
func TestConnectedInstall_NIMAndGrove(t *testing.T) {
	helm, _, evs := runFeaturesInstall(t, func(c *installconfig.Config) {
		c.Features = &installconfig.Features{
			NIMOperator: &installconfig.Dependency{Mode: "install"},
			Grove:       &installconfig.Dependency{Mode: "install"},
		}
	})

	assert.Len(t, helm.CallsByRelease("nim-operator"), 1)
	assert.Len(t, helm.CallsByRelease("grove"), 1)

	prereq := prereqSummaryEvent(t, evs)
	assert.EqualValues(t, 3, prereq["installed"], "prometheus + gpu + network")
	assert.Equal(t, "installed", prereq["outcome"])

	features := featuresSummaryEvent(t, evs)
	assert.EqualValues(t, 2, features["installed"], "nim + grove")
	assert.Equal(t, "installed", features["outcome"])
}

// TestConnectedInstall_Grove_AlreadyPresent guards the Grove sentinel: its
// controller Deployment is named "grove-operator", so an existing install must
// be detected and skipped rather than reinstalled.
func TestConnectedInstall_Grove_AlreadyPresent(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{Existing: map[client.ObjectKey]bool{
		{Namespace: "grove-system", Name: "grove-operator"}: true,
	}}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Features = &installconfig.Features{Grove: &installconfig.Dependency{Mode: "install"}}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.Empty(t, helm.CallsByRelease("grove"), "present Grove operator is not reinstalled")
	skipped := findEvent(readEvents(t, eventsPath), "step_skipped", "install-grove")
	require.NotNil(t, skipped)
	assert.Equal(t, "already_installed", skipped["skipReason"])
}

// TestConnectedInstall_NIMOperator_AlreadyPresent guards the NIM sentinel: the
// operator's controller Deployment is named "nim-operator-k8s-nim-operator"
// (Helm <release>-<chart> fullname), so an existing install must be detected
// and skipped rather than reinstalled.
func TestConnectedInstall_NIMOperator_AlreadyPresent(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{Existing: map[client.ObjectKey]bool{
		{Namespace: "nim-operator", Name: "nim-operator-k8s-nim-operator"}: true,
	}}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Features = &installconfig.Features{NIMOperator: &installconfig.Dependency{Mode: "install"}}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.Empty(t, helm.CallsByRelease("nim-operator"), "present NIM operator is not reinstalled")
	skipped := findEvent(readEvents(t, eventsPath), "step_skipped", "install-nim-operator")
	require.NotNil(t, skipped)
	assert.Equal(t, "already_installed", skipped["skipReason"])
}

// TestPreflight_NIMAdmissionWebhook_WarnsWhenEnabled verifies that enabling the
// NIM Operator with its admission controller left on raises a non-blocking
// preflight warning (its webhook needs cert-manager).
func TestPreflight_NIMAdmissionWebhook_WarnsWhenEnabled(t *testing.T) {
	_, _, evs := runFeaturesInstall(t, func(c *installconfig.Config) {
		c.Features = &installconfig.Features{NIMOperator: &installconfig.Dependency{Mode: "install"}}
	})

	step := findEvent(evs, "step_succeeded", "preflight-nim-admission-webhook")
	require.NotNil(t, step, "the NIM admission-webhook check must run when NIM is enabled")
	result := step["results"].([]any)[0].(map[string]any)
	assert.Equal(t, "warn", result["status"])
	assert.Contains(t, result["reason"], "cert-manager")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"), "a warning must not block the install")
}

// TestPreflight_NIMAdmissionWebhook_NoWarnWhenDisabled verifies the check is not
// assembled when the admission controller is explicitly disabled.
func TestPreflight_NIMAdmissionWebhook_NoWarnWhenDisabled(t *testing.T) {
	_, _, evs := runFeaturesInstall(t, func(c *installconfig.Config) {
		c.Features = &installconfig.Features{NIMOperator: &installconfig.Dependency{
			Mode:             "install",
			CustomHelmValues: map[string]any{"operator.admissionController.enabled": false},
		}}
	})

	assert.Nil(t, findEvent(evs, "step_succeeded", "preflight-nim-admission-webhook"),
		"no warning when the admission controller is disabled")
}

// TestConnectedInstall_NIMOperator_FailureFailsRun verifies that a NIM Operator
// helm failure fails the run, but only the features pipeline fails: features
// install on top of Run:ai, so Run:ai (control plane + cluster) is already up
// by the time the feature fails.
func TestConnectedInstall_NIMOperator_FailureFailsRun(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	helm.InstallErrors = map[string]error{"nim-operator": errors.New("helm boom")}
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Features = &installconfig.Features{NIMOperator: &installconfig.Dependency{Mode: "install"}}
	term, _, _ := newTestTerminal()

	require.Error(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	// Run:ai installed successfully before the feature ran.
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	assert.Len(t, helm.CallsByRelease("runai-backend"), 1, "control plane installs before features")
	assert.True(t, apiClient.ClusterCreated)
	// The feature failure surfaces as a features-pipeline failure.
	assert.True(t, hasPipelineEvent(evs, "pipeline_failed", "features"))
	assert.False(t, hasPipelineEvent(evs, "pipeline_failed", "prereq"))
}

// runKubeflowInstall runs a connected install whose Features block is set by
// mutate, wiring a FakeManifestProvider (the non-Helm prerequisites use it).
// The caller supplies the FakeK8sExecutor so it can seed cluster state (e.g. a
// pre-existing CRD).
func runKubeflowInstall(t *testing.T, k8s *harness.FakeK8sExecutor, mutate func(*installconfig.Config)) (*harness.FakeManifestProvider, []map[string]any) {
	t.Helper()
	helm := harness.NewFakeHelmWithReleases()
	manifests := &harness.FakeManifestProvider{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), ManifestProvider: manifests}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	mutate(cfg)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))
	return manifests, readEvents(t, eventsPath)
}

// TestConnectedInstall_KubeflowTraining verifies the training operator installs
// from its kustomize overlay (matrix version), with no MPI/coexistence work.
func TestConnectedInstall_KubeflowTraining(t *testing.T) {
	k8s := &harness.FakeK8sExecutor{}
	manifests, evs := runKubeflowInstall(t, k8s, func(c *installconfig.Config) {
		c.Features = &installconfig.Features{Kubeflow: &installconfig.Dependency{Mode: "install"}}
	})

	kustomize := manifests.CallsByOp("kustomize")
	require.Len(t, kustomize, 1)
	assert.Contains(t, kustomize[0].TarballURL, "training-operator")
	assert.Contains(t, kustomize[0].TarballURL, "v1.8.1", "version resolved from the matrix")
	assert.Equal(t, "manifests/overlays/standalone", kustomize[0].SubPath)

	require.NotNil(t, findEvent(evs, "step_succeeded", "install-training-operator"))
	assert.Empty(t, manifests.CallsByOp("fetch"), "no MPI operator when only training is requested")
	assert.Nil(t, findEvent(evs, "step_started", "reconcile-training-mpi-coexistence"), "no coexistence without MPI")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
}

// TestConnectedInstall_MPIOperator verifies the standalone MPI operator installs
// from its flat manifest, with no training operator or coexistence work.
func TestConnectedInstall_MPIOperator(t *testing.T) {
	k8s := &harness.FakeK8sExecutor{}
	manifests, evs := runKubeflowInstall(t, k8s, func(c *installconfig.Config) {
		c.Features = &installconfig.Features{MPIOperator: &installconfig.Dependency{Mode: "install"}}
	})

	fetched := manifests.CallsByOp("fetch")
	require.Len(t, fetched, 1)
	assert.Contains(t, fetched[0].URL, "mpi-operator")
	assert.Contains(t, fetched[0].URL, "v0.6.0", "version resolved from the matrix")

	require.NotNil(t, findEvent(evs, "step_succeeded", "install-mpi-operator"))
	assert.Empty(t, manifests.CallsByOp("kustomize"), "no training operator when only MPI is requested")
	assert.Nil(t, findEvent(evs, "step_started", "reconcile-training-mpi-coexistence"), "no coexistence without training operator")
}

// TestConnectedInstall_KubeflowTraining_AlreadyPresent guards the training
// operator sentinel: its controller Deployment is named "training-operator" in
// the "kubeflow" namespace, so an existing install must be detected and skipped
// rather than re-applied from the kustomize overlay.
func TestConnectedInstall_KubeflowTraining_AlreadyPresent(t *testing.T) {
	k8s := &harness.FakeK8sExecutor{Existing: map[client.ObjectKey]bool{
		{Namespace: "kubeflow", Name: "training-operator"}: true,
	}}
	manifests, evs := runKubeflowInstall(t, k8s, func(c *installconfig.Config) {
		c.Features = &installconfig.Features{Kubeflow: &installconfig.Dependency{Mode: "install"}}
	})

	assert.Empty(t, manifests.CallsByOp("kustomize"), "present training operator is not re-applied")
	skipped := findEvent(evs, "step_skipped", "install-training-operator")
	require.NotNil(t, skipped)
	assert.Equal(t, "already_installed", skipped["skipReason"])
}

// TestConnectedInstall_MPIOperator_AlreadyPresent guards the MPI operator
// sentinel: its controller Deployment is named "mpi-operator" in the
// "mpi-operator" namespace, so an existing install must be detected and skipped
// rather than re-applied from its flat manifest.
func TestConnectedInstall_MPIOperator_AlreadyPresent(t *testing.T) {
	k8s := &harness.FakeK8sExecutor{Existing: map[client.ObjectKey]bool{
		{Namespace: "mpi-operator", Name: "mpi-operator"}: true,
	}}
	manifests, evs := runKubeflowInstall(t, k8s, func(c *installconfig.Config) {
		c.Features = &installconfig.Features{MPIOperator: &installconfig.Dependency{Mode: "install"}}
	})

	assert.Empty(t, manifests.CallsByOp("fetch"), "present MPI operator is not re-applied")
	skipped := findEvent(evs, "step_skipped", "install-mpi-operator")
	require.NotNil(t, skipped)
	assert.Equal(t, "already_installed", skipped["skipReason"])
}

// TestConnectedInstall_KubeflowTraining_RenderFailureFailsRun verifies a
// kustomize render failure for the training operator fails the run via the
// features pipeline. Features install on top of Run:ai, so Run:ai (control
// plane + cluster) is already up by the time the feature fails.
func TestConnectedInstall_KubeflowTraining_RenderFailureFailsRun(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	manifests := &harness.FakeManifestProvider{RenderErr: errors.New("kustomize boom")}
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), ManifestProvider: manifests}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Features = &installconfig.Features{Kubeflow: &installconfig.Dependency{Mode: "install"}}
	term, _, _ := newTestTerminal()

	require.Error(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	require.NotNil(t, findEvent(evs, "step_failed", "install-training-operator"))
	// Run:ai installed successfully before the feature ran.
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	assert.Len(t, helm.CallsByRelease("runai-backend"), 1, "control plane installs before features")
	assert.True(t, apiClient.ClusterCreated)
	// The feature failure surfaces as a features-pipeline failure.
	assert.True(t, hasPipelineEvent(evs, "pipeline_failed", "features"))
	assert.False(t, hasPipelineEvent(evs, "pipeline_failed", "prereq"))
}

// TestConnectedInstall_MPIOperator_FetchFailureFailsRun verifies a manifest
// fetch failure for the MPI operator fails the run via the features pipeline,
// after Run:ai is already up.
func TestConnectedInstall_MPIOperator_FetchFailureFailsRun(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	manifests := &harness.FakeManifestProvider{FetchErr: errors.New("fetch boom")}
	k8s := &harness.FakeK8sExecutor{}
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), ManifestProvider: manifests}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Features = &installconfig.Features{MPIOperator: &installconfig.Dependency{Mode: "install"}}
	term, _, _ := newTestTerminal()

	require.Error(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	require.NotNil(t, findEvent(evs, "step_failed", "install-mpi-operator"))
	// Run:ai installed successfully before the feature ran.
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	assert.Len(t, helm.CallsByRelease("runai-backend"), 1, "control plane installs before features")
	assert.True(t, apiClient.ClusterCreated)
	// The feature failure surfaces as a features-pipeline failure.
	assert.True(t, hasPipelineEvent(evs, "pipeline_failed", "features"))
	assert.False(t, hasPipelineEvent(evs, "pipeline_failed", "prereq"))
}

// TestConnectedInstall_KubeflowAndMPI_Coexistence verifies that when both are
// requested, the coexistence step disables MPI v1 in the training operator
// (deployment patch) and removes the conflicting v1 mpijobs CRD. The CRD is
// seeded as present (the training operator's overlay creates it) so the
// detection-gated step runs.
func TestConnectedInstall_KubeflowAndMPI_Coexistence(t *testing.T) {
	k8s := &harness.FakeK8sExecutor{SeededObjects: map[client.ObjectKey]*unstructured.Unstructured{
		{Name: "mpijobs.kubeflow.org"}: mpijobsCRD("v1"), // training operator's v1 CRD
	}}
	manifests, evs := runKubeflowInstall(t, k8s, func(c *installconfig.Config) {
		c.Features = &installconfig.Features{
			Kubeflow:    &installconfig.Dependency{Mode: "install"},
			MPIOperator: &installconfig.Dependency{Mode: "install"},
		}
	})

	require.Len(t, manifests.CallsByOp("kustomize"), 1)
	require.Len(t, manifests.CallsByOp("fetch"), 1)
	require.NotNil(t, findEvent(evs, "step_succeeded", "install-training-operator"))
	require.NotNil(t, findEvent(evs, "step_succeeded", "install-mpi-operator"))
	require.NotNil(t, findEvent(evs, "step_succeeded", "reconcile-training-mpi-coexistence"))

	patches := k8s.CallsByVerb("patch")
	require.Len(t, patches, 1, "training operator deployment is patched to disable MPI v1")
	assert.Equal(t, "training-operator", patches[0].Name)
	assert.Equal(t, "kubeflow", patches[0].Namespace)
	// The patch is a strategic merge targeting the container by name (not index),
	// setting only the non-MPI schemes.
	assert.Equal(t, "application/strategic-merge-patch+json", patches[0].PatchType)
	assert.Contains(t, string(patches[0].PatchData), `"name":"training-operator"`)
	assert.Contains(t, string(patches[0].PatchData), "--enable-scheme=tfjob")
	assert.NotContains(t, string(patches[0].PatchData), "mpijob", "MPI scheme must not be enabled in the training operator")

	deletes := k8s.CallsByVerb("delete")
	require.Len(t, deletes, 1, "the v1 mpijobs CRD is removed")
	assert.Equal(t, "mpijobs.kubeflow.org", deletes[0].Name)

	// Ordering regression guard: the v1 mpijobs CRD must be deleted (coexistence)
	// before the MPI operator installs its v2beta1 CRD.
	coexDone := eventIndex(evs, "step_succeeded", "reconcile-training-mpi-coexistence")
	mpiStart := eventIndex(evs, "step_started", "install-mpi-operator")
	require.NotEqual(t, -1, coexDone)
	require.NotEqual(t, -1, mpiStart)
	assert.Less(t, coexDone, mpiStart, "coexistence must complete before the MPI operator installs")
}

// TestDryRun_KubeflowAndMPI_CoexistenceSurfacesDestructiveDelete verifies that a
// dry-run surfaces the coexistence step's destructive v1 mpijobs CRD deletion,
// rather than hiding it behind an "already_installed" skip on a cluster where
// the CRD has not actually been created.
func TestDryRun_KubeflowAndMPI_CoexistenceSurfacesDestructiveDelete(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	manifests := &harness.FakeManifestProvider{}
	k8s := &harness.FakeK8sExecutor{}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe(), ManifestProvider: manifests}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, true) // dry-run
	cfg.Features = &installconfig.Features{
		Kubeflow:    &installconfig.Dependency{Mode: "install"},
		MPIOperator: &installconfig.Dependency{Mode: "install"},
	}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	evs := readEvents(t, eventsPath)
	skipped := findEvent(evs, "step_skipped", "reconcile-training-mpi-coexistence")
	require.NotNil(t, skipped, "coexistence must appear in the dry-run plan")
	assert.Equal(t, "dry_run", skipped["skipReason"], "dry-run shows the actions, not an already_installed skip")
	output, _ := skipped["output"].(string)
	assert.Contains(t, output, "mpijobs.kubeflow.org")
	assert.Contains(t, output, "destructive", "the dry-run must flag the CRD deletion as destructive")
	assert.Empty(t, k8s.CallsByVerb("delete"), "dry-run performs no real deletion")
}

// TestConnectedInstall_KubeflowAndMPI_CoexistenceIdempotent verifies that on a
// re-run, the coexistence step does NOT delete the MPI operator's own v2beta1
// mpijobs CRD: with only a v2beta1 CRD present, the step is skipped and makes no
// patch/delete. This guards against deleting a live MPI CRD on every re-run.
func TestConnectedInstall_KubeflowAndMPI_CoexistenceIdempotent(t *testing.T) {
	k8s := &harness.FakeK8sExecutor{SeededObjects: map[client.ObjectKey]*unstructured.Unstructured{
		{Name: "mpijobs.kubeflow.org"}: mpijobsCRD("v2beta1"), // MPI operator's CRD, already reconciled
	}}
	_, evs := runKubeflowInstall(t, k8s, func(c *installconfig.Config) {
		c.Features = &installconfig.Features{
			Kubeflow:    &installconfig.Dependency{Mode: "install"},
			MPIOperator: &installconfig.Dependency{Mode: "install"},
		}
	})

	skipped := findEvent(evs, "step_skipped", "reconcile-training-mpi-coexistence")
	require.NotNil(t, skipped, "coexistence is skipped when only the v2beta1 CRD remains")
	assert.Equal(t, "already_installed", skipped["skipReason"])
	assert.Empty(t, k8s.CallsByVerb("patch"))
	assert.Empty(t, k8s.CallsByVerb("delete"), "the MPI operator's v2beta1 CRD must never be deleted")
}

// mpijobsCRD builds an mpijobs CustomResourceDefinition declaring the given
// served versions, for seeding the fake cluster in coexistence tests.
func mpijobsCRD(versions ...string) *unstructured.Unstructured {
	vs := make([]any, 0, len(versions))
	for _, v := range versions {
		vs = append(vs, map[string]any{"name": v})
	}
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "CustomResourceDefinition",
		"metadata":   map[string]any{"name": "mpijobs.kubeflow.org"},
		"spec":       map[string]any{"versions": vs},
	}}
}
