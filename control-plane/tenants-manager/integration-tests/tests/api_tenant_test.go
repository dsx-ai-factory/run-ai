/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"io"
	"net/http"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/require"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/generated"
)

// ---------------------------------------------------------------------------
// Get Tenants
// ---------------------------------------------------------------------------

func TestGetTenants_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	created := createAndCleanupTenant(t, ctx)

	tenants, resp, err := OpenapiClient.TenantAPI.GetTenants(newTestContext()).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotEmpty(t, tenants)

	found := false
	for _, tenant := range tenants {
		if tenant.Id == created.Tenant.Id {
			require.Equal(t, created.Tenant.Name, tenant.Name)
			found = true
			break
		}
	}
	require.True(t, found, "created tenant not found in list")
}

func TestGetTenants_FilterByName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	created := createAndCleanupTenant(t, ctx)

	tenants, resp, err := OpenapiClient.TenantAPI.GetTenants(newTestContext()).
		Name(created.Tenant.Name).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, tenants, 1)
	require.Equal(t, created.Tenant.Id, tenants[0].Id)
	require.Equal(t, created.Tenant.Name, tenants[0].Name)
}

func TestGetTenants_FilterByAlias(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	created := createAndCleanupTenant(t, ctx)

	alias := GenerateRandomString("alias-")
	updateReq := openapi.TenantUpdateRequest{}
	updateReq.SetAlias(alias)
	_, resp, err := OpenapiClient.TenantAPI.UpdateTenantById(newTestContext(), created.Tenant.Id).
		TenantUpdateRequest(updateReq).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	tenants, resp, err := OpenapiClient.TenantAPI.GetTenants(newTestContext()).
		Alias(alias).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, tenants, 1)
	require.Equal(t, created.Tenant.Id, tenants[0].Id)
}

func TestGetTenants_FilterByName_NotFound(t *testing.T) {
	t.Parallel()

	tenants, resp, err := OpenapiClient.TenantAPI.GetTenants(newTestContext()).
		Name("nonexistent-tenant-name").Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Empty(t, tenants)
}

// ---------------------------------------------------------------------------
// Delete Tenant By ID
// ---------------------------------------------------------------------------

func TestDeleteTenantById_SoftDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	name := GenerateRandomTenantName()
	email := GenerateRandomEmail(name)
	created, err := TestStepsClient.CreateTenantInternal(ctx, name, email)
	require.NoError(t, err)

	deleteResp, resp, err := OpenapiClient.TenantAPI.DeleteTenantById(newTestContext(), created.Tenant.Id).
		IsHardDelete(false).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, deleteResp)
	require.NotEmpty(t, deleteResp.Uid)

	_, getResp, getErr := OpenapiClient.TenantAPI.GetTenantById(newTestContext(), created.Tenant.Id).Execute()
	require.Error(t, getErr)
	require.Equal(t, http.StatusNotFound, getResp.StatusCode)
}

func TestDeleteTenantById_HardDelete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tenantName := GenerateRandomTenantName()
	email := GenerateRandomEmail(tenantName)
	created, err := TestStepsClient.CreateTenantInternal(ctx, tenantName, email)
	require.NoError(t, err)

	deleteResp, resp, err := OpenapiClient.TenantAPI.DeleteTenantById(newTestContext(), created.Tenant.Id).
		IsHardDelete(true).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, deleteResp)
	require.NotEmpty(t, deleteResp.Uid)

	_, getResp, getErr := OpenapiClient.TenantAPI.GetTenantById(newTestContext(), created.Tenant.Id).Execute()
	require.Error(t, getErr)
	require.Equal(t, http.StatusNotFound, getResp.StatusCode)
}

