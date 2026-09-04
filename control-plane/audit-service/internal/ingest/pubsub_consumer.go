/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ingest

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/audit-service/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
)

type PubSubConsumer struct{}

func NewPubSubConsumer() MessageConsumer {
	return &PubSubConsumer{}
}

func (psc *PubSubConsumer) Start(ctx context.Context, handleLog func(log *api.AuditLog) error) error {
	return runaipubsub.Subscribe(func(ctx context.Context, log *api.AuditLog) error {
		return handleLog(log)
	})
}

func (psc *PubSubConsumer) Stop() error {
	return runaipubsub.Close()
}

func (psc *PubSubConsumer) GetName() string {
	return "subscriber"
}
