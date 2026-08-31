/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/run-ai/runai/runai-common-packages/config"
	config_models "github.com/run-ai/runai/runai-common-packages/runaipubsub/config"
	kafka_mocks "github.com/run-ai/runai/runai-common-packages/runaipubsub/kafka/mocks"
	logger_mocks "github.com/run-ai/runai/runai-common-packages/runaipubsub/logger/mocks"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/pevent"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestClose_NoErrors(t *testing.T) {
	reader := kafka_mocks.NewMockkafkaReader(t)
	s := subscriber{
		reader:      reader,
		closeSignal: make(chan struct{}),
	}
	reader.On("Close").Return(nil)
	require.NoError(t, s.Close())
}

func TestClose_Errors(t *testing.T) {
	reader := kafka_mocks.NewMockkafkaReader(t)
	s := subscriber{
		reader:      reader,
		closeSignal: make(chan struct{}),
	}
	reader.On("Close").Return(fmt.Errorf("failed to Close reader"))
	require.Error(t, s.Close())
}

func TestReadMessages_ShouldTerminate(t *testing.T) {
	msgs := make(chan bool, 1)
	ctx := context.Background()
	called := make(chan struct{})
	callback := func(event []byte) error { return nil }
	logger := logger_mocks.NewLoggerer(t)
	loggerStub := func(args mock.Arguments) {
		called <- struct{}{}
	}
	logger.On("Infof", mock.Anything).Run(loggerStub)
	reader := kafka_mocks.NewMockkafkaReader(t)
	reader.On("FetchMessage", ctx).Run(func(args mock.Arguments) {
		<-msgs
	}).Return(kafka.Message{Value: []byte("some data here")}, nil).Maybe()

	s := subscriber{
		reader: reader,
		logger: logger,

		msgInProgress: sync.WaitGroup{},
		closeSignal:   make(chan struct{}),
	}
	s.Subscribe(ctx, callback)
	close(s.closeSignal)
	<-called
}

func TestReadMessages_ShouldCallCallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan []byte)
	logger := setupLoggerMock(t)
	reader, msgs := setupKafkaReaderMock(t, ctx)
	s := subscriber{
		reader: reader,
		logger: logger,

		msgInProgress: sync.WaitGroup{},
		closeSignal:   make(chan struct{}),
		maxHandleTime: 3 * time.Second,
		retryOnErr:    true,
	}
	callback := func(event []byte) error {
		events <- event
		return nil
	}

	msgs <- true
	s.Subscribe(ctx, callback)
	event := <-events
	cancel()
	close(s.closeSignal)
	require.Equal(t, string(event), "some data here")
}

func TestReadMessages_ShouldRetryIfCallbackFail(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan []byte)
	logger := setupLoggerMock(t)
	reader, msgs := setupKafkaReaderMock(t, ctx)
	s := subscriber{
		reader: reader,
		logger: logger,

		msgInProgress: sync.WaitGroup{},
		closeSignal:   make(chan struct{}),
		retries:       3,
		maxHandleTime: 3 * time.Second,
		retryOnErr:    true,
	}
	retry := 0
	callback := func(event []byte) error {
		events <- event
		retry++
		if retry > 2 {
			return nil
		}
		return fmt.Errorf("failed :(")
	}

	msgs <- true
	s.Subscribe(ctx, callback)
	<-events
	<-events
	event := <-events
	cancel()
	close(s.closeSignal)
	require.Equal(t, string(event), "some data here")
}

