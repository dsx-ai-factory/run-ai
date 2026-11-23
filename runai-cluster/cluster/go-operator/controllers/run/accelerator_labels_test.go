/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package run

import (
	"testing"

	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNodeAcceleratorResourcesFollowAllocatable(t *testing.T) {
	node := &corev1.Node{}
	node.Labels = map[string]string{"team": "research"}
	node.Status.Allocatable = corev1.ResourceList{
		corev1.ResourceCPU: resource.MustParse("8"),
		"nvidia.com/gpu":   resource.MustParse("4"),
		"amd.com/gpu":      resource.MustParse("0"),
	}
	nvidia := constants.AcceleratorResourceNodeLabel("nvidia.com/gpu")
	amd := constants.AcceleratorResourceNodeLabel("amd.com/gpu")

	labels, changed := acceleratorNodeLabels(node)
	require.True(t, changed)
	require.Equal(t, map[string]string{"team": "research", nvidia: "true"}, labels)
	node.Labels = labels
	_, changed = acceleratorNodeLabels(node)
	require.False(t, changed)

	// Simulate resource removal and later discovery by a different device plugin.
	delete(node.Status.Allocatable, "nvidia.com/gpu")
	node.Status.Allocatable["amd.com/gpu"] = resource.MustParse("2")
	labels, changed = acceleratorNodeLabels(node)
	require.True(t, changed)
	require.Equal(t, map[string]string{"team": "research", amd: "true"}, labels)
	node.Labels = labels
	node.Labels[constants.NodePoolNameLabel] = clusterconstants.ExcludedNodepoolName
	labels, changed = acceleratorNodeLabels(node)
	require.True(t, changed)
	require.NotContains(t, labels, amd)
	require.Equal(t, "research", labels["team"])
}
