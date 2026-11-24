/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/metadata"
)

// The node label holds the memory of a single GPU: 16384 MiB, which is 17179.869184 MB.
// The expected GPU memory values below are written out by hand (GPU count * 17179.869184 MB, rounded down), and are
// deliberately not computed with the production conversion functions.
const (
	draTestNamespace        = "dra-ns"
	draTestNode             = "gpu-node"
	draTestPodName          = "dra-pod"
	draTestPodUID           = types.UID("dra-pod-uid")
	draTestNodeGpuMemoryMiB = "16384"
	draTestGpuDeviceClass   = "gpu.nvidia.com"

	draTestMemoryOfOneGpuMb    = int64(17179)
	draTestMemoryOfTwoGpusMb   = int64(34359)
	draTestMemoryOfThreeGpusMb = int64(51539)
	draTestMemoryOfFourGpusMb  = int64(68719)
)

// stubDRAResolver returns a fixed GPU count. The first failures calls return an error.
type stubDRAResolver struct {
	gpus     int64
	failures int
	calls    int
}

func (s *stubDRAResolver) ResolvePodRequestedGpuCount(_ context.Context, _ *v1.Pod) (int64, error) {
	s.calls++
	if s.failures > 0 {
		s.failures--
		return 0, errors.New("claim is not readable yet")
	}
	return s.gpus, nil
}

func draTestClaimRef(name string) v1.PodResourceClaim {
	claimName := name
	return v1.PodResourceClaim{Name: name, ResourceClaimName: &claimName}
}

func draTestTemplateRef(name string) v1.PodResourceClaim {
	templateName := name
	return v1.PodResourceClaim{Name: name, ResourceClaimTemplateName: &templateName}
}

// draTestPod builds a pod of the given phase that was scheduled to the test node, with the given received resource
// type annotation and DRA claim references. It has no GPU request in its spec, like a pod that uses DRA.
func draTestPod(phase v1.PodPhase, receivedType string, claimRefs ...v1.PodResourceClaim) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        draTestPodName,
			Namespace:   draTestNamespace,
			UID:         draTestPodUID,
			Annotations: map[string]string{receivedResourceAnnotationName: receivedType},
		},
		Spec: v1.PodSpec{
			NodeName:       draTestNode,
			ResourceClaims: claimRefs,
			Containers:     []v1.Container{{Name: "main"}},
		},
		Status: v1.PodStatus{Phase: phase},
	}
}

func draTestNodeObject(gpuMemoryMiB string) *v1.Node {
	labels := map[string]string{}
	if gpuMemoryMiB != "" {
		labels[nodeGpuMemoryLabelName] = gpuMemoryMiB
	}
	return &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: draTestNode, Labels: labels}}
}

func draTestClaimSpec(deviceClass string, count int64) resourcev1.ResourceClaimSpec {
	return resourcev1.ResourceClaimSpec{Devices: resourcev1.DeviceClaim{Requests: []resourcev1.DeviceRequest{{
		Name: "gpu",
		Exactly: &resourcev1.ExactDeviceRequest{
			DeviceClassName: deviceClass,
			AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
			Count:           count,
		},
	}}}}
}

func calcDraTestMetrics(t *testing.T, pod *v1.Pod, kubeClient client.Client, resolver DRAGpuResolver) (
	*metadata.PodMetadata, error) {
	t.Helper()
	podMetrics := &metadata.PodMetadata{}
	err := CalcPodGpuResources(context.Background(), pod, podMetrics, kubeClient, resolver)
	return podMetrics, err
}

