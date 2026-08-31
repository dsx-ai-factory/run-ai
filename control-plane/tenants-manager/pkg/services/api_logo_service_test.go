/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/tenants-manager/generated/pkg/api"
	authMock "github.com/run-ai/runai/control-plane/tenants-manager/pkg/business/authorization/mocks"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/dao"
	dm "github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/dao/mocks"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/utils"
	um "github.com/run-ai/runai/control-plane/tenants-manager/pkg/utils/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const tenantID = "1111"

type logoServiceMockManager struct {
	logoDaoMock       *dm.Logo
	logoValidatorMock *um.LogoValidator
	serviceUtilsMock  *um.ErrorUtilser
	authMock          *authMock.AuthManagerInterface
}

func initLogoServiceTest(t *testing.T) (mm *logoServiceMockManager, serviceTestSubject *LogoApiService) {
	mm = &logoServiceMockManager{
		logoDaoMock:       dm.NewLogo(t),
		logoValidatorMock: um.NewLogoValidator(t),
		serviceUtilsMock:  um.NewErrorUtilser(t),
		authMock:          authMock.NewAuthManagerInterface(t),
	}

	serviceTestSubject = &LogoApiService{
		logoDao:       mm.logoDaoMock,
		logoValidator: mm.logoValidatorMock,
		serviceUtils:  mm.serviceUtilsMock,
		authManager:   mm.authMock,
	}
	return
}

func TestUploadLogo_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initLogoServiceTest(t)

	emptyLogo := api.LogoCreationRequest{}
	emptyByte := make([]byte, 128)
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantID)

	mm.authMock.On("GetTenantIdForSystemScope", ctx, "").Return(ptr.To(tenantID), nil)
	mm.logoValidatorMock.On("DecodeLogo", emptyLogo.Logo).Return(&emptyByte, nil)
	mm.logoValidatorMock.On("ValidateLogoSize", &emptyByte).Return(nil)
	mm.logoValidatorMock.On("ValidateMimeType", &emptyByte).Return(nil)
	mm.logoDaoMock.On("SetByTenantId", ctx, mock.Anything, &emptyByte).Return(nil)

	response, err := serviceTestSubject.UploadTenantLogo(ctx, "", emptyLogo)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestUploadLogo_InvalidLogo(t *testing.T) {
	mm, serviceTestSubject := initLogoServiceTest(t)

	emptyLogo := api.LogoCreationRequest{}
	emptyByte := make([]byte, 128)
	someError := errors.New("some error")
	utilsError := &utils.Error{Code: 1, Message: "some error", Err: someError}
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantID)

	mm.authMock.On("GetTenantIdForSystemScope", ctx, "").Return(ptr.To(tenantID), nil)
	mm.logoValidatorMock.On("DecodeLogo", emptyLogo.Logo).Return(&emptyByte, utilsError)

	response, err := serviceTestSubject.UploadTenantLogo(ctx, "", emptyLogo)
	assert.Equal(t, utilsError, err)
	assert.Equal(t, utilsError.Code, response.Code)
	assert.Equal(t, utilsError.Message, response.Body)
}

func TestUploadLogo_DbUpdateFailed(t *testing.T) {
	mm, serviceTestSubject := initLogoServiceTest(t)

	emptyLogo := api.LogoCreationRequest{}
	emptyByte := make([]byte, 128)
	daoError := &dao.DaoError{Reason: dao.Conflict}
	apiError := api.Error{Code: 1, Message: "some error"}
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantID)

	mm.authMock.On("GetTenantIdForSystemScope", ctx, "").Return(ptr.To(tenantID), nil)
	mm.logoValidatorMock.On("DecodeLogo", emptyLogo.Logo).Return(&emptyByte, nil)
	mm.logoValidatorMock.On("ValidateLogoSize", &emptyByte).Return(nil)
	mm.logoValidatorMock.On("ValidateMimeType", &emptyByte).Return(nil)
	mm.logoDaoMock.On("SetByTenantId", ctx, mock.Anything, &emptyByte).Return(daoError)
	mm.serviceUtilsMock.On("TenantErrorToResponse", daoError).Return(&apiError).Once()

	response, err := serviceTestSubject.UploadTenantLogo(ctx, "", emptyLogo)
	assert.Equal(t, err, daoError)
	assert.Equal(t, int(apiError.Code), response.Code)
	assert.Equal(t, apiError.Message, response.Body.(*api.Error).Message)
}

