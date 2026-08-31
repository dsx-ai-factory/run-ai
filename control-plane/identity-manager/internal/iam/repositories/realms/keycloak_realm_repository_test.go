/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package realms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	authMocks "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/auth/mocks"
	kcMocks "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/gocloak/mocks"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	helpers "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/repositories/realms/helpers"
	runaicloakMocks "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/runaicloak/mocks"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/utils"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"
)

const (
	realm              = "test-realm"
	idpAlias           = "oidc"
	authFlowID         = "authFlowID"
	authFlowProviderID = "identity-provider-redirector"
)

func initTest(t *testing.T) (authMock *authMocks.Auth, goCloakMock *kcMocks.GoCloak, runaicloakMock *runaicloakMocks.RunAICloakClient, testSubject *KeycloakRealmRepository) {
	goCloak := kcMocks.NewGoCloak(t)
	auth := authMocks.NewAuth(t)
	runaicloakClient := runaicloakMocks.NewRunAICloakClient(t)
	testSubject = &KeycloakRealmRepository{
		auth:             auth,
		kcClient:         goCloak,
		runaicloakClient: runaicloakClient,
	}
	return auth, goCloak, runaicloakClient, testSubject
}

func getMockRealms() []*models.Realm {
	realms := []*models.Realm{}
	for i := range 10 {
		realmName := fmt.Sprintf("test%d", i)
		theme := "runai-theme"
		defaults := &models.RealmDefaults{
			Theme:         &theme,
			SmtpHost:      "test-smtp-host",
			SmtpPort:      2525,
			SmtpFromEmail: "test@example.com",
		}
		realm := models.NewRunaiRealm(realmName, defaults)
		realms = append(realms, realm)
	}
	return realms
}

func getMockKcRealms() []*gocloak.RealmRepresentation {
	realms := getMockRealms()
	kcRealms := []*gocloak.RealmRepresentation{}
	for _, realm := range realms {
		kcRealm := helpers.FromRealmModel(realm, &gocloak.RealmRepresentation{})
		kcRealms = append(kcRealms, kcRealm)
	}
	return kcRealms
}

func TestCreateRealm_HappyFlow(t *testing.T) {
	authMock, iamMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	kcRealm := getMockKcRealms()[0]
	realm := getMockRealms()[0]
	realmName := *kcRealm.Realm
	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)
	iamMock.On("CreateRealm", context.Background(), *jwt, *kcRealm).Return(realmName, nil)

	httpReq := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			res.WriteHeader(http.StatusOK)
			res.Write(getUserProfileMetadataResponse())
		}
		if req.Method == http.MethodPut {
			res.WriteHeader(http.StatusOK)
			res.Write([]byte{})
		}
	}))
	defer httpReq.Close()
	authMock.On("GetHost").Return(httpReq.URL)

	createdRealm, err := testSubject.CreateRealm(context.Background(), realm)
	assert.Nil(t, err)
	assert.Equal(t, realmName, createdRealm)
}

func getUserProfileMetadataResponse() []byte {
	externalTokenReponse := models.UserProfileMetadata{
		Attributes: nil,
		Groups:     nil,
	}
	jsonResponse, _ := json.Marshal(externalTokenReponse)
	return jsonResponse
}

func TestCreateRealm_FailToCreateRealm(t *testing.T) {
	authMock, iamMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	kcRealm := getMockKcRealms()[0]
	realm := getMockRealms()[0]
	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)
	iamMock.On("CreateRealm", context.Background(), *jwt, *kcRealm).Return("", kcMocks.GetKeycloakError())
	createdRealm, err := testSubject.CreateRealm(context.Background(), realm)
	assert.NotNil(t, err)
	assert.Empty(t, createdRealm)
}

func TestCreateRealm_FailToGetAdminToken(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	authMock.On("GetAdminToken", context.Background()).Return(nil, kcMocks.GetKeycloakError())
	realmName := "test"
	createdRealm, err := testSubject.CreateRealm(context.Background(), &models.Realm{Name: &realmName})
	assert.NotNil(t, err)
	assert.Empty(t, createdRealm)
}

