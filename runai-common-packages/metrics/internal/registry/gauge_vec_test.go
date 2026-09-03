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

func TestGaugeVec_RegistryTwiceReturnsSame(t *testing.T) {
	prm := mocks.NewRegisterer(t)
	prm.On("MustRegister", mock.Anything)

	registry := NewGaugeVecRegistry(prm)
	m1 := registry.Register("test_gauge", "description", make([]string, 0))
	m2 := registry.Register("test_gauge", "description", make([]string, 0))
	assert.Same(t, m1, m2)
}

func TestGaugeVec_GetFromRegistryNotNilWhenRegistered(t *testing.T) {
	prm := mocks.NewRegisterer(t)
	prm.On("MustRegister", mock.Anything)

	registry := NewGaugeVecRegistry(prm)
	registry.Register("test_gauge", "description", make([]string, 0))
	metric := registry.Get("test_gauge")
	assert.NotNil(t, metric)
}

func TestGaugeVec_GetFromRegistryNotNilWithoutRegister(t *testing.T) {
	prm := mocks.NewRegisterer(t)

	registry := NewGaugeVecRegistry(prm)
	metric := registry.Get("test_gauge")
	assert.Equal(t, &DummyGaugeVector{}, metric)
}
