/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flat_workloads

import (
	"reflect"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestFlatToAssetsBasedInferenceFields(t *testing.T) {
	tests := []struct {
		name             string
		defaults         map[string]interface{}
		rules            map[string]interface{}
		expectedDefaults map[string]interface{}
		expectedRules    map[string]interface{}
	}{
		{
			name:             "empty maps",
			defaults:         map[string]interface{}{},
			rules:            map[string]interface{}{},
			expectedDefaults: map[string]interface{}{},
			expectedRules:    map[string]interface{}{},
		},
		{
			name: "autoscaling fields rename in defaults",
			defaults: map[string]interface{}{
				fldNameAutoscaling: map[string]interface{}{
					fldNameFlatAutoscalingMetric:          "cpu",
					fldNameFlatAutoscalingMetricThreshold: "80",
				},
			},
			rules: map[string]interface{}{},
			expectedDefaults: map[string]interface{}{
				fldNameAutoscaling: map[string]interface{}{
					fldNameAssetAutoscalingMetric:          "cpu",
					fldNameAssetAutoscalingMetricThreshold: "80",
				},
			},
			expectedRules: map[string]interface{}{},
		},
		{
			name:     "autoscaling fields rename in rules",
			defaults: map[string]interface{}{},
			rules: map[string]interface{}{
				fldNameAutoscaling: map[string]interface{}{
					fldNameFlatAutoscalingMetric:          "memory",
					fldNameFlatAutoscalingMetricThreshold: "70",
				},
			},
			expectedDefaults: map[string]interface{}{},
			expectedRules: map[string]interface{}{
				fldNameAutoscaling: map[string]interface{}{
					fldNameAssetAutoscalingMetric:          "memory",
					fldNameAssetAutoscalingMetricThreshold: "70",
				},
			},
		},
		{
			name: "serving port access fields restructure in defaults",
			defaults: map[string]interface{}{
				fldNameServingPort: map[string]interface{}{
					fldNameServingPortAuthorizationType: "all",
					fldNameServingPortAuthorizedUsers:   []string{"user1"},
					fldNameServingPortAuthorizedGroups:  []string{"group1"},
					fldNameServingPortLocalAccess:       true,
				},
			},
			rules: map[string]interface{}{},
			expectedDefaults: map[string]interface{}{
				fldNameServingPort: map[string]interface{}{}, // original fields should be removed
				fldNameAssetServingPortAccess: map[string]interface{}{
					fldNameServingPortAuthorizationType: "all",
					fldNameServingPortAuthorizedUsers:   []string{"user1"},
					fldNameServingPortAuthorizedGroups:  []string{"group1"},
					fldNameServingPortLocalAccess:       true,
				},
			},
			expectedRules: map[string]interface{}{},
		},
		{
			name:     "serving port access fields restructure in rules",
			defaults: map[string]interface{}{},
			rules: map[string]interface{}{
				fldNameServingPort: map[string]interface{}{
					fldNameServingPortAuthorizationType: "users",
					fldNameServingPortAuthorizedUsers:   []string{"user2", "user3"},
				},
			},
			expectedDefaults: map[string]interface{}{},
			expectedRules: map[string]interface{}{
				fldNameServingPort: map[string]interface{}{}, // original fields should be removed
				fldNameAssetServingPortAccess: map[string]interface{}{
					fldNameServingPortAuthorizationType: "users",
					fldNameServingPortAuthorizedUsers:   []string{"user2", "user3"},
				},
			},
		},
		{
			name: "mixed fields in defaults and rules",
			defaults: map[string]interface{}{
				fldNameAutoscaling: map[string]interface{}{
					fldNameFlatAutoscalingMetric: "gpu",
				},
			},
			rules: map[string]interface{}{
				fldNameServingPort: map[string]interface{}{
					fldNameServingPortLocalAccess: false,
				},
			},
			expectedDefaults: map[string]interface{}{
				fldNameAutoscaling: map[string]interface{}{
					fldNameAssetAutoscalingMetric: "gpu",
				},
			},
			expectedRules: map[string]interface{}{
				fldNameServingPort: map[string]interface{}{},
				fldNameAssetServingPortAccess: map[string]interface{}{
					fldNameServingPortLocalAccess: false,
				},
			},
		},
		{
			name: "no relevant fields",
			defaults: map[string]interface{}{
				"otherField": "value1",
			},
			rules: map[string]interface{}{
				"anotherField": "value2",
			},
			expectedDefaults: map[string]interface{}{
				"otherField": "value1",
			},
			expectedRules: map[string]interface{}{
				"anotherField": "value2",
			},
		},
		{
			name: "autoscaling exists but no metric fields",
			defaults: map[string]interface{}{
				fldNameAutoscaling: map[string]interface{}{
					"someOtherAutoscalingField": "value",
				},
			},
			rules: map[string]interface{}{},
			expectedDefaults: map[string]interface{}{
				fldNameAutoscaling: map[string]interface{}{
					"someOtherAutoscalingField": "value",
				},
			},
			expectedRules: map[string]interface{}{},
		},
		{
			name: "serving port exists but no auth fields",
			defaults: map[string]interface{}{
				fldNameServingPort: map[string]interface{}{
					"someOtherServingPortField": "value",
				},
			},
			rules: map[string]interface{}{},
			expectedDefaults: map[string]interface{}{
				fldNameServingPort: map[string]interface{}{
					"someOtherServingPortField": "value",
				},
			},
			expectedRules: map[string]interface{}{},
		},
		{
			name: "partial serving port access fields in defaults",
			defaults: map[string]interface{}{
				fldNameServingPort: map[string]interface{}{
					fldNameServingPortAuthorizationType: "groups",
					// fldNameServingPortAuthorizedUsers:   []string{"user1"}, // Missing
					fldNameServingPortAuthorizedGroups: []string{"groupX"},
					// fldNameServingPortLocalAccess:       true, // Missing
				},
			},
			rules: map[string]interface{}{},
			expectedDefaults: map[string]interface{}{
				fldNameServingPort: map[string]interface{}{},
				fldNameAssetServingPortAccess: map[string]interface{}{
					fldNameServingPortAuthorizationType: "groups",
					fldNameServingPortAuthorizedGroups:  []string{"groupX"},
				},
			},
			expectedRules: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create copies to avoid modifying the original test case maps
			currentDefaults := make(map[string]interface{})
			for k, v := range tt.defaults {
				currentDefaults[k] = v // Shallow copy is usually fine for test data
			}
			currentRules := make(map[string]interface{})
			for k, v := range tt.rules {
				currentRules[k] = v // Shallow copy
			}

			flatToAssetsBasedInferenceFields(currentDefaults, currentRules)
			assert.Equal(t, tt.expectedDefaults, currentDefaults, "Defaults mismatch")
			assert.Equal(t, tt.expectedRules, currentRules, "Rules mismatch")
		})
	}

	// in addition, we use these tests to protect again forgetting to include a newly added
	// field under ServingPortAccess to the flatToAssetsBasedInferenceFields function.
	t.Run("verify ServingPortAccess fields", func(t *testing.T) {
		var s api.ServingPortAccess
		sType := reflect.TypeOf(s)
		assert.Equal(t, 4, sType.NumField(), "ServingPortAccess should have 4 fields; if new fields are added, update the code and this test")
	})
}
