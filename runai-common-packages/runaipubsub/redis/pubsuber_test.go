/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/redis/go-redis/v9"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	config_models "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/logger"
	logger_mocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/logger/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/pevent"
	mocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/redis/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockStreamConnection struct {
	mockRedis      *redis.Client
	mockPublisher  *gochannel.GoChannel
	mockSubscriber *gochannel.GoChannel
}

func (m *mockStreamConnection) connect(logger *Logger, url, user, password string) (*redis.Client, message.Publisher, message.Subscriber, error) {
	m.mockRedis = redis.NewClient(&redis.Options{
		Addr: "mock-redis:6379",
	})
	pub := gochannel.NewGoChannel(
		gochannel.Config{Persistent: true},
		logger,
	)
	m.mockPublisher = pub
	m.mockSubscriber = pub
	return m.mockRedis, pub, pub, nil
}

func newMockRedisManager(logger logger.Loggerer, errFunc func(e error)) *redisPubSuber {
	return &redisPubSuber{
		connection: &mockStreamConnection{},
		logger:     NewLogger(logger),
	}
}

func TestConnection(t *testing.T) {
	conf := &config_models.Pubsub{}
	conf.SetDefaults()
	config.Set(conf)
	logger := logger_mocks.NewLoggerer(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything)
	logger.On("Infof", mock.Anything, mock.Anything, mock.Anything)
	pubSuber := newMockRedisManager(logger, func(e error) {})
	err := pubSuber.Connect(conf.Url, conf.User, conf.Password)
	assert.Nil(t, err)
	assert.True(t, pubSuber.router.IsRunning())
}

