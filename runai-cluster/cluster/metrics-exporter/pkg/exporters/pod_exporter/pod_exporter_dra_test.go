/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_exporter

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	kubefake "k8s.io/client-go/kubernetes/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/dra"
)

const draTestPodUID = types.UID("dra-pod-uid")

// draStubResolver returns a fixed GPU count. The first failures calls return an error.
type draStubResolver struct {
	gpus     int64
	failures int
	calls    int
}

func (s *draStubResolver) ResolvePodRequestedGpuCount(_ context.Context, _ *v1.Pod) (int64, error) {
	s.calls++
	if s.failures > 0 {
		s.failures--
		return 0, errors.New("claim is not readable yet")
	}
	return s.gpus, nil
}

func draClaimRef(name string) v1.PodResourceClaim {
	claimName := name
	return v1.PodResourceClaim{Name: name, ResourceClaimName: &claimName}
}

func draTemplateRef(name string) v1.PodResourceClaim {
	templateName := name
	return v1.PodResourceClaim{Name: name, ResourceClaimTemplateName: &templateName}
}

// newDRATestPod builds a pod of the given phase, scheduled to a node, with the given DRA claim references.
// It has no GPU request in its spec, like a pod that uses DRA.
func newDRATestPod(phase v1.PodPhase, claimRefs ...v1.PodResourceClaim) *v1.Pod {
	pod := createPod()
	pod.UID = draTestPodUID
	pod.Status.Phase = phase
	pod.Spec = v1.PodSpec{
		NodeName:       "gpu-node",
		Containers:     []v1.Container{{Name: "main"}},
		ResourceClaims: claimRefs,
	}
	return pod
}

func withSpecGpus(pod *v1.Pod, gpus string) *v1.Pod {
	pod.Spec.Containers[0].Resources.Requests = v1.ResourceList{
		clustercommonconstants.NvidiaGpuResourceName: resource.MustParse(gpus),
	}
	return pod
}

func TestGetGpuNumDevicesDRA(t *testing.T) {
	for _, tc := range []struct {
		name         string
		pod          *v1.Pod
		draGpus      int64
		want         string
		wantResolves int
	}{
		{
			name:         "Running pod with a two-GPU claim: reports 2 devices",
			pod:          newDRATestPod(v1.PodRunning, draClaimRef("claim")),
			draGpus:      2,
			want:         "2",
			wantResolves: 1,
		},
		{
			name:         "Bound pod that is not running yet with a one-GPU template: reports 1 device",
			pod:          newDRATestPod(v1.PodPending, draTemplateRef("template")),
			draGpus:      1,
			want:         "1",
			wantResolves: 1,
		},
		{
			name: "Terminating pod with a four-GPU claim: keeps reporting 4 devices",
			pod: func() *v1.Pod {
				pod := newDRATestPod(v1.PodRunning, draClaimRef("claim"))
				now := metav1.Now()
				pod.DeletionTimestamp = &now
				return pod
			}(),
			draGpus:      4,
			want:         "4",
			wantResolves: 1,
		},
		{
			name:         "One GPU in the pod spec and a two-GPU claim: reports 3 devices",
			pod:          withSpecGpus(newDRATestPod(v1.PodRunning, draClaimRef("claim")), "1"),
			draGpus:      2,
			want:         "3",
			wantResolves: 1,
		},
		{
			name:         "Non-GPU claim: reports 0 devices",
			pod:          newDRATestPod(v1.PodRunning, draClaimRef("claim")),
			draGpus:      0,
			want:         "0",
			wantResolves: 1,
		},
		{
			name:         "Traditional two-GPU pod without claims: reports 2 devices and does not read DRA",
			pod:          withSpecGpus(newDRATestPod(v1.PodRunning), "2"),
			draGpus:      7, // must be ignored
			want:         "2",
			wantResolves: 0,
		},
		{
			name: "Fraction pod with two devices: reports 2 devices and does not read DRA",
			pod: func() *v1.Pod {
				pod := newDRATestPod(v1.PodRunning, draClaimRef("claim"))
				pod.Annotations[receivedResourceTypeAnnotationName] = clustercommonconstants.ReceivedTypeFraction
				pod.Annotations[numberOfDevicesAnnotationName] = "2"
				return pod
			}(),
			draGpus:      7, // must be ignored
			want:         "2",
			wantResolves: 0,
		},
		{
			name: "Pod that is not scheduled yet: reports 0 devices and does not read DRA",
			pod: func() *v1.Pod {
				pod := newDRATestPod(v1.PodPending, draClaimRef("claim"))
				pod.Spec.NodeName = ""
				return pod
			}(),
			draGpus:      2,
			want:         "0",
			wantResolves: 0,
		},
		{
			name:         "Succeeded pod: reports 0 devices and does not read DRA",
			pod:          newDRATestPod(v1.PodSucceeded, draClaimRef("claim")),
			draGpus:      2,
			want:         "0",
			wantResolves: 0,
		},
		{
			name:         "Failed pod: reports 0 devices and does not read DRA",
			pod:          newDRATestPod(v1.PodFailed, draClaimRef("claim")),
			draGpus:      2,
			want:         "0",
			wantResolves: 0,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resolver := &draStubResolver{gpus: tc.draGpus}
			exporter := NewPodExporter(context.TODO(), nil, tc.pod, WithDRAGpuResolver(resolver))

			numDevices, err := exporter.GetGpuNumDevices()

			require.NoError(t, err)
			require.Equal(t, tc.want, numDevices)
			require.Equal(t, tc.wantResolves, resolver.calls)
		})
	}
}