func TestDisableRealm_HappyFlow(t *testing.T) {
	authMock, iamMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	kcRealm := getMockKcRealms()[0]
	realmName := *kcRealm.Realm
	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)
	iamMock.On("GetRealm", context.Background(), *jwt, realmName).Return(kcRealm, nil)

	disabledRealm := *kcRealm
	disabledRealm.Enabled = gocloak.BoolP(false)

	httpReq := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			res.WriteHeader(http.StatusOK)
			res.Write(getUserProfileMetadataResponse())
		}
		if req.Method == http.MethodPut {
			res.WriteHeader(http.StatusOK)
			res.Write([]byte{})
		}
	}))
	defer httpReq.Close()
	authMock.On("GetHost").Return(httpReq.URL)

	iamMock.On("UpdateRealm", context.Background(), *jwt, disabledRealm).Return(nil)
	err := testSubject.DisableRealm(context.Background(), realmName)
	assert.Nil(t, err)
}

func TestDisableRealm_FailToGetRealm(t *testing.T) {
	authMock, iamMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	kcRealm := getMockKcRealms()[0]
	realmName := *kcRealm.Realm
	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)
	iamMock.On("GetRealm", context.Background(), *jwt, realmName).Return(nil, kcMocks.GetKeycloakError())
	err := testSubject.DisableRealm(context.Background(), realmName)
	assert.NotNil(t, err)
}

func TestDisableRealm_FailToUpdateRealm(t *testing.T) {
	authMock, iamMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	kcRealm := getMockKcRealms()[0]
	realmName := *kcRealm.Realm
	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)
	iamMock.On("GetRealm", context.Background(), *jwt, realmName).Return(kcRealm, nil)

	disabledRealm := *kcRealm
	disabledRealm.Enabled = gocloak.BoolP(false)

	iamMock.On("UpdateRealm", context.Background(), *jwt, disabledRealm).Return(kcMocks.GetKeycloakError())
	err := testSubject.DisableRealm(context.Background(), realmName)
	assert.NotNil(t, err)
}

func TestDisableRealm_FailToGetAdminToken(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	authMock.On("GetAdminToken", context.Background()).Return(nil, kcMocks.GetKeycloakError())
	err := testSubject.DisableRealm(context.Background(), "test")
	assert.NotNil(t, err)
}

func TestGetRealm_HappyFlow(t *testing.T) {
	authMock, iamMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	kcRealm := getMockKcRealms()[0]
	realmName := *kcRealm.Realm
	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)
	iamMock.On("GetRealm", context.Background(), *jwt, realmName).Return(kcRealm, nil)

	expectedRealm := &models.Realm{
		Name:    kcRealm.Realm,
		Enabled: kcRealm.Enabled,
	}

	actualRealm, err := testSubject.GetRealm(context.Background(), realmName)
	assert.Nil(t, err)
	assert.ObjectsAreEqual(expectedRealm, actualRealm)
}

func TestGetRealm_FailToGetRealm(t *testing.T) {
	authMock, iamMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	kcRealm := getMockKcRealms()[0]
	realmName := *kcRealm.Realm
	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)
	iamMock.On("GetRealm", context.Background(), *jwt, realmName).Return(nil, kcMocks.GetKeycloakError())

	actualRealm, err := testSubject.GetRealm(context.Background(), realmName)
	assert.NotNil(t, err)
	assert.Nil(t, actualRealm)
}

func TestGetRealm_FailToGetAdminToken(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	authMock.On("GetAdminToken", context.Background()).Return(nil, kcMocks.GetKeycloakError())

	actualRealm, err := testSubject.GetRealm(context.Background(), "test")
	assert.NotNil(t, err)
	assert.Nil(t, actualRealm)
}

func TestGetRealms_HappyFlow(t *testing.T) {
	authMock, iamMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	kcRealm := getMockKcRealms()
	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)
	iamMock.On("GetRealms", context.Background(), *jwt).Return(kcRealm, nil)

	expectedRealms := []*models.Realm{
		{
			Name:    kcRealm[0].Realm,
			Enabled: kcRealm[0].Enabled,
		},
		{
			Name:    kcRealm[1].Realm,
			Enabled: kcRealm[1].Enabled,
		},
	}

	actualRealms, err := testSubject.GetRealms(context.Background())
	assert.Nil(t, err)
	assert.ObjectsAreEqual(expectedRealms, actualRealms)
}

func TestGetRealms_FailToGetRealms(t *testing.T) {
	authMock, iamMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)
	iamMock.On("GetRealms", context.Background(), *jwt).Return(nil, kcMocks.GetKeycloakError())

	actualRealms, err := testSubject.GetRealms(context.Background())
	assert.NotNil(t, err)
	assert.Nil(t, actualRealms)
}

