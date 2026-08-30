package services

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/pubsub/async-events/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	sm "github.com/run-ai/runai/control-plane/authorization/pkg/client/mocks"
	"github.com/run-ai/runai/control-plane/tenants-manager/generated/pkg/api"
	auth_manager_mocks "github.com/run-ai/runai/control-plane/tenants-manager/pkg/business/authorization/mocks"
	tm "github.com/run-ai/runai/control-plane/tenants-manager/pkg/business/tenant/mocks"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"

	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/utils"
	"k8s.io/utils/pointer"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/dao"
	daoMock "github.com/run-ai/runai/control-plane/tenants-manager/pkg/database/dao/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	NAME  = "name"
	EMAIL = "email"
)

func initTenantTest(t *testing.T) (tenantDaoMock *daoMock.Tenant, tenantMock *tm.TenantProvider,
	tenantPublisherMock *mocks.TenantPublishServicer, scopesMock *sm.ScopeValidator, authManagerMock *auth_manager_mocks.AuthManagerInterface, testSubject *TenantApiService,
) {
	tenantDaoMock = daoMock.NewTenant(t)
	tenantMock = tm.NewTenantProvider(t)
	scopesMock = sm.NewScopeValidator(t)
	tenantPublisherMock = mocks.NewTenantPublishServicer(t)
	authManagerMock = auth_manager_mocks.NewAuthManagerInterface(t)

	testSubject = &TenantApiService{
		tenantDao:       tenantDaoMock,
		serviceUtils:    utils.NewErrorUtils(),
		tenantProvider:  tenantMock,
		scopeValidator:  scopesMock,
		tenantPublisher: tenantPublisherMock,
		authManager:     authManagerMock,
	}
	return
}

func TestDeleteTenantById_HappyFlow(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	permittedScopes := &authorization_api.PermittedScopes{
		System: ptr.To(true),
	}
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, permittedScopes)

	_, tenantMock, _, _, _, testSubject := initTenantTest(t)
	tenantMock.On("DeleteTenantById", mock.Anything, mock.Anything, false).Return(pointer.String("tenantId"), nil)
	response, err := testSubject.DeleteTenantById(ctx, 1, false)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
}

func TestHardDeleteTenantById_HappyFlow(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	permittedScopes := &authorization_api.PermittedScopes{
		System: ptr.To(true),
	}
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, permittedScopes)

	_, tenantMock, tenantPublisherMock, _, _, testSubject := initTenantTest(t)
	tenantMock.On("DeleteTenantById", mock.Anything, mock.Anything, true).Return(pointer.String("tenantId"), nil)
	tenantPublisherMock.On("PublishTenantDeleteEvent", mock.Anything, mock.Anything).Return(nil)
	response, err := testSubject.DeleteTenantById(ctx, 1, true)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
}

func TestHardDeleteTenantById_FailedPublishFlow(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	permittedScopes := &authorization_api.PermittedScopes{
		System: ptr.To(true),
	}
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, permittedScopes)

	_, tenantMock, tenantPublisherMock, _, _, testSubject := initTenantTest(t)
	tenantMock.On("DeleteTenantById", mock.Anything, mock.Anything, true).Return(pointer.String("tenantId"), nil)
	tenantPublisherMock.On("PublishTenantDeleteEvent", mock.Anything, mock.Anything).Return(errors.New("test"))
	response, err := testSubject.DeleteTenantById(ctx, 1, true)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
}

func TestDeleteTenantById_500Error(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	permittedScopes := &authorization_api.PermittedScopes{
		System: ptr.To(true),
	}
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, permittedScopes)

	_, tenantMock, _, _, _, testSubject := initTenantTest(t)
	tenantMock.On("DeleteTenantById", mock.Anything, mock.Anything, false).Return(nil, getApiError(http.StatusInternalServerError))
	response, err := testSubject.DeleteTenantById(ctx, 1, false)
	assert.NotNil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetTenantById_HappyFlow(t *testing.T) {
	_, tenantMock, _, _, _, testSubject := initTenantTest(t)
	tenantMock.On("GetTenantById", mock.Anything, mock.Anything).Return(getTenantModel(), nil)
	response, err := testSubject.GetTenantById(context.Background(), 1)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
}