func TestDeleteTenantById_NotFound(t *testing.T) {
	t.Parallel()

	_, resp, err := OpenapiClient.TenantAPI.DeleteTenantById(newTestContext(), 999999).
		IsHardDelete(false).Execute()
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// Get Tenant by ID
// ---------------------------------------------------------------------------

func TestGetTenantById_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	created := createAndCleanupTenant(t, ctx)

	tenant, resp, err := OpenapiClient.TenantAPI.GetTenantById(newTestContext(), created.Tenant.Id).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, tenant)

	require.Equal(t, created.Tenant.Id, tenant.Id)
	require.Equal(t, created.Tenant.Name, tenant.Name)
	require.Equal(t, created.Tenant.Uuid, tenant.Uuid)
}

// ---------------------------------------------------------------------------
// Create Tenant
// ---------------------------------------------------------------------------

func TestCreateTenant_HappyFlow(t *testing.T) {
	t.Parallel()

	name := GenerateRandomTenantName()
	email := GenerateRandomEmail(name)
	req := openapi.TenantCreationRequest{
		Name:  name,
		Email: email,
	}

	created, resp, err := OpenapiClient.TenantAPI.CreateTenant(newTestContext()).
		TenantCreationRequest(req).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	require.NotNil(t, created)
	t.Cleanup(func() {
		_ = TestStepsClient.HardDeleteTenantInternal(context.Background(), created.Tenant.Id)
	})

	tenant := created.Tenant
	require.Equal(t, name, tenant.Name)
	require.NotZero(t, tenant.Id)
	require.NotEmpty(t, tenant.Uuid)
	require.NotZero(t, tenant.CreatedAt)
	require.NotZero(t, tenant.UpdatedAt)
}

