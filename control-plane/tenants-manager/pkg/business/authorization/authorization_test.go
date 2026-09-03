/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package business

import (
	"context"
	"errors"
	"net/http"
	"testing"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

const (
	tenantIdFromContext = "111"
	tenantIdFromParam   = "222"
)

func TestNewAuthManager(t *testing.T) {
	authManager := NewAuthManager()
	assert.NotNil(t, authManager)
}

func TestAuthManager_GetTenantIdForSystemScope(t *testing.T) {
	authManager := &AuthManager{}

	t.Run("empty tenantId - returns tenant from context", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantIdFromContext)

		result, err := authManager.GetTenantIdForSystemScope(ctx, "")

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, tenantIdFromContext, *result)
	})

	t.Run("empty tenantId - no tenant in context", func(t *testing.T) {
		ctx := context.Background()

		result, err := authManager.GetTenantIdForSystemScope(ctx, "")

		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("provided tenantId - management tenant with system scope", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantIdFromContext)
		ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

		permittedScopes := &authorization_api.PermittedScopes{
			System: ptr.To(true),
		}
		ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, permittedScopes)

		result, err := authManager.GetTenantIdForSystemScope(ctx, tenantIdFromParam)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, tenantIdFromParam, *result)
	})

	t.Run("provided tenantId - not management tenant", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "false")

		result, err := authManager.GetTenantIdForSystemScope(ctx, tenantIdFromParam)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unauthorized to manage this resource")
	})

	t.Run("provided tenantId - management tenant without system scope", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

		permittedScopes := &authorization_api.PermittedScopes{
			System: ptr.To(false),
		}
		managementCtx := context.WithValue(ctx, authorization_utils.PermittedScopesKey, permittedScopes)

		result, err := authManager.GetTenantIdForSystemScope(managementCtx, tenantIdFromParam)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unauthorized to manage this resource")
	})

	t.Run("provided tenantId - management tenant with no permitted scopes in context", func(t *testing.T) {
		ctx := context.Background()
		ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

		result, err := authManager.GetTenantIdForSystemScope(ctx, tenantIdFromParam)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unauthorized to manage this resource")
	})
}

func TestAuthManager_HandleSystemScopeError(t *testing.T) {
	authManager := &AuthManager{}

	t.Run("unauthorized error", func(t *testing.T) {
		inputError := errors.New("Unauthorized: access denied")
		errMessage := "Failed to access resource"

		response, err := authManager.HandleSystemScopeError(inputError, errMessage)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized error. Failed to access resource")
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	})

	t.Run("internal server error", func(t *testing.T) {
		inputError := errors.New("database connection failed")
		errMessage := "Internal processing error"

		response, err := authManager.HandleSystemScopeError(inputError, errMessage)

		assert.Error(t, err)
		assert.Equal(t, "Internal processing error", err.Error())
		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})

	t.Run("error with partial unauthorized text", func(t *testing.T) {
		inputError := errors.New("some error with unauthorized in the middle")
		errMessage := "Some error occurred"

		response, err := authManager.HandleSystemScopeError(inputError, errMessage)

		assert.Error(t, err)
		assert.Equal(t, "Some error occurred", err.Error())
		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})

	t.Run("unauthorized error with different message format", func(t *testing.T) {
		inputError := errors.New("Unauthorized to perform this action")
		errMessage := "Permission denied"

		response, err := authManager.HandleSystemScopeError(inputError, errMessage)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized error. Permission denied")
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	})

	t.Run("empty error message", func(t *testing.T) {
		inputError := errors.New("Unauthorized")
		errMessage := ""

		response, err := authManager.HandleSystemScopeError(inputError, errMessage)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized error. ")
		assert.Equal(t, http.StatusUnauthorized, response.Code)
	})
}
