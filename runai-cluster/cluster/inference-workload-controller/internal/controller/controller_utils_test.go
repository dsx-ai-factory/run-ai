/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"testing"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
)

func TestAddNimWorkloadLabel(t *testing.T) {
	tests := []struct {
		name           string
		initialLabels  map[string]string
		image          string
		expectedLabels map[string]string
	}{
		{
			name: "Existing labels with NIM image but wrong registry",
			initialLabels: map[string]string{
				"app": "test",
			},
			image: "gcr.io/nim/nvidia/llama-3.1-nemotron-70b-instruct:latest",
			expectedLabels: map[string]string{
				"app": "test",
			},
		},
		{
			name: "Existing labels with NIM image",
			initialLabels: map[string]string{
				"app": "test",
			},
			image: "nvcr.io/nim/nvidia/llama-3.1-nemotron-70b-instruct:latest",
			expectedLabels: map[string]string{
				"app":                      "test",
				constants.LabelNIMWorkload: "true",
			},
		},
		{
			name: "Non-NIM image does not add label",
			initialLabels: map[string]string{
				"app": "test",
			},
			image: "nginx:latest",
			expectedLabels: map[string]string{
				"app": "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize DictionaryField from the initial labels
			df := flds.NewDictionaryField(tt.initialLabels)

			AddNimWorkloadLabel(df, tt.image)

			// Convert back to plain map for easy assertion
			got := map[string]string{}
			for k, v := range df.Deref() {
				if v.Value != nil {
					got[k] = *v.Value
				}
			}

			// Check if the resulting labels match expected
			if len(got) != len(tt.expectedLabels) {
				t.Errorf("Expected %d labels, got %d", len(tt.expectedLabels), len(got))
			}

			for key, expectedValue := range tt.expectedLabels {
				if actualValue, exists := got[key]; !exists {
					t.Errorf("Expected label %q not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("Label %q: expected %q, got %q", key, expectedValue, actualValue)
				}
			}

			// Check for unexpected labels
			for key := range got {
				if _, expected := tt.expectedLabels[key]; !expected {
					t.Errorf("Unexpected label %q found with value %q", key, got[key])
				}
			}
		})
	}
}
