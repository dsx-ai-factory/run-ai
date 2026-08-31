/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tracing_test

import (
	"context"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/tracing"
	"github.com/stretchr/testify/assert"
)

func TestDummyTracerProvider_Shutdown(t *testing.T) {
	tp := &tracing.DummyTracerProvider{}
	err := tp.Shutdown(context.Background())
	assert.Equal(t, nil, err)
}

func TestDummyTracerProvider_ForceFlush(t *testing.T) {
	tp := &tracing.DummyTracerProvider{}
	err := tp.ForceFlush(context.Background())
	assert.Equal(t, nil, err)
}

func TestDummyTracerProvider_Tracer(t *testing.T) {
	tp := &tracing.DummyTracerProvider{}
	tracer := tp.Tracer()
	assert.IsType(t, &tracing.DummyTracer{}, tracer)
}

func TestDummyTracer_StartSpan(t *testing.T) {
	tracer := &tracing.DummyTracer{}
	c, span := tracer.StartSpan(context.Background(), "test")

	assert.IsType(t, context.Background(), c)
	assert.IsType(t, &tracing.DummySpan{}, span)
}

func TestDummyTracer_FromContext(t *testing.T) {
	tracer := &tracing.DummyTracer{}
	span := tracer.FromContext(context.Background())

	assert.IsType(t, &tracing.DummySpan{}, span)
}

func TestDummySpan_EndNotPanic(t *testing.T) {
	span := &tracing.DummySpan{}

	assert.NotPanics(t, span.End)
}

func TestDummySpan_SetAttributes(t *testing.T) {
	span := &tracing.DummySpan{}
	s := span.SetAttribute("test", "test")
	assert.Equal(t, span, s)
}
