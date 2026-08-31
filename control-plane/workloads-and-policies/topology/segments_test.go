/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package topology_test

import (
	"math"
	"testing"

	"github.com/run-ai/runai/control-plane/workloads-and-policies/topology"
	"github.com/stretchr/testify/require"
)

func TestParseSegmentSize(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		expected  int64
		wantError bool
	}{
		{name: "positive integer", value: "2", expected: 2},
		{name: "leading zero", value: "02", expected: 2},
		{name: "empty", value: "", wantError: true},
		{name: "non-numeric", value: "abc", wantError: true},
		{name: "zero", value: "0", wantError: true},
		{name: "negative", value: "-5", wantError: true},
		{name: "int32 overflow", value: "2147483648", wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := topology.ParseSegmentSize(test.value)
			if test.wantError {
				require.ErrorContains(t, err, "must be a positive integer")
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestSegmentCount(t *testing.T) {
	tests := []struct {
		name           string
		workerReplicas int64
		segmentSize    int64
		expected       int64
		wantError      bool
	}{
		{name: "exact segments", workerReplicas: 4, segmentSize: 2, expected: 2},
		{name: "partial final segment", workerReplicas: 5, segmentSize: 2, expected: 3},
		{name: "segment larger than workers", workerReplicas: 3, segmentSize: 4, expected: 1},
		{name: "maximum segments", workerReplicas: 10_000, segmentSize: 1, expected: 10_000},
		{name: "too many segments", workerReplicas: 10_001, segmentSize: 1, wantError: true},
		{name: "maximum int32 without overflow", workerReplicas: math.MaxInt32, segmentSize: 214_749, expected: 10_000},
		{name: "negative workers", workerReplicas: -1, segmentSize: 1, wantError: true},
		{name: "invalid segment size", workerReplicas: 4, segmentSize: 0, wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := topology.SegmentCount(test.workerReplicas, test.segmentSize)
			if test.wantError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, test.expected, actual)
		})
	}
}
