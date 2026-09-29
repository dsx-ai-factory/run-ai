/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	iamUtils "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/utils"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	"k8s.io/utils/ptr"
)

// app mock data
const (
	APP_NAME      = "app_name"
	APP_CLIENT_ID = "app_client_id"
	APP_ENABLED   = true
)

// client mock data
const (
	NON_APP_NAME          = "non_app_name"
	CLIENT_ID             = "client_id"
	ID_OF_CLIENT          = "403978e8-58b8-465f-be53-22a81831dbb2"
	CLIENT_SECRET         = "client_secret"
	TENANT_ID_MAPPER_NAME = "TenantId Mapper"
)

const (
	DATE       = "2020-01-01"
	TIME       = "0001-01-01 00:00:00 +0000"
	TENANT_ID  = "1001"
	REALM_NAME = "realm_name"
)

const (
	PROTOCOL_MAPPER_ID       = "111"
	PROTOCOL_MAPPER_NAME     = "mapper"
	PROTOCOL_MAPPER_PROTOCOL = "protocol"
)

func GetBasicClientArray() []*iamModels.Client {
	clientArray := []*iamModels.Client{
		{
			Enabled:  ptr.To(APP_ENABLED),
			ID:       ptr.To(CLIENT_ID),
			Name:     ptr.To(APP_NAME),
			ClientID: ptr.To(APP_CLIENT_ID),
			Secret:   ptr.To(CLIENT_SECRET),
		},
		{
			ID:       ptr.To(ID_OF_CLIENT),
			Enabled:  ptr.To(true),
			ClientID: ptr.To(utils.RunaiInstallerClientId),
			Name:     ptr.To(utils.RunaiInstallerClientId),
			Secret:   ptr.To(CLIENT_SECRET),
		},
		{
			ID:       ptr.To(ID_OF_CLIENT),
			Enabled:  ptr.To(true),
			ClientID: ptr.To(utils.RUNAI_ADMIN_UI_CLIENT_ID),
			Name:     ptr.To(utils.RUNAI_ADMIN_UI_CLIENT_ID),
			Secret:   ptr.To(CLIENT_SECRET),
		},
		{
			ID:       ptr.To(ID_OF_CLIENT),
			Enabled:  ptr.To(true),
			ClientID: ptr.To(utils.RUNAI_TOKEN_EXCHANGE_CLIENT_ID),
			Name:     ptr.To(utils.RUNAI_TOKEN_EXCHANGE_CLIENT_ID),
			Secret:   ptr.To(CLIENT_SECRET),
		},
	}
	return clientArray
}

func GetMultiClientsArray() []*iamModels.Client {
	clientArray := []*iamModels.Client{
		{
			Enabled:  ptr.To(APP_ENABLED),
			ID:       ptr.To(CLIENT_ID),
			Name:     ptr.To(APP_NAME),
			ClientID: ptr.To(APP_CLIENT_ID),
			Attributes: &iamModels.ClientAttributes{
				CreatedAt: ptr.To(TIME),
			},
		},
		{
			Enabled:  ptr.To(APP_ENABLED),
			ID:       ptr.To(CLIENT_ID),
			Name:     ptr.To(NON_APP_NAME),
			ClientID: ptr.To(NON_APP_NAME),
			Attributes: &iamModels.ClientAttributes{
				CreatedAt: ptr.To(TIME),
			},
		},
	}
	return clientArray
}

func GetApiClient() api.Client {
	attributeMap := make(map[string]interface{})
	attributeMap[iamUtils.CreationTime] = TIME
	attributeMap[iamUtils.ClientTypeAttr] = utils.ClientTypeWorkloadsAuth
	return api.Client{
		Id: CLIENT_ID,
		RedirectUris: []string{
			"http://url1",
			"http://url2",
		},
		Attributes: attributeMap,
	}
}

func GetBasicReconcilerClientWithTenantIdMapperArray() []*iamModels.Client {
	clientArray := []*iamModels.Client{
		{
			Enabled:  ptr.To(true),
			ID:       ptr.To("runai-admin-ui"),
			Name:     ptr.To("runai-admin-ui"),
			ClientID: ptr.To("runai-admin-ui"),
		},
		{
			Enabled:  ptr.To(true),
			ID:       ptr.To("runai-cli"),
			Name:     ptr.To("runai-cli"),
			ClientID: ptr.To("runai-cli"),
		},
		{
			Enabled:  ptr.To(true),
			ID:       ptr.To("runai-internal.installer"),
			Name:     ptr.To("runai-internal.installer"),
			ClientID: ptr.To("runai-internal.installer"),
		},
		{
			Enabled:  ptr.To(true),
			ID:       ptr.To("cli"),
			Name:     ptr.To("cli"),
			ClientID: ptr.To("cli"),
		},
		{
			Enabled:  ptr.To(true),
			ID:       ptr.To("token-exchange"),
			Name:     ptr.To("token-exchange"),
			ClientID: ptr.To("token-exchange"),
		},
		{
			Enabled:  ptr.To(true),
			ID:       ptr.To(utils.RUNAI_MCP_CLIENT_ID),
			Name:     ptr.To(utils.RUNAI_MCP_CLIENT_ID),
			ClientID: ptr.To(utils.RUNAI_MCP_CLIENT_ID),
		},
	}
	return clientArray
}

func GetServiceAccountsArray() []iamModels.Client {
	isExternal := true
	return []iamModels.Client{
		{
			Enabled:  ptr.To(true),
			ID:       ptr.To("external1-id"),
			Name:     ptr.To("external1-name"),
			ClientID: ptr.To("external1-client-id"),
			Attributes: &iamModels.ClientAttributes{
				IsExternal: &isExternal,
			},
		},
		{
			Enabled:  ptr.To(true),
			ID:       ptr.To("external2-id"),
			Name:     ptr.To("external2-name"),
			ClientID: ptr.To("external2-client-id"),
			Attributes: &iamModels.ClientAttributes{
				IsExternal: &isExternal,
			},
		},
	}
}

func GetInternalApplicationsArray() []iamModels.Client {
	isExternal := false
	clusterId := "cluster1"
	return []iamModels.Client{
		{
			Enabled:  ptr.To(true),
			ID:       ptr.To("internal1-id"),
			Name:     ptr.To("internal1-name"),
			ClientID: ptr.To("internal1-client-id"),
			Attributes: &iamModels.ClientAttributes{
				IsExternal: &isExternal,
				ClusterId:  &clusterId,
			},
		},
		{
			Enabled:  ptr.To(true),
			ID:       ptr.To("internal2-id"),
			Name:     ptr.To("internal2-name"),
			ClientID: ptr.To("internal2-client-id"),
			Attributes: &iamModels.ClientAttributes{
				IsExternal: &isExternal,
			},
		},
	}
}
