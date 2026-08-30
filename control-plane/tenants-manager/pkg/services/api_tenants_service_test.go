package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/pubsub/async-events/mocks"

	iamMock "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/mocks"
	"k8s.io/utils/pointer"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/tenants-manager/generated/pkg/api"
	tm "github.com/run-ai/runai/control-plane/tenants-manager/pkg/business/tenant/mocks"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
	um "github.com/run-ai/runai/control-plane/tenants-manager/pkg/utils/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type tenantServiceMockManager struct {
	serviceUtilsMock    *um.ErrorUtilser
	iamMock             *iamMock.IIdm
	tenantMock          *tm.TenantProvider
	tenantPublisherMock *mocks.TenantPublishServicer
}

func initTenantServiceTest(t *testing.T) (mm *tenantServiceMockManager, serviceTestSubject *TenantsApiService) {
	mm = &tenantServiceMockManager{
		serviceUtilsMock:    um.NewErrorUtilser(t),
		iamMock:             iamMock.NewIIdm(t),
		tenantMock:          tm.NewTenantProvider(t),
		tenantPublisherMock: mocks.NewTenantPublishServicer(t),
	}

	serviceTestSubject = &TenantsApiService{
		serviceUtils:    mm.serviceUtilsMock,
		iam:             mm.iamMock,
		tenantProvider:  mm.tenantMock,
		tenantPublisher: mm.tenantPublisherMock,
	}
	return
}

func TestCreateInternalTenant_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantServiceTest(t)
	mm.tenantMock.On("CreateTenant", mock.Anything, mock.Anything).Return(&api.TenantCreationResponse{}, nil)
	response, err := serviceTestSubject.CreateTenantInternal(context.Background(), api.TenantCreationRequest{
		Name:  NAME,
		Email: EMAIL,
	})
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Nil(t, err)
}

func TestCreateInternalTenant_404Error(t *testing.T) {
	mm, serviceTestSubject := initTenantServiceTest(t)
	mm.tenantMock.On("CreateTenant", mock.Anything, mock.Anything).Return(nil, getApiError(http.StatusBadRequest))
	response, err := serviceTestSubject.CreateTenantInternal(context.Background(), api.TenantCreationRequest{
		Name:  NAME,
		Email: EMAIL,
	})
	assert.NotNil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestCreateInternalTenant_500Error(t *testing.T) {
	mm, serviceTestSubject := initTenantServiceTest(t)
	mm.tenantMock.On("CreateTenant", mock.Anything, mock.Anything).Return(nil, getApiError(http.StatusInternalServerError))
	response, err := serviceTestSubject.CreateTenantInternal(context.Background(), api.TenantCreationRequest{
		Name:  NAME,
		Email: EMAIL,
	})
	assert.NotNil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateInternalTenant_409Error(t *testing.T) {
	mm, serviceTestSubject := initTenantServiceTest(t)
	mm.tenantMock.On("CreateTenant", mock.Anything, mock.Anything).Return(nil, getApiError(http.StatusConflict))
	response, err := serviceTestSubject.CreateTenantInternal(context.Background(), api.TenantCreationRequest{
		Name:  NAME,
		Email: EMAIL,
	})
	assert.NotNil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusConflict, response.Code)
}

func TestDeleteTenantByIdInternal_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantServiceTest(t)
	mm.tenantMock.On("DeleteTenantById", mock.Anything, mock.Anything, false).Return(pointer.String("tenantId"), nil)
	response, err := serviceTestSubject.DeleteTenantByIdInternal(context.Background(), 1, false)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
}

func TestDeleteTenantByIdInternal_500Error(t *testing.T) {
	mm, serviceTestSubject := initTenantServiceTest(t)
	mm.tenantMock.On("DeleteTenantById", mock.Anything, mock.Anything, false).Return(nil, getApiError(http.StatusInternalServerError))
	response, err := serviceTestSubject.DeleteTenantByIdInternal(context.Background(), 1, false)
	assert.NotNil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestInternalDeleteTenantById_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantServiceTest(t)
	mm.tenantMock.On("DeleteTenantById", mock.Anything, mock.Anything, false).Return(pointer.String("tenantId"), nil)
	response, err := serviceTestSubject.DeleteTenantByIdInternal(context.Background(), 1, false)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
}

func TestInternalHardDeleteTenantById_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantServiceTest(t)
	mm.tenantMock.On("DeleteTenantById", mock.Anything, mock.Anything, true).Return(pointer.String("tenantId"), nil)
	mm.tenantPublisherMock.On("PublishTenantDeleteEvent", mock.Anything, mock.Anything).Return(nil)
	response, err := serviceTestSubject.DeleteTenantByIdInternal(context.Background(), 1, true)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
}

func TestGetTenantByIdInternal_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantServiceTest(t)
	mm.tenantMock.On("DeleteTenantById", mock.Anything, mock.Anything, true).Return(pointer.String("tenantId"), nil)
	mm.tenantPublisherMock.On("PublishTenantDeleteEvent", mock.Anything, mock.Anything).Return(errors.New("test"))
	response, err := serviceTestSubject.DeleteTenantByIdInternal(context.Background(), 1, true)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
}

func TestGetTenantByIdInternal_500Error(t *testing.T) {
	mm, serviceTestSubject := initTenantServiceTest(t)
	mm.tenantMock.On("GetTenantById", mock.Anything, mock.Anything).Return(nil, getApiError(http.StatusInternalServerError))
	response, err := serviceTestSubject.GetTenantByIdInternal(context.Background(), 1)
	assert.NotNil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetTenantByNameInternal_HappyFlow(t *testing.T) {
	mm, serviceTestSubject := initTenantServiceTest(t)
	mm.tenantMock.On("GetTenantByName", mock.Anything, mock.Anything).Return(getTenantModel(), nil)
	response, err := serviceTestSubject.GetTenant(context.Background(), "name")
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
}

func TestGetTenantByNameInternal_500Error(t *testing.T) {
	mm, serviceTestSubject := initTenantServiceTest(t)
	mm.tenantMock.On("GetTenantByName", mock.Anything, mock.Anything).Return(nil, getApiError(http.StatusInternalServerError))
	response, err := serviceTestSubject.GetTenant(context.Background(), "name")
	assert.NotNil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func getApiError(code int) *api.Error {
	return api.NewError(code, "error", fmt.Errorf("error"))
}

func getTenantModel() *models.Tenant {
	alias := "alias"
	return &models.Tenant{
		Name:        "name",
		DisplayName: "display-name",
		Uuid:        uuid.New().String(),
		Status:      models.TenantCreating,
		Alias:       &alias,
	}
}

func getTenantArrayModel() *[]models.Tenant {
	alias1 := "alias1"
	alias2 := "alias2"
	return &[]models.Tenant{
		{
			Name:        "name1",
			DisplayName: "display-name1",
			Uuid:        uuid.New().String(),
			Alias:       &alias1,
		},
		{
			Name:        "name2",
			DisplayName: "display-name2",
			Uuid:        uuid.New().String(),
			Alias:       &alias2,
		},
	}
}
