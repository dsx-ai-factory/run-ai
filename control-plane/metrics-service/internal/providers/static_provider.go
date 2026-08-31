/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package providers

import (
	"context"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/metricstores"
	"github.com/run-ai/runai/runai-common-packages/config"
)

type StaticProvider struct{}

func NewStaticProvider() *StaticProvider {
	return &StaticProvider{}
}

func (sp *StaticProvider) Create(ctx context.Context, datasourceId string) (*metricstores.Datasource, error) {
	cfg := config.Get[*DatasourceProviderConfig]()
	return &cfg.StaticDatasource, nil
}

func (sp *StaticProvider) Delete(ctx context.Context, datasourceId string) error {
	return nil
}
