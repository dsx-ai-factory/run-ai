/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_identifier

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"
)

func TestGpuPriority(t *testing.T) {
	annotations := map[string]string{
		types.RunaiGpuPriority: "42",
	}

	hasPriority, priority, err := GetSharedGpuPriority(annotations)

	assert.True(t, hasPriority)
	assert.Equal(t, priority, 42)
	assert.Nil(t, err)
}

func TestGpuPriorityMalformed(t *testing.T) {
	annotations := map[string]string{
		types.RunaiGpuPriority: "fourtytwo",
	}

	hasPriority, priority, err := GetSharedGpuPriority(annotations)

	assert.False(t, hasPriority)
	assert.Equal(t, priority, -1)
	assert.NotNil(t, err)
}

func TestNoGpuPriority(t *testing.T) {
	annotations := map[string]string{}

	hasPriority, priority, err := GetSharedGpuPriority(annotations)

	assert.False(t, hasPriority)
	assert.Equal(t, priority, -1)
	assert.Nil(t, err)
}
