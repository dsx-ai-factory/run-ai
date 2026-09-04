/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGBMemoryInMigProduct(t *testing.T) {
	memory, err := GetGBMemoryInMigProduct("PRODUCT-A100-40GB")
	assert.Nil(t, err)
	assert.Equal(t, memory40GBValue, memory)

	memory, err = GetGBMemoryInMigProduct("PRODUCT-A100-80GB")
	assert.Nil(t, err)
	assert.Equal(t, memory80GBValue, memory)

	memory, err = GetGBMemoryInMigProduct("PRODUCT-H100")
	assert.Nil(t, err)
	assert.Equal(t, memory80GBValue, memory)
}

func TestGetGBMemoryInMigProductError(t *testing.T) {
	_, err := GetGBMemoryInMigProduct("PRODUCT-A100")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("unknown device memory for device %s", "PRODUCT-A100"), err.Error())

	_, err = GetGBMemoryInMigProduct("PRODUCT-A100-30GB")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("unknown device memory for device %s", "PRODUCT-A100-30GB"), err.Error())

	_, err = GetGBMemoryInMigProduct("PRODUCT-INVALID")
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("unknown device memory for device %s", "PRODUCT-INVALID"), err.Error())
}

func TestExtractGpuAndMemoryFromMigDeviceName(t *testing.T) {
	gpu, memory, err := ExtractGpuAndMemoryFromMigResourceName("nvidia.com/mig-1g.5gb")
	assert.Nil(t, err)
	assert.Equal(t, 1, gpu)
	assert.Equal(t, 5, memory)
}

func TestExtractGpuAndMemoryFromMigDeviceNameError(t *testing.T) {
	invalidDeviceName := "nvidia.com/mig-1g.5gb111"
	gpu, memory, err := ExtractGpuAndMemoryFromMigResourceName(invalidDeviceName)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf("failed to extract gpu/memory from %s", invalidDeviceName), err.Error())
	assert.Equal(t, -1, gpu)
	assert.Equal(t, -1, memory)
}
