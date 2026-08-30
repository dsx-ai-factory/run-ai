//go:build integration
// +build integration

package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/stretchr/testify/require"
)

func setUpAsyncInit(t *testing.T) <-chan error {
	setEnvs(t)

	errCh, err := cache.InitAsync(nil)
	if err != nil {
		fmt.Printf("Error when initializing cache async: %v", err)
		os.Exit(1)
	}

	return errCh
}

func waitForStoreRegistration(t *testing.T, regErrCh <-chan error) {
	select {
	case err := <-regErrCh:
		require.NoError(t, err)
		return
	case <-time.After(5 * time.Second):
		t.Fatalf("store registration did not complete")
	}
}

func waitForInitialized(t *testing.T) {
	require.Eventually(t, func() bool {
		return cache.IsInitialized()
	}, 10*time.Second, 100*time.Millisecond, "cache should be initialized")
}

func TestAsyncInit_RegisterStore(t *testing.T) {
	setUpAsyncInit(t)
	defer tearDownIntegrationTest()

	regErrCh := cache.RegisterStoreAsync[testStruct]()
	require.NotNil(t, regErrCh)

	waitForInitialized(t)
	waitForStoreRegistration(t, regErrCh)

	err := cache.Set[testStruct](context.Background(), "async-register-key", testStruct{Id: 1, Name: "async"})
	require.NoError(t, err)

	res, err := cache.Get[testStruct](context.Background(), "async-register-key")
	require.NoError(t, err)
	require.Equal(t, testStruct{Id: 1, Name: "async"}, *res)
}

func TestAsyncInit_RegisterSharedStore(t *testing.T) {
	setUpAsyncInit(t)
	defer tearDownIntegrationTest()

	regErrCh := cache.RegisterSharedStoreAsync[testStruct]("async-shared")
	require.NotNil(t, regErrCh)

	waitForInitialized(t)
	waitForStoreRegistration(t, regErrCh)

	err := cache.Set[testStruct](context.Background(), "async-shared-key", testStruct{Id: 1, Name: "async-shared"})
	require.NoError(t, err)

	res, err := cache.Get[testStruct](context.Background(), "async-shared-key")
	require.NoError(t, err)
	require.Equal(t, testStruct{Id: 1, Name: "async-shared"}, *res)
}

func TestAsyncInit_SetGetDel(t *testing.T) {
	setUpAsyncInit(t)
	defer tearDownIntegrationTest()

	waitForInitialized(t)

	key := "async-setgetdel"
	val := "hello-async"

	err := cache.Set[string](context.Background(), key, val)
	require.NoError(t, err)

	res, err := cache.Get[string](context.Background(), key)
	require.NoError(t, err)
	require.Equal(t, val, *res)

	err = cache.Del[string](context.Background(), key)
	require.NoError(t, err)

	res, err = cache.Get[string](context.Background(), key)
	require.NoError(t, err)
	require.Nil(t, res)
}

func TestAsyncInit_BadConfig(t *testing.T) {
	_, err := cache.InitAsync(nil)
	defer tearDownIntegrationTest()

	require.ErrorContains(t, err, "failed to set secrets folder")
}
