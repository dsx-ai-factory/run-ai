/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pre_hooks

import (
	"testing"
)

// test function to for compareVersions
func TestCompareVersions(t *testing.T) {
	// Define test cases
	testCases := []struct {
		controlPlaneVersion string
		clusterVersion      string
		expected            bool
	}{
		{"2.21.15", "2.21.15", true},
		{"2.22.1", "2.22.2", true},
		{"2.22.0-alpha", "2.22.0-beta", true},
		{"2.22.10000932-6002de9", "2.22.10001463-run-26838-222-9ffa636", true},
		{"2.22.0", "2.23.0", false},
		{"2.22.0", "3.0.0", false},
		{"0.0.0-master-asda.3", "0.0.0-master-asda.32", true},
	}

	for _, tc := range testCases {
		result := compareVersions(tc.controlPlaneVersion, tc.clusterVersion)
		if result != nil && tc.expected {
			t.Errorf("Expected no error for control plane version %s and cluster version %s, but got error: %v", tc.controlPlaneVersion, tc.clusterVersion, result)
		} else if result == nil && !tc.expected {
			t.Errorf("Expected error for control plane version %s and cluster version %s, but got no error", tc.controlPlaneVersion, tc.clusterVersion)
		}
	}
}
