/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nats

import (
	"context"
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/run-ai/runai/runai-common-packages/config"
	config_models "github.com/run-ai/runai/runai-common-packages/runaipubsub/config"
	logger_mocks "github.com/run-ai/runai/runai-common-packages/runaipubsub/logger/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/nats/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func initPubsuberForTesting() (*pubsuber, *mocks.JetStream) {
	cfg := &config_models.Pubsub{
		Type:                 "nats",
		Url:                  "test",
		MaxEntries:           100,
		MaxMessageAgeMinutes: 60,
	}
	config.Set(cfg)
	jsMock := &mocks.JetStream{}
	return &pubsuber{
		streams:   sync.Map{},
		jetStream: jsMock,
	}, jsMock
}

func TestPublish(t *testing.T) {
	subject, jsMock := initPubsuberForTesting()
	expectedStream := &mocks.Stream{}
	jsMock.On("CreateOrUpdateStream", mock.Anything, mock.Anything).Return(expectedStream, nil)
	jsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	err := subject.Publish("test", []byte("test"))
	assert.NoError(t, err)
	jsMock.AssertNumberOfCalls(t, "CreateOrUpdateStream", 1)
	jsMock.AssertNumberOfCalls(t, "Publish", 1)
}

func TestPublish_MultipleCalls(t *testing.T) {
	subject, jsMock := initPubsuberForTesting()
	expectedStream := &mocks.Stream{}
	jsMock.On("CreateOrUpdateStream", mock.Anything, mock.Anything).Return(expectedStream, nil)
	jsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	err := subject.Publish("test", []byte("test"))
	assert.NoError(t, err)
	subject.Publish("test", []byte("test1"))
	jsMock.AssertNumberOfCalls(t, "CreateOrUpdateStream", 1)
	jsMock.AssertNumberOfCalls(t, "Publish", 2)
}

func TestPublish_MultipleCalls_DifferentStreams(t *testing.T) {
	subject, jsMock := initPubsuberForTesting()
	expectedStream := &mocks.Stream{}
	jsMock.On("CreateOrUpdateStream", mock.Anything, mock.Anything).Return(expectedStream, nil)
	jsMock.On("Publish", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	err := subject.Publish("test", []byte("test"))
	assert.NoError(t, err)
	subject.Publish("test1", []byte("test1"))
	jsMock.AssertNumberOfCalls(t, "CreateOrUpdateStream", 2)
	jsMock.AssertNumberOfCalls(t, "Publish", 2)
}

func TestCreateStream_Cache(t *testing.T) {
	subject, jsMock := initPubsuberForTesting()
	expectedStream := &mocks.Stream{}
	jsMock.On("CreateOrUpdateStream", mock.Anything, mock.Anything).Return(expectedStream, nil)
	res, err := subject.createStreamWithCache(context.Background(), "test")
	assert.NoError(t, err)
	assert.Equal(t, expectedStream, res)
	subject.createStreamWithCache(context.Background(), "test")
	jsMock.AssertNumberOfCalls(t, "CreateOrUpdateStream", 1)
}

func TestCreateStream_Cache_DifferentStreams(t *testing.T) {
	subject, jsMock := initPubsuberForTesting()
	expectedStream := &mocks.Stream{}
	jsMock.On("CreateOrUpdateStream", mock.Anything, mock.Anything).Return(expectedStream, nil)
	res, err := subject.createStreamWithCache(context.Background(), "test")
	assert.NoError(t, err)
	assert.Equal(t, expectedStream, res)
	subject.createStreamWithCache(context.Background(), "test1")
	jsMock.AssertNumberOfCalls(t, "CreateOrUpdateStream", 2)
}

func TestCreateStream_NoCache(t *testing.T) {
	subject, jsMock := initPubsuberForTesting()
	expectedStream := &mocks.Stream{}
	jsMock.On("CreateOrUpdateStream", mock.Anything, mock.Anything).Return(expectedStream, nil)
	res, err := subject.createStreamWithCache(context.Background(), "test")
	assert.NoError(t, err)
	assert.Equal(t, expectedStream, res)
	subject.createOrUpdateStream(context.Background(), "test")
	jsMock.AssertNumberOfCalls(t, "CreateOrUpdateStream", 2)
}

func TestSubscribe(t *testing.T) {
	subject, jsMock := initPubsuberForTesting()
	expectedStream := &mocks.Stream{}
	jsMock.On("CreateOrUpdateStream", mock.Anything, mock.Anything).Return(expectedStream, nil)
	expectedConsumer := &mocks.Consumer{}
	expectedConsumer.On("Consume", mock.Anything, mock.Anything).Return(nil, nil)
	expectedStream.On("CreateOrUpdateConsumer", mock.Anything, mock.Anything).Return(expectedConsumer, nil)
	err := subject.Subscribe(context.Background(), "test", func(msg []byte) error {
		return nil
	})
	assert.NoError(t, err)
	jsMock.AssertNumberOfCalls(t, "CreateOrUpdateStream", 1)
	expectedStream.AssertNumberOfCalls(t, "CreateOrUpdateConsumer", 1)
	expectedConsumer.AssertNumberOfCalls(t, "Consume", 1)
}

func TestSubscribe_MultipleCalls(t *testing.T) {
	subject, jsMock := initPubsuberForTesting()
	expectedStream := &mocks.Stream{}
	jsMock.On("CreateOrUpdateStream", mock.Anything, mock.Anything).Return(expectedStream, nil)
	expectedConsumer := &mocks.Consumer{}
	expectedConsumer.On("Consume", mock.Anything, mock.Anything).Return(nil, nil)
	expectedStream.On("CreateOrUpdateConsumer", mock.Anything, mock.Anything).Return(expectedConsumer, nil)
	err := subject.Subscribe(context.Background(), "test", func(msg []byte) error {
		return nil
	})
	assert.NoError(t, err)
	subject.Subscribe(context.Background(), "test", func(msg []byte) error {
		return nil
	})
	jsMock.AssertNumberOfCalls(t, "CreateOrUpdateStream", 2)
	expectedStream.AssertNumberOfCalls(t, "CreateOrUpdateConsumer", 2)
	expectedConsumer.AssertNumberOfCalls(t, "Consume", 2)
}

func TestSubscribeAsync(t *testing.T) {
	subject, jsMock := initPubsuberForTesting()
	done := make(chan bool)
	expectedStream := &mocks.Stream{}
	expectedConsumer := &mocks.Consumer{}
	jsMock.On("CreateOrUpdateStream", mock.Anything, mock.Anything).Return(expectedStream, nil)
	expectedStream.On("CreateOrUpdateConsumer", mock.Anything, mock.Anything).Return(expectedConsumer, nil)
	expectedConsumer.On("Consume", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		done <- true
	}).Return(nil, nil)

	subject.SubscribeAsync(context.Background(), "test", func(msg []byte) error {
		return nil
	}, make(chan error))

	<-done
	jsMock.AssertNumberOfCalls(t, "CreateOrUpdateStream", 1)
	expectedStream.AssertNumberOfCalls(t, "CreateOrUpdateConsumer", 1)
	expectedConsumer.AssertNumberOfCalls(t, "Consume", 1)
}

