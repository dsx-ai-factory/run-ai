// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"testing"

	catalogconfig "github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/pkg/config"
	sharedconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/stretchr/testify/require"
)

func TestMetadataCacheTTLConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name, value string
		want        int
		valid       bool
	}{
		{"default", "", 900, true}, {"override", "60", 60, true},
		{"zero", "0", 0, false}, {"negative", "-1", 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("NGC_METADATA_CACHE_TTL_SECONDS", tc.value)
			sharedconfig.FlushAll()
			t.Cleanup(sharedconfig.FlushAll)
			err := sharedconfig.Register[*catalogconfig.NGC]()
			if !tc.valid {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, sharedconfig.Get[*catalogconfig.NGC]().MetadataCacheTTLSeconds)
		})
	}
}
