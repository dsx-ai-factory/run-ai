package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	runai_config "github.com/run-ai/runai/runai-common-packages/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	authMock "github.com/run-ai/runai/control-plane/authorization/pkg/client/mocks"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	iamError "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/errors"
	iamMock "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/mocks"
	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	authBusiness "github.com/run-ai/runai/control-plane/identity-manager/pkg/business/authorization"
	dalMock "github.com/run-ai/runai/control-plane/identity-manager/pkg/business/mocks"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/config"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	mockData "github.com/run-ai/runai/control-plane/identity-manager/pkg/utils/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

const (
	Username = "test@run.ai"
)

func initUserServiceTest(t *testing.T) (iamServicesContainerMock *iamMock.MockServicesContainer, authzClientMock *authMock.IAuthorizationClient, testSubject *UsersApiService) {
	config := setIamConfig()
	tenantConfig := setTenantConfig()
	mockIamServicesContainer := iamMock.NewMockServicesContainer()
	authzClientMock = authMock.NewIAuthorizationClient(t)
	realmDalMock := dalMock.NewRealmDal(t)
	authManagerMock := authBusiness.NewAuthManager(realmDalMock)
	notifyConfig := setNotifyConfig()
	testSubject = &UsersApiService{
		authzClient:          authzClientMock,
		iamServices:          *mockIamServicesContainer.ToServicesContainer(),
		config:               config,
		notifyConfig:         notifyConfig,
		tenantDomainTemplate: tenantConfig.TenantDomainTemplate,
		authManager:          authManagerMock,
	}
	return mockIamServicesContainer, authzClientMock, testSubject
}

func setIamConfig() config.Iam {
	conf := config.Iam{
		Url:               "url",
		AdminUsername:     "adminUsername",
		AdminPassword:     "adminPassword",
		IgnoredRealms:     []string{},
		TenantsManagerUrl: "tenantsManagerUrl",
	}
	runai_config.Set(&conf)
	return conf
}

func setNotifyConfig() config.Notify {
	conf := config.Notify{
		Enabled: true,
	}
	runai_config.Set(&conf)
	return conf
}

func setTenantConfig() config.TenantRealm {
	return config.TenantRealm{
		AppHost:              "https://app.run.ai",
		TenantDomainTemplate: "https://{{.}}.run.ai",
		Env:                  "test",
		DefaultClientId:      "runai-admin-ui",
		Name:                 "test-realm",
		IsMultiTenant:        false,
		LoginTheme:           "runai-theme",
	}
}

func TestCreateUser_HappyFlow(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")
	iamMock.MockUserService.On("CreateUser", mock.Anything, mock.Anything, mock.Anything).Return(mockData.USER_ID, nil)
	response, err := testSubject.CreateUser(withCtx, api.UserCreationRequest{
		Email: mockData.USER_EMAIL,
	})
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, mockData.USER_ID, response.Body.(api.UserPostResponse).Id)
	assert.Equal(t, mockData.USER_EMAIL, response.Body.(api.UserPostResponse).Username)
	assert.NotEmpty(t, response.Body.(api.UserPostResponse).TempPassword)
}

func TestCreateUser_InternalError(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")
	iamMock.MockUserService.On("CreateUser", mock.Anything, mock.Anything, mock.Anything).Return("", iamError.ErrInternal)
	response, err := testSubject.CreateUser(withCtx, api.UserCreationRequest{
		Email: mockData.USER_EMAIL,
	})
	assert.NotNil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateUser_UserAlreadyExists(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	iamMock.MockUserService.On("CreateUser", mock.Anything, mock.Anything, mock.Anything).Return("", iamError.ErrConflict)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")
	response, err := testSubject.CreateUser(withCtx, api.UserCreationRequest{
		Email: mockData.USER_EMAIL,
	})
	assert.NotNil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusConflict, response.Code)
}

func setupDeleteUserTest(t *testing.T) (context.Context, *iamMock.MockServicesContainer, *authMock.IAuthorizationClient, *UsersApiService) {
	iamMock, authzClientMock, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	user := mockData.GetUser(ptr.To(mockData.USER_EMAIL))
	iamMock.MockUserService.On("GetUserByID", mock.Anything, mock.Anything, mock.Anything).Return(user, nil)
	setupPrivilegeEscalationMock(t, authzClientMock, nil)
	return ctx, iamMock, authzClientMock, testSubject
}

