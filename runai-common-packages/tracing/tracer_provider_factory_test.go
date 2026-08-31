/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tracing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFactory_DummyWhenDisabled(t *testing.T) {
	factory := TracerProviderFactory{}
	config := &Config{
		Enabled: false,
	}

	provider, err := factory.CreateTracerProvider(config)
	assert.Nil(t, err)
	assert.IsType(t, &DummyTracerProvider{}, provider)
}

func TestFactory_DummyWhenUnknown(t *testing.T) {
	factory := TracerProviderFactory{}
	config := &Config{
		Enabled:         true,
		UnderlineTracer: "test",
	}

	provider, err := factory.CreateTracerProvider(config)
	assert.Nil(t, err)
	assert.IsType(t, &DummyTracerProvider{}, provider)
}

func TestFactory_OtelWhenOtel(t *testing.T) {
	factory := TracerProviderFactory{}
	config := &Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Enabled:           true,
		UnderlineTracer:   "otel",
	}

	provider, err := factory.CreateTracerProvider(config)
	assert.Nil(t, err)
	assert.IsType(t, &OtelTracerProvider{}, provider)
}
