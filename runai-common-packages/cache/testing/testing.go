/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"reflect"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2"
	cache_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/config"
	"github.com/stretchr/testify/mock"
)

func KeyFromType[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func SetMock(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCacher {
	m := NewMockCacher(t)
	cache.SetMock(m)
	return m
}

type TestingT interface {
	mock.TestingT
	Cleanup(func())
}

func SetLocksMock(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockLocker {
	mockLocker := NewMockLocker(t)
	m := newLocksMockManager(t, mockLocker)
	cache.SetLocksMock(m)
	return mockLocker
}

func EnableCache() {
	_ = cache.Close()
	e := true
	cfg := &cache_config.Cache{
		Type:            cache_config.CTInMemory,
		Enabled:         &e,
		ServiceName:     "test-service",
		MaxStoreSizeMB:  200,
		MaxRecordSizeKB: 2,
		TtlSeconds:      3600,
	}
	_ = cache.Init(cfg)
}

func DisableCache() {
	_ = cache.Close()
	e := false
	cfg := &cache_config.Cache{
		Enabled: &e,
	}
	_ = cache.Init(cfg)
}
