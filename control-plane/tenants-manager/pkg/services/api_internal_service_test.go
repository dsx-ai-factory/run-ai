/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/tenants-manager/generated/pkg/api"
	berrors "github.com/run-ai/runai/control-plane/tenants-manager/pkg/business/errors"
	msmocks "github.com/run-ai/runai/control-plane/tenants-manager/pkg/business/metric-store/mocks"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/dao/mocks"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
	um "github.com/run-ai/runai/control-plane/tenants-manager/pkg/utils/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Services tests")
}

func initInternalTest(t *testing.T) (tenantDaoMock *mocks.Tenant, tenantServiceMockManagerMock *um.ErrorUtilser, testSubject *InternalApiService) {
	tenantDaoMock = mocks.NewTenant(t)
	tenantServiceMockManagerMock = um.NewErrorUtilser(t)
	testSubject = &InternalApiService{
		tenantDao:      tenantDaoMock,
		serviceUtils:   tenantServiceMockManagerMock,
		domainTemplate: "http://{{.}}.run.ai",
	}
	return
}

func getDaoError() *dao.DaoError {
	return &dao.DaoError{
		Reason: dao.General,
		Err:    errors.Join(dao.ErrNotFound),
	}
}

func Test_GetTenantGetDomain_HappyFlow(t *testing.T) {
	tenant := getTenantModel()
	tenantMock, _, testSubject := initInternalTest(t)
	tenantMock.On("GetById", mock.Anything, mock.Anything).Return(tenant, nil)
	response, err := testSubject.GetTenantDomain(context.Background(), 1)
	assert.Equal(t, response.Body, "http://name.run.ai")
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
}

func Test_GetTenantGetDomain_500Error(t *testing.T) {
	tenantMock, _, testSubject := initInternalTest(t)
	tenantMock.On("GetById", mock.Anything, mock.Anything).Return(nil, getDaoError())
	response, err := testSubject.GetTenantDomain(context.Background(), 1)
	assert.NotNil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetTenantExtendedMetricStore(t *testing.T) {
	tenant := getTenantModel()
	msServiceMock := msmocks.NewMetricStoreService(t)
	service := &InternalApiService{
		metricsStoreService: msServiceMock,
	}
	msServiceMock.On("GetTenantExtendedMetricStore", context.Background(), tenant.ID).Return(&api.ExtendedMetricStore{
		Type:         "mimir",
		DatasourceId: "id",
		Read: api.MetricStoreEndpoint{
			Url: "url",
		},
		Write: api.MetricStoreEndpoint{
			Url: "url",
		},
	}, nil)

	res, err := service.GetTenantExtendedMetricStore(context.Background(), int32(tenant.ID))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "mimir", res.Body.(*api.ExtendedMetricStore).Type)
	require.Equal(t, "id", res.Body.(*api.ExtendedMetricStore).DatasourceId)
	require.Equal(t, "url", res.Body.(*api.ExtendedMetricStore).Read.Url)
	require.Equal(t, "url", res.Body.(*api.ExtendedMetricStore).Write.Url)
}

func TestGetTenantExtendedMetricStore_Error(t *testing.T) {
	tenant := getTenantModel()
	msServiceMock := msmocks.NewMetricStoreService(t)
	errorUtilserMock := um.NewErrorUtilser(t)

	service := &InternalApiService{
		serviceUtils:        errorUtilserMock,
		metricsStoreService: msServiceMock,
	}
	errorUtilserMock.On("BusinessErrorToApiErr", mock.Anything, mock.Anything).Return(&api.Error{
		Code:    http.StatusInternalServerError,
		Message: "error",
	})
	msServiceMock.On("GetTenantExtendedMetricStore", context.Background(), tenant.ID).Return(nil, &berrors.BusinessError{Reason: berrors.GeneralError, Err: errors.New("error")})

	res, err := service.GetTenantExtendedMetricStore(context.Background(), int32(tenant.ID))
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Equal(t, "error", res.Body.(*api.Error).Message)
}

func Test_GetTenantByNameOrAlias_ByName_HappyFlow(t *testing.T) {
	tenant := getTenantModel()
	tenantMock, _, testSubject := initInternalTest(t)
	tenantMock.On("GetByNameOrAlias", mock.Anything, "name").Return(tenant, nil)

	response, err := testSubject.GetTenantByNameOrAlias(context.Background(), "name")
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	apiTenant := response.Body.(api.Tenant)
	assert.Equal(t, tenant.Name, apiTenant.Name)
}

func Test_GetTenantByNameOrAlias_ByAlias_HappyFlow(t *testing.T) {
	tenant := getTenantModel()
	tenantMock, _, testSubject := initInternalTest(t)
	tenantMock.On("GetByNameOrAlias", mock.Anything, "alias").Return(tenant, nil)

	response, err := testSubject.GetTenantByNameOrAlias(context.Background(), "alias")
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	apiTenant := response.Body.(api.Tenant)
	assert.Equal(t, tenant.Name, apiTenant.Name)
	assert.Equal(t, *tenant.Alias, apiTenant.Alias)
}

