/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	"github.com/Nerzal/gocloak/v13"

	internalCache "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/cache"
	internalCacheMocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/cache/mocks"
	iamError "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/errors"
	iamMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/mocks"
	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	authMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business/authorization/mocks"
	businessMocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/realm/identity_provider"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/realm/providers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
)

const (
	URL              = "https://example.com/"
	NON_SAML_XML     = "hello world"
	INVALID_XML_FILE = "not!valid!base64"
	XML_URL          = "https://example.com/auth/realms/workAI/protocol/saml/descriptor"
	realm            = "test-realm"
	idpAlias         = "idpAlias"
)

// The upload path parses the metadata before handing it to Keycloak, so the
// happy-flow fixture has to be a real SAML descriptor.
const SAML_XML = `<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" entityID="https://idp.example.com"/>`

var (
	XML_FILE          = base64.StdEncoding.EncodeToString([]byte(SAML_XML))
	NON_SAML_XML_FILE = base64.StdEncoding.EncodeToString([]byte(NON_SAML_XML))
)

type IdpsApiServiceMocks struct {
	IamServicesContainerMock *iamMock.MockServicesContainer
	IdpMock                  *businessMocks.IdpDal
	AuthManagerMock          *authMock.AuthManagerInterface
	IdpCacheMock             *internalCacheMocks.IdpCache
	IdpMappersCacheMock      *internalCacheMocks.IdpMappersCache
	IdpsListCacheMock        *internalCacheMocks.IdpsListCache
}

func initIdpServiceTests(t *testing.T) (
	MocksList *IdpsApiServiceMocks,
	testSubject *IdpsApiService,
	ctx context.Context,
) {
	config.InitConfig()

	mockIamServicesContainer := iamMock.NewMockServicesContainer()
	idpMock := businessMocks.NewIdpDal(t)
	authManagerMock := authMock.NewAuthManagerInterface(t)
	idpCacheMock := internalCacheMocks.NewIdpCache(t)
	idpMappersCacheMock := internalCacheMocks.NewIdpMappersCache(t)
	idpsListCacheMock := internalCacheMocks.NewIdpsListCache(t)
	MocksList = &IdpsApiServiceMocks{
		IamServicesContainerMock: mockIamServicesContainer,
		IdpMock:                  idpMock,
		AuthManagerMock:          authManagerMock,
		IdpCacheMock:             idpCacheMock,
		IdpMappersCacheMock:      idpMappersCacheMock,
		IdpsListCacheMock:        idpsListCacheMock,
	}
	testSubject = &IdpsApiService{
		idpRepresentationFactory: identity_provider.NewIdpRepresentationFactory(),
		dal:                      idpMock,
		iamIdps:                  MocksList.IamServicesContainerMock.ToServicesContainer().IdpService,
		iamClients:               MocksList.IamServicesContainerMock.ToServicesContainer().ClientService,
		authManager:              MocksList.AuthManagerMock,
		idpsCacheWrapper: &internalCache.IdpCacheWrapper{
			IdpCache:        MocksList.IdpCacheMock,
			IdpMappersCache: MocksList.IdpMappersCacheMock,
			IdpsListCache:   MocksList.IdpsListCacheMock,
		},
	}
	ctx = context.WithValue(context.Background(), utils.AuthRealmKey, realm)
	ctx = context.WithValue(ctx, runaictx.CfTenantId, "tenant-id")
	return MocksList, testSubject, ctx
}

func TestCreateIdp_NoData(t *testing.T) {
	assert.Nil(t, nil)
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	authMock := MocksList.AuthManagerMock
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)

	request := api.IdpCreationRequest{
		Name: "test",
	}
	response, err := testSubject.CreateIdp(ctx, request, "")
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, errors.New(fmt.Sprintf("unable to determin idp configuration")), response.Body)
}

func TestCreateIdp_WrongData(t *testing.T) {
	assert.Nil(t, nil)
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	authMock := MocksList.AuthManagerMock
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)

	request := api.IdpCreationRequest{
		Type:     api.IdpCreationRequestTypeEnum(models.OCP),
		Name:     idpAlias,
		SamlData: &api.SamlCreationData{MetadataXmlUrl: "test"},
		OidcData: &api.OidcCreationData{},
		OcpData:  nil,
	}
	response, err := testSubject.CreateIdp(ctx, request, "")
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, errors.New(fmt.Sprintf("unable to determin idp configuration")), response.Body)
}

func TestCreateIdp_Ocp409Flow(t *testing.T) {
	assert.Nil(t, nil)
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	authMock := MocksList.AuthManagerMock
	iamMock := MocksList.IamServicesContainerMock
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("CreateIdp", ctx, realm, mock.Anything).Return(ptr.To(""), errors.New("409 Conflict"))

	request := api.IdpCreationRequest{
		Type:     api.IdpCreationRequestTypeEnum(models.OCP),
		Name:     idpAlias,
		SamlData: nil,
		OidcData: nil,
		OcpData: &api.OcpCreationData{
			ClientId:     "a",
			IdpBaseUrl:   URL,
			ClientSecret: "c",
		},
	}
	response, err := testSubject.CreateIdp(ctx, request, "")
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusConflict, response.Code)
	assert.Equal(t, errors.New("failed to create idp with alias idpAlias for tenant test-realm. err: name already exists"), response.Body)
}

