/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	realm_client "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/realm/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
)

func TestGenerateClientByType_ServiceAccountTypes(t *testing.T) {
	for _, tc := range []struct {
		name       string
		clientType api.ClientType
	}{
		{name: "ServiceAccount", clientType: api.CLIENTTYPE_SERVICE_ACCOUNT},
		{name: "deprecated Application alias", clientType: api.CLIENTTYPE_APPLICATION},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, _, err := generateClientByType(context.Background(), api.ClientCreationRequest{
				Type:     tc.clientType,
				ClientId: "my-service-account",
				TenantId: "tenant-1",
			})

			require.NoError(t, err)
			require.NotNil(t, client.Attributes)
			assert.Equal(t, realm_client.ClientTypeServiceAccount, *client.Attributes.ClientType)
			assert.True(t, *client.Attributes.IsExternal)
		})
	}
}

func TestGenerateClientByType_UnknownTypeIsRejected(t *testing.T) {
	client, _, err := generateClientByType(context.Background(), api.ClientCreationRequest{
		Type:     api.ClientType("NotAType"),
		ClientId: "x",
		TenantId: "tenant-1",
	})

	assert.Nil(t, client)
	assert.ErrorIs(t, err, utils.ErrInternal)
}
