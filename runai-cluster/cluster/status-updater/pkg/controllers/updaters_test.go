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

// These tests check the annotations that updatePod saves on the Pod, which are the values other services read.
// The node label holds the memory of a single GPU: 16384 MiB, which is 17179.869184 MB. The expected GPU memory
// values are written out by hand (GPU count * 17179.869184 MB, rounded down).
const (
	updatePodTestNamespace = "update-pod-ns"
	updatePodTestPodName   = "dra-pod"
	updatePodTestNodeName  = "gpu-node"

	// Literals of constants that are not exported by the pod metadata calculator.
	updatePodTestReceivedTypeAnnotation = "received-resource-type"
	updatePodTestReceivedTypeRegular    = "Regular"
	updatePodTestNodeGpuMemoryLabel     = "nvidia.com/gpu.memory"

	updatePodTestMemoryOfOneGpuMb  = "17179"
	updatePodTestMemoryOfTwoGpusMb = "34359"
)

// updatePodStubResolver returns a fixed GPU count. The first failures calls return an error.
type updatePodStubResolver struct {
	gpus     int64
	failures int
	calls    int
}

func (s *updatePodStubResolver) ResolvePodRequestedGpuCount(_ context.Context, _ *v1.Pod) (int64, error) {
	s.calls++
	if s.failures > 0 {
		s.failures--
		return 0, errors.New("claim is not readable yet")
	}
	return s.gpus, nil
}

// newUpdatePodTestClient returns a client holding a DRA pod of the given phase, with the given saved annotations, and
// the node it runs on.
func newUpdatePodTestClient(phase v1.PodPhase, savedAnnotations map[string]string) client.Client {
	annotations := map[string]string{updatePodTestReceivedTypeAnnotation: updatePodTestReceivedTypeRegular}
	for key, value := range savedAnnotations {
		annotations[key] = value
	}
	claimName := "gpu-claim"
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: updatePodTestPodName, Namespace: updatePodTestNamespace, Annotations: annotations,
		},
		Spec: v1.PodSpec{
			NodeName:       updatePodTestNodeName,
			Containers:     []v1.Container{{Name: "main"}},
			ResourceClaims: []v1.PodResourceClaim{{Name: "gpu", ResourceClaimName: &claimName}},
		},
		Status: v1.PodStatus{Phase: phase},
	}
	node := &v1.Node{ObjectMeta: metav1.ObjectMeta{
		Name:   updatePodTestNodeName,
		Labels: map[string]string{updatePodTestNodeGpuMemoryLabel: "16384"},
	}}
	return fake.NewClientBuilder().WithObjects(pod, node).Build()
}

func getUpdatePodTestPod(t *testing.T, kubeClient client.Client) *v1.Pod {
	t.Helper()
	pod := &v1.Pod{}
	key := client.ObjectKey{Namespace: updatePodTestNamespace, Name: updatePodTestPodName}
	require.NoError(t, kubeClient.Get(context.Background(), key, pod))
	return pod
}

// runUpdatePod runs updatePod on the stored pod, and returns the error with the pod as stored afterwards.
func runUpdatePod(t *testing.T, kubeClient client.Client, resolver *updatePodStubResolver) (*v1.Pod, error) {
	t.Helper()
	err := updatePod(context.Background(), getUpdatePodTestPod(t, kubeClient), kubeClient, resolver)
	return getUpdatePodTestPod(t, kubeClient), err
}

func TestUpdatePodSavesDRAAllocationOnPod(t *testing.T) {
	kubeClient := newUpdatePodTestClient(v1.PodRunning, nil)

	saved, err := runUpdatePod(t, kubeClient, &updatePodStubResolver{gpus: 2})

	require.NoError(t, err)
	require.Equal(t, "2", saved.Annotations[patcher.PodAllocatedGPUsAnnotationName])
	require.Equal(t, updatePodTestMemoryOfTwoGpusMb,
		saved.Annotations[patcher.PodAllocatedGpuMemoryAnnotationName])
}

