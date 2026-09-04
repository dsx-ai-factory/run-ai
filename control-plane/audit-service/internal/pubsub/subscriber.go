/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pubsub

import (
	"context"
	"fmt"
	"strconv"

	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
)

type AuditDeleter interface {
	HardDeleteTenantAuditLogs(ctx context.Context, tenantId int) error
}

type Subscriber struct {
	auditDeleter AuditDeleter
}

func NewSubscriber(auditDeleter AuditDeleter) *Subscriber {
	return &Subscriber{
		auditDeleter: auditDeleter,
	}
}

func (s *Subscriber) SubscribeTenantDelete(ctx context.Context) error {
	return runaipubsub.Subscribe[events.TenantDelete](s.tenantDeleteSubscribeHandler)
}

func (s *Subscriber) tenantDeleteSubscribeHandler(ctx context.Context, event *events.TenantDelete) error {
	logging.Infof("Handling tenant delete request: %v", event)
	intTenantId, err := strconv.Atoi(event.TenantId)
	if err != nil {
		return fmt.Errorf("failed to convert tenant id to int: %v", err)
	}
	return s.auditDeleter.HardDeleteTenantAuditLogs(ctx, intTenantId)
}