func TestDeleteUserById_HappyFlow(t *testing.T) {
	ctx, iamMock, authzClientMock, testSubject := setupDeleteUserTest(t)
	authzClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authzClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(nil, nil)
	iamMock.MockUserService.On("DeleteUser", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iamMock.MockClientService.On("DeleteAccessKeysByEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	response, err := testSubject.DeleteUserById(ctx, mockData.USER_ID)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, response.Code)
}

func TestDeleteUserById_InternalError(t *testing.T) {
	apiError := utils.Error{
		Code:    http.StatusInternalServerError,
		Message: "",
	}
	ctx, iamMock, authzClientMock, testSubject := setupDeleteUserTest(t)
	authzClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authzClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(nil, nil)
	iamMock.MockClientService.On("DeleteAccessKeysByEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iamMock.MockUserService.On("DeleteUser", mock.Anything, mock.Anything, mock.Anything).Return(apiError)
	response, err := testSubject.DeleteUserById(ctx, mockData.USER_ID)
	assert.NotNil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestDeleteUserById_UserNotFound(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	iamMock.MockUserService.On("GetUserByID", mock.Anything, mock.Anything, mock.Anything).Return(nil, iamError.ErrNotFound)
	response, err := testSubject.DeleteUserById(context.Background(), "noId")
	assert.NotNil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestDeleteUserById_UserWithoutEmail(t *testing.T) {
	userWithoutEmail := mockData.GetUser(nil)
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	iamMock.MockUserService.On("GetUserByID", mock.Anything, mock.Anything, mock.Anything).Return(userWithoutEmail, nil)
	response, err := testSubject.DeleteUserById(ctx, mockData.USER_ID)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestDeleteUserById_InsufficientPermissions(t *testing.T) {
	iamMock, authzClientMock, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	user := mockData.GetUser(ptr.To(mockData.USER_EMAIL))
	iamMock.MockUserService.On("GetUserByID", mock.Anything, mock.Anything, mock.Anything).Return(user, nil)
	forbiddenErr := errors.New("403 Forbidden")
	setupPrivilegeEscalationMock(t, authzClientMock, forbiddenErr)
	response, err := testSubject.DeleteUserById(ctx, mockData.USER_ID)
	assertErrorResponse(t, response, err, http.StatusForbidden, "You're not authorized to delete the selected user")
}

func TestDeleteUserById_PrivilegeEscalationGeneralError(t *testing.T) {
	iamMock, authzClientMock, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	user := mockData.GetUser(ptr.To(mockData.USER_EMAIL))
	iamMock.MockUserService.On("GetUserByID", mock.Anything, mock.Anything, mock.Anything).Return(user, nil)
	generalErr := errors.New("internal server error")
	setupPrivilegeEscalationMock(t, authzClientMock, generalErr)
	response, err := testSubject.DeleteUserById(ctx, mockData.USER_ID)
	assertErrorResponse(t, response, err, http.StatusInternalServerError, "")
}

func TestDeleteUserById_AuthClientError(t *testing.T) {
	httpResponse := &http.Response{
		StatusCode: 500,
	}
	ctx, _, authzClientMock, testSubject := setupDeleteUserTest(t)
	authzClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authzClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(httpResponse, iamError.ErrInternal)
	response, err := testSubject.DeleteUserById(ctx, mockData.USER_ID)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestDeleteUserById_ErrorDeletingApps(t *testing.T) {
	ctx, iamMock, authzClientMock, testSubject := setupDeleteUserTest(t)
	authzClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authzClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(nil, nil)
	iamMock.MockUserService.On("DeleteUser", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iamMock.MockClientService.On("DeleteAccessKeysByEmail", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
	response, err := testSubject.DeleteUserById(ctx, mockData.USER_ID)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func ctxWithSubject() context.Context {
	return context.WithValue(context.Background(), runaictx.CfSubjectID, "admin@run.ai")
}

func TestDeleteInternalUserByEmail_HappyFlow(t *testing.T) {
	iamMock, authzClientMock, testSubject := initUserServiceTest(t)
	user := mockData.GetUser(ptr.To(mockData.USER_EMAIL))
	iamMock.MockUserService.On("GetUserByEmail", mock.Anything, mock.Anything, mock.Anything).Return(user, nil)
	authzClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authzClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(nil, nil)
	iamMock.MockClientService.On("DeleteAccessKeysByEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iamMock.MockUserService.On("DeleteUser", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	response, err := testSubject.DeleteInternalUserByEmail(ctxWithSubject(), "test-realm", mockData.USER_EMAIL)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestDeleteInternalUserByEmail_NoSubjectID_SkipsAccessRulesAndDeletesUser(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	user := mockData.GetUser(ptr.To(mockData.USER_EMAIL))
	iamMock.MockUserService.On("GetUserByEmail", mock.Anything, mock.Anything, mock.Anything).Return(user, nil)
	iamMock.MockUserService.On("DeleteUser", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	response, err := testSubject.DeleteInternalUserByEmail(context.Background(), "test-realm", mockData.USER_EMAIL)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestDeleteInternalUserByEmail_UserNotFound(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	iamMock.MockUserService.On("GetUserByEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil, iamError.ErrNotFound)
	response, err := testSubject.DeleteInternalUserByEmail(context.Background(), "test-realm", mockData.USER_EMAIL)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestDeleteInternalUserByEmail_AuthClientError(t *testing.T) {
	iamMock, authzClientMock, testSubject := initUserServiceTest(t)
	user := mockData.GetUser(ptr.To(mockData.USER_EMAIL))
	iamMock.MockUserService.On("GetUserByEmail", mock.Anything, mock.Anything, mock.Anything).Return(user, nil)
	authzClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authzClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(&http.Response{StatusCode: 500}, iamError.ErrInternal)
	response, err := testSubject.DeleteInternalUserByEmail(ctxWithSubject(), "test-realm", mockData.USER_EMAIL)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestDeleteInternalUserByEmail_ErrorDeletingApps(t *testing.T) {
	iamMock, authzClientMock, testSubject := initUserServiceTest(t)
	user := mockData.GetUser(ptr.To(mockData.USER_EMAIL))
	iamMock.MockUserService.On("GetUserByEmail", mock.Anything, mock.Anything, mock.Anything).Return(user, nil)
	authzClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authzClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(nil, nil)
	iamMock.MockClientService.On("DeleteAccessKeysByEmail", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
	response, err := testSubject.DeleteInternalUserByEmail(ctxWithSubject(), "test-realm", mockData.USER_EMAIL)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestDeleteInternalUserByEmail_InternalError(t *testing.T) {
	iamMock, authzClientMock, testSubject := initUserServiceTest(t)
	user := mockData.GetUser(ptr.To(mockData.USER_EMAIL))
	iamMock.MockUserService.On("GetUserByEmail", mock.Anything, mock.Anything, mock.Anything).Return(user, nil)
	authzClientMock.On("DeleteSubjectAccessRules", mock.Anything).Return(authorization_api.ApiDeleteSubjectAccessRulesRequest{})
	authzClientMock.On("DeleteSubjectAccessRulesExecute", mock.Anything).Return(nil, nil)
	iamMock.MockClientService.On("DeleteAccessKeysByEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iamMock.MockUserService.On("DeleteUser", mock.Anything, mock.Anything, mock.Anything).Return(iamError.ErrInternal)
	response, err := testSubject.DeleteInternalUserByEmail(ctxWithSubject(), "test-realm", mockData.USER_EMAIL)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetUserById_HappyFlow(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	user := mockData.GetUser(ptr.To(mockData.USER_EMAIL))
	iamMock.MockUserService.On("GetUserByID", mock.Anything, mock.Anything, mock.Anything).Return(user, nil)
	response, err := testSubject.GetUserById(context.Background(), mockData.USER_ID)
	assert.Nil(t, err)
	assert.Equal(t, mockData.USER_EMAIL, response.Body.(api.User).Username)
}

func TestGetUserById_InternalError(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	iamMock.MockUserService.On("GetUserByID", mock.Anything, mock.Anything, mock.Anything).Return(nil, iamError.ErrInternal)
	response, err := testSubject.GetUserById(context.Background(), mockData.USER_ID)
	assert.NotNil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetUserById_UserNotFound(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	iamMock.MockUserService.On("GetUserByID", mock.Anything, mock.Anything, mock.Anything).Return(nil, iamError.ErrNotFound)
	response, err := testSubject.GetUserById(context.Background(), mockData.USER_ID)
	assert.NotNil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func TestGetInternalUserById_HappyFlow(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	user := mockData.GetUser(ptr.To(mockData.USER_EMAIL))
	iamMock.MockUserService.On("GetUserByID", mock.Anything, mock.Anything, mock.Anything).Return(user, nil)

	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, "runai")
	response, err := testSubject.GetInternalUserById(ctx, mockData.USER_ID)
	assert.Nil(t, err)
	assert.Equal(t, mockData.USER_EMAIL, response.Body.(api.User).Username)
}

func TestGetInternalUserById_MissingRealm(t *testing.T) {
	_, _, testSubject := initUserServiceTest(t)
	response, err := testSubject.GetInternalUserById(context.Background(), mockData.USER_ID)
	assert.NotNil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestGetUsers_HappyFlow(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	expectedUsers := mockData.GetUsers()
	iamMock.MockUserService.On("GetUsers", mock.Anything, mock.Anything, mock.Anything).Return(expectedUsers, nil)
	response, err := testSubject.GetUsers(context.Background(), []string{"isLocal==true"}, api.USERSFILTERSORTFIELDS_USERNAME, "asc", 0, 10, "searchVal")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, len(response.Body.([]api.User)), 2)
}

func TestGetUsers_InternalError(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	iamMock.MockUserService.On("GetUsers", mock.Anything, mock.Anything, mock.Anything).Return(nil, iamError.ErrInternal)
	response, err := testSubject.GetUsers(context.Background(), []string{"isLocal==true"}, api.USERSFILTERSORTFIELDS_USERNAME, "asc", 0, 10, "searchVal")
	assert.NotNil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestLogoutUser_HappyFlow(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	iamMock.MockUserService.On("LogoutUser", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	response, err := testSubject.LogoutUser(context.Background(), mockData.USER_ID)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestLogoutUser_InternalError(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	iamMock.MockUserService.On("LogoutUser", mock.Anything, mock.Anything, mock.Anything).Return(iamError.ErrInternal)
	response, err := testSubject.LogoutUser(context.Background(), mockData.USER_ID)
	assert.NotNil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestLogoutUser_UserNotFound(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	iamMock.MockUserService.On("LogoutUser", mock.Anything, mock.Anything, mock.Anything).Return(iamError.ErrNotFound)
	response, err := testSubject.LogoutUser(context.Background(), mockData.USER_ID)
	assert.NotNil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusNotFound, response.Code)
}

func setupUserPrivilegeEscalationTest(t *testing.T) (context.Context, *iamMock.MockServicesContainer, *authMock.IAuthorizationClient, *UsersApiService) {
	iamMock, authzClientMock, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)

	user := mockData.GetUser(ptr.To(mockData.USER_EMAIL))
	iamMock.MockUserService.On("GetUserByID", ctx, mockData.REALM_NAME, mockData.USER_ID).Return(user, nil)

	return ctx, iamMock, authzClientMock, testSubject
}

func TestResetUserPassword_HappyFlow(t *testing.T) {
	ctx, iamMock, authzClientMock, testSubject := setupUserPrivilegeEscalationTest(t)

	setupPrivilegeEscalationMock(t, authzClientMock, nil)

	iamMock.MockUserService.On("ResetUserPassword", ctx, mockData.REALM_NAME, mockData.USER_ID, mock.Anything).Return(nil)
	iamMock.MockUserService.On("LogoutUser", ctx, mockData.REALM_NAME, mockData.USER_ID).Return(nil)

	response, err := testSubject.ResetUserPassword(ctx, mockData.USER_ID)
	assert.Nil(t, err)
	assert.NotNil(t, response.Body)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestResetUserPassword_GetUserError(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)

	iamMock.MockUserService.On("GetUserByID", ctx, mockData.REALM_NAME, mockData.USER_ID).Return(nil, iamError.ErrNotFound)

	response, err := testSubject.ResetUserPassword(ctx, mockData.USER_ID)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusNotFound, response.Code)

	utilsError := err.(utils.Error)
	assert.Contains(t, utilsError.Message, "failed getting user for privilege escalation check")
	assert.Equal(t, http.StatusNotFound, utilsError.Code)
}

func TestResetUserPassword_InternalError(t *testing.T) {
	ctx, iamMock, authzClientMock, testSubject := setupUserPrivilegeEscalationTest(t)

	setupPrivilegeEscalationMock(t, authzClientMock, nil)

	iamMock.MockUserService.On("ResetUserPassword", ctx, mockData.REALM_NAME, mockData.USER_ID, mock.Anything).Return(iamError.ErrInternal)

	response, err := testSubject.ResetUserPassword(ctx, mockData.USER_ID)
	assert.NotNil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestResetUserPassword_InsufficientPermissions(t *testing.T) {
	ctx, _, authzClientMock, testSubject := setupUserPrivilegeEscalationTest(t)

	forbiddenErr := errors.New("403 Forbidden")
	setupPrivilegeEscalationMock(t, authzClientMock, forbiddenErr)

	response, err := testSubject.ResetUserPassword(ctx, mockData.USER_ID)
	assertErrorResponse(t, response, err, http.StatusForbidden, "You're not authorized to reset password for the selected user")
}

func TestResetUserPassword_PrivilegeEscalationGeneralError(t *testing.T) {
	ctx, _, authzClientMock, testSubject := setupUserPrivilegeEscalationTest(t)

	generalErr := errors.New("internal server error")
	setupPrivilegeEscalationMock(t, authzClientMock, generalErr)

	response, err := testSubject.ResetUserPassword(ctx, mockData.USER_ID)
	assertErrorResponse(t, response, err, http.StatusInternalServerError, "") // Empty for security
}

func TestResetUserPassword_UserEmailNil(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)

	userWithNilEmail := &iamModels.User{
		ID:        ptr.To(mockData.USER_ID),
		Username:  ptr.To("testuser"),
		Email:     nil,
		Enabled:   ptr.To(true),
		CreatedBy: ptr.To("admin"),
	}
	iamMock.MockUserService.On("GetUserByID", ctx, mockData.REALM_NAME, mockData.USER_ID).Return(userWithNilEmail, nil)

	response, err := testSubject.ResetUserPassword(ctx, mockData.USER_ID)
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusBadRequest, response.Code)

	utilsError := err.(utils.Error)
	assert.Contains(t, utilsError.Message, "User email is required for privilege escalation check")
}

func TestCreateInternalUser_HappyFlow(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	iamMock.MockUserService.On("CreateUser", mock.Anything, mock.Anything, mock.Anything).Return(mockData.USER_ID, nil)
	response, err := testSubject.CreateIntenralUser(context.Background(), api.InternalUserCreationRequest{
		Email:         mockData.USER_EMAIL,
		Realm:         mockData.REALM_NAME,
		Password:      ptr.To(mockData.USER_PASSWORD),
		ResetPassword: false,
	})
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, mockData.USER_ID, response.Body.(api.UserPostResponse).Id)
	assert.Equal(t, mockData.USER_EMAIL, response.Body.(api.UserPostResponse).Username)
	assert.NotNil(t, response.Body.(api.UserPostResponse).TempPassword)
}

func TestCreateInternalUser_HappyFlow_WithGeneratedPassword(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	iamMock.MockUserService.On("CreateUser", mock.Anything, mock.Anything, mock.Anything).Return(mockData.USER_ID, nil)
	response, err := testSubject.CreateIntenralUser(context.Background(), api.InternalUserCreationRequest{
		Email:         mockData.USER_EMAIL,
		Realm:         mockData.REALM_NAME,
		Password:      ptr.To(""),
		ResetPassword: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, mockData.USER_ID, response.Body.(api.UserPostResponse).Id)
	assert.Equal(t, mockData.USER_EMAIL, response.Body.(api.UserPostResponse).Username)
	assert.NotEmpty(t, response.Body.(api.UserPostResponse).TempPassword)
	assert.NotEqual(t, mockData.USER_PASSWORD, response.Body.(api.UserPostResponse).TempPassword)
}

func TestCreateInternalUser_Error(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	apiError := utils.Error{
		Code: http.StatusInternalServerError,
	}
	iamMock.MockUserService.On("CreateUser", mock.Anything, mock.Anything, mock.Anything).Return("", apiError)
	response, err := testSubject.CreateIntenralUser(context.Background(), api.InternalUserCreationRequest{
		Email:         mockData.USER_EMAIL,
		Realm:         mockData.REALM_NAME,
		Password:      ptr.To(mockData.USER_PASSWORD),
		ResetPassword: true,
	})
	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetLocalUsers(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	expectedUsers := mockData.GetLocalAndSSOUsers()
	iamMock.MockUserService.On("GetUsers", mock.Anything, mock.Anything, mock.Anything).Return(expectedUsers, nil)
	response, err := testSubject.GetUsers(context.Background(), []string{"isLocal==true"}, api.USERSFILTERSORTFIELDS_USERNAME, "asc", 0, 10, "searchVal")
	users := response.Body.([]api.User)

	var localUsers []api.User
	var ssoUsers []api.User

	for _, user := range users {
		if *user.IsLocal {
			localUsers = append(localUsers, user)
		} else {
			ssoUsers = append(ssoUsers, user)
		}
	}
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, len(ssoUsers), 3)
	assert.Equal(t, len(localUsers), 2)
}

func TestChangeUserPassword_HappyFlow(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")
	user := mockData.GetUser(ptr.To(mockData.USER_EMAIL))
	password := mockData.GetUserChangePasswordRequest()

	iamMock.MockUserService.On("GetUserByEmail", withCtx, mockData.REALM_NAME, "admin").Return(user, nil)
	iamMock.MockUserService.On("ChangeUserPassword", withCtx, mockData.REALM_NAME, user, password.CurrentPassword, password.NewPassword).Return(nil)
	iamMock.MockUserService.On("LogoutUser", withCtx, mockData.REALM_NAME, *user.ID).Return(nil)

	response, err := testSubject.ChangeUserPassword(withCtx, mockData.GetUserChangePasswordRequest())
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestChangeUserPassword_InternalError(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")
	user := mockData.GetUser(ptr.To(mockData.USER_EMAIL))
	password := mockData.GetUserChangePasswordRequest()

	iamMock.MockUserService.On("GetUserByEmail", withCtx, mockData.REALM_NAME, "admin").Return(user, nil)
	iamMock.MockUserService.On("ChangeUserPassword", withCtx, mockData.REALM_NAME, user, password.CurrentPassword, password.NewPassword).Return(iamError.ErrInvalidInput)

	response, err := testSubject.ChangeUserPassword(withCtx, mockData.GetUserChangePasswordRequest())
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestChangeUserPassword_GetUserByEmail_Error(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")

	iamMock.MockUserService.On("GetUserByEmail", withCtx, mockData.REALM_NAME, "admin").Return(nil, iamError.ErrInternal)
	response, err := testSubject.ChangeUserPassword(withCtx, mockData.GetUserChangePasswordRequest())

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestChangeUserPassword_FailedLogoutUser(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")
	user := mockData.GetUser(ptr.To(mockData.USER_EMAIL))
	password := mockData.GetUserChangePasswordRequest()

	iamMock.MockUserService.On("GetUserByEmail", withCtx, mockData.REALM_NAME, "admin").Return(user, nil)
	iamMock.MockUserService.On("ChangeUserPassword", withCtx, mockData.REALM_NAME, user, password.CurrentPassword, password.NewPassword).Return(nil)
	iamMock.MockUserService.On("LogoutUser", withCtx, mockData.REALM_NAME, *user.ID).Return(fmt.Errorf("error"))

	response, err := testSubject.ChangeUserPassword(withCtx, mockData.GetUserChangePasswordRequest())
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestChangeUserPassword_InvalidPasswordError(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)

	iamMock.MockUserService.On("GetUserByEmail", mock.Anything, mock.Anything, mock.Anything).
		Return(mockData.GetUser(ptr.To(mockData.USER_EMAIL)), nil)
	iamMock.MockUserService.On("ChangeUserPassword", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(fmt.Errorf("keycloack error: %s: %w", "invalidPasswordMinSpecialCharsMessage: Invalid password: must contain at least 1 special character.", iamError.ErrInvalidInput))

	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")
	response, err := testSubject.ChangeUserPassword(withCtx, mockData.GetUserChangePasswordRequest())

	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Contains(t, response.Body, "Invalid password: must contain at least 1 special character.")
}

func TestGroups_SanityEmpty(t *testing.T) {
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfTenantId, mockData.TENANT_ID)
	iamMock, authzMock, testSubject := initUserServiceTest(t)
	user := mockData.GetUserWithGroups()
	iamMock.MockUserService.On("GetUserByEmail", withCtx, mockData.REALM_NAME, Username).Return(user, nil)
	authzMock.On("GetGroupsForTenant", withCtx, mockData.TENANT_ID).Return([]string{"test"}, nil)

	res, err := testSubject.GetUserAccessGroups(withCtx, Username)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Equal(t, []string{}, res.Body)
}

func TestGroups_KcError(t *testing.T) {
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfTenantId, mockData.TENANT_ID)
	iamMock, _, testSubject := initUserServiceTest(t)
	iamMock.MockUserService.On("GetUserByEmail", withCtx, mockData.REALM_NAME, Username).Return(nil, errors.New("error"))

	res, err := testSubject.GetUserAccessGroups(withCtx, Username)
	assert.NotNil(t, res)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, nil, res.Body)
}

func TestGroups_utils(t *testing.T) {
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfTenantId, mockData.TENANT_ID)
	iamMock, authzMock, testSubject := initUserServiceTest(t)

	user := mockData.GetUserWithGroups()
	iamMock.MockUserService.On("GetUserByEmail", withCtx, mockData.REALM_NAME, Username).Return(user, nil)
	authzMock.On("GetGroupsForTenant", withCtx, mockData.TENANT_ID).Return(nil, errors.New("error"))

	res, err := testSubject.GetUserAccessGroups(withCtx, Username)
	assert.NotNil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.Equal(t, nil, res.Body)
}

func TestCountUsers_HappyFlow(t *testing.T) {
	realm := "test-realm"

	iamMock, _, testSubject := initUserServiceTest(t)
	expectedUsers := mockData.GetUsers()
	iamMock.MockUserService.On("GetUsers", mock.Anything, mock.Anything, mock.Anything).Return(expectedUsers, nil)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, realm)

	response, err := testSubject.CountUsers(ctx, []string{
		"username=@test",
		"creationTime>=2024-09-24T00:00:00.00Z",
		"lastLogin<=2024-09-24T00:00:00.00Z",
		"lastUpdated>=2024-09-24T00:00:00.00Z",
		"createdBy=@admin",
		"isLocal==true",
	}, "searchVal")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, int64(2), response.Body.(api.CountUsers200Response).Count)
}

func TestCountUsers_ParseFiltersError(t *testing.T) {
	_, _, testSubject := initUserServiceTest(t)

	realm := "test-realm"
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, realm)

	response, err := testSubject.CountUsers(ctx, []string{"invalid_filter"}, "searchVal")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestCountUsers_GetUsersError_InternalServerError(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)

	realm := "test-realm"
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, realm)
	iamMock.MockUserService.On("GetUsers", mock.Anything, mock.Anything, mock.Anything).Return(nil, iamError.ErrInternal)

	response, _ := testSubject.CountUsers(ctx, []string{"username=@value1"}, "searchVal")

	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCountUsers_NoUsersFound(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)

	realm := "test-realm"
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, realm)
	iamMock.MockUserService.On("GetUsers", mock.Anything, mock.Anything, mock.Anything).Return(mockData.GetNoUsers(), nil)

	response, err := testSubject.CountUsers(ctx, []string{"username=@ish"}, "ish")

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, int64(0), response.Body.(api.CountUsers200Response).Count)
}

func setupPrivilegeEscalationMock(t *testing.T, authzClientMock *authMock.IAuthorizationClient, returnErr error) {
	mockInternalAPI := authMock.NewInternalAPI(t)
	var response *http.Response
	if returnErr == nil {
		response = &http.Response{StatusCode: 204, Body: http.NoBody}
	}
	mockInternalAPI.On("PrivilegeEscalationEvaluateExecute", mock.AnythingOfType("openapi.ApiPrivilegeEscalationEvaluateRequest")).Return(response, returnErr)
	mockRequest := authorization_api.ApiPrivilegeEscalationEvaluateRequest{ApiService: mockInternalAPI}
	authzClientMock.On("PrivilegeEscalationEvaluate", mock.Anything).Return(mockRequest)
}

func assertErrorResponse(t *testing.T, response api.ImplResponse, err error, expectedStatusCode int, expectedMessage string) {
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, expectedStatusCode, response.Code)

	utilsError := err.(utils.Error)
	assert.Equal(t, expectedMessage, utilsError.Message)
	assert.Equal(t, expectedStatusCode, utilsError.Code)
}

func TestBatchCreateUsers_HappyFlow(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")
	userEmail1 := "user1@run.ai"
	userEmail2 := "user2@run.ai"

	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.MatchedBy(func(user *iamModels.User) bool {
		return *user.Email == userEmail1
	})).Return(mockData.USER_ID+"1", nil).Once()
	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.MatchedBy(func(user *iamModels.User) bool {
		return *user.Email == userEmail2
	})).Return(mockData.USER_ID+"2", nil).Once()

	request := api.BulkUserCreationRequest{
		Users: []api.UserCreationRequestBulk{
			{Email: userEmail1, ResetPassword: ptr.To(false)},
			{Email: userEmail2, ResetPassword: ptr.To(true)},
		},
	}

	response, err := testSubject.BatchCreateUsers(withCtx, request)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	bulkResponse := response.Body.(api.BulkUserCreationResponse)
	assert.Equal(t, int32(2), bulkResponse.TotalRequested)
	assert.Equal(t, int32(2), bulkResponse.TotalSuccessful)
	assert.Equal(t, int32(0), bulkResponse.TotalFailed)
	assert.Len(t, bulkResponse.Results, 2)

	assert.True(t, bulkResponse.Results[0].Success)
	assert.Equal(t, mockData.USER_ID+"1", bulkResponse.Results[0].User.Id)
	assert.Equal(t, userEmail1, bulkResponse.Results[0].User.Username)
	assert.NotEmpty(t, bulkResponse.Results[0].User.TempPassword)
	assert.Nil(t, bulkResponse.Results[0].Error)

	assert.True(t, bulkResponse.Results[1].Success)
	assert.Equal(t, mockData.USER_ID+"2", bulkResponse.Results[1].User.Id)
	assert.Equal(t, userEmail2, bulkResponse.Results[1].User.Username)
	assert.NotEmpty(t, bulkResponse.Results[1].User.TempPassword)
	assert.Nil(t, bulkResponse.Results[1].Error)
}

func TestBatchCreateUsers_AllFailures(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")
	userEmail1 := "existing@run.ai"
	userEmail2 := "error@run.ai"

	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.MatchedBy(func(user *iamModels.User) bool {
		return *user.Email == userEmail1
	})).Return("", iamError.ErrConflict).Once()
	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.MatchedBy(func(user *iamModels.User) bool {
		return *user.Email == userEmail2
	})).Return("", iamError.ErrInternal).Once()

	request := api.BulkUserCreationRequest{
		Users: []api.UserCreationRequestBulk{
			{Email: userEmail1, ResetPassword: ptr.To(false)},
			{Email: userEmail2, ResetPassword: ptr.To(true)},
		},
	}

	response, err := testSubject.BatchCreateUsers(withCtx, request)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)

	bulkResponse := response.Body.(api.BulkUserCreationResponse)
	assert.Equal(t, int32(2), bulkResponse.TotalRequested)
	assert.Equal(t, int32(0), bulkResponse.TotalSuccessful)
	assert.Equal(t, int32(2), bulkResponse.TotalFailed)
	assert.Len(t, bulkResponse.Results, 2)

	assert.False(t, bulkResponse.Results[0].Success)
	assert.Empty(t, bulkResponse.Results[0].User.Id)
	assert.Equal(t, userEmail1, bulkResponse.Results[0].User.Username)
	assert.Empty(t, bulkResponse.Results[0].User.TempPassword)
	assert.NotNil(t, bulkResponse.Results[0].Error)
	assert.Equal(t, int32(http.StatusConflict), bulkResponse.Results[0].Error.Code)
	assert.Equal(t, "user already exists", bulkResponse.Results[0].Error.Message)

	assert.False(t, bulkResponse.Results[1].Success)
	assert.Empty(t, bulkResponse.Results[1].User.Id)
	assert.Equal(t, userEmail2, bulkResponse.Results[1].User.Username)
	assert.Empty(t, bulkResponse.Results[1].User.TempPassword)
	assert.NotNil(t, bulkResponse.Results[1].Error)
	assert.Equal(t, int32(http.StatusInternalServerError), bulkResponse.Results[1].Error.Code)
}

func TestBatchCreateUsers_MixedResults(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")

	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.MatchedBy(func(user *iamModels.User) bool {
		return *user.Email == "success1@run.ai"
	})).Return(mockData.USER_ID+"1", nil).Once()
	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.MatchedBy(func(user *iamModels.User) bool {
		return *user.Email == "conflict@run.ai"
	})).Return("", iamError.ErrConflict).Once()
	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.MatchedBy(func(user *iamModels.User) bool {
		return *user.Email == "success2@run.ai"
	})).Return(mockData.USER_ID+"2", nil).Once()
	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.MatchedBy(func(user *iamModels.User) bool {
		return *user.Email == "badrequest@run.ai"
	})).Return("", iamError.ErrInvalidInput).Once()

	request := api.BulkUserCreationRequest{
		Users: []api.UserCreationRequestBulk{
			{Email: "success1@run.ai", ResetPassword: ptr.To(false)},
			{Email: "conflict@run.ai", ResetPassword: ptr.To(false)},
			{Email: "success2@run.ai", ResetPassword: ptr.To(true)},
			{Email: "badrequest@run.ai", ResetPassword: ptr.To(false)},
		},
	}

	response, err := testSubject.BatchCreateUsers(withCtx, request)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	bulkResponse := response.Body.(api.BulkUserCreationResponse)
	assert.Equal(t, int32(4), bulkResponse.TotalRequested)
	assert.Equal(t, int32(2), bulkResponse.TotalSuccessful)
	assert.Equal(t, int32(2), bulkResponse.TotalFailed)
	assert.Len(t, bulkResponse.Results, 4)

	assert.True(t, bulkResponse.Results[0].Success)
	assert.Equal(t, mockData.USER_ID+"1", bulkResponse.Results[0].User.Id)
	assert.Equal(t, "success1@run.ai", bulkResponse.Results[0].User.Username)

	assert.False(t, bulkResponse.Results[1].Success)
	assert.Empty(t, bulkResponse.Results[1].User.Id)
	assert.Equal(t, "conflict@run.ai", bulkResponse.Results[1].User.Username)
	assert.Equal(t, int32(http.StatusConflict), bulkResponse.Results[1].Error.Code)
	assert.Equal(t, "user already exists", bulkResponse.Results[1].Error.Message)

	assert.True(t, bulkResponse.Results[2].Success)
	assert.Equal(t, mockData.USER_ID+"2", bulkResponse.Results[2].User.Id)
	assert.Equal(t, "success2@run.ai", bulkResponse.Results[2].User.Username)

	assert.False(t, bulkResponse.Results[3].Success)
	assert.Empty(t, bulkResponse.Results[3].User.Id)
	assert.Equal(t, "badrequest@run.ai", bulkResponse.Results[3].User.Username)
	assert.Equal(t, int32(http.StatusBadRequest), bulkResponse.Results[3].Error.Code)
	assert.Equal(t, "bad request", bulkResponse.Results[3].Error.Message)
}

func TestBatchCreateUsers_SingleUser(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")

	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.Anything).Return(mockData.USER_ID, nil).Once()

	request := api.BulkUserCreationRequest{
		Users: []api.UserCreationRequestBulk{
			{Email: "single@run.ai", ResetPassword: ptr.To(true)},
		},
	}

	response, err := testSubject.BatchCreateUsers(withCtx, request)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	bulkResponse := response.Body.(api.BulkUserCreationResponse)
	assert.Equal(t, int32(1), bulkResponse.TotalRequested)
	assert.Equal(t, int32(1), bulkResponse.TotalSuccessful)
	assert.Equal(t, int32(0), bulkResponse.TotalFailed)
	assert.Len(t, bulkResponse.Results, 1)

	assert.True(t, bulkResponse.Results[0].Success)
	assert.Equal(t, mockData.USER_ID, bulkResponse.Results[0].User.Id)
	assert.Equal(t, "single@run.ai", bulkResponse.Results[0].User.Username)
	assert.NotEmpty(t, bulkResponse.Results[0].User.TempPassword)
}

func TestGetUserCreationError_ConflictError(t *testing.T) {
	_, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)

	err := iamError.ErrConflict
	result := testSubject.getUserCreationError(ctx, "test@run.ai", err)

	assert.NotNil(t, result)
	assert.Equal(t, int32(http.StatusConflict), result.Code)
	assert.Equal(t, "user already exists", result.Message)
}

func TestGetUserCreationError_BadRequestError(t *testing.T) {
	_, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)

	err := iamError.ErrInvalidInput
	result := testSubject.getUserCreationError(ctx, "test@run.ai", err)

	assert.NotNil(t, result)
	assert.Equal(t, int32(http.StatusBadRequest), result.Code)
	assert.Equal(t, "bad request", result.Message)
}

func TestGetUserCreationError_InternalServerError(t *testing.T) {
	_, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)

	err := errors.New("some internal error")
	result := testSubject.getUserCreationError(ctx, "test@run.ai", err)

	assert.NotNil(t, result)
	assert.Equal(t, int32(http.StatusInternalServerError), result.Code)
	assert.Equal(t, "some internal error", result.Message)
}

func TestCreateUser_WithNotifyTrue(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")

	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.Anything).Return(mockData.USER_ID, nil)
	iamMock.MockUserService.On("SendPasswordResetEmail", withCtx, mockData.REALM_NAME, mockData.USER_ID, resetPasswordExpirationTimeSeconds, "https://realm_name.run.ai/login/callback?runai_flow=invite", utils.RUNAI_ADMIN_UI_CLIENT_ID).Return(nil)

	response, err := testSubject.CreateUser(withCtx, api.UserCreationRequest{
		Email:  mockData.USER_EMAIL,
		Notify: ptr.To(true),
	})

	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, mockData.USER_ID, response.Body.(api.UserPostResponse).Id)
	assert.Equal(t, mockData.USER_EMAIL, response.Body.(api.UserPostResponse).Username)
	assert.NotEmpty(t, response.Body.(api.UserPostResponse).TempPassword)
}

