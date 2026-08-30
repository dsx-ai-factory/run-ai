package business

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	authMock "github.com/run-ai/runai/control-plane/authorization/pkg/client/mocks"
	businessMocks "github.com/run-ai/runai/control-plane/identity-manager/pkg/business/mocks"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	runaictx "github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

const (
	realmName = "test-realm"
	tenantID  = "test-tenant"
)

func initAuthManagerTests(t *testing.T) (ctx context.Context, realmDalMock *businessMocks.RealmDal, testSubject AuthManagerInterface) {
	realmDalMock = businessMocks.NewRealmDal(t)

	ctx = context.WithValue(context.Background(), utils.AuthRealmKey, realmName)
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantID)

	testSubject = NewAuthManager(realmDalMock)
	return ctx, realmDalMock, testSubject
}

func Test_GetRealmAndTenantID_EmptyTenantId_ShouldReturnFromContext(t *testing.T) {
	ctx, _, testSubject := initAuthManagerTests(t)

	realm, tenantId, err := testSubject.GetRealmAndTenantID(ctx, "")

	assert.NoError(t, err)
	assert.Equal(t, realmName, *realm)
	assert.Equal(t, tenantID, *tenantId)
}

func Test_GetRealmAndTenantID_NonEmptyTenantIdButNotManagementTenant_ShouldReturnUnauthorized(t *testing.T) {
	ctx, _, testSubject := initAuthManagerTests(t)

	realm, tenantId, err := testSubject.GetRealmAndTenantID(ctx, "some-tenant")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
	assert.Nil(t, realm)
	assert.Nil(t, tenantId)
}

func Test_GetRealmAndTenantID_ManagementTenantButNoSystemScope_ShouldReturnUnauthorized(t *testing.T) {
	ctx, _, testSubject := initAuthManagerTests(t)
	managementCtx := context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")

	realm, tenantId, err := testSubject.GetRealmAndTenantID(managementCtx, "some-tenant")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")
	assert.Nil(t, realm)
	assert.Nil(t, tenantId)
}

func Test_GetRealmAndTenantID_ManagementTenantWithSystemScope_ShouldGetRealmFromDB(t *testing.T) {
	ctx, realmDalMock, testSubject := initAuthManagerTests(t)

	managementCtx := context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")
	permittedScopes := &authorization_api.PermittedScopes{
		System: ptr.To(true),
	}
	managementCtx = context.WithValue(managementCtx, authorization_utils.PermittedScopesKey, permittedScopes)

	expectedRealm := &models.Realm{
		Realm:    "test-realm",
		TenantId: "test-tenant",
	}
	realmDalMock.On("GetRealmByTenantId", managementCtx, "test-tenant").Return(expectedRealm, nil).Once()

	realm, tenantID, err := testSubject.GetRealmAndTenantID(managementCtx, "test-tenant")

	assert.NoError(t, err)
	assert.Equal(t, expectedRealm.Realm, *realm)
	assert.Equal(t, "test-tenant", *tenantID)
	realmDalMock.AssertExpectations(t)
}

func Test_GetRealmAndTenantID_ManagementTenantWithSystemScope_DBError(t *testing.T) {
	ctx, realmDalMock, testSubject := initAuthManagerTests(t)

	managementCtx := context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")
	permittedScopes := &authorization_api.PermittedScopes{
		System: ptr.To(true),
	}
	managementCtx = context.WithValue(managementCtx, authorization_utils.PermittedScopesKey, permittedScopes)

	realmDalMock.On("GetRealmByTenantId", managementCtx, "test-tenant").Return(nil, fmt.Errorf("DB error")).Once()

	realm, tenantID, err := testSubject.GetRealmAndTenantID(managementCtx, "test-tenant")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DB error")
	assert.Nil(t, realm)
	assert.Nil(t, tenantID)
	realmDalMock.AssertExpectations(t)
}

func Test_GetRealmAndTenantID_ManagementTenantWithSystemScope_RealmNotFound(t *testing.T) {
	ctx, realmDalMock, testSubject := initAuthManagerTests(t)

	managementCtx := context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")
	permittedScopes := &authorization_api.PermittedScopes{
		System: ptr.To(true),
	}
	managementCtx = context.WithValue(managementCtx, authorization_utils.PermittedScopesKey, permittedScopes)

	realmDalMock.On("GetRealmByTenantId", managementCtx, "non-existent-tenant").Return(nil, fmt.Errorf("realm not found")).Once()

	realm, tenantID, err := testSubject.GetRealmAndTenantID(managementCtx, "non-existent-tenant")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "realm not found")
	assert.Nil(t, realm)
	assert.Nil(t, tenantID)
	realmDalMock.AssertExpectations(t)
}

func TestValidatePrivilegeEscalation_Success(t *testing.T) {
	ctx := context.Background()
	authzClient := authMock.NewIAuthorizationClient(t)
	authManager := &AuthManager{}
	subjectId := "user@example.com"

	mockRequest := authorization_api.ApiPrivilegeEscalationEvaluateRequest{ApiService: authzClient}
	authzClient.On("PrivilegeEscalationEvaluate", ctx).Return(mockRequest)
	authzClient.On("PrivilegeEscalationEvaluateExecute", mock.MatchedBy(func(req authorization_api.ApiPrivilegeEscalationEvaluateRequest) bool {
		return true // Accept any request that was properly configured
	})).Return(&http.Response{StatusCode: 200}, nil)

	err := authManager.ValidatePrivilegeEscalation(ctx, authzClient, subjectId)

	assert.NoError(t, err)
	authzClient.AssertExpectations(t)
}

func TestValidatePrivilegeEscalation_Forbidden(t *testing.T) {
	ctx := context.Background()
	authzClient := authMock.NewIAuthorizationClient(t)
	authManager := &AuthManager{}
	subjectId := "user@example.com"

	mockRequest := authorization_api.ApiPrivilegeEscalationEvaluateRequest{ApiService: authzClient}
	authzClient.On("PrivilegeEscalationEvaluate", ctx).Return(mockRequest)

	forbiddenErr := fmt.Errorf("403 Forbidden")
	authzClient.On("PrivilegeEscalationEvaluateExecute", mock.MatchedBy(func(req authorization_api.ApiPrivilegeEscalationEvaluateRequest) bool {
		return true
	})).Return(&http.Response{StatusCode: 403}, forbiddenErr)

	err := authManager.ValidatePrivilegeEscalation(ctx, authzClient, subjectId)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "403 Forbidden")
	authzClient.AssertExpectations(t)
}

func TestValidatePrivilegeEscalation_InternalError(t *testing.T) {
	ctx := context.Background()
	authzClient := authMock.NewIAuthorizationClient(t)
	authManager := &AuthManager{}
	subjectId := "user@example.com"

	mockRequest := authorization_api.ApiPrivilegeEscalationEvaluateRequest{ApiService: authzClient}
	authzClient.On("PrivilegeEscalationEvaluate", ctx).Return(mockRequest)

	internalErr := fmt.Errorf("500 Internal Server Error")
	authzClient.On("PrivilegeEscalationEvaluateExecute", mock.MatchedBy(func(req authorization_api.ApiPrivilegeEscalationEvaluateRequest) bool {
		return true
	})).Return(&http.Response{StatusCode: 500}, internalErr)

	err := authManager.ValidatePrivilegeEscalation(ctx, authzClient, subjectId)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500 Internal Server Error")
	authzClient.AssertExpectations(t)
}