func TestCalcPodGpuResourcesDRA(t *testing.T) {
	for _, tc := range []struct {
		name         string
		pod          *v1.Pod
		draGpus      int64
		wantGpus     float64
		wantMemory   int64
		wantResolves int
	}{
		{
			name:         "Running pod with a two-GPU claim: reports two GPUs and 34359 MB",
			pod:          draTestPod(v1.PodRunning, receivedTypeRegular, draTestClaimRef("claim")),
			draGpus:      2,
			wantGpus:     2,
			wantMemory:   draTestMemoryOfTwoGpusMb,
			wantResolves: 1,
		},
		{
			name:         "Bound pod that is not running yet: reports one GPU and 17179 MB",
			pod:          draTestPod(v1.PodPending, receivedTypeRegular, draTestTemplateRef("template")),
			draGpus:      1,
			wantGpus:     1,
			wantMemory:   draTestMemoryOfOneGpuMb,
			wantResolves: 1,
		},
		{
			name: "Terminating pod with a four-GPU claim: keeps reporting four GPUs and 68719 MB",
			pod: func() *v1.Pod {
				p := draTestPod(v1.PodRunning, receivedTypeRegular, draTestClaimRef("claim"))
				now := metav1.Now()
				p.DeletionTimestamp = &now
				return p
			}(),
			draGpus:      4,
			wantGpus:     4,
			wantMemory:   draTestMemoryOfFourGpusMb,
			wantResolves: 1,
		},
		{
			name: "One GPU in the pod spec and a two-GPU claim: reports three GPUs and 51539 MB",
			pod: func() *v1.Pod {
				p := draTestPod(v1.PodRunning, receivedTypeRegular, draTestClaimRef("claim"))
				gpu := map[v1.ResourceName]resource.Quantity{gpuResourceName: resource.MustParse("1")}
				p.Spec.Containers[0].Resources = v1.ResourceRequirements{Requests: gpu, Limits: gpu}
				return p
			}(),
			draGpus:      2,
			wantGpus:     3,
			wantMemory:   draTestMemoryOfThreeGpusMb,
			wantResolves: 1,
		},
		{
			name:         "Non-GPU claim: reports zero GPUs and zero memory",
			pod:          draTestPod(v1.PodRunning, receivedTypeRegular, draTestClaimRef("claim")),
			draGpus:      0,
			wantGpus:     0,
			wantMemory:   0,
			wantResolves: 1,
		},
		{
			name: "Traditional two-GPU pod without claims: reports two GPUs and does not read DRA",
			pod: func() *v1.Pod {
				p := draTestPod(v1.PodRunning, receivedTypeRegular)
				gpu := map[v1.ResourceName]resource.Quantity{gpuResourceName: resource.MustParse("2")}
				p.Spec.Containers[0].Resources = v1.ResourceRequirements{Requests: gpu, Limits: gpu}
				return p
			}(),
			draGpus:      7, // must be ignored
			wantGpus:     2,
			wantMemory:   draTestMemoryOfTwoGpusMb,
			wantResolves: 0,
		},
		{
			name: "Pod that is not scheduled yet: reports zero GPUs and does not read DRA",
			pod: func() *v1.Pod {
				p := draTestPod(v1.PodPending, receivedTypeRegular, draTestClaimRef("claim"))
				p.Spec.NodeName = ""
				return p
			}(),
			draGpus:      2,
			wantGpus:     0,
			wantMemory:   0,
			wantResolves: 0,
		},
		{
			name:         "Succeeded pod: reports zero GPUs and does not read DRA",
			pod:          draTestPod(v1.PodSucceeded, receivedTypeRegular, draTestClaimRef("claim")),
			draGpus:      2,
			wantGpus:     0,
			wantMemory:   0,
			wantResolves: 0,
		},
		{
			name:         "Pod without a received resource type: reports zero GPUs and does not read DRA",
			pod:          draTestPod(v1.PodRunning, receivedTypeNone, draTestClaimRef("claim")),
			draGpus:      2,
			wantGpus:     0,
			wantMemory:   0,
			wantResolves: 0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			kubeClient := fake.NewClientBuilder().WithObjects(draTestNodeObject(draTestNodeGpuMemoryMiB)).Build()
			resolver := &stubDRAResolver{gpus: tc.draGpus}

			podMetrics, err := calcDraTestMetrics(t, tc.pod, kubeClient, resolver)

			require.NoError(t, err)
			require.InDelta(t, tc.wantGpus, podMetrics.GPUs, 1e-12)
			require.Equal(t, tc.wantMemory, podMetrics.GpuMemory)
			require.Equal(t, tc.wantResolves, resolver.calls)
		})
	}
}

