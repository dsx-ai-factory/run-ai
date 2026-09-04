/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package providers

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/metricstores"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/stretchr/testify/require"
)

var datasourceMock = metricstores.Datasource{
	Type: metricstores.DatasourceTypeThanos,
	Read: metricstores.DatasourceEndpoint{
		Url: "http://thanos:9090",
		Auth: &metricstores.DatasourceAuth{
			Type: metricstores.AuthTypeBasic,
			Basic: &metricstores.AuthBasic{
				User:     "user",
				Password: "password",
			},
		},
	},
	Write: metricstores.DatasourceEndpoint{
		Url: "http://thanos:9090/write",
		Auth: &metricstores.DatasourceAuth{
			Type: metricstores.AuthTypeBasic,
			Basic: &metricstores.AuthBasic{
				User:     "user",
				Password: "password",
			},
		},
	},
}

func TestStaticProvider_Create(t *testing.T) {
	cfg := &DatasourceProviderConfig{
		StaticDatasource: datasourceMock,
	}
	config.Set(cfg)
	defer config.FlushAll()
	provider := NewStaticProvider()

	ds, err := provider.Create(context.Background(), "any-datasource-id")

	require.NoError(t, err)
	require.Equal(t, &datasourceMock, ds)
}

func TestStaticProvider_Create_ReturnsSameDatasourceRegardlessOfId(t *testing.T) {
	cfg := &DatasourceProviderConfig{
		StaticDatasource: datasourceMock,
	}
	config.Set(cfg)
	defer config.FlushAll()
	provider := NewStaticProvider()

	ds1, err := provider.Create(context.Background(), "tenant-1")
	require.NoError(t, err)
	ds2, err := provider.Create(context.Background(), "tenant-2")
	require.NoError(t, err)

	require.Equal(t, &datasourceMock, ds1)
	require.Equal(t, &datasourceMock, ds2)
}

func TestStaticProvider_Delete(t *testing.T) {
	provider := NewStaticProvider()

	err := provider.Delete(context.Background(), "any-datasource-id")
	require.NoError(t, err)
}
