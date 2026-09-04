/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inmemory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/maypok86/otter/v2"
)

type Client struct {
	stores sync.Map
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Init() error {
	return nil
}

func (c *Client) Store(cfg options.Store) error {
	if _, exists := c.stores.Load(cfg.Name); exists {
		return nil
	}

	c.stores.Store(cfg.Name, otter.Must(&otter.Options[string, []byte]{
		MaximumWeight:    uint64(cfg.Options.MaxStoreSizeMB * 1000 * 1000), //nolint:gosec // G115: value is bounded/non-negative in practice
		Weigher:          entryMemoryUsageBytes,
		ExpiryCalculator: otter.ExpiryWriting[string, []byte](cfg.Options.Ttl),
	}))
	return nil
}

func (c *Client) Close() error {
	return nil
}

func (c *Client) Set(ctx context.Context, storeName string, key string, val []byte) error {
	store, ok := c.stores.Load(storeName)
	if !ok {
		return fmt.Errorf("store %s not found", storeName)
	}

	cache := store.(*otter.Cache[string, []byte])
	cache.Set(key, val)
	return nil
}

func (c *Client) Get(ctx context.Context, storeName string, key string) ([]byte, error) {
	store, ok := c.stores.Load(storeName)
	if !ok {
		return nil, fmt.Errorf("store %s not found", storeName)
	}

	cache := store.(*otter.Cache[string, []byte])
	value, ok := cache.GetIfPresent(key)
	if !ok {
		return nil, nil
	}
	return value, nil
}

func (c *Client) Delete(ctx context.Context, storeName string, key string) error {
	store, ok := c.stores.Load(storeName)
	if !ok {
		return fmt.Errorf("store %s not found", storeName)
	}

	cache := store.(*otter.Cache[string, []byte])
	cache.Invalidate(key)
	return nil
}

func (c *Client) ClearStore(ctx context.Context, storeName string) error {
	store, ok := c.stores.Load(storeName)
	if !ok {
		return fmt.Errorf("store %s not found", storeName)
	}

	cache := store.(*otter.Cache[string, []byte])
	cache.InvalidateAll()
	return nil
}

func (c *Client) RetrieveMemory(storeName string) (int64, error) {
	store, ok := c.stores.Load(storeName)
	if !ok {
		return -1, fmt.Errorf("store %s not found", storeName)
	}

	cache := store.(*otter.Cache[string, []byte])
	return int64(cache.WeightedSize()), nil //nolint:gosec // G115: value is bounded/non-negative in practice
}

func (c *Client) RetrieveLength(storeName string) (int64, error) {
	store, ok := c.stores.Load(storeName)
	if !ok {
		return -1, fmt.Errorf("store %s not found", storeName)
	}

	cache := store.(*otter.Cache[string, []byte])
	return int64(cache.EstimatedSize()), nil
}

func (c *Client) SetIfAbsent(ctx context.Context, storeName string, key string, val []byte, ttl *time.Duration) (bool, error) {
	store, ok := c.stores.Load(storeName)
	if !ok {
		return false, fmt.Errorf("store %s not found", storeName)
	}

	cache := store.(*otter.Cache[string, []byte])
	_, wasAbsent := cache.SetIfAbsent(key, val)
	if !wasAbsent {
		return false, nil
	}

	if ttl != nil {
		cache.SetExpiresAfter(key, *ttl)
	}
	return true, nil
}

func (c *Client) DeleteIf(_ context.Context, storeName string, key string, cond func(val []byte) bool) error {
	store, ok := c.stores.Load(storeName)
	if !ok {
		return fmt.Errorf("store %s not found", storeName)
	}

	cache := store.(*otter.Cache[string, []byte])
	if val, ok := cache.GetIfPresent(key); !ok {
		return nil
	} else if cond(val) {
		cache.Invalidate(key)
	}

	return nil
}

func entryMemoryUsageBytes(key string, value []byte) uint32 {
	return uint32(len(key) + len(value)) //nolint:gosec // G115: value is bounded/non-negative in practice
}
