/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/run-ai/runai/control-plane/bff-service/pkg/authorization/mocks"
	"github.com/run-ai/runai/control-plane/bff-service/pkg/utils"
	"k8s.io/utils/ptr"
)

const (
	realmName = "test-realm"
	tenantID  = "test-tenant-id"
)

func TestSettingsHelper_HandleTenantAuthorization(t *testing.T) {
	tests := []struct {
		name                 string
		tenantId             string
		operationDescription string
		setupMock            func(*mocks.AuthManagerInterface, context.Context, string)
		wantTenantID         *string
		wantResponseNil      bool
		wantErrorNil         bool
		wantHTTPStatus       int
		wantErrorCode        int
	}{
		{
			name:                 "successful authorization",
			tenantId:             "tenant-123",
			operationDescription: "test operation",
			setupMock: func(m *mocks.AuthManagerInterface, ctx context.Context, tenantId string) {
				result := "tenant-123"
				m.On("GetTenantID", ctx, tenantId).Return(&result, nil)
			},
			wantTenantID:    ptr.To("tenant-123"),
			wantResponseNil: true,
			wantErrorNil:    true,
		},
		{
			name:                 "unauthorized error with capital U",
			tenantId:             "tenant-456",
			operationDescription: "test operation",
			setupMock: func(m *mocks.AuthManagerInterface, ctx context.Context, tenantId string) {
				m.On("GetTenantID", ctx, tenantId).Return(nil, errors.New("Unauthorized to manage UI settings"))
			},
			wantTenantID:    nil,
			wantResponseNil: false,
			wantErrorNil:    false,
			wantHTTPStatus:  http.StatusForbidden,
			wantErrorCode:   http.StatusForbidden,
		},
		{
			name:                 "unauthorized error with lowercase u",
			tenantId:             "tenant-999",
			operationDescription: "test operation",
			setupMock: func(m *mocks.AuthManagerInterface, ctx context.Context, tenantId string) {
				m.On("GetTenantID", ctx, tenantId).Return(nil, errors.New("unauthorized: access denied"))
			},
			wantTenantID:    nil,
			wantResponseNil: false,
			wantErrorNil:    false,
			wantHTTPStatus:  http.StatusForbidden,
			wantErrorCode:   http.StatusForbidden,
		},
		{
			name:                 "internal server error",
			tenantId:             "tenant-789",
			operationDescription: "test operation",
			setupMock: func(m *mocks.AuthManagerInterface, ctx context.Context, tenantId string) {
				m.On("GetTenantID", ctx, tenantId).Return(nil, errors.New("some other error"))
			},
			wantTenantID:    nil,
			wantResponseNil: false,
			wantErrorNil:    false,
			wantHTTPStatus:  http.StatusInternalServerError,
			wantErrorCode:   http.StatusInternalServerError,
		},
		{
			name:                 "empty tenant id with successful authorization",
			tenantId:             "",
			operationDescription: "get settings",
			setupMock: func(m *mocks.AuthManagerInterface, ctx context.Context, tenantId string) {
				result := "default-tenant"
				m.On("GetTenantID", ctx, tenantId).Return(&result, nil)
			},
			wantTenantID:    ptr.To("default-tenant"),
			wantResponseNil: true,
			wantErrorNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := mocks.NewAuthManagerInterface(t)
			ctx := context.Background()

			tt.setupMock(mockAuth, ctx, tt.tenantId)

			helper := &SettingsHelper{
				authManager: mockAuth,
			}

			tenantID, response, err := helper.HandleTenantAuthorization(ctx, tt.tenantId, tt.operationDescription)

			if tt.wantTenantID != nil {
				if tenantID == nil {
					t.Errorf("HandleTenantAuthorization() tenantID = nil, want %v", *tt.wantTenantID)
				} else if *tenantID != *tt.wantTenantID {
					t.Errorf("HandleTenantAuthorization() tenantID = %v, want %v", *tenantID, *tt.wantTenantID)
				}
			} else if tenantID != nil {
				t.Errorf("HandleTenantAuthorization() tenantID = %v, want nil", *tenantID)
			}

			if tt.wantResponseNil {
				if response != nil {
					t.Errorf("HandleTenantAuthorization() response = %v, want nil", response)
				}
			} else {
				if response == nil {
					t.Errorf("HandleTenantAuthorization() response = nil, want non-nil")
				} else if response.Code != tt.wantHTTPStatus {
					t.Errorf("HandleTenantAuthorization() response.Code = %v, want %v", response.Code, tt.wantHTTPStatus)
				}
			}

			if tt.wantErrorNil {
				if err != nil {
					t.Errorf("HandleTenantAuthorization() error = %v, want nil", err)
				}
			} else {
				if err == nil {
					t.Errorf("HandleTenantAuthorization() error = nil, want non-nil")
				} else {
					utilsErr, ok := err.(utils.Error)
					if !ok {
						t.Errorf("HandleTenantAuthorization() error is not utils.Error type")
					} else if utilsErr.Code != tt.wantErrorCode {
						t.Errorf("HandleTenantAuthorization() error.Code = %v, want %v", utilsErr.Code, tt.wantErrorCode)
					}
				}
			}
		})
	}
}