func TestGetTenantById_500Error(t *testing.T) {
	_, tenantMock, _, _, _, testSubject := initTenantTest(t)
	tenantMock.On("GetTenantById", mock.Anything, mock.Anything).Return(nil, getApiError(http.StatusInternalServerError))
	response, err := testSubject.GetTenantById(context.Background(), 1)
	assert.NotNil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateTenant_HappyFlow(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	permittedScopes := &authorization_api.PermittedScopes{
		System: ptr.To(true),
	}
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, permittedScopes)

	_, tenantMock, _, _, _, testSubject := initTenantTest(t)
	tenantMock.On("CreateTenant", mock.Anything, mock.Anything).Return(&api.TenantCreationResponse{}, nil)
	response, err := testSubject.CreateTenant(ctx, api.TenantCreationRequest{
		Name:  NAME,
		Email: EMAIL,
	})
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Nil(t, err)
}

func TestCreateTenant_404Error(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	permittedScopes := &authorization_api.PermittedScopes{
		System: ptr.To(true),
	}
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, permittedScopes)

	_, tenantMock, _, _, _, testSubject := initTenantTest(t)
	tenantMock.On("CreateTenant", mock.Anything, mock.Anything).Return(nil, getApiError(http.StatusBadRequest))
	response, err := testSubject.CreateTenant(ctx, api.TenantCreationRequest{
		Name:  NAME,
		Email: EMAIL,
	})
	assert.NotNil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestCreateTenant_500Error(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	permittedScopes := &authorization_api.PermittedScopes{
		System: ptr.To(true),
	}
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, permittedScopes)

	_, tenantMock, _, _, _, testSubject := initTenantTest(t)
	tenantMock.On("CreateTenant", mock.Anything, mock.Anything).Return(nil, getApiError(http.StatusInternalServerError))
	response, err := testSubject.CreateTenant(ctx, api.TenantCreationRequest{
		Name:  NAME,
		Email: EMAIL,
	})
	assert.NotNil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateTenant_409Error(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	permittedScopes := &authorization_api.PermittedScopes{
		System: ptr.To(true),
	}
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, permittedScopes)

	_, tenantMock, _, _, _, testSubject := initTenantTest(t)
	tenantMock.On("CreateTenant", mock.Anything, mock.Anything).Return(nil, getApiError(http.StatusConflict))
	response, err := testSubject.CreateTenant(ctx, api.TenantCreationRequest{
		Name:  NAME,
		Email: EMAIL,
	})
	assert.NotNil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusConflict, response.Code)
}

func TestCreateTenant_Authorization_NotManagementTenant(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "false")

	permittedScopes := &authorization_api.PermittedScopes{
		System: ptr.To(true),
	}
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, permittedScopes)

	_, _, _, _, authManagerMock, testSubject := initTenantTest(t)

	authManagerMock.On("HandleSystemScopeError", errors.New("unauthorized to manage this resource"), "failed to create tenant").Return(
		api.Response(http.StatusUnauthorized, &api.Error{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized error. failed to create tenant",
		}), errors.New("unauthorized error. failed to create tenant"))

	response, err := testSubject.CreateTenant(ctx, api.TenantCreationRequest{
		Name:  NAME,
		Email: EMAIL,
	})

	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Contains(t, err.Error(), "unauthorized error")
}

