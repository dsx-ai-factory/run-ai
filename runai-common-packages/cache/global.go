/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cache

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/aws/smithy-go/ptr"
	cache_config "github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/internal/observability"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/internal/registry"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/run-ai/runai/runai-common-packages/config"
)

type Cacher interface {
	Get(ctx context.Context, storeKey reflect.Type, key string) (any, error)
	Set(ctx context.Context, storeKey reflect.Type, key string, value any) error
	Del(ctx context.Context, storeKey reflect.Type, key string) error
	RegisterStore(storeKey reflect.Type, storeName string, opts ...options.Option) error
	RegisterSharedStore(storeKey reflect.Type, storeName string, opts ...options.Option) error
	ClearStore(ctx context.Context, storeKey reflect.Type) error
	ClearCache(ctx context.Context) error
	Close() error
}

type observableCacheClienter interface {
	cacheClienter
	observability.MemoryRetrieverer
	observability.LengthRetrieverer
}

type observabilityHandlerer interface {
	WithObservability(ctx context.Context, storeName, key string, operationType observability.OperationType, operation func(ctx context.Context) (bool, error)) error
}

type storeGetterer interface {
	Get(t string) *options.Store
}

var (
	once                 sync.Once
	observabilityHandler observabilityHandlerer
	defaultStore         *options.Store
	storeGetter          storeGetterer
	cacherInstance       Cacher
	observabilitySampler *observability.Sampler
	cacheEnabled         bool
	logger               Logger = &dummyLogger{}
	initDoneCh           <-chan struct{}
	closeContext         context.Context = context.Background()
	cancelCloseContext   context.CancelFunc
)

const (
	sharedStorePrefix = "shared"
)

func Init(cfg *cache_config.Cache) error {
	var err error
	once.Do(func() {
		if err = initConfig(cfg); err != nil {
			return
		}
		if cacheEnabled = ptr.ToBool(config.Get[*cache_config.Cache]().Enabled); !cacheEnabled {
			return
		}

		closeContext, cancelCloseContext = context.WithCancel(context.Background())
		initializerInstance := newCacheInitializer(newCacheClientFactory())
		err = initializerInstance.Initialize(closeContext)
		initDoneCh = initializerInstance.Done()
	})
	return err
}

func InitAsync(cfg *cache_config.Cache) (<-chan error, error) {
	var (
		configErr error
		errCh     <-chan error
	)

	once.Do(func() {
		if configErr = initConfig(cfg); configErr != nil {
			return
		}
		if cacheEnabled = ptr.ToBool(config.Get[*cache_config.Cache]().Enabled); !cacheEnabled {
			return
		}

		initializerInstance := newCacheInitializer(newCacheClientFactory())
		closeContext, cancelCloseContext = context.WithCancel(context.Background())
		errCh = initializerInstance.InitializeAsync(closeContext)
		initDoneCh = initializerInstance.Done()
	})

	return errCh, configErr
}

func initConfig(cfg *cache_config.Cache) error {
	if cfg != nil {
		return config.SetWithDefaults(cfg)
	}
	return cache_config.InitConfig()
}

func Get[T any](ctx context.Context, key string) (*T, error) {
	if ok, err := shouldExecute(); !ok {
		return nil, err
	}
	storeKey := getStoreKeyByType[T]()

	var err error
	var res any
	err = observabilityHandler.WithObservability(ctx, getStoreName(storeKey), key, observability.OTGet, func(ctx context.Context) (bool, error) {
		res, err = cacherInstance.Get(ctx, storeKey, key)
		if err == nil {
			return res == nil, nil
		}
		return false, err
	})

	if err != nil || res == nil {
		return nil, err
	}
	if casted, ok := res.(*T); ok {
		return casted, nil
	}
	return nil, fmt.Errorf("failed to cast %v to %s", res, storeKey.String())
}

func Set[T any](ctx context.Context, key string, value T) error {
	if ok, err := shouldExecute(); !ok {
		return err
	}
	storeKey := getStoreKeyByType[T]()

	return observabilityHandler.WithObservability(ctx, getStoreName(storeKey), key, observability.OTSet, func(ctx context.Context) (bool, error) {
		return false, cacherInstance.Set(ctx, storeKey, key, value)
	})
}

func Del[T any](ctx context.Context, key string) error {
	if ok, err := shouldExecute(); !ok {
		return err
	}
	storeKey := getStoreKeyByType[T]()

	return observabilityHandler.WithObservability(ctx, getStoreName(storeKey), key, observability.OTDel, func(ctx context.Context) (bool, error) {
		return false, cacherInstance.Del(ctx, storeKey, key)
	})
}

func RegisterStore[T any](opts ...options.Option) error {
	if ok, err := shouldExecute(); !ok {
		return err
	}
	storeKey := getStoreKeyByType[T]()
	storeName := buildStoreName(storeKey)

	return observabilityHandler.WithObservability(context.Background(), storeName, "", observability.OTStore, func(ctx context.Context) (bool, error) {
		return false, cacherInstance.RegisterStore(storeKey, storeName, opts...)
	})
}

