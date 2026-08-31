/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ingest

import (
	"testing"

	"github.com/run-ai/runai/runai-application-packages/dbmodels"
	"github.com/stretchr/testify/require"
)

func TestConvertToJsonBMap_DifferentScenarios(t *testing.T) {
	tests := []struct {
		name        string
		logMetadata map[string]string
		want        dbmodels.JsonBMap
	}{
		{
			name:        "Basic case",
			logMetadata: map[string]string{"key1": "value1", "key2": "value2"},
			want:        dbmodels.JsonBMap{"key1": "value1", "key2": "value2"},
		},
		{
			name:        "Empty map",
			logMetadata: map[string]string{},
			want:        dbmodels.JsonBMap{},
		},
		{
			name:        "Nil map",
			logMetadata: nil,
			want:        dbmodels.JsonBMap{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertToJsonBMap(tt.logMetadata)
			require.Equal(t, got, tt.want)
		})
	}
}