func TestCreateTenant_Authorization_NoSystemScope(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	permittedScopes := &authorization_api.PermittedScopes{
		System: ptr.To(false),
	}
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, permittedScopes)

	_, _, _, _, authManagerMock, testSubject := initTenantTest(t)

	authManagerMock.On("HandleSystemScopeError", errors.New("unauthorized to manage this resource"), "failed to create tenant").Return(
		api.Response(http.StatusUnauthorized, &api.Error{
			Code:    http.StatusUnauthorized,
			Message: "unauthorized error. failed to create tenant",
		}), errors.New("unauthorized error. failed to create tenant"))

	response, err := testSubject.CreateTenant(ctx, api.TenantCreationRequest{
		Name:  NAME,
		Email: EMAIL,
	})

	assert.Error(t, err)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.Contains(t, err.Error(), "unauthorized error")
}

func TestGetAllTenants_HappyFlow(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	permittedScopes := &authorization_api.PermittedScopes{
		System: ptr.To(true),
	}
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, permittedScopes)

	_, tenantMock, _, sm, _, testSubject := initTenantTest(t)
	sm.On("HasPermittedScopes", mock.Anything).Return(nil)
	tenantMock.On("GetAllTenants", mock.Anything, mock.Anything, mock.Anything).Return(getTenantArrayModel(), nil)
	response, err := testSubject.GetTenants(ctx, false, "", "")
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
}

func TestGetAllTenants_500Error(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	permittedScopes := &authorization_api.PermittedScopes{
		System: ptr.To(true),
	}
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, permittedScopes)

	_, tenantMock, _, sm, _, testSubject := initTenantTest(t)
	sm.On("HasPermittedScopes", mock.Anything).Return(nil)
	tenantMock.On("GetAllTenants", mock.Anything, mock.Anything, mock.Anything).Return(nil, getApiError(http.StatusInternalServerError))
	response, err := testSubject.GetTenants(ctx, false, "", "")
	assert.NotNil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetAllTenants_EmptyResult_ReturnsEmptyArray(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	permittedScopes := &authorization_api.PermittedScopes{
		System: ptr.To(true),
	}
	ctx = context.WithValue(ctx, authorization_utils.PermittedScopesKey, permittedScopes)

	_, tenantMock, _, sm, _, testSubject := initTenantTest(t)
	sm.On("HasPermittedScopes", mock.Anything).Return(nil)
	emptyTenants := &[]models.Tenant{}
	tenantMock.On("GetAllTenants", mock.Anything, mock.Anything, mock.Anything).Return(emptyTenants, nil)

	response, err := testSubject.GetTenants(ctx, false, "non-existent-name", "")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.NotNil(t, response.Body, "Response body should not be nil when no tenants found")

	tenantsList, ok := response.Body.([]api.Tenant)
	assert.True(t, ok, "Response body should be of type []api.Tenant")
	assert.NotNil(t, tenantsList, "Tenants list should not be nil")
	assert.Empty(t, tenantsList, "Tenants list should be empty")
}

func TestUpdateTenantById_HappyFlow(t *testing.T) {
	_, tenantMock, _, _, _, testSubject := initTenantTest(t)
	alias := "new-alias"
	tenantMock.On("UpdateTenantAlias", mock.Anything, int32(1), &alias).Return(getTenantModel(), nil)
	response, err := testSubject.UpdateTenantById(context.Background(), 1, api.TenantUpdateRequest{Alias: &alias})
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
}

func TestUpdateTenantById_NilAlias(t *testing.T) {
	_, tenantMock, _, _, _, testSubject := initTenantTest(t)
	tenantMock.On("UpdateTenantAlias", mock.Anything, int32(1), (*string)(nil)).Return(getTenantModel(), nil)
	response, err := testSubject.UpdateTenantById(context.Background(), 1, api.TenantUpdateRequest{Alias: nil})
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
}

func TestUpdateTenantById_NotFoundError(t *testing.T) {
	_, tenantMock, _, _, _, testSubject := initTenantTest(t)
	alias := "new-alias"
	tenantMock.On("UpdateTenantAlias", mock.Anything, int32(1), &alias).Return(nil, getApiError(http.StatusNotFound))
	response, err := testSubject.UpdateTenantById(context.Background(), 1, api.TenantUpdateRequest{Alias: &alias})
	assert.NotNil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestUpdateTenantById_ConflictError(t *testing.T) {
	_, tenantMock, _, _, _, testSubject := initTenantTest(t)
	alias := "existing-alias"
	tenantMock.On("UpdateTenantAlias", mock.Anything, int32(1), &alias).Return(nil, getApiError(http.StatusConflict))
	response, err := testSubject.UpdateTenantById(context.Background(), 1, api.TenantUpdateRequest{Alias: &alias})
	assert.NotNil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusConflict, response.Code)
}

