/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package install_v2_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/component-tests/harness"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/installconfig"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/specrun"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	nfdCSV = "nfd.v4.18.0"
	gpuCSV = "gpu-operator-certified.v25.3.0"
)

func olmSubscriptionWithCSV(csv string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "operators.coreos.com/v1alpha1",
		"kind":       "Subscription",
		"status":     map[string]any{"installedCSV": csv},
	}}
}

// olmCSVWithExamples seeds a Succeeded CSV that carries the operator's default
// custom resource in its alm-examples annotation, which the GPU flow reads to
// build the ClusterPolicy / NodeFeatureDiscovery (the empty-spec CR is rejected
// by the operator's required-field validation on a real cluster).
func olmCSVWithExamples(alm string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "operators.coreos.com/v1alpha1",
		"kind":       "ClusterServiceVersion",
		"metadata":   map[string]any{"annotations": map[string]any{"alm-examples": alm}},
		"status":     map[string]any{"phase": "Succeeded"},
	}}
}

const (
	nfdALMExamples = `[{"apiVersion":"nfd.openshift.io/v1","kind":"NodeFeatureDiscovery","metadata":{"name":"nfd-instance"},"spec":{"operand":{"imagePullPolicy":"IfNotPresent"}}}]`
	gpuALMExamples = `[{"apiVersion":"nvidia.com/v1","kind":"ClusterPolicy","metadata":{"name":"gpu-cluster-policy"},"spec":{"operator":{"defaultRuntime":"crio"},"driver":{"enabled":true},"toolkit":{"enabled":true}}}]`
)

func establishedCRDObject(name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "apiextensions.k8s.io/v1",
		"kind":       "CustomResourceDefinition",
		"metadata":   map[string]any{"name": name},
		"status": map[string]any{
			"conditions": []any{map[string]any{"type": "Established", "status": "True"}},
		},
	}}
}

// olmCSVSucceeded seeds a Succeeded CSV with no alm-examples, for operators that
// install no custom resource (SR-IOV, NIM, cert-manager) — WaitForCSV only reads
// the phase.
func olmCSVSucceeded() *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "operators.coreos.com/v1alpha1",
		"kind":       "ClusterServiceVersion",
		"status":     map[string]any{"phase": "Succeeded"},
	}}
}

// seedNFD seeds the Node Feature Discovery operator (shared by the GPU and
// network flows) as installed with a Succeeded CSV and an Established CRD.
func seedNFD(k8s *harness.FakeK8sExecutor) {
	if k8s.SeededObjects == nil {
		k8s.SeededObjects = map[client.ObjectKey]*unstructured.Unstructured{}
	}
	k8s.SeededObjects[client.ObjectKey{Namespace: "openshift-nfd", Name: "nfd"}] = olmSubscriptionWithCSV(nfdCSV)
	k8s.SeededObjects[client.ObjectKey{Namespace: "openshift-nfd", Name: nfdCSV}] = olmCSVWithExamples(nfdALMExamples)
	k8s.SeededObjects[client.ObjectKey{Name: "nodefeaturediscoveries.nfd.openshift.io"}] = establishedCRDObject("nodefeaturediscoveries.nfd.openshift.io")
}

// seedOpenShiftGPUOperator makes the OperatorHub GPU flow resolve immediately on
// the fake cluster: it seeds the NFD and GPU subscriptions as having installed a
// Succeeded CSV (so WaitForCSV returns) and their operator CRDs as Established
// (so the gated CR applies do not block).
func seedOpenShiftGPUOperator(k8s *harness.FakeK8sExecutor) {
	seedNFD(k8s)
	k8s.SeededObjects[client.ObjectKey{Namespace: "nvidia-gpu-operator", Name: "gpu-operator-certified"}] = olmSubscriptionWithCSV(gpuCSV)
	k8s.SeededObjects[client.ObjectKey{Namespace: "nvidia-gpu-operator", Name: gpuCSV}] = olmCSVWithExamples(gpuALMExamples)
	k8s.SeededObjects[client.ObjectKey{Name: "clusterpolicies.nvidia.com"}] = establishedCRDObject("clusterpolicies.nvidia.com")
}

func applyCallByKindName(k8s *harness.FakeK8sExecutor, kind, name string) *harness.K8sCall {
	for _, c := range k8s.CallsByVerb("apply") {
		if c.Kind == kind && c.Name == name {
			call := c
			return &call
		}
	}
	return nil
}

func subscriptionSource(t *testing.T, call *harness.K8sCall) string {
	t.Helper()
	require.NotNil(t, call)
	u, ok := call.Object.(*unstructured.Unstructured)
	require.True(t, ok, "subscription apply must carry an unstructured object")
	source, _, _ := unstructured.NestedString(u.Object, "spec", "source")
	return source
}

func subscriptionChannel(t *testing.T, call *harness.K8sCall) string {
	t.Helper()
	require.NotNil(t, call)
	u, ok := call.Object.(*unstructured.Unstructured)
	require.True(t, ok, "subscription apply must carry an unstructured object")
	channel, _, _ := unstructured.NestedString(u.Object, "spec", "channel")
	return channel
}

