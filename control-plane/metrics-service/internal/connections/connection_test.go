/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package connections

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/generated/pkg/api"

	"github.com/stretchr/testify/assert"
)

func TestNewConnection_NoAuth(t *testing.T) {
	metricStore := api.MetricStoreEndpoint{
		Url:     "url",
		Headers: map[string]string{"key": "value"},
	}
	connection, err := newConnection(metricStore)
	assert.NoError(t, err)
	assert.Equal(t, "url", connection.Url)
	assert.Equal(t, "value", connection.Headers["key"])
	assert.IsType(t, &EmptyAuthConfigurator{}, connection.AuthConfigurator)
}

func TestNewConnection_BasicAuth(t *testing.T) {
	metricStore := api.MetricStoreEndpoint{
		Url:     "url",
		Headers: map[string]string{"key": "value"},
		Auth: &api.MetricStoreEndpointAuth{
			Basic: &api.MetricStoreEndpointAuthBasic{
				Username: "user",
				Password: "pass",
			},
			Type: api.METRICSTOREENDPOINTAUTHTYPE_BASIC,
		},
	}
	connection, err := newConnection(metricStore)
	assert.NoError(t, err)
	assert.Equal(t, "url", connection.Url)
	assert.Equal(t, "value", connection.Headers["key"])
	assert.IsType(t, &BasicAuthConfigurator{}, connection.AuthConfigurator)
	assert.Equal(t, "user", connection.AuthConfigurator.(*BasicAuthConfigurator).Username)
	assert.Equal(t, "pass", connection.AuthConfigurator.(*BasicAuthConfigurator).Password)
}

func TestNewConnection_BasicAuthNilUsername(t *testing.T) {
	metricStore := api.MetricStoreEndpoint{
		Url:     "url",
		Headers: map[string]string{"key": "value"},
		Auth: &api.MetricStoreEndpointAuth{
			Basic: &api.MetricStoreEndpointAuthBasic{
				Username: "",
				Password: "pass",
			},
			Type: api.METRICSTOREENDPOINTAUTHTYPE_BASIC,
		},
	}
	connection, err := newConnection(metricStore)
	assert.NoError(t, err)
	assert.IsType(t, &EmptyAuthConfigurator{}, connection.AuthConfigurator)
}

func TestNewConnection_BasicAuthNilPassword(t *testing.T) {
	metricStore := api.MetricStoreEndpoint{
		Url:     "url",
		Headers: map[string]string{"key": "value"},
		Auth: &api.MetricStoreEndpointAuth{
			Basic: &api.MetricStoreEndpointAuthBasic{
				Username: "user",
				Password: "",
			},
			Type: api.METRICSTOREENDPOINTAUTHTYPE_BASIC,
		},
	}
	connection, err := newConnection(metricStore)
	assert.NoError(t, err)
	assert.IsType(t, &EmptyAuthConfigurator{}, connection.AuthConfigurator)
}

func TestNewConnection_EmptyAuthType(t *testing.T) {
	metricStore := api.MetricStoreEndpoint{
		Url:     "url",
		Headers: map[string]string{"key": "value"},
		Auth:    &api.MetricStoreEndpointAuth{},
	}
	connection, err := newConnection(metricStore)
	assert.NoError(t, err)
	assert.Equal(t, "url", connection.Url)
	assert.Equal(t, "value", connection.Headers["key"])
	assert.IsType(t, &EmptyAuthConfigurator{}, connection.AuthConfigurator)
}
