/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
)

var storeNameCleanerRegex, _ = regexp.Compile(`[^a-zA-Z0-9_\-]+`)

type cacheClienter interface {
	Init() error
	Store(storeConfig options.Store) error
	Close() error
	Set(ctx context.Context, storeName string, key string, val []byte) error
	SetIfAbsent(ctx context.Context, storeName string, key string, val []byte, ttl *time.Duration) (bool, error)
	DeleteIf(ctx context.Context, storeName string, key string, cond func(val []byte) bool) error
	Get(ctx context.Context, storeName string, key string) ([]byte, error)
	Delete(ctx context.Context, storeName string, key string) error
	ClearStore(ctx context.Context, storeName string) error
}

type storeRegistryer interface {
	Get(t string) *options.Store
	GetOrDefault(t string) *options.Store
	Set(t string, storeConfig *options.Store)
	List() []*options.Store
}

type cacher struct {
	cacheClient   cacheClienter
	storeRegistry storeRegistryer
}

func newCacher(cacheClient cacheClienter, storeRegistry storeRegistryer) *cacher {
	return &cacher{
		cacheClient:   cacheClient,
		storeRegistry: storeRegistry,
	}
}

func (c *cacher) Get(ctx context.Context, storeKey reflect.Type, key string) (any, error) {
	store := c.getStore(storeKey)

	resBytes, err := c.cacheClient.Get(ctx, store.Name, key)
	if err != nil {
		return nil, err
	}
	if resBytes == nil {
		return nil, nil
	}

	res := reflect.New(storeKey).Interface()
	err = json.Unmarshal(resBytes, res)
	if err == nil {
		return res, nil
	}

	if delErr := c.cacheClient.Delete(ctx, store.Name, key); delErr != nil {
		return nil, fmt.Errorf("failed to unmarshal value, key deletion failed! need to delete key [%s] manually, error: %w, delete error: %w", key, err, delErr)
	}

	return nil, fmt.Errorf("failed to unmarshal value: %w", err)
}

func (c *cacher) Set(ctx context.Context, storeKey reflect.Type, key string, value any) error {
	store := c.getStore(storeKey)
	if store.Options.ReadOnly {
		return fmt.Errorf("store is read only")
	}

	valBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value [%v]: %w", value, err)
	}

	if err := validateMaxRecordSize(store.Options.MaxRecordSizeKB, key, valBytes); err != nil {
		return err
	}

	return c.cacheClient.Set(ctx, store.Name, key, valBytes)
}

func (c *cacher) Del(ctx context.Context, storeKey reflect.Type, key string) error {
	store := c.getStore(storeKey)
	if store.Options.ReadOnly {
		return fmt.Errorf("store is read only")
	}

	return c.cacheClient.Delete(ctx, store.Name, key)
}

func (c *cacher) RegisterStore(storeKey reflect.Type, storeName string, opts ...options.Option) error {
	storeOpts := options.NewStoreOptions(opts...)
	if storeOpts.ReadOnly {
		return fmt.Errorf("read only is only allowed for shared stores")
	}
	return c.createStore(storeOpts, storeKey, storeName)
}

func (c *cacher) RegisterSharedStore(storeKey reflect.Type, storeName string, opts ...options.Option) error {
	if storeNameCleanerRegex.MatchString(storeName) {
		return fmt.Errorf("store name contains invalid characters")
	}

	storeOpts := options.NewStoreOptions(opts...)
	return c.createStore(storeOpts, storeKey, storeName)
}

func (c *cacher) ClearStore(ctx context.Context, storeKey reflect.Type) error {
	storeConf := c.storeRegistry.Get(storeKey.String())
	if storeConf == nil {
		return fmt.Errorf("store not found")
	}
	if storeConf.Options.ReadOnly {
		return fmt.Errorf("cannot clear read only store")
	}

	err := c.cacheClient.ClearStore(ctx, storeConf.Name)
	if err != nil {
		return fmt.Errorf("failed to clear store %s: %w", storeConf.Name, err)
	}
	return nil
}

func (c *cacher) ClearCache(ctx context.Context) error {
	storeConfigs := c.storeRegistry.List()
	var errors []error

	for _, storeConf := range storeConfigs {
		if storeConf.Options.ReadOnly {
			continue
		}
		err := c.cacheClient.ClearStore(ctx, storeConf.Name)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to clear store %s: %w", storeConf.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to clear stores: %v", errors)
	}
	return nil
}

func (c *cacher) Close() error {
	once = sync.Once{}

	err := c.cacheClient.Close()
	if err != nil {
		return fmt.Errorf("failed to close cache client: %w", err)
	}

	return nil
}

func (c *cacher) getStore(storeKey reflect.Type) *options.Store {
	return c.storeRegistry.GetOrDefault(storeKey.String())
}

func (c *cacher) createStore(storeOpts *options.StoreOptions, storeKey reflect.Type, storeName string) error {
	if err := validateRegistryStoreType(storeKey); err != nil {
		return err
	}

	if store := c.storeRegistry.Get(storeKey.String()); store != nil && store.Name != storeName {
		return fmt.Errorf("store of type %s already registered, can not be registered again with other name, current %s new %s", storeKey.String(), store.Name, storeName)
	}

	store := &options.Store{
		Name:    storeName,
		Options: storeOpts,
	}

	err := c.cacheClient.Store(*store)
	if err != nil {
		return fmt.Errorf("failed to create store with name %s: %w", storeName, err)
	}

	c.storeRegistry.Set(storeKey.String(), store)
	return nil
}

func validateMaxRecordSize(maxRecordSizeKb int, key string, val []byte) error {
	recordSize := len([]byte(key)) + len(val)
	maxRecordSizeBytes := maxRecordSizeKb * 1000
	if recordSize > maxRecordSizeBytes {
		return fmt.Errorf("record size %.3f KB exceeds max allowed size of %d KB", float64(recordSize)/1000, maxRecordSizeKb)
	}
	return nil
}

func validateRegistryStoreType(storeKey reflect.Type) error {
	if storeKey.Kind() != reflect.Struct {
		return fmt.Errorf("only structs are accepted as store types")
	}
	return nil
}
