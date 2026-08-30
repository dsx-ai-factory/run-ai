package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	openapi "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	realm_client "github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/client"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

func TestAccessKey_CreateAccessKey_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	name := GenerateRandomString(accessKeyNamePrefix)
	accessKeyRequest := openapi.AccessKeyCreationRequest{
		Name: name,
	}
	resp, httpResp, err := OpenapiClient.AccessKeysAPI.CreateAccessKey(ctx).AccessKeyCreationRequest(accessKeyRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, httpResp.StatusCode)
	require.NotNil(t, resp)
	defer clientCleanup(t, ctx, resp.Id)

	require.NotNil(t, resp.Id)
	require.NotNil(t, resp.ClientSecret)
	require.NotNil(t, resp.ClientId)
	require.NotNil(t, resp.CreatedAt)
	require.Equal(t, name, resp.Name)
	require.NotEqual(t, resp.ClientId, resp.Name)

	newClient, err := IamServices.ClientService.GetClient(ctx, BaseRealmName, resp.Id)
	require.NoError(t, err)
	require.NotNil(t, newClient)
	require.Nil(t, newClient.Attributes.IsExternal)
	require.NotNil(t, newClient.Attributes.UpdatedAt)
	require.NotNil(t, newClient.Attributes.CreatedAt)
	require.Nil(t, newClient.Attributes.LastLogin)
	require.Equal(t, 30*time.Minute, *newClient.Attributes.TokenLifespan)
	require.Equal(t, realm_client.ClientTypeUserApp, *newClient.Attributes.ClientType)
	require.Equal(t, createdByUser, *newClient.Attributes.CreatedBy)
}

func TestAccessKey_Token_NoEmailVerifiedClaim(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	createdByUser := GenerateUsername()
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, createdByUser)

	name := GenerateRandomString(accessKeyNamePrefix)
	resp, httpResp, err := OpenapiClient.AccessKeysAPI.CreateAccessKey(ctx).AccessKeyCreationRequest(openapi.AccessKeyCreationRequest{Name: name}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, httpResp.StatusCode)
	defer clientCleanup(t, ctx, resp.Id)

	ctx2 := newTestContext()
	tokenCtx := TestStepsClient.SetFakeHostHeader(ctx2, OpenapiClient, BaseRealmName)
	tokenResp, _, err := OpenapiClient.TokensAPI.GrantToken(tokenCtx).TokenRequest(openapi.TokenRequest{
		GrantType:    ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_CLIENT_CREDENTIALS),
		ClientID:     ptr.To(resp.ClientId),
		ClientSecret: ptr.To(resp.ClientSecret),
	}).Execute()
	require.NoError(t, err)
	require.NotNil(t, tokenResp.AccessToken)

	parser := jwt.NewParser()
	parsedToken, _, err := parser.ParseUnverified(*tokenResp.AccessToken, &jwt.MapClaims{})
	require.NoError(t, err)

	claims, ok := parsedToken.Claims.(*jwt.MapClaims)
	require.True(t, ok)
	_, hasEmailVerified := (*claims)["email_verified"]
	require.False(t, hasEmailVerified, "access key token must not contain email_verified claim")
}

func TestAccessKey_CreateAccessKey_InvalidName(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	accessKeyRequest := openapi.AccessKeyCreationRequest{
		Name: "",
	}
	resp, httpResp, err := OpenapiClient.AccessKeysAPI.CreateAccessKey(ctx).AccessKeyCreationRequest(accessKeyRequest).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestAccessKey_CreateAccessKey_TooManyAccessKeys(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	var accessKeyIds []string
	for index := 0; index < 20; index++ {
		accessKeyName := fmt.Sprintf("access-key-%d", index)
		accessKeyId, err := TestStepsClient.CreateAccessKey(ctx, BaseRealmName, BaseTenantId, ptr.To(accessKeyName), nil)
		if err == nil {
			accessKeyIds = append(accessKeyIds, accessKeyId)
			defer clientCleanup(t, ctx, accessKeyId)
		}
	}
	name := GenerateRandomString(accessKeyNamePrefix)
	accessKeyRequest := openapi.AccessKeyCreationRequest{
		Name: name,
	}
	resp, httpResp, err := OpenapiClient.AccessKeysAPI.CreateAccessKey(ctx).AccessKeyCreationRequest(accessKeyRequest).Execute()

	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
	require.Nil(t, resp)
	if openapiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
		errMsg := string(openapiErr.Body())
		require.Contains(t, errMsg, "access key limit exceeded. You can only create up to 20 access keys")
	}
}

