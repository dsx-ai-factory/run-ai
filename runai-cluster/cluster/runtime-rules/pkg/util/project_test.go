/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package util

import (
	"testing"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func kaiProjectWithLimits() *kaiorg.Project {
	return &kaiorg.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: "proj-a",
			Annotations: map[string]string{
				commonconstants.InteractiveJobTimeLimitSecondsAnnotation:                  "10",
				commonconstants.InteractiveJobMaxIdleDurationSecondsAnnotation:            "20",
				commonconstants.InteractivePreemptibleJobMaxIdleDurationSecondsAnnotation: "30",
				commonconstants.TrainingJobTimeLimitSecondsAnnotation:                     "40",
				commonconstants.TrainingJobMaxIdleDurationSecondsAnnotation:               "50",
			},
		},
	}
}

func TestKaiProjectLimitAccessors(t *testing.T) {
	project := kaiProjectWithLimits()

	cases := map[string]struct {
		got  int64
		want int64
	}{
		"InteractiveJobTimeLimitSecs":                  {InteractiveJobTimeLimitSecs(project), 10},
		"InteractiveJobMaxIdleDurationSecs":            {InteractiveJobMaxIdleDurationSecs(project), 20},
		"InteractivePreemptibleJobMaxIdleDurationSecs": {InteractivePreemptibleJobMaxIdleDurationSecs(project), 30},
		"TrainingJobTimeLimitSecs":                     {TrainingJobTimeLimitSecs(project), 40},
		"TrainingJobMaxIdleDurationSecs":               {TrainingJobMaxIdleDurationSecs(project), 50},
	}
	for name, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", name, c.got, c.want)
		}
	}
}

func TestKaiProjectHasLimits(t *testing.T) {
	if !ProjectHasLimits(kaiProjectWithLimits()) {
		t.Error("ProjectHasLimits = false, want true for a project with limit annotations")
	}

	// No annotations -> no limits.
	if ProjectHasLimits(&kaiorg.Project{ObjectMeta: metav1.ObjectMeta{Name: "empty"}}) {
		t.Error("ProjectHasLimits = true, want false for an annotation-less project")
	}

	// Unparseable value -> treated as 0 (no panic, no limit).
	bad := &kaiorg.Project{ObjectMeta: metav1.ObjectMeta{
		Name:        "bad",
		Annotations: map[string]string{commonconstants.TrainingJobTimeLimitSecondsAnnotation: "not-a-number"},
	}}
	if TrainingJobTimeLimitSecs(bad) != 0 {
		t.Errorf("TrainingJobTimeLimitSecs = %d, want 0 for unparseable annotation", TrainingJobTimeLimitSecs(bad))
	}
	if ProjectHasLimits(bad) {
		t.Error("ProjectHasLimits = true, want false when the only annotation is unparseable")
	}
}
