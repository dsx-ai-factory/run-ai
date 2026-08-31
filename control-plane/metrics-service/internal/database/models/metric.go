/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"github.com/google/uuid"
)

const MetricsTable = "report_metrics_queries"

type Metric struct {
	ReportID    uuid.UUID
	DisplayName string
	Query       string
}

func (m *Metric) TableName() string {
	return MetricsTable
}