func TestCreateUser_WithNotifyFalse(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")

	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.Anything).Return(mockData.USER_ID, nil)

	response, err := testSubject.CreateUser(withCtx, api.UserCreationRequest{
		Email:  mockData.USER_EMAIL,
		Notify: ptr.To(false),
	})

	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, mockData.USER_ID, response.Body.(api.UserPostResponse).Id)
	assert.Equal(t, mockData.USER_EMAIL, response.Body.(api.UserPostResponse).Username)
	assert.NotEmpty(t, response.Body.(api.UserPostResponse).TempPassword)

	iamMock.MockUserService.AssertNotCalled(t, "SendPasswordResetEmail")
}

func TestCreateUser_WithNotifyNil(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")

	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.Anything).Return(mockData.USER_ID, nil)

	response, err := testSubject.CreateUser(withCtx, api.UserCreationRequest{
		Email:  mockData.USER_EMAIL,
		Notify: nil,
	})

	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, mockData.USER_ID, response.Body.(api.UserPostResponse).Id)
	assert.Equal(t, mockData.USER_EMAIL, response.Body.(api.UserPostResponse).Username)
	assert.NotEmpty(t, response.Body.(api.UserPostResponse).TempPassword)

	iamMock.MockUserService.AssertNotCalled(t, "SendPasswordResetEmail")
}

