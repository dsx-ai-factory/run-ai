/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	authMocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/auth/mocks"
	kcMocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/gocloak/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	helpers "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/repositories/users/helpers"
	runaicloakMocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/runaicloak/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	password      string = "password"
	wrongPassword string = "wrongPassword"
	realm         string = "testRealm"
	sessionID     string = "test-session-id"
)

const (
	TIME_UNIX = "1687774767"
)

func initTest(t *testing.T) (authMock *authMocks.Auth, goCloakMock *kcMocks.GoCloak, runaicloakMock *runaicloakMocks.RunAICloakClient, testSubject *KeycloakUserRepository) {
	goCloak := kcMocks.NewGoCloak(t)
	auth := authMocks.NewAuth(t)
	runaicloak := runaicloakMocks.NewRunAICloakClient(t)
	testSubject = &KeycloakUserRepository{
		auth:       auth,
		kcClient:   goCloak,
		runaicloak: runaicloak,
	}
	return auth, goCloak, runaicloak, testSubject
}

func TestCreateUser_HappyFlow(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	kcUser := getMockKcUsers()[0]
	user := getMockUsers()[0]
	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)
	goCloakMock.On("CreateUser", context.Background(), *jwt, realm, mock.Anything).Return(*kcUser.ID, nil)
	createdUserID, err := testSubject.CreateUser(context.Background(), realm, user)
	assert.Nil(t, err)
	assert.Equal(t, *kcUser.ID, createdUserID)
}

func TestCreateUser_FailToGetAdminToken(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	authMock.On("GetAdminToken", context.Background()).Return(nil, assert.AnError)
	createdUserID, err := testSubject.CreateUser(context.Background(), realm, &models.User{})
	assert.Error(t, err)
	assert.Empty(t, createdUserID)
}

func TestCreateUser_FailToCreateUser(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	user := getMockUsers()[0]
	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)
	goCloakMock.On("CreateUser", context.Background(), *jwt, realm, mock.Anything).Return("", assert.AnError)
	createdUserID, err := testSubject.CreateUser(context.Background(), realm, user)
	assert.Error(t, err)
	assert.Empty(t, createdUserID)
}

