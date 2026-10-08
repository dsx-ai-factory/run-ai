/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validation

import (
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/runtime-rules/pkg/frameworks"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
)

const (
	FrameworkErrorMsg = "Cluster prerequisites were not met for %s framework. Contact your administrator for help"
)

type DistributedValidator struct {
	frameworksStatus map[string]frameworks.DistributedJobFrameworksStatus
}

func NewDistributedValidator(frameworksStatus map[string]frameworks.DistributedJobFrameworksStatus) *DistributedValidator {
	return &DistributedValidator{frameworksStatus: frameworksStatus}
}

func (dv *DistributedValidator) Validate(workload *jobs.JobWorkload) error {
	if workload.Kind != constants.KindDistributedWorkload {
		return nil
	}

	framework := workload.Spec.JobType
	status, available := dv.frameworksStatus[framework]
	if !available || !status.SuspendEnabled {
		return fmt.Errorf(FrameworkErrorMsg, framework) //nolint:staticcheck
	}

	return nil
}
