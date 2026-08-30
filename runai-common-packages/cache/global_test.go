package cache

import (
	"context"
	"testing"
	"time"

	cache_config "github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/internal/observability"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/mocks"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testType struct{}

func TestClearStore_ShouldNotExecute_CacheClientNil(t *testing.T) {
	err := ClearStore[testType]()
	assert.Nil(t, err)
}

func TestClearStore_ShouldExecute_CacheClientNil(t *testing.T) {
	enableCache()

	err := ClearStore[testType]()
	assert.EqualError(t, err, "cache is not initialized")
}

func TestRegisterStoreGlobal(t *testing.T) {
	cacheMock := initTest(t)
	defer config.FlushAll()

	storeKey := getStoreKeyByType[testType]()
	cacheMock.On("RegisterStore", storeKey, "test-cache_testType", mock.AnythingOfType("options.Option"), mock.AnythingOfType("options.Option"), mock.AnythingOfType("options.Option")).Return(nil)

	err := RegisterStore[testType](options.Ttl(time.Hour), options.MaxRecordSizeKB(100), options.MaxStoreSizeMB(100))
	assert.NoError(t, err)
}

func TestRegisterSharedStoreGlobal(t *testing.T) {
	cacheMock := initTest(t)
	defer config.FlushAll()

	cacheMock.On("RegisterSharedStore", getStoreKeyByType[testType](), "shared-name-cache_testType", mock.AnythingOfType("options.Option"), mock.AnythingOfType("options.Option"), mock.AnythingOfType("options.Option"), mock.AnythingOfType("options.Option")).Return(nil)

	assert.NoError(t, RegisterSharedStore[testType]("name", options.Ttl(time.Hour), options.MaxRecordSizeKB(100), options.MaxStoreSizeMB(100), options.ReadOnly()))
}

func TestRegisterStoreAsync_Global(t *testing.T) {
	cacheMock := initTest(t)
	defer config.FlushAll()

	initDoneCh = closedStructCh()

	storeKey := getStoreKeyByType[testType]()
	cacheMock.On("RegisterStore", storeKey, "test-cache_testType").Return(nil)

	errCh := RegisterStoreAsync[testType]()
	require.NoError(t, <-errCh)
}

func TestRegisterSharedStoreAsync_Global(t *testing.T) {
	cacheMock := initTest(t)
	defer config.FlushAll()

	initDoneCh = closedStructCh()

	storeKey := getStoreKeyByType[testType]()
	cacheMock.On("RegisterSharedStore", storeKey, "shared-name-cache_testType").Return(nil)

	errCh := RegisterSharedStoreAsync[testType]("name")
	require.NoError(t, <-errCh)
}

func TestWaitCacheReady_NilChan(t *testing.T) {
	initDoneCh = nil
	errCh := waitCacheReady(func() error { return nil })
	require.EqualError(t, <-errCh, "cache is not initialized or during initialization: call InitAsync or Init first")
}

func TestWaitCacheReady_CloseBeforeReady(t *testing.T) {
	initDoneCh = make(chan struct{}) // open channel - never closed, simulates cache not ready
	closeContext, cancelCloseContext = context.WithCancel(context.Background())
	defer func() {
		initDoneCh = nil
		closeContext = context.Background()
		cancelCloseContext = nil
	}()

	errCh := waitCacheReady(func() error { return nil })
	cancelCloseContext()

	require.EqualError(t, <-errCh, "cache closed before initialization completed")
}

func TestSetMock(t *testing.T) {
	cacherMock := mocks.NewCacher(t)

	SetMock(cacherMock)
	defer config.FlushAll()

	assert.True(t, cacheEnabled)
	assert.NotNil(t, initDoneCh)
	assert.Equal(t, cacherMock, cacherInstance)
	assert.NotNil(t, observabilityHandler)
	assert.NotNil(t, storeGetter)
	assert.Equal(t, &options.Store{Name: "test-default-store", Options: options.NewStoreOptions()}, defaultStore)
}

func TestClose(t *testing.T) {
	cacherMock := mocks.NewCacher(t)
	cacherMock.On("Close").Return(nil)

	SetMock(cacherMock)
	defer config.FlushAll()

	err := Close()
	assert.NoError(t, err)

	assert.Nil(t, initDoneCh)
}

func initTest(t *testing.T) *mocks.Cacher {
	cacherMock := mocks.NewCacher(t)

	enableCache()

	isEnabled := true
	config.Set(&cache_config.Cache{
		Enabled:     &isEnabled,
		ServiceName: "test",
	})

	cacherInstance = cacherMock
	observabilityHandler = &observability.DummyObservabilityHandler{}

	return cacherMock
}

func enableCache() {
	cacheEnabled = true
}

func closedStructCh() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}
