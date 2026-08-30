package nats

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	cache_config "github.com/run-ai/runai/runai-common-packages/cache/v2/config"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/internal/nats/mocks"
	"github.com/run-ai/runai/runai-common-packages/cache/v2/options"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testPort = 8369

func TestNewClient(t *testing.T) {
	natsClient := NewClient()
	require.NotNil(t, natsClient)
}

func TestInit(t *testing.T) {
	s := runTestServer()
	defer s.Shutdown()

	config.Set(&cache_config.Cache{
		Url: fmt.Sprintf("localhost:%d", testPort),
	})

	natsClient := NewClient()
	err := natsClient.Init()
	require.NoError(t, err)
}

func TestInit_FailedToConnect(t *testing.T) {
	config.Set(&cache_config.Cache{
		Url: "non_existent_host:123",
	})

	natsClient := NewClient()
	err := natsClient.Init()
	require.Contains(t, err.Error(), "could not connect to nats: dial tcp: lookup non_existent_host")
}

func TestStore(t *testing.T) {
	mockKvStore := mocks.NewKeyValue(t)
	mockJetStream := mocks.NewJetStream(t)
	mockJetStream.On("CreateOrUpdateKeyValue", mock.Anything, mock.Anything).Return(mockKvStore, nil)

	natsClient := &Client{
		jetStream: mockJetStream,
	}

	storeConfig := getMockStoreConfig()
	err := natsClient.Store(storeConfig)
	require.NoError(t, err)
}

func TestStore_Failed(t *testing.T) {
	mockJetStream := mocks.NewJetStream(t)
	mockJetStream.On("CreateOrUpdateKeyValue", mock.Anything, mock.Anything).Return(nil, errors.New("invalid bucket name"))

	natsClient := &Client{
		jetStream: mockJetStream,
	}

	storeConfig := options.Store{
		Name:    "test:Store",
		Options: &options.StoreOptions{},
	}
	err := natsClient.Store(storeConfig)
	require.Equal(t, "failed to create kv store test:Store in cache: failed to create or update kv store: invalid bucket name", err.Error())
}

func TestClose(t *testing.T) {
	s := runTestServer()
	defer s.Shutdown()

	config.Set(&cache_config.Cache{
		Url: fmt.Sprintf("localhost:%d", testPort),
	})

	natsClient := NewClient()
	err := natsClient.Init()
	require.NoError(t, err)

	err = natsClient.Close()
	require.NoError(t, err)
}

func TestClose_Failed(t *testing.T) {
	s := runTestServer()
	defer s.Shutdown()

	config.Set(&cache_config.Cache{
		Url: fmt.Sprintf("localhost:%d", testPort),
	})

	natsClient := NewClient()
	err := natsClient.Init()
	require.NoError(t, err)

	natsClient.natsConnection.Close() // simulate a closed connection

	err = natsClient.Close()
	require.Equal(t, "could not drain nats connection: nats: connection closed", err.Error())
}

func TestSet(t *testing.T) {
	ctx := context.Background()
	key := "key1"
	value := []byte("value1")

	mockKvStore := mocks.NewKeyValue(t)
	mockKvStore.On("Put", ctx, key, value).Return(uint64(0), nil)

	mockJetStream := mocks.NewJetStream(t)
	mockJetStream.On("CreateOrUpdateKeyValue", mock.Anything, mock.Anything).Return(mockKvStore, nil)

	natsClient := &Client{
		jetStream: mockJetStream,
	}

	storeConfig := getMockStoreConfig()
	err := natsClient.Store(storeConfig)
	require.NoError(t, err)

	err = natsClient.Set(ctx, storeConfig.Name, key, value)
	require.NoError(t, err)
}

