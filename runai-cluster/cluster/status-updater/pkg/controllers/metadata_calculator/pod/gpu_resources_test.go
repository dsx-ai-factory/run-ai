/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/metadata"
)

const a100GpuProductName = "NVIDIA-A100-SXM4-40GB"

func TestGetDeviceComputePortions_A100(t *testing.T) {
	portion, err := getDeviceComputePortions(a100GpuProductName)
	assert.Nil(t, err)
	assert.Equal(t, defaultComputePortions, portion)
}

func TestGetDeviceComputePortions_NotMIG(t *testing.T) {
	_, err := getDeviceComputePortions("Tesla-T4")
	assert.NotNil(t, err)
}

// Value of nvidia.com/gpu.product published by GPU Feature Discovery on an RTX PRO 6000 Max-Q node.
const rtxPro6000GpuProductName = "NVIDIA-RTX-PRO-6000-Blackwell-Max-Q-Workstation-Edition"

func TestGetDeviceComputePortions_RTXPro6000(t *testing.T) {
	portion, err := getDeviceComputePortions(rtxPro6000GpuProductName)
	assert.Nil(t, err)
	assert.Equal(t, float64(4), portion)
}

// newRunningMigPod returns a running pod scheduled on mixed-node that requests quantity of migResource in its limits
// (no MIG request when migResource is empty).
func newRunningMigPod(annotations map[string]string, migResource string, quantity string) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "mig-pod", Namespace: "ns", Annotations: annotations},
		Spec: v1.PodSpec{
			NodeName: "mixed-node",
			Containers: []v1.Container{{
				Name: "c",
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{},
				},
			}},
		},
		Status: v1.PodStatus{Phase: v1.PodRunning},
	}
	if migResource != "" {
		pod.Spec.Containers[0].Resources.Limits[v1.ResourceName(migResource)] = resource.MustParse(quantity)
	}
	return pod
}

// TestGetDeviceComputePortions_Blackwell checks the compute slices per GPU product against the NVIDIA MIG user guide.
func TestGetDeviceComputePortions_Blackwell(t *testing.T) {
	// Compute slices per GPU from the NVIDIA MIG user guide (supported MIG profiles).
	tests := map[string]float64{
		"NVIDIA-H200":                   7,
		"NVIDIA-B200":                   7,
		"NVIDIA-GB200":                  7,
		"NVIDIA-B300-SXM6-AC":           7,
		"NVIDIA-RTX-PRO-5000-Blackwell": 2,
		rtxPro6000GpuProductName:        4,
		"NVIDIA-RTX-PRO-6000-Blackwell-Server-Edition": 4,
	}
	for product, expected := range tests {
		t.Run(product, func(t *testing.T) {
			portion, err := getDeviceComputePortions(product)
			assert.Nil(t, err)
			assert.Equal(t, expected, portion)
		})
	}
}

