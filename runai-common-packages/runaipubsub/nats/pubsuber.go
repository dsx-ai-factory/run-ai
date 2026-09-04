/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	config_models "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/logger"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/pevent"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/retry"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type CallbackFunc func(event []byte) error
type pubsuber struct {
	natsConnection  *nats.Conn
	jetStream       jetstream.JetStream
	connCtx         context.Context
	streams         sync.Map
	logger          logger.Loggerer
	consumers       sync.Map
	publishInterval time.Duration
	publishAttempts int
}

func NewPubSuber(logger logger.Loggerer) *pubsuber {
	cfg := config.Get[*config_models.Pubsub]()
	return &pubsuber{
		logger:          logger,
		streams:         sync.Map{},
		consumers:       sync.Map{},
		publishInterval: time.Duration(cfg.PublishRetryInterval) * time.Second,
		publishAttempts: cfg.PublishRetries,
	}
}

func (p *pubsuber) Connect(url, user, password string) error {
	var err error
	p.natsConnection, err = nats.Connect(url, p.getNatsConnectOptions(user, password)...)
	if err != nil {
		return fmt.Errorf("could not connect to NATS, err: %w", err)
	}
	connCtx, cancel := context.WithCancel(context.Background())
	p.connCtx = connCtx

	p.natsConnection.SetDisconnectErrHandler(func(nc *nats.Conn, err error) {
		if err != nil {
			p.logger.Errorf("nats connection closed, err: %w", err)
		}
	})
	p.natsConnection.SetClosedHandler(func(*nats.Conn) {
		cancel()
	})
	p.jetStream, err = jetstream.New(p.natsConnection, func(opt *jetstream.JetStreamOptions) error {

		return nil
	})
	if err != nil {
		return fmt.Errorf("could not create jet stream, err: %w", err)
	}
	return nil
}

func (p *pubsuber) Publish(stream string, payload []byte) error {
	_, err := p.createStreamWithCache(context.Background(), stream)
	if err != nil {
		return fmt.Errorf("failed to create stream %s on publish, err %w", stream, err)
	}
	_, err = p.jetStream.Publish(context.Background(), stream, payload, jetstream.WithRetryWait(p.publishInterval), jetstream.WithRetryAttempts(p.publishAttempts))
	if err != nil {
		return fmt.Errorf("failed to publish message to stream %s, err %w", stream, err)
	}
	return nil
}

func (p *pubsuber) Subscribe(ctx context.Context, stream string, callback func(event []byte) error) error {
	cfg := config.Get[*config_models.Pubsub]()
	totalAttempts := cfg.SubscriptionInitRetries + 1

	return p.subscribe(ctx, stream, callback, make(chan error), totalAttempts)
}

func (p *pubsuber) SubscribeAsync(ctx context.Context, stream string, callback func(event []byte) error, errChan chan error) {
	go p.subscribe(ctx, stream, callback, errChan, math.MaxInt16)
}

func (p *pubsuber) Close() error {
	p.consumers.Range(func(key, value any) bool {
		cc := value.(jetstream.ConsumeContext)
		cc.Drain()
		return true
	})
	if err := p.natsConnection.Drain(); err != nil {
		return fmt.Errorf("could not drain nats connection, err: %w", err)
	}
	// Wait for the connection to be closed
	<-p.connCtx.Done()
	return nil
}

func (p *pubsuber) subscribe(ctx context.Context, stream string, callback CallbackFunc, errChan chan error, totalAttempts int) error {
	cons, err := p.initSubscriber(ctx, stream, totalAttempts, errChan)
	if err != nil {
		tryAppendToChan(errChan, err)
		return fmt.Errorf("failed to initialize subscriber for stream %s: %w", stream, err)
	}

	if err = p.startConsuming(cons, stream, callback); err != nil {
		tryAppendToChan(errChan, err)
	}
	return err
}

func (p *pubsuber) initSubscriber(ctx context.Context, stream string, totalAttempts int, errChan chan error) (jetstream.Consumer, error) {
	var cons jetstream.Consumer
	err := retry.RetryStreamErrors(
		ctx, errChan,
		func(attempt uint) error {
			var err error
			cons, err = p.subscribeToStream(ctx, stream)
			if err != nil {
				p.logger.Warnf("Failed to init subscription to stream %s (attempt %d/%d): %v", stream, attempt+1, totalAttempts, err)
			}
			return err
		},
		totalAttempts,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to initialize subscription to stream %s: %w", stream, err)
	}
	return cons, nil
}

func (p *pubsuber) subscribeToStream(ctx context.Context, stream string) (jetstream.Consumer, error) {
	s, err := p.createOrUpdateStream(ctx, stream)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream %s on subscribeToStream, err %w", stream, err)
	}
	cfg := config.Get[*config_models.Pubsub]()
	cons, err := s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		// Durable consumers are not deleted automatically when they have no more messages.
		Durable:           cfg.ServiceName,
		MaxDeliver:        cfg.SubscribeRetries + 1,
		AckWait:           time.Duration(cfg.WaitForAckSec) * time.Second,
		InactiveThreshold: time.Duration(cfg.ConsumerTimeoutSec) * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer err %w", err)
	}

	return cons, nil
}

