/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package write

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/api"
	auditconfig "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/config"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/generated"
	client_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
)

const (
	maxWriteRetries = 3
	writeRetryDelay = 2 * time.Second
)

type AuditSaver interface {
	SaveAuditLog(ctx context.Context, auditLog openapi.AuditLog) error
}

type APIWriter struct {
	client          AuditSaver
	maxWriteRetries uint
	writeRetryDelay time.Duration
}

func NewAPIWriter() *APIWriter {
	client := initAuditClient()
	return &APIWriter{client: client, maxWriteRetries: maxWriteRetries, writeRetryDelay: writeRetryDelay}
}

func (w *APIWriter) Write(log openapi.AuditLog) error {
	go func() {
		err := retry.Retry(
			func(u uint) error {
				return w.client.SaveAuditLog(context.Background(), log)
			},
			strategy.Wait(w.writeRetryDelay),
			strategy.Limit(w.maxWriteRetries),
		)
		if err != nil {
			logging.Warnf("Failed writing audit log using API writer: %v", err)
		}
	}()
	return nil
}

func initAuditClient() AuditSaver {
	auditServiceUrl := config.Get[*auditconfig.Config]().AuditServiceUrl
	config := openapi.NewConfiguration()
	config.Servers = openapi.ServerConfigurations{{URL: client_utils.GetInternalUrl(auditServiceUrl)}}
	return &api.AuditClient{ApiClient: openapi.NewAPIClient(config)}
}
