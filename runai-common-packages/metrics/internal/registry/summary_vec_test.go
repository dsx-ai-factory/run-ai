/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package registry

import (
	"testing"

	"github.com/run-ai/runai/runai-common-packages/metrics/internal/registry/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSummaryVec_RegistryTwiceReturnsSame(t *testing.T) {
	prm := mocks.NewRegisterer(t)
	prm.On("MustRegister", mock.Anything)

	registry := NewSummaryVecRegistry(prm)
	m1 := registry.Register("test_Summary", "description", make([]string, 0), make(map[float64]float64, 0))
	m2 := registry.Register("test_Summary", "description", make([]string, 0), make(map[float64]float64, 0))
	assert.Same(t, m1, m2)
}

func TestSummaryVec_GetFromRegistryNotNilWhenRegistered(t *testing.T) {
	prm := mocks.NewRegisterer(t)
	prm.On("MustRegister", mock.Anything)

	registry := NewSummaryVecRegistry(prm)
	registry.Register("test_Summary", "description", make([]string, 0), make(map[float64]float64, 0))
	metric := registry.Get("test_Summary")
	assert.NotNil(t, metric)
}

func TestSummaryVec_GetFromRegistryNotNilWithoutRegister(t *testing.T) {
	prm := mocks.NewRegisterer(t)

	registry := NewSummaryVecRegistry(prm)
	metric := registry.Get("test_Summary")
	assert.Equal(t, &dummySummaryVector{}, metric)
}
