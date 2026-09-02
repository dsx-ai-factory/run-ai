/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kafka

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	kafka_mocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/kafka/mocks"
	logger_mocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/logger/mocks"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var topic, msg, ctx = uuid.NewString(), uuid.NewString(), context.Background()
var subscribeCallback = func(event []byte) error { return nil }

func TestPublish_NotInitialized(t *testing.T) {
	ps := pubSuber{}
	require.Error(t, fmt.Errorf("kafka publisher is not initialized"), ps.Publish(topic, []byte(msg)))
}

func TestPublish_FailedToCreateTopic(t *testing.T) {
	p := kafka_mocks.NewMockpublisherer(t)
	tc := kafka_mocks.NewMocktopicManagerer(t)
	ps := pubSuber{
		publisher:    p,
		topicManager: tc,
	}
	tc.On("CreateIfNotExist", topic).Return(fmt.Errorf("failed to crerate topic"))
	require.Error(t, fmt.Errorf("failed to crerate topic"), ps.Publish(topic, []byte(msg)))
}

func TestPublish_PublishFailed(t *testing.T) {
	p := kafka_mocks.NewMockpublisherer(t)
	tc := kafka_mocks.NewMocktopicManagerer(t)
	tc.On("CreateIfNotExist", topic).Return(nil)
	ps := pubSuber{
		topicManager: tc,
		publisher:    p,
	}
	p.On("Publish", topic, []byte(msg)).Return(fmt.Errorf("failed to pubsh msg"))
	require.Error(t, fmt.Errorf("failed to pubsh msg"), ps.Publish(topic, []byte(msg)))
}

func TestPublish_Succeed(t *testing.T) {
	p := kafka_mocks.NewMockpublisherer(t)
	tc := kafka_mocks.NewMocktopicManagerer(t)
	tc.On("CreateIfNotExist", topic).Return(nil)
	ps := pubSuber{
		topicManager: tc,
		publisher:    p,
	}
	p.On("Publish", topic, []byte(msg)).Return(nil)
	require.Nil(t, ps.Publish(topic, []byte(msg)))
}