func TestCalcPodGpuResourcesDRAFractionPodDoesNotReadDRA(t *testing.T) {
	pod := draTestPod(v1.PodRunning, receivedTypeFraction, draTestClaimRef("claim"))
	pod.Annotations[gpuFractionAnnotation] = "0.5"
	kubeClient := fake.NewClientBuilder().WithObjects(draTestNodeObject(draTestNodeGpuMemoryMiB)).Build()
	resolver := &stubDRAResolver{gpus: 3}

	podMetrics, err := calcDraTestMetrics(t, pod, kubeClient, resolver)

	require.NoError(t, err)
	require.Equal(t, 0.5, podMetrics.GPUs)
	require.Zero(t, resolver.calls)
}

func TestCalcPodGpuResourcesDRALookupFailureIsReturnedAndRecovers(t *testing.T) {
	pod := draTestPod(v1.PodRunning, receivedTypeRegular, draTestClaimRef("claim"))
	kubeClient := fake.NewClientBuilder().WithObjects(draTestNodeObject(draTestNodeGpuMemoryMiB)).Build()
	resolver := &stubDRAResolver{gpus: 2, failures: 1}

	// The first reconcile must fail, so that it is retried and no wrong (zero) GPU count is written.
	podMetrics, err := calcDraTestMetrics(t, pod, kubeClient, resolver)
	require.Error(t, err)
	require.Zero(t, podMetrics.GPUs)
	require.Zero(t, podMetrics.GpuMemory)

	// The retry succeeds.
	podMetrics, err = calcDraTestMetrics(t, pod, kubeClient, resolver)
	require.NoError(t, err)
	require.Equal(t, float64(2), podMetrics.GPUs)
	require.Equal(t, draTestMemoryOfTwoGpusMb, podMetrics.GpuMemory)
}

func TestCalcPodGpuResourcesWithoutResolverIgnoresClaims(t *testing.T) {
	pod := draTestPod(v1.PodRunning, receivedTypeRegular, draTestClaimRef("claim"))
	kubeClient := fake.NewClientBuilder().WithObjects(draTestNodeObject(draTestNodeGpuMemoryMiB)).Build()

	podMetrics, err := calcDraTestMetrics(t, pod, kubeClient, nil)

	require.NoError(t, err)
	require.Zero(t, podMetrics.GPUs)
}

// The GPU memory is the GPU count times the memory of a single GPU, taken from the node label.
// A node without the label keeps the GPU count and reports no memory, as it does for traditional pods.
func TestCalcPodGpuResourcesDRANodeWithoutGpuMemoryLabel(t *testing.T) {
	pod := draTestPod(v1.PodRunning, receivedTypeRegular, draTestClaimRef("claim"))
	kubeClient := fake.NewClientBuilder().WithObjects(draTestNodeObject("")).Build()

	podMetrics, err := calcDraTestMetrics(t, pod, kubeClient, &stubDRAResolver{gpus: 2})

	require.NoError(t, err)
	require.Equal(t, float64(2), podMetrics.GPUs)
	require.Zero(t, podMetrics.GpuMemory)
}

// The tests below use the real DRA parser over a fake cache, to cover how claims and templates are read.
func newRealDRAParser(objs ...client.Object) (*dra.DRAParser, client.Client) {
	objs = append(objs, draTestNodeObject(draTestNodeGpuMemoryMiB))
	kubeClient := fake.NewClientBuilder().WithObjects(objs...).Build()
	return dra.NewDRAParser(kubeClient, kubefake.NewClientset()), kubeClient
}

