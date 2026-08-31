/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package helmclient

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsHelmWaitTimeout(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "helm wait timeout error",
			err:      fmt.Errorf("uninstallation completed with 1 error(s): context deadline exceeded"),
			expected: true,
		},
		{
			name:     "wrapped helm wait timeout error",
			err:      fmt.Errorf("helm uninstall of 'runai' failed: %w", fmt.Errorf("uninstallation completed with 1 error(s): context deadline exceeded")),
			expected: true,
		},
		{
			name:     "genuine helm failure",
			err:      errors.New("failed to delete configmap: forbidden"),
			expected: false,
		},
		{
			name:     "connection error",
			err:      errors.New("dial tcp: connection refused"),
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isHelmWaitTimeout(tt.err))
		})
	}
}