func TestSet_Failed(t *testing.T) {
	ctx := context.Background()
	key := "key1"
	value := []byte("value1")

	mockKvStore := mocks.NewKeyValue(t)
	mockKvStore.On("Put", ctx, key, value).Return(uint64(0), errors.New("failed to put"))

	mockJetStream := mocks.NewJetStream(t)
	mockJetStream.On("CreateOrUpdateKeyValue", mock.Anything, mock.Anything).Return(mockKvStore, nil)

	natsClient := &Client{
		jetStream: mockJetStream,
	}

	storeConfig := getMockStoreConfig()
	err := natsClient.Store(storeConfig)
	require.NoError(t, err)

	err = natsClient.Set(ctx, storeConfig.Name, key, value)
	require.Equal(t, "failed to set key <testStore,key1> in cache: failed to put", err.Error())
}

func TestSet_Failed_StoreNotFound(t *testing.T) {
	mockJetStream := mocks.NewJetStream(t)
	natsClient := &Client{
		jetStream: mockJetStream,
	}

	err := natsClient.Set(context.Background(), "notExistingStore", "key", []byte("value"))
	require.Equal(t, "store notExistingStore not found", err.Error())
}

func TestSetIfAbsent(t *testing.T) {
	ctx := context.Background()
	key := "key1"
	value := []byte("value1")
	ttl := time.Duration(10 * time.Second)

	mockKvStore := mocks.NewKeyValue(t)
	mockKvStore.On("Create", ctx, key, value, mock.AnythingOfType("jetstream.createOptFn")).Return(uint64(0), nil)

	mockJetStream := mocks.NewJetStream(t)
	mockJetStream.On("CreateOrUpdateKeyValue", mock.Anything, mock.Anything).Return(mockKvStore, nil)

	natsClient := &Client{
		jetStream: mockJetStream,
	}

	storeConfig := getMockStoreConfig()
	err := natsClient.Store(storeConfig)
	require.NoError(t, err)

	acquired, err := natsClient.SetIfAbsent(ctx, storeConfig.Name, key, value, &ttl)
	require.True(t, acquired)
	require.NoError(t, err)
}

func TestSetIfAbsent_NoTTL(t *testing.T) {
	ctx := context.Background()
	key := "key1"
	value := []byte("value1")

	mockKvStore := mocks.NewKeyValue(t)
	mockKvStore.On("Create", ctx, key, value).Return(uint64(0), nil)

	mockJetStream := mocks.NewJetStream(t)
	mockJetStream.On("CreateOrUpdateKeyValue", mock.Anything, mock.Anything).Return(mockKvStore, nil)

	natsClient := &Client{
		jetStream: mockJetStream,
	}

	storeConfig := getMockStoreConfig()
	err := natsClient.Store(storeConfig)
	require.NoError(t, err)

	acquired, err := natsClient.SetIfAbsent(ctx, storeConfig.Name, key, value, nil)
	require.True(t, acquired)
	require.NoError(t, err)
}

func TestSetIfAbsent_Failed(t *testing.T) {
	ctx := context.Background()
	key := "key1"
	value := []byte("value1")
	ttl := time.Duration(10 * time.Second)

	mockKvStore := mocks.NewKeyValue(t)
	mockKvStore.On("Create", ctx, key, value, mock.AnythingOfType("jetstream.createOptFn")).Return(uint64(0), errors.New("failed to create"))

	mockJetStream := mocks.NewJetStream(t)
	mockJetStream.On("CreateOrUpdateKeyValue", mock.Anything, mock.Anything).Return(mockKvStore, nil)

	natsClient := &Client{
		jetStream: mockJetStream,
	}

	storeConfig := getMockStoreConfig()
	err := natsClient.Store(storeConfig)
	require.NoError(t, err)

	acquired, err := natsClient.SetIfAbsent(ctx, storeConfig.Name, key, value, &ttl)
	require.False(t, acquired)
	require.Error(t, err)
}

func TestSetIfAbsent_Failed_StoreNotFound(t *testing.T) {
	mockJetStream := mocks.NewJetStream(t)
	natsClient := &Client{
		jetStream: mockJetStream,
	}
	acquired, err := natsClient.SetIfAbsent(context.Background(), "notExistingStore", "key", []byte("value"), nil)
	require.Equal(t, "store notExistingStore not found", err.Error())
	require.False(t, acquired)
}

