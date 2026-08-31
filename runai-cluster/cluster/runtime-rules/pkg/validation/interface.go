/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validation

import "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"

type Validator interface {
	Validate(workload *jobs.JobWorkload) error
}
