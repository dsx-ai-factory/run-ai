/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package report_generator

import (
	"github.com/dsx-ai-factory/run-ai/scale-tests/sdk/reports"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/scale"
)

type Generator interface {
	GenerateReport(testResults *scale.SamplesReport) (*reports.SuiteReport, error)
}
