/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package idps

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"

	"github.com/Nerzal/gocloak/v13"
	authMocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/auth/mocks"
	iamError "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/errors"
	kcMocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/gocloak/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"
)

const (
	realm       string = "testRealm"
	metadataUrl string = "https://idp.example.com/metadata"
)

func initTest(t *testing.T) (authMock *authMocks.Auth, goCloakMock *kcMocks.GoCloak, testSubject *KeycloakIdpRepository) {
	goCloak := kcMocks.NewGoCloak(t)
	auth := authMocks.NewAuth(t)
	testSubject = &KeycloakIdpRepository{
		auth:     auth,
		kcClient: goCloak,
	}
	return auth, goCloak, testSubject
}

func getMockKcIdps() []*gocloak.IdentityProviderRepresentation {
	idps := []*gocloak.IdentityProviderRepresentation{
		{
			Alias:      gocloak.StringP("aliastest1"),
			ProviderID: gocloak.StringP("providertest1"),
			TrustEmail: gocloak.BoolP(true),
		},
		{
			Alias:      gocloak.StringP("aliastest2"),
			ProviderID: gocloak.StringP("providertest2"),
			TrustEmail: gocloak.BoolP(true),
		},
	}
	return idps
}

func getMockIdps() []*models.Idp {
	idps := []*models.Idp{
		{
			Alias:      gocloak.StringP("aliastest1"),
			ProviderID: gocloak.StringP("providertest1"),
		},
		{
			Alias:      gocloak.StringP("aliastest2"),
			ProviderID: gocloak.StringP("providertest2"),
		},
	}
	return idps
}

//#endregion helpers

