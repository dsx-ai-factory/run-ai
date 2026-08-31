/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/builders"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/config"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/pointer"
	"k8s.io/utils/ptr"
)

func TestShouldResetUserPassword_NormalTenant(t *testing.T) {
	tenantInfo := builders.TenantInfo{
		Name:         "Name",
		Id:           1,
		MimirId:      "MimirId",
		Email:        "Email",
		Password:     ptr.To("Password"),
		ContractType: pointer.String(models.ContractTypeNormal),
	}
	shouldReset := shouldResetUserPassword(tenantInfo)
	assert.False(t, shouldReset)
}

func TestShouldResetUserPassword_NilContractType(t *testing.T) {
	tenantInfo := builders.TenantInfo{
		Name:         "Name",
		Id:           1,
		MimirId:      "MimirId",
		Email:        "Email",
		Password:     ptr.To("Password"),
		ContractType: nil,
	}
	shouldReset := shouldResetUserPassword(tenantInfo)
	assert.False(t, shouldReset)
}

func TestShouldResetUserPassword_EmptyPassword(t *testing.T) {
	tenantInfo := builders.TenantInfo{
		Name:         "Name",
		Id:           1,
		MimirId:      "MimirId",
		Email:        "Email",
		Password:     ptr.To(""),
		ContractType: nil,
	}
	shouldReset := shouldResetUserPassword(tenantInfo)
	assert.True(t, shouldReset)
}

// TestSetTenantData_OneClientFailure_DoesNotCancelOthers is a regression test for
// setTenantData canceling the shared ctx (via errgroup.WithContext) the moment one
// client's permission call fails, aborting other clients' still in-flight requests.
func TestSetTenantData_OneClientFailure_DoesNotCancelOthers(t *testing.T) {
	testSubject, authClientMock, _ := initTest(t, "")

	clients := []config.Client{
		{ClientId: AgentClientName + NAME},
		{ClientId: InstallerClientName},
	}

	var mu sync.Mutex
	var otherClientCtxErr error

	// setInstallerPermissions routes through request.Execute(), which calls back
	// into r.ApiService, so the returned request must point at the mock itself.
	requestForMock := authorization_api.ApiCreateInternalAccessRuleRequest{ApiService: authClientMock}

	// First goroutine to reach the mock fails immediately.
	authClientMock.On("CreateInternalAccessRule", mock.Anything).
		Once().
		Return(requestForMock)
	authClientMock.On("CreateInternalAccessRuleExecute", mock.Anything).
		Once().
		Return(nil, getHttpRespone(http.StatusInternalServerError), errors.New("boom"))

	// Second goroutine is still in-flight when the first one fails; its ctx must
	// still be usable once it gets around to making its request.
	authClientMock.On("CreateInternalAccessRule", mock.Anything).
		Once().
		Run(func(args mock.Arguments) {
			ctx := args.Get(0).(context.Context)
			time.Sleep(50 * time.Millisecond)
			mu.Lock()
			otherClientCtxErr = ctx.Err()
			mu.Unlock()
		}).
		Return(requestForMock)
	authClientMock.On("CreateInternalAccessRuleExecute", mock.Anything).
		Once().
		Return(nil, getHttpRespone(http.StatusOK), nil)

	err := testSubject.setTenantData(context.Background(), *getTenantInfo(), clients)
	assert.Error(t, err)

	mu.Lock()
	defer mu.Unlock()
	assert.NoError(t, otherClientCtxErr, "one client's failure should not cancel the ctx used by other in-flight clients")
}