func TestReadMessages_FailedProcessing_DlqEnabled(t *testing.T) {
	conf := config.Get[*config_models.Pubsub]()
	conf.ServiceName = "test-service"
	conf.EnableDLQ = true
	config.Set(conf)
	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan []byte)

	var capturedMsg []byte
	publisher := kafka_mocks.NewMockpublisherer(t)
	publisher.On("Publish", pevent.GetDlqTopicName(), mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		capturedMsg = args.Get(1).([]byte) // payload is the second arg
	})
	dlqHandler := newDLQHandler(publisher)

	logger := setupLoggerMock(t)
	reader, msgs := setupKafkaReaderMock(t, ctx)
	s := subscriber{
		reader:        reader,
		logger:        logger,
		msgInProgress: sync.WaitGroup{},
		closeSignal:   make(chan struct{}),
		retries:       3,
		maxHandleTime: 3 * time.Second,
		retryOnErr:    true,
		dlqHandler:    dlqHandler,
		topic:         "test-topic",
	}
	retry := 0
	callback := func(event []byte) error {
		events <- event
		retry++
		return fmt.Errorf("failed")
	}

	msgs <- true
	s.Subscribe(ctx, callback)
	<-events
	<-events
	event := <-events
	require.Equal(t, string(event), "some data here")

	require.Eventually(t, func() bool {
		return capturedMsg != nil
	}, 5*time.Second, 100*time.Millisecond)

	cancel()
	close(s.closeSignal)

	var dlqEvent pevent.Event[pevent.DlqMessage]
	require.NoError(t, json.Unmarshal(capturedMsg, &dlqEvent))
	require.Equal(t, string(dlqEvent.Payload.Body), "some data here")
	require.Equal(t, dlqEvent.Payload.OriginTopic, "test-topic")
	require.Equal(t, dlqEvent.Payload.Consumer, "test-service")
}

func TestReadMessages_FailedProcessing_DlqDisabled(t *testing.T) {
	conf := config.Get[*config_models.Pubsub]()
	conf.ServiceName = "test-service"
	conf.EnableDLQ = false
	config.Set(conf)
	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan []byte)

	publisher := kafka_mocks.NewMockpublisherer(t)
	dlqHandler := newDLQHandler(publisher)

	logger := setupLoggerMock(t)
	reader, msgs := setupKafkaReaderMock(t, ctx)
	s := subscriber{
		reader:        reader,
		logger:        logger,
		msgInProgress: sync.WaitGroup{},
		closeSignal:   make(chan struct{}),
		retries:       3,
		maxHandleTime: 3 * time.Second,
		retryOnErr:    true,
		dlqHandler:    dlqHandler,
		topic:         "test-topic",
	}
	retry := 0
	callback := func(event []byte) error {
		events <- event
		retry++
		if retry > 2 {
			return nil
		}
		return fmt.Errorf("failed")
	}

	msgs <- true
	s.Subscribe(ctx, callback)
	<-events
	<-events
	ev := <-events
	require.Equal(t, string(ev), "some data here")
	require.Never(t, func() bool {
		return !publisher.AssertNotCalled(t, "Publish", pevent.GetDlqTopicName(), mock.Anything)
	}, 3*time.Second, 1*time.Second)
	cancel()
	close(s.closeSignal)
}

func TestReadMessages_ShouldNotRetryIfCallbackFail(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan []byte)
	logger := setupLoggerMock(t)
	reader, msgs := setupKafkaReaderMock(t, ctx)
	s := subscriber{
		reader: reader,
		logger: logger,

		msgInProgress: sync.WaitGroup{},
		closeSignal:   make(chan struct{}),
		retries:       3,
		maxHandleTime: 3 * time.Second,
		retryOnErr:    false,
	}
	retry := 0
	callback := func(event []byte) error {
		events <- event
		retry++
		if retry > 2 {
			return nil
		}
		return fmt.Errorf("failed :(")
	}

	msgs <- true
	s.Subscribe(ctx, callback)
	event := <-events
	cancel()
	close(s.closeSignal)
	require.Equal(t, len(events), 0)
	require.Equal(t, string(event), "some data here")
}

func setupKafkaReaderMock(t *testing.T, ctx context.Context) (*kafka_mocks.MockkafkaReader, chan bool) {
	msgs := make(chan bool, 1)

	reader := kafka_mocks.NewMockkafkaReader(t)
	reader.On("CommitMessages", ctx, mock.Anything).Maybe().Return(nil)
	reader.On("FetchMessage", ctx).Run(func(args mock.Arguments) {
		<-msgs
	}).Return(kafka.Message{Value: []byte("some data here")}, nil)

	return reader, msgs
}

func setupLoggerMock(t *testing.T) *logger_mocks.Loggerer {
	logger := logger_mocks.NewLoggerer(t)
	logger.On("Infof", mock.Anything).Maybe()
	logger.On("Warnf", mock.Anything, mock.Anything).Maybe()
	return logger
}
