//go:build integration
// +build integration


/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"net/http"
	"testing"

	openapi "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

func TestCreateEmail_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)

	conf := openapi.NewEmailChannelConfigWithDefaults()
	conf.SetUseAttachments(true)
	conf.SetUsername("user")
	conf.SetPassword("pass")
	conf.SetSmtpHost("host")
	conf.SetSmtpPort(123)
	conf.SetFrom("from")

	_, response, err := client.EmailAPI.CreateEmail(ctx).EmailChannel(openapi.EmailChannel{
		Config: *conf,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	defer deleteEmail(t, ctx)

	email, _, err := client.EmailAPI.GetEmail(ctx).Execute()
	require.NoError(t, err)
	require.NotNil(t, email)
	require.True(t, *email.Config.UseAttachments.Get())
	require.Equal(t, "user", *email.Config.Username)
	require.Nil(t, email.Config.Password) // Password is not returned
	require.Equal(t, "host", email.Config.SmtpHost)
	require.Equal(t, int32(123), *email.Config.SmtpPort)
	require.Equal(t, "from", *email.Config.From)
}

func TestCreateEmail_Conflict(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)

	conf := openapi.NewEmailChannelConfigWithDefaults()
	conf.SetUsername("user")
	conf.SetPassword("pass")
	conf.SetSmtpHost("host")
	conf.SetSmtpPort(587)
	conf.SetFrom("from")

	_, response, err := client.EmailAPI.CreateEmail(ctx).EmailChannel(openapi.EmailChannel{
		Config: *conf,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	defer deleteEmail(t, ctx)

	_, conflictResponse, err := client.EmailAPI.CreateEmail(ctx).EmailChannel(openapi.EmailChannel{
		Config: *conf,
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusConflict, conflictResponse.StatusCode)
}

func TestCreateEmail_NoSystemScope(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithoutSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithoutSystemScope)

	_, createResponse, err := client.EmailAPI.CreateEmail(ctx).EmailChannel(openapi.EmailChannel{
		Config: *openapi.NewEmailChannelConfigWithDefaults(),
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, createResponse.StatusCode)
}

func TestPatchEmail(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)

	conf := openapi.NewEmailChannelConfigWithDefaults()
	conf.SetUsername("user")
	conf.SetPassword("pass")
	conf.SetSmtpHost("host")
	conf.SetSmtpPort(123)
	conf.SetFrom("from")

	_, response, err := client.EmailAPI.CreateEmail(ctx).EmailChannel(openapi.EmailChannel{
		Config: *conf,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	defer deleteEmail(t, ctx)

	patchConf := openapi.EmailChannelConfigForPatch{
		Username:        ptr.To("patched-user"),
		FromDisplayName: ptr.To("patched-display-name"),
	}

	email, patchResponse, err := client.EmailAPI.PatchEmail(ctx).EmailChannelForPatch(openapi.EmailChannelForPatch{
		Config: patchConf,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, patchResponse.StatusCode)

	require.NotNil(t, email)
	require.Equal(t, "patched-user", *email.Config.Username)
	require.Equal(t, "patched-display-name", *email.Config.FromDisplayName)
	require.Equal(t, "host", email.Config.SmtpHost)
	require.Equal(t, int32(123), *email.Config.SmtpPort)
	require.Equal(t, "from", *email.Config.From)
}

func TestUpdateEmail_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)

	conf := openapi.NewEmailChannelConfigWithDefaults()
	conf.SetSmtpHost("host")
	conf.SetSmtpPort(123)
	conf.SetFrom("from")

	_, response, err := client.EmailAPI.CreateEmail(ctx).EmailChannel(openapi.EmailChannel{
		Config: *conf,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	defer deleteEmail(t, ctx)

	updateConf := openapi.NewEmailChannelConfigWithDefaults()
	updateConf.SetSmtpHost("new-host")

	email, updateResponse, err := client.EmailAPI.UpdateEmail(ctx).EmailChannel(openapi.EmailChannel{
		Config: *updateConf,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, "new-host", email.Config.SmtpHost)
	require.Equal(t, int32(587), *email.Config.SmtpPort) // default port is 587
	require.Equal(t, http.StatusOK, updateResponse.StatusCode)
	require.Nil(t, email.Config.Username)
	require.Nil(t, email.Config.From)
}

func TestUpdateEmail_NoSystemScope(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithoutSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithoutSystemScope)

	_, updateResponse, err := client.EmailAPI.UpdateEmail(ctx).EmailChannel(openapi.EmailChannel{
		Config: *openapi.NewEmailChannelConfigWithDefaults(),
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, updateResponse.StatusCode)
}

func TestDeleteEmail_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)

	conf := openapi.NewEmailChannelConfigWithDefaults()
	conf.SetUsername("user")
	conf.SetPassword("pass")
	conf.SetSmtpHost("host")
	conf.SetSmtpPort(587)
	conf.SetFrom("from")

	_, response, err := client.EmailAPI.CreateEmail(ctx).EmailChannel(openapi.EmailChannel{
		Config: *conf,
	}).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)

	deleteResponse, err := client.EmailAPI.DeleteEmail(ctx).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, deleteResponse.StatusCode)

	_, getResponse, err := client.EmailAPI.GetEmail(ctx).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, getResponse.StatusCode)
}

func TestDeleteEmail_NotFound(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)

	deleteResponse, err := client.EmailAPI.DeleteEmail(ctx).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, deleteResponse.StatusCode)
}

func TestGetEmail_NoSystemScope(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithoutSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithoutSystemScope)

	_, getResponse, err := client.EmailAPI.GetEmail(ctx).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, getResponse.StatusCode)
}

func TestPatchEmail_NoSystemScope(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithoutSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithoutSystemScope)

	_, patchResponse, err := client.EmailAPI.PatchEmail(ctx).EmailChannelForPatch(openapi.EmailChannelForPatch{
		Config: *openapi.NewEmailChannelConfigForPatch(),
	}).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, patchResponse.StatusCode)
}

func TestDeleteEmail_NoSystemScope(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithoutSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithoutSystemScope)

	deleteResponse, err := client.EmailAPI.DeleteEmail(ctx).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusForbidden, deleteResponse.StatusCode)
}

func TestValidateEmail(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, tenantIdWithSystemScope)
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, tenantIdWithSystemScope)

	conf := &openapi.EmailChannelConfig{
		Username:       ptr.To("user"),
		Password:       ptr.To("password"),
		SmtpTlsEnabled: *openapi.NewNullableBool(ptr.To(false)),
		SmtpHost:       SmtpHost,
		SmtpPort:       ptr.To(int32(SmtpPort)),
		AuthType:       openapi.EMAILAUTHTYPE_AUTH_LOGIN.Ptr(),
	}

	resp, response, err := client.EmailAPI.ValidateEmail(ctx).EmailChannelConfig(*conf).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, response.StatusCode)
	require.NotNil(t, resp)
	require.True(t, resp.GetStatus())
}

func deleteEmail(t *testing.T, ctx context.Context) {
	response, err := client.EmailAPI.DeleteEmail(ctx).Execute()
	if err != nil && response.StatusCode != http.StatusNotFound {
		require.NoError(t, err)
	}
}
