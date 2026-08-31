/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metricsctr

import (
	"github.com/run-ai/runai/runai-common-packages/metrics"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

func Init() {
	ctrlmetrics.Registry = metrics.GetPrometheusRegistry()
}
