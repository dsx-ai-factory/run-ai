/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaipubsub

import (
	"context"

	config_models "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/logger"
)

type PubSuber interface {
	Connect(url, user, password string) error
	Publish(topic string, payload []byte) error
	Subscribe(ctx context.Context, topic string, callback func(event []byte) error) error
	SubscribeAsync(ctx context.Context, topic string, callback func(event []byte) error, errChan chan error)
	Close() error
}

type PubSubFactoryer interface {
	Create(logger logger.Loggerer, psType config_models.PubSubType) (PubSuber, error)
}

type PubsubConnectorer interface {
	Connect() error
	ConnectAsync(ctx context.Context) chan error
}
