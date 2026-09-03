/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package patch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"
)

func TestRemoveGpuFromLimits(t *testing.T) {
	requirements := &v1.ResourceRequirements{
		Limits: map[v1.ResourceName]resource.Quantity{types.GpuResource: resource.MustParse("1")},
	}
	RemoveGpuResourceFromLimits(requirements)
	assert.Empty(t, requirements.Limits)
}

func TestRemoveGpuFromLimitsNotPresent(t *testing.T) {
	requirements := &v1.ResourceRequirements{
		Limits: map[v1.ResourceName]resource.Quantity{types.RunaiGpuMemory: resource.MustParse("1")},
	}
	RemoveGpuResourceFromLimits(requirements)
	assert.Equal(t, len(requirements.Limits), 1)
}

func TestAddRequest(t *testing.T) {
	requirements := &v1.ResourceRequirements{}
	AddResourcesToRequests(requirements, types.GpuResource, resource.MustParse("1"))

	assert.Equal(t, len(requirements.Requests), 1)
	request, found := requirements.Requests[types.GpuResource]
	assert.True(t, found)
	assert.True(t, request.Equal(resource.MustParse("1")))
}

func TestOverrideRequest(t *testing.T) {
	requirements := &v1.ResourceRequirements{
		Requests: map[v1.ResourceName]resource.Quantity{types.GpuResource: resource.MustParse("1")},
	}
	AddResourcesToRequests(requirements, types.GpuResource, resource.MustParse("2"))

	assert.Equal(t, len(requirements.Requests), 1)
	request, found := requirements.Requests[types.GpuResource]
	assert.True(t, found)
	assert.True(t, request.Equal(resource.MustParse("2")))
}

func TestAddLimit(t *testing.T) {
	requirements := &v1.ResourceRequirements{}
	AddResourcesToLimits(requirements, types.GpuResource, resource.MustParse("1"))

	assert.Equal(t, len(requirements.Limits), 1)
	request, found := requirements.Limits[types.GpuResource]
	assert.True(t, found)
	assert.True(t, request.Equal(resource.MustParse("1")))
}

func TestOverrideLimit(t *testing.T) {
	requirements := &v1.ResourceRequirements{
		Limits: map[v1.ResourceName]resource.Quantity{types.GpuResource: resource.MustParse("1")},
	}
	AddResourcesToLimits(requirements, types.GpuResource, resource.MustParse("2"))

	assert.Equal(t, len(requirements.Limits), 1)
	request, found := requirements.Limits[types.GpuResource]
	assert.True(t, found)
	assert.True(t, request.Equal(resource.MustParse("2")))
}
