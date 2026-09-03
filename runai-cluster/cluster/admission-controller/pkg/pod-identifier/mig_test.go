/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_identifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasMigDeviceRequest(t *testing.T) {
	annotations := map[string]string{
		"nvidia.com/mig": "3g.20gb",
	}

	assert.True(t, HasMigDeviceRequest(annotations))
}

func TestDoesntHaveMigDeviceRequest(t *testing.T) {
	annotations := map[string]string{}
	assert.False(t, HasMigDeviceRequest(annotations))
}
