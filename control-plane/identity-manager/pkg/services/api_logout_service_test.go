/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	iamError "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/errors"
	iamMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const (
	testRealm     = "test-realm"
	testSessionId = "test-session-id"
)

func initLogoutServiceTest(t *testing.T) (*iamMock.MockServicesContainer, *LogoutAPIService, context.Context) {
	mockIamServicesContainer := iamMock.NewMockServicesContainer()
	testSubject := &LogoutAPIService{
		iamServices: *mockIamServicesContainer.ToServicesContainer(),
	}
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, testRealm)
	ctx = context.WithValue(ctx, runaictx.CfSId, testSessionId)
	return mockIamServicesContainer, testSubject, ctx
}

func TestLogout_HappyFlow(t *testing.T) {
	iamMock, testSubject, ctx := initLogoutServiceTest(t)
	iamMock.MockUserService.On("LogoutUserBySessionID", ctx, testRealm, testSessionId).Return(nil)

	response, err := testSubject.Logout(ctx)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, response.Body)
}

func TestLogout_InternalError(t *testing.T) {
	iamMock, testSubject, ctx := initLogoutServiceTest(t)
	iamMock.MockUserService.On("LogoutUserBySessionID", ctx, testRealm, testSessionId).Return(iamError.ErrInternal)

	response, err := testSubject.Logout(ctx)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Nil(t, response.Body)
}

func TestLogout_UserNotFound(t *testing.T) {
	iamMock, testSubject, ctx := initLogoutServiceTest(t)
	iamMock.MockUserService.On("LogoutUserBySessionID", ctx, testRealm, testSessionId).Return(iamError.ErrNotFound)

	response, err := testSubject.Logout(ctx)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Nil(t, response.Body)
}

func TestLogout_MissingSessionId(t *testing.T) {
	_, testSubject, _ := initLogoutServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, testRealm)

	response, err := testSubject.Logout(ctx)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Nil(t, response.Body)
}
