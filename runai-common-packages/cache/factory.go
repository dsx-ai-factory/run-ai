/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cache

import (
	"fmt"

	"github.com/run-ai/runai/runai-common-packages/cache/v2/internal/valkey"

	"github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/internal/inmemory"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/internal/keydb"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/internal/nats"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/internal/redis"
)

type cacheClientFactoryer interface {
	CreateClient(cacheType config.CacheType) (observableCacheClienter, error)
}

type cacheClientFactory struct{}

func newCacheClientFactory() *cacheClientFactory {
	return &cacheClientFactory{}
}

func (f *cacheClientFactory) CreateClient(cacheType config.CacheType) (observableCacheClienter, error) {
	switch cacheType {
	case config.CTRedis:
		return redis.NewClient(), nil
	case config.CTKeyDB:
		return keydb.NewClient(), nil
	case config.CTNats:
		return nats.NewClient(), nil
	case config.CTInMemory:
		return inmemory.NewClient(), nil
	case config.CTValkey:
		return valkey.NewClient(), nil
	default:
		return nil, fmt.Errorf("cache type %s is not supported", cacheType)
	}

}
