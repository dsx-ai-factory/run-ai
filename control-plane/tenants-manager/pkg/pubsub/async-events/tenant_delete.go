/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package async_events

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"

	event "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub"
)

type TenantPublishServicer interface {
	PublishTenantDeleteEvent(ctx context.Context, event *event.TenantDelete) error
}

type TenantPublishService struct{}

func NewTenantPublishService() TenantPublishServicer {
	return &TenantPublishService{}
}

func (pubsub *TenantPublishService) PublishTenantDeleteEvent(ctx context.Context, e *event.TenantDelete) error {
	logging.Infof("publishing delete tenant event for tenant: %s", e.TenantId)
	err := runaipubsub.Publish[event.TenantDelete](ctx, e)
	if err != nil {
		return err
	}
	return nil
}
