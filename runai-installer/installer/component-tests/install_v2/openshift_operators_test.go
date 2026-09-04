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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	sriovCSV   = "sriov-network-operator.v1.5.0"
	networkCSV = "nvidia-network-operator.v25.10.0"

	nicALMExamples = `[{"apiVersion":"mellanox.com/v1alpha1","kind":"NicClusterPolicy","metadata":{"name":"nic-cluster-policy"},"spec":{"ofedDriver":{"image":"doca-driver"}}}]`
)

// seedOpenShiftNetworkOperator resolves the OperatorHub network flow on the fake
// cluster: NFD + SR-IOV (no CR) + the network operator with its NicClusterPolicy.
func seedOpenShiftNetworkOperator(k8s *harness.FakeK8sExecutor) {
	seedNFD(k8s)
	k8s.SeededObjects[client.ObjectKey{Namespace: "openshift-sriov-network-operator", Name: "sriov-network-operator"}] = olmSubscriptionWithCSV(sriovCSV)
	k8s.SeededObjects[client.ObjectKey{Namespace: "openshift-sriov-network-operator", Name: sriovCSV}] = olmCSVSucceeded()
	k8s.SeededObjects[client.ObjectKey{Namespace: "nvidia-network-operator", Name: "nvidia-network-operator"}] = olmSubscriptionWithCSV(networkCSV)
	k8s.SeededObjects[client.ObjectKey{Namespace: "nvidia-network-operator", Name: networkCSV}] = olmCSVWithExamples(nicALMExamples)
	k8s.SeededObjects[client.ObjectKey{Name: "nicclusterpolicies.mellanox.com"}] = establishedCRDObject("nicclusterpolicies.mellanox.com")
}

// seedOpenShiftPrereqs seeds every OpenShift prerequisite operator the default
// connectedConfig enables (GPU + network), so a full OpenShift install resolves.
func seedOpenShiftPrereqs(k8s *harness.FakeK8sExecutor) {
	seedOpenShiftGPUOperator(k8s)
	seedOpenShiftNetworkOperator(k8s)
}

// TestOpenShiftInstall_NetworkOperatorViaOperatorHub verifies the network operator
// installs through OperatorHub on OpenShift: NFD and the SR-IOV operator (its
// dependencies) plus the certified network operator and its NicClusterPolicy.
func TestOpenShiftInstall_NetworkOperatorViaOperatorHub(t *testing.T) {
	helm := harness.NewFakeHelmWithReleases()
	k8s := &harness.FakeK8sExecutor{}
	seedOpenShiftNetworkOperator(k8s)
	_, provider := harness.NewFakeRunAIClientProvider(helm)
	deps := orchestrator2.Deps{K8sExecutor: k8s, HelmInstaller: helm, RunAIClientProvider: provider, ClusterProbe: okProbe()}

	eventsPath := filepath.Join(t.TempDir(), "events.ndjson")
	cfg, opts := connectedConfig(t, eventsPath, false)
	cfg.Install.KubernetesDistribution = "openshift"
	cfg.ExternalConnection = nil
	cfg.Prerequisites = &installconfig.Prerequisites{NetworkOperator: &installconfig.Dependency{
		Mode: "install",
		// Operator-pod settings (with no custom resource) flow through the OLM
		// Subscription's spec.config.
		OpenShift: &installconfig.DependencyOpenShift{
			SubscriptionConfig: &installconfig.SubscriptionConfig{Env: map[string]string{"LOG_LEVEL": "debug"}},
		},
	}}
	term, _, _ := newTestTerminal()

	require.NoError(t, specrun.Execute(context.Background(), cfg.ToSpec(opts), deps, term, execOpts()))

	assert.Empty(t, helm.CallsByRelease("network-operator"), "on OpenShift the network operator must not install via Helm")

	sriovSub := applyCallByKindName(k8s, "Subscription", "sriov-network-operator")
	require.NotNil(t, sriovSub, "SR-IOV operator is subscribed via OLM")
	assert.Equal(t, "redhat-operators", subscriptionSource(t, sriovSub))
	assert.Equal(t, "stable", subscriptionChannel(t, sriovSub), "SR-IOV channel resolves from the compat matrix")

	netSub := applyCallByKindName(k8s, "Subscription", "nvidia-network-operator")
	require.NotNil(t, netSub, "network operator is subscribed via OLM")
	assert.Equal(t, "certified-operators", subscriptionSource(t, netSub))
	assert.Equal(t, "nvidia-network-operator", netSub.Namespace)
	assert.Equal(t, "v25.10", subscriptionChannel(t, netSub), "network channel derives from the matrix version v25.10.0")
	assert.Equal(t, networkCSV, subscriptionField(t, netSub, "startingCSV"), "network subscription pins the matrix version via startingCSV")
	assert.Equal(t, "Manual", subscriptionField(t, netSub, "installPlanApproval"), "default approval strategy is Manual")
	assert.Equal(t, "Manual", subscriptionField(t, sriovSub, "installPlanApproval"), "the flow's approval strategy applies to the SR-IOV dependency too")

	netSubObj := netSub.Object.(*unstructured.Unstructured)
	env, _, _ := unstructured.NestedSlice(netSubObj.Object, "spec", "config", "env")
	require.Len(t, env, 1, "subscriptionConfig env flows into the network operator Subscription")
	assert.Equal(t, map[string]any{"name": "LOG_LEVEL", "value": "debug"}, env[0])

	require.NotNil(t, applyCallByKindName(k8s, "Subscription", "nfd"), "NFD is installed as a network dependency")
	require.NotNil(t, applyCallByKindName(k8s, "NicClusterPolicy", "nic-cluster-policy"), "the NicClusterPolicy CR is applied")

	evs := readEvents(t, eventsPath)
	for _, step := range []string{
		"install-nfd-operator",
		"install-sriov-network-operator",
		"install-network-operator",
		"create-nic-cluster-policy",
	} {
		require.NotNil(t, findEvent(evs, "step_succeeded", step), "step %s must succeed", step)
	}

	summary := prereqSummaryEvent(t, evs)
	assert.Equal(t, "installed", summary["outcome"])
	assert.EqualValues(t, 1, summary["installed"], "only the network operator prerequisite is enabled")
}
