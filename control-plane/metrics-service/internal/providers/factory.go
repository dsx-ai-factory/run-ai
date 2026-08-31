/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package providers

import (
	"github.com/run-ai/runai/control-plane/metrics-service/internal/metricstores"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
)

func NewDatasourceProvider() metricstores.DatasourceProvider {
	cfg := config.Get[*DatasourceProviderConfig]()
	switch cfg.Type {
	case DPTGrafanaLabs:
		return NewGrafanaLabsProvider()
	case DPTStatic:
		return NewStaticProvider()
	default:
		logging.Warnf("unknown datasource provider type %q, defaulting to static", cfg.Type)
		return NewStaticProvider()
	}
}