func TestCreateUser_WithNotifyTrue_SendEmailError(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")

	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.Anything).Return(mockData.USER_ID, nil)
	iamMock.MockUserService.On("SendPasswordResetEmail", withCtx, mockData.REALM_NAME, mockData.USER_ID, resetPasswordExpirationTimeSeconds, "https://realm_name.run.ai/login/callback?runai_flow=invite", utils.RUNAI_ADMIN_UI_CLIENT_ID).Return(fmt.Errorf("email service error"))

	response, err := testSubject.CreateUser(withCtx, api.UserCreationRequest{
		Email:  mockData.USER_EMAIL,
		Notify: ptr.To(true),
	})

	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, response.Code)
	assert.Equal(t, mockData.USER_ID, response.Body.(api.UserPostResponse).Id)
	assert.Equal(t, mockData.USER_EMAIL, response.Body.(api.UserPostResponse).Username)
	assert.NotEmpty(t, response.Body.(api.UserPostResponse).TempPassword)

	iamMock.MockUserService.AssertCalled(t, "SendPasswordResetEmail", mock.Anything, mockData.REALM_NAME, mockData.USER_ID, resetPasswordExpirationTimeSeconds, "https://realm_name.run.ai/login/callback?runai_flow=invite", utils.RUNAI_ADMIN_UI_CLIENT_ID)
}

