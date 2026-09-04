/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kafka

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	config_models "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/logger"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type kafkaWriter interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type publisher struct {
	writer kafkaWriter

	logger logger.Loggerer

	retries  int
	interval time.Duration
	inWrite  sync.WaitGroup
	stopCh   chan struct{}
}

func newPublisher(urls []string, dialer *kafka.Dialer, logger logger.Loggerer) *publisher {
	cfg := config.Get[*config_models.Pubsub]()

	return &publisher{
		writer:   createWriter(urls, dialer),
		logger:   logger,
		retries:  cfg.PublishRetries,
		interval: time.Duration(cfg.PublishRetryInterval) * time.Second,
		inWrite:  sync.WaitGroup{},
		stopCh:   make(chan struct{}),
	}
}

func (p *publisher) Publish(topic string, payload []byte) error {
	p.inWrite.Add(1)
	defer p.inWrite.Done()

	select {
	case <-p.stopCh:
		return errors.New("kafka publisher closed")
	default:
		return p.publish(topic, payload)
	}

}

func (p *publisher) Close() error {
	close(p.stopCh)
	p.inWrite.Wait()
	return p.writer.Close()
}

func createWriter(urls []string, dialer *kafka.Dialer) *kafka.Writer {
	cfg := config.Get[*config_models.Pubsub]()
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      urls,
		Balancer:     &kafka.LeastBytes{},
		MaxAttempts:  1,
		WriteTimeout: time.Duration(cfg.PublishRetryInterval) * time.Second,
		Dialer:       dialer,
	})
	writer.AllowAutoTopicCreation = true
	return writer
}

func messageWithUniqueKey(topic string, payload []byte) *kafka.Message {
	return &kafka.Message{
		Topic: topic,
		Key:   generateMessageId(),
		Value: payload,
	}
}

// segmentio/kafka-go does not have a built-in enable.idempotence, key must be added to message.
// Kafka ensures that messages with the same key always go to the same partition.
// If a message is retried, it will overwrite the previous one if the consumer is configured correctly.
func generateMessageId() []byte {
	return []byte(uuid.NewString())
}

func (p *publisher) publish(topic string, payload []byte) error {
	var err error
	msg := messageWithUniqueKey(topic, payload)
	for i := 0; i <= p.retries; i++ {
		err = p.writer.WriteMessages(context.Background(), *msg)
		if err == nil {
			break
		}

		p.logger.Warnf("Failed to publish message to topic %s (attempt %d/%d): %v", topic, i+1, p.retries+1, err)
		if i < p.retries && p.interval > 0 {
			time.Sleep(p.interval)
		}
	}
	return err
}
