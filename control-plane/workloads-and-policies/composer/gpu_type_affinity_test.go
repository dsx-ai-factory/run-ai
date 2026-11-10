// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
package composer

import (
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/component-helpers/scheduling/corev1/nodeaffinity"
	"testing"
)

func TestGPUTypeAffinityIntersectsPlacement(t *testing.T) {
	for _, tc := range []struct {
		name     string
		required *corev1.NodeSelector
		allow    bool
	}{
		{"unrestricted", nil, true},
		{"alternative pools", &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{
			{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "pool", Operator: corev1.NodeSelectorOpIn, Values: []string{"a"}}}},
			{MatchExpressions: []corev1.NodeSelectorRequirement{{Key: "pool", Operator: corev1.NodeSelectorOpIn, Values: []string{"b"}}}},
		}}, true},
		{"empty selector", &corev1.NodeSelector{}, false},
		{"empty term", &corev1.NodeSelector{NodeSelectorTerms: []corev1.NodeSelectorTerm{{}}}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			affinity := &corev1.Affinity{PodAntiAffinity: &corev1.PodAntiAffinity{}, NodeAffinity: &corev1.NodeAffinity{RequiredDuringSchedulingIgnoredDuringExecution: tc.required}}
			require.Nil(t, GPUTypeToNodeAffinity(convutils.Ref("H100"), &affinity, nil))
			require.NotNil(t, affinity.PodAntiAffinity)
			pod := &corev1.Pod{Spec: corev1.PodSpec{Affinity: affinity, NodeSelector: map[string]string{"zone": "east"}}}
			for _, pool := range []string{"a", "b"} {
				for _, gpu := range []string{"H100", "A100", ""} {
					for _, zone := range []string{"east", "west"} {
						match, err := nodeaffinity.GetRequiredNodeAffinity(pod).Match(&corev1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"pool": pool, NvidiaGPUProductLabel: gpu, "zone": zone}}})
						require.NoError(t, err)
						require.Equal(t, tc.allow && gpu == "H100" && zone == "east", match, "pool=%s gpu=%s zone=%s", pool, gpu, zone)
					}
				}
			}
		})
	}
}
