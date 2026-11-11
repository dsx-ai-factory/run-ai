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

func newRunningMigPod(annotations map[string]string, migResource string, quantity string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "mig-pod", Namespace: "ns", Annotations: annotations},
		Spec: v1.PodSpec{
			NodeName: "mixed-node",
			Containers: []v1.Container{{
				Name: "c",
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{v1.ResourceName(migResource): resource.MustParse(quantity)},
				},
			}},
		},
		Status: v1.PodStatus{Phase: v1.PodRunning},
	}
}

func TestCalcPodGpuResources_MigOnMixedRTXPro6000Node(t *testing.T) {
	node := &v1.Node{ObjectMeta: metav1.ObjectMeta{
		Name:   "mixed-node",
		Labels: map[string]string{nodeGpuTypeLabelName: rtxPro6000GpuProductName, "nvidia.com/mig.strategy": "mixed"},
	}}
	kubeClient := fake.NewClientBuilder().WithObjects(node).Build()

	tests := map[string]struct {
		annotations map[string]string
		migResource string
		quantity    string
		// second MIG profile requested by the same container, with quantity 1
		extraMigResource string
		expected         float64
	}{
		"mig-device annotation": {
			annotations: map[string]string{
				receivedResourceAnnotationName: receivedTypeMigInstance,
				migDeviceAnnotationName:        `{"name":"nvidia.com/mig-1g.24gb"}`,
			},
			migResource: "nvidia.com/mig-1g.24gb", quantity: "1", expected: 0.25,
		},
		"mig resource request without mig-device annotation": {
			annotations: map[string]string{receivedResourceAnnotationName: receivedTypeMigInstance},
			migResource: "nvidia.com/mig-1g.24gb", quantity: "1", expected: 0.25,
		},
		"empty mig-device annotation falls back to the mig resource request": {
			annotations: map[string]string{
				receivedResourceAnnotationName: receivedTypeMigInstance,
				migDeviceAnnotationName:        "",
			},
			migResource: "nvidia.com/mig-2g.48gb", quantity: "1", expected: 0.5,
		},
		"every requested mig profile is summed": {
			annotations: map[string]string{receivedResourceAnnotationName: receivedTypeMigInstance},
			migResource: "nvidia.com/mig-1g.24gb", quantity: "1", extraMigResource: "nvidia.com/mig-2g.48gb",
			expected: 0.75,
		},
		"requested quantity is taken into account": {
			annotations: map[string]string{receivedResourceAnnotationName: receivedTypeMigInstance},
			migResource: "nvidia.com/mig-1g.24gb", quantity: "2", expected: 0.5,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			pod := newRunningMigPod(tc.annotations, tc.migResource, tc.quantity)
			if tc.extraMigResource != "" {
				pod.Spec.Containers[0].Resources.Limits[v1.ResourceName(tc.extraMigResource)] = resource.MustParse("1")
			}
			podMeta := metadata.PodMetadata{Status: TranslatePodStatus(pod)}

			err := CalcPodGpuResources(context.Background(), pod, &podMeta, kubeClient)

			assert.Nil(t, err)
			assert.Equal(t, tc.expected, podMeta.MigGPUs)
			assert.Equal(t, running, podMeta.Status)
		})
	}
}