func TestGetRealms_FailToGetAdminToken(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	authMock.On("GetAdminToken", context.Background()).Return(nil, kcMocks.GetKeycloakError())

	actualRealms, err := testSubject.GetRealms(context.Background())
	assert.NotNil(t, err)
	assert.Nil(t, actualRealms)
}

func TestUpdateRealm_HappyFlow(t *testing.T) {
	authMock, iamMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	kcRealm := getMockKcRealms()[0]
	realm := getMockRealms()[0]
	realmName := *realm.Name

	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)
	iamMock.On("GetRealm", context.Background(), *jwt, realmName).Return(kcRealm, nil)

	realm.Enabled = gocloak.BoolP(false)

	expectedkcRealm := kcRealm
	expectedkcRealm.Enabled = gocloak.BoolP(false)
	iamMock.On("UpdateRealm", context.Background(), *jwt, *expectedkcRealm).Return(nil)

	httpReq := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			res.WriteHeader(http.StatusOK)
			res.Write(getUserProfileMetadataResponse())
		}
		if req.Method == http.MethodPut {
			res.WriteHeader(http.StatusOK)
			res.Write([]byte{})
		}
	}))
	defer httpReq.Close()
	authMock.On("GetHost").Return(httpReq.URL)

	err := testSubject.UpdateRealm(context.Background(), *realm)
	assert.Nil(t, err)
}

func TestUpdateRealm_FailToGetRealm(t *testing.T) {
	authMock, iamMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	realm := getMockRealms()[0]
	realmName := *realm.Name

	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)
	iamMock.On("GetRealm", context.Background(), *jwt, realmName).Return(nil, kcMocks.GetKeycloakError())

	err := testSubject.UpdateRealm(context.Background(), *realm)
	assert.NotNil(t, err)
}

func TestUpdateRealm_FailToUpdateRealm(t *testing.T) {
	authMock, iamMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	kcRealm := getMockKcRealms()[0]
	realm := getMockRealms()[0]
	realmName := *realm.Name

	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)
	iamMock.On("GetRealm", context.Background(), *jwt, realmName).Return(kcRealm, nil)

	realm.Enabled = gocloak.BoolP(false)

	expectedkcRealm := kcRealm
	expectedkcRealm.Enabled = gocloak.BoolP(false)

	iamMock.On("UpdateRealm", context.Background(), *jwt, *expectedkcRealm).Return(kcMocks.GetKeycloakError())

	err := testSubject.UpdateRealm(context.Background(), *realm)
	assert.NotNil(t, err)
}

func TestUpdateRealm_FailToGetAdminToken(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	authMock.On("GetAdminToken", context.Background()).Return(nil, kcMocks.GetKeycloakError())

	err := testSubject.UpdateRealm(context.Background(), models.Realm{})
	assert.NotNil(t, err)
}

func TestDeleteRealm_HappyFlow(t *testing.T) {
	authMock, iamMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	realm := getMockRealms()[0]
	realmName := *realm.Name

	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)

	iamMock.On("DeleteRealm", context.Background(), *jwt, realmName).Return(nil)

	err := testSubject.DeleteRealm(context.Background(), *realm.Name)
	assert.Nil(t, err)
}

func TestDeleteRealm_FailToDeleteRealm(t *testing.T) {
	authMock, iamMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	realm := getMockRealms()[0]
	realmName := *realm.Name

	authMock.On("GetAdminToken", context.Background()).Return(jwt, nil)

	iamMock.On("DeleteRealm", context.Background(), *jwt, realmName).Return(kcMocks.GetKeycloakError())

	err := testSubject.DeleteRealm(context.Background(), *realm.Name)
	assert.NotNil(t, err)
}

func TestDeleteRealm_FailToGetAdminToken(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)
	authMock.On("GetAdminToken", context.Background()).Return(nil, kcMocks.GetKeycloakError())

	err := testSubject.DeleteRealm(context.Background(), "test")
	assert.NotNil(t, err)
}

func TestShouldAddCustomAttribute_HappyFlow(t *testing.T) {
	name := "existingAttr"
	customAttribute := models.UserProfileAttributeMetadata{Name: &name}
	existingAttributesMap := map[string]struct{}{name: {}}
	assert.False(t, shouldAddCustomAttribute(customAttribute, existingAttributesMap))
}

