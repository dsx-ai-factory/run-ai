//go:build integration
// +build integration


/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	cache_config "github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	runaiconfig "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/stretchr/testify/require"
)

func TestLock(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	err := cache.Lock("test-lock").Lock(context.Background(), time.Duration(time.Second))
	require.NoError(t, err)

	err = cache.Lock("test-lock").Unlock(context.Background())
	require.NoError(t, err)
}

func TestLock_NegativeTTL(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	err := cache.Lock("TestLock_NegativeTTL").Lock(context.Background(), time.Duration(-1))
	require.Error(t, err)
	require.ErrorContains(t, err, "ttl > 0 is required")
}

func TestUnLock_NotLocked(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	err := cache.Lock("TestUnLock_NotLocked").Unlock(context.Background())
	require.NoError(t, err)
}

func TestLock_AlreadyLocked(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	err := cache.Lock("TestLock_AlreadyLocked").Lock(context.Background(), time.Duration(10*time.Second))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	err = cache.Lock("TestLock_AlreadyLocked").Lock(ctx, time.Duration(time.Minute))
	require.ErrorContains(t, err, "already locked")

	err = cache.Lock("TestLock_AlreadyLocked").Unlock(context.Background())
	require.NoError(t, err)
}

func TestLock_Expired(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	err := cache.Lock("TestLock_Expired").Lock(context.Background(), time.Duration(time.Second))
	require.NoError(t, err)

	time.Sleep(time.Second * 2)
	err = cache.Lock("TestLock_Expired").Lock(context.Background(), time.Second)
	require.NoError(t, err)
}

func TestLock_LockReleasedLock(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	err := cache.Lock("TestLock_LockReleasedLock").Lock(context.Background(), time.Duration(time.Second))
	require.NoError(t, err)
	err = cache.Lock("TestLock_LockReleasedLock").Unlock(context.Background())
	require.NoError(t, err)
	err = cache.Lock("TestLock_LockReleasedLock").Lock(context.Background(), time.Second)
	require.NoError(t, err)
	err = cache.Lock("TestLock_LockReleasedLock").Unlock(context.Background())
	require.NoError(t, err)
}

func TestLock_LockReleasedLock_Expired(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	err := cache.Lock("TestLock_LockReleasedLock_Expired").Lock(context.Background(), time.Duration(time.Second))
	require.NoError(t, err)
	time.Sleep(time.Second * 2)
	err = cache.Lock("TestLock_LockReleasedLock_Expired").Unlock(context.Background())
	require.NoError(t, err)
}

func TestLock_ValidatesOnlyOneRoutineInCriticalSection(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	var count atomic.Int32
	wg := sync.WaitGroup{}
	for i := range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Printf("trying to lock: %d \n", i)
			err := cache.Lock("TestLock_ValidatesOnlyOneRoutineInCriticalSection").Lock(context.Background(), time.Duration(time.Second))
			require.NoError(t, err)

			require.Equal(t, int32(1), count.Add(1))
			time.Sleep(time.Millisecond * 500)
			require.Equal(t, int32(0), count.Add(-1))
			fmt.Printf("unlocked: %d \n", i)

			err = cache.Lock("TestLock_ValidatesOnlyOneRoutineInCriticalSection").Unlock(context.Background())
			require.NoError(t, err)
		}()
	}
	wg.Wait()
	require.Equal(t, int32(0), count.Load())
}

func TestSharedLock_AlreadyLocked(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	err := cache.SharedLock("TestSharedLock_AlreadyLocked").Lock(context.Background(), time.Duration(10*time.Second))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	err = cache.SharedLock("TestSharedLock_AlreadyLocked").Lock(ctx, time.Duration(time.Minute))
	require.ErrorContains(t, err, "already locked")

	err = cache.SharedLock("TestSharedLock_AlreadyLocked").Unlock(context.Background())
	require.NoError(t, err)
}

