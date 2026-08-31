/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package lock

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	cache_config "github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/internal/observability"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/run-ai/runai/runai-common-packages/config"
)

type cacheClienter interface {
	SetIfAbsent(ctx context.Context, storeName string, key string, val []byte, ttl *time.Duration) (bool, error)
	DeleteIf(ctx context.Context, storeName string, key string, cond func(val []byte) bool) error
	Delete(ctx context.Context, storeName string, key string) error
}

type observabilityHandlerer interface {
	WithObservability(ctx context.Context, storeName, key string, operationType observability.OperationType, operation func(ctx context.Context) (bool, error)) error
}

type locker struct {
	cacheClient          cacheClienter
	observabilityHandler observabilityHandlerer
	store                options.Store
	lockUniqueId         string
	name                 string
}

func newLocker(cacheClient cacheClienter, observabilityHandler observabilityHandlerer, store options.Store, lockUniqueId string, name string) *locker {
	return &locker{
		cacheClient:          cacheClient,
		observabilityHandler: observabilityHandler,
		store:                store,
		lockUniqueId:         fmt.Sprintf("%s-%s", config.Get[*cache_config.Cache]().ServiceName, lockUniqueId),
		name:                 name,
	}
}

func (l *locker) Lock(ctx context.Context, ttl time.Duration) error {
	if ttl <= time.Duration(0) {
		return fmt.Errorf("ttl > 0 is required")
	}
	if err := validateLockName(l.name); err != nil {
		return err
	}

	return l.observabilityHandler.WithObservability(ctx, l.store.Name, l.name, observability.OTLock, func(ctx context.Context) (bool, error) {
		return false, l.lock(ctx, ttl)
	})
}

func (l *locker) lock(ctx context.Context, ttl time.Duration) error {
	lock := Record{
		LockedAt: time.Now(),
		LockedBy: l.lockUniqueId,
	}

	valBytes, err := json.Marshal(lock)
	if err != nil {
		return fmt.Errorf("failed to marshal lock [%v]: %w", lock, err)
	}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("key %s.%s already locked", l.store.Name, l.name)
		default:
			if acquired, err := l.cacheClient.SetIfAbsent(context.WithoutCancel(ctx), l.store.Name, l.name, valBytes, &ttl); acquired {
				return nil
			} else if err != nil {
				return fmt.Errorf("failed to lock key %s.%s: %w", l.store.Name, l.name, err)
			}

			time.Sleep(time.Millisecond * 100)
		}
	}

}

func (l *locker) Unlock(ctx context.Context) error {
	if err := validateLockName(l.name); err != nil {
		return err
	}
	return l.observabilityHandler.WithObservability(ctx, l.store.Name, l.name, observability.OTUnlock, func(ctx context.Context) (bool, error) {
		return false, l.unlock(ctx)
	})
}

func (l *locker) unlock(ctx context.Context) error {
	if err := l.cacheClient.DeleteIf(ctx, l.store.Name, l.name, l.lockIdComparator()); err != nil {
		return fmt.Errorf("failed to unlock key %s.%s: %w", l.store.Name, l.name, err)
	}
	return nil
}

func (l *locker) lockIdComparator() func(val []byte) bool {
	return func(val []byte) bool {
		r := Record{}
		if err := json.Unmarshal(val, &r); err != nil {
			return false
		}

		return r.LockedBy == l.lockUniqueId
	}
}

func validateLockName(name string) error {
	if name == "" {
		return fmt.Errorf("lock name is required")
	}
	if len(name) > 50 {
		return fmt.Errorf("lock name must be less than 50 characters")
	}
	return nil
}
