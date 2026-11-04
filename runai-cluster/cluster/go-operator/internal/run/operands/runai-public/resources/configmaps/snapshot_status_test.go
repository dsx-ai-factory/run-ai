/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configmaps

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	discoveryfake "k8s.io/client-go/discovery/fake"
	clienttesting "k8s.io/client-go/testing"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	kc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client"
	kcfake "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/kube-client/fake"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	snapshotapi "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/snapshot"
)

func snapshotScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	s := runtime.NewScheme()
	if err := appsv1.AddToScheme(s); err != nil {
		t.Fatalf("register apps: %v", err)
	}
	if err := apiextensionsv1.AddToScheme(s); err != nil {
		t.Fatalf("register apiextensions: %v", err)
	}
	return s
}

func snapshotConfig() *runv1.RunaiConfig {
	return &runv1.RunaiConfig{
		Spec: runv1.RunaiConfigSpec{
			Snapshot: &snapshotapi.Snapshot{
				Namespace:   ptr.To("snapshot"),
				ReleaseName: ptr.To("snapshot"),
			},
		},
	}
}

func snapshotCRD() *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "snapshotjobs.nvidia.com"},
	}
}

func snapshotOperator(availableReplicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "snapshot-operator", Namespace: "snapshot"},
		Status:     appsv1.DeploymentStatus{AvailableReplicas: availableReplicas},
	}
}

func snapshotAgent(selector map[string]string) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: "snapshot-agent", Namespace: "snapshot"},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{NodeSelector: selector},
			},
		},
	}
}

// buildSnapshotClients wraps a controller-runtime fake in the kube-client fake,
// so the reader can also hand out a discovery client. A CustomResourceDefinition
// among the objects stands for "Snapshot's API is served": detection asks
// discovery, not the object store, so the CRD becomes a discovery entry rather
// than being stored.
func buildSnapshotClients(t *testing.T, objects ...client.Object) (kc.HybridClient, discovery.DiscoveryInterface) {
	t.Helper()

	stored := make([]client.Object, 0, len(objects))
	served := false
	for _, object := range objects {
		if _, isCRD := object.(*apiextensionsv1.CustomResourceDefinition); isCRD {
			served = true
			continue
		}
		stored = append(stored, object)
	}

	discoveryClient := snapshotDiscovery(served)
	cached := ctrlfake.NewClientBuilder().WithScheme(snapshotScheme(t)).WithObjects(stored...).Build()
	return kcfake.New(cached, cached, discoveryClient), discoveryClient
}

func snapshotDiscovery(served bool) discovery.DiscoveryInterface {
	discoveryClient := &discoveryfake.FakeDiscovery{Fake: &clienttesting.Fake{}}
	if served {
		discoveryClient.Resources = []*metav1.APIResourceList{{
			GroupVersion: runv1.SnapshotGroupVersion,
			APIResources: []metav1.APIResource{{Name: "snapshotjobs", Kind: runv1.SnapshotJobKind}},
		}}
	}
	return discoveryClient
}

func componentOf(status runv1.PublicStatus, name string) runv1.PublicStatus {
	return status.Components[name]
}

func available(status runv1.PublicStatus) bool {
	return ptr.Deref((*bool)(status.Available), false)
}

// A cluster with no Snapshot at all is the case that must keep behaving exactly
// as it does today, so it gets its own test rather than a table row.
func TestSnapshotStatusWithoutSnapshot(t *testing.T) {
	kbClient, discoveryClient := buildSnapshotClients(t)

	status := getSnapshotStatus(context.Background(), discoveryClient, kbClient, snapshotConfig())

	if available(status) {
		t.Fatal("snapshot reported available on a cluster with nothing installed")
	}
	if available(componentOf(status, runv1.SnapshotComponentCRD)) {
		t.Fatal("CRD reported installed when it is absent")
	}
	if status.Reason == "" {
		t.Fatal("unavailable status carries no reason")
	}
}