func TestSetIfAbsent_Exists(t *testing.T) {
	ctx := context.Background()
	key := "key1"
	value := []byte("value1")
	ttl := time.Duration(10 * time.Second)

	mockKvStore := mocks.NewKeyValue(t)
	mockKvStore.On("Create", ctx, key, value, mock.AnythingOfType("jetstream.createOptFn")).Return(uint64(0), jetstream.ErrKeyExists)

	mockJetStream := mocks.NewJetStream(t)
	mockJetStream.On("CreateOrUpdateKeyValue", mock.Anything, mock.Anything).Return(mockKvStore, nil)

	natsClient := &Client{
		jetStream: mockJetStream,
	}

	storeConfig := getMockStoreConfig()
	err := natsClient.Store(storeConfig)
	require.NoError(t, err)

	acquired, err := natsClient.SetIfAbsent(ctx, storeConfig.Name, key, value, &ttl)
	require.False(t, acquired)
	require.NoError(t, err)
}

func TestGet(t *testing.T) {
	storeConfig := getMockStoreConfig()
	ctx := context.Background()
	key := "key1"
	value := []byte("value1")

	mockKvEntry := mocks.NewKeyValueEntry(t)
	mockKvEntry.On("Value").Return(value)

	mockKvStore := mocks.NewKeyValue(t)
	mockKvStore.On("Get", ctx, key).Return(mockKvEntry, nil)

	mockJetStream := mocks.NewJetStream(t)
	mockJetStream.On("CreateOrUpdateKeyValue", mock.Anything, mock.Anything).Return(mockKvStore, nil)

	natsClient := &Client{
		jetStream: mockJetStream,
	}

	err := natsClient.Store(storeConfig)
	require.NoError(t, err)

	v, err := natsClient.Get(ctx, storeConfig.Name, key)
	require.NoError(t, err)
	require.Equal(t, value, v)
}

func TestGet_Failed(t *testing.T) {
	storeConfig := getMockStoreConfig()
	ctx := context.Background()
	key := "key1"

	mockKvStore := mocks.NewKeyValue(t)
	mockKvStore.On("Get", ctx, key).Return(nil, errors.New("failed to get"))

	mockJetStream := mocks.NewJetStream(t)
	mockJetStream.On("CreateOrUpdateKeyValue", mock.Anything, mock.Anything).Return(mockKvStore, nil)

	natsClient := &Client{
		jetStream: mockJetStream,
	}

	err := natsClient.Store(storeConfig)
	require.NoError(t, err)

	v, err := natsClient.Get(ctx, storeConfig.Name, key)
	require.Equal(t, "failed to get key <testStore,key1> from cache: failed to get", err.Error())
	require.Nil(t, v)
}

func TestGet_Failed_KeyNotFound(t *testing.T) {
	storeConfig := getMockStoreConfig()
	ctx := context.Background()
	key := "key1"

	mockKvStore := mocks.NewKeyValue(t)
	mockKvStore.On("Get", ctx, "key1").Return(nil, jetstream.ErrKeyNotFound)

	mockJetStream := mocks.NewJetStream(t)
	mockJetStream.On("CreateOrUpdateKeyValue", mock.Anything, mock.Anything).Return(mockKvStore, nil)

	natsClient := &Client{
		jetStream: mockJetStream,
	}

	err := natsClient.Store(storeConfig)
	require.NoError(t, err)

	v, err := natsClient.Get(ctx, storeConfig.Name, key)
	require.NoError(t, err)
	require.Nil(t, v)
}

func TestGet_Failed_StoreNotFound(t *testing.T) {
	mockJetStream := mocks.NewJetStream(t)
	natsClient := &Client{
		jetStream: mockJetStream,
	}

	v, err := natsClient.Get(context.Background(), "notExistingStore", "key")
	require.Equal(t, "store notExistingStore not found", err.Error())
	require.Nil(t, v)
}

