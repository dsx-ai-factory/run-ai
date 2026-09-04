/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package util

import (
	"strconv"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
)

// The kai.resources project carries the job time/idle limits as annotations
// (run.ai/*-seconds) rather than dedicated spec fields (as the legacy run.ai project did).
// These accessors read them back as seconds, returning 0 when absent or unparseable
// (matching the writer, which omits zero-valued limits).

func InteractiveJobTimeLimitSecs(project *kaiorg.Project) int64 {
	return projectLimitSeconds(project, commonconstants.InteractiveJobTimeLimitSecondsAnnotation)
}

func InteractiveJobMaxIdleDurationSecs(project *kaiorg.Project) int64 {
	return projectLimitSeconds(project, commonconstants.InteractiveJobMaxIdleDurationSecondsAnnotation)
}

func InteractivePreemptibleJobMaxIdleDurationSecs(project *kaiorg.Project) int64 {
	return projectLimitSeconds(project, commonconstants.InteractivePreemptibleJobMaxIdleDurationSecondsAnnotation)
}

func TrainingJobTimeLimitSecs(project *kaiorg.Project) int64 {
	return projectLimitSeconds(project, commonconstants.TrainingJobTimeLimitSecondsAnnotation)
}

func TrainingJobMaxIdleDurationSecs(project *kaiorg.Project) int64 {
	return projectLimitSeconds(project, commonconstants.TrainingJobMaxIdleDurationSecondsAnnotation)
}

// ProjectHasLimits reports whether the kai project sets any job time/idle limit.
func ProjectHasLimits(project *kaiorg.Project) bool {
	return InteractiveJobMaxIdleDurationSecs(project) > 0 ||
		InteractiveJobTimeLimitSecs(project) > 0 ||
		TrainingJobMaxIdleDurationSecs(project) > 0 ||
		TrainingJobTimeLimitSecs(project) > 0 ||
		InteractivePreemptibleJobMaxIdleDurationSecs(project) > 0
}

func projectLimitSeconds(project *kaiorg.Project, annotationKey string) int64 {
	value, err := strconv.ParseInt(project.Annotations[annotationKey], 10, 64)
	if err != nil {
		return 0
	}
	return value
}
