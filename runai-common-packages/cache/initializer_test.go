package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	cache_config "github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupConnectorTestConfig(t *testing.T, connectRetries int) {
	metricsEnabled := false
	conf := config.Get[*cache_config.Cache]()
	conf.ConnectRetries = connectRetries
	conf.InitRetryIntervalDuration = 10 * time.Millisecond
	conf.Type = cache_config.CTRedis
	conf.ServiceName = "test"
	conf.MetricsEnabled = &metricsEnabled
	config.Set(conf)
	t.Cleanup(func() {
		config.FlushAll()
		cacherInstance = nil
		observabilityHandler = nil
		defaultStore = nil
		storeGetter = nil
		lockManagerInstance = nil
	})
}

func TestCacheConnector(t *testing.T) {
	factoryMock := NewMockCacheClientFactoryer(t)
	connector := newCacheInitializer(factoryMock)

	require.NotNil(t, connector)
	require.Equal(t, factoryMock, connector.clientFactory)
	require.NotNil(t, connector.errorsCh)
	require.NotNil(t, connector.doneCh)
}

func TestConnect(t *testing.T) {
	factoryMock := NewMockCacheClientFactoryer(t)
	clientMock := NewMockObservableCacheClienter(t)
	connector := newCacheInitializer(factoryMock)

	setupConnectorTestConfig(t, 2)

	clientMock.On("Init").Return(nil)
	clientMock.On("Store", mock.Anything).Return(nil)
	factoryMock.On("CreateClient", cache_config.CTRedis).Return(clientMock, nil)

	err := connector.Initialize(context.Background())
	require.NoError(t, err)
	require.NotNil(t, cacherInstance)

	_, ok := <-connector.errorsCh
	require.False(t, ok, "errors channel should be closed after Connect")

	_, ok = <-connector.doneCh
	require.False(t, ok, "done channel should be closed after Connect")
}

func TestConnect_ClientInitFailed(t *testing.T) {
	factoryMock := NewMockCacheClientFactoryer(t)
	clientMock := NewMockObservableCacheClienter(t)
	connector := newCacheInitializer(factoryMock)

	setupConnectorTestConfig(t, 2)

	clientMock.On("Init").Return(fmt.Errorf("init error"))
	factoryMock.On("CreateClient", cache_config.CTRedis).Return(clientMock, nil)

	err := connector.Initialize(context.Background())
	require.EqualError(t, err, "failed client init after 3 attempts: init error")

	for i := range 3 {
		errMsg, ok := <-connector.errorsCh
		require.True(t, ok, "errors should be sent to channel")
		require.EqualError(t, errMsg, fmt.Sprintf("cache client init attempt %d/%d failed: init error", i+1, 3))
	}

	_, ok := <-connector.errorsCh
	require.False(t, ok, "errors channel should be closed after Connect")

	_, ok = <-connector.doneCh
	require.False(t, ok, "done channel should be closed after Connect")
}

func TestConnect_InitCacherFailed(t *testing.T) {
	factoryMock := NewMockCacheClientFactoryer(t)
	clientMock := NewMockObservableCacheClienter(t)
	connector := newCacheInitializer(factoryMock)

	setupConnectorTestConfig(t, 2)

	clientMock.On("Init").Return(nil)
	clientMock.On("Store", mock.Anything).Return(fmt.Errorf("jetstream unavailable"))
	factoryMock.On("CreateClient", cache_config.CTRedis).Return(clientMock, nil)

	err := connector.Initialize(context.Background())
	require.ErrorContains(t, err, "jetstream unavailable")

	clientMock.AssertNumberOfCalls(t, "Init", 1)
}

func TestConnectAsync(t *testing.T) {
	factoryMock := NewMockCacheClientFactoryer(t)
	clientMock := NewMockObservableCacheClienter(t)
	connector := newCacheInitializer(factoryMock)

	setupConnectorTestConfig(t, 0)

	clientMock.On("Init").Return(nil)
	clientMock.On("Store", mock.Anything).Return(nil)
	factoryMock.On("CreateClient", cache_config.CTRedis).Return(clientMock, nil)

	errCh := connector.InitializeAsync(context.Background())
	require.NotNil(t, errCh)

	<-connector.Done()

	require.NotNil(t, cacherInstance)

	_, ok := <-errCh
	require.False(t, ok, "errors channel should be closed after ConnectAsync")

	_, ok = <-connector.doneCh
	require.False(t, ok, "done channel should be closed after ConnectAsync")
}