func TestCreateIdp_HappyFlow(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	idp := getMockIdps()[0]
	kcIdp := getMockKcIdps()[0]
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("CreateIdentityProvider", ctx, *token, realm, *kcIdp).Return(*idp.Alias, nil)
	_, err := testSubject.CreateIdp(ctx, realm, *idp)
	assert.Nil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestCreateIdp_FailToGetToken(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	idp := getMockIdps()[0]
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	_, err := testSubject.CreateIdp(ctx, realm, *idp)
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestCreateIdp_FailToCreateIdp(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	idp := getMockIdps()[0]
	kcIdp := getMockKcIdps()[0]
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("CreateIdentityProvider", ctx, *token, realm, *kcIdp).Return("", assert.AnError)
	_, err := testSubject.CreateIdp(ctx, realm, *idp)
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetIdp_HappyFlow(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	idp := getMockIdps()[0]
	kcIdp := getMockKcIdps()[0]
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProvider", ctx, *token, realm, *idp.Alias).Return(kcIdp, nil)
	_, err := testSubject.GetIdp(ctx, realm, *idp.Alias)
	assert.Nil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetIdp_FailToGetToken(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	idp := getMockIdps()[0]
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	_, err := testSubject.GetIdp(ctx, realm, *idp.Alias)
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetIdp_FailToGetIdp(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	idp := getMockIdps()[0]
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProvider", ctx, *token, realm, *idp.Alias).Return(nil, assert.AnError)
	_, err := testSubject.GetIdp(ctx, realm, *idp.Alias)
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestUpdateIdp_HappyFlow(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	idp := getMockIdps()[0]
	kcIdp := getMockKcIdps()[0]
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProvider", ctx, *token, realm, *idp.Alias).Return(kcIdp, nil)
	goCloakMock.On("UpdateIdentityProvider", ctx, *token, realm, *idp.Alias, *kcIdp).Return(nil)
	err := testSubject.UpdateIdp(ctx, realm, *idp)
	assert.Nil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestUpdateIdp_FailToGetToken(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	idp := getMockIdps()[0]
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	err := testSubject.UpdateIdp(ctx, realm, *idp)
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestUpdateIdp_FailToGetIdp(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	idp := getMockIdps()[0]
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProvider", ctx, *token, realm, *idp.Alias).Return(nil, assert.AnError)
	err := testSubject.UpdateIdp(ctx, realm, *idp)
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestUpdateIdp_FailToUpdateIdp(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	idp := getMockIdps()[0]
	kcIdp := getMockKcIdps()[0]
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProvider", ctx, *token, realm, *idp.Alias).Return(kcIdp, nil)
	goCloakMock.On("UpdateIdentityProvider", ctx, *token, realm, *idp.Alias, *kcIdp).Return(assert.AnError)
	err := testSubject.UpdateIdp(ctx, realm, *idp)
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestDeleteIdp_HappyFlow(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	idp := getMockIdps()[0]
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("DeleteIdentityProvider", ctx, *token, realm, *idp.Alias).Return(nil)
	err := testSubject.DeleteIdp(ctx, realm, *idp.Alias)
	assert.Nil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestDeleteIdp_FailToGetToken(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	idp := getMockIdps()[0]
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	err := testSubject.DeleteIdp(ctx, realm, *idp.Alias)
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestDeleteIdp_FailToDeleteIdp(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	idp := getMockIdps()[0]
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("DeleteIdentityProvider", ctx, *token, realm, *idp.Alias).Return(assert.AnError)
	err := testSubject.DeleteIdp(ctx, realm, *idp.Alias)
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetIdps_HappyFlow(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProviders", ctx, *token, realm).Return(getMockKcIdps(), nil)
	_, err := testSubject.GetIdps(ctx, realm)
	assert.Nil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetIdps_FailToGetToken(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	_, err := testSubject.GetIdps(ctx, realm)
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetIdps_FailToGetIdps(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProviders", ctx, *token, realm).Return(nil, assert.AnError)
	_, err := testSubject.GetIdps(ctx, realm)
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetIdps_EmptyIdps(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProviders", ctx, *token, realm).Return([]*gocloak.IdentityProviderRepresentation{}, nil)
	idpsResponse, err := testSubject.GetIdps(ctx, realm)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(idpsResponse))
	assert.Nil(t, idpsResponse)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestIsIdpConfigured_HappyFlow(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProviders", ctx, *token, realm).Return(getMockKcIdps(), nil)
	configured, err := testSubject.IsIdpConfigured(ctx, realm)
	assert.Nil(t, err)
	assert.True(t, configured)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestIsIdpConfigured_FailToGetToken(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	configured, err := testSubject.IsIdpConfigured(ctx, realm)
	assert.NotNil(t, err)
	assert.False(t, configured)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestIsIdpConfigured_EmptyIdps(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProviders", ctx, *token, realm).Return([]*gocloak.IdentityProviderRepresentation{}, nil)
	configured, err := testSubject.IsIdpConfigured(ctx, realm)
	assert.Nil(t, err)
	assert.False(t, configured)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestIsIdpConfigured_FailToGetIdps(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProviders", ctx, *token, realm).Return(nil, assert.AnError)
	configured, err := testSubject.IsIdpConfigured(ctx, realm)
	assert.NotNil(t, err)
	assert.False(t, configured)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetSPSamlMetadataUrl(t *testing.T) {
	_, _, testSubject := initTest(t)
	// The new implementation uses config.TenantsRealm().AppHost
	// We need to set up the config for testing
	config.InitConfig()
	originalConfig := runai_config.Get[*config.TenantRealm]()
	defer func() {
		runai_config.Set(originalConfig) // Restore original config
	}()

	testAppHost := "https://test-domain.com"
	assert.NoError(t, runai_config.Set(&config.TenantRealm{
		AppHost: testAppHost,
	}))

	expectedURL := fmt.Sprintf("%s/auth/realms/%s/broker/%s/endpoint/descriptor", testAppHost, realm, "saml-a1b2c3")

	result := testSubject.GetSPSamlMetadataUrl(context.Background(), realm, "saml-a1b2c3")

	assert.Equal(t, expectedURL, result)
}

func TestSetMandatotyClaimMapper_Create_HappyFlow(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	idp := getMockIdps()[0]
	idpMapper := &models.IdpMapper{Name: ptr.To("name")}
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProviderMappers", ctx, *token, realm, *idp.Alias).Return(nil, nil)
	goCloakMock.On("CreateIdentityProviderMapper", ctx, *token, realm, *idp.Alias, mock.Anything).Return("", nil)
	err := testSubject.SetMandatotyClaimMapper(ctx, realm, *idp.Alias, idpMapper)
	assert.Nil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestSetMandatotyClaimMapper_Update_HappyFlow(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	idp := getMockIdps()[0]
	idpMapper := &models.IdpMapper{Name: ptr.To("Name")}
	identityProviderMappers := []*gocloak.IdentityProviderMapper{
		{
			ID:   ptr.To("ID"),
			Name: ptr.To("Name"),
		},
	}
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProviderMappers", ctx, *token, realm, *idp.Alias).Return(identityProviderMappers, nil)
	goCloakMock.On("UpdateIdentityProviderMapper", ctx, *token, realm, *idp.Alias, mock.Anything).Return(nil)
	err := testSubject.SetMandatotyClaimMapper(ctx, realm, *idp.Alias, idpMapper)
	assert.Nil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestSetMandatotyClaimMapper_FailToGetToken(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	idp := getMockIdps()[0]
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	err := testSubject.SetMandatotyClaimMapper(ctx, realm, *idp.Alias, &models.IdpMapper{})
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestSetMandatotyClaimMapper_FailToCreateMapper(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	idp := getMockIdps()[0]
	idpMapper := &models.IdpMapper{Name: ptr.To("Name")}
	identityProviderMappers := []*gocloak.IdentityProviderMapper{
		{
			ID:   ptr.To("ID"),
			Name: ptr.To("Name1"),
		},
	}
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProviderMappers", ctx, *token, realm, *idp.Alias).Return(identityProviderMappers, nil)
	goCloakMock.On("CreateIdentityProviderMapper", ctx, *token, realm, *idp.Alias, mock.Anything).Return("", fmt.Errorf("error"))
	err := testSubject.SetMandatotyClaimMapper(ctx, realm, *idp.Alias, idpMapper)
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestDeleteMandatoryClaimMapper_HappyFlow(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	idp := getMockIdps()[0]
	mapperName := "test-mapper"
	identityProviderMappers := []*gocloak.IdentityProviderMapper{
		{
			ID:   ptr.To("ID"),
			Name: ptr.To(mapperName),
		},
	}
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProviderMappers", ctx, *token, realm, *idp.Alias).Return(identityProviderMappers, nil)
	goCloakMock.On("DeleteIdentityProviderMapper", ctx, *token, realm, *idp.Alias, "ID").Return(nil)
	err := testSubject.DeleteMandatoryClaimMapper(ctx, realm, *idp.Alias, mapperName)
	assert.Nil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestDeleteMandatoryClaimMapper_FailToGetToken(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	idp := getMockIdps()[0]
	mapperName := "test-mapper"
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	err := testSubject.DeleteMandatoryClaimMapper(ctx, realm, *idp.Alias, mapperName)
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestDeleteMandatoryClaimMapper_FailToDeleteMapper(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	idp := getMockIdps()[0]
	mapperName := "test-mapper"
	identityProviderMappers := []*gocloak.IdentityProviderMapper{
		{
			ID:   ptr.To("ID"),
			Name: ptr.To(mapperName),
		},
	}
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProviderMappers", ctx, *token, realm, *idp.Alias).Return(identityProviderMappers, nil)
	goCloakMock.On("DeleteIdentityProviderMapper", ctx, *token, realm, *idp.Alias, "ID").Return(assert.AnError)
	err := testSubject.DeleteMandatoryClaimMapper(ctx, realm, *idp.Alias, mapperName)
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetIdpConfig_HappyFlow(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	kcConfig := map[string]string{"entityId": "https://idp.example.com"}
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("ImportIdentityProviderConfig", ctx, *token, realm, metadataUrl, "saml").Return(kcConfig, nil)
	config, err := testSubject.GetIdpConfig(ctx, realm, metadataUrl, "saml")
	assert.Nil(t, err)
	assert.Equal(t, kcConfig, config)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetIdpConfig_MetadataRejectedByKeycloak(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("ImportIdentityProviderConfig", ctx, *token, realm, metadataUrl, "saml").
		Return(nil, getKcApiError(http.StatusBadRequest))
	_, err := testSubject.GetIdpConfig(ctx, realm, metadataUrl, "saml")
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, iamError.ErrInvalidInput))
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetIdpConfig_FailToGetToken(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	_, err := testSubject.GetIdpConfig(ctx, realm, metadataUrl, "saml")
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetIdpConfigFromFile_HappyFlow(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	kcConfig := map[string]string{"entityId": "https://idp.example.com"}
	reader := strings.NewReader("<EntityDescriptor/>")
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("ImportIdentityProviderConfigFromFile", ctx, *token, realm, "saml", "metadata.xml", reader).Return(kcConfig, nil)
	config, err := testSubject.GetIdpConfigFromFile(ctx, realm, "saml", "metadata.xml", reader)
	assert.Nil(t, err)
	assert.Equal(t, kcConfig, config)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetIdpConfigFromFile_MetadataRejectedByKeycloak(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	reader := strings.NewReader("not xml")
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("ImportIdentityProviderConfigFromFile", ctx, *token, realm, "saml", "metadata.xml", reader).
		Return(nil, getKcApiError(http.StatusBadRequest))
	_, err := testSubject.GetIdpConfigFromFile(ctx, realm, "saml", "metadata.xml", reader)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, iamError.ErrInvalidInput))
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetIdpConfigFromFile_FailToGetToken(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	authMock.On("GetAdminToken", ctx).Return(nil, assert.AnError)
	_, err := testSubject.GetIdpConfigFromFile(ctx, realm, "saml", "metadata.xml", strings.NewReader("not xml"))
	assert.NotNil(t, err)
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestGetIdpConfig_KeycloakServerError(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("ImportIdentityProviderConfig", ctx, *token, realm, metadataUrl, "saml").
		Return(nil, getKcApiError(http.StatusInternalServerError))
	_, err := testSubject.GetIdpConfig(ctx, realm, metadataUrl, "saml")
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, iamError.ErrInternal))
	assert.False(t, errors.Is(err, iamError.ErrInvalidInput))
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func TestUpdateIdp_IdpNotFound(t *testing.T) {
	authMock, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	token := kcMocks.GetMockAccessToken()
	idp := getMockIdps()[0]
	authMock.On("GetAdminToken", ctx).Return(token, nil)
	goCloakMock.On("GetIdentityProvider", ctx, *token, realm, *idp.Alias).Return(nil, getKcApiError(http.StatusNotFound))
	err := testSubject.UpdateIdp(ctx, realm, *idp)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, iamError.ErrNotFound))
	authMock.AssertExpectations(t)
	goCloakMock.AssertExpectations(t)
}

func getKcApiError(code int) error {
	return &gocloak.APIError{Code: code, Message: fmt.Sprintf("%d %s", code, http.StatusText(code))}
}
