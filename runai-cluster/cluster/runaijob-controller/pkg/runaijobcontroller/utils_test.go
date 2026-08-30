// /*
// Copyright 2016 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */
package runaijobcontroller

import (
	"testing"

	runaijobv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
)

func TestIsJobFinished(t *testing.T) {
	testCases := map[string]struct {
		conditionType        runaijobv1.JobConditionType
		conditionStatus      v1.ConditionStatus
		expectJobNotFinished bool
	}{
		"RunaiJob is completed and condition is true": {
			runaijobv1.JobComplete,
			v1.ConditionTrue,
			false,
		},
		"RunaiJob is completed and condition is false": {
			runaijobv1.JobComplete,
			v1.ConditionFalse,
			true,
		},
		"RunaiJob is completed and condition is unknown": {
			runaijobv1.JobComplete,
			v1.ConditionUnknown,
			true,
		},
		"RunaiJob is failed and condition is true": {
			runaijobv1.JobFailed,
			v1.ConditionTrue,
			false,
		},
		"RunaiJob is failed and condition is false": {
			runaijobv1.JobFailed,
			v1.ConditionFalse,
			true,
		},
		"RunaiJob is failed and condition is unknown": {
			runaijobv1.JobFailed,
			v1.ConditionUnknown,
			true,
		},
	}

	for name, tc := range testCases {
		job := &runaijobv1.RunaiJob{
			Status: runaijobv1.JobStatus{
				Conditions: []runaijobv1.JobCondition{{
					Type:   tc.conditionType,
					Status: tc.conditionStatus,
				}},
			},
		}

		if tc.expectJobNotFinished == IsJobFinished(job) {
			if tc.expectJobNotFinished {
				t.Errorf("test name: %s, job was not expected to be finished", name)
			} else {
				t.Errorf("test name: %s, job was expected to be finished", name)
			}
		}
	}
}