func TestUpdateTenantById_500Error(t *testing.T) {
	_, tenantMock, _, _, _, testSubject := initTenantTest(t)
	alias := "new-alias"
	tenantMock.On("UpdateTenantAlias", mock.Anything, int32(1), &alias).Return(nil, getApiError(http.StatusInternalServerError))
	response, err := testSubject.UpdateTenantById(context.Background(), 1, api.TenantUpdateRequest{Alias: &alias})
	assert.NotNil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

// ---------------------------------------------------------------------------
// GetTenantById — diagnostics from provisioning_runs
// ---------------------------------------------------------------------------

func initTenantTestWithRuns(t *testing.T) (*daoMock.Tenant, *daoMock.ProvisioningRun, *tm.TenantProvider, *TenantApiService) {
	tenantDaoMock := daoMock.NewTenant(t)
	runDaoMock := daoMock.NewProvisioningRun(t)
	tenantMock := tm.NewTenantProvider(t)
	testSubject := &TenantApiService{
		tenantDao:          tenantDaoMock,
		provisioningRunDao: runDaoMock,
		serviceUtils:       utils.NewErrorUtils(),
		tenantProvider:     tenantMock,
		scopeValidator:     sm.NewScopeValidator(t),
		tenantPublisher:    mocks.NewTenantPublishServicer(t),
		authManager:        auth_manager_mocks.NewAuthManagerInterface(t),
	}
	return tenantDaoMock, runDaoMock, tenantMock, testSubject
}

func failedTenantModel() *models.Tenant {
	return &models.Tenant{
		Name:   NAME,
		Status: models.TenantFailed,
	}
}

func TestGetTenantById_FailedTenant_ReturnsDiagnosticsFromLatestRun(t *testing.T) {
	_, runDaoMock, tenantMock, testSubject := initTenantTestWithRuns(t)

	tenantMock.On("GetTenantById", mock.Anything, int32(1)).Return(failedTenantModel(), nil)
	runDaoMock.On("GetLatestForTenant", mock.Anything, uint(1)).Return(
		&models.ProvisioningRun{Diagnostics: "metric_store: connection refused"},
		nil,
	)

	response, err := testSubject.GetTenantById(context.Background(), 1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	tenant := response.Body.(api.Tenant)
	assert.Equal(t, "metric_store: connection refused", tenant.Diagnostics)
}

func TestGetTenantById_FailedTenant_NoRuns_NoDiagnostics(t *testing.T) {
	_, runDaoMock, tenantMock, testSubject := initTenantTestWithRuns(t)

	tenantMock.On("GetTenantById", mock.Anything, int32(1)).Return(failedTenantModel(), nil)
	runDaoMock.On("GetLatestForTenant", mock.Anything, uint(1)).Return(nil, &dao.DaoError{Reason: dao.NotFound})

	response, err := testSubject.GetTenantById(context.Background(), 1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	tenant := response.Body.(api.Tenant)
	assert.Empty(t, tenant.Diagnostics)
}

func TestGetTenantById_ReadyTenant_NoDiagnosticsLookup(t *testing.T) {
	_, runDaoMock, tenantMock, testSubject := initTenantTestWithRuns(t)

	tenantMock.On("GetTenantById", mock.Anything, int32(1)).Return(getTenantModel(), nil)

	response, err := testSubject.GetTenantById(context.Background(), 1)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	tenant := response.Body.(api.Tenant)
	assert.Empty(t, tenant.Diagnostics)
	runDaoMock.AssertNotCalled(t, "GetLatestForTenant")
}
