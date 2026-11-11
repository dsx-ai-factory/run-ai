/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package snapshots

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
)

var optIn = map[string]string{optInAnnotation: "true"}

func TestStatus(t *testing.T) {
	gpu := gpuContainer("main")
	ownProfile := corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault}

	tests := []struct {
		name      string
		karta     *kartav1alpha1.Karta
		workload  *unstructured.Unstructured
		wantPhase dromosv1alpha1.PluginPhase
		// wantIssue is part of the warning; empty means no warning at all.
		wantIssue string
	}{
		{
			name:      "not opted in",
			karta:     deploymentKarta(),
			workload:  deployment(t, nil, gpu),
			wantPhase: dromosv1alpha1.PluginPhaseDisabled,
		},
		{
			// The signal is on the workload (REQ-02-1); a pod label alone is not it.
			name:      "only a pod label",
			karta:     deploymentKarta(),
			workload:  withPodLabel(t, deployment(t, nil, gpu), constants.SnapshotOptInLabel, "true"),
			wantPhase: dromosv1alpha1.PluginPhaseDisabled,
		},
		{
			name:      "eligible",
			karta:     deploymentKarta(),
			workload:  deployment(t, optIn, gpu),
			wantPhase: dromosv1alpha1.PluginPhaseReady,
		},
		{
			name:      "two pod templates",
			karta:     pairKarta(),
			workload:  pair(t, optIn, gpu),
			wantPhase: dromosv1alpha1.PluginPhaseDisabled,
			wantIssue: "2 pod templates",
		},
		{
			name:      "pods cannot carry labels",
			karta:     bareKarta(),
			workload:  bare(t, optIn, gpu),
			wantPhase: dromosv1alpha1.PluginPhaseDisabled,
			wantIssue: "cannot carry labels",
		},
		{
			name:      "no GPU container",
			karta:     deploymentKarta(),
			workload:  deployment(t, optIn, corev1.Container{Name: "main", Image: "busybox"}),
			wantPhase: dromosv1alpha1.PluginPhaseDisabled,
			wantIssue: "no container requests",
		},
		{
			name:      "two GPU containers",
			karta:     deploymentKarta(),
			workload:  deployment(t, optIn, gpu, gpuContainer("sidecar")),
			wantPhase: dromosv1alpha1.PluginPhaseDisabled,
			wantIssue: "both request",
		},
		{
			name:      "own seccomp profile",
			karta:     deploymentKarta(),
			workload:  deployment(t, optIn, withSeccomp(gpu, ownProfile)),
			wantPhase: dromosv1alpha1.PluginPhaseDisabled,
			wantIssue: "seccomp",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plugin := newPlugin(kartaStub{karta: tc.karta}, true)
			status := &dromosv1alpha1.PluginStatus{Phase: dromosv1alpha1.PluginPhaseUnknown}
			state := reconcile.New()

			if err := plugin.Status(context.Background(), runnerFor(t, tc.workload), nil, status, state); err != nil {
				t.Fatalf("Status: %v", err)
			}

			if status.Phase != tc.wantPhase {
				t.Errorf("phase = %s, want %s", status.Phase, tc.wantPhase)
			}
			issues := state.PluginIssues(plugin)
			if tc.wantIssue == "" {
				if len(issues) != 0 {
					t.Errorf("issues = %v, want none", issues)
				}
				return
			}
			if len(issues) != 1 {
				t.Fatalf("issues = %v, want one warning", issues)
			}
			if issues[0].IsCritical {
				t.Error("the warning is critical; an ineligible workload's phase must not change")
			}
			if !strings.Contains(issues[0].Error(), tc.wantIssue) {
				t.Errorf("warning %q does not mention %q", issues[0].Error(), tc.wantIssue)
			}
		})
	}
}