func TestSnapshotStatus(t *testing.T) {
	tests := []struct {
		name          string
		objects       []client.Object
		wantAvailable bool
		wantCRD       bool
		wantOperator  bool
	}{
		{
			name:    "crd only",
			objects: []client.Object{snapshotCRD()},
			wantCRD: true,
		},
		{
			name:    "operator with no available replica",
			objects: []client.Object{snapshotCRD(), snapshotOperator(0)},
			wantCRD: true,
		},
		{
			// The operator cannot be useful without the CRDs, so a missing CRD
			// is the reason reported even though the deployment itself is
			// ready - each component still reports its own truth.
			name:         "operator ready but no crd",
			objects:      []client.Object{snapshotOperator(1)},
			wantOperator: true,
		},
		{
			name:          "fully installed",
			objects:       []client.Object{snapshotCRD(), snapshotOperator(1), snapshotAgent(nil)},
			wantAvailable: true,
			wantCRD:       true,
			wantOperator:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kbClient, discoveryClient := buildSnapshotClients(t, tc.objects...)

			status := getSnapshotStatus(context.Background(), discoveryClient, kbClient, snapshotConfig())

			if available(status) != tc.wantAvailable {
				t.Fatalf("Available = %v, want %v (reason: %q)", available(status), tc.wantAvailable, status.Reason)
			}
			if got := available(componentOf(status, runv1.SnapshotComponentCRD)); got != tc.wantCRD {
				t.Errorf("%s available = %v, want %v", runv1.SnapshotComponentCRD, got, tc.wantCRD)
			}
			if got := available(componentOf(status, runv1.SnapshotComponentOperator)); got != tc.wantOperator {
				t.Errorf("%s available = %v, want %v", runv1.SnapshotComponentOperator, got, tc.wantOperator)
			}
			if !tc.wantAvailable && status.Reason == "" {
				t.Error("unavailable status carries no reason")
			}
			if !tc.wantCRD && componentOf(status, runv1.SnapshotComponentCRD).Reason == "" {
				t.Error("absent CRD reported with no reason")
			}
		})
	}
}

func TestSnapshotsAvailable(t *testing.T) {
	tests := []struct {
		name           string
		crdReason      string
		operatorReason string
		wantAvailable  bool
		wantReason     string
	}{
		{name: "both fine", wantAvailable: true},
		{
			// A missing CRD wins: without it the operator's state says nothing.
			name:           "crd reason wins",
			crdReason:      "no crd",
			operatorReason: "no replicas",
			wantReason:     "no crd",
		},
		{name: "operator only", operatorReason: "no replicas", wantReason: "no replicas"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotAvailable, gotReason := snapshotsAvailable(tc.crdReason, tc.operatorReason)
			if gotAvailable != tc.wantAvailable {
				t.Errorf("available = %v, want %v", gotAvailable, tc.wantAvailable)
			}
			if gotReason != tc.wantReason {
				t.Errorf("reason = %q, want %q", gotReason, tc.wantReason)
			}
		})
	}
}

func TestSnapshotsOperatorReady(t *testing.T) {
	tests := []struct {
		name    string
		objects []client.Object
		want    bool
	}{
		{name: "deployment missing", want: false},
		{name: "no available replica", objects: []client.Object{snapshotOperator(0)}, want: false},
		{name: "ready", objects: []client.Object{snapshotOperator(1)}, want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kbClient, _ := buildSnapshotClients(t, tc.objects...)

			got, reason := snapshotsOperatorReady(context.Background(), kbClient, snapshotConfig())
			if got != tc.want {
				t.Fatalf("ready = %v, want %v (reason: %q)", got, tc.want, reason)
			}
			if !got && reason == "" {
				t.Error("not ready, but no reason given")
			}
		})
	}
}

// The agent's node selector travels to the anyworkload-controller, which can
// only schedule capture and restore pods where the agent runs.
func TestSnapshotAgentNodeSelectorReachesTheConfiguration(t *testing.T) {
	kbClient, _ := buildSnapshotClients(t, snapshotCRD(), snapshotOperator(1),
		snapshotAgent(map[string]string{"nvidia.com/gpu.present": "true"}))

	got := snapshotsAgentNodeSelector(context.Background(), kbClient, snapshotConfig())

	const want = "nvidia.com/gpu.present=true"
	if got != want {
		t.Errorf("snapshotsAgentNodeSelector() = %q, want %q", got, want)
	}
}

// Keys are sorted so an unchanged cluster renders an unchanged value: the
// anyworkload-controller takes this through its environment, so a reordering
// would roll the deployment for no reason.
func TestSnapshotAgentNodeSelectorIsSorted(t *testing.T) {
	kbClient, _ := buildSnapshotClients(t, snapshotAgent(map[string]string{
		"zone": "b", "nvidia.com/gpu.present": "true", "arch": "amd64",
	}))

	got := snapshotsAgentNodeSelector(context.Background(), kbClient, snapshotConfig())

	const want = "arch=amd64,nvidia.com/gpu.present=true,zone=b"
	if got != want {
		t.Errorf("snapshotsAgentNodeSelector() = %q, want %q", got, want)
	}
}

// No agent DaemonSet, or one with no selector, means no constraint.
func TestSnapshotAgentNodeSelectorEmpty(t *testing.T) {
	cases := map[string][]client.Object{
		"no daemonset":   nil,
		"no selector":    {snapshotAgent(nil)},
		"empty selector": {snapshotAgent(map[string]string{})},
	}

	for name, objects := range cases {
		t.Run(name, func(t *testing.T) {
			kbClient, _ := buildSnapshotClients(t, objects...)
			if got := snapshotsAgentNodeSelector(context.Background(), kbClient, snapshotConfig()); got != "" {
				t.Errorf("snapshotsAgentNodeSelector() = %q, want empty", got)
			}
		})
	}
}
