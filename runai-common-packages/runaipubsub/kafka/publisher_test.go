package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	kafka_mocks "github.com/run-ai/runai/runai-common-packages/runaipubsub/kafka/mocks"
	logger_mocks "github.com/run-ai/runai/runai-common-packages/runaipubsub/logger/mocks"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewWriter(t *testing.T) {
	logger := logger_mocks.NewLoggerer(t)
	writer := newPublisher([]string{""}, &kafka.Dialer{}, logger)
	require.NotNil(t, writer.writer)
	require.Equal(t, writer.retries, 0)
	require.NotNil(t, writer.interval)
}

func TestClose_NoError(t *testing.T) {
	kw := kafka_mocks.NewMockkafkaWriter(t)
	kw.On("Close").Return(nil)
	writer := publisher{
		writer: kw,
		stopCh: make(chan struct{}),
	}
	err := writer.Close()
	require.NoError(t, err)
}

func TestClose_Error(t *testing.T) {
	kw := kafka_mocks.NewMockkafkaWriter(t)
	kw.On("Close").Return(fmt.Errorf("failed to Close"))
	writer := publisher{
		writer:  kw,
		inWrite: sync.WaitGroup{},
		stopCh:  make(chan struct{}),
	}
	err := writer.Close()
	require.Error(t, err)
	require.Equal(t, err.Error(), "failed to Close")
}

func TestWriteMessages_NoError(t *testing.T) {
	ctx := context.Background()
	logger := logger_mocks.NewLoggerer(t)
	kw := kafka_mocks.NewMockkafkaWriter(t)
	kw.On("WriteMessages", ctx, mock.Anything).Return(nil)
	writer := publisher{
		writer:   kw,
		logger:   logger,
		retries:  0,
		interval: time.Second,
		inWrite:  sync.WaitGroup{},
		stopCh:   make(chan struct{}),
	}
	err := writer.Publish("some topic", make([]byte, 0))
	require.NoError(t, err)
}

func TestWriteMessages_FailAllAttempts(t *testing.T) {
	ctx := context.Background()
	logger := logger_mocks.NewLoggerer(t)
	logger.On("Warnf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	kw := kafka_mocks.NewMockkafkaWriter(t)
	kw.On("WriteMessages", ctx, mock.Anything).Return(fmt.Errorf("failed to write")).Times(4)
	writer := publisher{
		writer:   kw,
		logger:   logger,
		retries:  3,
		interval: time.Second,
		inWrite:  sync.WaitGroup{},
		stopCh:   make(chan struct{}),
	}
	err := writer.Publish("some topic", make([]byte, 0))
	require.Error(t, err)
	require.Equal(t, err.Error(), "failed to write")
}

func TestWriteMessages_FailAllAttemptsClosed(t *testing.T) {
	ctx := context.Background()
	logger := logger_mocks.NewLoggerer(t)
	kw := kafka_mocks.NewMockkafkaWriter(t)
	logger.On("Warnf", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	kw.On("WriteMessages", ctx, mock.Anything).Return(fmt.Errorf("failed to write")).Times(4)
	kw.On("Close").Return(nil)
	writer := publisher{
		writer:   kw,
		logger:   logger,
		retries:  3,
		interval: time.Second,
		inWrite:  sync.WaitGroup{},
		stopCh:   make(chan struct{}),
	}
	var err error
	called := make(chan struct{})
	go func() {
		err = writer.Publish("some topic", make([]byte, 0))
		called <- struct{}{}
	}()
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, writer.Close())
	<-called
	require.Error(t, err)
	require.Equal(t, err.Error(), "failed to write")
	require.Equal(t, errors.New("kafka publisher closed"), writer.Publish("some topic", make([]byte, 0)))
}

func TestWriteMessages_Closed(t *testing.T) {
	logger := logger_mocks.NewLoggerer(t)
	kw := kafka_mocks.NewMockkafkaWriter(t)
	called := make(chan struct{})
	called2 := make(chan struct{})
	kw.On("Close").Return(nil)
	writer := publisher{
		writer:   kw,
		logger:   logger,
		retries:  0,
		interval: time.Second,
		inWrite:  sync.WaitGroup{},
		stopCh:   make(chan struct{}),
	}
	var err error
	go func() {
		<-called
		err = writer.Publish("some topic", make([]byte, 0))
		called2 <- struct{}{}
	}()
	writer.Close()
	called <- struct{}{}
	<-called2
	require.Equal(t, errors.New("kafka publisher closed"), err)
}
