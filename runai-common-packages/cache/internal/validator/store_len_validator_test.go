package validator

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/require"
)

func TestNewStoreLenValidator(t *testing.T) {
	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen("storeName").SetVal(10)

	hlenFetcher := func(ctx context.Context) (int64, error) {
		return clientRW.HLen(ctx, "storeName").Result()
	}

	validator, err := NewStoreLenValidator("storeName", 100, hlenFetcher)
	require.NoError(t, err)
	require.Equal(t, "storeName", validator.storeName)
	require.Equal(t, int64(100), validator.maxKeys)
	require.Equal(t, int64(10), validator.lastLen.Load())
	require.Equal(t, int64(0), validator.setsSinceLastCheck.Load())
	require.NotEmpty(t, validator.lastCheckedAtUnix.Load())
}

func TestNewStoreLenValidator_FailedToFetchStoreLen(t *testing.T) {
	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen("storeName").SetErr(errors.New("error"))

	hlenFetcher := func(ctx context.Context) (int64, error) {
		return clientRW.HLen(ctx, "storeName").Result()
	}

	validator, err := NewStoreLenValidator("storeName", 100, hlenFetcher)
	require.Equal(t, "failed to get HLEN for store <storeName>: error", err.Error())
	require.Nil(t, validator)
}

func TestValidate_NoNeedToGetStoreLen(t *testing.T) {
	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen("storeName").SetVal(10)

	hlenFetcher := func(ctx context.Context) (int64, error) {
		return clientRW.HLen(ctx, "storeName").Result()
	}
	validator, err := NewStoreLenValidator("storeName", 100, hlenFetcher)
	require.NoError(t, err)

	err = validator.Validate(context.Background())
	require.NoError(t, err)

	require.Equal(t, int64(1), validator.setsSinceLastCheck.Load())
}

func TestValidate_HasExceededSetThreshold(t *testing.T) {
	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen("storeName").SetVal(10)

	hlenFetcher := func(ctx context.Context) (int64, error) {
		return clientRW.HLen(ctx, "storeName").Result()
	}
	validator, err := NewStoreLenValidator("storeName", 100, hlenFetcher)
	require.NoError(t, err)

	validator.setsSinceLastCheck.Store(serialSetsThreshold)

	mockRW.ExpectHLen("storeName").SetVal(20)

	err = validator.Validate(context.Background())
	require.NoError(t, err)

	require.Equal(t, int64(20), validator.lastLen.Load())
	require.Equal(t, int64(1), validator.setsSinceLastCheck.Load())
}

func TestValidate_CheckIntervalExpired(t *testing.T) {
	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen("storeName").SetVal(10)

	hlenFetcher := func(ctx context.Context) (int64, error) {
		return clientRW.HLen(ctx, "storeName").Result()
	}
	validator, err := NewStoreLenValidator("storeName", 100, hlenFetcher)
	require.NoError(t, err)

	validator.lastCheckedAtUnix.Store(time.Now().Add(-2 * time.Minute).Unix())

	mockRW.ExpectHLen("storeName").SetVal(20)

	err = validator.Validate(context.Background())
	require.NoError(t, err)

	require.Equal(t, int64(20), validator.lastLen.Load())
	require.Equal(t, int64(1), validator.setsSinceLastCheck.Load())
}

func TestValidate_LastStoreLenExceedsMaxKeys(t *testing.T) {
	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen("storeName").SetVal(100)

	hlenFetcher := func(ctx context.Context) (int64, error) {
		return clientRW.HLen(ctx, "storeName").Result()
	}
	validator, err := NewStoreLenValidator("storeName", 100, hlenFetcher)
	require.NoError(t, err)

	validator.lastCheckedAtUnix.Store(time.Now().Add(-2 * time.Minute).Unix())

	mockRW.ExpectHLen("storeName").SetVal(99)

	err = validator.Validate(context.Background())
	require.NoError(t, err)

	require.Equal(t, int64(99), validator.lastLen.Load())
	require.Equal(t, int64(1), validator.setsSinceLastCheck.Load())
}

