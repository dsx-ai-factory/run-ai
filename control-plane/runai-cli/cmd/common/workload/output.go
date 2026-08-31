/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
)

func SuccessDeleteMessage(workloadName string, workloadType api.WorkloadType) string {
	return fmt.Sprintf(consts.SuccessDelete, workloadType, workloadName, consts.FollowingWorkspace)
}

func SuccessInferenceDeleteMessage(workloadName string, workloadType api.WorkloadType) string {
	return fmt.Sprintf(consts.SuccessDelete, workloadType, workloadName, consts.FollowingInference)
}

func SuccessResumeMessage(workloadName string, workloadType api.WorkloadType) string {
	return fmt.Sprintf(consts.SuccessResume, workloadType, workloadName, consts.FollowingWorkspace)
}

func SuccessSuspendMessage(workloadName string, workloadType api.WorkloadType) string {
	return fmt.Sprintf(consts.SuccessSuspend, workloadType, workloadName, consts.FollowingWorkspace)
}
