/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/generated"
)

type TestSteps struct {
	client *openapi.APIClient
}

func NewTestSteps(client *openapi.APIClient) *TestSteps {
	return &TestSteps{client: client}
}

func (ts *TestSteps) CreateTenantInternal(ctx context.Context, name, email string) (*openapi.TenantCreationResponse, error) {
	req := openapi.TenantCreationRequest{
		Name:  name,
		Email: email,
	}
	result, resp, err := ts.client.TenantsAPI.CreateTenantInternal(ctx).TenantCreationRequest(req).Execute()
	if err != nil {
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode
		}
		return nil, fmt.Errorf("failed to create tenant %s: %v (status: %d)", name, err, statusCode)
	}
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d when creating tenant %s", resp.StatusCode, name)
	}
	return result, nil
}

func (ts *TestSteps) HardDeleteTenantInternal(ctx context.Context, tenantId int32) error {
	_, resp, err := ts.client.TenantsAPI.DeleteTenantByIdInternal(ctx, tenantId).IsHardDelete(true).Execute()
	if err != nil {
		return fmt.Errorf("failed to hard delete tenant %d: %v", tenantId, err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d when deleting tenant %d", resp.StatusCode, tenantId)
	}
	return nil
}

func createAndCleanupTenant(t *testing.T, ctx context.Context, waitReady bool) *openapi.TenantCreationResponse {
	t.Helper()
	name := GenerateRandomTenantName()
	email := GenerateRandomEmail(name)
	created, err := TestStepsClient.CreateTenantInternal(ctx, name, email)
	require.NoError(t, err)
	require.NotNil(t, created)

	t.Cleanup(func() {
		_ = TestStepsClient.HardDeleteTenantInternal(context.Background(), created.Tenant.Id)
	})

	if waitReady {
		pollTenantStatus(t, created.Tenant.Id, "Ready", 30*time.Second)
	}

	return created
}
