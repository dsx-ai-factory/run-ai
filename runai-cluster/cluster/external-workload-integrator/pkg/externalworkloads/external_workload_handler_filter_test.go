/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	"context"
	"reflect"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/supportedhandlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type gvkRecordingFilterHandler struct {
	filterGVK          schema.GroupVersionKind
	secondaryFilterGVK schema.GroupVersionKind
}

func (h *gvkRecordingFilterHandler) Parse(context.Context, client.Object, *v1alpha1.ExternalWorkload) error {
	return nil
}

func (h *gvkRecordingFilterHandler) GetObject(context.Context, string, string) (client.Object, error) {
	return nil, nil
}

func (h *gvkRecordingFilterHandler) Filter(object client.Object) bool {
	h.filterGVK = object.GetObjectKind().GroupVersionKind()
	return true
}

func (h *gvkRecordingFilterHandler) SecondaryFilter(object client.Object) bool {
	h.secondaryFilterGVK = object.GetObjectKind().GroupVersionKind()
	return true
}

func TestEventHandlerFilterDoesNotMutateUnstructuredObject(t *testing.T) {
	const namespaceName = "runai-project-test"

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add core types to scheme: %v", err)
	}

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
			Labels: map[string]string{
				constants.RunaiQueueLabel: "test-project",
			},
		},
	}
	cachedClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(namespace).Build()

	gvk := schema.GroupVersionKind{
		Group:   "leaderworkerset.x-k8s.io",
		Version: "v1",
		Kind:    "LeaderWorkerSet",
	}
	filterHandler := &gvkRecordingFilterHandler{}
	handler := &ExternalWorkloadHandler{
		cachedClient: cachedClient,
		supportedGVKsHandlers: map[schema.GroupVersionKind]supportedhandlers.TopOwnerHandler{
			gvk: filterHandler,
		},
	}
	workload := &unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{
			"name":      "test-lws",
			"namespace": namespaceName,
		},
		"spec": map[string]any{"replicas": int64(1)},
	}}
	before := workload.DeepCopy()

	if !handler.EventHandlerFilter(context.Background(), workload, gvk) {
		t.Fatal("expected LeaderWorkerSet to pass the event filter")
	}
	if !reflect.DeepEqual(before.Object, workload.Object) {
		t.Fatalf("event filter mutated informer-owned object: before=%v after=%v", before.Object, workload.Object)
	}
	if filterHandler.filterGVK != gvk {
		t.Fatalf("scheduler filter saw GVK %s, expected %s", filterHandler.filterGVK, gvk)
	}
	if filterHandler.secondaryFilterGVK != gvk {
		t.Fatalf("secondary filter saw GVK %s, expected %s", filterHandler.secondaryFilterGVK, gvk)
	}
}