func TestCreateIdp_Ocp403Flow(t *testing.T) {
	assert.Nil(t, nil)
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	authMock := MocksList.AuthManagerMock
	iamMock := MocksList.IamServicesContainerMock
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("CreateIdp", ctx, realm, mock.Anything).Return(ptr.To(""), errors.New("not 409"))

	request := api.IdpCreationRequest{
		Type:     api.IdpCreationRequestTypeEnum(models.OCP),
		Name:     idpAlias,
		SamlData: nil,
		OidcData: nil,
		OcpData: &api.OcpCreationData{
			ClientId:     "a",
			IdpBaseUrl:   URL,
			ClientSecret: "c",
		},
	}
	response, err := testSubject.CreateIdp(ctx, request, "")
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateIdp_OcpHappyFlow(t *testing.T) {
	assert.Nil(t, nil)
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	authMock := MocksList.AuthManagerMock
	iamMock := MocksList.IamServicesContainerMock
	idpCacheMock := MocksList.IdpCacheMock
	idpsListCacheMock := MocksList.IdpsListCacheMock
	idpMappersCacheMock := MocksList.IdpMappersCacheMock

	iamMock.MockIdpService.On("CreateIdp", ctx, realm, mock.Anything).Return(ptr.To(idpAlias), nil)
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("UpdateIdpMappers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	idpCacheMock.On("DeleteIdpFromRealm", ctx, realm, idpAlias, mock.Anything).Return(nil)
	idpsListCacheMock.On("DeleteIdpsListFromRealm", ctx, realm).Return(nil)
	idpMappersCacheMock.On("DeleteIdpMappersFromRealm", ctx, realm, idpAlias).Return(nil)
	request := api.IdpCreationRequest{
		Type:     api.IdpCreationRequestTypeEnum(models.OCP),
		Name:     idpAlias,
		SamlData: nil,
		OidcData: nil,
		OcpData: &api.OcpCreationData{
			ClientId:     "a",
			IdpBaseUrl:   URL,
			ClientSecret: "c",
		},
	}
	response, err := testSubject.CreateIdp(ctx, request, "")

	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	res := response.Body.(api.IdpPostResponse)
	assert.Equal(t, idpAlias, res.Alias)
}

func TestCreateIdp_OidcHappyFlow(t *testing.T) {
	assert.Nil(t, nil)
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	idpCacheMock := MocksList.IdpCacheMock
	idpsListCacheMock := MocksList.IdpsListCacheMock
	idpMappersCacheMock := MocksList.IdpMappersCacheMock
	dalMock := MocksList.IdpMock
	wg := sync.WaitGroup{}
	wg.Add(1)
	config := map[string]string{"clientSecret": "test"}
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("GetIdpConfig", mock.Anything, mock.Anything, mock.Anything, "oidc").Return(config, nil)
	iamMock.MockIdpService.On("CreateIdp", ctx, realm, mock.Anything).Return(ptr.To(idpAlias), nil)
	iamMock.MockClientService.On("CreateExchangePolicy", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iamMock.MockIdpService.On("UpdateIdpMappers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	idpCacheMock.On("DeleteIdpFromRealm", ctx, realm, idpAlias, mock.Anything).Return(nil)
	idpsListCacheMock.On("DeleteIdpsListFromRealm", ctx, realm).Return(nil)
	idpMappersCacheMock.On("DeleteIdpMappersFromRealm", ctx, realm, idpAlias).Return(nil)
	dalMock.On("GetIdpMetadata", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	dalMock.On("SaveIdpMetadata", mock.Anything, mock.Anything).Run(func(mock.Arguments) {
		defer wg.Done()
	}).Return(nil)

	request := api.IdpCreationRequest{
		Type:     api.IdpCreationRequestTypeEnum(models.OIDC),
		Name:     idpAlias,
		SamlData: nil,
		OidcData: &api.OidcCreationData{
			ClientId:            "a",
			ClientSecret:        "a",
			DiscoverDocumentUrl: URL,
		},
		OcpData: nil,
	}
	response, err := testSubject.CreateIdp(ctx, request, "")
	wg.Wait()

	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	res := response.Body.(api.IdpPostResponse)
	assert.Equal(t, idpAlias, res.Alias)
}

func TestCreateIdp_OidcMissingFlow(t *testing.T) {
	assert.Nil(t, nil)
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	authMock := MocksList.AuthManagerMock
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)

	request := api.IdpCreationRequest{
		Type:     api.IdpCreationRequestTypeEnum(models.OIDC),
		Name:     idpAlias,
		SamlData: nil,
		OidcData: &api.OidcCreationData{
			ClientSecret:        "a",
			DiscoverDocumentUrl: URL,
		},
		OcpData: nil,
	}
	response, err := testSubject.CreateIdp(ctx, request, "")
	assert.NotNil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, errors.New(fmt.Sprintf("invalid oidc configuration")), response.Body)
}

func TestCreateIdp_SamlHappyFlow(t *testing.T) {
	assert.Nil(t, nil)
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	idpCacheMock := MocksList.IdpCacheMock
	idpsListCacheMock := MocksList.IdpsListCacheMock
	idpMappersCacheMock := MocksList.IdpMappersCacheMock
	dalMock := MocksList.IdpMock
	wg := sync.WaitGroup{}
	wg.Add(1)
	config := signedSamlConfig(t)
	config["clientSecret"] = "test"
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("GetIdpConfig", mock.Anything, mock.Anything, mock.Anything, "saml").Return(config, nil)
	iamMock.MockIdpService.On("CreateIdp", ctx, realm, mock.Anything).Return(ptr.To(idpAlias), nil)
	iamMock.MockIdpService.On("UpdateIdpMappers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	idpCacheMock.On("DeleteIdpFromRealm", ctx, realm, idpAlias, mock.Anything).Return(nil)
	idpsListCacheMock.On("DeleteIdpsListFromRealm", ctx, realm).Return(nil)
	idpMappersCacheMock.On("DeleteIdpMappersFromRealm", ctx, realm, idpAlias).Return(nil)
	dalMock.On("GetIdpMetadata", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	dalMock.On("SaveIdpMetadata", mock.Anything, mock.Anything).Run(func(mock.Arguments) {
		defer wg.Done()
	}).Return(nil)
	request := api.IdpCreationRequest{
		Type:     api.IdpCreationRequestTypeEnum(models.SAML),
		Name:     idpAlias,
		SamlData: &api.SamlCreationData{MetadataXmlUrl: URL, MetadataXmlType: utils.Url},
		OidcData: nil,
		OcpData:  nil,
	}
	response, err := testSubject.CreateIdp(ctx, request, "")
	wg.Wait()
	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.Code)
	res := response.Body.(api.IdpPostResponse)
	assert.Equal(t, idpAlias, res.Alias)
}

func TestCreateIdp_ConfigUrlError(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("GetIdpConfig", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))

	request := api.IdpCreationRequest{
		Type:     api.IdpCreationRequestTypeEnum(models.SAML),
		Name:     idpAlias,
		SamlData: &api.SamlCreationData{MetadataXmlUrl: URL, MetadataXmlType: utils.Url},
	}
	response, err := testSubject.CreateIdp(ctx, request, "")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateIdp_ConfigFileError(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("GetIdpConfigFromFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))

	response, err := testSubject.CreateIdp(ctx, getSAMLIdpCreationRequest(utils.File, XML_FILE), "")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestCreateIdp_InvalidBase64MetadataFile(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	authMock := MocksList.AuthManagerMock
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)

	response, err := testSubject.CreateIdp(ctx, getSAMLIdpCreationRequest(utils.File, INVALID_XML_FILE), "")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, http.StatusBadRequest, err.(utils.Error).Code)
}

func TestCreateIdp_NonSamlMetadataFile(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	authMock := MocksList.AuthManagerMock
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)

	response, err := testSubject.CreateIdp(ctx, getSAMLIdpCreationRequest(utils.File, NON_SAML_XML_FILE), "")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, http.StatusBadRequest, err.(utils.Error).Code)
}

func TestCreateIdp_MetadataRejectedByKeycloak(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("GetIdpConfigFromFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, iamError.HandleKeyCloakError(&gocloak.APIError{Code: http.StatusBadRequest, Message: "400 Bad Request"}))

	response, err := testSubject.CreateIdp(ctx, getSAMLIdpCreationRequest(utils.File, XML_FILE), "")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, http.StatusBadRequest, err.(utils.Error).Code)
}

func TestCreateIdp_KeycloakServerError(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("GetIdpConfigFromFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, iamError.HandleKeyCloakError(&gocloak.APIError{Code: http.StatusInternalServerError, Message: "500 Internal Server Error"}))

	response, err := testSubject.CreateIdp(ctx, getSAMLIdpCreationRequest(utils.File, XML_FILE), "")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestDeleteIdp_HappyFlow(t *testing.T) {
	assert.Nil(t, nil)
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	idpMock := MocksList.IdpMock
	idpCacheMock := MocksList.IdpCacheMock
	idpsListCacheMock := MocksList.IdpsListCacheMock
	idpMappersCacheMock := MocksList.IdpMappersCacheMock
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("DeleteIdp", mock.Anything, mock.Anything, idpAlias).Return(nil)
	idpMock.On("DeleteIdpMetadata", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	idpCacheMock.On("DeleteIdpFromRealm", ctx, realm, idpAlias).Return(nil)
	idpsListCacheMock.On("DeleteIdpsListFromRealm", ctx, realm).Return(nil)
	idpMappersCacheMock.On("DeleteIdpMappersFromRealm", ctx, realm, idpAlias).Return(nil)
	response, err := testSubject.DeleteIdp(ctx, idpAlias, "")

	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Nil(t, response.Body)
}

func TestDeleteIdp_FailedReconcile(t *testing.T) {
	assert.Nil(t, nil)
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	idpMock := MocksList.IdpMock
	idpCacheMock := MocksList.IdpCacheMock
	idpsListCacheMock := MocksList.IdpsListCacheMock
	idpMappersCacheMock := MocksList.IdpMappersCacheMock
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("DeleteIdp", mock.Anything, mock.Anything, idpAlias).Return(nil)
	idpMock.On("DeleteIdpMetadata", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	idpCacheMock.On("DeleteIdpFromRealm", ctx, realm, idpAlias).Return(nil)
	idpsListCacheMock.On("DeleteIdpsListFromRealm", ctx, realm).Return(nil)
	idpMappersCacheMock.On("DeleteIdpMappersFromRealm", ctx, realm, idpAlias).Return(nil)
	response, err := testSubject.DeleteIdp(ctx, idpAlias, "")

	assert.Nil(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusNoContent, response.Code)
	assert.Nil(t, response.Body)
}

func TestDeleteIdp_Error(t *testing.T) {
	assert.Nil(t, nil)
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	idpCacheMock := MocksList.IdpCacheMock
	idpsListCacheMock := MocksList.IdpsListCacheMock
	idpMappersCacheMock := MocksList.IdpMappersCacheMock
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("DeleteIdp", mock.Anything, mock.Anything, idpAlias).Return(assert.AnError)
	idpCacheMock.On("DeleteIdpFromRealm", ctx, realm, idpAlias).Return(nil)
	idpsListCacheMock.On("DeleteIdpsListFromRealm", ctx, realm).Return(nil)
	idpMappersCacheMock.On("DeleteIdpMappersFromRealm", ctx, realm, idpAlias).Return(nil)
	response, err := testSubject.DeleteIdp(ctx, idpAlias, "")
	assert.NotNil(t, err)
	assert.NotNil(t, response)
}

func TestUpdateIdpMappers_HappyFlow(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	idpCacheMock := MocksList.IdpCacheMock
	idpMappersCacheMock := MocksList.IdpMappersCacheMock
	mappers := generateNewMappers()
	mappersConfig := generateMapperConfigArray()
	idpRep := generateIdentityProviderRepresentation()
	idpRep.Enabled = ptr.To(true)

	idpCacheMock.On("GetIdpForRealm", ctx, realm, idpAlias, mock.Anything).Return(idpRep, nil)
	idpMappersCacheMock.On("GetIdpMappersForRealm", ctx, realm, idpAlias, mock.Anything).Return(&internalCache.IdpMappers{Mappers: mappersConfig}, nil)
	iamMock.MockIdpService.On("UpdateIdpMappers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := testSubject.updateIdpMappers(ctx, realm, idpAlias, mappers)
	assert.Nil(t, err)
}

func TestUpdateIdpMappers_Error(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	idpCacheMock := MocksList.IdpCacheMock
	idpMappersCacheMock := MocksList.IdpMappersCacheMock
	mappers := generateNewMappers()
	mappersConfig := generateMapperConfigArray()
	idpRep := generateIdentityProviderRepresentation()
	idpRep.Enabled = ptr.To(true)

	idpCacheMock.On("GetIdpForRealm", ctx, realm, idpAlias, mock.Anything).Return(idpRep, nil)
	idpMappersCacheMock.On("GetIdpMappersForRealm", ctx, realm, idpAlias, mock.Anything).Return(&internalCache.IdpMappers{Mappers: mappersConfig}, nil)
	iamMock.MockIdpService.On("UpdateIdpMappers", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(assert.AnError)
	err := testSubject.updateIdpMappers(ctx, realm, idpAlias, mappers)
	assert.NotNil(t, err)
}

func TestGetMappersForUpdate_HappyFlow(t *testing.T) {
	_, testSubject, ctx := initIdpServiceTests(t)
	mappers := generateNewMappers()
	mappersConfig := generateMapperConfigArray()
	idpRep := generateIdentityProviderRepresentation()
	mappperToUpdate := testSubject.getMappersForUpdate(ctx, mappers, mappersConfig, idpRep)
	assert.NotNil(t, mappperToUpdate)
	assert.NotNil(t, 1, len(mappperToUpdate))
}

func TestGetEmptyMappersForUpdate_HappyFlow(t *testing.T) {
	_, testSubject, ctx := initIdpServiceTests(t)
	mappers := generateEmptyNewMappers()
	mappersConfig := generateMapperConfigArray()
	idpRep := generateIdentityProviderRepresentation()
	mapperToUpdate := testSubject.getMappersForUpdate(ctx, mappers, mappersConfig, idpRep)
	mapConfig := *mapperToUpdate[0].Config
	assert.NotNil(t, mapConfig["user.attribute"])
}

func TestGetIdpConfig_SAML_File(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	conf := &map[string]string{
		utils.MetadataXmlUrl:  "",
		utils.MetadataXmlFile: XML_FILE,
		utils.EntityId:        "entityId",
		utils.FileName:        "fileName",
		utils.MetadataXmlType: utils.File,
	}
	idpData := getIdpMetadataModel(idpAlias, XML_FILE, utils.File, models.SAML, map[string]string{}, conf)
	iamMock := MocksList.IamServicesContainerMock
	iamMock.MockIdpService.On("GetIdpConfigFromFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(signedSamlConfig(t), nil)
	config, err := testSubject.getIdpConfig(ctx, "runai", &idpData)
	assert.Nil(t, err)
	assert.Equal(t, "true", config[utils.ValidateSignature])
	assert.NoError(t, utils.ValidateSigningCertificate(config[utils.SigningCertificate]))
}

func TestGetIdpConfig_SAML_Url(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	conf := &map[string]string{
		utils.MetadataXmlUrl:  "metadataXmlUrl",
		utils.MetadataXmlFile: "",
		utils.EntityId:        "entityId",
		utils.FileName:        "",
		utils.MetadataXmlType: utils.Url,
	}
	idpData := getIdpMetadataModel(idpAlias, "fromUrl", utils.Url, models.SAML, map[string]string{}, conf)
	iamMock := MocksList.IamServicesContainerMock
	iamMock.MockIdpService.On("GetIdpConfig", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(signedSamlConfig(t), nil)
	config, err := testSubject.getIdpConfig(ctx, "runai", &idpData)
	assert.Nil(t, err)
	assert.Equal(t, "true", config[utils.ValidateSignature])
	assert.NoError(t, utils.ValidateSigningCertificate(config[utils.SigningCertificate]))
}

func TestGetIdpConfig_OIDC_Url(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	idpData := getIdpMetadataModel(idpAlias, "fromUrl", utils.Url, models.OIDC, map[string]string{}, &map[string]string{})
	iamMock := MocksList.IamServicesContainerMock
	iamMock.MockIdpService.On("GetIdpConfig", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]string{}, nil)
	config, err := testSubject.getIdpConfig(ctx, "runai", &idpData)
	assert.Nil(t, err)
	assert.Equal(t, map[string]string{}, config)
}

func TestGetIdpConfig_OCP(t *testing.T) {
	_, testSubject, ctx := initIdpServiceTests(t)
	idpData := getIdpMetadataModel(idpAlias, "", "", models.OCP, map[string]string{}, &map[string]string{})
	config, err := testSubject.getIdpConfig(ctx, "runai", &idpData)
	assert.Nil(t, err)
	assert.Equal(t, map[string]string{}, config)
}

func TestGetConfigFromFile_HappyFlow(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	iamMock.MockIdpService.On("GetIdpConfigFromFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]string{}, nil)
	config, err := testSubject.getConfigFromFile(ctx, realm, strings.NewReader("Hello, testing!"), "fileName", "saml")
	assert.Nil(t, err)
	assert.Equal(t, map[string]string{}, config)
}

func TestGetConfigFromFile_error(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	iamMock.MockIdpService.On("GetIdpConfigFromFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)
	config, err := testSubject.getConfigFromFile(ctx, realm, strings.NewReader("Hello, testing!"), "fileName", "saml")
	assert.NotNil(t, err)
	assert.Nil(t, config)
}

func TestGetConfigFromUrl_HappyFlow(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	iamMock.MockIdpService.On("GetIdpConfig", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(map[string]string{}, nil)
	config, err := testSubject.getConfigFromUrl(ctx, realm, "url", "saml")
	assert.Nil(t, err)
	assert.Equal(t, map[string]string{}, config)
}

func TestGetConfigFromUrl_error(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	iamMock.MockIdpService.On("GetIdpConfig", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)
	config, err := testSubject.getConfigFromUrl(ctx, realm, "url", "saml")
	assert.NotNil(t, err)
	assert.Nil(t, config)
}

func TestGetIdpResources_HappyFlow(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	authMock := MocksList.AuthManagerMock
	idpMock := MocksList.IdpMock
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)
	idpMock.On("GetIdpMetadata", mock.Anything, mock.Anything).Return(&models.IdpMetadata{
		ResourceName: "name",
		Resource:     "file",
		ResourceType: utils.File,
	}, nil)
	res, err := testSubject.GetIdpResources(ctx, mock.Anything, "")
	assert.Equal(t, "name", res.Body.(api.IdpResources).ResourceName)
	assert.Equal(t, "file", res.Body.(api.IdpResources).Resource)
	assert.Equal(t, utils.File, res.Body.(api.IdpResources).Type)
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Nil(t, err)
}

func TestGetIdpResources_dbError(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	authMock := MocksList.AuthManagerMock
	idpMock := MocksList.IdpMock
	authMock.On("GetRealmAndTenantID", ctx, "").Return(ptr.To(realm), ptr.To(tenantID), nil)
	idpMock.On("GetIdpMetadata", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))
	res, err := testSubject.GetIdpResources(ctx, mock.Anything, "")
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.NotNil(t, err)
}

func TestUpdateIdp_HappyFlow(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	dalMock := MocksList.IdpMock
	authMock := MocksList.AuthManagerMock
	idpCacheMock := MocksList.IdpCacheMock
	idpsListCacheMock := MocksList.IdpsListCacheMock
	idpMappersCacheMock := MocksList.IdpMappersCacheMock
	wg := sync.WaitGroup{}
	wg.Add(1)

	mappers := generateMapperConfigArray()
	authMock.On("GetRealmAndTenantID", mock.Anything, mock.Anything).Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("GetIdpConfigFromFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(signedSamlConfig(t), nil)
	iamMock.MockIdpService.On("UpdateIdp", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iamMock.MockIdpService.On("GetIdp", mock.Anything, mock.Anything, mock.Anything).Return(&iamModels.Idp{}, nil)
	iamMock.MockIdpService.On("GetIdpMappers", mock.Anything, mock.Anything, mock.Anything).Return(mappers, nil)
	iamMock.MockIdpService.On("UpdateIdpMappers", mock.Anything, mock.Anything, mock.Anything, mappers).Return(nil)
	idpCacheMock.On("GetIdpForRealm", ctx, mock.Anything, mock.Anything, mock.Anything).Return(&iamModels.Idp{}, nil)
	idpCacheMock.On("DeleteIdpFromRealm", ctx, mock.Anything, mock.Anything).Return(nil)
	idpsListCacheMock.On("DeleteIdpsListFromRealm", ctx, mock.Anything).Return(nil)
	idpMappersCacheMock.On("DeleteIdpMappersFromRealm", ctx, mock.Anything, mock.Anything).Return(nil)
	dalMock.On("GetIdpMetadata", mock.Anything, mock.Anything, mock.Anything).Run(func(arg mock.Arguments) {
	}).Return(getIdpMetadataFromDB(), nil)
	dalMock.On("UpdateIdpMetadata", mock.Anything, mock.Anything).Run(func(arg mock.Arguments) {
		defer wg.Done()
	}).Return(nil)

	res, err := testSubject.UpdateIdp(ctx, mock.Anything, getSAMLIdpCreationRequest(utils.File, XML_FILE), "")
	wg.Wait()
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Nil(t, err)
}

func TestUpdateIdp_MergeConfig_HappyFlow(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	idpCacheMock := MocksList.IdpCacheMock
	idpsListCacheMock := MocksList.IdpsListCacheMock
	idpMappersCacheMock := MocksList.IdpMappersCacheMock
	authMock.On("GetRealmAndTenantID", mock.Anything, mock.Anything).Return(ptr.To(realm), ptr.To(tenantID), nil)
	idpCacheMock.On("GetIdpForRealm", ctx, mock.Anything, mock.Anything, mock.Anything).Return(&iamModels.Idp{}, nil)
	idpsListCacheMock.On("DeleteIdpsListFromRealm", ctx, mock.Anything).Return(nil)
	idpMappersCacheMock.On("DeleteIdpMappersFromRealm", ctx, mock.Anything, mock.Anything).Return(nil)
	idpCacheMock.On("DeleteIdpFromRealm", ctx, mock.Anything, mock.Anything).Return(nil)
	iamMock.MockIdpService.On("UpdateIdp", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iamMock.MockIdpService.On("GetIdp", mock.Anything, mock.Anything, mock.Anything).Return(generateIdentityOidcProviderRepresentation(), nil)
	iamMock.MockIdpService.On("DeleteMandatoryClaimMapper", mock.Anything, realm, idpAlias, utils.OidcMandatoryClaim).Return(nil)

	res, err := testSubject.UpdateIdp(ctx, mock.Anything, getOIDCIdpCreationRequest(), "")
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Nil(t, err)
}

func TestGetIdp_oidc(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock

	mappers := generateMapperConfigArray()
	iamMock.MockIdpService.On("GetIdpMappers", mock.Anything, mock.Anything, mock.Anything).Return(mappers, nil)
	oidc := generateIdentityOidcProviderRepresentation()

	idp := testSubject.iamIdpToApi(ctx, oidc, "test")
	assert.NotNil(t, idp.OidcData)
	assert.NotNil(t, idp.Mappers)
	assert.NotNil(t, idp.OidcData.Scopes)
	assert.Equal(t, []string{"scope1", "scope2"}, idp.OidcData.Scopes)
	assert.Nil(t, idp.SamlData)
	assert.Nil(t, idp.OcpData)
}

func TestGetIdp_redirectUriUsesAliasNotProviderType(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock

	iamMock.MockIdpService.On("GetIdpMappers", mock.Anything, mock.Anything, mock.Anything).Return(generateMapperConfigArray(), nil)
	oidc := generateIdentityOidcProviderRepresentation()
	oidc.Alias = ptr.To("oidc-a1b2c3")

	idp := testSubject.iamIdpToApi(ctx, oidc, realm)
	assert.Equal(t, fmt.Sprintf("https://app.example.com/auth/realms/%s/broker/oidc-a1b2c3/endpoint", realm), idp.RedirectUri)
}

func TestGetIdp_oidc_noScope(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	mappers := generateMapperConfigArray()
	iamMock.MockIdpService.On("GetIdpMappers", mock.Anything, mock.Anything, mock.Anything).Return(mappers, nil)
	oidc := generateIdentityOidcProviderRepresentationNoScope()
	idp := testSubject.iamIdpToApi(ctx, oidc, "test")
	assert.NotNil(t, idp.OidcData)
	assert.NotNil(t, idp.Mappers)
	assert.Nil(t, idp.OidcData.Scopes)
	assert.Nil(t, idp.SamlData)
	assert.Nil(t, idp.OcpData)
}

func TestGetIdp_saml(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock

	mappers := generateMapperConfigArray()
	iamMock.MockIdpService.On("GetIdpMappers", mock.Anything, mock.Anything, mock.Anything).Return(mappers, nil)
	iamMock.MockIdpService.On("GetSPSamlMetadataUrl", mock.Anything, mock.Anything, mock.Anything).Return("https://app.example.com/auth/realms/test/broker/saml-a1b2c3/endpoint/descriptor")
	saml := generateIdentitySamlProviderRepresentation()
	saml.Alias = ptr.To("saml-a1b2c3")

	idp := testSubject.iamIdpToApi(ctx, saml, "test")
	assert.NotNil(t, idp.SamlData)
	assert.NotNil(t, idp.Mappers)
	assert.Equal(t, "https://app.example.com/auth/realms/test/broker/saml-a1b2c3/endpoint/descriptor", idp.SamlData.ServiceProviderMetadataUrl)
	assert.Equal(t, fmt.Sprintf("https://app.example.com/auth/realms/%s/broker/saml-a1b2c3/endpoint", "test"), idp.RedirectUri)
	assert.Nil(t, idp.OidcData)
	assert.Nil(t, idp.OcpData)
}

func TestGetIdp_ocp(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock

	mappers := generateMapperConfigArray()
	iamMock.MockIdpService.On("GetIdpMappers", mock.Anything, mock.Anything, mock.Anything).Return(mappers, nil)
	oidc := generateIdentityOcpProviderRepresentation()

	idp := testSubject.iamIdpToApi(ctx, oidc, "test")
	assert.NotNil(t, idp.OcpData)
	assert.NotNil(t, idp.Mappers)
	assert.NotNil(t, idp.OcpData.Scopes)
	assert.Equal(t, []string{"scope1", "scope2"}, idp.OcpData.Scopes)
	assert.Nil(t, idp.OidcData)
	assert.Nil(t, idp.SamlData)
}

func TestUpdateIdp_MergeConfigWithUrl_HappyFlow(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	idpCacheMock := MocksList.IdpCacheMock
	idpsListCacheMock := MocksList.IdpsListCacheMock
	idpMappersCacheMock := MocksList.IdpMappersCacheMock
	idpCacheMock.On("GetIdpForRealm", ctx, mock.Anything, mock.Anything, mock.Anything).Return(&iamModels.Idp{}, nil)
	idpsListCacheMock.On("DeleteIdpsListFromRealm", ctx, mock.Anything).Return(nil)
	idpMappersCacheMock.On("DeleteIdpMappersFromRealm", ctx, mock.Anything, mock.Anything).Return(nil)
	idpCacheMock.On("DeleteIdpFromRealm", ctx, mock.Anything, mock.Anything).Return(nil)
	authMock.On("GetRealmAndTenantID", mock.Anything, mock.Anything).Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("UpdateIdp", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	iamMock.MockIdpService.On("GetIdp", mock.Anything, mock.Anything, mock.Anything).Return(generateIdentityOidcWithUrlProviderRepresentation(), nil)
	iamMock.MockIdpService.On("DeleteMandatoryClaimMapper", mock.Anything, realm, idpAlias, utils.OidcMandatoryClaim).Return(nil)

	res, err := testSubject.UpdateIdp(ctx, mock.Anything, getOIDCIdpCreationRequest(), "")
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Nil(t, err)
}

func TestUpdateIdp_Update_OnlyMappers(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	idpCacheMock := MocksList.IdpCacheMock
	idpsListCacheMock := MocksList.IdpsListCacheMock
	idpMappersCacheMock := MocksList.IdpMappersCacheMock
	idpCacheMock.On("GetIdpForRealm", ctx, mock.Anything, mock.Anything, mock.Anything).Return(&iamModels.Idp{}, nil)
	idpsListCacheMock.On("DeleteIdpsListFromRealm", ctx, mock.Anything).Return(nil)
	idpMappersCacheMock.On("DeleteIdpMappersFromRealm", ctx, mock.Anything, mock.Anything).Return(nil)
	idpCacheMock.On("DeleteIdpFromRealm", ctx, mock.Anything, mock.Anything).Return(nil)
	mappers := generateMapperConfigArray()
	authMock.On("GetRealmAndTenantID", mock.Anything, mock.Anything).Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("GetIdp", mock.Anything, mock.Anything, mock.Anything).Return(&iamModels.Idp{}, nil)
	iamMock.MockIdpService.On("GetIdpMappers", mock.Anything, mock.Anything, mock.Anything).Return(mappers, nil)
	iamMock.MockIdpService.On("UpdateIdpMappers", mock.Anything, mock.Anything, mock.Anything, mappers).Return(nil)

	res, err := testSubject.UpdateIdp(ctx, mock.Anything, getEmptyDataIdpCreationRequest(), "")
	assert.Equal(t, http.StatusOK, res.Code)
	assert.Nil(t, err)
}

func TestUpdateIdp_Config_Error(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	authMock.On("GetRealmAndTenantID", mock.Anything, mock.Anything).Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("GetIdpConfigFromFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))
	res, err := testSubject.UpdateIdp(ctx, mock.Anything, getSAMLIdpCreationRequest(utils.File, XML_FILE), "")
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.NotNil(t, err)
}

func TestUpdateIdp_InvalidBase64MetadataFile(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	authMock := MocksList.AuthManagerMock
	authMock.On("GetRealmAndTenantID", mock.Anything, mock.Anything).Return(ptr.To(realm), ptr.To(tenantID), nil)

	res, err := testSubject.UpdateIdp(ctx, mock.Anything, getSAMLIdpCreationRequest(utils.File, INVALID_XML_FILE), "")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, http.StatusBadRequest, err.(utils.Error).Code)
}

func TestUpdateIdp_MetadataRejectedByKeycloak(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	authMock.On("GetRealmAndTenantID", mock.Anything, mock.Anything).Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("GetIdpConfigFromFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, iamError.HandleKeyCloakError(&gocloak.APIError{Code: http.StatusBadRequest, Message: "400 Bad Request"}))

	res, err := testSubject.UpdateIdp(ctx, mock.Anything, getSAMLIdpCreationRequest(utils.File, XML_FILE), "")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusBadRequest, res.Code)
	assert.Equal(t, http.StatusBadRequest, err.(utils.Error).Code)
}

func TestUpdateIdp_IdpNotFound(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	authMock.On("GetRealmAndTenantID", mock.Anything, mock.Anything).Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("GetIdpConfigFromFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(signedSamlConfig(t), nil)
	iamMock.MockIdpService.On("UpdateIdp", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(fmt.Errorf("failed to get idp: %w", iamError.HandleKeyCloakError(&gocloak.APIError{Code: http.StatusNotFound, Message: "404 Not Found"})))

	res, err := testSubject.UpdateIdp(ctx, mock.Anything, getSAMLIdpCreationRequest(utils.File, XML_FILE), "")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusNotFound, res.Code)
	assert.Equal(t, http.StatusNotFound, err.(utils.Error).Code)
}

func TestUpdateIdp_KeycloakServerError(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	authMock.On("GetRealmAndTenantID", mock.Anything, mock.Anything).Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("GetIdpConfigFromFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil, iamError.HandleKeyCloakError(&gocloak.APIError{Code: http.StatusInternalServerError, Message: "500 Internal Server Error"}))

	res, err := testSubject.UpdateIdp(ctx, mock.Anything, getSAMLIdpCreationRequest(utils.File, XML_FILE), "")

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.Code)
}

func TestUpdateIdp_Update_Error(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	authMock := MocksList.AuthManagerMock
	authMock.On("GetRealmAndTenantID", mock.Anything, mock.Anything).Return(ptr.To(realm), ptr.To(tenantID), nil)
	iamMock.MockIdpService.On("GetIdpConfigFromFile", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(signedSamlConfig(t), nil)
	iamMock.MockIdpService.On("UpdateIdp", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
	res, err := testSubject.UpdateIdp(ctx, mock.Anything, getSAMLIdpCreationRequest(utils.File, XML_FILE), "")
	assert.Equal(t, http.StatusInternalServerError, res.Code)
	assert.NotNil(t, err)
}

func TestGetIdpFileResourceAsString_HappyFlow(t *testing.T) {
	_, testSubject, _ := initIdpServiceTests(t)
	idpRequest := getIdpMetadataModel(idpAlias, XML_FILE, utils.File, models.SAML, nil, nil)
	res, err := testSubject.getIdpResourceAsString(context.Background(), idpRequest)
	assert.Equal(t, SAML_XML, res)
	assert.Nil(t, err)
}

func TestGetIdpResourceAsString_Error(t *testing.T) {
	_, testSubject, _ := initIdpServiceTests(t)
	idpRequest := getIdpMetadataModel(idpAlias, "0", utils.File, models.SAML, nil, nil)
	res, err := testSubject.getIdpResourceAsString(context.Background(), idpRequest)
	assert.Equal(t, "", res)
	assert.NotNil(t, err)
}

func TestGetIdpURLResourceAsString_HappyFlow(t *testing.T) {
	_, testSubject, _ := initIdpServiceTests(t)
	idpRequest := getIdpMetadataModel(idpAlias, XML_URL, utils.Url, models.SAML, nil, nil)
	res, err := testSubject.getIdpResourceAsString(context.Background(), idpRequest)
	assert.NotNil(t, res)
	assert.Nil(t, err)
}

func TestGetIdpURLResourceAsString_Error(t *testing.T) {
	_, testSubject, _ := initIdpServiceTests(t)
	idpRequest := getIdpMetadataModel(idpAlias, "123", utils.Url, models.SAML, nil, nil)
	res, err := testSubject.getIdpResourceAsString(context.Background(), idpRequest)
	assert.Equal(t, "", res)
	assert.NotNil(t, err)
}

func TestSaveIdpMetadata_HappyFlow(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	idpMock := MocksList.IdpMock
	idpMock.On("SaveIdpMetadata", mock.Anything, mock.Anything).Return(nil)
	idpMock.On("GetIdpMetadata", mock.Anything, mock.Anything).Return(nil, nil)
	err := testSubject.saveIdpMetadata(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	assert.Nil(t, err)
}

func TestSaveIdpMetadata_DB_Error(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	idpMock := MocksList.IdpMock
	idpMock.On("SaveIdpMetadata", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
	idpMock.On("GetIdpMetadata", mock.Anything, mock.Anything).Return(nil, nil)
	err := testSubject.saveIdpMetadata(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	assert.NotNil(t, err)
}

func TestShouldSaveIdpResources(t *testing.T) {
	assert.True(t, shouldSaveIdpResources("save"))
	assert.False(t, shouldSaveIdpResources(""))
}

func TestShouldGetConfigFromUrl(t *testing.T) {
	assert.True(t, shouldGetConfigFromUrl(getIdpMetadataModel(idpAlias, XML_URL, utils.Url, models.SAML, nil, nil)))
	assert.True(t, shouldGetConfigFromUrl(getIdpMetadataModel(idpAlias, XML_URL, utils.Url, models.OIDC, nil, nil)))

	assert.False(t, shouldGetConfigFromUrl(getIdpMetadataModel(idpAlias, XML_FILE, utils.File, models.SAML, nil, nil)))
	assert.False(t, shouldGetConfigFromUrl(getIdpMetadataModel(idpAlias, "", utils.Url, models.SAML, nil, nil)))
	assert.False(t, shouldGetConfigFromUrl(getIdpMetadataModel(idpAlias, "", "", models.OCP, nil, nil)))
}

func TestShouldMergeConfigs(t *testing.T) {
	assert.True(t, shouldMergeConfigs(getIdpMetadataModel(idpAlias, "", utils.Url, models.OIDC, nil, nil)))
	assert.True(t, shouldMergeConfigs(getIdpMetadataModel(idpAlias, "", utils.Url, models.OCP, nil, nil)))

	assert.False(t, shouldMergeConfigs(getIdpMetadataModel(idpAlias, "", utils.Url, models.SAML, nil, nil)))
	assert.False(t, shouldMergeConfigs(getIdpMetadataModel(idpAlias, XML_URL, "", models.OIDC, nil, nil)))
	assert.False(t, shouldMergeConfigs(getIdpMetadataModel(idpAlias, XML_URL, "", models.OCP, nil, nil)))
}

func TestHasResources_true(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	idpMock := MocksList.IdpMock
	idpMock.On("GetIdpMetadata", mock.Anything, mock.Anything, mock.Anything).Return(getIdpMetadataFromDB(), nil)
	assert.True(t, testSubject.hasResources(ctx, "tenantId", idpAlias))
}

func TestHasResources_false(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	idpMock := MocksList.IdpMock
	idpMock.On("GetIdpMetadata", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	assert.False(t, testSubject.hasResources(ctx, "tenantid", idpAlias))
}

func getIdpMetadataFromDB() *models.IdpMetadata {
	return &models.IdpMetadata{
		TenantId:     "tenantId",
		Alias:        idpAlias,
		ResourceName: "name",
		Resource:     "file",
		ResourceType: utils.File,
	}
}

func getEmptyDataIdpCreationRequest() api.IdpCreationRequest {
	return api.IdpCreationRequest{
		Type:     "",
		Name:     "",
		SamlData: nil,
		OidcData: nil,
		OcpData:  nil,
		Mappers: api.Mappers{
			Gid:    "gid",
			Groups: "groups",
			Email:  "email",
		},
	}
}

func getSAMLIdpCreationRequest(metadataXmlType string, metadataXmlFile string) api.IdpCreationRequest {
	return api.IdpCreationRequest{
		Type: api.IdpCreationRequestTypeEnum(models.SAML),
		Name: "saml-idp",
		SamlData: &api.SamlCreationData{
			MetadataXmlType: api.SamlCreationDataMetadataXmlTypeEnum(metadataXmlType),
			MetadataXmlFile: metadataXmlFile,
		},
		OidcData: nil,
		OcpData:  nil,
		Mappers: api.Mappers{
			Gid:    "gid",
			Groups: "groups",
			Email:  "email",
		},
	}
}

func getOIDCIdpCreationRequest() api.IdpCreationRequest {
	return api.IdpCreationRequest{
		Type:     api.IdpCreationRequestTypeEnum(models.OIDC),
		Name:     "oidc",
		SamlData: nil,
		OidcData: &api.OidcCreationData{
			ClientId:            "a",
			ClientSecret:        "a",
			DiscoverDocumentUrl: "",
			Scopes:              []string{"openid", "profile", "email"},
		},
		OcpData: nil,
		Mappers: api.Mappers{
			Uid:                 "uid",
			Gid:                 "gid",
			Groups:              "groups",
			Email:               "email",
			SupplementaryGroups: "supplementaryGroups",
		},
	}
}

func generateNewMappers() api.Mappers {
	return api.Mappers{
		Gid: "new-gid",
	}
}

func generateEmptyNewMappers() api.Mappers {
	return api.Mappers{
		Gid: "",
	}
}

func generateMapperConfigArray() []*iamModels.IdpMapper {
	return []*iamModels.IdpMapper{
		{
			Name:                   ptr.To("name"),
			IdentityProviderMapper: ptr.To("identityProviderMapper"),
			IdentityProviderAlias:  ptr.To("identityProviderAlias"),
			Config: &map[string]string{
				providers.MapperAttributeName: "GID",
				providers.MapperUserAttribute: "old-gid",
			},
		},
	}
}

func generateIdentityProviderRepresentation() *iamModels.Idp {
	return &iamModels.Idp{
		ProviderID: ptr.To("saml"),
		Alias:      ptr.To(idpAlias),
	}
}

func generateIdentityOidcProviderRepresentation() *iamModels.Idp {
	return &iamModels.Idp{
		ProviderID: ptr.To("oidc"),
		Alias:      ptr.To("oidc"),
		Config: &map[string]string{
			"clientId":             "a",
			"clientSecret":         "b",
			"discoveryDocumentUrl": "",
			"defaultScope":         "scope1 scope2",
		},
	}
}

func generateIdentityOidcProviderRepresentationNoScope() *iamModels.Idp {
	return &iamModels.Idp{
		ProviderID: ptr.To("oidc"),
		Alias:      ptr.To("oidc"),
		Config: &map[string]string{
			"clientId":             "a",
			"clientSecret":         "b",
			"discoveryDocumentUrl": "",
		},
	}
}

func generateIdentitySamlProviderRepresentation() *iamModels.Idp {
	return &iamModels.Idp{
		ProviderID: ptr.To("saml"),
		Alias:      ptr.To("saml"),
		Config: &map[string]string{
			"signingCertificate":     "signingCertificate",
			"singleSignOnServiceUrl": "singleSignOnServiceUrl",
			"entityId":               "entityId",
		},
	}
}

func generateIdentityOcpProviderRepresentation() *iamModels.Idp {
	return &iamModels.Idp{
		ProviderID: ptr.To("openshift-v4"),
		Alias:      ptr.To("openshift-v4"),
		Config: &map[string]string{
			"clientId":             "a",
			"clientSecret":         "b",
			"discoveryDocumentUrl": "",
			"defaultScope":         "scope1 scope2",
		},
	}
}

func generateIdentityOidcWithUrlProviderRepresentation() *iamModels.Idp {
	return &iamModels.Idp{
		ProviderID: ptr.To("oidc"),
		Alias:      ptr.To("oidc"),
		Config: &map[string]string{
			"idpBaseUrl":   "clientSecret",
			"clientId":     "clientSecret",
			"clientSecret": "clientSecret",
			"defaultScope": "scope1 scope2",
		},
	}
}

func getIdpMetadataModel(alias, resource, resourceType string, idp models.ProviderID, mappers map[string]string, config *map[string]string) models.IdpData {
	return models.IdpData{
		Alias:        alias,
		Idp:          idp,
		Resource:     resource,
		ResourceType: resourceType,
		Mappers:      mappers,
		Config:       config,
	}
}

func TestSetIdpSettings_HappyFlow(t *testing.T) {
	assert.Nil(t, nil)
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock

	request := getOidcIdpCreationRequestWithMandatoryClaim([]string{"valid-value"})
	iamMock.MockIdpService.On("SetMandatotyClaimMapper", ctx, realm, idpAlias, mock.Anything).Return(nil)

	err := testSubject.setIdpSettings(ctx, realm, idpAlias, true, request)
	assert.Nil(t, err)
	iamMock.MockIdpService.AssertExpectations(t)
}

func TestSetIdpSettings_SetClaimError(t *testing.T) {
	assert.Nil(t, nil)
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock

	request := getOidcIdpCreationRequestWithMandatoryClaim([]string{"valid-value"})
	iamMock.MockIdpService.On("SetMandatotyClaimMapper", ctx, realm, idpAlias, mock.Anything).Return(fmt.Errorf("error"))

	err := testSubject.setIdpSettings(ctx, realm, idpAlias, true, request)
	assert.NotNil(t, err)
	iamMock.MockIdpService.AssertExpectations(t)
}

func TestSetIdpSettings_DeleteMapper(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	request := api.IdpCreationRequest{
		Type: api.IdpCreationRequestTypeEnum(models.OIDC),
		OidcData: &api.OidcCreationData{
			ClientId:            "test-client",
			ClientSecret:        "test-secret",
			DiscoverDocumentUrl: URL,
		},
	}
	iamMock.MockIdpService.On("DeleteMandatoryClaimMapper", mock.Anything, realm, idpAlias, utils.OidcMandatoryClaim).Return(nil)
	err := testSubject.setIdpSettings(ctx, realm, idpAlias, false, request)
	assert.Nil(t, err)
	iamMock.MockIdpService.AssertExpectations(t)
}

func TestSetIdpSettings_DeleteMapperError(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock

	request := api.IdpCreationRequest{
		Type: api.IdpCreationRequestTypeEnum(models.OIDC),
		OidcData: &api.OidcCreationData{
			ClientId:            "test-client",
			ClientSecret:        "test-secret",
			DiscoverDocumentUrl: URL,
		},
	}
	iamMock.MockIdpService.On("DeleteMandatoryClaimMapper", mock.Anything, realm, idpAlias, utils.OidcMandatoryClaim).Return(fmt.Errorf("error"))
	err := testSubject.setIdpSettings(ctx, realm, idpAlias, false, request)
	assert.NotNil(t, err)
	iamMock.MockIdpService.AssertExpectations(t)
}

func TestSetIdpSettings_EmptyValuesArray(t *testing.T) {
	_, testSubject, ctx := initIdpServiceTests(t)

	request := getOidcIdpCreationRequestWithMandatoryClaim([]string{})

	err := testSubject.setIdpSettings(ctx, realm, idpAlias, true, request)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid mandatory claim values", err.Error())
}

func TestSetIdpSettings_NilValuesArray(t *testing.T) {
	_, testSubject, ctx := initIdpServiceTests(t)

	request := getOidcIdpCreationRequestWithMandatoryClaim(nil)

	err := testSubject.setIdpSettings(ctx, realm, idpAlias, true, request)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid mandatory claim values", err.Error())
}

func TestSetIdpSettings_NonOidcIdp(t *testing.T) {
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock

	request := api.IdpCreationRequest{
		Type: api.IdpCreationRequestTypeEnum(models.SAML),
		SamlData: &api.SamlCreationData{
			MetadataXmlType: "url",
			MetadataXmlUrl:  URL,
		},
	}

	err := testSubject.setIdpSettings(ctx, realm, idpAlias, true, request)
	assert.Nil(t, err)
	iamMock.MockIdpService.AssertNotCalled(t, "SetMandatotyClaimMapper")
	iamMock.MockIdpService.AssertNotCalled(t, "DeleteMandatoryClaimMapper")
}

func TestSetIdpSettings_NilMandatoryClaim(t *testing.T) {
	_, testSubject, ctx := initIdpServiceTests(t)

	request := api.IdpCreationRequest{
		Type: api.IdpCreationRequestTypeEnum(models.OIDC),
		OidcData: &api.OidcCreationData{
			ClientId:            "test-client",
			ClientSecret:        "test-secret",
			DiscoverDocumentUrl: URL,
			MandatoryClaim:      nil,
		},
	}

	err := testSubject.setIdpSettings(ctx, realm, idpAlias, true, request)
	assert.Nil(t, err)
}

func getOidcIdpCreationRequestWithMandatoryClaim(values []string) api.IdpCreationRequest {
	return api.IdpCreationRequest{
		Type: api.IdpCreationRequestTypeEnum(models.OIDC),
		OidcData: &api.OidcCreationData{
			ClientId:            "test-client",
			ClientSecret:        "test-secret",
			DiscoverDocumentUrl: URL,
			MandatoryClaim: &api.MandatoryClaim{
				Claim:  "test-claim",
				Values: values,
			},
		},
	}
}

func TestSetIdpSettings_SpacesInMandatoryClaimValues(t *testing.T) {
	assert.Nil(t, nil)
	MocksList, testSubject, ctx := initIdpServiceTests(t)
	iamMock := MocksList.IamServicesContainerMock
	request := getOidcIdpCreationRequestWithMandatoryClaim([]string{"value with space", "another value"})
	err := testSubject.setIdpSettings(ctx, "realm", "test-alias", true, request)
	assert.NotNil(t, err)
	assert.Equal(t, "invalid mandatory claim value - spaces not allowed", err.Error())
	iamMock.MockIdpService.AssertNotCalled(t, "SetMandatotyClaimMapper")
}

func signedSamlConfig(t *testing.T) map[string]string {
	t.Helper()
	return map[string]string{
		utils.SigningCertificate: selfSignedSigningCertificate(t),
	}
}

func selfSignedSigningCertificate(t *testing.T) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	assert.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test-idp"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	assert.NoError(t, err)
	return base64.StdEncoding.EncodeToString(der)
}