func TestCloseOnRouterContext(t *testing.T) {
	conf := &config_models.Pubsub{}
	conf.SetDefaults()
	config.Set(conf)
	logger := logger_mocks.NewLoggerer(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything)
	logger.On("Infof", mock.Anything, mock.Anything, mock.Anything)
	pubSuber := newMockRedisManager(logger, func(e error) {})
	err := pubSuber.Connect(conf.Url, conf.User, conf.Password)
	assert.Nil(t, err)
	assert.True(t, pubSuber.router.IsRunning())
	ctx1 := context.Background()
	ctx2 := context.Background()
	err = pubSuber.Subscribe(ctx1, "test", func(event []byte) error {
		return nil
	})
	assert.Equal(t, 1, len(pubSuber.router.Handlers()))
	assert.Nil(t, err)
	err = pubSuber.Subscribe(ctx2, "test1", func(event []byte) error {
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(pubSuber.router.Handlers()))
	time.Sleep(time.Second * 2)
	pubSuber.Close()
	assert.Eventually(t, func() bool {
		return len(pubSuber.router.Handlers()) == 0 && pubSuber.router.IsClosed() && err == nil
	}, (time.Second * 5), time.Second, "router is not closed")
}

func TestCloseOneSubscriber(t *testing.T) {
	conf := &config_models.Pubsub{}
	conf.SetDefaults()
	config.Set(conf)
	logger := logger_mocks.NewLoggerer(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything)
	logger.On("Infof", mock.Anything, mock.Anything, mock.Anything)
	pubSuber := newMockRedisManager(logger, func(e error) {})
	err := pubSuber.Connect(conf.Url, conf.User, conf.Password)
	assert.Nil(t, err)
	assert.True(t, pubSuber.router.IsRunning())
	ctx1, cancel := context.WithCancel(context.Background())
	ctx2 := context.Background()
	err = pubSuber.Subscribe(ctx1, "test", func(event []byte) error {
		return nil
	})
	assert.Equal(t, 1, len(pubSuber.router.Handlers()))
	assert.Nil(t, err)
	err = pubSuber.Subscribe(ctx2, "test1", func(event []byte) error {
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(pubSuber.router.Handlers()))
	time.Sleep(time.Second * 2)
	cancel()
	assert.Eventually(t, func() bool {
		return len(pubSuber.router.Handlers()) == 1 && !pubSuber.router.IsClosed() && err == nil
	}, (time.Second * 5), time.Second, "router is not closed")
}

func TestCloseRouterNoSubscriber(t *testing.T) {
	conf := &config_models.Pubsub{}
	conf.SetDefaults()
	config.Set(conf)
	logger := logger_mocks.NewLoggerer(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything)
	logger.On("Infof", mock.Anything, mock.Anything, mock.Anything)
	pubSuber := newMockRedisManager(logger, func(e error) {})
	err := pubSuber.Connect(conf.Url, conf.User, conf.Password)
	assert.Nil(t, err)
	assert.True(t, pubSuber.router.IsRunning())
	pubSuber.Close()
	assert.Eventually(t, func() bool {
		return len(pubSuber.router.Handlers()) == 0 && pubSuber.router.IsClosed() && err == nil
	}, (time.Second * 5), time.Second, "router is not closed")
}

func TestCloseAllSubscribersClosesRouter(t *testing.T) {
	conf := &config_models.Pubsub{}
	conf.SetDefaults()
	config.Set(conf)
	logger := logger_mocks.NewLoggerer(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything)
	logger.On("Infof", mock.Anything, mock.Anything, mock.Anything)
	logger.On("Errorf", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	pubSuber := newMockRedisManager(logger, func(e error) {})
	err := pubSuber.Connect(conf.Url, conf.User, conf.Password)
	assert.Nil(t, err)
	assert.True(t, pubSuber.router.IsRunning())
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	err = pubSuber.Subscribe(ctx1, "test", func(event []byte) error {
		return nil
	})
	assert.Equal(t, 1, len(pubSuber.router.Handlers()))
	assert.Nil(t, err)
	err = pubSuber.Subscribe(ctx2, "test1", func(event []byte) error {
		return nil
	})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(pubSuber.router.Handlers()))
	time.Sleep(time.Second * 2)
	cancel1()
	cancel2()
	assert.Eventually(t, func() bool {
		return len(pubSuber.router.Handlers()) == 0 && pubSuber.router.IsClosed() && err == nil
	}, (time.Second * 5), time.Second, "router is not closed")
}

func TestConfigureRouterWithDeadLetterQueue(t *testing.T) {
	var psConf = &config_models.Pubsub{}
	psConf.SetDefaults()
	psConf.SubscribeRetries = 2
	psConf.ServiceName = "test-service"
	config.Set(psConf)
	ctx := context.Background()
	errCounter := 0
	dlqCounter := 0

	logger := logger_mocks.NewLoggerer(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything)
	logger.On("Infof", mock.Anything, mock.Anything, mock.Anything)
	logger.On("Errorf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		errCounter++
	})
	pubSuber := newMockRedisManager(logger, func(e error) {})
	err := pubSuber.Connect(psConf.Url, psConf.User, psConf.Password)
	assert.Nil(t, err)
	assert.True(t, pubSuber.router.IsRunning())
	err = pubSuber.Subscribe(ctx, "test", func(event []byte) error {
		return fmt.Errorf("throwing error")
	})
	assert.NoError(t, err)
	err = pubSuber.Subscribe(ctx, pevent.GetDlqTopicName(), func(event []byte) error {
		var dlqRunaiEvent pevent.Event[pevent.DlqMessage]
		err := json.Unmarshal(event, &dlqRunaiEvent)
		assert.NoError(t, err)
		assert.Equal(t, "throwing error", dlqRunaiEvent.Payload.Error)
		assert.Equal(t, "test", dlqRunaiEvent.Payload.OriginTopic)
		assert.Equal(t, "test-service", dlqRunaiEvent.Payload.Consumer)
		assert.Equal(t, "test body", string(dlqRunaiEvent.Payload.Body))
		dlqCounter++
		return nil
	})
	assert.NoError(t, err)
	_ = pubSuber.Publish("test", []byte("test body"))
	assert.Eventually(t, func() bool {
		return errCounter == 2 && dlqCounter == 1
	}, (time.Second * 5), time.Second, "router middleware not configured correctly")
}

func TestConfigureRouterWithOutDeadLetterQueue(t *testing.T) {
	var psConf = &config_models.Pubsub{}
	psConf.SetDefaults()
	psConf.SubscribeRetries = 5
	psConf.EnableDLQ = false
	psConf.ServiceName = "test"
	config.Set(psConf)
	ctx := context.Background()
	errCounter := 0
	dlqCounter := 0
	logger := logger_mocks.NewLoggerer(t)
	logger.On("Debugf", mock.Anything, mock.Anything, mock.Anything)
	logger.On("Infof", mock.Anything, mock.Anything, mock.Anything)
	logger.On("Errorf", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		errCounter++
	})
	pubSuber := newMockRedisManager(logger, func(e error) {})
	err := pubSuber.Connect(psConf.Url, psConf.User, psConf.Password)
	assert.Nil(t, err)
	assert.True(t, pubSuber.router.IsRunning())
	err = pubSuber.Subscribe(ctx, "test", func(event []byte) error {
		return fmt.Errorf("throwing error")
	})
	assert.NoError(t, err)
	err = pubSuber.Subscribe(ctx, "test_dlq", func(event []byte) error {
		dlqCounter++
		return nil
	})
	assert.NoError(t, err)
	_ = pubSuber.Publish("test", []byte("test"))
	assert.Eventually(t, func() bool {
		return errCounter == 5 && dlqCounter == 0
	}, (time.Second * 10), time.Second, "router middleware not configured correctly")
}

