/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
)

type WorkloadValidator interface {
	Validate(_ context.Context, job *jobs.JobWorkload) error
}
