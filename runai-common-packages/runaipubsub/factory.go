/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaipubsub

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/kafka"

	config_models "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/logger"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/nats"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/redis"
)

type pubSubFactory struct {
}

func newPubSubFactory() *pubSubFactory {
	return &pubSubFactory{}
}

func (f *pubSubFactory) Create(logger logger.Loggerer, psType config_models.PubSubType) (PubSuber, error) {
	switch psType {
	case config_models.PSTRedis:
		return redis.NewRedisPubSuber(logger), nil
	case config_models.PSTNats:
		return nats.NewPubSuber(logger), nil
	case config_models.PSTKafka:
		return kafka.NewPubSuber(logger), nil
	default:
		return nil, fmt.Errorf("pubsub type %s is not supported", psType)
	}
}