func TestValidate_StoreLenExceedsMaxKeys(t *testing.T) {
	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen("storeName").SetVal(100)

	hlenFetcher := func(ctx context.Context) (int64, error) {
		return clientRW.HLen(ctx, "storeName").Result()
	}
	validator, err := NewStoreLenValidator("storeName", 100, hlenFetcher)
	require.NoError(t, err)

	validator.lastCheckedAtUnix.Store(time.Now().Add(-2 * time.Minute).Unix())

	mockRW.ExpectHLen("storeName").SetVal(100)

	err = validator.Validate(context.Background())
	require.Equal(t, "store <storeName> exceeded max size (100 keys)", err.Error())

	require.Equal(t, int64(100), validator.lastLen.Load())
	require.Equal(t, int64(0), validator.setsSinceLastCheck.Load())
}

func TestValidateByLen_FailedToFetchStoreLen(t *testing.T) {
	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen("storeName").SetVal(100)

	hlenFetcher := func(ctx context.Context) (int64, error) {
		return clientRW.HLen(ctx, "storeName").Result()
	}
	validator, err := NewStoreLenValidator("storeName", 100, hlenFetcher)
	require.NoError(t, err)

	validator.setsSinceLastCheck.Store(serialSetsThreshold)

	mockRW.ExpectHLen("storeName").SetErr(errors.New("error"))

	err = validator.Validate(context.Background())
	require.Equal(t, "failed to get HLEN for store <storeName>: error", err.Error())

	require.Equal(t, int64(100), validator.lastLen.Load())
	require.Equal(t, int64(20), validator.setsSinceLastCheck.Load())
}

func TestValidateByLen_LenFetchedOnce(t *testing.T) {
	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen("storeName").SetVal(10)

	hlenFetcher := func(ctx context.Context) (int64, error) {
		time.Sleep(50 * time.Millisecond) // simulate a network call
		return clientRW.HLen(ctx, "storeName").Result()
	}
	validator, err := NewStoreLenValidator("storeName", 100, hlenFetcher)
	require.NoError(t, err)

	validator.setsSinceLastCheck.Store(serialSetsThreshold)

	mockRW.ExpectHLen("storeName").SetVal(0)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := validator.Validate(context.Background())
		require.NoError(t, err)
	}()
	go func() {
		defer wg.Done()
		err := validator.Validate(context.Background())
		require.NoError(t, err)
	}()

	wg.Wait()

	require.Equal(t, int64(0), validator.lastLen.Load())
	require.Equal(t, int64(2), validator.setsSinceLastCheck.Load())
}

func TestValidateByTime_FailedToFetchStoreLen(t *testing.T) {
	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen("storeName").SetVal(100)

	hlenFetcher := func(ctx context.Context) (int64, error) {
		return clientRW.HLen(ctx, "storeName").Result()
	}
	validator, err := NewStoreLenValidator("storeName", 100, hlenFetcher)
	require.NoError(t, err)

	validator.lastCheckedAtUnix.Store(time.Now().Add(-2 * time.Minute).Unix())

	mockRW.ExpectHLen("storeName").SetErr(errors.New("error"))

	err = validator.Validate(context.Background())
	require.Equal(t, "failed to get HLEN for store <storeName>: error", err.Error())

	require.Equal(t, int64(100), validator.lastLen.Load())
	require.Equal(t, int64(0), validator.setsSinceLastCheck.Load())
}

func TestValidateByTime_LenFetchedOnce(t *testing.T) {
	clientRW, mockRW := redismock.NewClientMock()
	mockRW.ExpectHLen("storeName").SetVal(10)

	hlenFetcher := func(ctx context.Context) (int64, error) {
		time.Sleep(50 * time.Millisecond) // simulate a network call
		return clientRW.HLen(ctx, "storeName").Result()
	}
	validator, err := NewStoreLenValidator("storeName", 100, hlenFetcher)
	require.NoError(t, err)

	validator.lastCheckedAtUnix.Store(time.Now().Add(-2 * time.Minute).Unix())

	mockRW.ExpectHLen("storeName").SetVal(0)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		err := validator.Validate(context.Background())
		require.NoError(t, err)
	}()
	go func() {
		defer wg.Done()
		err := validator.Validate(context.Background())
		require.NoError(t, err)
	}()

	wg.Wait()

	require.Equal(t, int64(0), validator.lastLen.Load())
	require.Equal(t, int64(2), validator.setsSinceLastCheck.Load())
}
