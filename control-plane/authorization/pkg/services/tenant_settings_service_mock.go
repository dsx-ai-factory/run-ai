/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
)

// MockTenantSettingsService is a mock implementation of TenantSettingsService for testing
type MockTenantSettingsService struct {
	IsScopedAuthorizationEnabledFunc func(ctx context.Context) bool
}

// NewMockTenantSettingsService creates a new mock instance
func NewMockTenantSettingsService(values map[string]bool) *MockTenantSettingsService {
	getValueOrDefault := func(key string, defaultValue bool) bool {
		value, ok := values[key]
		if !ok {
			return defaultValue
		}
		return value
	}
	return &MockTenantSettingsService{
		IsScopedAuthorizationEnabledFunc: func(ctx context.Context) bool {
			return getValueOrDefault(EnableScopedAuthorizationKey, true)
		},
	}
}

// IsScopedAuthorizationEnabled calls the mock function
func (m *MockTenantSettingsService) IsScopedAuthorizationEnabled(ctx context.Context) bool {
	return m.IsScopedAuthorizationEnabledFunc(ctx)
}

// SetIsScopedAuthorizationEnabled sets the mock function for IsScopedAuthorizationEnabled
func (m *MockTenantSettingsService) SetIsScopedAuthorizationEnabled(fn func(ctx context.Context) bool) {
	m.IsScopedAuthorizationEnabledFunc = fn
}
