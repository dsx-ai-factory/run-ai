package kafka

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/run-ai/runai/runai-common-packages/config"
	config_models "github.com/run-ai/runai/runai-common-packages/runaipubsub/config"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/logger"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/pevent"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/retry"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/plain"
	"github.com/segmentio/kafka-go/sasl/scram"
)

type publisherer interface {
	Publish(topic string, payload []byte) error
	Close() error
}

type subscriberer interface {
	Subscribe(ctx context.Context, callback func(event []byte) error)
	Close() error
}

type topicManagerer interface {
	CreateIfNotExist(topic string) error
}

type dlqHandlerer interface {
	sendToDLQ(msg *pevent.Event[pevent.DlqMessage]) error
}

type pubSuber struct {
	urls         []string
	topicManager topicManagerer
	publisher    publisherer
	subscribers  sync.Map
	dlqHandler   dlqHandlerer
	dialer       *kafka.Dialer
	logger       logger.Loggerer
}

func NewPubSuber(logger logger.Loggerer) *pubSuber {
	return &pubSuber{
		logger:      logger,
		subscribers: sync.Map{},
	}
}

func (kps *pubSuber) Connect(url, user, password string) error {
	err := config.Register[*Config]()
	if err != nil {
		return fmt.Errorf("failed to register kafka config: %w", err)
	}
	kps.setDialer(user, password)
	kps.urls = strings.Split(url, ",")
	tc, err := newTopicManager(url, kps.dialer, kps.logger)
	if err != nil {
		return err
	}

	kps.topicManager = tc
	kps.publisher = newPublisher(kps.urls, kps.dialer, kps.logger)
	kps.dlqHandler = newDLQHandler(kps.publisher)
	return nil
}

func (kps *pubSuber) Publish(topic string, payload []byte) error {
	if kps.publisher == nil {
		return fmt.Errorf("kafka publisher is not initialized")
	}
	if err := kps.topicManager.CreateIfNotExist(topic); err != nil {
		return err
	}
	if err := kps.publisher.Publish(topic, payload); err != nil {
		return fmt.Errorf("failed to Publish message: %w", err)
	}

	return nil
}

func (kps *pubSuber) Subscribe(ctx context.Context, topic string, callback func(event []byte) error) error {
	cfg := config.Get[*config_models.Pubsub]()
	n := cfg.SubscriptionInitRetries + 1
	return kps.subscribe(ctx, topic, callback, make(chan error, 1), n)
}

func (kps *pubSuber) SubscribeAsync(ctx context.Context, topic string, callback func(event []byte) error, errChan chan error) {
	go kps.subscribe(ctx, topic, callback, errChan, math.MaxInt16)
}

func (kps *pubSuber) subscribe(ctx context.Context, topic string, callback func(event []byte) error, errChan chan error, totalAttempts int) error {
	err := kps.initSubscriber(ctx, topic, totalAttempts, errChan)
	if err != nil {
		return err
	}

	if ctx.Err() != nil {
		kps.logger.Infof("context canceled, not subscribed to topic: %s", topic)
		return nil
	}

	s, loaded := kps.subscribers.LoadOrStore(topic, newSubscriber(kps.urls, kps.dialer, topic, kps.logger, kps.dlqHandler))
	if loaded {
		kps.logger.Warnf("already subscribed to topic: %s", topic)
		return nil
	}
	s.(subscriberer).Subscribe(ctx, callback)
	return nil
}

func (kps *pubSuber) Close() error {
	var err error
	kps.subscribers.Range(func(key, value any) bool {
		reader := value.(subscriberer)
		if e := reader.Close(); e != nil {
			kps.logger.Errorf("failed to Close subscriber to topic: %s: %+v", key, e)
			err = errors.Join(err, e)
		}
		return true
	})
	if kps.publisher == nil {
		return err
	}

	if e := kps.publisher.Close(); e != nil {
		kps.logger.Errorf("failed to Close publisher %w", e)
		err = errors.Join(err, e)
	}

	return err
}

func (kps *pubSuber) initSubscriber(ctx context.Context, topic string, totalAttempts int, errChan chan error) error {
	return retry.RetryStreamErrors(ctx, errChan,
		func(attempt uint) error {
			err := kps.topicManager.CreateIfNotExist(topic)
			if err != nil {
				kps.logger.Warnf("Failed to create topic %s (attempt %d/%d): %v", topic, attempt+1, totalAttempts, err)
			}
			return err
		},
		totalAttempts,
	)
}

func (kps *pubSuber) setDialer(user, password string) {
	cfg := config.Get[*config_models.Pubsub]()
	kps.dialer = &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}
	if cfg.EnableTLS {
		kps.dialer.TLS = &tls.Config{MinVersion: tls.VersionTLS12}
	}
	mechanism, err := kps.getAuthMechanism(user, password)
	if err != nil {
		kps.logger.Errorf("failed to get auth mechanism: %w", err)
		return
	}
	kps.dialer.SASLMechanism = mechanism
}

func (kps *pubSuber) getAuthMechanism(user, password string) (sasl.Mechanism, error) {
	if user == "" && password == "" {
		return nil, nil
	}
	if user == "" || password == "" {
		kps.logger.Warnf("Must set both user and password for SASL authentication")
		return nil, nil
	}
	cfg := config.Get[*Config]()
	switch cfg.AuthType {
	case KafkaAuthTypePlain:
		return &plain.Mechanism{
			Username: user,
			Password: password,
		}, nil
	case KafkaAuthTypeScram:
		return scram.Mechanism(scram.SHA512, user, password)
	}
	return nil, fmt.Errorf("invalid SASL mechanism: %s", cfg.AuthType)
}