func TestGetLogo_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initLogoServiceTest(t)

	logoData := []byte("test")
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantID)

	mm.authMock.On("GetTenantIdForSystemScope", ctx, "").Return(ptr.To(tenantID), nil)
	mm.logoDaoMock.On("GetByTenantId", ctx, mock.Anything).Return(&logoData, nil)

	response, err := serviceTestSubject.GetTenantLogo(ctx, "")
	assert.Nil(t, err)
	dataUrl := response.Body.(api.GetTenantLogo200Response).Logo
	parts := strings.Split(dataUrl, ",")
	assert.Equal(t, 2, len(parts), "Data URL should have base64 part after comma")
	result, err := base64.StdEncoding.DecodeString(parts[1])
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, logoData, result)
}

func TestGetLogo_DbGetFailed(t *testing.T) {
	mm, serviceTestSubject := initLogoServiceTest(t)

	daoError := &dao.DaoError{Reason: dao.General}
	apiError := api.Error{Code: 1, Message: "some error"}
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantID)

	mm.authMock.On("GetTenantIdForSystemScope", ctx, mock.Anything).Return(ptr.To(tenantID), nil)
	mm.logoDaoMock.On("GetByTenantId", ctx, mock.Anything).Return(nil, daoError)
	mm.serviceUtilsMock.On("TenantErrorToResponse", daoError).Return(&apiError).Once()

	response, err := serviceTestSubject.GetTenantLogo(ctx, "")
	assert.Equal(t, err, daoError)
	assert.Equal(t, int(apiError.Code), response.Code)
	assert.Equal(t, apiError.Message, response.Body.(*api.Error).Message)
}

func TestGetLogo_AuthError_Unauthorized(t *testing.T) {
	mm, serviceTestSubject := initLogoServiceTest(t)

	tenantIdParam := "222"
	authError := errors.New("Unauthorized to manage this resource")
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantID)

	mm.authMock.On("GetTenantIdForSystemScope", ctx, tenantIdParam).Return(nil, authError)
	mm.authMock.On("HandleSystemScopeError", authError, "failed").Return(
		api.Response(http.StatusUnauthorized, &api.Error{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized error. failed",
		}),
		errors.New("unauthorized error. failed"),
	)

	response, err := serviceTestSubject.GetTenantLogo(ctx, tenantIdParam)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unauthorized error")
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestGetLogo_AuthError_InternalServerError(t *testing.T) {
	mm, serviceTestSubject := initLogoServiceTest(t)

	tenantIdParam := "222"
	authError := errors.New("some internal error")
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantID)

	mm.authMock.On("GetTenantIdForSystemScope", ctx, tenantIdParam).Return(nil, authError)
	mm.authMock.On("HandleSystemScopeError", authError, "failed").Return(
		api.Response(http.StatusInternalServerError, &api.Error{
			Code:    http.StatusInternalServerError,
			Message: "failed",
		}),
		errors.New("failed"),
	)

	response, err := serviceTestSubject.GetTenantLogo(ctx, tenantIdParam)
	assert.NotNil(t, err)
	assert.Equal(t, "failed", err.Error())
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetLogo_InvalidTenantIdConversion(t *testing.T) {
	mm, serviceTestSubject := initLogoServiceTest(t)

	invalidTenantId := "invalid_id"
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, invalidTenantId)

	mm.authMock.On("GetTenantIdForSystemScope", ctx, "").Return(&invalidTenantId, nil)

	response, err := serviceTestSubject.GetTenantLogo(ctx, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestUploadLogo_AuthError_Unauthorized(t *testing.T) {
	mm, serviceTestSubject := initLogoServiceTest(t)

	tenantIdParam := "222"
	emptyLogo := api.LogoCreationRequest{}
	authError := errors.New("Unauthorized to manage this resource")
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantID)

	mm.authMock.On("GetTenantIdForSystemScope", ctx, tenantIdParam).Return(nil, authError)
	mm.authMock.On("HandleSystemScopeError", authError, "failed").Return(
		api.Response(http.StatusUnauthorized, &api.Error{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized error. failed",
		}),
		errors.New("unauthorized error. failed"),
	)

	response, err := serviceTestSubject.UploadTenantLogo(ctx, tenantIdParam, emptyLogo)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unauthorized error")
	assert.Equal(t, http.StatusUnauthorized, response.Code)
}