func subscriptionField(t *testing.T, call *harness.K8sCall, field string) string {
	t.Helper()
	require.NotNil(t, call)
	u, ok := call.Object.(*unstructured.Unstructured)
	require.True(t, ok, "subscription apply must carry an unstructured object")
	v, _, _ := unstructured.NestedString(u.Object, "spec", field)
	return v
}

// TestOpenShiftInstall_GPUOperatorViaOperatorHub verifies the real GPU operator
// installs through OperatorHub on OpenShift instead of the NGC Helm chart: NFD
// and the GPU operator are installed as OLM Subscriptions (each with its own
// OperatorGroup), the GPU namespace is labeled for OpenShift monitoring, and the
// NodeFeatureDiscovery + ClusterPolicy CRs are applied.
func TestOpenShiftInstall_GPUOperatorViaOperatorHub(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	seedOpenShiftPrereqs(k8s)
	apiClient, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Install.KubernetesDistribution = "openshift"
	cfg.ExternalConnection = nil
	skipPrometheus(cfg)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.Empty(t, helm.CallsByRelease("gpu-operator"), "on OpenShift the GPU operator must not install via Helm")

	nfdSub := applyCallByKindName(k8s, "Subscription", "nfd")
	require.NotNil(t, nfdSub, "the NFD operator is subscribed via OLM")
	assert.Equal(t, "redhat-operators", subscriptionSource(t, nfdSub))
	assert.Equal(t, "openshift-nfd", nfdSub.Namespace)

	assert.Equal(t, "stable", subscriptionChannel(t, nfdSub), "NFD channel resolves from the compat matrix")

	gpuSub := applyCallByKindName(k8s, "Subscription", "gpu-operator-certified")
	require.NotNil(t, gpuSub, "the GPU operator is subscribed via OLM")
	assert.Equal(t, "certified-operators", subscriptionSource(t, gpuSub))
	assert.Equal(t, "nvidia-gpu-operator", gpuSub.Namespace)
	assert.Equal(t, "v25.3", subscriptionChannel(t, gpuSub), "GPU channel derives from the matrix GPU version v25.3.0")
	assert.Equal(t, gpuCSV, subscriptionField(t, gpuSub, "startingCSV"), "GPU subscription pins the matrix version via startingCSV")
	assert.Equal(t, "Manual", subscriptionField(t, gpuSub, "installPlanApproval"), "default approval strategy is Manual so OLM does not upgrade in the background")
	assert.Equal(t, "Manual", subscriptionField(t, nfdSub, "installPlanApproval"), "the flow's approval strategy applies to the NFD dependency too")

	require.NotNil(t, applyCallByKindName(k8s, "OperatorGroup", "openshift-nfd-og"), "NFD gets an OperatorGroup")
	require.NotNil(t, applyCallByKindName(k8s, "OperatorGroup", "nvidia-gpu-operator-og"), "GPU gets an OperatorGroup")

	gpuNS := applyCallByKindName(k8s, "Namespace", "nvidia-gpu-operator")
	require.NotNil(t, gpuNS)
	ns, ok := gpuNS.Object.(*corev1.Namespace)
	require.True(t, ok)
	assert.Equal(t, "true", ns.Labels["openshift.io/cluster-monitoring"], "GPU namespace is labeled for OpenShift monitoring")

	require.NotNil(t, applyCallByKindName(k8s, "NodeFeatureDiscovery", "nfd-instance"), "the NFD instance CR is applied")

	cp := applyCallByKindName(k8s, "ClusterPolicy", "gpu-cluster-policy")
	require.NotNil(t, cp, "the GPU ClusterPolicy CR is applied")
	cpObj := cp.Object.(*unstructured.Unstructured)
	runtime, _, _ := unstructured.NestedString(cpObj.Object, "spec", "operator", "defaultRuntime")
	assert.Equal(t, "crio", runtime, "the ClusterPolicy spec comes from the operator's CSV default (alm-examples), not an empty spec")

	evs := readEvents(t, eventsPath)
	for _, step := range []string{
		"install-nfd-operator",
		"create-node-feature-discovery",
		"install-gpu-operator",
		"create-gpu-cluster-policy",
	} {
		require.NotNil(t, findEvent(evs, "step_succeeded", step), "step %s must succeed", step)
	}

	summary := prereqSummaryEvent(t, evs)
	assert.Equal(t, "installed", summary["outcome"])
	assert.EqualValues(t, 2, summary["installed"], "gpu-operator (OLM) + network (Prometheus is not installed on OpenShift)")
	assert.True(t, hasPipelineEvent(evs, "pipeline_succeeded", "runai"))
	assert.True(t, apiClient.ClusterCreated, "cluster registered via control-plane API")
}

