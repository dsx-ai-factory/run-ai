/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"

type ExportableMetric struct {
	Metric        metrics_client.Metric
	Name          string
	Description   string
	LabelsMapping map[string]string
}