func TestBatchCreateUsers_WithNotifyTrue(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")
	userEmail1 := "user1@run.ai"
	userEmail2 := "user2@run.ai"

	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.MatchedBy(func(user *iamModels.User) bool {
		return *user.Email == userEmail1
	})).Return(mockData.USER_ID+"1", nil).Once()
	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.MatchedBy(func(user *iamModels.User) bool {
		return *user.Email == userEmail2
	})).Return(mockData.USER_ID+"2", nil).Once()

	iamMock.MockUserService.On("SendPasswordResetEmail", withCtx, mockData.REALM_NAME, mockData.USER_ID+"1", resetPasswordExpirationTimeSeconds, "https://realm_name.run.ai/login/callback?runai_flow=invite", utils.RUNAI_ADMIN_UI_CLIENT_ID).Return(nil).Once()
	iamMock.MockUserService.On("SendPasswordResetEmail", withCtx, mockData.REALM_NAME, mockData.USER_ID+"2", resetPasswordExpirationTimeSeconds, "https://realm_name.run.ai/login/callback?runai_flow=invite", utils.RUNAI_ADMIN_UI_CLIENT_ID).Return(nil).Once()

	request := api.BulkUserCreationRequest{
		Users: []api.UserCreationRequestBulk{
			{Email: userEmail1, ResetPassword: ptr.To(false)},
			{Email: userEmail2, ResetPassword: ptr.To(true)},
		},
		Notify: ptr.To(true),
	}

	response, err := testSubject.BatchCreateUsers(withCtx, request)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	bulkResponse := response.Body.(api.BulkUserCreationResponse)
	assert.Equal(t, int32(2), bulkResponse.TotalRequested)
	assert.Equal(t, int32(2), bulkResponse.TotalSuccessful)
	assert.Equal(t, int32(0), bulkResponse.TotalFailed)
	assert.Len(t, bulkResponse.Results, 2)

	iamMock.MockUserService.AssertNumberOfCalls(t, "SendPasswordResetEmail", 2)
}