func TestDelete(t *testing.T) {
	storeConfig := getMockStoreConfig()
	ctx := context.Background()
	key := "key1"

	mockKvStore := mocks.NewKeyValue(t)
	mockKvStore.On("Purge", ctx, key).Return(nil)

	mockJetStream := mocks.NewJetStream(t)
	mockJetStream.On("CreateOrUpdateKeyValue", mock.Anything, mock.Anything).Return(mockKvStore, nil)

	natsClient := &Client{
		jetStream: mockJetStream,
	}

	err := natsClient.Store(storeConfig)
	require.NoError(t, err)

	err = natsClient.Delete(ctx, storeConfig.Name, key)
	require.NoError(t, err)
}

func TestDelete_Failed(t *testing.T) {
	storeConfig := getMockStoreConfig()
	ctx := context.Background()
	key := "key1"

	mockKvStore := mocks.NewKeyValue(t)
	mockKvStore.On("Purge", ctx, key).Return(errors.New("failed to purge"))

	mockJetStream := mocks.NewJetStream(t)
	mockJetStream.On("CreateOrUpdateKeyValue", mock.Anything, mock.Anything).Return(mockKvStore, nil)

	natsClient := &Client{
		jetStream: mockJetStream,
	}

	err := natsClient.Store(storeConfig)
	require.NoError(t, err)

	err = natsClient.Delete(ctx, storeConfig.Name, key)
	require.Equal(t, "failed to delete key <testStore,key1> from cache: failed to purge", err.Error())
}

func TestDelete_Failed_StoreNotFound(t *testing.T) {
	mockJetStream := mocks.NewJetStream(t)
	natsClient := &Client{
		jetStream: mockJetStream,
	}

	err := natsClient.Delete(context.Background(), "notExistingStore", "key")
	require.Equal(t, "store notExistingStore not found", err.Error())
}

func TestClearStore(t *testing.T) {
	storeConfig := getMockStoreConfig()
	ctx := context.Background()
	key := "key1"

	keysChan := make(chan string, 1)
	keysChan <- key

	go func() {
		// necessary to close the channel when key is read, and not block forever
		for len(keysChan) > 0 {
		}
		close(keysChan)
	}()

	mockKeyLister := mocks.NewKeyLister(t)
	mockKeyLister.On("Keys").Return((<-chan string)(keysChan))

	mockKvStore := mocks.NewKeyValue(t)
	mockKvStore.On("ListKeys", ctx).Return(mockKeyLister, nil)
	mockKvStore.On("Purge", ctx, key).Return(nil)

	mockJetStream := mocks.NewJetStream(t)
	mockJetStream.On("CreateOrUpdateKeyValue", mock.Anything, mock.Anything).Return(mockKvStore, nil)

	natsClient := &Client{
		jetStream: mockJetStream,
	}

	err := natsClient.Store(storeConfig)
	require.NoError(t, err)

	err = natsClient.ClearStore(ctx, storeConfig.Name)
	require.NoError(t, err)
}

func TestClearStore_Failed(t *testing.T) {
	storeConfig := getMockStoreConfig()
	ctx := context.Background()
	key := "key1"

	keysChan := make(chan string, 1)
	keysChan <- key

	go func() {
		// necessary to close the channel when key is read, and not block forever
		for len(keysChan) > 0 {
		}
		close(keysChan)
	}()

	mockKeyLister := mocks.NewKeyLister(t)
	mockKeyLister.On("Keys").Return((<-chan string)(keysChan))

	mockKvStore := mocks.NewKeyValue(t)
	mockKvStore.On("ListKeys", ctx).Return(mockKeyLister, nil)
	mockKvStore.On("Purge", ctx, key).Return(errors.New("failed to purge"))

	mockJetStream := mocks.NewJetStream(t)
	mockJetStream.On("CreateOrUpdateKeyValue", mock.Anything, mock.Anything).Return(mockKvStore, nil)

	natsClient := &Client{
		jetStream: mockJetStream,
	}

	err := natsClient.Store(storeConfig)
	require.NoError(t, err)

	err = natsClient.ClearStore(ctx, storeConfig.Name)
	require.Equal(t, "failed to clear kv store testStore: [failed to purge key key1 from kv store testStore: failed to purge]", err.Error())
}

