/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package events_listeners

import (
	"context"
	"fmt"
	"testing"

	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/events_listeners/mocks"
)

const tenantId = "1"

func TestTenantDeleteCallback_Success(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewTenantDeleteClienter(t)
	l := &TenantDeleteListener{client: client}

	client.On("Delete", ctx, tenantId).Return(nil)

	err := l.tenantDeleteCallback(ctx, &events.TenantDelete{TenantId: tenantId})
	require.NoError(t, err)
}

func TestTenantDeleteCallback_ManagerReturnsError(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewTenantDeleteClienter(t)
	l := &TenantDeleteListener{client: client}

	client.On("Delete", ctx, tenantId).Return(fmt.Errorf("delete failed"))

	err := l.tenantDeleteCallback(ctx, &events.TenantDelete{TenantId: tenantId})
	require.EqualError(t, err, "delete failed")
}
