/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"testing"
	"time"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	"github.com/stretchr/testify/require"
)

var samlMetadataUrl string

const samlEntityId = "urn:example:idp"

func init() {
	samlMetadataUrl = "http://" + KeycloakDomain + ":8090/auth/realms/master/protocol/saml/descriptor"
}

func TestSAMLIdp_CreateAndDelete(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-create")
	alias, entityId, metadataUrl, mappers := newRandomSamlIdpTestDataWithRealm()
	idpRequest := makeSamlIdpRequest(alias, entityId, metadataUrl, mappers)
	resp, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(idpRequest).Execute()
	defer deleteIDP(t, ctx, alias)

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.Equal(t, alias, resp.Alias)
}

func TestSAMLIdp_Get(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-get")
	alias, entityId, metadataUrl, mappers := newRandomSamlIdpTestDataWithRealm()
	createSamlIdp(t, ctx, alias, entityId, metadataUrl, mappers)
	defer deleteIDP(t, ctx, alias)
	resp, err := OpenapiClient.IdpsAPI.GetIdp(ctx, alias).Execute()

	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestSAMLIdp_Update(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-update")
	alias, entityId, metadataUrl, mappers := newRandomSamlIdpTestDataWithRealm()
	createSamlIdp(t, ctx, alias, entityId, metadataUrl, mappers)
	defer deleteIDP(t, ctx, alias)
	newEntityId := entityId + "-updated"
	updateRequest := makeSamlIdpRequest(alias, newEntityId, metadataUrl, mappers)

	_, err := OpenapiClient.IdpsAPI.UpdateIdp(ctx, alias).IdpCreationRequest(updateRequest).Execute()
	require.NoError(t, err)
}

func TestSAMLIdp_List(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-list")
	alias, entityId, metadataUrl, mappers := newRandomSamlIdpTestDataWithRealm()
	createSamlIdp(t, ctx, alias, entityId, metadataUrl, mappers)
	defer deleteIDP(t, ctx, alias)

	listResp, httpResp, err := OpenapiClient.IdpsAPI.GetIdps(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	found := false
	for _, idp := range listResp {
		if idp.Alias == alias {
			found = true
			break
		}
	}
	require.True(t, found, "Created SAML IdP should be in the list")
}

func TestSAMLIdp_Get_NotFound(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-get-notfound")
	alias := "non-existent-saml-idp-alias"
	httpResp, err := OpenapiClient.IdpsAPI.GetIdp(ctx, alias).Execute()
	OpenapiClient.IdpsAPI.GetIdp(ctx, alias)
	requireApiError(t, err, httpResp, http.StatusNotFound)
}

func TestSAMLIdp_Create_Conflict(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-conflict")
	alias, entityId, metadataUrl, mappers := newRandomSamlIdpTestDataWithRealm()
	createSamlIdp(t, ctx, alias, entityId, metadataUrl, mappers)
	defer deleteIDP(t, ctx, alias)

	idpRequest := makeSamlIdpRequest(alias, entityId, metadataUrl, mappers)
	_, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(idpRequest).Execute()
	requireApiError(t, err, httpResp, http.StatusConflict)
}

func TestSAMLIdp_Create_InvalidBase64MetadataFile(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-create-invalid-base64")
	alias := GenerateRandomString("saml-idp-invalid-base64")
	idpRequest := makeSamlIdpRawFileRequest(alias, samlEntityId, "metadata.xml", "not!valid!base64", GetDefaultIDPMappers())

	_, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(idpRequest).Execute()
	requireApiError(t, err, httpResp, http.StatusBadRequest)
}

func TestSAMLIdp_Create_UnparseableMetadataFile(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-create-unparseable")
	alias := GenerateRandomString("saml-idp-unparseable")
	idpRequest := makeSamlIdpFileRequest(alias, samlEntityId, "metadata.xml", "this is not saml metadata", GetDefaultIDPMappers())

	_, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(idpRequest).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
}

func TestSAMLIdp_Create_UnresolvableMetadataUrl(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-create-unresolvable-url")
	alias := GenerateRandomString("saml-idp-unresolvable-url")
	idpRequest := makeSamlIdpRequest(alias, samlEntityId, "http://no-such-host.invalid/descriptor", GetDefaultIDPMappers())

	_, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(idpRequest).Execute()
	requireApiError(t, err, httpResp, http.StatusBadRequest)
}

func TestSAMLIdp_Update_NotFound(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-update-notfound")
	alias, entityId, metadataUrl, mappers := newRandomSamlIdpTestDataWithRealm()
	idpRequest := makeSamlIdpRequest(alias, entityId, metadataUrl, mappers)
	httpResp, err := OpenapiClient.IdpsAPI.UpdateIdp(ctx, alias).IdpCreationRequest(idpRequest).Execute()
	requireApiError(t, err, httpResp, http.StatusNotFound)
}

func TestSAMLIdp_Update_InvalidBase64MetadataFile(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-update-invalid-base64")
	alias, entityId, metadataUrl, mappers := newRandomSamlIdpTestDataWithRealm()
	createSamlIdp(t, ctx, alias, entityId, metadataUrl, mappers)
	defer deleteIDP(t, ctx, alias)
	invalidRequest := makeSamlIdpRawFileRequest(alias, entityId, "metadata.xml", "not!valid!base64", mappers)

	httpResp, err := OpenapiClient.IdpsAPI.UpdateIdp(ctx, alias).IdpCreationRequest(invalidRequest).Execute()
	requireApiError(t, err, httpResp, http.StatusBadRequest)
}

func TestSAMLIdp_Update_UnparseableMetadataFile(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-update-unparseable")
	alias, entityId, metadataUrl, mappers := newRandomSamlIdpTestDataWithRealm()
	createSamlIdp(t, ctx, alias, entityId, metadataUrl, mappers)
	defer deleteIDP(t, ctx, alias)
	invalidRequest := makeSamlIdpFileRequest(alias, entityId, "metadata.xml", "this is not saml metadata", mappers)

	httpResp, err := OpenapiClient.IdpsAPI.UpdateIdp(ctx, alias).IdpCreationRequest(invalidRequest).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
}

func TestSAMLIdp_Update_Invalid(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-invalid")
	alias, entityId, metadataUrl, mappers := newRandomSamlIdpTestDataWithRealm()
	createSamlIdp(t, ctx, alias, entityId, metadataUrl, mappers)
	defer deleteIDP(t, ctx, alias)
	invalidRequest := makeSamlIdpRequest(alias, entityId, "", mappers) // missing entityId

	httpResp, err := OpenapiClient.IdpsAPI.UpdateIdp(ctx, alias).IdpCreationRequest(invalidRequest).Execute()
	requireApiError(t, err, httpResp, http.StatusBadRequest)
}

func TestSAMLIdp_Delete_NotFound(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-delete-notfound")
	alias := "non-existent-saml-idp-alias"
	httpResp, err := OpenapiClient.IdpsAPI.DeleteIdp(ctx, alias).Execute()
	requireApiError(t, err, httpResp, http.StatusNotFound)
}

func TestSAMLIdp_Create_MissingFields(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-missingfields")
	_, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(openapi.IdpCreationRequest{
		Name: StrPtr("missing-metadataxmltype"),
		Type: openapi.IDPCREATIONREQUESTTYPEENUM_SAML,
		SamlData: *openapi.NewNullableSamlCreationData(&openapi.SamlCreationData{
			MetadataXmlType: nil,
			MetadataXmlUrl:  StrPtr(samlMetadataUrl),
			EntityId:        StrPtr(samlEntityId),
		}),
		Mappers: GetDefaultIDPMappers(),
	}).Execute()
	requireApiError(t, err, httpResp, http.StatusBadRequest)
}

func TestSAMLIdp_List_Empty(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-listempty")
	listResp, _, err := OpenapiClient.IdpsAPI.GetIdps(ctx).Execute()
	require.NoError(t, err)
	require.Empty(t, listResp)
}

func TestSAMLIdp_Get_AfterDelete(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-get-afterdelete")
	alias, entityId, metadataUrl, mappers := newRandomSamlIdpTestDataWithRealm()
	createSamlIdp(t, ctx, alias, entityId, metadataUrl, mappers)
	deleteIDP(t, ctx, alias)

	_, err := OpenapiClient.IdpsAPI.GetIdp(ctx, alias).Execute()
	require.Error(t, err)
}

func TestSAMLIdp_GetIdpMappers(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-get-idp-mappers")
	alias, entityId, metadataUrl, mappers := newRandomSamlIdpTestDataWithRealm()
	createSamlIdp(t, ctx, alias, entityId, metadataUrl, mappers)
	defer deleteIDP(t, ctx, alias)

	result, _, err := OpenapiClient.IdpsAPI.GetIdpMappers(ctx, alias).Execute()
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, mappers.Gid, result.Gid)
	require.Equal(t, mappers.Uid, result.Uid)
	require.Equal(t, mappers.Groups, result.Groups)
	require.Equal(t, mappers.SupplementaryGroups, result.SupplementaryGroups)
	require.Equal(t, mappers.Email, result.Email)
	require.Equal(t, mappers.FirstName, result.FirstName)
	require.Equal(t, mappers.LastName, result.LastName)
}

func TestSAMLIdp_Create_FileUpload_FromDisk(t *testing.T) {
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-file-disk")
	alias := GenerateRandomString("saml-idp-file-disk")
	entityId := "urn:test:idp"
	fileName := "test-metadata.xml"
	xmlContent, _ := samlMetadataWithValidity(t, entityId, "https://idp.example.com/sso", time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	mappers := GetDefaultIDPMappers()
	idpRequest := makeSamlIdpFileRequest(alias, entityId, fileName, xmlContent, mappers)
	defer deleteIDP(t, ctx, alias)

	resp, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(idpRequest).Execute()
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.Equal(t, alias, resp.Alias)
	require.NotEmpty(t, samlSigningCertificate(t, ctx, alias))
}

func TestSAMLIdp_Create_MetadataWithoutSigningCertificate(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-no-cert")
	alias := GenerateRandomString("saml-idp-no-cert")
	entityID := "urn:test:no-cert"
	xml := `<EntityDescriptor entityID="` + entityID + `" xmlns="urn:oasis:names:tc:SAML:2.0:metadata">
  <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol"/>
</EntityDescriptor>`
	idpRequest := makeSamlIdpFileRequest(alias, entityID, "metadata.xml", xml, GetDefaultIDPMappers())

	_, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(idpRequest).Execute()
	requireApiError(t, err, httpResp, http.StatusBadRequest)
	require.Empty(t, samlSigningCertificate(t, ctx, alias))
}

func TestSAMLIdp_Create_CertificateSignedByPrivateCA(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-untrusted-cert")
	alias := GenerateRandomString("saml-idp-untrusted-cert")
	entityID := "urn:test:untrusted-cert"
	idpRequest := makeSamlIdpFileRequest(alias, entityID, "metadata.xml", samlMetadataWithUntrustedCertificate(t, entityID), GetDefaultIDPMappers())
	defer deleteIDP(t, ctx, alias)

	resp, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(idpRequest).Execute()
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.Equal(t, alias, resp.Alias)
	require.NotEmpty(t, samlSigningCertificate(t, ctx, alias))
}

func TestSAMLIdp_Update_UnsignedCertificate(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-update-unsigned")
	alias, entityID, metadataURL, mappers := newRandomSamlIdpTestDataWithRealm()
	createSamlIdp(t, ctx, alias, entityID, metadataURL, mappers)
	defer deleteIDP(t, ctx, alias)
	require.NotEmpty(t, samlSigningCertificate(t, ctx, alias))

	xml := `<EntityDescriptor entityID="` + entityID + `" xmlns="urn:oasis:names:tc:SAML:2.0:metadata">
  <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol"/>
</EntityDescriptor>`
	updateRequest := makeSamlIdpFileRequest(alias, entityID, "metadata.xml", xml, mappers)

	httpResp, err := OpenapiClient.IdpsAPI.UpdateIdp(ctx, alias).IdpCreationRequest(updateRequest).Execute()
	requireApiError(t, err, httpResp, http.StatusBadRequest)
	require.NotEmpty(t, samlSigningCertificate(t, ctx, alias))
}

func TestSAMLIdp_CreateDryRun_ReturnsConfigWithoutSaving(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-dry-run")
	alias := GenerateRandomString("saml-idp-dry-run")
	entityID := "urn:test:dry-run"
	ssoURL := "https://idp.example.com/sso"
	xml, certificate := samlMetadataWithValidity(t, entityID, ssoURL, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	idpRequest := makeSamlIdpFileRequest(alias, entityID, "metadata.xml", xml, GetDefaultIDPMappers())
	idpRequest.SetDryRun(true)

	resp, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(idpRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.Equal(t, alias, resp.GetAlias())
	samlData := resp.GetSamlData()
	require.Equal(t, entityID, samlData.GetEntityId())
	require.Equal(t, ssoURL, samlData.GetSingleSignOnServiceUrl())
	require.Equal(t, certificate, samlData.GetSigningCertificate())
	require.Empty(t, samlSigningCertificate(t, ctx, alias))
}

func TestSAMLIdp_CreateDryRun_ExpiredCertificate(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-dry-run-expired")
	alias := GenerateRandomString("saml-idp-dry-run-expired")
	entityID := "urn:test:dry-run-expired"
	xml, _ := samlMetadataWithValidity(t, entityID, "https://idp.example.com/sso", time.Now().Add(-2*time.Hour), time.Now().Add(-time.Hour))
	idpRequest := makeSamlIdpFileRequest(alias, entityID, "metadata.xml", xml, GetDefaultIDPMappers())
	idpRequest.SetDryRun(true)

	_, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(idpRequest).Execute()
	requireApiError(t, err, httpResp, http.StatusBadRequest)
	require.Empty(t, samlSigningCertificate(t, ctx, alias))
}

func TestSAMLIdp_CreateDryRun_InvalidSamlReturnsParsedBody(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-dry-run-invalid-body")
	alias := GenerateRandomString("saml-idp-dry-run-invalid-body")
	entityID := "https://ci-84de-saml-signature-valida.runailabs.com/auth/realms/saml"
	ssoURL := entityID + "/protocol/saml"
	idpRequest := makeSamlIdpFileRequest(alias, entityID, "metadata.xml", expiredSamlMetadata, GetDefaultIDPMappers())
	idpRequest.SetDryRun(true)

	_, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(idpRequest).Execute()
	requireApiError(t, err, httpResp, http.StatusBadRequest)
	body, readErr := io.ReadAll(httpResp.Body)
	require.NoError(t, readErr)
	var resp openapi.IdpPostResponse
	require.NoError(t, json.Unmarshal(body, &resp))
	require.Equal(t, alias, resp.GetAlias())
	samlData := resp.GetSamlData()
	require.Equal(t, entityID, samlData.GetEntityId())
	require.Equal(t, ssoURL, samlData.GetSingleSignOnServiceUrl())
	require.Equal(t, expiredSamlSigningCertificate, samlData.GetSigningCertificate())
	require.True(t, samlData.GetValidateSignature())
	require.True(t, samlData.GetCertificateExpirationDate().Equal(time.Date(2025, time.January, 2, 0, 0, 0, 0, time.UTC)))
	require.Empty(t, samlSigningCertificate(t, ctx, alias))
}

func TestSAMLIdp_Create_DryRun_MetadataWithoutSigningCertificate(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-dry-run-no-cert")
	alias := GenerateRandomString("saml-idp-dry-run-no-cert")
	entityID := "urn:test:dry-run-no-cert"
	xml := `<EntityDescriptor entityID="` + entityID + `" xmlns="urn:oasis:names:tc:SAML:2.0:metadata">
  <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol"/>
</EntityDescriptor>`
	idpRequest := makeSamlIdpFileRequest(alias, entityID, "metadata.xml", xml, GetDefaultIDPMappers())
	idpRequest.SetDryRun(true)

	_, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(idpRequest).Execute()
	requireApiError(t, err, httpResp, http.StatusBadRequest)
	require.Empty(t, samlSigningCertificate(t, ctx, alias))
}

func TestSAMLIdp_Update_DryRun_ReturnsConfigWithoutSaving(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("saml-idp-update-dry-run")
	alias, entityID, metadataURL, mappers := newRandomSamlIdpTestDataWithRealm()
	createSamlIdp(t, ctx, alias, entityID, metadataURL, mappers)
	defer deleteIDP(t, ctx, alias)
	storedCertificate := samlSigningCertificate(t, ctx, alias)
	require.NotEmpty(t, storedCertificate)

	newEntityID := "urn:test:update-dry-run"
	ssoURL := "https://idp.example.com/sso"
	xml, certificate := samlMetadataWithValidity(t, newEntityID, ssoURL, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
	updateRequest := makeSamlIdpFileRequest(alias, newEntityID, "metadata.xml", xml, mappers)
	updateRequest.SetDryRun(true)

	httpResp, err := OpenapiClient.IdpsAPI.UpdateIdp(ctx, alias).IdpCreationRequest(updateRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	body, readErr := io.ReadAll(httpResp.Body)
	require.NoError(t, readErr)
	var resp openapi.IdpPostResponse
	require.NoError(t, json.Unmarshal(body, &resp))
	require.Equal(t, alias, resp.GetAlias())
	samlData := resp.GetSamlData()
	require.Equal(t, newEntityID, samlData.GetEntityId())
	require.Equal(t, ssoURL, samlData.GetSingleSignOnServiceUrl())
	require.Equal(t, certificate, samlData.GetSigningCertificate())
	require.Equal(t, storedCertificate, samlSigningCertificate(t, ctx, alias))
}

func makeSamlIdpRequest(alias, entityId, metadataUrl string, mappers *openapi.Mappers) openapi.IdpCreationRequest {
	metadataType := openapi.SAMLCREATIONDATAMETADATAXMLTYPEENUM_URL
	return openapi.IdpCreationRequest{
		Name: StrPtr(alias),
		Type: openapi.IDPCREATIONREQUESTTYPEENUM_SAML,
		SamlData: *openapi.NewNullableSamlCreationData(&openapi.SamlCreationData{
			MetadataXmlType: &metadataType,
			MetadataXmlUrl:  StrPtr(metadataUrl),
			EntityId:        StrPtr(entityId),
		}),
		Mappers: mappers,
	}
}

func makeSamlIdpFileRequest(alias, entityId, fileName, xmlContent string, mappers *openapi.Mappers) openapi.IdpCreationRequest {
	metadataType := openapi.SAMLCREATIONDATAMETADATAXMLTYPEENUM_FILE
	xmlBase64 := base64.StdEncoding.EncodeToString([]byte(xmlContent))
	return openapi.IdpCreationRequest{
		Name: StrPtr(alias),
		Type: openapi.IDPCREATIONREQUESTTYPEENUM_SAML,
		SamlData: *openapi.NewNullableSamlCreationData(&openapi.SamlCreationData{
			MetadataXmlType: &metadataType,
			MetadataXmlFile: StrPtr(xmlBase64),
			FileName:        StrPtr(fileName),
			EntityId:        StrPtr(entityId),
		}),
		Mappers: mappers,
	}
}

func makeSamlIdpRawFileRequest(alias, entityId, fileName, metadataXmlFile string, mappers *openapi.Mappers) openapi.IdpCreationRequest {
	metadataType := openapi.SAMLCREATIONDATAMETADATAXMLTYPEENUM_FILE
	return openapi.IdpCreationRequest{
		Name: StrPtr(alias),
		Type: openapi.IDPCREATIONREQUESTTYPEENUM_SAML,
		SamlData: *openapi.NewNullableSamlCreationData(&openapi.SamlCreationData{
			MetadataXmlType: &metadataType,
			MetadataXmlFile: StrPtr(metadataXmlFile),
			FileName:        StrPtr(fileName),
			EntityId:        StrPtr(entityId),
		}),
		Mappers: mappers,
	}
}

func newRandomSamlIdpTestDataWithRealm() (string, string, string, *openapi.Mappers) {
	alias := GenerateRandomString("saml")
	entityId := samlEntityId
	metadataUrl := samlMetadataUrl
	mappers := GetDefaultIDPMappers()
	return alias, entityId, metadataUrl, mappers
}

func createSamlIdp(t *testing.T, ctx context.Context, alias, entityId, metadataUrl string, mappers *openapi.Mappers) {
	idpRequest := makeSamlIdpRequest(alias, entityId, metadataUrl, mappers)
	createIDP(t, ctx, idpRequest)
}

func samlMetadataWithValidity(t *testing.T, entityID, ssoURL string, notBefore, notAfter time.Time) (string, string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: entityID},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	certificate := base64.StdEncoding.EncodeToString(der)
	xml := `<EntityDescriptor entityID="` + entityID + `" xmlns="urn:oasis:names:tc:SAML:2.0:metadata">
  <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <KeyDescriptor use="signing">
      <KeyInfo xmlns="http://www.w3.org/2000/09/xmldsig#">
        <X509Data>
          <X509Certificate>` + certificate + `</X509Certificate>
        </X509Data>
      </KeyInfo>
    </KeyDescriptor>
    <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="` + ssoURL + `"/>
  </IDPSSODescriptor>
</EntityDescriptor>`
	return xml, certificate
}

func samlMetadataWithSigningCertificate(t *testing.T, entityID string) string {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: entityID},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	return `<EntityDescriptor entityID="` + entityID + `" xmlns="urn:oasis:names:tc:SAML:2.0:metadata">
  <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <KeyDescriptor use="signing">
      <KeyInfo xmlns="http://www.w3.org/2000/09/xmldsig#">
        <X509Data>
          <X509Certificate>` + base64.StdEncoding.EncodeToString(der) + `</X509Certificate>
        </X509Data>
      </KeyInfo>
    </KeyDescriptor>
  </IDPSSODescriptor>
</EntityDescriptor>`
}

func samlMetadataWithUntrustedCertificate(t *testing.T, entityID string) string {
	t.Helper()
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "untrusted-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	require.NoError(t, err)
	caCert, err := x509.ParseCertificate(caDER)
	require.NoError(t, err)

	leafKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	leafTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: entityID},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	leafDER, err := x509.CreateCertificate(rand.Reader, leafTemplate, caCert, &leafKey.PublicKey, caKey)
	require.NoError(t, err)
	return samlMetadataWithCertificate(entityID, leafDER)
}

func samlMetadataWithCertificate(entityID string, der []byte) string {
	return `<EntityDescriptor entityID="` + entityID + `" xmlns="urn:oasis:names:tc:SAML:2.0:metadata">
  <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
    <KeyDescriptor use="signing">
      <KeyInfo xmlns="http://www.w3.org/2000/09/xmldsig#">
        <X509Data>
          <X509Certificate>` + base64.StdEncoding.EncodeToString(der) + `</X509Certificate>
        </X509Data>
      </KeyInfo>
    </KeyDescriptor>
  </IDPSSODescriptor>
</EntityDescriptor>`
}

func samlSigningCertificate(t *testing.T, ctx context.Context, alias string) string {
	t.Helper()
	idps, _, err := OpenapiClient.IdpsAPI.GetIdps(ctx).Execute()
	require.NoError(t, err)
	for _, idp := range idps {
		if idp.Alias == alias {
			samlData := idp.GetSamlData()
			return samlData.GetSigningCertificate()
		}
	}
	return ""
}

func createIDP(t *testing.T, ctx context.Context, idpCreationRequest openapi.IdpCreationRequest) {
	err := TestStepsClient.CreateIDP(ctx, idpCreationRequest)
	require.NoError(t, err)
}

func deleteIDP(t *testing.T, ctx context.Context, alias string) {
	err := TestStepsClient.DeleteIDP(ctx, alias)
	require.NoError(t, err)
}

const expiredSamlSigningCertificate = "MIIClzCCAX+gAwIBAgIBATANBgkqhkiG9w0BAQsFADAPMQ0wCwYDVQQDDARzYW1sMB4XDTI1MDEwMTAwMDAwMFoXDTI1MDEwMjAwMDAwMFowDzENMAsGA1UEAwwEc2FtbDCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBAMFzjVhXt5Jc8h9qE8+767WxOHj3umxMBqDmR4l7wzN4cLNB32q5Oc45jwwVlctCf3ropsc6EyblJaSoIfSVq90joOzPOnLMtHBZ/Sp+Z05rqssQ7xXBm/p9MhMvaGxbkgCqqvj6VCHdbNlQCz5jHV45KySkAZ7ETwSPoGB9L5ioYbI+zai4nNlKTlqfm/KKG0C2snK2Ep2sMNeKl/wDtk8cXOKrbPfCpseBU73iETJmgAURv9vGEjRSXo99kBGX+xfSJtEgFdPu9NdLNLkzopvtSjc9AsM1PL3i+y1WofSbN/PplVbgM0oQ+fCoKwEqwAmR3d9NEUYoPuJnf9AZuekCAwEAATANBgkqhkiG9w0BAQsFAAOCAQEAH9TgB1kIUCuf5kWjGj8Pa5VQHxXMslrUKjvAUt0XnR2YNExCIlCezyP9N7ht+nRy1tIAWa4rqOIyY4NIziGKtKiAWigWicA3j3aFxJhajL8Z33UQlY6vUPGvK2zxpZQXaeQiUrrSBlNBBISZjgkHg7TfzR+TYQhpnoik52q4zspZ722q6zmqzayaGD+zPDu9IwoY9Z7Px6BNAwzJGU/4G4IPnqeVh3uGduQ3eslAoFxvADXhMPBwSCTHSFVWwZBgRyTzQQlrToQGDp26jDLzJTcqMQu75hIbe7ICxBmAXqpK7Z23j1AiH3TYN7e3T07pIi9ZtB1K/6DAjYMd6uIyrg=="

const expiredSamlMetadata = `<md:EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" xmlns:md="urn:oasis:names:tc:SAML:2.0:metadata" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" xmlns:ds="http://www.w3.org/2000/09/xmldsig#" entityID="https://ci-84de-saml-signature-valida.runailabs.com/auth/realms/saml"><md:IDPSSODescriptor WantAuthnRequestsSigned="true" protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol"><md:KeyDescriptor use="signing"><ds:KeyInfo><ds:KeyName>2KUdnE6VSX0MRZEuGyTp3uGiRf2yNDwwlWhhmopbuc4</ds:KeyName><ds:X509Data><ds:X509Certificate>` + expiredSamlSigningCertificate + `</ds:X509Certificate></ds:X509Data></ds:KeyInfo></md:KeyDescriptor><md:ArtifactResolutionService Binding="urn:oasis:names:tc:SAML:2.0:bindings:SOAP" Location="https://ci-84de-saml-signature-valida.runailabs.com/auth/realms/saml/protocol/saml/resolve" index="0"></md:ArtifactResolutionService><md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://ci-84de-saml-signature-valida.runailabs.com/auth/realms/saml/protocol/saml"></md:SingleLogoutService><md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://ci-84de-saml-signature-valida.runailabs.com/auth/realms/saml/protocol/saml"></md:SingleLogoutService><md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Artifact" Location="https://ci-84de-saml-signature-valida.runailabs.com/auth/realms/saml/protocol/saml"></md:SingleLogoutService><md:SingleLogoutService Binding="urn:oasis:names:tc:SAML:2.0:bindings:SOAP" Location="https://ci-84de-saml-signature-valida.runailabs.com/auth/realms/saml/protocol/saml"></md:SingleLogoutService><md:NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:persistent</md:NameIDFormat><md:NameIDFormat>urn:oasis:names:tc:SAML:2.0:nameid-format:transient</md:NameIDFormat><md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified</md:NameIDFormat><md:NameIDFormat>urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress</md:NameIDFormat><md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" Location="https://ci-84de-saml-signature-valida.runailabs.com/auth/realms/saml/protocol/saml"></md:SingleSignOnService><md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect" Location="https://ci-84de-saml-signature-valida.runailabs.com/auth/realms/saml/protocol/saml"></md:SingleSignOnService><md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:SOAP" Location="https://ci-84de-saml-signature-valida.runailabs.com/auth/realms/saml/protocol/saml"></md:SingleSignOnService><md:SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Artifact" Location="https://ci-84de-saml-signature-valida.runailabs.com/auth/realms/saml/protocol/saml"></md:SingleSignOnService></md:IDPSSODescriptor></md:EntityDescriptor>`
