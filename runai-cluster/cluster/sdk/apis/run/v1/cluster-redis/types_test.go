/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_redis

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestSetDefaultsWhereNeeded(t *testing.T) {
	t.Run("defaults image to chart-provided valkey", func(t *testing.T) {
		t.Setenv("IMG_THIRD_PARTY_VALKEY_REGISTRY", "docker.io")
		t.Setenv("IMG_THIRD_PARTY_VALKEY_REPOSITORY", "valkey/valkey")
		t.Setenv("IMG_THIRD_PARTY_VALKEY_TAG", "9.0.4")
		config := &ClusterRedis{}

		config.SetDefaultsWhereNeeded(&internal.Internal{
			Global: &internal.Global{ImageRegistry: "nvcr.io/nvidia/runai"},
		}, &global.Global{})

		require.NotNil(t, config.Image)
		require.Equal(t, "docker.io", *config.Image.Registry)
		require.Equal(t, "valkey/valkey", *config.Image.Name)
		require.Equal(t, "9.0.4", *config.Image.Tag)
	})

	t.Run("preserves explicit image override", func(t *testing.T) {
		config := &ClusterRedis{
			Image: &common.Image{
				Registry: ptr.To("registry.example.com"),
				Name:     ptr.To("custom/valkey"),
				Tag:      ptr.To("1.2.3"),
			},
		}

		config.SetDefaultsWhereNeeded(&internal.Internal{}, &global.Global{})

		require.Equal(t, "registry.example.com", *config.Image.Registry)
		require.Equal(t, "custom/valkey", *config.Image.Name)
		require.Equal(t, "1.2.3", *config.Image.Tag)
	})

	t.Run("defaults enabled to true", func(t *testing.T) {
		config := &ClusterRedis{}

		config.SetDefaultsWhereNeeded(&internal.Internal{}, &global.Global{})

		require.NotNil(t, config.Enabled)
		require.True(t, *config.Enabled)
	})

	t.Run("preserves explicit false enabled", func(t *testing.T) {
		config := &ClusterRedis{Enabled: ptr.To(false)}

		config.SetDefaultsWhereNeeded(&internal.Internal{}, &global.Global{})

		require.False(t, *config.Enabled)
	})
}
