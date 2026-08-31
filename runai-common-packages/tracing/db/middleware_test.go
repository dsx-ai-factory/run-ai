/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package db

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/tracing"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDBMiddleware(t *testing.T) {
	// Create a new TracerProvider for testing
	tp, err := tracing.Init(&tracing.Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Enabled:           true,
		UnderlineTracer:   "otel",
	})

	// Check if there was an error initializing the TracerProvider
	if err != nil {
		t.Fatalf("Failed to initialize tracer: %v", err)
	}
	// Call the DBMiddleware function
	plugin, ok := DBMiddleware(tp)

	// Check if the plugin is not nil
	if plugin == nil {
		t.Fatal("Expected non-nil gorm.Plugin, but got nil")
	}

	// Ensure the plugin is of the correct type
	_, ok = plugin.(gorm.Plugin)
	if !ok {
		t.Fatalf("Expected gorm.Plugin, but got %T", plugin)
	}
}

func TestDBMiddleware_InvalidTracerProvider(t *testing.T) {
	// Initialize gorm.DB
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Create a new TracerProvider for testing
	tp, err := tracing.Init(&tracing.Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Enabled:           false,
		UnderlineTracer:   "otel",
	})

	if err != nil {
		panic("failed to initialize tracer")
	}

	// Call the DBMiddleware function
	plugin, ok := DBMiddleware(tp)
	if ok {
		db.Use(plugin)
	}
}

func TestDBMiddleware_GetTracingProvider(t *testing.T) {
	// Create a new TracerProvider for testing
	tracing.Init(&tracing.Config{
		ServiceName:       "test-service",
		CollectorEndpoint: "localhost:4317",
		Enabled:           true,
		UnderlineTracer:   "otel",
	})

	tracingProvider := tracing.GetTracerProvider()

	// Make sure tracingProvider is not nil
	if tracingProvider == nil {
		t.Fatal("Expected non-nil TracerProvider, but got nil")
	}

	// Call the DBMiddleware function
	plugin, ok := DBMiddleware(tracingProvider)

	// Check if the plugin is not nil
	if plugin == nil {
		t.Fatal("Expected non-nil gorm.Plugin, but got nil")
	}

	// Ensure the plugin is of the correct type
	_, ok = plugin.(gorm.Plugin)
	if !ok {
		t.Fatalf("Expected gorm.Plugin, but got %T", plugin)
	}
}
