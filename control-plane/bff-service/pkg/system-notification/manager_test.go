package system_notification

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	cachemock "github.com/run-ai/runai/runai-common-packages/cache/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var messageStoreType = reflect.TypeOf((*Message)(nil)).Elem()

func TestSet_Success(t *testing.T) {
	cacher := cachemock.NewCacher(t)
	cache.SetMock(cacher)
	manager := &SystemNotificationManager{}

	ctx := context.Background()
	msg := &Message{Msg: "hello"}
	cacher.On("Set", ctx, messageStoreType, SystemNotificationSettingKey, mock.MatchedBy(func(v interface{}) bool {
		m, ok := v.(Message)
		return ok && m.Msg == "hello" && m.ID != ""
	})).Return(nil)
	err := manager.Set(ctx, msg)
	require.NoError(t, err)
	assert.NotEmpty(t, msg.ID)
}

func TestSet_CacheFails(t *testing.T) {
	cacher := cachemock.NewCacher(t)
	cache.SetMock(cacher)
	manager := &SystemNotificationManager{}

	ctx := context.Background()
	msg := &Message{Msg: "hello"}
	cacher.On("Set", ctx, messageStoreType, SystemNotificationSettingKey, mock.MatchedBy(func(v interface{}) bool {
		m, ok := v.(Message)
		return ok && m.Msg == "hello" && m.ID != ""
	})).Return(errors.New("cache error")).Once()
	err := manager.Set(ctx, msg)
	require.Error(t, err)
	assert.Equal(t, "cache error", err.Error())
}

func TestGet_CacheHit(t *testing.T) {
	cacher := cachemock.NewCacher(t)
	cache.SetMock(cacher)
	manager := &SystemNotificationManager{}
	ctx := context.Background()
	cached := &Message{ID: "id1", Msg: "cached", CreatedAt: 123}

	cacher.On("Get", ctx, messageStoreType, SystemNotificationSettingKey).Return(cached, nil)
	got, err := manager.Get(ctx)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "id1", got.ID)
	assert.Equal(t, "cached", got.Msg)
	assert.Equal(t, int64(123), got.CreatedAt)
}

func TestGet_CacheMiss(t *testing.T) {
	cacher := cachemock.NewCacher(t)
	cache.SetMock(cacher)
	manager := &SystemNotificationManager{}
	ctx := context.Background()

	cacher.On("Get", ctx, messageStoreType, SystemNotificationSettingKey).Return(nil, nil)
	got, err := manager.Get(ctx)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestGet_CacheReturnsError(t *testing.T) {
	cacher := cachemock.NewCacher(t)
	cache.SetMock(cacher)
	manager := &SystemNotificationManager{}
	ctx := context.Background()
	cacheErr := errors.New("cache error")

	cacher.On("Get", ctx, messageStoreType, SystemNotificationSettingKey).Return(nil, cacheErr)
	got, err := manager.Get(ctx)
	require.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, cacheErr, err)
}

func TestInvalidate_Success(t *testing.T) {
	cacher := cachemock.NewCacher(t)
	cache.SetMock(cacher)
	manager := &SystemNotificationManager{}
	ctx := context.Background()

	cacher.On("Del", ctx, messageStoreType, SystemNotificationSettingKey).Return(nil)
	err := manager.Invalidate(ctx)
	require.NoError(t, err)
}

func TestInvalidate_CacheDelFails(t *testing.T) {
	cacher := cachemock.NewCacher(t)
	cache.SetMock(cacher)
	manager := &SystemNotificationManager{}
	ctx := context.Background()

	cacher.On("Del", ctx, messageStoreType, SystemNotificationSettingKey).Return(errors.New("cache del failed"))
	err := manager.Invalidate(ctx)
	require.Error(t, err)
	assert.Equal(t, "cache del failed", err.Error())
}
