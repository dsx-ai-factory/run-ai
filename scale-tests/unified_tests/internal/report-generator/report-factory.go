/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package report_generator

import (
	"fmt"
	workloadreportgenerator "github.com/run-ai/runai/scale-tests/unified_tests/internal/workload-scale/aggregated_report"
)

type ReportType string

const (
	WorkloadScale ReportType = "workload_scale"
)

func GetReportGenerator(reportType ReportType) (Generator, error) {
	if reportType == WorkloadScale {
		return workloadreportgenerator.NewWorkloadReportGenerator(), nil
	}
	return nil, fmt.Errorf("unsupported report type: %s", reportType)
}
