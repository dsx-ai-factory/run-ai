/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kai_operator

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestSetDefaultsWhereNeeded(t *testing.T) {
	t.Run("defaults enabled to true", func(t *testing.T) {
		config := &KAIOperator{}

		config.SetDefaultsWhereNeeded()

		require.NotNil(t, config.Enabled)
		require.True(t, *config.Enabled)
	})

	t.Run("preserves explicit false", func(t *testing.T) {
		config := &KAIOperator{Enabled: ptr.To(false)}

		config.SetDefaultsWhereNeeded()

		require.False(t, *config.Enabled)
	})
}