func TestShouldAddCustomAttribute_NameIsNil(t *testing.T) {
	name := "existingAttr"
	customAttribute := models.UserProfileAttributeMetadata{Name: nil}
	existingAttributesMap := map[string]struct{}{name: {}}
	assert.False(t, shouldAddCustomAttribute(customAttribute, existingAttributesMap))
}

func TestShouldAddCustomAttribute_AddName(t *testing.T) {
	name := "newAttr"
	customAttribute := models.UserProfileAttributeMetadata{Name: &name}
	existingAttributesMap := map[string]struct{}{"existingAttr": {}}
	assert.True(t, shouldAddCustomAttribute(customAttribute, existingAttributesMap))
}

func TestGetAttributeNameMap_HappyFlow(t *testing.T) {
	name1 := "attribute1"
	name2 := "attribute2"
	attributes := []models.UserProfileAttributeMetadata{
		{Name: &name1},
		{Name: &name2},
	}
	result := getAttributeNameMap(attributes)
	assert.Len(t, result, 2)
	assert.Contains(t, result, name1)
	assert.Contains(t, result, name2)
}

func TestGetAttributeNameMap_DuplicateNames(t *testing.T) {
	name1 := "attribute1"
	name2 := "attribute2"
	attributes := []models.UserProfileAttributeMetadata{
		{Name: &name1},
		{Name: &name1},
		{Name: &name2},
	}
	result := getAttributeNameMap(attributes)
	assert.Len(t, result, 2)
	assert.Contains(t, result, name1)
	assert.Contains(t, result, name2)
}

func TestGetAttributeNameMap_EmptySlice(t *testing.T) {
	attributes := []models.UserProfileAttributeMetadata{}
	assert.Empty(t, getAttributeNameMap(attributes))
}

func TestEnrichUserProfileAttributes_HappyFlow(t *testing.T) {
	userProfile := []models.UserProfileAttributeMetadata{}
	result := enrichUserProfileAttributes(userProfile)
	assert.Len(t, result, 10)

	expectedAttributes := map[string]bool{
		utils.IsLocalUser:         true,
		utils.LastLoginTime:       true,
		utils.CreationTime:        true,
		utils.UpdateTime:          true,
		utils.Upn:                 true,
		utils.UID:                 true,
		utils.GID:                 true,
		utils.Groups:              true,
		utils.SupplementaryGroups: true,
		utils.CreatedBy:           true,
	}

	for _, attr := range result {
		assert.True(t, expectedAttributes[*attr.Name], "Unexpected attribute: %s", *attr.Name)
	}
}

func TestEnrichUserProfileAttributes_AddAttribute(t *testing.T) {
	name := "existingAttr"
	userProfile := []models.UserProfileAttributeMetadata{
		{Name: &name},
	}
	result := enrichUserProfileAttributes(userProfile)
	assert.Len(t, result, 11)
}

func TestGetAuthExecutions_HappyFlow(t *testing.T) {
	authMock, gocloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	flow := models.ModifyAuthenticationExecutionRepresentation{
		Alias:       gocloak.StringP("Alias"),
		Description: gocloak.StringP("Description"),
		ID:          gocloak.StringP("ID"),
		ProviderID:  gocloak.StringP("ProviderID"),
	}

	var authFlows []*gocloak.ModifyAuthenticationExecutionRepresentation
	authFlows = append(authFlows, helpers.FromModifyAutheExecutionRepModel(&flow))

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationExecutions", ctx, *jwt, realm, "Alias").Return(authFlows, nil)
	resp, err := testSubject.getAuthExecutions(ctx, realm, "Alias")
	assert.Nil(t, err)
	assert.NotNil(t, resp)
}

func Test_GetAuthFlowByProviderID_HappyFlow(t *testing.T) {
	authMock, gocloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	authExec := []*gocloak.ModifyAuthenticationExecutionRepresentation{
		{
			ProviderID: ptr.To(authFlowProviderID),
		},
		{
			ProviderID: ptr.To(authFlowProviderID + "1"),
		},
	}
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationExecutions", ctx, *jwt, realm, "browser").Return(authExec, nil)

	res, err := testSubject.getAuthFlowByProviderID(ctx, realm, authFlowProviderID)
	assert.Nil(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, authFlowProviderID, *res.ProviderID)
}

