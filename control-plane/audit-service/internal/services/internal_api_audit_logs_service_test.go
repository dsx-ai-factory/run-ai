/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/audit-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/audit-service/internal/services/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInternalApiAuditLogsService_SaveAuditLog(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	batcher := mocks.NewBatcher(t)
	batcher.On("Add", mock.Anything).Return(nil)

	alf := InternalApiAuditLogsService{
		logsBatcher: batcher,
	}

	res, err := alf.SaveAuditLog(ctx, api.AuditLog{TenantId: "1", RequestId: uuid.NewString(), ClusterId: uuid.NewString()})
	require.NoError(t, err)
	require.Equal(t, res.Code, http.StatusOK)
}

func TestInternalApiAuditLogsService_SaveAuditLogNoRequestId(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	batcher := mocks.NewBatcher(t)
	batcher.On("Add", mock.Anything).Return(nil)

	alf := InternalApiAuditLogsService{
		logsBatcher: batcher,
	}

	res, err := alf.SaveAuditLog(ctx, api.AuditLog{TenantId: "1", ClusterId: uuid.NewString()})
	require.NoError(t, err)
	require.Equal(t, res.Code, http.StatusOK)
}

func TestInternalApiAuditLogsService_SaveAuditLog_ParseFailed(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	batcher := mocks.NewBatcher(t)

	alf := InternalApiAuditLogsService{
		logsBatcher: batcher,
	}

	res, err := alf.SaveAuditLog(ctx, api.AuditLog{TenantId: "1", RequestId: uuid.NewString(), ClusterId: "uuid.NewString()"})
	require.Error(t, err)
	require.Equal(t, err.Error(), "failed to parse log")
	require.Equal(t, res.Code, http.StatusBadRequest)
}

func TestInternalApiAuditLogsService_SaveAuditLog_BatchFailed(t *testing.T) {
	ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "1")
	batcher := mocks.NewBatcher(t)
	batcher.On("Add", mock.Anything).Return(fmt.Errorf("failed"))

	alf := InternalApiAuditLogsService{
		logsBatcher: batcher,
	}

	res, err := alf.SaveAuditLog(ctx, api.AuditLog{TenantId: "1", RequestId: uuid.NewString(), ClusterId: uuid.NewString()})
	require.Error(t, err)
	require.Equal(t, err.Error(), "failed to batch audit log for tenant [1]")
	require.Equal(t, res.Code, http.StatusInternalServerError)
}
