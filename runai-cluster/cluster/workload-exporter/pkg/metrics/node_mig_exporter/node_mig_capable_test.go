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

// TestIsMigNodeWithMigCapableLabel checks that the GFD nvidia.com/mig.capable label decides MIG support,
// with the product family list kept as a fallback when the label is absent.
func TestIsMigNodeWithMigCapableLabel(t *testing.T) {
	tests := []struct {
		name   string
		labels map[string]string
		want   bool
	}{
		{
			name: "mig.capable true with unknown product",
			labels: map[string]string{
				MigCapableLabel: TrueLabelValue, GpuProductLabel: "NVIDIA-RTX-PRO-6000-Blackwell-Server-Edition",
				MigStrategyLabel: MigStrategyMixed, MigEnabledLabel: TrueLabelValue,
			},
			want: true,
		},
		{
			name: "mig.capable false with non MIG product",
			labels: map[string]string{
				MigCapableLabel: "false", GpuProductLabel: "NVIDIA-L40S",
				MigStrategyLabel: MigStrategyMixed, MigEnabledLabel: TrueLabelValue,
			},
			want: false,
		},
		{
			name: "no mig.capable label falls back to product list",
			labels: map[string]string{
				GpuProductLabel:  "NVIDIA-A100-SXM4-80GB",
				MigStrategyLabel: MigStrategySingle, MigEnabledLabel: TrueLabelValue,
			},
			want: true,
		},
		{
			name: "mig.capable true without runai mig enabled label",
			labels: map[string]string{
				MigCapableLabel: TrueLabelValue, MigStrategyLabel: MigStrategyMixed,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &v1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node", Labels: tt.labels}}
			if got := NewMigStrategyExporter(node).IsMigNode(); got != tt.want {
				t.Errorf("IsMigNode() = %v, want %v", got, tt.want)
			}
		})
	}
}
