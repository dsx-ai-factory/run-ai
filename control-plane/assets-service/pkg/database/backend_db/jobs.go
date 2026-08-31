/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package backend_db

import (
	"fmt"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
)

const tableNameJobs = "jobs"

// JobsRec - we need some info from jobs table, so we define a partial record with the things we need and leave out the rest
type JobsRec struct {
	WorkloadId string
	Status     string
}

func (JobsRec) TableName() string {
	return fmt.Sprintf("%s.%s", runai_config.Get[*config.ServicesConfig]().BackendSchema, tableNameJobs)
}
