/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"

	client_utils "github.com/run-ai/runai/runai-application-packages/client-utils"
	"github.com/run-ai/runai/control-plane/audit-service/pkg/client/api"
	openapi "github.com/run-ai/runai/control-plane/audit-service/pkg/client/generated"
)

type IAuditExternalClient interface {
	DownloadAuditLogs(ctx context.Context) error
	GetAuditLogs(ctx context.Context) (*openapi.GetAuditLogs200Response, error)
}

type IAuditInternalClient interface {
	SaveAuditLog(ctx context.Context, auditLog openapi.AuditLog) error
}

//go:generate mockery --name IAuditClient
type IAuditClient interface {
	IAuditExternalClient
	IAuditInternalClient
}

func NewAuditClient(baseURL string) IAuditClient {
	config := openapi.NewConfiguration()
	config.Servers = openapi.ServerConfigurations{{URL: client_utils.GetInternalUrl(baseURL)}}
	return &api.AuditClient{ApiClient: openapi.NewAPIClient(config)}
}