func RegisterSharedStore[T any](name string, opts ...options.Option) error {
	if ok, err := shouldExecute(); !ok {
		return err
	}
	storeKey := getStoreKeyByType[T]()
	storeName := buildSharedStoreName(name, storeKey)

	return observabilityHandler.WithObservability(context.Background(), storeName, "", observability.OTStore, func(ctx context.Context) (bool, error) {
		return false, cacherInstance.RegisterSharedStore(storeKey, storeName, opts...)
	})
}

func RegisterStoreAsync[T any](opts ...options.Option) <-chan error {
	return waitCacheReady(func() error {
		return RegisterStore[T](opts...)
	})
}

func RegisterSharedStoreAsync[T any](name string, opts ...options.Option) <-chan error {
	return waitCacheReady(func() error {
		return RegisterSharedStore[T](name, opts...)
	})
}

func ClearStore[T any]() error {
	if ok, err := shouldExecute(); !ok {
		return err
	}
	storeKey := getStoreKeyByType[T]()

	return observabilityHandler.WithObservability(context.Background(), getStoreName(storeKey), "", observability.OTClearStore, func(ctx context.Context) (bool, error) {
		return false, cacherInstance.ClearStore(ctx, storeKey)
	})
}

func ClearCache() error {
	if ok, err := shouldExecute(); !ok {
		return err
	}

	return observabilityHandler.WithObservability(context.Background(), "", "", observability.OTClearCache, func(ctx context.Context) (bool, error) {
		return false, cacherInstance.ClearCache(ctx)
	})
}

func Close() error {
	once = sync.Once{}
	if cancelCloseContext != nil {
		cancelCloseContext()
	}

	if !cacheEnabled {
		return nil
	}

	initDoneCh = nil

	if observabilitySampler != nil {
		observabilitySampler.Stop()
		observabilitySampler = nil
	}

	if cacherInstance == nil {
		return nil
	}

	err := cacherInstance.Close()
	cacherInstance = nil
	return err
}

func SetLogger(l Logger) {
	if l == nil {
		return
	}
	logger = l
}

func SetMock(cacher Cacher) {
	cacheEnabled = true
	ch := make(chan struct{})
	close(ch)
	initDoneCh = ch
	_ = config.Set(&cache_config.Cache{
		ServiceName: "test",
		Enabled:     ptr.Bool(true),
	})
	cacherInstance = cacher
	observabilityHandler = observability.NewDummyObservabilityHandler()
	storeGetter = registry.NewStoreRegistry(&options.Store{Name: config.Get[*cache_config.Cache]().ServiceName, Options: options.NewStoreOptions()})
	defaultStore = &options.Store{Name: "test-default-store", Options: options.NewStoreOptions()}
}

func IsInitialized() bool {
	return cacherInstance != nil
}

func shouldExecute() (bool, error) {
	if !cacheEnabled {
		return false, nil
	}
	if cacherInstance == nil {
		return false, fmt.Errorf("cache is not initialized")
	}

	return true, nil
}

func waitCacheReady(onReady func() error) <-chan error {
	errCh := make(chan error, 1)
	if initDoneCh == nil {
		errCh <- fmt.Errorf("cache is not initialized or during initialization: call InitAsync or Init first")
		close(errCh)
		return errCh
	}

	go func() {
		defer close(errCh)
		select {
		case <-initDoneCh:
			err := onReady()
			if err != nil {
				errCh <- err
			}
		case <-closeContext.Done():
			errCh <- fmt.Errorf("cache closed before initialization completed")
		}
	}()
	return errCh
}

func initDefaultStore(cacheClient cacheClienter, observabilityHandler observabilityHandlerer) (defaultStore *options.Store, err error) {
	defaultStore = &options.Store{Name: config.Get[*cache_config.Cache]().ServiceName, Options: options.NewStoreOptions()}

	return defaultStore, observabilityHandler.WithObservability(context.Background(), defaultStore.Name, "", observability.OTStore, func(ctx context.Context) (bool, error) {
		return false, cacheClient.Store(*defaultStore)
	})
}

func getStoreKeyByType[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func buildStoreName(storeKey reflect.Type) string {
	storeTypeName := strings.ReplaceAll(storeKey.String(), ".", "_")
	serviceName := config.Get[*cache_config.Cache]().ServiceName
	return fmt.Sprintf("%s-%s", serviceName, storeTypeName)
}

func buildSharedStoreName(name string, storeType reflect.Type) string {
	storeTypeName := strings.ReplaceAll(storeType.String(), ".", "_")
	return fmt.Sprintf("%s-%s-%s", sharedStorePrefix, name, storeTypeName)
}

func getStoreName(storeKey reflect.Type) string {
	if store := storeGetter.Get(storeKey.String()); store != nil {
		return store.Name
	}

	return defaultStore.Name
}