func TestGetUserByID_HappyFlow(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()

	kcUser := getMockKcUsers()[0]
	user := getMockUsers()[0]

	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	goCloakMock.On("GetUserByID", ctx, *jwt, realm, *kcUser.ID).Return(kcUser, nil)

	result, err := testSubject.GetUserByID(ctx, realm, *kcUser.ID)
	assert.NoError(t, err)
	assert.Equal(t, user, result)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetUserByID_FailToGetAdminToken(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	ctx := context.Background()

	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	result, err := testSubject.GetUserByID(ctx, realm, "1")
	assert.Error(t, err)
	assert.Nil(t, result)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetUserByID_FailToGetUserByID(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()

	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	goCloakMock.On("GetUserByID", ctx, *jwt, realm, "1").Return(nil, assert.AnError)

	result, err := testSubject.GetUserByID(ctx, realm, "1")
	assert.Error(t, err)
	assert.Nil(t, result)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetUserByEmail_HappyFlow(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()

	kcUsers := getMockKcUsers()
	user := getMockUsers()[0]

	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	searchParams := gocloak.GetUsersParams{
		Email: kcUsers[0].Email,
	}
	goCloakMock.On("GetUsers", ctx, *jwt, realm, searchParams).Return(kcUsers, nil)

	result, err := testSubject.GetUserByEmail(ctx, realm, *kcUsers[0].Email)
	assert.NoError(t, err)
	assert.Equal(t, user, result)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetUserByEmail_FailToGetAdminToken(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	ctx := context.Background()

	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	result, err := testSubject.GetUserByEmail(ctx, realm, "")
	assert.Error(t, err)
	assert.Nil(t, result)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetUserByEmail_FailToGetUserByEmail(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	kcUsers := getMockKcUsers()
	wrongEmail := gocloak.StringP("notexist@gmail.com")
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	searchParams := gocloak.GetUsersParams{
		Email: wrongEmail,
	}
	goCloakMock.On("GetUsers", ctx, *jwt, realm, searchParams).Return(kcUsers, nil)

	result, err := testSubject.GetUserByEmail(ctx, realm, *wrongEmail)
	assert.ErrorContains(t, err, "not found")
	assert.Nil(t, result)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestDeleteUser_HappyFlow(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	kcUser := getMockKcUsers()[0]

	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	goCloakMock.On("DeleteUser", ctx, *jwt, realm, *kcUser.ID).Return(nil)
	err := testSubject.DeleteUser(ctx, realm, *kcUser.ID)
	assert.NoError(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestDeleteUser_FailToGetAdminToken(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	err := testSubject.DeleteUser(ctx, realm, "")
	assert.Error(t, err)
	authMock.AssertExpectations(t)
}

func TestDeleteUser_FailToDeleteUser(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	kcUser := getMockKcUsers()[0]

	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	goCloakMock.On("DeleteUser", ctx, *jwt, realm, *kcUser.ID).Return(assert.AnError)
	err := testSubject.DeleteUser(ctx, realm, *kcUser.ID)
	assert.Error(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestLogoutUser_HappyFlow(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	kcUser := getMockKcUsers()[0]

	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	goCloakMock.On("LogoutAllSessions", ctx, *jwt, realm, *kcUser.ID).Return(nil)
	err := testSubject.LogoutUser(ctx, realm, *kcUser.ID)
	assert.NoError(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestLogoutUser_FailToGetAdminToken(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	err := testSubject.LogoutUser(ctx, realm, "")
	assert.Error(t, err)
	authMock.AssertExpectations(t)
}

func TestLogoutUser_FailToLogoutUser(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	kcUser := getMockKcUsers()[0]

	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	goCloakMock.On("LogoutAllSessions", ctx, *jwt, realm, *kcUser.ID).Return(assert.AnError)
	err := testSubject.LogoutUser(ctx, realm, *kcUser.ID)
	assert.Error(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestChangeUserPassword_HappyFlow(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	LoginResponse := kcMocks.GetMockJWT()
	ctx := context.Background()
	user := getMockUsers()[0]

	goCloakMock.On("Login", ctx, "runai-cli", "", realm, *user.Username, password).Return(LoginResponse, nil)
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	goCloakMock.On("SetPassword", ctx, *jwt, *user.ID, realm, password, false).Return(nil)
	err := testSubject.ChangeUserPassword(ctx, realm, user, password, password)
	assert.NoError(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestChangeUserPassword_FailedToLogin(t *testing.T) {
	_, goCloakMock, _, testSubject := initTest(t)
	ctx := context.Background()
	user := getMockUsers()[0]

	goCloakMock.On("Login", ctx, "runai-cli", "", realm, *user.Username, password).Return(nil, assert.AnError)
	err := testSubject.ChangeUserPassword(ctx, realm, user, password, password)
	assert.Error(t, err)
	goCloakMock.AssertExpectations(t)
}

func TestChangeUserPassword_FailToGetAdminToken(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	ctx := context.Background()
	LoginResponse := kcMocks.GetMockJWT()
	user := getMockUsers()[0]

	goCloakMock.On("Login", ctx, "runai-cli", "", realm, *user.Username, password).Return(LoginResponse, nil)
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	err := testSubject.ChangeUserPassword(ctx, realm, user, password, password)
	assert.Error(t, err)
	authMock.AssertExpectations(t)
}

func TestChangeUserPassword_FailToChangePassword(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	LoginResponse := kcMocks.GetMockJWT()
	ctx := context.Background()
	user := getMockUsers()[0]

	goCloakMock.On("Login", ctx, "runai-cli", "", realm, *user.Username, password).Return(LoginResponse, nil)
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	goCloakMock.On("SetPassword", ctx, *jwt, *user.ID, realm, password, false).Return(assert.AnError)
	err := testSubject.ChangeUserPassword(ctx, realm, user, password, password)
	assert.Error(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestResetPassword_HappyFlow(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	userID := getMockUsers()[0].ID

	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	goCloakMock.On("SetPassword", ctx, *jwt, *userID, realm, password, true).Return(nil)
	err := testSubject.ResetUserPassword(ctx, realm, *userID, password)
	assert.NoError(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestResetPassword_FailToGetAdminToken(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	err := testSubject.ResetUserPassword(ctx, realm, "", password)
	assert.Error(t, err)
	authMock.AssertExpectations(t)
}

func TestResetPassword_FailToResetPassword(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	userID := getMockUsers()[0].ID

	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	goCloakMock.On("SetPassword", ctx, *jwt, *userID, realm, password, true).Return(assert.AnError)
	err := testSubject.ResetUserPassword(ctx, realm, *userID, password)
	assert.Error(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestSendPasswordResetEmail_HappyFlow(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	userID := getMockUsers()[0].ID
	expirationSeconds := 604800 // 1 week

	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	testClientId := "test-client-id"
	goCloakMock.On("ExecuteActionsEmail", ctx, *jwt, realm, mock.MatchedBy(func(params gocloak.ExecuteActionsEmail) bool {
		return *params.UserID == *userID &&
			len(*params.Actions) == 1 &&
			(*params.Actions)[0] == helpers.UPDATE_PASSWORD_ACTION &&
			*params.Lifespan == expirationSeconds &&
			*params.ClientID == testClientId
	})).Return(nil)

	err := testSubject.SendPasswordResetEmail(ctx, realm, *userID, expirationSeconds, "https://mock-q7z3k.example.com", testClientId)
	assert.NoError(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestSendPasswordResetEmail_FailToGetAdminToken(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	err := testSubject.SendPasswordResetEmail(ctx, realm, "user-id", 3600, "https://mock-q7z3k.example.com", "test-client-id")
	assert.Error(t, err)
	authMock.AssertExpectations(t)
}

func TestSendPasswordResetEmail_FailToExecuteActionsEmail(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	userID := getMockUsers()[0].ID
	expirationSeconds := 604800 // 1 week

	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	testClientId := "test-client-id"
	goCloakMock.On("ExecuteActionsEmail", ctx, *jwt, realm, mock.MatchedBy(func(params gocloak.ExecuteActionsEmail) bool {
		return *params.UserID == *userID &&
			len(*params.Actions) == 1 &&
			(*params.Actions)[0] == helpers.UPDATE_PASSWORD_ACTION &&
			*params.Lifespan == expirationSeconds &&
			*params.ClientID == testClientId
	})).Return(assert.AnError)

	err := testSubject.SendPasswordResetEmail(ctx, realm, *userID, expirationSeconds, "https://mock-q7z3k.example.com", testClientId)
	assert.Error(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestToggleRequiredAction_HappyFlow(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	requiredAction := getKcRequiredActionsMock()[0]
	alias := models.RequiredActionVerifyProfile
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	goCloakMock.On("GetRequiredAction", ctx, *jwt, realm, *requiredAction.Alias).Return(requiredAction, nil)
	goCloakMock.On("UpdateRequiredAction", ctx, *jwt, realm, *requiredAction).Return(nil)
	err := testSubject.ToggleRequiredAction(ctx, realm, alias, true)
	assert.NoError(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestToggleRequiredAction_FailToGetAdminToken(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	err := testSubject.ToggleRequiredAction(ctx, realm, "", true)
	assert.Error(t, err)
	authMock.AssertExpectations(t)
}

func TestToggleRequiredAction_FailToGetRequiredAction(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	requiredAction := getKcRequiredActionsMock()[0]
	alias := models.RequiredActionVerifyProfile
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	goCloakMock.On("GetRequiredAction", ctx, *jwt, realm, *requiredAction.Alias).Return(nil, assert.AnError)
	err := testSubject.ToggleRequiredAction(ctx, realm, alias, true)
	assert.Error(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestToggleRequiredAction_FailToUpdateRequiredAction(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	requiredAction := getKcRequiredActionsMock()[0]
	alias := models.RequiredActionVerifyProfile
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	goCloakMock.On("GetRequiredAction", ctx, *jwt, realm, *requiredAction.Alias).Return(requiredAction, nil)
	goCloakMock.On("UpdateRequiredAction", ctx, *jwt, realm, *requiredAction).Return(assert.AnError)
	err := testSubject.ToggleRequiredAction(ctx, realm, alias, true)
	assert.Error(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetCustomUsers_HappyFlow(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	httpReq := getHttpServer(getUsersResponse())
	defer httpReq.Close()
	authMock.On("GetHost").Return(httpReq.URL)

	users, err := testSubject.getCustomUsers(ctx, "runai", map[string]string{})
	assert.Nil(t, err)
	assert.Equal(t, 2, users.Count)
	assert.NotNil(t, users.Users)
}

func TestGetCustomUsers_AdminTokenError(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, fmt.Errorf("error"))

	users, err := testSubject.getCustomUsers(ctx, "runai", map[string]string{})
	assert.Nil(t, users)
	assert.NotNil(t, err)
}

func TestGetCustomUsers_MarshalFailure(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	httpReq := getHttpServer([]byte("fail"))
	defer httpReq.Close()
	authMock.On("GetHost").Return(httpReq.URL)

	users, err := testSubject.getCustomUsers(ctx, "runai", map[string]string{})
	assert.Nil(t, users)
	assert.NotNil(t, err)
}

func TestGetUsers_HappyFlow(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	httpReq := getHttpServer(getUsersResponse())
	defer httpReq.Close()
	authMock.On("GetHost").Return(httpReq.URL)

	users, err := testSubject.GetUsers(ctx, "runai", getUserFilters(gocloak.StringP("username"), gocloak.StringP("q"), gocloak.IntP(1), gocloak.IntP(1), gocloak.StringP("sortBy"), gocloak.StringP("sortOrder"), gocloak.StringP("creationFilter"), gocloak.StringP("loginFilter"), gocloak.StringP("updateFilter"), gocloak.StringP("createdBy")))
	assert.Nil(t, err)
	assert.Equal(t, 2, len(users))
}

func TestGetUsers_KcError(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, fmt.Errorf("error"))

	users, err := testSubject.GetUsers(ctx, "runai", getUserFilters(gocloak.StringP("username"), gocloak.StringP("q"), gocloak.IntP(1), gocloak.IntP(1), gocloak.StringP("sortBy"), gocloak.StringP("sortOrder"), gocloak.StringP("creationFilter"), gocloak.StringP("loginFilter"), gocloak.StringP("updateFilter"), gocloak.StringP("createdBy")))
	assert.Nil(t, users)
	assert.NotNil(t, err)
}

func TestSetQueryParam_HappyFlow(t *testing.T) {
	params := setQueryParamMap(getUserFilters(gocloak.StringP("username"), gocloak.StringP("q"), gocloak.IntP(1), gocloak.IntP(1), gocloak.StringP("sortBy"), gocloak.StringP("sortOrder"), gocloak.StringP("creationFilter"), gocloak.StringP("loginFilter"), gocloak.StringP("updateFilter"), gocloak.StringP("createdBy")))
	assert.Equal(t, 10, len(params))
	assert.Equal(t, "username", params[Username])
	assert.Equal(t, "q", params[Q])
	assert.Equal(t, "1", params[Max])
	assert.Equal(t, "1", params[Offset])
	assert.Equal(t, "sortBy", params[SortBy])
	assert.Equal(t, "sortOrder", params[SortOrder])
	assert.Equal(t, "creationFilter", params[CreationFilter])
	assert.Equal(t, "loginFilter", params[LoginFilter])
	assert.Equal(t, "updateFilter", params[UpdateFilter])
	assert.Equal(t, "createdBy", params[CreatedBy])
}

func TestLogoutUserBySessionID_HappyFlow(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()

	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	goCloakMock.On("LogoutUserSession", ctx, *jwt, realm, sessionID).Return(nil)
	err := testSubject.LogoutUserBySessionID(ctx, realm, sessionID)
	assert.NoError(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestLogoutUserBySessionID_FailToGetAdminToken(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	err := testSubject.LogoutUserBySessionID(ctx, realm, "")
	assert.Error(t, err)
	authMock.AssertExpectations(t)
}

func TestLogoutUserBySessionID_FailToLogoutSession(t *testing.T) {
	authMock, goCloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()

	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	goCloakMock.On("LogoutUserSession", ctx, *jwt, realm, sessionID).Return(assert.AnError)
	err := testSubject.LogoutUserBySessionID(ctx, realm, sessionID)
	assert.Error(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestToggleUnmanagedUsersAttributes_HappyFlow_Enable(t *testing.T) {
	authMock, goCloakMock, runaicloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	attributesJson := map[string]interface{}{
		"existingKey":              "existingValue",
		"unmanagedAttributePolicy": "DISABLED",
	}
	runaicloakMock.On("GetRealmUserAttributes", ctx, realm, *jwt).Return(attributesJson, nil)
	runaicloakMock.On("ToggleUserAttribute", ctx, realm, *jwt, attributesJson, UnmanagedAttributesKey, true).Return(nil)

	authMock.On("GetAdminToken", ctx).Return(jwt, nil)

	err := testSubject.ToggleUnmanagedUsersAttributes(ctx, realm, true)
	assert.NoError(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
	runaicloakMock.AssertExpectations(t)
}

func TestToggleUnmanagedUsersAttributes_HappyFlow_Disable(t *testing.T) {
	authMock, goCloakMock, runaicloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	attributesJson := map[string]interface{}{
		"existingKey":              "existingValue",
		"unmanagedAttributePolicy": "ENABLED",
	}
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	runaicloakMock.On("GetRealmUserAttributes", ctx, realm, *jwt).Return(attributesJson, nil)
	runaicloakMock.On("ToggleUserAttribute", ctx, realm, *jwt, attributesJson, UnmanagedAttributesKey, false).Return(nil)
	err := testSubject.ToggleUnmanagedUsersAttributes(ctx, realm, false)
	assert.NoError(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestToggleUnmanagedUsersAttributes_FailToGetAdminToken(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	ctx := context.Background()

	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	err := testSubject.ToggleUnmanagedUsersAttributes(ctx, realm, true)
	assert.Error(t, err)
	authMock.AssertExpectations(t)
}

func TestToggleUnmanagedUsersAttributes_FailToGetRealmUserAttributes(t *testing.T) {
	authMock, goCloakMock, runaicloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()

	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	runaicloakMock.On("GetRealmUserAttributes", ctx, realm, *jwt).Return(nil, assert.AnError)
	err := testSubject.ToggleUnmanagedUsersAttributes(ctx, realm, true)
	assert.Error(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestToggleUnmanagedUsersAttributes_FailToUpdateAttributes(t *testing.T) {
	authMock, goCloakMock, runaicloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	attributesJson := map[string]interface{}{
		"existingKey": "existingValue",
	}
	runaicloakMock.On("GetRealmUserAttributes", ctx, realm, *jwt).Return(attributesJson, nil)
	runaicloakMock.On("ToggleUserAttribute", ctx, realm, *jwt, attributesJson, UnmanagedAttributesKey, true).Return(assert.AnError)

	authMock.On("GetAdminToken", ctx).Return(jwt, nil)

	err := testSubject.ToggleUnmanagedUsersAttributes(ctx, realm, true)
	assert.Error(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func getUserFilters(username *string, q *string, max *int, offset *int, sortBy *string, sortOrder *string, creationFilter *string, loginFilter *string, updateFilter *string, createdBy *string) models.UserFilters {
	return models.UserFilters{
		Username:       username,
		Q:              q,
		Max:            max,
		Offset:         offset,
		SortBy:         sortBy,
		SortOrder:      sortOrder,
		CreationFilter: creationFilter,
		LoginFilter:    loginFilter,
		UpdateFilter:   updateFilter,
		CreatedBy:      createdBy,
	}
}

func getHttpServer(respJson []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		res.Write(respJson)
	}))
}

func getUsersResponse() []byte {
	respJson, _ := json.Marshal(UsersResponse{
		Count: 2,
		Users: []map[string]string{{"createdAt": "1726744334", "lastLogin": "", "createdBy": "", "groups": "", "id": "56ad39cd-3ef3-46b0-bc2f-b431134535dc", "email": "runai-test@run.ai", "username": "runai-test@run.ai", "isLocal": "false", "updatedAt": "1726744334"}, {"createdAt": "1726744334", "lastLogin": "1726746236", "createdBy": "", "groups": "", "id": "adf65d15-1bc7-4884-a763-895f65c0ca58", "email": "test@run.ai", "username": "test@run.ai", "isLocal": "true", "updatedAt": "1726744334"}},
	})
	return respJson
}

func getMockKcUsers() []*gocloak.User {
	return []*gocloak.User{
		{
			ID:        gocloak.StringP("1"),
			Username:  gocloak.StringP("test1@run.ai"),
			FirstName: gocloak.StringP("first"),
			LastName:  gocloak.StringP("last"),
			Email:     gocloak.StringP("test1@run.ai"),
			Attributes: &map[string][]string{
				utils.UpdateTime:    {TIME_UNIX},
				utils.LastLoginTime: {TIME_UNIX},
				utils.CreationTime:  {TIME_UNIX},
				utils.CreatedBy:     {"admin@run.ai"},
				utils.IsLocalUser:   {"true"},
			},
		},
		{
			ID:        gocloak.StringP("2"),
			Username:  gocloak.StringP("test1@run.ai"),
			FirstName: gocloak.StringP("first"),
			LastName:  gocloak.StringP("last"),
			Email:     gocloak.StringP("test2@run.ai"),
			Attributes: &map[string][]string{
				utils.UpdateTime:    {TIME_UNIX},
				utils.LastLoginTime: {TIME_UNIX},
				utils.CreationTime:  {TIME_UNIX},
				utils.CreatedBy:     {"admin@run.ai"},
				utils.IsLocalUser:   {"false"},
			},
		},
	}
}

func getMockUsers() []*models.User {
	return []*models.User{
		{
			Username:  gocloak.StringP("test1@run.ai"),
			Email:     gocloak.StringP("test1@run.ai"),
			ID:        gocloak.StringP("1"),
			CreatedBy: gocloak.StringP("admin@run.ai"),
			CreatedAt: utils.FormatDate(TIME_UNIX),
			UpdatedAt: utils.FormatDate(TIME_UNIX),
			LastLogin: utils.FormatDate(TIME_UNIX),
			IsLocal:   gocloak.BoolP(true),
			FirstName: gocloak.StringP("first"),
			LastName:  gocloak.StringP("last"),
		},
		{
			Username:  gocloak.StringP("test2@run.ai"),
			Email:     gocloak.StringP("test2@run.ai"),
			ID:        gocloak.StringP("2"),
			CreatedBy: gocloak.StringP("admin@run.ai"),
			CreatedAt: utils.FormatDate(TIME_UNIX),
			UpdatedAt: utils.FormatDate(TIME_UNIX),
			LastLogin: utils.FormatDate(TIME_UNIX),
			IsLocal:   gocloak.BoolP(false),
			FirstName: gocloak.StringP("first"),
			LastName:  gocloak.StringP("last"),
		},
	}
}

func getKcRequiredActionsMock() []*gocloak.RequiredActionProviderRepresentation {
	return []*gocloak.RequiredActionProviderRepresentation{
		{
			Alias:   gocloak.StringP(string(models.RequiredActionVerifyProfile)),
			Enabled: gocloak.BoolP(true),
		},
	}
}
