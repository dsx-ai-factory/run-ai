/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package safemapcache

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultCleanupPeriod      = time.Minute
	defaultInvalidationPeriod = time.Minute
)

var (
	ErrStoreInInactive    = fmt.Errorf("cannot store in inactive safemapcache")
	ErrLoadFromInactive   = fmt.Sprintf("cannot load from inactive safemapcache")
	ErrDeleteFromInactive = fmt.Sprintf("cannot to delete from inactive safemapcache")
	ErrDestroyInactive    = fmt.Sprintf("cannot destroy inactive safemapcache")
)

type SafeMapCache[K any, V any] struct {
	syncMap          sync.Map
	active           atomic.Bool
	cleanupPeriod    time.Duration
	cleanupTicker    *time.Ticker
	expirationPeriod time.Duration
}

type mapValue[V any] struct {
	value          V
	lastUpdateTime time.Time
}

func New[K, V any]() *SafeMapCache[K, V] {
	return NewWithPeriods[K, V](defaultCleanupPeriod, defaultInvalidationPeriod)
}

func NewWithPeriods[K, V any](cleanupPeriod, expirationPeriod time.Duration) *SafeMapCache[K, V] {
	mapCache := &SafeMapCache[K, V]{
		cleanupPeriod:    cleanupPeriod,
		cleanupTicker:    time.NewTicker(cleanupPeriod),
		expirationPeriod: expirationPeriod,
	}

	mapCache.active.Store(true)

	go mapCache.cleanupRoutine()

	return mapCache
}

func (u *SafeMapCache[K, V]) Store(key K, value V) {
	if !u.active.Load() {
		panic(ErrStoreInInactive)
	}

	u.syncMap.Store(key, &mapValue[V]{
		value:          value,
		lastUpdateTime: time.Now(),
	})
}

func (u *SafeMapCache[K, V]) Load(key K) (value V, found bool, valid bool) {
	if !u.active.Load() {
		panic(ErrLoadFromInactive)
	}

	val, found := u.syncMap.Load(key)
	if !found {
		var v V
		return v, false, false
	}

	realValue := val.(*mapValue[V])

	expired := time.Now().Sub(realValue.lastUpdateTime) > u.expirationPeriod

	return realValue.value, true, !expired

}

func (u *SafeMapCache[K, V]) Delete(key K) {
	if !u.active.Load() {
		panic(ErrDeleteFromInactive)
	}

	u.syncMap.Delete(key)
}

func (u *SafeMapCache[K, V]) Clear() {
	u.syncMap.Range(func(key any, value any) bool {
		u.syncMap.Delete(key)
		return true
	})
}

func (u *SafeMapCache[K, V]) Destroy() {
	if !u.active.Load() {
		panic(ErrDestroyInactive)
	}

	u.active.Store(false)
	u.cleanupFunc()
}

func (u *SafeMapCache[K, V]) Active() bool {
	return u.active.Load()
}

func (u *SafeMapCache[K, V]) cleanupRoutine() {
	for _ = range u.cleanupTicker.C {
		if !u.active.Load() {
			panic(ErrStoreInInactive)
		}

		u.cleanupFunc()
	}
}

func (u *SafeMapCache[K, V]) cleanupFunc() {
	u.syncMap.Range(func(key, value any) bool {
		realValue := value.(*mapValue[V])
		if time.Now().Sub(realValue.lastUpdateTime) > u.expirationPeriod {
			u.syncMap.Delete(key)
		}

		return true
	})
}