func (p *pubsuber) getNatsConnectOptions(user, password string) []nats.Option {
	cfg := config.Get[*config_models.Pubsub]()
	maxReconnectDuration := 24 * time.Hour
	reconnectInterval := 5 * time.Second

	return []nats.Option{
		nats.Timeout(time.Duration(cfg.ConnectTimeoutSeconds) * time.Second),
		nats.MaxReconnects(int(maxReconnectDuration / reconnectInterval)),
		nats.ReconnectWait(reconnectInterval),
		nats.ReconnectBufSize(-1),
		p.getNatsAuthOptions(user, password),
	}
}

func (p *pubsuber) getNatsAuthOptions(user, password string) nats.Option {
	if user != "" && password != "" {
		return nats.UserInfo(user, password)
	}
	if password != "" {
		return nats.Token(password)
	}
	if user != "" {
		p.logger.Warnf("nats user is set but password is not set")
	}
	return nil
}

func (p *pubsuber) createStreamWithCache(ctx context.Context, stream string) (jetstream.Stream, error) {
	if s := p.getCachedStream(stream); s != nil {
		return s, nil
	}
	jetStream, err := p.createOrUpdateStream(ctx, stream)
	if err != nil {
		return nil, err
	}
	p.cacheStream(stream, jetStream)
	return jetStream, nil
}

func (p *pubsuber) createOrUpdateStream(ctx context.Context, stream string) (jetstream.Stream, error) {
	cfg := config.Get[*config_models.Pubsub]()
	return p.jetStream.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     stream,
		Subjects: []string{stream},
		MaxMsgs:  cfg.MaxEntries,
		MaxAge:   time.Duration(cfg.MaxMessageAgeMinutes) * time.Minute,
		Replicas: cfg.Replicas,
	})
}

func (p *pubsuber) getCachedStream(stream string) jetstream.Stream {
	if s, ok := p.streams.Load(stream); ok {
		return s.(jetstream.Stream)
	}
	return nil
}

func (p *pubsuber) cacheStream(stream string, jetStream jetstream.Stream) {
	p.streams.Store(stream, jetStream)
}

func (p *pubsuber) startConsuming(cons jetstream.Consumer, stream string, callback CallbackFunc) error {
	cc, err := cons.Consume(func(msg jetstream.Msg) {
		p.consume(msg, callback)
	}, jetstream.ConsumeErrHandler(func(consumeCtx jetstream.ConsumeContext, err error) {
		p.logger.Errorf("error while consuming message: %w", err)
	}))
	p.consumers.Store(stream, cc)
	return err
}

func (p *pubsuber) consume(msg jetstream.Msg, callback CallbackFunc) {
	defer func() {
		if r := recover(); r != nil {
			p.logger.Errorf("panic while processing message: %v", r)
			p.handleFailedMessage(msg, fmt.Sprintf("panic while processing message: %v", r))
		}
	}()
	if err := callback(msg.Data()); err != nil {
		p.logger.Errorf("error while processing message: %w", err)
		p.handleFailedMessage(msg, err.Error())
		return
	}

	p.acknowledge(msg)
}

func (p *pubsuber) handleFailedMessage(msg jetstream.Msg, errMsg string) {
	cfg := config.Get[*config_models.Pubsub]()

	if cfg.RetryOnReject && p.retrySendMessage(msg) {
		return
	}

	if cfg.EnableDLQ {
		p.sendDlqMessage(msg, errMsg)
	}

	p.reject(msg, errMsg)
}

func (p *pubsuber) retrySendMessage(msg jetstream.Msg) bool {
	cfg := config.Get[*config_models.Pubsub]()

	metadata, err := msg.Metadata()
	if err != nil {
		p.logger.Errorf("failed to get message metadata: %v", err)
		return false
	}

	if metadata.NumDelivered < uint64(cfg.SubscribeRetries+1) { //nolint:gosec // G115: value is bounded/non-negative
		err := msg.Nak()
		if err != nil {
			p.logger.Errorf("failed to nak message: %v", err)
		}
		return true
	}

	return false
}

func (p *pubsuber) sendDlqMessage(msg jetstream.Msg, errMsg string) {
	dlqMsg := pevent.NewDlq(errMsg, msg.Subject(), msg.Data())
	dlqPayload, err := json.Marshal(dlqMsg)
	if err != nil {
		p.logger.Errorf("failed to marshal dlq message: %w", err)
		return
	}

	if err := p.Publish(pevent.GetDlqTopicName(), dlqPayload); err != nil {
		p.logger.Errorf("failed to publish message to dlq: %w", err)
	}
}

func (p *pubsuber) acknowledge(msg jetstream.Msg) {
	err := msg.Ack()
	if err != nil {
		p.logger.Errorf("error while ack message: %w", err)
	}
}

func (p *pubsuber) reject(msg jetstream.Msg, reason string) {
	err := msg.TermWithReason(reason)
	if err != nil {
		p.logger.Errorf("error while terminating message: %w", err)
	}
}

func tryAppendToChan(errChan chan error, err error) {
	select {
	case errChan <- err:
	default:
	}
}
