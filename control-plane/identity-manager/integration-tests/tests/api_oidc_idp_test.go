/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"testing"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	"github.com/stretchr/testify/require"
)

var discoveryUrl string

func init() {
	discoveryUrl = "http://" + KeycloakDomain + ":8090/auth/realms/master/.well-known/openid-configuration"
}

func TestOIDCIdp_CreateAndDelete(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("oidc-idp-create")
	_, clientId, clientSecret, idpAlias, mappers := newRandomIdpTestData(ctx, "oidc-idp-create")
	idpRequest := makeOidcIdpRequest(idpAlias, clientId, clientSecret, discoveryUrl, []string{"openid"}, mappers)
	resp, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(idpRequest).Execute()
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.Equal(t, idpAlias, resp.Alias)
}

func TestOIDCIdp_Get(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("oidc-idp-get")
	_, clientId, clientSecret, idpAlias, mappers := newRandomIdpTestData(ctx, "oidc-idp-get")
	createOidcIdp(t, ctx, idpAlias, clientId, clientSecret, mappers)

	resp, err := OpenapiClient.IdpsAPI.GetIdp(ctx, idpAlias).Execute()
	require.NoError(t, err)
	require.NotNil(t, resp)
}

func TestOIDCIdp_Update(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("oidc-idp-update")
	_, clientId, clientSecret, idpAlias, mappers := newRandomIdpTestData(ctx, "oidc-idp-update")
	createOidcIdp(t, ctx, idpAlias, clientId, clientSecret, mappers)
	newSecret := fmt.Sprintf("secret-updated-%d", rand.Int())
	newScopes := []string{"openid", "profile"}
	updateRequest := makeOidcIdpRequest(idpAlias, clientId, newSecret, discoveryUrl, newScopes, mappers)

	_, err := OpenapiClient.IdpsAPI.UpdateIdp(ctx, idpAlias).IdpCreationRequest(updateRequest).Execute()
	require.NoError(t, err)
}