func TestCreateTenant_DuplicateName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	created := createAndCleanupTenant(t, ctx)

	req := openapi.TenantCreationRequest{
		Name:  created.Tenant.Name,
		Email: GenerateRandomEmail(created.Tenant.Name),
	}
	_, resp, err := OpenapiClient.TenantAPI.CreateTenant(newTestContext()).
		TenantCreationRequest(req).Execute()
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestCreateTenant_SameNameAndAlias(t *testing.T) {
	t.Parallel()

	name := GenerateRandomTenantName()
	email := GenerateRandomEmail(name)
	req := openapi.TenantCreationRequest{
		Name:  name,
		Email: email,
	}
	req.SetAlias(name)

	_, resp, err := OpenapiClient.TenantAPI.CreateTenant(newTestContext()).
		TenantCreationRequest(req).Execute()
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateTenant_DuplicateAlias(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tenant1 := createAndCleanupTenant(t, ctx)

	alias := GenerateRandomString("alias-")
	updateReq := openapi.TenantUpdateRequest{}
	updateReq.SetAlias(alias)
	_, resp, err := OpenapiClient.TenantAPI.UpdateTenantById(newTestContext(), tenant1.Tenant.Id).
		TenantUpdateRequest(updateReq).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	name2 := GenerateRandomTenantName()
	createReq := openapi.TenantCreationRequest{
		Name:  name2,
		Email: GenerateRandomEmail(name2),
	}
	createReq.SetAlias(alias)

	_, resp, err = OpenapiClient.TenantAPI.CreateTenant(newTestContext()).
		TenantCreationRequest(createReq).Execute()
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestCreateTenant_NameConflictsWithExistingAlias(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tenant1 := createAndCleanupTenant(t, ctx)

	alias := GenerateRandomString("alias-")
	updateReq := openapi.TenantUpdateRequest{}
	updateReq.SetAlias(alias)
	_, resp, err := OpenapiClient.TenantAPI.UpdateTenantById(newTestContext(), tenant1.Tenant.Id).
		TenantUpdateRequest(updateReq).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	createReq := openapi.TenantCreationRequest{
		Name:  alias,
		Email: GenerateRandomEmail(alias),
	}

	_, resp, err = OpenapiClient.TenantAPI.CreateTenant(newTestContext()).
		TenantCreationRequest(createReq).Execute()
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestCreateTenant_AliasConflictsWithExistingName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tenant1 := createAndCleanupTenant(t, ctx)

	name2 := GenerateRandomTenantName()
	createReq := openapi.TenantCreationRequest{
		Name:  name2,
		Email: GenerateRandomEmail(name2),
	}
	createReq.SetAlias(tenant1.Tenant.Name)

	_, resp, err := OpenapiClient.TenantAPI.CreateTenant(newTestContext()).
		TenantCreationRequest(createReq).Execute()
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// Update Tenant
// ---------------------------------------------------------------------------

func TestUpdateTenantAlias(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	created := createAndCleanupTenant(t, ctx)

	alias := GenerateRandomString("alias-")
	updateReq := openapi.TenantUpdateRequest{}
	updateReq.SetAlias(alias)

	updated, resp, err := OpenapiClient.TenantAPI.UpdateTenantById(newTestContext(), created.Tenant.Id).
		TenantUpdateRequest(updateReq).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, updated)

	require.Equal(t, alias, updated.GetAlias())
	require.Equal(t, created.Tenant.Id, updated.Id)
	require.Equal(t, created.Tenant.Name, updated.Name)
}

func TestUpdateTenantAlias_ClearAlias(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	created := createAndCleanupTenant(t, ctx)

	alias := GenerateRandomString("alias-")
	setReq := openapi.TenantUpdateRequest{}
	setReq.SetAlias(alias)
	_, resp, err := OpenapiClient.TenantAPI.UpdateTenantById(newTestContext(), created.Tenant.Id).
		TenantUpdateRequest(setReq).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	clearReq := openapi.TenantUpdateRequest{}
	clearReq.SetAliasNil()
	updated, resp, err := OpenapiClient.TenantAPI.UpdateTenantById(newTestContext(), created.Tenant.Id).
		TenantUpdateRequest(clearReq).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.False(t, updated.HasAlias())
}

func TestUpdateTenantAlias_DuplicateAlias(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tenant1 := createAndCleanupTenant(t, ctx)
	tenant2 := createAndCleanupTenant(t, ctx)

	alias := GenerateRandomString("alias-")
	setReq := openapi.TenantUpdateRequest{}
	setReq.SetAlias(alias)

	_, resp, err := OpenapiClient.TenantAPI.UpdateTenantById(newTestContext(), tenant1.Tenant.Id).
		TenantUpdateRequest(setReq).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	_, resp, err = OpenapiClient.TenantAPI.UpdateTenantById(newTestContext(), tenant2.Tenant.Id).
		TenantUpdateRequest(setReq).Execute()
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestUpdateTenantAlias_ConflictsWithExistingName(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tenant1 := createAndCleanupTenant(t, ctx)
	tenant2 := createAndCleanupTenant(t, ctx)

	updateReq := openapi.TenantUpdateRequest{}
	updateReq.SetAlias(tenant1.Tenant.Name)

	_, resp, err := OpenapiClient.TenantAPI.UpdateTenantById(newTestContext(), tenant2.Tenant.Id).
		TenantUpdateRequest(updateReq).Execute()
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestUpdateTenantById_NotFound(t *testing.T) {
	t.Parallel()

	updateReq := openapi.TenantUpdateRequest{}
	updateReq.SetAlias("some-alias")

	_, resp, err := OpenapiClient.TenantAPI.UpdateTenantById(newTestContext(), 999999).
		TenantUpdateRequest(updateReq).Execute()
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// Create Tenant — response field validation
// ---------------------------------------------------------------------------

func TestCreateTenant_InitialStatus_IsCreating(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	created := createAndCleanupTenant(t, ctx)

	require.Equal(t, "Creating", created.Tenant.Status)
}

func TestCreateTenant_MimirId_IsPresent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	created := createAndCleanupTenant(t, ctx)

	require.True(t, created.Tenant.HasMimirId(), "mimirId should be set after tenant creation")
	require.NotEmpty(t, created.Tenant.GetMimirId())
}

func TestCreateTenant_ContractType_IsNormal(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	created := createAndCleanupTenant(t, ctx)

	require.Equal(t, "normal", created.Tenant.ContractType)
}

func TestCreateTenant_EmptyName_Returns400(t *testing.T) {
	t.Parallel()

	req := openapi.TenantCreationRequest{
		Name:  "",
		Email: GenerateRandomEmail("empty-name"),
	}
	_, resp, err := OpenapiClient.TenantAPI.CreateTenant(newTestContext()).
		TenantCreationRequest(req).Execute()
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	require.Contains(t, string(body), "Name")
}

// ---------------------------------------------------------------------------
// Delete Tenant — lifecycle edge cases
// ---------------------------------------------------------------------------

func TestDeleteTenantById_ForceDeleteTwice_Returns404(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	name := GenerateRandomTenantName()
	created, err := TestStepsClient.CreateTenantInternal(ctx, name, GenerateRandomEmail(name))
	require.NoError(t, err)

	_, resp, err := OpenapiClient.TenantAPI.DeleteTenantById(newTestContext(), created.Tenant.Id).
		IsHardDelete(true).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	_, resp, err = OpenapiClient.TenantAPI.DeleteTenantById(newTestContext(), created.Tenant.Id).
		IsHardDelete(true).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteTenantById_SoftDeleteTwice_Returns404(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	name := GenerateRandomTenantName()
	created, err := TestStepsClient.CreateTenantInternal(ctx, name, GenerateRandomEmail(name))
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = TestStepsClient.HardDeleteTenantInternal(context.Background(), created.Tenant.Id)
	})

	_, resp, err := OpenapiClient.TenantAPI.DeleteTenantById(newTestContext(), created.Tenant.Id).
		IsHardDelete(false).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	_, resp, err = OpenapiClient.TenantAPI.DeleteTenantById(newTestContext(), created.Tenant.Id).
		IsHardDelete(false).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteTenantById_SoftThenHardDelete_Succeeds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	name := GenerateRandomTenantName()
	created, err := TestStepsClient.CreateTenantInternal(ctx, name, GenerateRandomEmail(name))
	require.NoError(t, err)

	_, resp, err := OpenapiClient.TenantAPI.DeleteTenantById(newTestContext(), created.Tenant.Id).
		IsHardDelete(false).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	deleteResp, resp, err := OpenapiClient.TenantAPI.DeleteTenantById(newTestContext(), created.Tenant.Id).
		IsHardDelete(true).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, created.Tenant.Uuid, deleteResp.Uid)
}

// ---------------------------------------------------------------------------
// Non-management tenant — 403
// ---------------------------------------------------------------------------

func newNonManagementTestContext() context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, runaictx.CfSubjectID, TestUserName)
	ctx = context.WithValue(ctx, runaictx.CfIsManagementTenant, "false")
	return ctx
}

func TestGetTenants_NonManagementTenant_Returns403(t *testing.T) {
	t.Parallel()

	_, resp, err := OpenapiClient.TenantAPI.GetTenants(newNonManagementTestContext()).Execute()
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// ---------------------------------------------------------------------------
// Get Tenant By Name (Internal API)
// ---------------------------------------------------------------------------

func TestGetTenantByNameInternal(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	created := createAndCleanupTenant(t, ctx)

	tenant, resp, err := OpenapiClient.TenantsAPI.GetTenant(ctx).TenantName(created.Tenant.Name).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, tenant)

	require.Equal(t, created.Tenant.Id, tenant.Id)
	require.Equal(t, created.Tenant.Name, tenant.Name)
}

func TestGetTenantByNameInternal_NotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	_, resp, err := OpenapiClient.TenantsAPI.GetTenant(ctx).TenantName("nonexistent-tenant-name").Execute()
	require.Error(t, err)
	require.NotNil(t, resp)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
}