func TestUploadLogo_AuthError_InternalServerError(t *testing.T) {
	mm, serviceTestSubject := initLogoServiceTest(t)

	tenantIdParam := "222"
	emptyLogo := api.LogoCreationRequest{}
	authError := errors.New("some internal error")
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantID)

	mm.authMock.On("GetTenantIdForSystemScope", ctx, tenantIdParam).Return(nil, authError)
	mm.authMock.On("HandleSystemScopeError", authError, "failed").Return(
		api.Response(http.StatusInternalServerError, &api.Error{
			Code:    http.StatusInternalServerError,
			Message: "failed",
		}),
		errors.New("failed"),
	)

	response, err := serviceTestSubject.UploadTenantLogo(ctx, tenantIdParam, emptyLogo)
	assert.NotNil(t, err)
	assert.Equal(t, "failed", err.Error())
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestUploadLogo_InvalidTenantIdConversion(t *testing.T) {
	mm, serviceTestSubject := initLogoServiceTest(t)

	emptyLogo := api.LogoCreationRequest{}
	invalidTenantId := "invalid_id"
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, invalidTenantId)

	mm.authMock.On("GetTenantIdForSystemScope", ctx, "").Return(&invalidTenantId, nil)

	response, err := serviceTestSubject.UploadTenantLogo(ctx, "", emptyLogo)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestUploadLogo_LogoSizeValidationFailed(t *testing.T) {
	mm, serviceTestSubject := initLogoServiceTest(t)

	emptyLogo := api.LogoCreationRequest{}
	emptyByte := make([]byte, 128)
	sizeError := errors.New("logo size too large")
	utilsError := &utils.Error{Code: http.StatusRequestEntityTooLarge, Message: "logo size is too big, limit is 128kb", Err: sizeError}
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantID)

	mm.authMock.On("GetTenantIdForSystemScope", ctx, "").Return(ptr.To(tenantID), nil)
	mm.logoValidatorMock.On("DecodeLogo", emptyLogo.Logo).Return(&emptyByte, nil)
	mm.logoValidatorMock.On("ValidateLogoSize", &emptyByte).Return(utilsError)

	response, err := serviceTestSubject.UploadTenantLogo(ctx, "", emptyLogo)
	assert.Equal(t, utilsError, err)
	assert.Equal(t, utilsError.Code, response.Code)
	assert.Equal(t, utilsError.Message, response.Body)
}

func TestUploadLogo_MimeTypeValidationFailed(t *testing.T) {
	mm, serviceTestSubject := initLogoServiceTest(t)

	emptyLogo := api.LogoCreationRequest{}
	emptyByte := make([]byte, 128)
	mimeTypeError := errors.New("unsupported file type")
	utilsError := &utils.Error{Code: http.StatusUnsupportedMediaType, Message: "unsupported file type, only png is supported", Err: mimeTypeError}
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantID)

	mm.authMock.On("GetTenantIdForSystemScope", ctx, "").Return(ptr.To(tenantID), nil)
	mm.logoValidatorMock.On("DecodeLogo", emptyLogo.Logo).Return(&emptyByte, nil)
	mm.logoValidatorMock.On("ValidateLogoSize", &emptyByte).Return(nil)
	mm.logoValidatorMock.On("ValidateMimeType", &emptyByte).Return(utilsError)

	response, err := serviceTestSubject.UploadTenantLogo(ctx, "", emptyLogo)
	assert.Equal(t, utilsError, err)
	assert.Equal(t, utilsError.Code, response.Code)
	assert.Equal(t, utilsError.Message, response.Body)
}
