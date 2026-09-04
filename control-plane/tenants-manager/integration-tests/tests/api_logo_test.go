/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/require"
)

// logoTenantCounter provides unique numeric tenant IDs for logo tests so that
// parallel tests do not share the same logo state in BFF.
var logoTenantCounter atomic.Int64

func init() {
	logoTenantCounter.Store(10000)
}

func nextLogoTenantID() string {
	id := logoTenantCounter.Add(1)
	return strconvItoa(id)
}

func strconvItoa(i int64) string {
	return fmt.Sprintf("%d", i)
}

// newLogoTestContext returns a context with a numeric CF tenant ID and management
// tenant flag set — required for TM's logo auth check and BFF's system-scope check.
func newLogoTestContext(tenantID string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, TestUserName)
	ctx = context.WithValue(ctx, runaictx.CfTenantId, tenantID)
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "true")
	return ctx
}

func redCircleSVGDataURL() string {
	const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><circle cx="5" cy="5" r="5" fill="#ff0000"/></svg>`
	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))
}

// ---------------------------------------------------------------------------
// Upload logo
// ---------------------------------------------------------------------------

func TestUploadTenantLogo_HappyFlow(t *testing.T) {
	t.Parallel()
	tenantID := nextLogoTenantID()
	ctx := newLogoTestContext(tenantID)

	httpResp, err := OpenapiClient.LogoAPI.
		UploadTenantLogo(ctx).
		LogoCreationRequest(*openapi.NewLogoCreationRequest(redCircleSVGDataURL())).
		Execute()

	require.NoError(t, err)
	require.NotNil(t, httpResp)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
}

func TestUploadTenantLogo_InvalidBase64_Returns400(t *testing.T) {
	t.Parallel()
	ctx := newLogoTestContext(nextLogoTenantID())

	httpResp, err := OpenapiClient.LogoAPI.
		UploadTenantLogo(ctx).
		LogoCreationRequest(*openapi.NewLogoCreationRequest("not-valid-base64!!!")).
		Execute()

	require.Error(t, err)
	require.NotNil(t, httpResp)
	require.Equal(t, http.StatusBadRequest, httpResp.StatusCode)
}

func TestUploadTenantLogo_UnsupportedMimeType_Returns415(t *testing.T) {
	t.Parallel()
	ctx := newLogoTestContext(nextLogoTenantID())

	// Minimal JPEG data URL — valid base64 but unsupported MIME type.
	jpegDataURL := "data:image/jpeg;base64,/9j/"

	httpResp, err := OpenapiClient.LogoAPI.
		UploadTenantLogo(ctx).
		LogoCreationRequest(*openapi.NewLogoCreationRequest(jpegDataURL)).
		Execute()

	require.Error(t, err)
	require.NotNil(t, httpResp)
	require.Equal(t, http.StatusUnsupportedMediaType, httpResp.StatusCode)
}

// ---------------------------------------------------------------------------
// Get logo
// ---------------------------------------------------------------------------

func TestGetTenantLogo_AfterUpload_ReturnsLogo(t *testing.T) {
	t.Parallel()
	tenantID := nextLogoTenantID()
	ctx := newLogoTestContext(tenantID)

	logoDataURL := redCircleSVGDataURL()

	httpResp, err := OpenapiClient.LogoAPI.
		UploadTenantLogo(ctx).
		LogoCreationRequest(*openapi.NewLogoCreationRequest(logoDataURL)).
		Execute()
	require.NoError(t, err)
	require.NotNil(t, httpResp)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)

	result, httpResp, err := OpenapiClient.LogoAPI.GetTenantLogo(ctx).Execute()
	require.NoError(t, err)
	require.NotNil(t, httpResp)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
	require.NotNil(t, result)
	require.Equal(t, logoDataURL, result.GetLogo())
}

func TestUploadTenantLogo_PNG_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newLogoTestContext(nextLogoTenantID())

	// Minimal 1×1 transparent PNG — valid PNG, supported MIME type.
	pngDataURL := "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="

	httpResp, err := OpenapiClient.LogoAPI.
		UploadTenantLogo(ctx).
		LogoCreationRequest(*openapi.NewLogoCreationRequest(pngDataURL)).
		Execute()

	require.NoError(t, err)
	require.NotNil(t, httpResp)
	require.Equal(t, http.StatusOK, httpResp.StatusCode)
}

func TestGetTenantLogo_MutationIsVisible(t *testing.T) {
	t.Parallel()
	tenantID := nextLogoTenantID()
	ctx := newLogoTestContext(tenantID)

	initialLogo := redCircleSVGDataURL()
	_, err := OpenapiClient.LogoAPI.
		UploadTenantLogo(ctx).
		LogoCreationRequest(*openapi.NewLogoCreationRequest(initialLogo)).
		Execute()
	require.NoError(t, err)

	firstGet, _, err := OpenapiClient.LogoAPI.GetTenantLogo(ctx).Execute()
	require.NoError(t, err)
	require.NotNil(t, firstGet)
	previousValue := firstGet.GetLogo()

	// Upload a different logo.
	const updatedSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><circle cx="5" cy="5" r="5" fill="#0000ff"/></svg>`
	updatedLogoDataURL := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(updatedSVG))

	_, err = OpenapiClient.LogoAPI.
		UploadTenantLogo(ctx).
		LogoCreationRequest(*openapi.NewLogoCreationRequest(updatedLogoDataURL)).
		Execute()
	require.NoError(t, err)

	secondGet, _, err := OpenapiClient.LogoAPI.GetTenantLogo(ctx).Execute()
	require.NoError(t, err)
	require.NotNil(t, secondGet)
	require.NotEqual(t, previousValue, secondGet.GetLogo(), "logo should reflect the update")
}