func TestBatchCreateUsers_WithNotifyFalse(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")
	userEmail1 := "user1@run.ai"
	userEmail2 := "user2@run.ai"

	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.MatchedBy(func(user *iamModels.User) bool {
		return *user.Email == userEmail1
	})).Return(mockData.USER_ID+"1", nil).Once()
	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.MatchedBy(func(user *iamModels.User) bool {
		return *user.Email == userEmail2
	})).Return(mockData.USER_ID+"2", nil).Once()

	request := api.BulkUserCreationRequest{
		Users: []api.UserCreationRequestBulk{
			{Email: userEmail1, ResetPassword: ptr.To(false)},
			{Email: userEmail2, ResetPassword: ptr.To(true)},
		},
		Notify: ptr.To(false),
	}

	response, err := testSubject.BatchCreateUsers(withCtx, request)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	bulkResponse := response.Body.(api.BulkUserCreationResponse)
	assert.Equal(t, int32(2), bulkResponse.TotalRequested)
	assert.Equal(t, int32(2), bulkResponse.TotalSuccessful)
	assert.Equal(t, int32(0), bulkResponse.TotalFailed)
	assert.Len(t, bulkResponse.Results, 2)

	iamMock.MockUserService.AssertNotCalled(t, "SendPasswordResetEmail")
}