// TestOpenShiftInstall_GPUOperator_CustomResourceOverride verifies that a user's
// openshift.customResource override is merged onto the GPU operator's
// ClusterPolicy CR — the OpenShift/OLM equivalent of custom Helm values.
func TestOpenShiftInstall_GPUOperator_CustomResourceOverride(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	seedOpenShiftPrereqs(k8s)
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Install.KubernetesDistribution = "openshift"
	cfg.ExternalConnection = nil
	skipPrometheus(cfg)
	cfg.Prerequisites.GPUOperator.OpenShift = &installconfig.DependencyOpenShift{
		CustomResource: map[string]any{
			"driver": map[string]any{"version": "550.127.08"},
			"mig":    map[string]any{"strategy": "single"},
		},
	}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	cp := applyCallByKindName(k8s, "ClusterPolicy", "gpu-cluster-policy")
	require.NotNil(t, cp)
	u, ok := cp.Object.(*unstructured.Unstructured)
	require.True(t, ok)

	driverVersion, _, _ := unstructured.NestedString(u.Object, "spec", "driver", "version")
	assert.Equal(t, "550.127.08", driverVersion, "customResource is merged onto the ClusterPolicy spec")
	migStrategy, _, _ := unstructured.NestedString(u.Object, "spec", "mig", "strategy")
	assert.Equal(t, "single", migStrategy)
}

// TestOpenShiftInstall_GPUOperator_AutomaticApprovalOverride verifies that a
// user can opt the GPU operator back into OLM auto-upgrades by setting
// openshift.installPlanApproval=automatic; the Subscription then carries the
// Automatic strategy instead of the default Manual pin.
func TestOpenShiftInstall_GPUOperator_AutomaticApprovalOverride(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	seedOpenShiftPrereqs(k8s)
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Install.KubernetesDistribution = "openshift"
	cfg.ExternalConnection = nil
	skipPrometheus(cfg)
	cfg.Prerequisites.GPUOperator.OpenShift = &installconfig.DependencyOpenShift{InstallPlanApproval: "automatic"}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	gpuSub := applyCallByKindName(k8s, "Subscription", "gpu-operator-certified")
	assert.Equal(t, "Automatic", subscriptionField(t, gpuSub, "installPlanApproval"), "the automatic override flows into the Subscription")
	assert.Equal(t, gpuCSV, subscriptionField(t, gpuSub, "startingCSV"), "startingCSV still pins the initial version even under Automatic")
}

// TestOpenShiftInstall_GPUOperator_AirgapCatalogOverride verifies the OLM catalog
// destination is configurable: an air-gapped cluster mirrors operators into one
// internal CatalogSource, and both the NFD and GPU subscriptions are pulled from
// it instead of the public catalogs.
func TestOpenShiftInstall_GPUOperator_AirgapCatalogOverride(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	seedOpenShiftPrereqs(k8s)
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Install.KubernetesDistribution = "openshift"
	cfg.ExternalConnection = nil
	skipPrometheus(cfg)
	cfg.OpenShift = &installconfig.OpenShift{
		OLMCatalog: &installconfig.OLMCatalog{Source: "runai-mirror-catalog", SourceNamespace: "openshift-marketplace"},
	}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.Equal(t, "runai-mirror-catalog", subscriptionSource(t, applyCallByKindName(k8s, "Subscription", "nfd")))
	assert.Equal(t, "runai-mirror-catalog", subscriptionSource(t, applyCallByKindName(k8s, "Subscription", "gpu-operator-certified")))
}

// TestOpenShiftInstall_GPUOperator_AlreadyPresent verifies the OperatorHub GPU
// flow is idempotent: when the NFD and GPU operators (and their CRs) are already
// present, every step is detected and skipped rather than re-subscribed.
func TestOpenShiftInstall_GPUOperator_AlreadyPresent(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{Existing: map[client.ObjectKey]bool{
		{Namespace: "openshift-nfd", Name: "nfd-controller-manager"}: true,
		{Namespace: "openshift-nfd", Name: "nfd-instance"}:           true,
		{Namespace: "nvidia-gpu-operator", Name: "gpu-operator"}:     true,
		{Name: "gpu-cluster-policy"}:                                 true,
		// The network operator is enabled too (connectedConfig); mark it present
		// so its OLM flow is also detected and skipped rather than blocking.
		{Namespace: "openshift-sriov-network-operator", Name: "sriov-network-operator"}:            true,
		{Namespace: "nvidia-network-operator", Name: "nvidia-network-operator-controller-manager"}: true,
		{Name: "nic-cluster-policy"}: true,
	}}
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Install.KubernetesDistribution = "openshift"
	cfg.ExternalConnection = nil
	skipPrometheus(cfg)
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.Nil(t, applyCallByKindName(k8s, "Subscription", "nfd"), "present NFD operator is not re-subscribed")
	assert.Nil(t, applyCallByKindName(k8s, "Subscription", "gpu-operator-certified"), "present GPU operator is not re-subscribed")

	evs := readEvents(t, eventsPath)
	for _, step := range []string{
		"install-nfd-operator",
		"create-node-feature-discovery",
		"install-gpu-operator",
		"create-gpu-cluster-policy",
	} {
		skipped := findEvent(evs, "step_skipped", step)
		require.NotNil(t, skipped, "step %s must be skipped", step)
		assert.Equal(t, "already_installed", skipped["skipReason"])
	}
}