func TestClearStore_Failed_ListStoreKeys(t *testing.T) {
	storeConfig := getMockStoreConfig()
	ctx := context.Background()

	mockKvStore := mocks.NewKeyValue(t)
	mockKvStore.On("ListKeys", ctx).Return(nil, errors.New("failed to list keys"))

	mockJetStream := mocks.NewJetStream(t)
	mockJetStream.On("CreateOrUpdateKeyValue", mock.Anything, mock.Anything).Return(mockKvStore, nil)

	natsClient := &Client{
		jetStream: mockJetStream,
	}

	err := natsClient.Store(storeConfig)
	require.NoError(t, err)

	err = natsClient.ClearStore(ctx, storeConfig.Name)
	require.Equal(t, "failed to list keys for kv store testStore: failed to list keys", err.Error())
}

func TestClearStore_Failed_StoreNotFound(t *testing.T) {
	mockJetStream := mocks.NewJetStream(t)
	natsClient := &Client{
		jetStream: mockJetStream,
	}

	err := natsClient.ClearStore(context.Background(), "notExistingStore")
	require.Equal(t, "store notExistingStore not found", err.Error())
}

func TestGetStore(t *testing.T) {
	c := Client{}
	mockKvStore := mocks.NewKeyValue(t)
	c.storesMap.Store("testStore", mockKvStore)

	store, err := c.getStore("testStore")
	require.NoError(t, err)
	require.Equal(t, mockKvStore, store)
}

func TestGetStore_Failed_StoreNotFound(t *testing.T) {
	c := Client{}
	store, err := c.getStore("testStore")
	require.Equal(t, "store testStore not found", err.Error())
	require.Nil(t, store)
}

func TestGetAuthOption_WithUserAndPassword(t *testing.T) {
	defer config.FlushAll()

	config.Set(&cache_config.Cache{
		User:     "user",
		Password: "password",
	})

	opt := getAuthOption()
	require.NotNil(t, opt)

	opts := &nats.Options{}
	err := opt(opts)
	require.NoError(t, err)

	require.Equal(t, "user", opts.User)
	require.Equal(t, "password", opts.Password)
}

func TestGetAuthOption_WithToken(t *testing.T) {
	defer config.FlushAll()

	config.Set(&cache_config.Cache{
		Password: "token",
	})

	opt := getAuthOption()
	require.NotNil(t, opt)

	opts := &nats.Options{}
	err := opt(opts)
	require.NoError(t, err)

	require.Equal(t, "token", opts.Token)
}

func TestGetAuthOption_Empty(t *testing.T) {
	opt := getAuthOption()
	require.Nil(t, opt)
}

func TestGetConnectionUrl(t *testing.T) {
	defer config.FlushAll()
	config.Set(&cache_config.Cache{
		Url: "localhost:4222",
	})
	connUrl := getConnectionUrl()
	require.Equal(t, "localhost:4222", connUrl)
}

func TestGetConnectionUrl_WithReadUrl(t *testing.T) {
	defer config.FlushAll()
	config.Set(&cache_config.Cache{
		Url:     "localhost:4222",
		ReadUrl: "localhost:4223",
	})
	connUrl := getConnectionUrl()
	require.Equal(t, "localhost:4222,localhost:4223", connUrl)
}

func getMockStoreConfig() options.Store {
	return options.Store{
		Name: "testStore",
		Options: &options.StoreOptions{
			Ttl: 10 * time.Second,
		},
	}
}

func runTestServer() *server.Server {
	opts := natsserver.DefaultTestOptions
	opts.Port = testPort
	opts.JetStream = true
	return natsserver.RunServer(&opts)
}
