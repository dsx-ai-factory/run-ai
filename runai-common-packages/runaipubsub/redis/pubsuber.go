/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package redis

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	config_models "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/logger"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/retry"
)

type PubSuberConnection interface {
	connect(logger *Logger, url string, user, password string) (*redis.Client, message.Publisher, message.Subscriber, error)
}

type redisPubSuber struct {
	connection       PubSuberConnection
	client           *redis.Client
	publisher        message.Publisher
	subscriber       message.Subscriber
	router           *message.Router
	routerCancelFunc context.CancelFunc
	logger           *Logger
}

func NewRedisPubSuber(logger logger.Loggerer) *redisPubSuber {
	return &redisPubSuber{
		connection: NewConnection(),
		logger:     NewLogger(logger),
	}
}

func (p *redisPubSuber) Connect(url, user, password string) error {
	ctx := p.initContext()

	var err error
	p.client, p.publisher, p.subscriber, err = p.connection.connect(p.logger, url, user, password)
	if err != nil {
		return err
	}

	p.router, err = p.configureRouter()
	if err != nil {
		return err
	}

	err = p.runAndWaitForRouter(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (p *redisPubSuber) Publish(topic string, payload []byte) error {
	redisMessage := message.NewMessage(watermill.NewUUID(), payload)
	middleware.SetCorrelationID(watermill.NewUUID(), redisMessage)

	return p.publishWithRetry(topic, redisMessage)
}

func (p *redisPubSuber) publishWithRetry(topic string, msg *message.Message) error {
	cfg := config.Get[*config_models.Pubsub]()
	publishRetries := cfg.PublishRetries
	publishInterval := time.Duration(cfg.PublishRetryInterval) * time.Second

	var err error
	for i := 0; i <= publishRetries; i++ {
		err = p.publisher.Publish(topic, msg)
		if err == nil {
			break
		}

		p.logger.Warnf("Failed to publish message to topic %s (attempt %d/%d): %v", topic, i+1, publishRetries+1, err)
		if i < publishRetries && publishInterval > 0 {
			time.Sleep(publishInterval)
		}
	}

	return err
}

func (p *redisPubSuber) Subscribe(ctx context.Context, topic string, callback func(event []byte) error) error {
	cfg := config.Get[*config_models.Pubsub]()
	return p.subscribe(ctx, topic, callback, make(chan error), cfg.SubscriptionInitRetries+1)
}

func (p *redisPubSuber) SubscribeAsync(ctx context.Context, topic string, callback func(event []byte) error, errChan chan error) {
	go p.subscribe(ctx, topic, callback, errChan, math.MaxInt16)
}

func (p *redisPubSuber) Close() error {
	p.routerCancelFunc()
	return nil
}

func (p *redisPubSuber) initContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	p.routerCancelFunc = cancel
	return ctx
}

func (p *redisPubSuber) subscribe(ctx context.Context, topic string, callback func(event []byte) error, errChan chan error, totalAttempts int) error {
	subscriberCallBack := p.subscriberCallback(callback)
	p.router.AddConsumerHandler(topic, topic, p.subscriber, subscriberCallBack)

	if err := p.initSubscriber(ctx, errChan, totalAttempts); err != nil {
		tryAppendToChan(errChan, fmt.Errorf("failed to subscribe to topic: %s, err:%v", topic, err))
		return fmt.Errorf("failed to subscribe to topic: %s, err:%v", topic, err)
	}

	go p.cleanupSubscriber(ctx, topic)
	return nil
}

func (p *redisPubSuber) initSubscriber(ctx context.Context, errChan chan error, totalAttempts int) error {
	err := retry.RetryStreamErrors(
		ctx,
		errChan,
		func(attempt uint) error {
			err := p.router.RunHandlers(ctx)
			if err != nil {
				p.logger.Warnf("Failed to start Redis handlers (attempt %d/%d): %v", attempt+1, totalAttempts, err)
			}
			return err
		},
		totalAttempts,
	)

	return err
}

func (p *redisPubSuber) cleanupSubscriber(ctx context.Context, topic string) {
	conf := config.Get[*config_models.Pubsub]()
	ticker := time.NewTicker(time.Second * time.Duration(conf.CleanerIntervalSeconds))
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			trimmedEntriesCount, err := p.client.XTrimMaxLen(ctx, topic, conf.MaxEntries).Result()
			if err != nil {
				p.logger.Warnf("cleaning up subscriber- failed to trim for topic %s, err:%v", topic, err)
				continue
			}
			p.logger.Debugf("cleaning up subscriber- trimmed %d entries for topic: %s", trimmedEntriesCount, topic)
		}
	}
}

func tryAppendToChan(errChan chan error, err error) {
	select {
	case errChan <- err:
	default:
	}
}
