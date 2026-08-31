/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datavolume

// Direct unit coverage for the kai.resources leaves of the datavolume mapper/filter.

import (
	"context"
	"testing"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func kaiTestScheme(t *testing.T) *runtime.Scheme {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := v1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add core/v1 to scheme: %v", err)
	}
	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add run/v1alpha1 to scheme: %v", err)
	}
	if err := kaiorg.AddToScheme(scheme); err != nil {
		t.Fatalf("failed to add kai.resources to scheme: %v", err)
	}
	return scheme
}

// MapKaiProjectToDatavolumeEvent should resolve the kai Project's namespace and enqueue every
// datavolume shared (by department = Spec.Parent) with that project.
func TestMapKaiProjectToDatavolumeEvent(t *testing.T) {
	scheme := kaiTestScheme(t)

	project := &kaiorg.Project{
		ObjectMeta: metav1.ObjectMeta{Name: "proj-a"},
		Spec:       kaiorg.ProjectSpec{Parent: "dep-a"},
		Status:     kaiorg.ProjectStatus{Namespace: "ns-a"},
	}
	namespace := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns-a"}}
	sharedDv := &v1alpha1.DataVolume{
		ObjectMeta: metav1.ObjectMeta{Name: "dv-shared-with-dep"},
		Spec: v1alpha1.DataVolumeSpec{
			SharedScopes: []v1alpha1.DataVolumeSharedScope{
				{ScopeType: v1alpha1.DepartmentScope, ScopeName: "dep-a"},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(project, namespace, sharedDv).Build()

	reconciler := NewDatavolumeReconciler(client)
	requests := reconciler.MapProjectToDatavolumeEvent(context.Background(), project)

	if len(requests) != 1 {
		t.Fatalf("expected 1 reconcile request, got %d: %v", len(requests), requests)
	}
	if requests[0].Name != "dv-shared-with-dep" {
		t.Errorf("expected request for dv-shared-with-dep, got %q", requests[0].Name)
	}
}

func TestMapKaiProjectToDatavolumeEvent_WrongType(t *testing.T) {
	client := fake.NewClientBuilder().WithScheme(kaiTestScheme(t)).Build()
	reconciler := NewDatavolumeReconciler(client)

	// A legacy-typed object must not be handled by the kai mapper.
	requests := reconciler.MapProjectToDatavolumeEvent(context.Background(),
		&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "not-a-kai-project"}})
	if len(requests) != 0 {
		t.Fatalf("expected 0 requests for a non-kai object, got %d", len(requests))
	}
}

func TestFilterCreateEvents_KaiProject(t *testing.T) {
	withNamespace := &kaiorg.Project{Status: kaiorg.ProjectStatus{Namespace: "ns-a"}}
	if !filterCreateEvents(withNamespace) {
		t.Error("expected create event for kai project with a namespace to pass the filter")
	}

	withoutNamespace := &kaiorg.Project{}
	if filterCreateEvents(withoutNamespace) {
		t.Error("expected create event for kai project without a namespace to be filtered out")
	}
}

func TestFilterUpdateEvents_KaiProject(t *testing.T) {
	oldProj := &kaiorg.Project{Status: kaiorg.ProjectStatus{Namespace: ""}}
	newProj := &kaiorg.Project{Status: kaiorg.ProjectStatus{Namespace: "ns-a"}}
	if !filterUpdateEvents(oldProj, newProj) {
		t.Error("expected update event with a namespace status change to pass the filter")
	}

	unchanged := &kaiorg.Project{Status: kaiorg.ProjectStatus{Namespace: "ns-a"}}
	if filterUpdateEvents(unchanged, unchanged.DeepCopy()) {
		t.Error("expected update event with no namespace change to be filtered out")
	}
}
