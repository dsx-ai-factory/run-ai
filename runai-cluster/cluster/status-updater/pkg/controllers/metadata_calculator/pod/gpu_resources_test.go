/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const a100GpuProductName = "NVIDIA-A100-SXM4-40GB"

func TestGetDeviceComputePortions_A100(t *testing.T) {
	portion, err := getDeviceComputePortions(a100GpuProductName)
	assert.Nil(t, err)
	assert.Equal(t, defaultComputePortions, portion)
}

func TestGetDeviceComputePortions_NotMIG(t *testing.T) {
	_, err := getDeviceComputePortions("Tesla-T4")
	assert.NotNil(t, err)
}