func TestBatchCreateUsers_WithNotifyNil(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	ctx := context.WithValue(context.Background(), utils.AuthRealmKey, mockData.REALM_NAME)
	withCtx := context.WithValue(ctx, runaictx.CfSubjectID, "admin")
	userEmail1 := "user1@run.ai"
	userEmail2 := "user2@run.ai"

	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.MatchedBy(func(user *iamModels.User) bool {
		return *user.Email == userEmail1
	})).Return(mockData.USER_ID+"1", nil).Once()
	iamMock.MockUserService.On("CreateUser", withCtx, mockData.REALM_NAME, mock.MatchedBy(func(user *iamModels.User) bool {
		return *user.Email == userEmail2
	})).Return(mockData.USER_ID+"2", nil).Once()

	request := api.BulkUserCreationRequest{
		Users: []api.UserCreationRequestBulk{
			{Email: userEmail1, ResetPassword: ptr.To(false)},
			{Email: userEmail2, ResetPassword: ptr.To(true)},
		},
		Notify: nil,
	}

	response, err := testSubject.BatchCreateUsers(withCtx, request)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)

	bulkResponse := response.Body.(api.BulkUserCreationResponse)
	assert.Equal(t, int32(2), bulkResponse.TotalRequested)
	assert.Equal(t, int32(2), bulkResponse.TotalSuccessful)
	assert.Equal(t, int32(0), bulkResponse.TotalFailed)
	assert.Len(t, bulkResponse.Results, 2)

	iamMock.MockUserService.AssertNotCalled(t, "SendPasswordResetEmail")
}