// TestCalcPodGpuResources_MigOnMixedNode covers MIG accounting on a mig.strategy=mixed node, where the scheduler may
// leave the runai-mig-device annotation empty and the allocation comes from the requested MIG resources.
func TestCalcPodGpuResources_MigOnMixedNode(t *testing.T) {
	mig := map[string]string{receivedResourceAnnotationName: receivedTypeMigInstance}
	withDevice := func(device string) map[string]string {
		return map[string]string{receivedResourceAnnotationName: receivedTypeMigInstance, migDeviceAnnotationName: device}
	}

	tests := map[string]struct {
		gpuProduct  string
		annotations map[string]string
		migResource string
		quantity    string
		// second MIG profile requested by the same container, with quantity 1
		extraMigResource string
		expected         float64
		expectErr        bool
	}{
		"mig-device annotation wins over the request": {
			annotations: withDevice(`{"name":"nvidia.com/mig-2g.48gb"}`),
			migResource: "nvidia.com/mig-1g.24gb", quantity: "1", expected: 0.5,
		},
		"missing mig-device annotation falls back to the request": {
			annotations: mig, migResource: "nvidia.com/mig-1g.24gb", quantity: "1", expected: 0.25,
		},
		"empty mig-device annotation falls back to the request": {
			annotations: withDevice(""), migResource: "nvidia.com/mig-2g.48gb", quantity: "1", expected: 0.5,
		},
		"mig-device annotation without a name falls back to the request": {
			annotations: withDevice(`{}`), migResource: "nvidia.com/mig-1g.24gb", quantity: "1", expected: 0.25,
		},
		"mig-device annotation with an empty name falls back to the request": {
			annotations: withDevice(`{"name":""}`), migResource: "nvidia.com/mig-1g.24gb", quantity: "1", expected: 0.25,
		},
		"invalid mig-device annotation is an error": {
			annotations: withDevice(`{`), migResource: "nvidia.com/mig-1g.24gb", quantity: "1", expectErr: true,
		},
		"no mig-device annotation and no mig request is an error": {
			annotations: mig, expectErr: true,
		},
		"every requested mig profile is summed": {
			annotations: mig, migResource: "nvidia.com/mig-1g.24gb", quantity: "1",
			extraMigResource: "nvidia.com/mig-2g.48gb", expected: 0.75,
		},
		"requested quantity is taken into account": {
			annotations: mig, migResource: "nvidia.com/mig-1g.24gb", quantity: "2", expected: 0.5,
		},
		"whole gpu profile": {
			annotations: mig, migResource: "nvidia.com/mig-4g.96gb", quantity: "1", expected: 1,
		},
		"annotation profile with media extensions suffix": {
			annotations: withDevice(`{"name":"nvidia.com/mig-1g.24gb+me"}`),
			migResource: "nvidia.com/mig-1g.24gb.me", quantity: "1", expected: 0.25,
		},
		// the NVIDIA device plugin publishes "+" as "." in the resource name (strings.ReplaceAll in internal/rm)
		"requested profile with media extensions suffix": {
			annotations: mig, migResource: "nvidia.com/mig-1g.24gb.me", quantity: "1", expected: 0.25,
		},
		"requested profile with all media extensions suffix": {
			annotations: mig, migResource: "nvidia.com/mig-2g.48gb.me.all", quantity: "1", expected: 0.5,
		},
		"requested profile without media extensions suffix": {
			annotations: mig, migResource: "nvidia.com/mig-1g.24gb-me", quantity: "1", expected: 0.25,
		},
		"seven slice gpu": {
			gpuProduct:  "NVIDIA-B200",
			annotations: mig, migResource: "nvidia.com/mig-1g.23gb", quantity: "7", expected: 1,
		},
		"unsupported gpu product is an error": {
			gpuProduct:  "Tesla-T4",
			annotations: mig, migResource: "nvidia.com/mig-1g.5gb", quantity: "1", expectErr: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			gpuProduct := tc.gpuProduct
			if gpuProduct == "" {
				gpuProduct = rtxPro6000GpuProductName
			}
			node := &v1.Node{ObjectMeta: metav1.ObjectMeta{
				Name:   "mixed-node",
				Labels: map[string]string{nodeGpuTypeLabelName: gpuProduct, "nvidia.com/mig.strategy": "mixed"},
			}}
			kubeClient := fake.NewClientBuilder().WithObjects(node).Build()
			pod := newRunningMigPod(tc.annotations, tc.migResource, tc.quantity)
			if tc.extraMigResource != "" {
				pod.Spec.Containers[0].Resources.Limits[v1.ResourceName(tc.extraMigResource)] = resource.MustParse("1")
			}
			podMeta := metadata.PodMetadata{Status: TranslatePodStatus(pod)}

			err := CalcPodGpuResources(context.Background(), pod, &podMeta, kubeClient)

			if tc.expectErr {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
			assert.InDelta(t, tc.expected, podMeta.MigGPUs, 1e-9)
			assert.Equal(t, running, podMeta.Status)
		})
	}
}
