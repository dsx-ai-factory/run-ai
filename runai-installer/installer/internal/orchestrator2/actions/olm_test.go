/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-installer/installer/component-tests/harness"
	"github.com/run-ai/runai/runai-installer/installer/internal/orchestrator2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestSubscriptionObject(t *testing.T) {
	catalog := OLMCatalog{Source: "certified-operators", SourceNamespace: "openshift-marketplace"}
	sub := SubscriptionObject("gpu-operator-certified", "nvidia-gpu-operator", "gpu-operator-certified", "v25.3", "gpu-operator-certified.v25.3.0", "Manual", catalog, nil)

	assert.Equal(t, "Subscription", sub.GetKind())
	assert.Equal(t, "nvidia-gpu-operator", sub.GetNamespace())

	get := func(field string) string {
		v, _, _ := unstructured.NestedString(sub.Object, "spec", field)
		return v
	}
	assert.Equal(t, "gpu-operator-certified", get("name"))
	assert.Equal(t, "v25.3", get("channel"))
	assert.Equal(t, "certified-operators", get("source"))
	assert.Equal(t, "openshift-marketplace", get("sourceNamespace"))
	assert.Equal(t, "Manual", get("installPlanApproval"))
	assert.Equal(t, "gpu-operator-certified.v25.3.0", get("startingCSV"))

	_, hasConfig, _ := unstructured.NestedMap(sub.Object, "spec", "config")
	assert.False(t, hasConfig, "no spec.config without subscription env")
}

func TestSubscriptionObject_DefaultsApprovalToAutomatic(t *testing.T) {
	sub := SubscriptionObject("nfd", "openshift-nfd", "nfd", "stable", "", "", OLMCatalog{Source: "redhat-operators", SourceNamespace: "openshift-marketplace"}, nil)

	approval, _, _ := unstructured.NestedString(sub.Object, "spec", "installPlanApproval")
	assert.Equal(t, "Automatic", approval, "an empty approval strategy defaults to OLM's own default")
}

func TestSubscriptionObject_OmitsEmptyStartingCSV(t *testing.T) {
	sub := SubscriptionObject("nfd", "openshift-nfd", "nfd", "stable", "", "Manual", OLMCatalog{Source: "redhat-operators", SourceNamespace: "openshift-marketplace"}, nil)

	_, found, _ := unstructured.NestedString(sub.Object, "spec", "startingCSV")
	assert.False(t, found, "an empty starting CSV must be omitted so OLM tracks the channel head")
}

func TestSubscriptionObject_RendersEnvConfig(t *testing.T) {
	sub := SubscriptionObject("nim-operator", "nvidia-nim-operator", "nim-operator", "v3.1", "", "Automatic", OLMCatalog{Source: "certified-operators", SourceNamespace: "openshift-marketplace"}, map[string]string{"ENABLE_WEBHOOKS": "false"})

	env, found, err := unstructured.NestedSlice(sub.Object, "spec", "config", "env")
	require.NoError(t, err)
	require.True(t, found, "subscription env must render under spec.config.env")
	require.Len(t, env, 1)
	assert.Equal(t, map[string]any{"name": "ENABLE_WEBHOOKS", "value": "false"}, env[0])
}

func TestOperatorGroupObject(t *testing.T) {
	og := OperatorGroupObject("nvidia-gpu-operator-og", "nvidia-gpu-operator", []string{"nvidia-gpu-operator"})

	assert.Equal(t, "OperatorGroup", og.GetKind())
	targets, _, _ := unstructured.NestedStringSlice(og.Object, "spec", "targetNamespaces")
	assert.Equal(t, []string{"nvidia-gpu-operator"}, targets)
}

func TestOperatorGroupObject_AllNamespaces(t *testing.T) {
	og := OperatorGroupObject("cert-manager-operator-og", "cert-manager-operator", nil)

	_, found, _ := unstructured.NestedSlice(og.Object, "spec", "targetNamespaces")
	assert.False(t, found, "an empty target list yields a cluster-wide OperatorGroup")
}

func TestWaitForCSV_SucceedsWhenCSVReady(t *testing.T) {
	k8s := &harness.FakeK8sExecutor{SeededObjects: map[client.ObjectKey]*unstructured.Unstructured{
		{Namespace: "nvidia-gpu-operator", Name: "gpu-operator-certified"}:         subscriptionStatus("gpu-operator-certified.v25.3.0"),
		{Namespace: "nvidia-gpu-operator", Name: "gpu-operator-certified.v25.3.0"}: csvPhase("Succeeded"),
	}}
	deps := orchestrator2.Deps{K8sExecutor: k8s}

	_, err := WaitForCSV{Namespace: "nvidia-gpu-operator", SubscriptionName: "gpu-operator-certified"}.Execute(context.Background(), deps)
	require.NoError(t, err)
}