func TestUpdatePodLookupFailureKeepsSavedAllocationAndRecovers(t *testing.T) {
	for _, tc := range []struct {
		name             string
		savedAnnotations map[string]string
		wantGpusAfterErr string
		wantMemAfterErr  string
	}{
		{
			name: "Pod with a saved one-GPU allocation: failure keeps one GPU, recovery saves two GPUs",
			savedAnnotations: map[string]string{
				patcher.PodAllocatedGPUsAnnotationName:      "1",
				patcher.PodAllocatedGpuMemoryAnnotationName: updatePodTestMemoryOfOneGpuMb,
			},
			wantGpusAfterErr: "1",
			wantMemAfterErr:  updatePodTestMemoryOfOneGpuMb,
		},
		{
			name: "Pod without a saved allocation: failure writes nothing, recovery saves two GPUs",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			kubeClient := newUpdatePodTestClient(v1.PodRunning, tc.savedAnnotations)
			resolver := &updatePodStubResolver{gpus: 2, failures: 1}

			// Temporary lookup failure: the error is returned, so the reconcile is retried.
			saved, err := runUpdatePod(t, kubeClient, resolver)
			require.Error(t, err)
			require.Equal(t, tc.wantGpusAfterErr, saved.Annotations[patcher.PodAllocatedGPUsAnnotationName])
			require.Equal(t, tc.wantMemAfterErr, saved.Annotations[patcher.PodAllocatedGpuMemoryAnnotationName])

			// Recovery: the retry saves the real allocation.
			saved, err = runUpdatePod(t, kubeClient, resolver)
			require.NoError(t, err)
			require.Equal(t, "2", saved.Annotations[patcher.PodAllocatedGPUsAnnotationName])
			require.Equal(t, updatePodTestMemoryOfTwoGpusMb,
				saved.Annotations[patcher.PodAllocatedGpuMemoryAnnotationName])
		})
	}
}

func TestUpdatePodFinishedPodClearsAllocation(t *testing.T) {
	for _, phase := range []v1.PodPhase{v1.PodSucceeded, v1.PodFailed} {
		t.Run(string(phase)+" pod with a saved two-GPU allocation: clears GPUs and memory to zero", func(t *testing.T) {
			kubeClient := newUpdatePodTestClient(phase, map[string]string{
				patcher.PodAllocatedGPUsAnnotationName:      "2",
				patcher.PodAllocatedGpuMemoryAnnotationName: updatePodTestMemoryOfTwoGpusMb,
			})
			// The claims of a finished pod may no longer exist, so no lookup may be made, and none may fail the update.
			resolver := &updatePodStubResolver{gpus: 2, failures: 1000}

			saved, err := runUpdatePod(t, kubeClient, resolver)

			require.NoError(t, err)
			require.Zero(t, resolver.calls)
			require.Equal(t, "0", saved.Annotations[patcher.PodAllocatedGPUsAnnotationName])
			require.Equal(t, "0", saved.Annotations[patcher.PodAllocatedGpuMemoryAnnotationName])
		})
	}
}

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

	err := updatePod(context.Background(), pod.DeepCopy(), kubeClient, nil)

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

	err := updatePod(context.Background(), pod.DeepCopy(), kubeClient, nil)

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

	require.NoError(t, updatePod(context.Background(), pod.DeepCopy(), kubeClient, nil))

	updated := &v1.Pod{}
	require.NoError(t, kubeClient.Get(context.Background(), client.ObjectKeyFromObject(pod), updated))
	assert.Equal(t, "Running", updated.Annotations[patcher.PodStatusAnnotationName])
	assert.Equal(t, "0", updated.Annotations[patcher.PodAllocatedGPUsAnnotationName])
	assert.Equal(t, "0", updated.Annotations[patcher.PodAllocatedGpuMemoryAnnotationName])
	assert.Equal(t, "0", updated.Annotations[patcher.PodAllocatedMigGPUsAnnotationName])
}

// TestUpdatePodLookupFailureWritesStatusAndKeepsAllocation checks that a DRA lookup failure on a pod whose saved
// status is stale still writes the current status, keeps the saved allocation and returns a gpuCalcError, so
// Reconcile still updates the workload and the pod is retried.
func TestUpdatePodLookupFailureWritesStatusAndKeepsAllocation(t *testing.T) {
	kubeClient := newUpdatePodTestClient(v1.PodRunning, map[string]string{
		patcher.PodStatusAnnotationName:             "Pending",
		patcher.PodAllocatedGPUsAnnotationName:      "1",
		patcher.PodAllocatedGpuMemoryAnnotationName: updatePodTestMemoryOfOneGpuMb,
	})

	saved, err := runUpdatePod(t, kubeClient, &updatePodStubResolver{gpus: 2, failures: 1})

	var gpuErr *gpuCalcError
	require.True(t, errors.As(err, &gpuErr), "expected a gpuCalcError, got %v", err)
	require.Equal(t, "Running", saved.Annotations[patcher.PodStatusAnnotationName])
	require.Equal(t, "1", saved.Annotations[patcher.PodAllocatedGPUsAnnotationName])
	require.Equal(t, updatePodTestMemoryOfOneGpuMb, saved.Annotations[patcher.PodAllocatedGpuMemoryAnnotationName])
}
