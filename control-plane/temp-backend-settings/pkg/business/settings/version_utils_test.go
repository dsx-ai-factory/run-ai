/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsVersionSupported(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		minimumVersion string
		expected       bool
	}{
		{"empty version", "", "2.10", false},
		{"dev version 0.0.0", "0.0.0-dev", "2.10", true},
		{"master version", "1.0.0-master", "2.10", true},
		{"exact match", "2.10", "2.10", true},
		{"above minimum", "2.11", "2.10", true},
		{"below minimum", "2.9", "2.10", false},
		{"major above", "3.0", "2.10", true},
		{"major below", "1.0", "2.10", false},
		{"patch above", "2.10.1", "2.10.0", true},
		{"patch below", "2.10.0", "2.10.1", false},
		{"rc suffix stripped", "2.11-rc1", "2.10", true},
		{"v prefix stripped", "v2.11", "2.10", true},
		{"v prefix and rc stripped", "v2.11-rc3", "2.10", true},
		{"below with rc suffix", "v2.9-rc1", "2.10", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsVersionSupported(tt.version, tt.minimumVersion)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"equal", "2.10.0", "2.10.0", 0},
		{"a greater major", "3.0.0", "2.0.0", 1},
		{"b greater major", "1.0.0", "2.0.0", -1},
		{"a greater minor", "2.11.0", "2.10.0", 1},
		{"b greater minor", "2.9.0", "2.10.0", -1},
		{"a greater patch", "2.10.1", "2.10.0", 1},
		{"b greater patch", "2.10.0", "2.10.1", -1},
		{"different lengths a shorter", "2.10", "2.10.1", -1},
		{"different lengths b shorter", "2.10.1", "2.10", 1},
		{"equal different lengths", "2.10", "2.10.0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareVersions(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseVersionParts(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected []int
	}{
		{"standard three part", "2.10.1", []int{2, 10, 1}},
		{"two part", "2.10", []int{2, 10}},
		{"single part", "5", []int{5}},
		{"non-numeric part defaults to 0", "2.abc.1", []int{2, 0, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersionParts(tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{"plain version", "2.10.0", "2.10.0"},
		{"with v prefix", "v2.10.0", "2.10.0"},
		{"with rc suffix", "2.10.0-rc1", "2.10.0"},
		{"with v prefix and rc suffix", "v2.10.0-rc2", "2.10.0"},
		{"no rc no v", "3.0", "3.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractVersion(tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}