func TestOIDCIdp_List(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("oidc-idp-list")
	_, clientId, clientSecret, idpAlias, mappers := newRandomIdpTestData(ctx, "oidc-idp-list")
	createOidcIdp(t, ctx, idpAlias, clientId, clientSecret, mappers)

	listResp, httpResp, err := OpenapiClient.IdpsAPI.GetIdps(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	found := false
	for _, idp := range listResp {
		if idp.Alias == idpAlias {
			found = true
			break
		}
	}
	require.True(t, found, "Created IdP should be in the list")
}

func TestOIDCIdp_Get_NotFound(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("oidc-idp-get-notfound")
	alias := "non-existent-idp-alias"
	resp, err := OpenapiClient.IdpsAPI.GetIdp(ctx, alias).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestOIDCIdp_Delete_NotFound(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("oidc-idp-delete-notfound")
	alias := "non-existent-idp-alias"
	resp, err := OpenapiClient.IdpsAPI.DeleteIdp(ctx, alias).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestOIDCIdp_Update_Invalid(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("oidc-idp-invalid")
	_, clientId, clientSecret, idpAlias, mappers := newRandomIdpTestData(ctx, "oidc-idp-invalid")
	createOidcIdp(t, ctx, idpAlias, clientId, clientSecret, mappers)
	invalidRequest := makeOidcIdpRequest(idpAlias, "", clientSecret, discoveryUrl, []string{"openid"}, mappers)

	httpResp, err := OpenapiClient.IdpsAPI.UpdateIdp(ctx, idpAlias).IdpCreationRequest(invalidRequest).Execute()
	requireApiError(t, err, httpResp, http.StatusBadRequest)
}

func TestOIDCIdp_Create_Conflict(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("oidc-idp-conflict")
	_, clientId, clientSecret, idpAlias, mappers := newRandomIdpTestData(ctx, "oidc-idp-conflict")
	createOidcIdp(t, ctx, idpAlias, clientId, clientSecret, mappers)
	idpRequest := makeOidcIdpRequest(idpAlias, clientId, clientSecret, discoveryUrl, []string{"openid"}, mappers)

	_, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(idpRequest).Execute()
	requireApiError(t, err, httpResp, http.StatusConflict)
}

func TestOIDCIdp_Update_NotFound(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("oidc-idp-update-notfound")
	_, clientId, clientSecret, idpAlias, mappers := newRandomIdpTestData(ctx, "oidc-idp-update-notfound")
	idpRequest := makeOidcIdpRequest(idpAlias, clientId, clientSecret, discoveryUrl, []string{"openid"}, mappers)
	_, err := OpenapiClient.IdpsAPI.UpdateIdp(ctx, idpAlias).IdpCreationRequest(idpRequest).Execute()
	require.Error(t, err)
}

func TestOIDCIdp_Get_AfterDelete(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("oidc-idp-get-after-delete")
	_, clientId, clientSecret, idpAlias, mappers := newRandomIdpTestData(ctx, "oidc-idp-get-after-delete")
	createOidcIdp(t, ctx, idpAlias, clientId, clientSecret, mappers)
	deleteIDP(t, ctx, idpAlias)
	_, err := OpenapiClient.IdpsAPI.GetIdp(ctx, idpAlias).Execute()
	require.Error(t, err)
}

func TestOIDCIdp_Create_MissingFields(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	_, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(openapi.IdpCreationRequest{
		Name: StrPtr("missing-clientid"),
		Type: openapi.IDPCREATIONREQUESTTYPEENUM_OIDC,
		OidcData: *openapi.NewNullableOidcCreationData(&openapi.OidcCreationData{
			ClientId:            "",
			ClientSecret:        "secret",
			DiscoverDocumentUrl: discoveryUrl,
			Scopes:              []string{"openid"},
		}),
		Mappers: GetDefaultIDPMappers(),
	}).Execute()
	requireApiError(t, err, httpResp, http.StatusBadRequest)
}

func TestOIDCIdp_List_Empty(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("oidc-idp-list-empty")
	idps, _, err := OpenapiClient.IdpsAPI.GetIdps(ctx).Execute()
	require.NoError(t, err)
	require.Empty(t, idps)
}

func TestOIDCIdp_Get_Mapper(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("oidc-idp-get-mapper")
	_, clientId, clientSecret, idpAlias, mappers := newRandomIdpTestData(ctx, "oidc-idp-get-mapper")
	createOidcIdp(t, ctx, idpAlias, clientId, clientSecret, mappers)

	idps, _, err := OpenapiClient.IdpsAPI.GetIdps(ctx).Execute()
	require.NoError(t, err)
	var found *openapi.Idp
	for _, idp := range idps {
		if idp.Alias == idpAlias {
			found = &idp
			break
		}
	}
	require.NotNil(t, found)
	require.NotNil(t, found.Mappers)
	require.Equal(t, mappers.Gid, found.Mappers.Gid)
	require.Equal(t, mappers.Uid, found.Mappers.Uid)
	require.Equal(t, mappers.Groups, found.Mappers.Groups)
	require.Equal(t, mappers.SupplementaryGroups, found.Mappers.SupplementaryGroups)
	require.Equal(t, mappers.Email, found.Mappers.Email)
	require.Equal(t, mappers.FirstName, found.Mappers.FirstName)
	require.Equal(t, mappers.LastName, found.Mappers.LastName)
}

func TestOIDCIdp_UpdateIdpMappers_FirstLastName(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("oidc-idp-update-mappers-name")
	_, clientId, clientSecret, idpAlias, mappers := newRandomIdpTestData(ctx, "oidc-idp-update-mappers-name")
	createOidcIdp(t, ctx, idpAlias, clientId, clientSecret, mappers)
	defer deleteIDP(t, ctx, idpAlias)

	updatedMappers := openapi.Mappers{
		Gid:                 StrPtr("GID"),
		Uid:                 StrPtr("UID"),
		Groups:              StrPtr("GROUPS"),
		SupplementaryGroups: StrPtr("SUPPLEMENTARYGROUPS"),
		Email:               StrPtr("email"),
		FirstName:           StrPtr("given_name"),
		LastName:            StrPtr("family_name"),
	}
	_, err := OpenapiClient.IdpsAPI.UpdateIdpMappers(ctx, idpAlias).Mappers(updatedMappers).Execute()
	require.NoError(t, err)

	result, _, err := OpenapiClient.IdpsAPI.GetIdpMappers(ctx, idpAlias).Execute()
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, StrPtr("given_name"), result.FirstName)
	require.Equal(t, StrPtr("family_name"), result.LastName)
}

func TestOIDCIdp_Create_WrongType(t *testing.T) {
	t.Parallel()
	ctx, _, _ := TestStepsClient.NewTestContextWithRealm("oidc-idp-wrong-type")
	_, clientId, clientSecret, idpAlias, mappers := newRandomIdpTestData(ctx, "oidc-idp-wrong-type")
	// Attempt to create an IdP with type SAML instead of OIDC
	idpRequest := openapi.IdpCreationRequest{
		Name: StrPtr(idpAlias),
		Type: openapi.IDPCREATIONREQUESTTYPEENUM_SAML, // Wrong type
		OidcData: *openapi.NewNullableOidcCreationData(&openapi.OidcCreationData{
			ClientId:            clientId,
			ClientSecret:        clientSecret,
			DiscoverDocumentUrl: discoveryUrl,
			Scopes:              []string{"openid"},
		}),
		Mappers: mappers,
	}
	_, httpResp, err := OpenapiClient.IdpsAPI.CreateIdp(ctx).IdpCreationRequest(idpRequest).Execute()
	requireApiError(t, err, httpResp, http.StatusBadRequest)
}

func randomIdpData(prefix string) (string, string, string) {
	return fmt.Sprintf("client-%d", rand.Int()), fmt.Sprintf("secret-%d", rand.Int()), fmt.Sprintf("%s-%d", prefix, rand.Int())
}

func makeOidcIdpRequest(alias, clientId, clientSecret, discoverUrl string, scopes []string, mappers *openapi.Mappers) openapi.IdpCreationRequest {
	return openapi.IdpCreationRequest{
		Name: StrPtr(alias),
		Type: openapi.IDPCREATIONREQUESTTYPEENUM_OIDC,
		OidcData: *openapi.NewNullableOidcCreationData(&openapi.OidcCreationData{
			ClientId:            clientId,
			ClientSecret:        clientSecret,
			DiscoverDocumentUrl: discoverUrl,
			Scopes:              scopes,
		}),
		Mappers: mappers,
	}
}

func newRandomIdpTestData(ctx context.Context, prefix string) (context.Context, string, string, string, *openapi.Mappers) {
	clientId, clientSecret, idpAlias := randomIdpData(prefix)
	mappers := GetDefaultIDPMappers()
	return ctx, clientId, clientSecret, idpAlias, mappers
}

func createOidcIdp(t *testing.T, ctx context.Context, idpAlias, clientId, clientSecret string, mappers *openapi.Mappers) {
	idpRequest := makeOidcIdpRequest(idpAlias, clientId, clientSecret, discoveryUrl, []string{"openid"}, mappers)
	createIDP(t, ctx, idpRequest)
}

func requireApiError(t *testing.T, err error, httpResp *http.Response, expectedStatus int) {
	require.Error(t, err)
	require.Equal(t, expectedStatus, httpResp.StatusCode)
}
