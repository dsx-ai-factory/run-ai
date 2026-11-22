/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import "testing"

// Fake-GPU mode labels every worker node so the simulator can place GPUs there.
// A control-plane node must never be labeled, and kubeadm has moved from the
// legacy master role label to the control-plane one.
func TestIsControlPlaneNode(t *testing.T) {
	for _, tc := range []struct {
		name   string
		labels map[string]string
		want   bool
	}{
		{name: "no labels", labels: nil, want: false},
		{name: "empty labels", labels: map[string]string{}, want: false},
		{
			name:   "worker node",
			labels: map[string]string{"kubernetes.io/hostname": "worker-1"},
			want:   false,
		},
		{
			name:   "worker node in the runai node pool",
			labels: map[string]string{"run.ai/simulated-gpu-node-pool": "default"},
			want:   false,
		},
		{
			name:   "control-plane role label",
			labels: map[string]string{"node-role.kubernetes.io/control-plane": ""},
			want:   true,
		},
		{
			name:   "legacy master role label",
			labels: map[string]string{"node-role.kubernetes.io/master": ""},
			want:   true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isControlPlaneNode(tc.labels); got != tc.want {
				t.Fatalf("isControlPlaneNode(%v) = %t, want %t", tc.labels, got, tc.want)
			}
		})
	}
}
