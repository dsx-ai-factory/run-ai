package cache

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/cenkalti/backoff/v4"
	cache_config "github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/internal/observability"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/internal/registry"
	"github.com/run-ai/runai/runai-common-packages/config"
)

type cacheInitializer struct {
	clientFactory cacheClientFactoryer
	errorsCh      chan error
	doneCh        chan struct{}
}

func newCacheInitializer(factory cacheClientFactoryer) *cacheInitializer {
	return &cacheInitializer{
		errorsCh:      make(chan error, 10),
		doneCh:        make(chan struct{}),
		clientFactory: factory,
	}
}

func (cc *cacheInitializer) Initialize(ctx context.Context) error {
	defer close(cc.doneCh)
	defer close(cc.errorsCh)

	conf := config.Get[*cache_config.Cache]()
	return cc.initWithRetry(ctx, uint64(conf.ConnectRetries)) //nolint:gosec // G115: value is bounded/non-negative in practice
}

func (cc *cacheInitializer) InitializeAsync(ctx context.Context) <-chan error {
	go func() {
		defer close(cc.doneCh)
		defer close(cc.errorsCh)

		_ = cc.initWithRetry(ctx, math.MaxInt64)
	}()
	return cc.errorsCh
}

func (cc *cacheInitializer) Done() <-chan struct{} {
	return cc.doneCh
}

func (cc *cacheInitializer) initWithRetry(ctx context.Context, retries uint64) error {
	cacheClient, err := cc.initClientWithRetry(ctx, retries)
	if err != nil {
		return err
	}

	return cc.initCacherWithRetry(ctx, cacheClient, retries)
}

func (cc *cacheInitializer) initClientWithRetry(ctx context.Context, retries uint64) (observableCacheClienter, error) {
	var cacheClient observableCacheClienter

	err := cc.retryWithBackoff(ctx, retries, "client init", func() error {
		var initErr error
		cacheClient, initErr = cc.initClient()
		return initErr
	})
	if err != nil {
		return nil, err
	}

	return cacheClient, nil
}

func (cc *cacheInitializer) initCacherWithRetry(ctx context.Context, cacheClient observableCacheClienter, retries uint64) error {
	err := cc.retryWithBackoff(ctx, retries, "cacher init", func() error {
		return cc.initCacher(cacheClient)
	})
	if err != nil {
		return err
	}

	return nil
}

func (cc *cacheInitializer) retryWithBackoff(ctx context.Context, retries uint64, stage string, operation func() error) error {
	conf := config.Get[*cache_config.Cache]()
	attempt := 0
	maxAttempts := retries + 1

	err := backoff.Retry(func() error {
		attempt++
		if opErr := operation(); opErr != nil {
			cc.handleError(stage, attempt, maxAttempts, opErr)
			return opErr
		}
		return nil
	}, cc.getBackoff(ctx, retries, conf.InitRetryIntervalDuration))

	if err != nil {
		return fmt.Errorf("failed %s after %d attempts: %w", stage, attempt, err)
	}
	return nil
}

func (cc *cacheInitializer) initClient() (observableCacheClienter, error) {
	conf := config.Get[*cache_config.Cache]()
	cacheClient, err := cc.clientFactory.CreateClient(conf.Type)
	if err != nil {
		return nil, err
	}

	if err = cacheClient.Init(); err != nil {
		return nil, err
	}

	return cacheClient, nil
}

func (cc *cacheInitializer) initCacher(cacheClient observableCacheClienter) error {
	observabilityHandler = observability.NewObservabilityHandler()
	var err error
	defaultStore, err = initDefaultStore(cacheClient, observabilityHandler)
	if err != nil {
		return err
	}

	storeRegistry := registry.NewStoreRegistry(defaultStore)
	cacherInstance = newCacher(cacheClient, storeRegistry)
	storeGetter = storeRegistry

	conf := config.Get[*cache_config.Cache]()
	if *conf.MetricsEnabled {
		observabilitySampler = observability.NewSampler(cacheClient, cacheClient, storeRegistry)
		observabilitySampler.Start()
	}
	if err := initLocks(cacheClient, observabilityHandler); err != nil {
		return err
	}
	return nil
}

func (cc *cacheInitializer) getBackoff(ctx context.Context, retries uint64, intervalDuration time.Duration) backoff.BackOff {
	var b backoff.BackOff = backoff.NewConstantBackOff(intervalDuration)
	b = backoff.WithMaxRetries(b, retries)
	return backoff.WithContext(b, ctx)
}

func (cc *cacheInitializer) handleError(stage string, attempt int, maxAttempts uint64, err error) {
	var wrappedErr error
	if maxAttempts == math.MaxInt64+1 {
		wrappedErr = fmt.Errorf("cache %s attempt %d failed: %w", stage, attempt, err)
	} else {
		wrappedErr = fmt.Errorf("cache %s attempt %d/%d failed: %w", stage, attempt, maxAttempts, err)
	}

	logger.Warnf(wrappedErr.Error())
	tryAppendToChan(cc.errorsCh, wrappedErr)
}

func tryAppendToChan[T any](channel chan T, element T) {
	if channel == nil {
		return
	}
	select {
	case channel <- element:
	default:
	}
}
