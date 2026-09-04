/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	config_models "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/pevent"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
)

func (p *redisPubSuber) runAndWaitForRouter(ctx context.Context) error {
	go p.runRouter(ctx)
	err := p.waitForRouterToStartListening()
	if err != nil {
		return err
	}
	return nil
}

func (p *redisPubSuber) runRouter(ctx context.Context) {
	go func() {
		err := p.router.Run(ctx)
		if err != nil {
			p.logger.Error("failed to run redis subscriber: %v", err, watermill.LogFields{})
		}
	}()
	<-ctx.Done()
	err := p.router.Close()
	if err != nil {
		p.logger.Error("failed to close router: %v", err, watermill.LogFields{})
	}
}

func (p *redisPubSuber) waitForRouterToStartListening() error {
	connectTimeout := config.Get[*config_models.Pubsub]().ConnectTimeoutSeconds
	timeout := time.Now().Add(time.Duration(connectTimeout) * time.Second)
	for time.Now().Before(timeout) {
		if p.router.IsRunning() {
			return nil
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fmt.Errorf("failed to start router after %d seconds", connectTimeout)
}

func (p *redisPubSuber) subscriberCallback(callback func(event []byte) error) func(msg *message.Message) error {
	return func(msg *message.Message) error {
		return callback(msg.Payload)
	}
}

func (p *redisPubSuber) configureRouter() (*message.Router, error) {
	router, err := message.NewRouter(message.RouterConfig{}, p.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to configure redis subscriber: %v", err)
	}
	err = p.configureMiddleware(router)
	if err != nil {
		return nil, err
	}
	return router, nil
}

func (p *redisPubSuber) configureMiddleware(router *message.Router) error {
	cfg := config.Get[*config_models.Pubsub]()
	if cfg.EnableDLQ {
		router.AddMiddleware(p.dlqMiddleware())
	} else {
		router.AddMiddleware(p.ackWhenMessageCompletedMiddleware())
	}
	if cfg.SubscribeRetries > 0 {
		router.AddMiddleware(
			middleware.Retry{
				MaxRetries:      cfg.SubscribeRetries,
				Multiplier:      1,
				InitialInterval: time.Duration(cfg.PublishRetryInterval) * time.Second,
				MaxInterval:     time.Duration(cfg.PublishRetryInterval) * time.Second,
				Logger:          p.logger,
			}.Middleware)
	}
	router.AddMiddleware(middleware.Recoverer)
	return nil
}

// This is a middleware that ack the message when it is completed, completion is defined as completion of
// the handler function and any other middleware that is added after this middleware
// This handler is useful when we don't want to send the message to DLQ on error, we retry and when retry is exhausted
// we ack the message
func (p *redisPubSuber) ackWhenMessageCompletedMiddleware() func(h message.HandlerFunc) message.HandlerFunc {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(message *message.Message) ([]*message.Message, error) {
			_, _ = h(message)
			message.Ack()
			return nil, nil
		}
	}
}

func (p *redisPubSuber) dlqMiddleware() func(h message.HandlerFunc) message.HandlerFunc {
	return func(h message.HandlerFunc) message.HandlerFunc {
		return func(msg *message.Message) ([]*message.Message, error) {
			messages, err := h(msg)
			if err == nil {
				return messages, nil
			}

			dlqMsg := pevent.NewDlq(err.Error(), message.SubscribeTopicFromCtx(msg.Context()), []byte(msg.Payload))
			payload, err := json.Marshal(dlqMsg)
			if err != nil {
				return messages, fmt.Errorf("failed to marshal DLQ message: %w", err)
			}
			err = p.Publish(pevent.GetDlqTopicName(), payload)
			return messages, err
		}
	}
}
