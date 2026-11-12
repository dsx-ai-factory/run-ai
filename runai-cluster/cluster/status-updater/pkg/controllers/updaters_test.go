/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/patcher"
)

// TestUpdatePodWritesStatusOnGpuCalcError checks that a GPU resource calculation failure still
// updates the pod status annotation, leaves the GPU annotations unset and returns a gpuCalcError.
func TestUpdatePodWritesStatusOnGpuCalcError(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod",
			Namespace: "ns",
			Annotations: map[string]string{
				"received-resource-type":        "UnknownType",
				patcher.PodStatusAnnotationName: "Pending",
			},
		},
		Status: v1.PodStatus{Phase: v1.PodRunning},
	}
	kubeClient := fake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).WithObjects(pod.DeepCopy()).Build()

	err := updatePod(context.Background(), pod.DeepCopy(), kubeClient)

	var gpuErr *gpuCalcError
	require.True(t, errors.As(err, &gpuErr), "expected a gpuCalcError, got %v", err)

	updated := &v1.Pod{}
	require.NoError(t, kubeClient.Get(context.Background(), client.ObjectKeyFromObject(pod), updated))
	assert.Equal(t, "Running", updated.Annotations[patcher.PodStatusAnnotationName])
	assert.NotContains(t, updated.Annotations, patcher.PodAllocatedGPUsAnnotationName)
	assert.NotContains(t, updated.Annotations, patcher.PodAllocatedMigGPUsAnnotationName)
}

// TestUpdatePodPatchErrorIsNotContinuable checks that when both the GPU calculation and the status patch
// fail, updatePod returns the patch error without the gpuCalcError marker, so Reconcile returns early.
func TestUpdatePodPatchErrorIsNotContinuable(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod",
			Namespace: "ns",
			Annotations: map[string]string{
				"received-resource-type":        "UnknownType",
				patcher.PodStatusAnnotationName: "Pending",
			},
		},
		Status: v1.PodStatus{Phase: v1.PodRunning},
	}
	patchErr := errors.New("patch failed")
	kubeClient := fake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).WithObjects(pod.DeepCopy()).
		WithInterceptorFuncs(interceptor.Funcs{
			Patch: func(context.Context, client.WithWatch, client.Object, client.Patch, ...client.PatchOption) error {
				return patchErr
			},
		}).Build()

	err := updatePod(context.Background(), pod.DeepCopy(), kubeClient)

	require.ErrorIs(t, err, patchErr)
	var gpuErr *gpuCalcError
	assert.False(t, errors.As(err, &gpuErr), "a failed status patch must not be continuable: %v", err)
}

// TestUpdatePodWritesAllAnnotations checks that a pod without a GPU calculation error still gets
// the status and all GPU annotations in one patch.
func TestUpdatePodWritesAllAnnotations(t *testing.T) {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod", Namespace: "ns"},
		Status:     v1.PodStatus{Phase: v1.PodRunning},
	}
	kubeClient := fake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).WithObjects(pod.DeepCopy()).Build()

	require.NoError(t, updatePod(context.Background(), pod.DeepCopy(), kubeClient))

	updated := &v1.Pod{}
	require.NoError(t, kubeClient.Get(context.Background(), client.ObjectKeyFromObject(pod), updated))
	assert.Equal(t, "Running", updated.Annotations[patcher.PodStatusAnnotationName])
	assert.Equal(t, "0", updated.Annotations[patcher.PodAllocatedGPUsAnnotationName])
	assert.Equal(t, "0", updated.Annotations[patcher.PodAllocatedGpuMemoryAnnotationName])
	assert.Equal(t, "0", updated.Annotations[patcher.PodAllocatedMigGPUsAnnotationName])
}
