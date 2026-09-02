/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nats

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	cache_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/cache/v2/options"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

type Client struct {
	natsConnection *nats.Conn
	jetStream      jetstream.JetStream
	connCtx        context.Context
	storesMap      sync.Map
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Init() error {
	var err error
	c.natsConnection, err = nats.Connect(getConnectionUrl(), getAuthOption())
	if err != nil {
		return fmt.Errorf("could not connect to nats: %w", err)
	}

	connCtx, cancel := context.WithCancel(context.Background())
	c.connCtx = connCtx
	c.natsConnection.SetClosedHandler(func(*nats.Conn) {
		cancel()
	})

	c.jetStream, err = jetstream.New(c.natsConnection)
	if err != nil {
		return fmt.Errorf("could not create jet stream: %w", err)
	}

	return nil
}

func (c *Client) Store(conf options.Store) error {
	_, err := c.createStore(context.Background(), conf)
	if err != nil {
		return fmt.Errorf("failed to create kv store %s in cache: %w", conf.Name, err)
	}

	return nil
}

func (c *Client) Close() error {
	if err := c.natsConnection.Drain(); err != nil {
		return fmt.Errorf("could not drain nats connection: %w", err)
	}
	// Wait for the connection to be closed
	<-c.connCtx.Done()
	return nil
}

func (c *Client) SetIfAbsent(ctx context.Context, storeName string, key string, val []byte, ttl *time.Duration) (bool, error) {
	kvStore, err := c.getStore(storeName)
	if err != nil {
		return false, err
	}
	if ttl != nil {
		_, err = kvStore.Create(ctx, key, val, jetstream.KeyTTL(*ttl))
	} else {
		_, err = kvStore.Create(ctx, key, val)
	}
	if err == nil {
		return true, nil
	}
	if errors.Is(err, jetstream.ErrKeyExists) {
		return false, nil
	}
	return false, err
}

func (c *Client) DeleteIf(ctx context.Context, storeName string, key string, cond func(val []byte) bool) error {
	kvStore, err := c.getStore(storeName)
	if err != nil {
		return err
	}
	val, err := kvStore.Get(ctx, key)
	if errors.Is(err, jetstream.ErrKeyNotFound) || errors.Is(err, jetstream.ErrKeyDeleted) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get key <%s,%s> from cache: %w", storeName, key, err)
	}

	if !cond(val.Value()) {
		return nil
	}

	err = kvStore.Delete(ctx, key, jetstream.LastRevision(val.Revision()))
	if err == nil || errors.Is(err, jetstream.ErrKeyNotFound) || errors.Is(err, jetstream.ErrKeyDeleted) {
		return nil
	}
	return fmt.Errorf("failed to delete key <%s,%s> from cache: %w", storeName, key, err)
}

func (c *Client) Set(ctx context.Context, storeName string, key string, val []byte) error {
	kvStore, err := c.getStore(storeName)
	if err != nil {
		return err
	}

	_, err = kvStore.Put(ctx, key, val)
	if err != nil {
		return fmt.Errorf("failed to set key <%s,%s> in cache: %w", storeName, key, err)
	}
	return nil
}

func (c *Client) Get(ctx context.Context, storeName string, key string) ([]byte, error) {
	kvStore, err := c.getStore(storeName)
	if err != nil {
		return nil, err
	}

	val, err := kvStore.Get(ctx, key)
	if errors.Is(err, jetstream.ErrKeyNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get key <%s,%s> from cache: %w", storeName, key, err)
	}

	return val.Value(), nil
}

func (c *Client) Delete(ctx context.Context, storeName string, key string) error {
	kvStore, err := c.getStore(storeName)
	if err != nil {
		return err
	}

	err = kvStore.Purge(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete key <%s,%s> from cache: %w", storeName, key, err)
	}
	return nil
}

func (c *Client) ClearStore(ctx context.Context, storeName string) error {
	kvStore, err := c.getStore(storeName)
	if err != nil {
		return err
	}

	errors := make([]error, 0)
	keysChannel, err := kvStore.ListKeys(ctx)
	if err != nil {
		return fmt.Errorf("failed to list keys for kv store %s: %w", storeName, err)
	}
	for key := range keysChannel.Keys() {
		if err := kvStore.Purge(ctx, key); err != nil {
			errors = append(errors, fmt.Errorf("failed to purge key %s from kv store %s: %w", key, storeName, err))
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("failed to clear kv store %s: %v", storeName, errors)
	}
	return nil
}

func (c *Client) RetrieveMemory(storeName string) (int64, error) {
	kvStore, err := c.getStore(storeName)
	if err != nil {
		return -1, err
	}

	s, err := kvStore.Status(context.Background())
	if err != nil {
		return -1, err
	}
	return int64(s.Bytes()), nil //nolint:gosec // G115: value is bounded/non-negative in practice
}

func (c *Client) RetrieveLength(storeName string) (int64, error) {
	kvStore, err := c.getStore(storeName)
	if err != nil {
		return -1, err
	}

	keyLister, err := kvStore.ListKeys(context.Background())
	if err != nil {
		return -1, err
	}

	return int64(len(keyLister.Keys())), nil
}

func (c *Client) createStore(ctx context.Context, conf options.Store) (jetstream.KeyValue, error) {
	kvStore, err := c.jetStream.CreateOrUpdateKeyValue(ctx, *getStoreConfig(conf))
	if err != nil {
		return nil, fmt.Errorf("failed to create or update kv store: %w", err)
	}

	c.storesMap.Store(conf.Name, kvStore)
	return kvStore, nil
}

func (c *Client) getStore(name string) (jetstream.KeyValue, error) {
	if kvStore, ok := c.storesMap.Load(name); ok {
		return kvStore.(jetstream.KeyValue), nil
	}
	return nil, fmt.Errorf("store %s not found", name)
}

func getAuthOption() nats.Option {
	cfg := config.Get[*cache_config.Cache]()
	if cfg.User != "" && cfg.Password != "" {
		return nats.UserInfo(cfg.User, cfg.Password)
	}
	if cfg.Password != "" {
		return nats.Token(cfg.Password)
	}
	return nil
}

func getConnectionUrl() string {
	cfg := config.Get[*cache_config.Cache]()
	if cfg.ReadUrl == "" {
		return cfg.Url
	}
	return cfg.Url + "," + cfg.ReadUrl
}

func getStoreConfig(conf options.Store) *jetstream.KeyValueConfig {
	return &jetstream.KeyValueConfig{
		Bucket:         conf.Name,
		TTL:            conf.Options.Ttl,
		MaxBytes:       int64(conf.Options.MaxStoreSizeMB) * 1000 * 1000,
		MaxValueSize:   int32(conf.Options.MaxRecordSizeKB) * 1000, //nolint:gosec // G115: value is bounded/non-negative in practice
		Storage:        jetstream.MemoryStorage,
		LimitMarkerTTL: time.Minute,
	}
}