func Test_GetTenantByNameOrAlias_NotFound(t *testing.T) {
	tenantMock, errorUtilsMock, testSubject := initInternalTest(t)
	daoErr := &dao.DaoError{
		Reason: dao.NotFound,
		Err:    dao.ErrNotFound,
	}
	tenantMock.On("GetByNameOrAlias", mock.Anything, "unknown").Return(nil, daoErr)
	errorUtilsMock.On("TenantErrorToResponse", daoErr).Return(&api.Error{
		Code:    http.StatusNotFound,
		Message: "Tenant not found",
	})

	response, err := testSubject.GetTenantByNameOrAlias(context.Background(), "unknown")
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusNotFound, response.Code)
	assert.Equal(t, "Tenant not found", response.Body.(*api.Error).Message)
}

func Test_GetTenantByNameOrAlias_InternalError(t *testing.T) {
	tenantMock, errorUtilsMock, testSubject := initInternalTest(t)
	daoErr := &dao.DaoError{
		Reason: dao.General,
		Err:    errors.New("db connection failed"),
	}
	tenantMock.On("GetByNameOrAlias", mock.Anything, "name").Return(nil, daoErr)
	errorUtilsMock.On("TenantErrorToResponse", daoErr).Return(&api.Error{
		Code:    http.StatusInternalServerError,
		Message: "Tenant operation Failed",
	})

	response, err := testSubject.GetTenantByNameOrAlias(context.Background(), "name")
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.Equal(t, "Tenant operation Failed", response.Body.(*api.Error).Message)
}

func Test_GetAllLogos_NoLogos(t *testing.T) {
	logoDaoMock := mocks.NewLogo(t)
	svc := &InternalApiService{logoDao: logoDaoMock}

	logoDaoMock.On("GetAll", mock.Anything, 0, 50).Return(&[]models.Logo{}, nil)

	resp, err := svc.GetAllLogos(context.Background(), 0, 50)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.Code)
	body := resp.Body.(api.TenantLogosResponse)
	require.Empty(t, body.Logos)
}

func Test_GetAllLogos_SinglePage(t *testing.T) {
	logoDaoMock := mocks.NewLogo(t)
	svc := &InternalApiService{logoDao: logoDaoMock}

	svg1 := []byte(`<svg xmlns="http://www.w3.org/2000/svg"></svg>`)
	svg2 := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect/></svg>`)
	logos := []models.Logo{
		{TenantId: 1, LogoPng: &svg1},
		{TenantId: 2, LogoPng: &svg2},
	}
	logoDaoMock.On("GetAll", mock.Anything, 0, 50).Return(&logos, nil)

	resp, err := svc.GetAllLogos(context.Background(), 0, 50)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.Code)
	body := resp.Body.(api.TenantLogosResponse)
	require.Len(t, body.Logos, 2)
	require.Equal(t, int64(1), body.Logos[0].TenantId)
	require.Equal(t, int64(2), body.Logos[1].TenantId)
	require.True(t, strings.HasPrefix(body.Logos[0].Logo, "data:image/svg+xml;base64,"))
}

func Test_GetAllLogos_SecondPage(t *testing.T) {
	logoDaoMock := mocks.NewLogo(t)
	svc := &InternalApiService{logoDao: logoDaoMock}

	svg3 := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><circle/></svg>`)
	logos := []models.Logo{
		{TenantId: 3, LogoPng: &svg3},
	}
	logoDaoMock.On("GetAll", mock.Anything, 2, 2).Return(&logos, nil)

	resp, err := svc.GetAllLogos(context.Background(), 2, 2)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.Code)
	body := resp.Body.(api.TenantLogosResponse)
	require.Len(t, body.Logos, 1)
	require.Equal(t, int64(3), body.Logos[0].TenantId)
}

func Test_GetAllLogos_OffsetBeyondTotal(t *testing.T) {
	logoDaoMock := mocks.NewLogo(t)
	svc := &InternalApiService{logoDao: logoDaoMock}

	logoDaoMock.On("GetAll", mock.Anything, 100, 50).Return(&[]models.Logo{}, nil)

	resp, err := svc.GetAllLogos(context.Background(), 100, 50)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.Code)
	body := resp.Body.(api.TenantLogosResponse)
	require.Empty(t, body.Logos)
}

func Test_GetAllLogos_DaoError(t *testing.T) {
	logoDaoMock := mocks.NewLogo(t)
	svc := &InternalApiService{logoDao: logoDaoMock}

	daoErr := &dao.DaoError{Reason: dao.General, Err: errors.New("db error")}
	logoDaoMock.On("GetAll", mock.Anything, 0, 50).Return((*[]models.Logo)(nil), daoErr)

	resp, err := svc.GetAllLogos(context.Background(), 0, 50)
	require.Error(t, err)
	require.Equal(t, http.StatusInternalServerError, resp.Code)
}
