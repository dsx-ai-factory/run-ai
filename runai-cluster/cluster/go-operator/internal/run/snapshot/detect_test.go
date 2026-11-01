/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshot

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	snapshotapi "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/snapshot"
)

func scheme(t *testing.T) *runtime.Scheme {
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

func config() *runv1.RunaiConfig {
	return &runv1.RunaiConfig{
		Spec: runv1.RunaiConfigSpec{
			Snapshot: &snapshotapi.Snapshot{
				Namespace:   ptr.To("snapshot"),
				ReleaseName: ptr.To("snapshot"),
			},
		},
	}
}

func crd() *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: SnapshotJobCRDName},
	}
}

func operator(availableReplicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "snapshot-operator", Namespace: "snapshot"},
		Status:     appsv1.DeploymentStatus{AvailableReplicas: availableReplicas},
	}
}

func agent(selector map[string]string) *appsv1.DaemonSet {
	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: "snapshot-agent", Namespace: "snapshot"},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{NodeSelector: selector},
			},
		},
	}
}

func build(t *testing.T, objects ...client.Object) client.Reader {
	t.Helper()
	return fake.NewClientBuilder().WithScheme(scheme(t)).WithObjects(objects...).Build()
}

// A cluster with no Snapshot at all is the case that must keep behaving exactly
// as it does today, so it gets its own test rather than a table row.
func TestDetectReportsUnavailableWithoutSnapshot(t *testing.T) {
	status := Detect(context.Background(), build(t), config())

	if status.Available {
		t.Fatal("snapshot reported available on a cluster with nothing installed")
	}
	if status.CRDInstalled {
		t.Fatal("CRD reported installed when it is absent")
	}
	if status.Reason == "" {
		t.Fatal("unavailable status carries no reason")
	}
	if status.AgentNodeSelector != nil {
		t.Fatalf("node selector read from an unavailable install: %v", status.AgentNodeSelector)
	}
}

func TestDetect(t *testing.T) {
	tests := []struct {
		name          string
		objects       []client.Object
		wantAvailable bool
		wantCRD       bool
		wantOperator  bool
		wantSelector  map[string]string
	}{
		{
			name:    "CRD present but operator missing",
			objects: []client.Object{crd()},
			wantCRD: true,
		},
		{
			name:    "operator present but has no ready replica",
			objects: []client.Object{crd(), operator(0)},
			wantCRD: true,
		},
		{
			// The operator being up without the CRD should not read as a
			// usable install: Run:ai would create SnapshotJobs the API server
			// rejects.
			name:         "operator ready but CRD missing",
			objects:      []client.Object{operator(1)},
			wantOperator: true,
		},
		{
			name:          "both halves present, agent carries a selector",
			objects:       []client.Object{crd(), operator(1), agent(map[string]string{"nvidia.com/gpu.present": "true"})},
			wantAvailable: true,
			wantCRD:       true,
			wantOperator:  true,
			wantSelector:  map[string]string{"nvidia.com/gpu.present": "true"},
		},
		{
			// An agent DaemonSet with no selector runs everywhere, which is a
			// legitimate configuration and not a detection failure.
			name:          "both halves present, agent has no selector",
			objects:       []client.Object{crd(), operator(1), agent(nil)},
			wantAvailable: true,
			wantCRD:       true,
			wantOperator:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status := Detect(context.Background(), build(t, tc.objects...), config())

			if status.Available != tc.wantAvailable {
				t.Fatalf("Available = %v, want %v (reason: %q)", status.Available, tc.wantAvailable, status.Reason)
			}
			if status.CRDInstalled != tc.wantCRD {
				t.Errorf("CRDInstalled = %v, want %v", status.CRDInstalled, tc.wantCRD)
			}
			if status.OperatorReady != tc.wantOperator {
				t.Errorf("OperatorReady = %v, want %v", status.OperatorReady, tc.wantOperator)
			}
			if !tc.wantAvailable && status.Reason == "" {
				t.Error("unavailable status carries no reason")
			}
			if len(status.AgentNodeSelector) != len(tc.wantSelector) {
				t.Fatalf("AgentNodeSelector = %v, want %v", status.AgentNodeSelector, tc.wantSelector)
			}
			for key, want := range tc.wantSelector {
				if status.AgentNodeSelector[key] != want {
					t.Errorf("AgentNodeSelector[%q] = %q, want %q", key, status.AgentNodeSelector[key], want)
				}
			}
		})
	}
}

// A nil client is what the operand passes in unit tests and before the manager
// is started; it must not panic.
func TestDetectToleratesNilClient(t *testing.T) {
	status := Detect(context.Background(), nil, config())

	if status.Available {
		t.Fatal("snapshot reported available with no client")
	}
	if status.Reason == "" {
		t.Fatal("unavailable status carries no reason")
	}
}

func TestFormatNodeSelector(t *testing.T) {
	tests := []struct {
		name     string
		selector map[string]string
		want     string
	}{
		{name: "empty", selector: nil, want: ""},
		{name: "single", selector: map[string]string{"a": "1"}, want: "a=1"},
		{
			// Sorted, so an unchanged cluster never produces a changed
			// deployment and an endless reconcile loop.
			name:     "sorted by key",
			selector: map[string]string{"c": "3", "a": "1", "b": "2"},
			want:     "a=1,b=2,c=3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatNodeSelector(tc.selector); got != tc.want {
				t.Fatalf("FormatNodeSelector() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestObjectNamesFollowTheChartFullnameRule(t *testing.T) {
	tests := []struct {
		name         string
		release      string
		wantOperator string
		wantAgent    string
	}{
		{
			// The chart does not prefix when the release already contains the
			// chart name, so neither do we.
			name:         "release contains the chart name",
			release:      "snapshot",
			wantOperator: "snapshot-operator",
			wantAgent:    "snapshot-agent",
		},
		{
			name:         "release does not contain the chart name",
			release:      "nvidia",
			wantOperator: "nvidia-snapshot-operator",
			wantAgent:    "nvidia-snapshot-agent",
		},
		{
			name:         "release embeds the chart name",
			release:      "my-snapshot-prod",
			wantOperator: "my-snapshot-prod-operator",
			wantAgent:    "my-snapshot-prod-agent",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &snapshotapi.Snapshot{ReleaseName: ptr.To(tc.release)}
			if got := cfg.OperatorDeploymentName(); got != tc.wantOperator {
				t.Errorf("OperatorDeploymentName() = %q, want %q", got, tc.wantOperator)
			}
			if got := cfg.AgentDaemonSetName(); got != tc.wantAgent {
				t.Errorf("AgentDaemonSetName() = %q, want %q", got, tc.wantAgent)
			}
		})
	}
}

// The gate reads from the operator's environment. Anything other than an
// explicit true leaves it shut, including a value the chart never writes.
func TestFeatureGateEnabled(t *testing.T) {
	tests := []struct {
		name  string
		value string
		set   bool
		want  bool
	}{
		{name: "unset defaults shut", want: false},
		{name: "true opens it", value: "true", set: true, want: true},
		{name: "1 opens it", value: "1", set: true, want: true},
		{name: "false leaves it shut", value: "false", set: true, want: false},
		{name: "empty leaves it shut", value: "", set: true, want: false},
		{name: "garbage leaves it shut", value: "yes-please", set: true, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.set {
				t.Setenv(FeatureGateEnvVar, tc.value)
			}
			if got := FeatureGateEnabled(); got != tc.want {
				t.Fatalf("FeatureGateEnabled() = %v, want %v", got, tc.want)
			}
		})
	}
}