// The framework starts a plugin at Unknown, which holds the workload in
// Initializing; a failed lookup must not leave it there.
func TestStatusIsDisabledWhenTheLookupFails(t *testing.T) {
	plugin := newPlugin(kartaStub{err: errors.New("cache not synced")}, true)
	status := &dromosv1alpha1.PluginStatus{Phase: dromosv1alpha1.PluginPhaseUnknown}

	err := plugin.Status(context.Background(), runnerFor(t, deployment(t, optIn, gpuContainer("main"))),
		nil, status, reconcile.New())

	if err == nil {
		t.Fatal("Status succeeded, want the lookup error")
	}
	if status.Phase != dromosv1alpha1.PluginPhaseDisabled {
		t.Errorf("phase = %s, want %s", status.Phase, dromosv1alpha1.PluginPhaseDisabled)
	}
}

// The plugin runs for every workload, so one that never opted in must not get
// an issue from a lookup it doesn't need.
func TestStatusSkipsTheLookupWhenNotOptedIn(t *testing.T) {
	plugin := newPlugin(kartaStub{err: errors.New("cache not synced")}, true)
	status := &dromosv1alpha1.PluginStatus{Phase: dromosv1alpha1.PluginPhaseUnknown}

	err := plugin.Status(context.Background(), runnerFor(t, deployment(t, nil, gpuContainer("main"))),
		nil, status, reconcile.New())

	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.Phase != dromosv1alpha1.PluginPhaseDisabled {
		t.Errorf("phase = %s, want %s", status.Phase, dromosv1alpha1.PluginPhaseDisabled)
	}
}

func TestStatusIsDisabledWithoutSnapshot(t *testing.T) {
	plugin := newPlugin(kartaStub{karta: deploymentKarta()}, false)
	status := &dromosv1alpha1.PluginStatus{Phase: dromosv1alpha1.PluginPhaseUnknown}
	state := reconcile.New()

	err := plugin.Status(context.Background(), runnerFor(t, deployment(t, optIn, gpuContainer("main"))),
		nil, status, state)

	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.Phase != dromosv1alpha1.PluginPhaseDisabled {
		t.Errorf("phase = %s, want %s", status.Phase, dromosv1alpha1.PluginPhaseDisabled)
	}
	if issues := state.PluginIssues(plugin); len(issues) != 0 {
		t.Errorf("issues = %v, want none", issues)
	}
}

type kartaStub struct {
	karta *kartav1alpha1.Karta
	err   error
}

func (s kartaStub) GetKartaForGVK(context.Context, schema.GroupVersionKind) (*kartav1alpha1.Karta, error) {
	return s.karta, s.err
}

func runnerFor(t *testing.T, workload *unstructured.Unstructured) *dromosv1alpha1.WorkloadRunner {
	t.Helper()
	raw, err := json.Marshal(workload.Object)
	if err != nil {
		t.Fatalf("marshal workload: %v", err)
	}
	return &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{Name: "llama", Namespace: "runai-proj"},
		Spec:       dromosv1alpha1.WorkloadRunnerSpec{Template: runtime.RawExtension{Raw: raw}},
	}
}

func deploymentKarta() *kartav1alpha1.Karta {
	return kartaWith(kartav1alpha1.ComponentDefinition{
		Name:           "deployment",
		Kind:           &kartav1alpha1.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		SpecDefinition: &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: ptr.To(".spec.template")},
	})
}

// pairKarta is a kind with a leader and a worker template, like LeaderWorkerSet.
func pairKarta() *kartav1alpha1.Karta {
	return kartaWith(
		kartav1alpha1.ComponentDefinition{
			Name: "pair",
			Kind: &kartav1alpha1.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "Pair"},
		},
		kartav1alpha1.ComponentDefinition{
			Name:           "leader",
			OwnerRef:       ptr.To("pair"),
			SpecDefinition: &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: ptr.To(".spec.leader")},
		},
		kartav1alpha1.ComponentDefinition{
			Name:           "worker",
			OwnerRef:       ptr.To("pair"),
			SpecDefinition: &kartav1alpha1.SpecDefinition{PodTemplateSpecPath: ptr.To(".spec.worker")},
		},
	)
}