func Test_GetAuthFlowByProviderID_MissingAuthFlow(t *testing.T) {
	authMock, gocloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()
	ctx := context.Background()
	authExec := []*gocloak.ModifyAuthenticationExecutionRepresentation{
		{
			ProviderID: ptr.To(authFlowProviderID + "1"),
		},
		{
			ProviderID: ptr.To(authFlowProviderID + "2"),
		},
	}
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationExecutions", ctx, *jwt, realm, "browser").Return(authExec, nil)

	res, err := testSubject.getAuthFlowByProviderID(context.Background(), realm, authFlowProviderID)
	assert.NotNil(t, err)
	assert.Nil(t, res)
}

func Test_GetAuthFlowByProviderID_KC_Error(t *testing.T) {
	authMock, gocloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationExecutions", ctx, *jwt, realm, utils.BrowserFlowAlias).Return(nil, assert.AnError)

	resp, err := testSubject.getAuthFlowByProviderID(ctx, realm, utils.IdpProviderID)
	assert.NotNil(t, err)
	assert.Nil(t, resp)
}

func TestUpdateAutoRedirectSSO_EnableWithValidIdpAlias(t *testing.T) {
	authMock, gocloakMock, runaicloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	var authFlows []*gocloak.ModifyAuthenticationExecutionRepresentation
	authFlows = append(authFlows, helpers.FromModifyAutheExecutionRepModel(getMockAuthFlow()))

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationExecutions", ctx, *jwt, realm, utils.BrowserFlowAlias).Return(authFlows, nil)

	runaicloakMock.On("UpdateAuthFlowExecutionConfig", ctx, realm, *jwt, authFlowID, idpAlias).Return(nil)

	err := testSubject.UpdateAutoRedirectSSO(ctx, realm, ptr.To(idpAlias), true)
	assert.Nil(t, err)
}

func TestUpdateAutoRedirectSSO_DisableWithExistingConfig(t *testing.T) {
	authMock, gocloakMock, runaicloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	authFlow := &models.ModifyAuthenticationExecutionRepresentation{
		ID:                   gocloak.StringP(authFlowID),
		ProviderID:           gocloak.StringP(utils.IdpProviderID),
		AuthenticationConfig: gocloak.StringP("config-id"),
		DisplayName:          gocloak.StringP("Identity Provider Redirector"),
	}

	var authFlows []*gocloak.ModifyAuthenticationExecutionRepresentation
	authFlows = append(authFlows, helpers.FromModifyAutheExecutionRepModel(authFlow))

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationExecutions", ctx, *jwt, realm, utils.BrowserFlowAlias).Return(authFlows, nil)

	runaicloakMock.On("DeleteAuthFlowExecutionConfig", ctx, realm, *jwt, *authFlow.AuthenticationConfig).Return(nil)

	err := testSubject.UpdateAutoRedirectSSO(ctx, realm, nil, false)
	assert.Nil(t, err)
}

func TestUpdateAutoRedirectSSO_DisableWithNoConfig(t *testing.T) {
	authMock, gocloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	var authFlows []*gocloak.ModifyAuthenticationExecutionRepresentation
	authFlows = append(authFlows, helpers.FromModifyAutheExecutionRepModel(getMockAuthFlow()))

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationExecutions", ctx, *jwt, realm, utils.BrowserFlowAlias).Return(authFlows, nil)

	err := testSubject.UpdateAutoRedirectSSO(ctx, realm, nil, false)
	assert.Nil(t, err)
}

func TestUpdateAutoRedirectSSO_EnableWithoutIdpAlias(t *testing.T) {
	authMock, gocloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	var authFlows []*gocloak.ModifyAuthenticationExecutionRepresentation
	authFlows = append(authFlows, helpers.FromModifyAutheExecutionRepModel(getMockAuthFlow()))

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationExecutions", ctx, *jwt, realm, utils.BrowserFlowAlias).Return(authFlows, nil)

	err := testSubject.UpdateAutoRedirectSSO(ctx, realm, nil, true)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "invalid settings fields")
}