func TestSubscribeAsync_MultipleCalls(t *testing.T) {
	done := make(chan bool)
	subject, jsMock := initPubsuberForTesting()
	expectedStream := &mocks.Stream{}
	expectedConsumer := &mocks.Consumer{}
	jsMock.On("CreateOrUpdateStream", mock.Anything, mock.Anything).Return(expectedStream, nil)
	expectedStream.On("CreateOrUpdateConsumer", mock.Anything, mock.Anything).Return(expectedConsumer, nil)
	expectedConsumer.On("Consume", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		done <- true
	}).Return(nil, nil)

	subject.SubscribeAsync(context.Background(), "test", func(msg []byte) error {
		return nil
	}, make(chan error))
	subject.SubscribeAsync(context.Background(), "test", func(msg []byte) error {
		return nil
	}, make(chan error))

	<-done
	<-done
	subject.SubscribeAsync(context.Background(), "test", func(msg []byte) error {
		return nil
	}, make(chan error))
	jsMock.AssertNumberOfCalls(t, "CreateOrUpdateStream", 2)
	expectedStream.AssertNumberOfCalls(t, "CreateOrUpdateConsumer", 2)
	expectedConsumer.AssertNumberOfCalls(t, "Consume", 2)
}

func TestSubscribeAsync_FailToCreateStream(t *testing.T) {
	subject, jsMock := initPubsuberForTesting()
	done := make(chan bool)
	errs := make(chan error, 1)
	expectedStream := &mocks.Stream{}
	expectedConsumer := &mocks.Consumer{}

	logger := logger_mocks.NewLoggerer(t)
	subject.logger = logger
	jsMock.On("CreateOrUpdateStream", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to create stream")).Once()
	logger.On("Warnf", "Failed to init subscription to stream %s (attempt %d/%d): %v", "test", mock.Anything, math.MaxInt16, mock.Anything)
	jsMock.On("CreateOrUpdateStream", mock.Anything, mock.Anything).Return(expectedStream, nil).Once()
	expectedStream.On("CreateOrUpdateConsumer", mock.Anything, mock.Anything).Return(expectedConsumer, nil)
	expectedConsumer.On("Consume", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		done <- true
	}).Return(nil, nil)

	subject.SubscribeAsync(context.Background(), "test", func(msg []byte) error {
		return nil
	}, errs)

	<-done
	require.Error(t, <-errs)
	jsMock.AssertNumberOfCalls(t, "CreateOrUpdateStream", 2)
	expectedStream.AssertNumberOfCalls(t, "CreateOrUpdateConsumer", 1)
	expectedConsumer.AssertNumberOfCalls(t, "Consume", 1)
}

func TestSubscribeAsync_FailToConsumer(t *testing.T) {
	subject, jsMock := initPubsuberForTesting()
	done := make(chan bool)
	errs := make(chan error, 1)
	expectedStream := &mocks.Stream{}
	expectedConsumer := &mocks.Consumer{}

	jsMock.On("CreateOrUpdateStream", mock.Anything, mock.Anything).Return(expectedStream, nil).Once()
	expectedStream.On("CreateOrUpdateConsumer", mock.Anything, mock.Anything).Return(expectedConsumer, nil).Once()
	expectedConsumer.On("Consume", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		done <- true
	}).Return(nil, fmt.Errorf("failed to start comsume, bad options, no reasons for retrying"))

	subject.SubscribeAsync(context.Background(), "test", func(msg []byte) error {
		return nil
	}, errs)

	<-done
	require.Error(t, <-errs)
	jsMock.AssertNumberOfCalls(t, "CreateOrUpdateStream", 1)
	expectedStream.AssertNumberOfCalls(t, "CreateOrUpdateConsumer", 1)
	expectedConsumer.AssertNumberOfCalls(t, "Consume", 1)
}