// bareKarta is a kind that embeds a pod spec with no metadata.
func bareKarta() *kartav1alpha1.Karta {
	return kartaWith(kartav1alpha1.ComponentDefinition{
		Name:           "bare",
		Kind:           &kartav1alpha1.GroupVersionKind{Group: "example.com", Version: "v1", Kind: "Bare"},
		SpecDefinition: &kartav1alpha1.SpecDefinition{PodSpecPath: ptr.To(".spec.podSpec")},
	})
}

func kartaWith(root kartav1alpha1.ComponentDefinition, children ...kartav1alpha1.ComponentDefinition) *kartav1alpha1.Karta {
	return &kartav1alpha1.Karta{Spec: kartav1alpha1.KartaSpec{
		StructureDefinition: kartav1alpha1.StructureDefinition{RootComponent: root, ChildComponents: children},
	}}
}

// deployment is a Deployment whose own annotations carry the opt-in, and whose
// pod template has a label of its own, so stamping can be seen to keep it.
func deployment(t *testing.T, annotations map[string]string, containers ...corev1.Container) *unstructured.Unstructured {
	t.Helper()
	return toUnstructured(t, &appsv1.Deployment{
		TypeMeta:   metav1.TypeMeta{APIVersion: "apps/v1", Kind: "Deployment"},
		ObjectMeta: metav1.ObjectMeta{Name: "llama", Namespace: "runai-proj", Annotations: annotations},
		Spec: appsv1.DeploymentSpec{Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "llama"}},
			Spec:       corev1.PodSpec{Containers: containers},
		}},
	})
}

func withPodLabel(t *testing.T, workload *unstructured.Unstructured, key, value string) *unstructured.Unstructured {
	t.Helper()
	if err := unstructured.SetNestedField(workload.Object, value, "spec", "template", "metadata", "labels", key); err != nil {
		t.Fatalf("set pod label: %v", err)
	}
	return workload
}

func pair(t *testing.T, annotations map[string]string, containers ...corev1.Container) *unstructured.Unstructured {
	t.Helper()
	template := toMap(t, &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: containers}})
	workload := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "example.com/v1",
		"kind":       "Pair",
		"metadata":   map[string]any{"name": "llama", "namespace": "runai-proj"},
		"spec":       map[string]any{"leader": template, "worker": runtime.DeepCopyJSON(template)},
	}}
	workload.SetAnnotations(annotations)
	return workload
}

func bare(t *testing.T, annotations map[string]string, containers ...corev1.Container) *unstructured.Unstructured {
	t.Helper()
	workload := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "example.com/v1",
		"kind":       "Bare",
		"metadata":   map[string]any{"name": "llama", "namespace": "runai-proj"},
		"spec":       map[string]any{"podSpec": toMap(t, &corev1.PodSpec{Containers: containers})},
	}}
	workload.SetAnnotations(annotations)
	return workload
}

func toUnstructured(t *testing.T, object runtime.Object) *unstructured.Unstructured {
	t.Helper()
	return &unstructured.Unstructured{Object: toMap(t, object)}
}

func toMap(t *testing.T, object any) map[string]any {
	t.Helper()
	converted, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		t.Fatalf("convert %T: %v", object, err)
	}
	return converted
}

func gpuContainer(name string) corev1.Container {
	return corev1.Container{
		Name:  name,
		Image: "vllm/vllm-openai:v0.31.0",
		Resources: corev1.ResourceRequirements{
			Limits: corev1.ResourceList{"nvidia.com/gpu": resource.MustParse("1")},
		},
	}
}

func withSeccomp(container corev1.Container, profile corev1.SeccompProfile) corev1.Container {
	container.SecurityContext = &corev1.SecurityContext{SeccompProfile: &profile}
	return container
}
