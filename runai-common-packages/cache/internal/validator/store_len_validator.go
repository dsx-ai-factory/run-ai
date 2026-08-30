package validator

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	serialSetsThreshold = 20
	lenCheckIntervalSec = 60
)

type storeLenValidator struct {
	hlenFetcher        func(ctx context.Context) (int64, error)
	maxKeys            int64
	storeName          string
	setsSinceLastCheck atomic.Int64
	lastCheckedAtUnix  atomic.Int64
	lastLen            atomic.Int64
	mu                 sync.Mutex
}

func NewStoreLenValidator(storeName string, maxKeys int64, hlenFetcher func(ctx context.Context) (int64, error)) (*storeLenValidator, error) {
	v := &storeLenValidator{
		hlenFetcher: hlenFetcher,
		maxKeys:     maxKeys,
		storeName:   storeName,
	}
	if err := v.updateStoreLen(context.Background(), false); err != nil {
		return nil, err
	}
	return v, nil
}

func (v *storeLenValidator) Validate(ctx context.Context) error {
	for {
		currentSetsSinceLastCheck := v.setsSinceLastCheck.Load()

		if err := v.validateByLen(ctx); err != nil {
			return err
		}
		if err := v.validateByTime(ctx); err != nil {
			return err
		}

		if v.setsSinceLastCheck.CompareAndSwap(currentSetsSinceLastCheck, currentSetsSinceLastCheck+1) {
			break
		}
	}
	return nil
}

func (v *storeLenValidator) Clear() {
	v.mu.Lock()
	v.lastLen.Store(0)
	v.setsSinceLastCheck.Store(0)
	v.lastCheckedAtUnix.Store(time.Now().Unix())
	v.mu.Unlock()
}

func (v *storeLenValidator) validateByLen(ctx context.Context) error {
	newVal := v.setsSinceLastCheck.Load() + 1
	if !v.shouldFetchByLen(newVal, v.lastLen.Load()) {
		return nil
	}

	v.mu.Lock()
	if v.shouldFetchByLen(v.setsSinceLastCheck.Load(), v.lastLen.Load()) {
		err := v.updateStoreLen(ctx, true)
		v.mu.Unlock()
		return err
	}
	v.mu.Unlock()

	return nil
}

func (v *storeLenValidator) validateByTime(ctx context.Context) error {
	now := time.Now().Unix()
	lastChecked := v.lastCheckedAtUnix.Load()
	if !v.shouldFetchByTime(lastChecked, now) {
		return nil
	}

	v.mu.Lock()
	if v.shouldFetchByTime(v.lastCheckedAtUnix.Load(), now) {
		err := v.updateStoreLen(ctx, true)
		v.mu.Unlock()
		return err
	}
	v.mu.Unlock()
	return nil
}

func (v *storeLenValidator) shouldFetchByLen(counterVal, lastLen int64) bool {
	return counterVal >= serialSetsThreshold || lastLen+counterVal > v.maxKeys
}

func (v *storeLenValidator) shouldFetchByTime(lastChecked int64, now int64) bool {
	return lastChecked+lenCheckIntervalSec < now
}

func (v *storeLenValidator) updateStoreLen(ctx context.Context, validateMaxKeys bool) error {
	newLen, err := v.getStoreLen(ctx)
	if err != nil {
		return err
	}

	if validateMaxKeys && newLen >= v.maxKeys {
		return fmt.Errorf("store <%s> exceeded max size (%d keys)", v.storeName, v.maxKeys)
	}

	v.lastLen.Store(newLen)
	v.setsSinceLastCheck.Store(0)
	v.lastCheckedAtUnix.Store(time.Now().Unix())
	return nil
}

func (v *storeLenValidator) getStoreLen(ctx context.Context) (int64, error) {
	hlen, err := v.hlenFetcher(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get HLEN for store <%s>: %w", v.storeName, err)
	}

	return hlen, nil
}
