/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package users

import (
	"strconv"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils"
	"github.com/Nerzal/gocloak/v13"
	"github.com/stretchr/testify/assert"
)

const (
	CREATED_BY_USER = "test@run.ai"
	EMAIL           = "test@run.ai"
	USERNAME        = "tese"
	ID              = "id"
	TIME_UNIX       = "1687774767"
)

const (
	Group1 = "Group1"
	Group2 = "Group2"
	Group3 = "Group3"
)

func TestGetRequiredUserActions_resetPassword_True(t *testing.T) {
	requiredActions := getRequiredUserActions(gocloak.BoolP(true))
	assert.Equal(t, 1, len(*requiredActions))
	assert.Equal(t, UPDATE_PASSWORD_ACTION, (*requiredActions)[0])
}

func TestGetRequiredUserActions_resetPassword_Nil(t *testing.T) {
	requiredActions := getRequiredUserActions(nil)
	assert.Equal(t, 1, len(*requiredActions))
	assert.Equal(t, UPDATE_PASSWORD_ACTION, (*requiredActions)[0])
}

func TestGetRequiredUserActions_resetPassword_False(t *testing.T) {
	requiredActions := getRequiredUserActions(gocloak.BoolP(false))
	assert.Nil(t, requiredActions)
}

func TestGenerateUserAttributes_HappyFlow(t *testing.T) {
	attributes := generateUserAttributes(CREATED_BY_USER)
	assert.NotNil(t, attributes)
	assert.Equal(t, (*attributes)[utils.CreatedBy], []string{CREATED_BY_USER})
	assert.Equal(t, (*attributes)[utils.IsLocalUser], []string{"true"})
	assert.NotEmpty(t, (*attributes)[utils.CreationTime])
	assert.NotEmpty(t, (*attributes)[utils.UpdateTime])
}

func TestGenerateUserAttributes_EmptyCreatedBy(t *testing.T) {
	attributes := generateUserAttributes("")
	assert.NotNil(t, attributes)
	assert.Nil(t, (*attributes)[utils.CreatedBy])
	assert.Equal(t, (*attributes)[utils.IsLocalUser], []string{"true"})
	assert.NotEmpty(t, (*attributes)[utils.CreationTime])
	assert.NotEmpty(t, (*attributes)[utils.UpdateTime])
}

func TestGetEmailWithFallback_HappyFlow(t *testing.T) {
	kcUser := getGoCloakUser(gocloak.StringP(EMAIL), gocloak.StringP(USERNAME))
	res := getEmailWithFallback(kcUser)
	assert.Equal(t, EMAIL, res)
}

func TestGetEmailWithFallback_UsernameFallback_nil_email(t *testing.T) {
	kcUser := getGoCloakUser(nil, gocloak.StringP(USERNAME))
	res := getEmailWithFallback(kcUser)
	assert.Equal(t, USERNAME, res)
}

func TestGetEmailWithFallback_UsernameFallback_empty_email(t *testing.T) {
	kcUser := getGoCloakUser(gocloak.StringP(""), gocloak.StringP(USERNAME))
	res := getEmailWithFallback(kcUser)
	assert.Equal(t, USERNAME, res)
}

func TestGetEmailWithFallback_empty_data(t *testing.T) {
	kcUser := getGoCloakUser(nil, nil)
	res := getEmailWithFallback(kcUser)
	assert.Equal(t, "", res)
}

func Test_GetUserAttributeSlice(t *testing.T) {
	kcUser := getGoCloakSSOUser()
	groups := getUserAttributeSlice(kcUser.Attributes, GROUPS)
	assert.Equal(t, &[]string{"Group1", "Group2", "Group3"}, groups)
}

func Test_GetUserAttributeSliceNotExist(t *testing.T) {
	kcUser := getGoCloakSSOUser()
	groups := getUserAttributeSlice(kcUser.Attributes, "Not exist")
	assert.Equal(t, (*[]string)(nil), groups)
}

func Test_GetUserAttributeSliceNil(t *testing.T) {
	groups := getUserAttributeSlice(nil, "nil")
	assert.Equal(t, (*[]string)(nil), groups)
}

func Test_GetUserAttribute(t *testing.T) {
	kcUser := getGoCloakUser(gocloak.StringP(EMAIL), gocloak.StringP(USERNAME))
	isLocal := getUserAttribute(kcUser.Attributes, utils.IsLocalUser)
	assert.Equal(t, "true", isLocal)
}

func Test_GetUserAttributeNotExist(t *testing.T) {
	kcUser := getGoCloakUser(gocloak.StringP(EMAIL), gocloak.StringP(USERNAME))
	groups := getUserAttribute(kcUser.Attributes, "Not exist")
	assert.Equal(t, "", groups)
}

func Test_GetUserAttributeNil(t *testing.T) {
	groups := getUserAttribute(nil, "nil")
	assert.Equal(t, "", groups)
}

func Test_ToUserModelFromCustomUser_HappyFlow(t *testing.T) {
	input := map[string]string{
		"username":  EMAIL,
		"id":        ID,
		"email":     EMAIL,
		"createdBy": CREATED_BY_USER,
		"createdAt": TIME_UNIX,
		"updatedAt": TIME_UNIX,
		"lastLogin": TIME_UNIX,
		"isLocal":   "true",
		"groups":    "group1,group2",
	}
	user := ToUserModelFromCustomUser(input)
	assert.Equal(t, ID, *user.ID)
	assert.Equal(t, EMAIL, *user.Username)
	assert.Equal(t, EMAIL, *user.Email)
	assert.Equal(t, CREATED_BY_USER, *user.CreatedBy)
}

func Test_ToUserModel(t *testing.T) {
	kcUser := getGoCloakSSOUser()
	user := ToUserModel(kcUser)
	assert.Equal(t, ID, *user.ID)
	assert.Equal(t, EMAIL, *user.Username)
	assert.Equal(t, EMAIL, *user.Email)
	assert.Equal(t, CREATED_BY_USER, *user.CreatedBy)
	unixTime, err := strconv.ParseInt(TIME_UNIX, 10, 64)
	if err != nil {
		t.Fatalf("Failed to parse TIME_UNIX: %v", err)
	}
	assert.Equal(t, unixTime, user.CreatedAt.Unix())
	assert.Equal(t, unixTime, user.LastLogin.Unix())
	assert.Equal(t, unixTime, user.UpdatedAt.Unix())
}

func getGoCloakUser(email, username *string) gocloak.User {
	return gocloak.User{
		ID:       gocloak.StringP(ID),
		Email:    email,
		Username: username,
		Attributes: &map[string][]string{
			utils.IsLocalUser: {"true"},
		},
	}
}

func getGoCloakSSOUser() gocloak.User {
	return gocloak.User{
		ID:       gocloak.StringP(ID),
		Username: gocloak.StringP(EMAIL),
		Email:    gocloak.StringP(EMAIL),
		Attributes: &map[string][]string{
			utils.UpdateTime:    {TIME_UNIX},
			utils.LastLoginTime: {TIME_UNIX},
			utils.CreationTime:  {TIME_UNIX},
			utils.CreatedBy:     {CREATED_BY_USER},
			GROUPS:              {Group1, Group2, Group3},
		},
	}
}
