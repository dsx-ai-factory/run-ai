/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tracing

import (
	"context"
	"sync"
)

var (
	tracingRegistry TracerProvider
	syncOnce        sync.Once
)

// InitTracer initializes an OpenTelemetry Tracer and returns a TracerProvider interface.
func Init(config *Config) (TracerProvider, error) {
	var err error = nil
	syncOnce.Do(func() {
		if config == nil {
			config, err = readConfig()
		}
		if err != nil {
			return
		}
		factory := &TracerProviderFactory{}
		tracingRegistry, err = factory.CreateTracerProvider(config)
	})
	return tracingRegistry, err
}

// Shutdown shuts down the tracer provider.
func Shutdown(ctx context.Context) error {
	if tracingRegistry == nil {
		return nil
	}
	return tracingRegistry.Shutdown(ctx)
}

// GetTracer returns an instance of GetTracer.
func GetTracer() Tracer {
	if tracingRegistry == nil {
		return &DummyTracer{}
	}
	return tracingRegistry.Tracer()
}

// GetTracerProvider returns an instance of TracerProvider.
func GetTracerProvider() TracerProvider {
	if tracingRegistry == nil {
		return &DummyTracerProvider{}
	}
	return tracingRegistry
}

// ForceFlush flushes the tracer.
func ForceFlush(ctx context.Context) error {
	if tracingRegistry == nil {
		return nil
	}
	return tracingRegistry.ForceFlush(ctx)
}