func TestConnectAsync_ClientInitFailed(t *testing.T) {
	factoryMock := NewMockCacheClientFactoryer(t)
	clientMock := NewMockObservableCacheClienter(t)
	connector := newCacheInitializer(factoryMock)

	setupConnectorTestConfig(t, 0)

	clientMock.On("Init").Return(fmt.Errorf("init error")).Times(20)
	clientMock.On("Init").Return(nil)
	clientMock.On("Store", mock.Anything).Return(nil)
	factoryMock.On("CreateClient", cache_config.CTRedis).Return(clientMock, nil)

	errCh := connector.InitializeAsync(context.Background())
	require.NotNil(t, errCh)

	<-connector.Done()

	require.NotNil(t, cacherInstance)

	for i := range 10 {
		errMsg, ok := <-errCh
		require.True(t, ok, "errors should be sent to channel")
		require.EqualError(t, errMsg, fmt.Sprintf("cache client init attempt %d failed: init error", i+1))
	}

	_, ok := <-errCh
	require.False(t, ok, "errors channel should be closed after ConnectAsync")

	_, ok = <-connector.doneCh
	require.False(t, ok, "done channel should be closed after ConnectAsync")
}

func TestConnectAsync_InitCacherRetries(t *testing.T) {
	factoryMock := NewMockCacheClientFactoryer(t)
	clientMock := NewMockObservableCacheClienter(t)
	connector := newCacheInitializer(factoryMock)

	setupConnectorTestConfig(t, 0)

	clientMock.On("Init").Return(nil)
	clientMock.On("Store", mock.Anything).Return(fmt.Errorf("jetstream unavailable")).Times(3)
	clientMock.On("Store", mock.Anything).Return(nil)
	factoryMock.On("CreateClient", cache_config.CTRedis).Return(clientMock, nil)

	errCh := connector.InitializeAsync(context.Background())
	require.NotNil(t, errCh)

	<-connector.Done()

	require.NotNil(t, cacherInstance)
	clientMock.AssertNumberOfCalls(t, "Init", 1)
}

func TestConnectWithRetry_Failed(t *testing.T) {
	factoryMock := NewMockCacheClientFactoryer(t)
	clientMock := NewMockObservableCacheClienter(t)

	setupConnectorTestConfig(t, 2)

	clientMock.On("Init").Return(fmt.Errorf("init error"))
	factoryMock.On("CreateClient", cache_config.CTRedis).Return(clientMock, nil)

	connector := newCacheInitializer(factoryMock)
	err := connector.initWithRetry(context.Background(), 2)
	require.Error(t, err)
}

func TestConnectWithRetry_FailedToCreateClient(t *testing.T) {
	factoryMock := NewMockCacheClientFactoryer(t)
	clientMock := NewMockObservableCacheClienter(t)

	setupConnectorTestConfig(t, 2)

	clientMock.On("Init").Return(fmt.Errorf("init error")).Times(3)
	factoryMock.On("CreateClient", cache_config.CTRedis).Return(clientMock, nil)

	connector := newCacheInitializer(factoryMock)
	err := connector.initWithRetry(context.Background(), 2)
	require.Error(t, err)
}

func TestConnectWithRetry_ContextCancellation(t *testing.T) {
	factoryMock := NewMockCacheClientFactoryer(t)
	clientMock := NewMockObservableCacheClienter(t)

	setupConnectorTestConfig(t, 100)

	clientMock.On("Init").Return(fmt.Errorf("init error")).Once()
	factoryMock.On("CreateClient", cache_config.CTRedis).Return(clientMock, nil)

	connector := newCacheInitializer(factoryMock)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := connector.initWithRetry(ctx, 100)
	require.EqualError(t, err, "failed client init after 1 attempts: context canceled")
}