// approvingExecutor simulates OLM reacting to an InstallPlan approval: once the
// pending plan is patched approved, the subscription reports its installedCSV,
// letting a Manual-approval WaitForCSV complete.
type approvingExecutor struct {
	*harness.FakeK8sExecutor
	subKey       client.ObjectKey
	installedCSV string
}

func (e *approvingExecutor) Patch(ctx context.Context, obj client.Object, patch client.Patch) error {
	if err := e.FakeK8sExecutor.Patch(ctx, obj, patch); err != nil {
		return err
	}
	if sub, ok := e.SeededObjects[e.subKey]; ok {
		_ = unstructured.SetNestedField(sub.Object, e.installedCSV, "status", "installedCSV")
	}
	return nil
}

func TestWaitForCSV_ApprovesPendingManualInstallPlan(t *testing.T) {
	prev := csvReadyPollInterval
	csvReadyPollInterval = time.Millisecond
	defer func() { csvReadyPollInterval = prev }()

	subKey := client.ObjectKey{Namespace: "nvidia-gpu-operator", Name: "gpu-operator-certified"}
	sub := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": olmAPIVersionV1alpha1,
		"kind":       kindSubscription,
		"status": map[string]any{
			"installPlanRef": map[string]any{"name": "install-abc", "namespace": "nvidia-gpu-operator"},
		},
	}}
	ip := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": olmAPIVersionV1alpha1,
		"kind":       kindInstallPlan,
		"metadata":   map[string]any{"name": "install-abc", "namespace": "nvidia-gpu-operator"},
		"spec":       map[string]any{"approved": false, "approval": "Manual"},
	}}
	fake := &harness.FakeK8sExecutor{SeededObjects: map[client.ObjectKey]*unstructured.Unstructured{
		subKey: sub,
		{Namespace: "nvidia-gpu-operator", Name: "install-abc"}:                    ip,
		{Namespace: "nvidia-gpu-operator", Name: "gpu-operator-certified.v25.3.0"}: csvPhase("Succeeded"),
	}}
	k8s := &approvingExecutor{FakeK8sExecutor: fake, subKey: subKey, installedCSV: "gpu-operator-certified.v25.3.0"}
	deps := orchestrator2.Deps{K8sExecutor: k8s}

	_, err := WaitForCSV{Namespace: "nvidia-gpu-operator", SubscriptionName: "gpu-operator-certified"}.Execute(context.Background(), deps)
	require.NoError(t, err)

	patches := fake.CallsByVerb("patch")
	require.Len(t, patches, 1, "the pending Manual InstallPlan is approved exactly once")
	assert.Equal(t, "InstallPlan", patches[0].Kind)
	assert.Equal(t, "install-abc", patches[0].Name)
	assert.Contains(t, string(patches[0].PatchData), `"approved":true`)
}

func TestWaitForCSV_FailsFastOnResolutionFailure(t *testing.T) {
	sub := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": olmAPIVersionV1alpha1,
		"kind":       kindSubscription,
		"status": map[string]any{
			"conditions": []any{map[string]any{
				"type":    "ResolutionFailed",
				"status":  "True",
				"reason":  "ConstraintsNotSatisfiable",
				"message": "no operators found in package leader-worker-set in the catalog",
			}},
		},
	}}
	k8s := &harness.FakeK8sExecutor{SeededObjects: map[client.ObjectKey]*unstructured.Unstructured{
		{Namespace: "openshift-lws-operator", Name: "leader-worker-set"}: sub,
	}}
	deps := orchestrator2.Deps{K8sExecutor: k8s}

	_, err := WaitForCSV{Namespace: "openshift-lws-operator", SubscriptionName: "leader-worker-set"}.Execute(context.Background(), deps)
	require.Error(t, err, "an unresolvable subscription must fail fast, not poll until timeout")
	assert.Contains(t, err.Error(), "cannot be resolved")
	assert.Contains(t, err.Error(), "no operators found in package leader-worker-set")
}

func TestWaitForCSV_BlocksUntilContextWhenNotReady(t *testing.T) {
	k8s := &harness.FakeK8sExecutor{}
	deps := orchestrator2.Deps{K8sExecutor: k8s}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := WaitForCSV{Namespace: "nvidia-gpu-operator", SubscriptionName: "gpu-operator-certified"}.Execute(ctx, deps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "CSV to succeed")
}