func TestSharedLock_LockService(t *testing.T) {
	setUpIntegrationTest(t)
	defer tearDownIntegrationTest()

	err := cache.SharedLock("TestSharedLock_LockService").Lock(context.Background(), time.Duration(10*time.Second))
	require.NoError(t, err)

	err = cache.Lock("TestSharedLock_LockService").Lock(context.Background(), time.Duration(10*time.Second))
	require.NoError(t, err)

	err = cache.SharedLock("TestSharedLock_LockService").Unlock(context.Background())
	require.NoError(t, err)

	err = cache.Lock("TestSharedLock_LockService").Unlock(context.Background())
	require.NoError(t, err)
}

func TestLock_ValidatesOneLockerDoesNotReleaseOtherAfterTtl(t *testing.T) {
	setUpIntegrationTest(t)
	conf := runaiconfig.Get[*cache_config.Cache]()
	defer tearDownIntegrationTest()

	wg := sync.WaitGroup{}
	wg.Add(2)
	counter := atomic.Int32{}

	conf.ServiceName = "first-test-service"
	runaiconfig.Set(conf)
	lock := cache.SharedLock("TestLock_ValidatesOneLockerDoesNotReleaseOther")
	go func() {
		err := lock.Lock(context.Background(), time.Duration(time.Second))
		require.NoError(t, err)
		require.Equal(t, int32(1), counter.Add(1))
		time.Sleep(time.Millisecond * 990)
		lock.Unlock(context.Background())
		require.Equal(t, int32(0), counter.Add(-1))
		require.NoError(t, err)
		wg.Done()
	}()

	go func() {
		conf.ServiceName = "second-test-service"
		runaiconfig.Set(conf)

		err := cache.SharedLock("TestLock_ValidatesOneLockerDoesNotReleaseOther").Lock(context.Background(), time.Duration(time.Second))
		require.NoError(t, err)
		require.Equal(t, int32(1), counter.Add(1))
		time.Sleep(time.Millisecond * 10)
		err = cache.SharedLock("TestLock_ValidatesOneLockerDoesNotReleaseOther").Unlock(context.Background())
		require.NoError(t, err)
		require.Equal(t, int32(0), counter.Add(-1))
		wg.Done()
	}()

	conf.ServiceName = "test-service"
	runaiconfig.Set(conf)

	wg.Wait()
}

func TestLock_ValidatesOneLockerDoesNotReleaseOther(t *testing.T) {
	setUpIntegrationTest(t)
	conf := runaiconfig.Get[*cache_config.Cache]()
	defer tearDownIntegrationTest()

	counter := atomic.Int32{}
	startEndChan := make(chan byte, 2)
	expectedErr := make(chan error, 1)
	go func() {
		lock := cache.SharedLock("TestLock_ValidatesOneLockerDoesNotReleaseOther")
		startEndChan <- byte(1)
		err := lock.Lock(context.Background(), time.Duration(time.Second))
		require.NoError(t, err)

		require.Equal(t, int32(1), counter.Add(1))
		time.Sleep(time.Millisecond * 900)

		err = lock.Unlock(context.Background())
		require.Equal(t, int32(0), counter.Add(-1))
		require.NoError(t, err)
		startEndChan <- byte(0)
	}()

	go func() {
		time.Sleep(time.Millisecond * 10)
		lock := cache.SharedLock("TestLock_ValidatesOneLockerDoesNotReleaseOther")
		startEndChan <- byte(1)

		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		err := lock.Lock(ctx, time.Duration(time.Second))
		expectedErr <- err
	}()

	<-startEndChan
	<-startEndChan
	conf.ServiceName = "second-test-service"
	runaiconfig.Set(conf)
	cache.SharedLock("TestLock_ValidatesOneLockerDoesNotReleaseOther").Unlock(context.Background())

	conf.ServiceName = "test-service"
	runaiconfig.Set(conf)

	err := <-expectedErr
	require.ErrorContains(t, err, "already locked")
	<-startEndChan
}