func TestAccessKey_CreateInvalidName_LongOrSymbols(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	resp, httpResp, err := OpenapiClient.AccessKeysAPI.CreateAccessKey(ctx).AccessKeyCreationRequest(openapi.AccessKeyCreationRequest{
		Name: "sdfsdfsdfsdfsdfsdfsdfdsf!@@$@$454364ehfhghjfgjfjfgjfjfgjfgjfjfjfgjgfbdfgbcdb",
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestAccessKey_CreateInvalidName_CapitalLetters(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	resp, httpResp, err := OpenapiClient.AccessKeysAPI.CreateAccessKey(ctx).AccessKeyCreationRequest(openapi.AccessKeyCreationRequest{
		Name: "AAAA",
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestAccessKey_CreateInvalidName_StartsWithNumber(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	resp, httpResp, err := OpenapiClient.AccessKeysAPI.CreateAccessKey(ctx).AccessKeyCreationRequest(openapi.AccessKeyCreationRequest{
		Name: "2accesskey",
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestAccessKey_CreateInvalidName_XSS(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	resp, httpResp, err := OpenapiClient.AccessKeysAPI.CreateAccessKey(ctx).AccessKeyCreationRequest(openapi.AccessKeyCreationRequest{
		Name: "<script>alert(123)</script>",
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestAccessKey_SameNameDifferentUsers(t *testing.T) {
	t.Parallel()
	name := GenerateRandomString(accessKeyNamePrefix)

	user1Ctx := TestStepsClient.SetContextValuesWithUser(newTestContext(), BaseTenantId, GenerateUsername())
	user2Ctx := TestStepsClient.SetContextValuesWithUser(newTestContext(), BaseTenantId, GenerateUsername())

	resp1, httpResp1, err := OpenapiClient.AccessKeysAPI.CreateAccessKey(user1Ctx).AccessKeyCreationRequest(openapi.AccessKeyCreationRequest{Name: name}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, httpResp1.StatusCode)
	defer clientCleanup(t, newTestContext(), resp1.Id)

	resp2, httpResp2, err := OpenapiClient.AccessKeysAPI.CreateAccessKey(user2Ctx).AccessKeyCreationRequest(openapi.AccessKeyCreationRequest{Name: name}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, httpResp2.StatusCode)
	defer clientCleanup(t, newTestContext(), resp2.Id)

	require.Equal(t, name, resp1.Name)
	require.Equal(t, name, resp2.Name)
	require.NotEqual(t, resp1.Id, resp2.Id)
}

func TestAccessKey_Delete_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	accessKeyId, err := TestStepsClient.CreateAccessKey(ctx, BaseRealmName, BaseTenantId, nil, nil)
	require.NoError(t, err)
	httpResp, err := OpenapiClient.AccessKeysAPI.DeleteAccessKeyById(ctx, accessKeyId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, httpResp.StatusCode)
}

func TestAccessKey_Delete_InvalidID(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	httpResp, err := OpenapiClient.AccessKeysAPI.DeleteAccessKeyById(ctx, "invalid").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
}

func TestAccessKey_Delete_VerifyGone(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	accessKeyId, err := TestStepsClient.CreateAccessKey(ctx, BaseRealmName, BaseTenantId, nil, nil)
	require.NoError(t, err)

	httpResp, err := OpenapiClient.AccessKeysAPI.DeleteAccessKeyById(ctx, accessKeyId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, httpResp.StatusCode)

	_, httpResp, err = OpenapiClient.AccessKeysAPI.GetAccessKeyById(ctx, accessKeyId).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
}

func TestAccessKey_DoubleDelete(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	accessKeyId, err := TestStepsClient.CreateAccessKey(ctx, BaseRealmName, BaseTenantId, nil, nil)
	require.NoError(t, err)

	httpResp, err := OpenapiClient.AccessKeysAPI.DeleteAccessKeyById(ctx, accessKeyId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, httpResp.StatusCode)

	httpResp, err = OpenapiClient.AccessKeysAPI.DeleteAccessKeyById(ctx, accessKeyId).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
}

func TestAccessKey_GetAccessKeyById_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	name := GenerateRandomString(accessKeyNamePrefix)
	accessKeyId, err := TestStepsClient.CreateAccessKey(ctx, BaseRealmName, BaseTenantId, ptr.To(name), nil)
	require.NoError(t, err)
	defer clientCleanup(t, ctx, accessKeyId)

	resp, httpResp, err := OpenapiClient.AccessKeysAPI.GetAccessKeyById(ctx, accessKeyId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, resp)

	require.Equal(t, accessKeyId, resp.Id)
	require.Equal(t, name, resp.Name)
	require.NotEmpty(t, resp.ClientId)
	require.NotNil(t, resp.CreatedAt.Get())
	require.Nil(t, resp.LastLogin.Get())
}

func TestAccessKey_GetAccessKeyById_InvalidID(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	resp, httpResp, err := OpenapiClient.AccessKeysAPI.GetAccessKeyById(ctx, "").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestAccessKey_GetById_OtherUserKey(t *testing.T) {
	t.Parallel()
	adminCtx := newTestContext()
	accessKeyId, err := TestStepsClient.CreateAccessKey(adminCtx, BaseRealmName, BaseTenantId, nil, nil)
	require.NoError(t, err)
	defer clientCleanup(t, adminCtx, accessKeyId)

	otherCtx := TestStepsClient.SetContextValuesWithUser(newTestContext(), BaseTenantId, GenerateUsername())
	resp, httpResp, err := OpenapiClient.AccessKeysAPI.GetAccessKeyById(otherCtx, accessKeyId).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestAccessKey_GetById_NoRoles_OtherUserKey(t *testing.T) {
	t.Parallel()
	adminCtx := newTestContext()
	accessKeyId, err := TestStepsClient.CreateAccessKey(adminCtx, BaseRealmName, BaseTenantId, nil, nil)
	require.NoError(t, err)
	defer clientCleanup(t, adminCtx, accessKeyId)

	noRolesCtx := TestStepsClient.SetContextValuesWithUser(newTestContext(), BaseTenantId, GenerateUsername())
	resp, httpResp, err := OpenapiClient.AccessKeysAPI.GetAccessKeyById(noRolesCtx, accessKeyId).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestAccessKey_RegenerateAccessKeySecret_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	accessKeyId, err := TestStepsClient.CreateAccessKey(ctx, BaseRealmName, BaseTenantId, nil, nil)
	require.NoError(t, err)
	defer clientCleanup(t, ctx, accessKeyId)
	accessKey, err := IamServices.ClientService.GetClient(ctx, BaseRealmName, accessKeyId)
	require.NoError(t, err)
	require.NotNil(t, accessKey.Secret)

	resp, httpResp, err := OpenapiClient.AccessKeysAPI.RegenerateAccessKeySecret(ctx, accessKeyId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, resp)
	require.NotNil(t, resp.ClientSecret)

	require.NotEqual(t, accessKey.Secret, resp.ClientSecret)
}

func TestAccessKey_RegenerateAccessKeySecret_InvalidID(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	resp, httpResp, err := OpenapiClient.AccessKeysAPI.RegenerateAccessKeySecret(ctx, "").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestAccessKey_RegenerateSecret_NonExisting(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	resp, httpResp, err := OpenapiClient.AccessKeysAPI.RegenerateAccessKeySecret(ctx, "non-existent-access-key-id-12345").Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestAccessKey_RegenerateSecret_OtherUserKey(t *testing.T) {
	t.Parallel()
	adminCtx := newTestContext()
	accessKeyId, err := TestStepsClient.CreateAccessKey(adminCtx, BaseRealmName, BaseTenantId, nil, nil)
	require.NoError(t, err)
	defer clientCleanup(t, adminCtx, accessKeyId)

	otherCtx := TestStepsClient.SetContextValuesWithUser(newTestContext(), BaseTenantId, GenerateUsername())
	resp, httpResp, err := OpenapiClient.AccessKeysAPI.RegenerateAccessKeySecret(otherCtx, accessKeyId).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestAccessKey_RegenerateSecret_NoRoles_OtherUserKey(t *testing.T) {
	t.Parallel()
	adminCtx := newTestContext()
	accessKeyId, err := TestStepsClient.CreateAccessKey(adminCtx, BaseRealmName, BaseTenantId, nil, nil)
	require.NoError(t, err)
	defer clientCleanup(t, adminCtx, accessKeyId)

	noRolesCtx := TestStepsClient.SetContextValuesWithUser(newTestContext(), BaseTenantId, GenerateUsername())
	resp, httpResp, err := OpenapiClient.AccessKeysAPI.RegenerateAccessKeySecret(noRolesCtx, accessKeyId).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestAccessKey_GetAccessKeys_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	createdByOtherUser := GenerateUsername()
	accessKey1Name := GenerateRandomString("accessKey1")
	accessKey2Name := GenerateRandomString("accessKey2")
	accessKey3Name := GenerateRandomString("accessKey3")
	createdByUser := GenerateUsername()
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, createdByUser)
	accessKeyId1, err := TestStepsClient.CreateAccessKey(ctx, BaseRealmName, BaseTenantId, ptr.To(accessKey1Name), ptr.To(createdByOtherUser))
	require.NoError(t, err)
	defer clientCleanup(t, ctx, accessKeyId1)
	accessKeyId2, err := TestStepsClient.CreateAccessKey(ctx, BaseRealmName, BaseTenantId, ptr.To(accessKey2Name), ptr.To(createdByUser))
	require.NoError(t, err)
	defer clientCleanup(t, ctx, accessKeyId2)
	accessKeyId3, err := TestStepsClient.CreateAccessKey(ctx, BaseRealmName, BaseTenantId, ptr.To(accessKey3Name), ptr.To(createdByUser))
	require.NoError(t, err)
	defer clientCleanup(t, ctx, accessKeyId3)

	resp, httpResp, err := OpenapiClient.AccessKeysAPI.GetAccessKeys(ctx).Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, resp)
	require.Len(t, resp.AccessKeys, 2)

	ids := make(map[string]bool, len(resp.AccessKeys))
	for _, ak := range resp.AccessKeys {
		ids[ak.Id] = true
	}
	require.True(t, ids[accessKeyId2], "created access key %s should appear in list", accessKey2Name)
	require.True(t, ids[accessKeyId3], "created access key %s should appear in list", accessKey3Name)

	for _, ak := range resp.AccessKeys {
		if ak.Id == accessKeyId2 || ak.Id == accessKeyId3 {
			require.Nil(t, ak.LastLogin.Get(), "newly created access key should have nil lastLogin")
		}
	}
}

func TestAccessKey_GetAccessKeysAdministration_CreatedByFilter_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	createdBy := GenerateUsername()
	createdBy2 := GenerateUsername()
	accessKey1Name := GenerateRandomString("accessKey1")
	accessKey2Name := GenerateRandomString("accessKey2")
	accessKeyId1, err := TestStepsClient.CreateAccessKey(ctx, BaseRealmName, BaseTenantId, ptr.To(accessKey1Name), ptr.To(createdBy))
	require.NoError(t, err)
	defer clientCleanup(t, ctx, accessKeyId1)
	accessKeyId2, err := TestStepsClient.CreateAccessKey(ctx, BaseRealmName, BaseTenantId, ptr.To(accessKey2Name), ptr.To(createdBy2))
	require.NoError(t, err)
	defer clientCleanup(t, ctx, accessKeyId2)

	resp, httpResp, err := OpenapiClient.AccessKeysAPI.GetAccessKeysAdministration(ctx).CreatedBy(createdBy).Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, resp)
	require.Len(t, resp.AccessKeys, 1)
	require.Equal(t, createdBy, resp.AccessKeys[0].CreatedBy)
}

func TestAccessKey_GetAccessKeysAdministration_ClientIdFilter_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	clientID := GenerateRandomString("clientID")
	accessKeyName := GenerateRandomString("accessKeyName")
	createdBy := GenerateUsername()
	accessKeyId1, err := TestStepsClient.CreateAccessKey(ctx, BaseRealmName, BaseTenantId, ptr.To(accessKeyName), nil)
	require.NoError(t, err)
	defer clientCleanup(t, ctx, accessKeyId1)

	accessKeyId2, err := IamServices.ClientService.CreateClient(ctx, BaseRealmName, &models.Client{
		Name:     ptr.To(accessKeyName),
		ClientID: ptr.To(clientID),
		Enabled:  ptr.To(true),
		TenantId: ptr.To(BaseTenantId),
		Attributes: &models.ClientAttributes{
			CreatedBy:  ptr.To(createdBy),
			ClientType: ptr.To(realm_client.ClientTypeUserApp),
		},
	})
	require.NoError(t, err)
	defer clientCleanup(t, ctx, accessKeyId2)

	resp, httpResp, err := OpenapiClient.AccessKeysAPI.GetAccessKeysAdministration(ctx).ClientId(clientID).Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, resp)
	require.Len(t, resp.AccessKeys, 1)
	require.Equal(t, clientID, resp.AccessKeys[0].ClientId)
}

func TestAccessKey_GetAccessKeysAdministration_ClientIdFilter_NotFound(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	resp, httpResp, err := OpenapiClient.AccessKeysAPI.GetAccessKeysAdministration(ctx).ClientId("nonexistent-client-id").Execute()

	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestAccessKey_GetAccessKeysAdministration_MoreThan1Filter(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	resp, httpResp, err := OpenapiClient.AccessKeysAPI.GetAccessKeysAdministration(ctx).ClientId("clientID").CreatedBy("createdBy").Execute()

	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
	require.Nil(t, resp)
}

func TestAccessKey_DeleteAccessKeyByIdAdministration_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	accessKeyId1, err := TestStepsClient.CreateAccessKey(ctx, BaseRealmName, BaseTenantId, nil, nil)
	require.NoError(t, err)
	httpResp, err := OpenapiClient.AccessKeysAPI.DeleteAccessKeyByIdAdministration(ctx, accessKeyId1).Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, httpResp.StatusCode)
}

func TestAccessKey_DeleteAccessKeyByIdAdministration_InvalidID(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	httpResp, err := OpenapiClient.AccessKeysAPI.DeleteAccessKeyByIdAdministration(ctx, "").Execute()

	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
}

func TestAccessKey_DeleteAccessKeyByIdAdministration_NoAccessKey(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	accessKeyId1, err := TestStepsClient.CreateServiceAccount(ctx, BaseRealmName, BaseTenantId, nil)
	require.NoError(t, err)

	httpResp, err := OpenapiClient.AccessKeysAPI.DeleteAccessKeyByIdAdministration(ctx, accessKeyId1).Execute()

	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, httpResp.StatusCode)
}

func clientCleanup(t *testing.T, ctx context.Context, accessKeyId string) {
	if accessKeyId == "" {
		return
	}
	err := TestStepsClient.DeleteClient(ctx, BaseRealmName, accessKeyId)
	require.NoError(t, err)
}
