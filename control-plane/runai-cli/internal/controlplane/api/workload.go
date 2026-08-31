/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
)

type Workload struct {
	Type      api.WorkloadType
	Name      string
	Project   string
	ClusterID string
	ID        string
	Framework api.Framework
}