func TestGetGpuNumDevicesDRALookupFailureIsReturnedAndRecovers(t *testing.T) {
	pod := newDRATestPod(v1.PodRunning, draClaimRef("claim"))
	resolver := &draStubResolver{gpus: 2, failures: 1}
	exporter := NewPodExporter(context.TODO(), nil, pod, WithDRAGpuResolver(resolver))

	// The first reconcile must fail, so that it is retried and no wrong (zero) device count is exported.
	numDevices, err := exporter.GetGpuNumDevices()
	require.Error(t, err)
	require.Equal(t, "", numDevices)

	// The retry succeeds.
	numDevices, err = exporter.GetGpuNumDevices()
	require.NoError(t, err)
	require.Equal(t, "2", numDevices)
}

func TestGetGpuNumDevicesWithoutResolverIgnoresClaims(t *testing.T) {
	pod := newDRATestPod(v1.PodRunning, draClaimRef("claim"))
	exporter := NewPodExporter(context.TODO(), nil, pod)

	numDevices, err := exporter.GetGpuNumDevices()

	require.NoError(t, err)
	require.Equal(t, "0", numDevices)
}

func newDRATestClaimSpec(deviceClass string, count int64) resourcev1.ResourceClaimSpec {
	return resourcev1.ResourceClaimSpec{Devices: resourcev1.DeviceClaim{Requests: []resourcev1.DeviceRequest{{
		Name: "gpu",
		Exactly: &resourcev1.ExactDeviceRequest{
			DeviceClassName: deviceClass,
			AllocationMode:  resourcev1.DeviceAllocationModeExactCount,
			Count:           count,
		},
	}}}}
}

func newDRATestClient(objects ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	utilruntime.Must(v1.AddToScheme(scheme))
	utilruntime.Must(resourcev1.AddToScheme(scheme))
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

// The tests below use the real DRA parser over a fake cache, to cover how claims and templates are read.
func TestGetGpuNumDevicesDRAWithRealParser(t *testing.T) {
	existingClaim := &resourcev1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "existing", Namespace: podNamespace},
		Spec:       newDRATestClaimSpec("gpu.nvidia.com", 2),
	}
	template := &resourcev1.ResourceClaimTemplate{
		ObjectMeta: metav1.ObjectMeta{Name: "tmpl", Namespace: podNamespace},
		Spec:       resourcev1.ResourceClaimTemplateSpec{Spec: newDRATestClaimSpec("gpu.nvidia.com", 1)},
	}
	nonGpuClaim := &resourcev1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{Name: "non-gpu", Namespace: podNamespace},
		Spec:       newDRATestClaimSpec("network.example.com", 8),
	}
	// The template (one GPU) and the claim generated from it (three GPUs) deliberately differ, so the test shows
	// which of them was read.
	isController := true
	generatedClaimName := "my-pod-tmpl-abcde"
	generatedClaim := &resourcev1.ResourceClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      generatedClaimName,
			Namespace: podNamespace,
			OwnerReferences: []metav1.OwnerReference{{
				APIVersion: "v1", Kind: "Pod", Name: podName, UID: draTestPodUID, Controller: &isController,
			}},
		},
		Spec: newDRATestClaimSpec("gpu.nvidia.com", 3),
	}
	generatedStatus := []v1.PodResourceClaimStatus{{Name: "tmpl", ResourceClaimName: &generatedClaimName}}

	for _, tc := range []struct {
		name     string
		claims   []v1.PodResourceClaim
		statuses []v1.PodResourceClaimStatus
		want     string
		wantErr  bool
	}{
		{
			name:   "Existing two-GPU claim: reports 2 devices",
			claims: []v1.PodResourceClaim{draClaimRef("existing")},
			want:   "2",
		},
		{
			name:   "One-GPU template without a generated claim yet: reports 1 device",
			claims: []v1.PodResourceClaim{draTemplateRef("tmpl")},
			want:   "1",
		},
		{
			name:     "Generated three-GPU claim named in the pod status: reports 3 devices",
			claims:   []v1.PodResourceClaim{draTemplateRef("tmpl")},
			statuses: generatedStatus,
			want:     "3",
		},
		{
			name:   "Independent two-GPU claim and one-GPU template: reports 3 devices",
			claims: []v1.PodResourceClaim{draClaimRef("existing"), draTemplateRef("tmpl")},
			want:   "3",
		},
		{
			name:   "Non-GPU device class (network.example.com): reports 0 devices",
			claims: []v1.PodResourceClaim{draClaimRef("non-gpu")},
			want:   "0",
		},
		{
			name:    "Claim that cannot be found: returns an error",
			claims:  []v1.PodResourceClaim{draClaimRef("missing")},
			wantErr: true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cachedClient := newDRATestClient(existingClaim.DeepCopy(), template.DeepCopy(),
				nonGpuClaim.DeepCopy(), generatedClaim.DeepCopy())
			parser := dra.NewDRAParser(cachedClient, kubefake.NewClientset())
			pod := newDRATestPod(v1.PodRunning, tc.claims...)
			pod.Status.ResourceClaimStatuses = tc.statuses
			exporter := NewPodExporter(context.TODO(), cachedClient, pod, WithDRAGpuResolver(parser))

			numDevices, err := exporter.GetGpuNumDevices()

			if tc.wantErr {
				require.Error(t, err)
				require.Equal(t, "", numDevices)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, numDevices)
		})
	}
}