func TestGetRedirectUriForEmailNotification_HappyFlow(t *testing.T) {
	_, _, testSubject := initUserServiceTest(t)
	result := testSubject.getRedirectUriForEmailNotification(mockData.REALM_NAME)
	assert.Equal(t, "https://realm_name.run.ai/login/callback?runai_flow=invite", result)
}

func TestGetRedirectUriForEmailNotification_InvalidTemplate(t *testing.T) {
	_, _, testSubject := initUserServiceTest(t)
	testSubject.tenantDomainTemplate = "{{invalid template"
	result := testSubject.getRedirectUriForEmailNotification(mockData.REALM_NAME)
	assert.Equal(t, "", result)
}

func TestShouldNotify_WithTrue(t *testing.T) {
	_, _, testSubject := initUserServiceTest(t)

	result := testSubject.shouldNotify(ptr.To(true))
	assert.True(t, result)
}

func TestShouldNotify_WithFalse(t *testing.T) {
	_, _, testSubject := initUserServiceTest(t)

	result := testSubject.shouldNotify(ptr.To(false))
	assert.False(t, result)
}

func TestShouldNotify_WithNil(t *testing.T) {
	_, _, testSubject := initUserServiceTest(t)

	result := testSubject.shouldNotify(nil)
	assert.False(t, result)
}

func TestGetUsersMinimal_HappyFlow(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	expectedUsers := mockData.GetUsers()
	iamMock.MockUserService.On("GetUsers", mock.Anything, mock.Anything, mock.Anything).Return(expectedUsers, nil)
	response, err := testSubject.GetUsersMinimal(context.Background(), []string{}, "test", nil, 10)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, len(response.Body.([]api.UserMinimal)), 2)
}

func TestGetUsersMinimal_WithUsernameFilter(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	expectedUsers := mockData.GetUsers()
	iamMock.MockUserService.On("GetUsers", mock.Anything, mock.Anything, mock.Anything).Return(expectedUsers, nil)
	response, err := testSubject.GetUsersMinimal(context.Background(), []string{"username==test@run.ai"}, "", nil, 10)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestGetUsersMinimal_InvalidFilter_BadRequest(t *testing.T) {
	_, _, testSubject := initUserServiceTest(t)
	response, err := testSubject.GetUsersMinimal(context.Background(), []string{"firstName=@john"}, "", nil, 10)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestGetUsersMinimal_InternalError(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	iamMock.MockUserService.On("GetUsers", mock.Anything, mock.Anything, mock.Anything).Return(nil, iamError.ErrInternal)
	response, err := testSubject.GetUsersMinimal(context.Background(), []string{}, "test", nil, 10)
	assert.NotNil(t, err)
	assert.Nil(t, response.Body)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetUsersMinimal_NoParams_BadRequest(t *testing.T) {
	_, _, testSubject := initUserServiceTest(t)
	response, err := testSubject.GetUsersMinimal(context.Background(), nil, "", nil, 10)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestGetUsersMinimal_TooManyIDs_BadRequest(t *testing.T) {
	_, _, testSubject := initUserServiceTest(t)
	ids := []string{"u1", "u2", "u3", "u4", "u5", "u6", "u7", "u8", "u9", "u10", "u11", "u12", "u13", "u14", "u15", "u16", "u17", "u18", "u19", "u20", "u21", "u22", "u23", "u24", "u25", "u26"}
	response, err := testSubject.GetUsersMinimal(context.Background(), nil, "", ids, 10)
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestGetUsersMinimal_IDOnly(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	expectedUsers := mockData.GetUsers()
	iamMock.MockUserService.On("GetUserByID", mock.Anything, mock.Anything, "u1").Return(expectedUsers[0], nil)
	iamMock.MockUserService.On("GetUserByID", mock.Anything, mock.Anything, "u2").Return(expectedUsers[1], nil)
	response, err := testSubject.GetUsersMinimal(context.Background(), nil, "", []string{"u1", "u2"}, 10)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, 2, len(response.Body.([]api.UserMinimal)))
	iamMock.MockUserService.AssertNotCalled(t, "GetUsers", mock.Anything, mock.Anything, mock.Anything)
}

func TestGetUsersMinimal_IDLookupErrorIsSkipped(t *testing.T) {
	iamMock, _, testSubject := initUserServiceTest(t)
	expectedUsers := mockData.GetUsers()
	iamMock.MockUserService.On("GetUserByID", mock.Anything, mock.Anything, "u1").Return(expectedUsers[0], nil)
	iamMock.MockUserService.On("GetUserByID", mock.Anything, mock.Anything, "missing").Return(nil, iamError.ErrInternal)
	response, err := testSubject.GetUsersMinimal(context.Background(), nil, "", []string{"u1", "missing"}, 10)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Equal(t, 1, len(response.Body.([]api.UserMinimal)))
}
