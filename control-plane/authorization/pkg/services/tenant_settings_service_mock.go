package services

import (
	"context"
)

// MockTenantSettingsService is a mock implementation of TenantSettingsService for testing
type MockTenantSettingsService struct {
	IsDGXCManagedFunc                func(ctx context.Context) bool
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
		IsDGXCManagedFunc: func(ctx context.Context) bool {
			return getValueOrDefault(CSPSettingKey, true)
		},
		IsScopedAuthorizationEnabledFunc: func(ctx context.Context) bool {
			return getValueOrDefault(EnableScopedAuthorizationKey, true)
		},
	}
}

// IsDGXCManaged calls the mock function
func (m *MockTenantSettingsService) IsDGXCManaged(ctx context.Context) bool {
	return m.IsDGXCManagedFunc(ctx)
}

// IsScopedAuthorizationEnabled calls the mock function
func (m *MockTenantSettingsService) IsScopedAuthorizationEnabled(ctx context.Context) bool {
	return m.IsScopedAuthorizationEnabledFunc(ctx)
}

// SetIsDGXCManaged sets the mock function for IsDGXCManaged
func (m *MockTenantSettingsService) SetIsDGXCManaged(fn func(ctx context.Context) bool) {
	m.IsDGXCManagedFunc = fn
}

// SetIsScopedAuthorizationEnabled sets the mock function for IsScopedAuthorizationEnabled
func (m *MockTenantSettingsService) SetIsScopedAuthorizationEnabled(fn func(ctx context.Context) bool) {
	m.IsScopedAuthorizationEnabledFunc = fn
}
