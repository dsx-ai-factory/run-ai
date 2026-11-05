// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package gpuinventory

import "testing"

func TestResolvePCIDeviceUsesExactUnambiguousNVIDIANames(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		wantError  bool
	}{
		{"NVIDIA-A100-SXM4-40GB", "20b0:10de", false},
		{"NVIDIA-A100-SXM4-80GB", "20b2:10de", false},
		{"NVIDIA-H100-80GB-HBM3", "2330:10de", false},
		{"NVIDIA-H100-PCIe", "2331:10de", false},
		{"NVIDIA-GB200", "2941:10de", false},
		{"NVIDIA-B200", "", true},
		{"Tesla-V100-SXM2-16GB", "", true},
		{"Tesla-V100", "", true},
		{"A100", "", true},
		{"NVIDIA-A100-SXM4-80GB-MIG-1g.10gb", "", true},
		{"NVIDIA-A100-SXM4-80GB-SHARED", "", true},
		{"unreported-new-GPU", "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ResolvePCIDevice(tc.name)
			if got != tc.want || (err != nil) != tc.wantError {
				t.Fatalf("PCI=%q error=%v; want %q error=%v", got, err, tc.want, tc.wantError)
			}
		})
	}
}
