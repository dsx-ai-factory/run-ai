/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package providers

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/metricstores"
)

type DatasourceProviderType string

const (
	DPTGrafanaLabs DatasourceProviderType = "grafanalabs"
	DPTStatic      DatasourceProviderType = "static"
)

type DatasourceProviderAuth struct {
	Token string `mapstructure:"token" config:"secret"`
}

type DatasourceProviderConfig struct {
	Type             DatasourceProviderType  `mapstructure:"datasource_type" validate:"required"`
	Auth             DatasourceProviderAuth  `mapstructure:"datasource_auth"`
	StaticDatasource metricstores.Datasource `mapstructure:"static_datasource"`
}

func (c *DatasourceProviderConfig) SetDefaults() {
	c.Type = DPTStatic
	c.Auth = DatasourceProviderAuth{}
	c.StaticDatasource = metricstores.Datasource{
		Type: metricstores.DatasourceTypeThanos,
		Read: metricstores.DatasourceEndpoint{
			Url: "http://runai-backend-thanos-query:9090",
		},
		Write: metricstores.DatasourceEndpoint{},
	}
}
