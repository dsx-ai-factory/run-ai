/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"slices"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAcceleratorResources(t *testing.T) {
	node := corev1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
		"nvidia.com/gpu.product": "H100", "example.org/device.product": "X1",
	}}, Status: corev1.NodeStatus{Allocatable: corev1.ResourceList{
		"nvidia.com/gpu": resource.MustParse("2"), "amd.com/gpu": resource.MustParse("1"),
		"example.org/device": resource.MustParse("1"), "example.org/network": resource.MustParse("1"),
		"gpu.intel.com/i915": resource.MustParse("2"), "kubernetes.io/native": resource.MustParse("8"),
		"empty.com/gpu": resource.MustParse("0"), corev1.ResourceCPU: resource.MustParse("8"),
	}}}
	got := acceleratorResources(&node)
	want := []string{"amd.com/gpu", "example.org/device", "example.org/network", "gpu.intel.com/i915", "nvidia.com/gpu"}
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