func TestSettingsHelper_HandleTenantAuthorization_ErrorMessages(t *testing.T) {
	tests := []struct {
		name                 string
		operationDescription string
		tenantId             string
		mockError            error
		wantErrorContains    string
	}{
		{
			name:                 "unauthorized error message format",
			operationDescription: "get settings",
			tenantId:             "tenant-123",
			mockError:            errors.New("Unauthorized to manage UI settings"),
			wantErrorContains:    "unauthorized to get settings. for tenant id: tenant-123.",
		},
		{
			name:                 "internal error message format",
			operationDescription: "update setting by key: 'enableNim'",
			tenantId:             "tenant-456",
			mockError:            errors.New("database connection failed"),
			wantErrorContains:    "failed to update setting by key: 'enableNim' for tenant id: tenant-456.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuth := mocks.NewAuthManagerInterface(t)
			ctx := context.Background()

			mockAuth.On("GetTenantID", ctx, tt.tenantId).Return(nil, tt.mockError)

			helper := &SettingsHelper{
				authManager: mockAuth,
			}

			_, _, err := helper.HandleTenantAuthorization(ctx, tt.tenantId, tt.operationDescription)

			if err == nil {
				t.Errorf("HandleTenantAuthorization() error = nil, want error containing %q", tt.wantErrorContains)
				return
			}

			utilsErr, ok := err.(utils.Error)
			if !ok {
				t.Errorf("HandleTenantAuthorization() error is not utils.Error type")
				return
			}

			if utilsErr.Message != tt.wantErrorContains {
				t.Errorf("HandleTenantAuthorization() error.Message = %q, want %q", utilsErr.Message, tt.wantErrorContains)
			}
		})
	}
}

func TestIsUnauthorizedError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "unauthorized error with prefix",
			err:  errors.New("Unauthorized to manage UI settings"),
			want: true,
		},
		{
			name: "unauthorized error exact match",
			err:  errors.New("Unauthorized"),
			want: true,
		},
		{
			name: "not unauthorized error",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "error with unauthorized in middle",
			err:  errors.New("this is Unauthorized error"),
			want: false,
		},
		{
			name: "lowercase unauthorized",
			err:  errors.New("unauthorized"),
			want: true,
		},
		{
			name: "lowercase unauthorized with message",
			err:  errors.New("unauthorized: access denied"),
			want: true,
		},
		{
			name: "mixed case unauthorized",
			err:  errors.New("UnAuthorized to perform action"),
			want: true,
		},
		{
			name: "empty error message",
			err:  errors.New(""),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isUnauthorizedError(tt.err); got != tt.want {
				t.Errorf("isUnauthorizedError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSettingsHelper(t *testing.T) {
	helper := NewSettingsHelper()

	if helper == nil {
		t.Error("NewSettingsHelper() returned nil")
	}

	if helper.authManager == nil {
		t.Error("NewSettingsHelper() authManager is nil")
	}
}

func TestSettingsHelper_HandleSettingsError(t *testing.T) {
	tests := []struct {
		name                 string
		err                  error
		tenantId             string
		operationDescription string
		wantHTTPStatus       int
		wantErrorCode        int
		wantErrorMessage     string
	}{
		{
			name:                 "unknown setting key error",
			err:                  errors.New("unknown setting key: invalidKey"),
			tenantId:             "tenant-123",
			operationDescription: "get setting by key: 'invalidKey'",
			wantHTTPStatus:       http.StatusBadRequest,
			wantErrorCode:        http.StatusBadRequest,
			wantErrorMessage:     "unknown setting key: invalidKey",
		},
		{
			name:                 "other error for get operation",
			err:                  errors.New("database connection failed"),
			tenantId:             "tenant-456",
			operationDescription: "get setting by key: 'enableNim'",
			wantHTTPStatus:       http.StatusInternalServerError,
			wantErrorCode:        http.StatusInternalServerError,
			wantErrorMessage:     "failed to get setting by key: 'enableNim' for tenant id: tenant-456.",
		},
		{
			name:                 "other error for update operation",
			err:                  errors.New("failed to save to database"),
			tenantId:             "tenant-789",
			operationDescription: "update setting by key: 'enableNim'",
			wantHTTPStatus:       http.StatusInternalServerError,
			wantErrorCode:        http.StatusInternalServerError,
			wantErrorMessage:     "failed to update setting by key: 'enableNim' for tenant id: tenant-789.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := &SettingsHelper{}

			response, err := helper.HandleSettingsError(tt.err, tt.tenantId, tt.operationDescription)

			if response.Code != tt.wantHTTPStatus {
				t.Errorf("HandleSettingsError() response.Code = %v, want %v", response.Code, tt.wantHTTPStatus)
			}

			if err == nil {
				t.Errorf("HandleSettingsError() error = nil, want non-nil")
				return
			}

			utilsErr, ok := err.(utils.Error)
			if !ok {
				t.Errorf("HandleSettingsError() error is not utils.Error type")
				return
			}

			if utilsErr.Code != tt.wantErrorCode {
				t.Errorf("HandleSettingsError() error.Code = %v, want %v", utilsErr.Code, tt.wantErrorCode)
			}

			if utilsErr.Message != tt.wantErrorMessage {
				t.Errorf("HandleSettingsError() error.Message = %q, want %q", utilsErr.Message, tt.wantErrorMessage)
			}
		})
	}
}
