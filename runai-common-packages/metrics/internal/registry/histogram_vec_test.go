/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package registry

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics/internal/registry/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHistogramVec_RegistryTwiceReturnsSame(t *testing.T) {
	prm := mocks.NewRegisterer(t)
	prm.On("MustRegister", mock.Anything)

	registry := NewHistogramVecRegistry(prm)
	m1 := registry.Register("test_Histogram", "description", make([]string, 0), make([]float64, 0))
	m2 := registry.Register("test_Histogram", "description", make([]string, 0), make([]float64, 0))
	assert.Same(t, m1, m2)
}

func TestHistogramVec_GetFromRegistryNotNilWhenRegistered(t *testing.T) {
	prm := mocks.NewRegisterer(t)
	prm.On("MustRegister", mock.Anything)

	registry := NewHistogramVecRegistry(prm)
	registry.Register("test_Histogram", "description", make([]string, 0), make([]float64, 0))
	metric := registry.Get("test_Histogram")
	assert.NotNil(t, metric)
}

func TestHistogramVec_GetFromRegistryNotNilWithoutRegister(t *testing.T) {
	prm := mocks.NewRegisterer(t)

	registry := NewHistogramVecRegistry(prm)
	metric := registry.Get("test_Histogram")
	assert.Equal(t, &DummyHistogramVector{}, metric)
}