func TestSubscriber_FailedToCreateTopic(t *testing.T) {
	tc := kafka_mocks.NewMocktopicManagerer(t)
	l := logger_mocks.NewLoggerer(t)
	ps := pubSuber{
		subscribers:  sync.Map{},
		topicManager: tc,
		logger:       l,
	}
	l.On("Warnf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	tc.On("CreateIfNotExist", topic).Return(fmt.Errorf("failed to crerate topic"))
	require.Error(t, fmt.Errorf("failed to crerate topic"), ps.Subscribe(ctx, topic, subscribeCallback))
}

func TestSubscriber_SubscriberFailed(t *testing.T) {
	s := kafka_mocks.NewMocksubscriberer(t)
	tc := kafka_mocks.NewMocktopicManagerer(t)
	l := logger_mocks.NewLoggerer(t)
	ps := pubSuber{
		subscribers:  sync.Map{},
		topicManager: tc,
		logger:       l,
	}
	l.On("Warnf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	ps.subscribers.Store(topic, s)
	tc.On("CreateIfNotExist", topic).Return(fmt.Errorf("failed to crerate topic"))
	require.Error(t, fmt.Errorf("failed to crerate topic"), ps.Subscribe(ctx, topic, subscribeCallback))
}

func TestSubscriber_Succeed(t *testing.T) {
	l := logger_mocks.NewLoggerer(t)
	s := kafka_mocks.NewMocksubscriberer(t)
	tc := kafka_mocks.NewMocktopicManagerer(t)
	ps := pubSuber{
		subscribers:  sync.Map{},
		topicManager: tc,
		logger:       l,
		urls:         []string{"something"},
	}
	ps.subscribers.Store(topic, s)
	tc.On("CreateIfNotExist", topic).Return(nil)
	l.On("Warnf", "already subscribed to topic: %s", topic).Return()
	err := ps.Subscribe(ctx, topic, subscribeCallback)
	require.Nil(t, err)
}

func TestSubscribeAsync_Succeed(t *testing.T) {
	l := logger_mocks.NewLoggerer(t)
	s := kafka_mocks.NewMocksubscriberer(t)
	tc := kafka_mocks.NewMocktopicManagerer(t)
	done := make(chan bool)
	ps := pubSuber{
		subscribers:  sync.Map{},
		topicManager: tc,
		logger:       l,
		urls:         []string{"something"},
	}
	ps.subscribers.Store(topic, s)
	tc.On("CreateIfNotExist", topic).Run(func(args mock.Arguments) {
		done <- true
	}).Return(nil)
	l.On("Warnf", "already subscribed to topic: %s", topic).Once()
	ps.SubscribeAsync(ctx, topic, subscribeCallback, make(chan error))
	<-done
}

func TestSubscribeAsync_FailedExitOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	l := logger_mocks.NewLoggerer(t)
	s := kafka_mocks.NewMocksubscriberer(t)
	tc := kafka_mocks.NewMocktopicManagerer(t)
	done := make(chan bool)
	logDone := make(chan bool)
	ps := pubSuber{
		subscribers:  sync.Map{},
		topicManager: tc,
		logger:       l,
		urls:         []string{"something"},
	}
	ps.subscribers.Store(topic, s)
	tc.On("CreateIfNotExist", topic).Run(func(args mock.Arguments) {
		done <- true
	}).Return(fmt.Errorf("failed to create topic"))

	l.On("Warnf", "Failed to create topic %s (attempt %d/%d): %v", topic, mock.Anything, math.MaxInt16, mock.Anything)
	l.On("Infof", "context canceled, not subscribed to topic: %s", topic).Run(func(args mock.Arguments) {
		logDone <- true
	}).Once()
	ps.SubscribeAsync(ctx, topic, subscribeCallback, make(chan error))
	<-done
	cancel()
	<-logDone
}

func TestClose_FailedToClosePublisher(t *testing.T) {
	l := logger_mocks.NewLoggerer(t)
	p := kafka_mocks.NewMockpublisherer(t)
	s := kafka_mocks.NewMocksubscriberer(t)
	ps := pubSuber{
		publisher:   p,
		subscribers: sync.Map{},
		logger:      l,
	}
	ps.subscribers.Store(topic, s)
	l.On("Errorf", "failed to Close publisher %w", mock.Anything).Return()
	p.On("Close").Return(fmt.Errorf("failed to Close"))
	s.On("Close").Return(nil)
	require.Equal(t, errors.Join(fmt.Errorf("failed to Close")), ps.Close())
}

func TestClose_FailedToCloseSubscriber(t *testing.T) {
	l := logger_mocks.NewLoggerer(t)
	p := kafka_mocks.NewMockpublisherer(t)
	s := kafka_mocks.NewMocksubscriberer(t)
	ps := pubSuber{
		publisher:   p,
		subscribers: sync.Map{},
		logger:      l,
	}
	ps.subscribers.Store(topic, s)
	l.On("Errorf", "failed to Close subscriber to topic: %s: %+v", topic, mock.Anything).Return()
	p.On("Close").Return(nil)
	s.On("Close").Return(fmt.Errorf("failed to Close"))
	require.Equal(t, errors.Join(fmt.Errorf("failed to Close")), ps.Close())
}

func TestClose_Succeed(t *testing.T) {
	p := kafka_mocks.NewMockpublisherer(t)
	s := kafka_mocks.NewMocksubscriberer(t)
	ps := pubSuber{
		publisher:   p,
		subscribers: sync.Map{},
	}
	ps.subscribers.Store(topic, s)
	p.On("Close").Return(nil)
	s.On("Close").Return(nil)
	require.NoError(t, ps.Close())
}

func TestGetAuthMechanism_Plain(t *testing.T) {
	config.Set(&Config{
		AuthType: KafkaAuthTypePlain,
	})
	ps := pubSuber{}
	mechanism, err := ps.getAuthMechanism("user", "password")
	require.NoError(t, err)
	expectedMechanism := &plain.Mechanism{
		Username: "user",
		Password: "password",
	}
	assert.NoError(t, err)
	require.Equal(t, expectedMechanism.Name(), mechanism.Name())
}

func TestGetAuthMechanism_Scram(t *testing.T) {
	config.Set(&Config{
		AuthType: KafkaAuthTypeScram,
	})
	ps := pubSuber{}
	mechanism, err := ps.getAuthMechanism("user", "password")
	require.NoError(t, err)
	expectedMechanism, err := scram.Mechanism(scram.SHA512, "user", "password")
	assert.NoError(t, err)
	require.Equal(t, expectedMechanism.Name(), mechanism.Name())
}