func TestCalcPodGpuResourcesDRAWithRealParser(t *testing.T) {
	existingClaim := &resourcev1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "existing", Namespace: draTestNamespace},
		Spec:       draTestClaimSpec(draTestGpuDeviceClass, 2),
	}
	// The template (one GPU) and the claim generated from it (three GPUs) deliberately differ, so the tests show
	// which of them was read.
	template := &resourcev1.ResourceClaimTemplate{
		ObjectMeta: metav1.ObjectMeta{Name: "tmpl", Namespace: draTestNamespace},
		Spec:       resourcev1.ResourceClaimTemplateSpec{Spec: draTestClaimSpec(draTestGpuDeviceClass, 1)},
	}
	isController := true
	generatedClaimOf := func(name string, ownerUID types.UID) *resourcev1.ResourceClaim {
		return &resourcev1.ResourceClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: draTestNamespace,
				OwnerReferences: []metav1.OwnerReference{{
					APIVersion: "v1", Kind: "Pod", Name: draTestPodName, UID: ownerUID, Controller: &isController,
				}},
			},
			Spec: draTestClaimSpec(draTestGpuDeviceClass, 3),
		}
	}
	otherClassClaim := &resourcev1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "other-class", Namespace: draTestNamespace},
		Spec:       draTestClaimSpec("network.example.com", 8),
	}
	generatedClaimName := "dra-pod-tmpl-abcde"
	foreignClaimName := "dra-pod-tmpl-foreign"
	generatedStatus := []v1.PodResourceClaimStatus{{Name: "tmpl", ResourceClaimName: &generatedClaimName}}
	foreignStatus := []v1.PodResourceClaimStatus{{Name: "tmpl", ResourceClaimName: &foreignClaimName}}

	for _, tc := range []struct {
		name       string
		claims     []v1.PodResourceClaim
		statuses   []v1.PodResourceClaimStatus
		wantGpus   float64
		wantMemory int64
		wantErr    bool
	}{
		{
			name:       "Existing two-GPU claim: reports two GPUs and 34359 MB",
			claims:     []v1.PodResourceClaim{draTestClaimRef("existing")},
			wantGpus:   2,
			wantMemory: draTestMemoryOfTwoGpusMb,
		},
		{
			name:       "One-GPU template without a generated claim yet: reports one GPU and 17179 MB",
			claims:     []v1.PodResourceClaim{draTestTemplateRef("tmpl")},
			wantGpus:   1,
			wantMemory: draTestMemoryOfOneGpuMb,
		},
		{
			name:       "Generated three-GPU claim named in the pod status: reports three GPUs and 51539 MB",
			claims:     []v1.PodResourceClaim{draTestTemplateRef("tmpl")},
			statuses:   generatedStatus,
			wantGpus:   3,
			wantMemory: draTestMemoryOfThreeGpusMb,
		},
		{
			name:     "Generated claim owned by another pod: returns an error",
			claims:   []v1.PodResourceClaim{draTestTemplateRef("tmpl")},
			statuses: foreignStatus,
			wantErr:  true,
		},
		{
			name:       "Independent two-GPU claim and one-GPU template: reports three GPUs and 51539 MB",
			claims:     []v1.PodResourceClaim{draTestClaimRef("existing"), draTestTemplateRef("tmpl")},
			wantGpus:   3,
			wantMemory: draTestMemoryOfThreeGpusMb,
		},
		{
			name:       "Non-GPU device class (network.example.com): reports zero GPUs",
			claims:     []v1.PodResourceClaim{draTestClaimRef("other-class")},
			wantGpus:   0,
			wantMemory: 0,
		},
		{
			name:    "Claim that cannot be found: returns an error",
			claims:  []v1.PodResourceClaim{draTestClaimRef("missing")},
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			parser, kubeClient := newRealDRAParser(existingClaim.DeepCopy(), template.DeepCopy(),
				otherClassClaim.DeepCopy(), generatedClaimOf(generatedClaimName, draTestPodUID),
				generatedClaimOf(foreignClaimName, types.UID("another-pod-uid")))
			pod := draTestPod(v1.PodRunning, receivedTypeRegular, tc.claims...)
			pod.Status.ResourceClaimStatuses = tc.statuses

			podMetrics, err := calcDraTestMetrics(t, pod, kubeClient, parser)

			if tc.wantErr {
				require.Error(t, err)
				require.Zero(t, podMetrics.GPUs)
				require.Zero(t, podMetrics.GpuMemory)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.wantGpus, podMetrics.GPUs)
			require.Equal(t, tc.wantMemory, podMetrics.GpuMemory)
		})
	}
}
