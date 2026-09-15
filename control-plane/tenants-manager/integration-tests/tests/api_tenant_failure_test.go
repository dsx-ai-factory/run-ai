/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Path 1: Failed state is persisted and exposed via the API
// ---------------------------------------------------------------------------

func TestTenantFailure_StatusAndDiagnosticsExposedViaGetById(t *testing.T) {
	ctx := context.Background()
	created := createAndCleanupTenant(t, ctx, false)
	tenantId := created.Tenant.Id

	const wantDiagnostics = "builder: auth realm creation failed"
	setTenantFailed(t, tenantId, wantDiagnostics)

	tenant, resp, err := OpenapiClient.TenantAPI.GetTenantById(newTestContext(), tenantId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Failed", tenant.Status)
	assert.Equal(t, wantDiagnostics, tenant.GetDiagnostics())
}

func TestTenantFailure_FailedStatusVisibleInList(t *testing.T) {
	ctx := context.Background()
	created := createAndCleanupTenant(t, ctx, false)
	tenantId := created.Tenant.Id

	setTenantFailed(t, tenantId, "builder: mimir tenant creation failed")

	tenants, resp, err := OpenapiClient.TenantAPI.GetTenants(newTestContext()).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var found bool
	for _, ten := range tenants {
		if ten.Id == tenantId {
			assert.Equal(t, "Failed", ten.Status)
			assert.Empty(t, ten.GetDiagnostics(), "list should not return diagnostics")
			found = true
			break
		}
	}
	require.True(t, found, "Failed tenant not found in tenant list")
}

func TestTenantFailure_FailedTenantCanBeHardDeleted(t *testing.T) {
	ctx := context.Background()

	name := GenerateRandomTenantName()
	created, err := TestStepsClient.CreateTenantInternal(ctx, name, GenerateRandomEmail(name))
	require.NoError(t, err)
	tenantId := created.Tenant.Id

	setTenantFailed(t, tenantId, "builder: idp unreachable")

	deleteResp, resp, err := OpenapiClient.TenantAPI.DeleteTenantById(newTestContext(), tenantId).
		IsHardDelete(true).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, created.Tenant.Uuid, deleteResp.Uid)

	_, resp, err = OpenapiClient.TenantAPI.GetTenantById(newTestContext(), tenantId).Execute()
	require.Error(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestTenantFailure_NameIsFreedAfterHardDelete(t *testing.T) {
	ctx := context.Background()

	name := GenerateRandomTenantName()
	created, err := TestStepsClient.CreateTenantInternal(ctx, name, GenerateRandomEmail(name))
	require.NoError(t, err)

	setTenantFailed(t, created.Tenant.Id, "builder: timeout")

	_, resp, err := OpenapiClient.TenantAPI.DeleteTenantById(newTestContext(), created.Tenant.Id).
		IsHardDelete(true).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// same name can now be reused
	recreated, err := TestStepsClient.CreateTenantInternal(ctx, name, GenerateRandomEmail(name))
	require.NoError(t, err, "tenant name should be reusable after hard-deleting a Failed tenant")
	t.Cleanup(func() {
		_ = TestStepsClient.HardDeleteTenantInternal(context.Background(), recreated.Tenant.Id)
	})
	assert.Equal(t, name, recreated.Tenant.Name)
}

func TestTenantFailure_DiagnosticsEmptyForReadyTenant(t *testing.T) {
	ctx := context.Background()
	created := createAndCleanupTenant(t, ctx, false)

	tenant, resp, err := OpenapiClient.TenantAPI.GetTenantById(newTestContext(), created.Tenant.Id).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, tenant.GetDiagnostics(), "diagnostics should be empty for a non-Failed tenant")
}

// ---------------------------------------------------------------------------
// Path 2: Subscriber resets Failed → Creating → Ready on redelivery
// ---------------------------------------------------------------------------

func TestTenantFailure_SubscriberRecovery_FailedTenantBecomesReadyOnRedelivery(t *testing.T) {
	ctx := context.Background()
	created := createAndCleanupTenant(t, ctx, true)
	tenantId := created.Tenant.Id

	// simulate a builder failure by forcing the tenant back to Failed
	const failDiagnostics = "builder: auth realm failed"
	setTenantFailed(t, tenantId, failDiagnostics)

	// verify the Failed state is visible before recovery
	status, diag := getTenantStatusFromDB(t, tenantId)
	require.Equal(t, "Failed", status)
	require.Equal(t, failDiagnostics, diag)

	// re-publish TenantProvision — the subscriber will reset Failed → Creating → Ready
	event := buildProvisionEvent(t, tenantId,
		created.Tenant.Name,
		created.Tenant.GetMimirId(),
		GenerateRandomEmail(created.Tenant.Name),
	)
	publishTenantProvision(t, event)

	// subscriber should reset diagnostics and reach Ready
	pollTenantStatus(t, tenantId, "Ready", 30*time.Second)

	_, finalDiag := getTenantStatusFromDB(t, tenantId)
	assert.Empty(t, finalDiag, "diagnostics should be cleared after successful recovery")

	// verify via API
	tenant, resp, err := OpenapiClient.TenantAPI.GetTenantById(newTestContext(), tenantId).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "Ready", tenant.Status)
	assert.Empty(t, tenant.GetDiagnostics())
}

// ---------------------------------------------------------------------------
// provisioning_runs table behavior
// ---------------------------------------------------------------------------

func TestTenantFailure_ListDoesNotReturnDiagnostics(t *testing.T) {
	ctx := context.Background()
	created := createAndCleanupTenant(t, ctx, false)
	tenantId := created.Tenant.Id

	setTenantFailed(t, tenantId, "metric_store: connection refused")

	tenants, resp, err := OpenapiClient.TenantAPI.GetTenants(newTestContext()).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	for _, ten := range tenants {
		if ten.Id == tenantId {
			assert.Empty(t, ten.GetDiagnostics(), "list should not return diagnostics")
			return
		}
	}
	t.Fatal("created tenant not found in list")
}

func TestTenantFailure_HardDeleteCascadesProvisioningRuns(t *testing.T) {
	ctx := context.Background()

	name := GenerateRandomTenantName()
	created, err := TestStepsClient.CreateTenantInternal(ctx, name, GenerateRandomEmail(name))
	require.NoError(t, err)
	tenantId := created.Tenant.Id

	setTenantFailed(t, tenantId, "builder: test failure")

	db := openDB(t)
	var runCount int
	row := db.QueryRow(`SELECT COUNT(*) FROM provisioning_runs WHERE tenant_id = $1`, tenantId)
	require.NoError(t, row.Scan(&runCount))
	require.Greater(t, runCount, 0, "expected at least one provisioning_run row before delete")

	_, resp, err := OpenapiClient.TenantAPI.DeleteTenantById(newTestContext(), tenantId).
		IsHardDelete(true).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	row = db.QueryRow(`SELECT COUNT(*) FROM provisioning_runs WHERE tenant_id = $1`, tenantId)
	require.NoError(t, row.Scan(&runCount))
	assert.Equal(t, 0, runCount, "provisioning_runs should be cascade-deleted with the tenant")
}