func TestUpdateAutoRedirectSSO_DisableWithIdpAlias(t *testing.T) {
	authMock, gocloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	var authFlows []*gocloak.ModifyAuthenticationExecutionRepresentation
	authFlows = append(authFlows, helpers.FromModifyAutheExecutionRepModel(getMockAuthFlow()))

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationExecutions", ctx, *jwt, realm, utils.BrowserFlowAlias).Return(authFlows, nil)

	err := testSubject.UpdateAutoRedirectSSO(ctx, realm, ptr.To(idpAlias), false)
	assert.Nil(t, err)
}

func TestUpdateAutoRedirectSSO_AuthFlowNotFound(t *testing.T) {
	authMock, gocloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	authFlow := &models.ModifyAuthenticationExecutionRepresentation{
		ID:          gocloak.StringP(authFlowID),
		ProviderID:  gocloak.StringP("different-provider"),
		DisplayName: gocloak.StringP("Different Provider"),
	}

	var authFlows []*gocloak.ModifyAuthenticationExecutionRepresentation
	authFlows = append(authFlows, helpers.FromModifyAutheExecutionRepModel(authFlow))

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationExecutions", ctx, *jwt, realm, utils.BrowserFlowAlias).Return(authFlows, nil)

	testIdpAlias := idpAlias
	err := testSubject.UpdateAutoRedirectSSO(ctx, realm, &testIdpAlias, true)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "couldn't find auth flow by provider id")
}

func TestUpdateAutoRedirectSSO_GetAuthExecutionsError(t *testing.T) {
	authMock, gocloakMock, _, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationExecutions", ctx, *jwt, realm, utils.BrowserFlowAlias).Return(nil, assert.AnError)

	testIdpAlias := idpAlias
	err := testSubject.UpdateAutoRedirectSSO(ctx, realm, &testIdpAlias, true)
	assert.NotNil(t, err)
}

func TestUpdateAutoRedirectSSO_GetAdminTokenError(t *testing.T) {
	authMock, _, _, testSubject := initTest(t)

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)

	testIdpAlias := idpAlias
	err := testSubject.UpdateAutoRedirectSSO(ctx, realm, &testIdpAlias, true)
	assert.NotNil(t, err)
}

func TestUpdateAutoRedirectSSO_UpdateConfigError(t *testing.T) {
	authMock, gocloakMock, runaicloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	var authFlows []*gocloak.ModifyAuthenticationExecutionRepresentation
	authFlows = append(authFlows, helpers.FromModifyAutheExecutionRepModel(getMockAuthFlow()))

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationExecutions", ctx, *jwt, realm, utils.BrowserFlowAlias).Return(authFlows, nil)

	runaicloakMock.On("UpdateAuthFlowExecutionConfig", ctx, realm, *jwt, authFlowID, idpAlias).Return(errors.New("error"))

	testIdpAlias := idpAlias
	err := testSubject.UpdateAutoRedirectSSO(ctx, realm, &testIdpAlias, true)
	assert.NotNil(t, err)
}

func TestUpdateAutoRedirectSSO_DeleteConfigError(t *testing.T) {
	authMock, gocloakMock, runaicloakMock, testSubject := initTest(t)
	jwt := kcMocks.GetMockAccessToken()

	authFlow := &models.ModifyAuthenticationExecutionRepresentation{
		ID:                   gocloak.StringP(authFlowID),
		ProviderID:           gocloak.StringP(utils.IdpProviderID),
		AuthenticationConfig: gocloak.StringP("config-id"),
		DisplayName:          gocloak.StringP("Identity Provider Redirector"),
	}

	var authFlows []*gocloak.ModifyAuthenticationExecutionRepresentation
	authFlows = append(authFlows, helpers.FromModifyAutheExecutionRepModel(authFlow))

	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(jwt, nil)
	gocloakMock.On("GetAuthenticationExecutions", ctx, *jwt, realm, utils.BrowserFlowAlias).Return(authFlows, nil)

	runaicloakMock.On("DeleteAuthFlowExecutionConfig", ctx, realm, *jwt, *authFlow.AuthenticationConfig).Return(errors.New("error"))

	err := testSubject.UpdateAutoRedirectSSO(ctx, realm, nil, false)
	assert.NotNil(t, err)
}

func getMockAuthFlow() *models.ModifyAuthenticationExecutionRepresentation {
	return &models.ModifyAuthenticationExecutionRepresentation{
		ID:          gocloak.StringP(authFlowID),
		ProviderID:  gocloak.StringP(utils.IdpProviderID),
		DisplayName: gocloak.StringP("Identity Provider Redirector"),
	}
}
