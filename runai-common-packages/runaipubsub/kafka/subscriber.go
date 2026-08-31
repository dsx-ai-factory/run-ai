/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kafka

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/run-ai/runai/runai-common-packages/config"
	config_models "github.com/run-ai/runai/runai-common-packages/runaipubsub/config"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/logger"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/pevent"
	"github.com/segmentio/kafka-go"
)

type kafkaReader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type subscriber struct {
	reader        kafkaReader
	logger        logger.Loggerer
	dlqHandler    dlqHandlerer
	topic         string
	msgInProgress sync.WaitGroup
	closeSignal   chan struct{}
	retries       int
	maxHandleTime time.Duration
	retryOnErr    bool
}

func newSubscriber(urls []string, dialer *kafka.Dialer, topic string, logger logger.Loggerer, dlqHandler dlqHandlerer) *subscriber {
	cfg := config.Get[*config_models.Pubsub]()

	return &subscriber{
		reader:        createReader(urls, dialer, topic, logger),
		logger:        logger,
		dlqHandler:    dlqHandler,
		topic:         topic,
		msgInProgress: sync.WaitGroup{},
		closeSignal:   make(chan struct{}),
		retries:       cfg.SubscribeRetries,
		maxHandleTime: time.Duration(cfg.WaitForAckSec) * time.Second,
		retryOnErr:    cfg.RetryOnReject,
	}
}

func (s *subscriber) Subscribe(ctx context.Context, callback func(event []byte) error) {
	go s.fetchAndHandleMessages(ctx, callback)
}

func (s *subscriber) Close() error {
	close(s.closeSignal)
	err := s.reader.Close()
	s.msgInProgress.Wait()
	return err
}

func (s *subscriber) fetchAndHandleMessages(ctx context.Context, callback func(event []byte) error) {
	for {
		select {
		case <-s.closeSignal:
			s.logger.Infof("Stopping fetchAndHandleMessages due to shutdown signal.")
			return
		default:
			msg, err := s.reader.FetchMessage(ctx)
			if err != nil {
				if !connectionClosed(err) {
					s.logger.Errorf("Kafka fetching error: %+v", err)
				}
				continue
			}

			s.handleMessage(&msg, callback, ctx)
		}
	}
}

func (s *subscriber) handleMessage(msg *kafka.Message, callback func(event []byte) error, ctx context.Context) {
	s.msgInProgress.Add(1)
	defer func() {
		s.msgInProgress.Done()
	}()

	cwr := newCallbackWithRetry(s.retries, callback, s.maxHandleTime, s.retryOnErr, s.logger)
	if err := cwr.run(msg.Value); err != nil && config.Get[*config_models.Pubsub]().EnableDLQ {
		s.sendDlqMessage(pevent.NewDlq(err.Error(), s.topic, msg.Value))
	}
	if err := s.reader.CommitMessages(ctx, *msg); err != nil {
		s.logger.Errorf("failed to commit message at offset: %d: %v\n", msg.Offset, err)
	}
}

func (s *subscriber) sendDlqMessage(msg *pevent.Event[pevent.DlqMessage]) {
	err := s.dlqHandler.sendToDLQ(msg)
	if err != nil {
		s.logger.Errorf("failed to send message to DLQ: %v", err)
	}
}

func createReader(urls []string, dialer *kafka.Dialer, topic string, logger logger.Loggerer) *kafka.Reader {
	cfg := config.Get[*config_models.Pubsub]()
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:       urls,
		Topic:         topic,
		GroupID:       cfg.ServiceName,
		MaxBytes:      100e6, // 100MB
		MaxWait:       1 * time.Second,
		RetentionTime: time.Duration(cfg.ConsumerTimeoutSec) * time.Second,
		MaxAttempts:   cfg.SubscribeRetries,
		StartOffset:   kafka.FirstOffset,
		ErrorLogger:   NewErrorLogger(logger),
		Dialer:        dialer,
	})
}

type callbackWithRetry struct {
	callback      func(event []byte) error
	retries       int
	maxHandleTime time.Duration
	retryOnErr    bool
	logger        logger.Loggerer
}

func newCallbackWithRetry(retries int, callback func(event []byte) error, maxHandleTime time.Duration, retryOnErr bool, logger logger.Loggerer) *callbackWithRetry {
	return &callbackWithRetry{
		callback:      callback,
		retries:       retries,
		maxHandleTime: maxHandleTime,
		retryOnErr:    retryOnErr,
		logger:        logger,
	}
}

func (cwr *callbackWithRetry) run(msg []byte) error {
	var err error

	for retryNum := 0; retryNum <= cwr.retries; retryNum++ {
		err = cwr.runWithDeadline(msg)
		if err == nil {
			return nil
		}

		if !cwr.shouldRetry(err) {
			return err
		}
	}

	return err
}

func (cwr *callbackWithRetry) runWithDeadline(msg []byte) error {
	errCh := make(chan error, 1)
	callbackContext, cancel := context.WithTimeout(context.Background(), cwr.maxHandleTime)
	defer cancel()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errCh <- fmt.Errorf("panic occurred: %v", r)
			}
		}()
		errCh <- cwr.callback(msg)
	}()

	select {
	case <-callbackContext.Done():
		return context.DeadlineExceeded
	case err := <-errCh:
		return err
	}
}

func (cwr *callbackWithRetry) shouldRetry(err error) bool {
	// We always retry if the error is a context.DeadlineExceeded
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	return cwr.retryOnErr
}

func connectionClosed(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
