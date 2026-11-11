/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package node_mig_exporter

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestIsMigNodeByGpuProduct(t *testing.T) {
	tests := []struct {
		gpuProduct string
		want       bool
	}{
		{"NVIDIA-A100-SXM4-80GB", true},
		{"NVIDIA-H100-80GB-HBM3", true},
		{"NVIDIA-H200", true},
		{"NVIDIA-B200", true},
		{"NVIDIA-GB200", true},
		{"NVIDIA-B300-SXM6-AC", true},
		{"NVIDIA-RTX-PRO-6000-Blackwell-Max-Q-Workstation-Edition", true},
		{"NVIDIA-RTX-PRO-5000-Blackwell", true},
		{"NVIDIA-L40S", false},
	}
	for _, tt := range tests {
		t.Run(tt.gpuProduct, func(t *testing.T) {
			node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
				GpuProductLabel:  tt.gpuProduct,
				MigStrategyLabel: MigStrategyMixed,
				MigEnabledLabel:  TrueLabelValue,
			}}}
			if got := NewMigStrategyExporter(node).IsMigNode(); got != tt.want {
				t.Errorf("IsMigNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