func TestMergeCRSpec_DeepMergesAndReplaces(t *testing.T) {
	base := map[string]any{
		"driver":  map[string]any{"enabled": true, "version": "default"},
		"toolkit": map[string]any{"enabled": true},
	}
	override := map[string]any{
		"driver": map[string]any{"version": "550.127.08"},
		"mig":    map[string]any{"strategy": "single"},
	}

	got := MergeCRSpec(base, override)

	driver := got["driver"].(map[string]any)
	assert.Equal(t, "550.127.08", driver["version"], "override replaces a leaf")
	assert.Equal(t, true, driver["enabled"], "sibling base keys are preserved by the deep merge")
	assert.Equal(t, map[string]any{"enabled": true}, got["toolkit"], "untouched base subtrees survive")
	assert.Equal(t, map[string]any{"strategy": "single"}, got["mig"], "new override subtrees are added")
}

func TestMergeCRSpec_DoesNotMutateInputs(t *testing.T) {
	base := map[string]any{"driver": map[string]any{"version": "default"}}
	override := map[string]any{"driver": map[string]any{"version": "550"}}

	_ = MergeCRSpec(base, override)

	require.Equal(t, "default", base["driver"].(map[string]any)["version"], "base must not be mutated")
	require.Equal(t, "550", override["driver"].(map[string]any)["version"], "override must not be mutated")
}

func TestApplyCSVExampleCR_AppliesDefaultMergedWithOverride(t *testing.T) {
	alm := `[{"apiVersion":"nvidia.com/v1","kind":"ClusterPolicy","metadata":{"name":"gpu-cluster-policy"},"spec":{"operator":{"defaultRuntime":"crio"},"driver":{"enabled":true}}}]`
	k8s := &harness.FakeK8sExecutor{SeededObjects: map[client.ObjectKey]*unstructured.Unstructured{
		{Namespace: "nvidia-gpu-operator", Name: "gpu-operator-certified"}:         subscriptionStatus("gpu-operator-certified.v25.3.0"),
		{Namespace: "nvidia-gpu-operator", Name: "gpu-operator-certified.v25.3.0"}: csvWithExamples(alm),
	}}
	deps := orchestrator2.Deps{K8sExecutor: k8s}

	act := ApplyCSVExampleCR{
		CSVNamespace:     "nvidia-gpu-operator",
		SubscriptionName: "gpu-operator-certified",
		CRAPIVersion:     "nvidia.com/v1",
		CRKind:           "ClusterPolicy",
		CRName:           "gpu-cluster-policy",
		SpecOverride:     map[string]any{"driver": map[string]any{"version": "550.127.08"}},
	}
	_, err := act.Execute(context.Background(), deps)
	require.NoError(t, err)

	applies := k8s.CallsByVerb("apply")
	require.Len(t, applies, 1)
	u := applies[0].Object.(*unstructured.Unstructured)
	runtime, _, _ := unstructured.NestedString(u.Object, "spec", "operator", "defaultRuntime")
	assert.Equal(t, "crio", runtime, "the CSV default spec is the base")
	version, _, _ := unstructured.NestedString(u.Object, "spec", "driver", "version")
	assert.Equal(t, "550.127.08", version, "the user override is merged on top")
}

func TestApplyCSVExampleCR_ErrorsWhenExampleMissing(t *testing.T) {
	k8s := &harness.FakeK8sExecutor{SeededObjects: map[client.ObjectKey]*unstructured.Unstructured{
		{Namespace: "nvidia-gpu-operator", Name: "gpu-operator-certified"}:         subscriptionStatus("gpu-operator-certified.v25.3.0"),
		{Namespace: "nvidia-gpu-operator", Name: "gpu-operator-certified.v25.3.0"}: csvWithExamples(`[]`),
	}}
	deps := orchestrator2.Deps{K8sExecutor: k8s}

	_, err := ApplyCSVExampleCR{
		CSVNamespace:     "nvidia-gpu-operator",
		SubscriptionName: "gpu-operator-certified",
		CRKind:           "ClusterPolicy",
		CRName:           "gpu-cluster-policy",
	}.Execute(context.Background(), deps)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no ClusterPolicy example")
}

func csvWithExamples(alm string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": olmAPIVersionV1alpha1,
		"kind":       kindClusterServiceVersion,
		"metadata":   map[string]any{"annotations": map[string]any{almExamplesAnnotation: alm}},
		"status":     map[string]any{"phase": "Succeeded"},
	}}
}

func subscriptionStatus(installedCSV string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": olmAPIVersionV1alpha1,
		"kind":       kindSubscription,
		"status":     map[string]any{"installedCSV": installedCSV},
	}}
}

func csvPhase(phase string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": olmAPIVersionV1alpha1,
		"kind":       kindClusterServiceVersion,
		"status":     map[string]any{"phase": phase},
	}}
}