func TestPublish(t *testing.T) {
	var psConf = &config_models.Pubsub{}
	psConf.PublishRetries = 2
	psConf.ServiceName = "test"
	config.SetWithDefaults(psConf)

	publisherMock := mocks.NewPublisher(t)
	publisherMock.On("Publish", mock.Anything, mock.Anything).Return(nil).Once()

	pubsuber := &redisPubSuber{
		publisher: publisherMock,
	}

	err := pubsuber.Publish("test-topic", []byte("test payload"))
	assert.NoError(t, err)
}

func TestPublish_FailedToPublishOnFirstTry(t *testing.T) {
	var psConf = &config_models.Pubsub{}
	psConf.PublishRetries = 2
	psConf.ServiceName = "test"
	config.SetWithDefaults(psConf)

	loggerMock := logger_mocks.NewLoggerer(t)
	loggerMock.On("Warnf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once()

	publisherMock := mocks.NewPublisher(t)
	publisherMock.On("Publish", mock.Anything, mock.Anything).Return(fmt.Errorf("failed")).Once()
	publisherMock.On("Publish", mock.Anything, mock.Anything).Return(nil).Once()

	pubsuber := &redisPubSuber{
		publisher: publisherMock,
		logger:    NewLogger(loggerMock),
	}

	err := pubsuber.Publish("test-topic", []byte("test payload"))
	assert.NoError(t, err)
}

func TestPublish_AllRetriesFailed(t *testing.T) {
	var psConf = &config_models.Pubsub{}
	psConf.PublishRetries = 2
	psConf.ServiceName = "test"
	config.SetWithDefaults(psConf)

	loggerMock := logger_mocks.NewLoggerer(t)
	loggerMock.On("Warnf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Times(3)

	publisherMock := mocks.NewPublisher(t)
	publisherMock.On("Publish", mock.Anything, mock.Anything).Return(fmt.Errorf("failed")).Times(3)

	pubsuber := &redisPubSuber{
		publisher: publisherMock,
		logger:    NewLogger(loggerMock),
	}

	err := pubsuber.Publish("test-topic", []byte("test payload"))
	assert.EqualError(t, err, "failed")
}
