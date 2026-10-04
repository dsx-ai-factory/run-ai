/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"

	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
)

type Client interface {
	GetClientID() string
	GetClientData(ctx context.Context, client *iamModels.Client) *iamModels.Client
	NeedsEmailScopeDefault() bool
}

type profileClient struct {
	clientType string
	input      ClientParams
}

func newProfileClient(clientType string, input ClientParams) Client {
	return &profileClient{clientType: clientType, input: input}
}

func (c *profileClient) GetClientID() string {
	if id := ProfileClientID(c.clientType); id != "" {
		return id
	}
	return c.input.ClientID
}

func (c *profileClient) GetClientData(ctx context.Context, client *iamModels.Client) *iamModels.Client {
	in := c.input
	in.CurrentState = client
	return BuildClientData(ctx, c.clientType, in)
}

func (c *profileClient) NeedsEmailScopeDefault() bool {
	return ProfileEmailScopeDefault(c.clientType)
}

func NewAdminUiClient(realm, tenantId, appHost, hostTemplate string) Client {
	return newProfileClient(ClientTypeRunaiAdminUI, ClientParams{
		Realm: realm, TenantId: tenantId, AppHost: appHost, HostTemplate: hostTemplate,
	})
}

func NewRunaiCliClient(realm, tenantId, appHost, hostTemplate string) Client {
	return newProfileClient(ClientTypeRunaiCLI, ClientParams{
		Realm: realm, TenantId: tenantId, AppHost: appHost, HostTemplate: hostTemplate,
	})
}

func NewCLIClient(realm, tenantId, appHost, hostTemplate string) Client {
	return newProfileClient(ClientTypeCLI, ClientParams{
		Realm: realm, TenantId: tenantId, AppHost: appHost, HostTemplate: hostTemplate,
	})
}

func NewRunaiClient(realm, tenantId, appHost, hostTemplate string) Client {
	return newProfileClient(ClientTypeRunai, ClientParams{
		Realm: realm, TenantId: tenantId, AppHost: appHost, HostTemplate: hostTemplate,
	})
}

func NewTokenExchangeClient(realm, tenantId, appHost, hostTemplate string) Client {
	return newProfileClient(ClientTypeTokenExchange, ClientParams{
		Realm: realm, TenantId: tenantId, AppHost: appHost, HostTemplate: hostTemplate,
	})
}

func NewMcpClient(realm, tenantId string) Client {
	return newProfileClient(ClientTypeMCP, ClientParams{
		Realm: realm, TenantId: tenantId,
	})
}

func NewWorkloadsAuthClient(clientId, tenantId, clusterId, appHost, hostTemplate string) Client {
	return newProfileClient(ClientTypeWorkloadsAuth, ClientParams{
		ClientID: clientId, TenantId: tenantId, ClusterId: &clusterId, AppHost: appHost, HostTemplate: hostTemplate,
	})
}

func NewRunaiServiceAccountClient(clientId, tenantId string, clusterId *string, clientType string) Client {
	return newProfileClient(clientType, ClientParams{
		ClientID: clientId, TenantId: tenantId, ClusterId: clusterId,
	})
}
