/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package business

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

type AuthManagerInterface interface {
	GetTenantIdForSystemScope(ctx context.Context, tenantId string) (*string, error)
	HandleSystemScopeError(err error, errMessage string) (api.ImplResponse, error)
}

type AuthManager struct{}

func NewAuthManager() AuthManagerInterface {
	return &AuthManager{}
}

func (a *AuthManager) GetTenantIdForSystemScope(ctx context.Context, tenantId string) (*string, error) {
	if tenantId == "" {
		tenant := runaictx.GetString(ctx, runaictx.CfTenantId)
		return tenant, nil
	}
	if !authorization_utils.HasSystemPermissions(ctx) {
		return nil, errors.New("unauthorized to manage this resource")
	}
	return &tenantId, nil
}

func (a *AuthManager) HandleSystemScopeError(err error, errMessage string) (api.ImplResponse, error) {
	if strings.HasPrefix(err.Error(), "Unauthorized") {
		errMsg := fmt.Sprintf("unauthorized error. %s", errMessage)
		apiError := &api.Error{
			Code:    http.StatusUnauthorized,
			Message: errMsg,
		}
		return api.Response(http.StatusUnauthorized, apiError), errors.New(errMsg)
	}
	apiError := &api.Error{
		Code:    http.StatusInternalServerError,
		Message: errMessage,
	}
	return api.Response(http.StatusInternalServerError, apiError), errors.New(errMessage)
}
