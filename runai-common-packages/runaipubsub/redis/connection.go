package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-redisstream/pkg/redisstream"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/redis/go-redis/v9"
	"github.com/run-ai/runai/runai-common-packages/config"
	config_models "github.com/run-ai/runai/runai-common-packages/runaipubsub/config"
)

type RedisStreamConnection struct {
}

func NewConnection() *RedisStreamConnection {
	return &RedisStreamConnection{}
}

func (rsc *RedisStreamConnection) connect(logger *Logger, url, user, password string) (*redis.Client, message.Publisher, message.Subscriber, error) {
	opts, err := createRedisOptions(logger, url, user, password)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create redis options: %w", err)
	}

	redisClient := redis.NewClient(opts)
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	publisher, err := rsc.createPublisher(redisClient, logger)
	if err != nil {
		return nil, nil, nil, err
	}
	subscriber, err := rsc.createSubscriber(redisClient, logger)
	if err != nil {
		return nil, nil, nil, err
	}
	return redisClient, publisher, subscriber, nil
}

func (rsc *RedisStreamConnection) createPublisher(redisClient *redis.Client, logger *Logger) (message.Publisher, error) {
	publisher, err := redisstream.NewPublisher(
		redisstream.PublisherConfig{
			Client:     redisClient,
			Marshaller: redisstream.DefaultMarshallerUnmarshaller{},
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis publisher: %w", err)
	}
	return publisher, nil
}

func (rsc *RedisStreamConnection) createSubscriber(redisClient *redis.Client, logger *Logger) (message.Subscriber, error) {
	conf := config.Get[*config_models.Pubsub]()
	subscriber, err := redisstream.NewSubscriber(
		redisstream.SubscriberConfig{
			Client:          redisClient,
			Unmarshaller:    redisstream.DefaultMarshallerUnmarshaller{},
			ConsumerGroup:   config.Get[*config_models.Pubsub]().ServiceName,
			ConsumerTimeout: time.Duration(conf.ConsumerTimeoutSec) * time.Second,
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create redis subscriber: %w", err)
	}
	return subscriber, nil
}

func createRedisOptions(logger *Logger, url, user, password string) (*redis.Options, error) {
	conf := config.Get[*config_models.Pubsub]()
	if !strings.HasPrefix(url, "redis://") {
		url = "redis://" + url
	}
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	opt.Username = user
	opt.Password = password
	opt.ConnMaxIdleTime = time.Duration(conf.ConnMaxIdleTimeSec) * time.Second
	opt.ConnMaxLifetime = time.Hour
	opt.OnConnect = func(ctx context.Context, connection *redis.Conn) error {
		err := connection.Ping(ctx).Err()
		if err != nil {
			logger.Error("Error to ping Redis connection", err, watermill.LogFields{})
			return err
		}
		logger.Debug("OnConnect called,succeed", watermill.LogFields{})
		return nil
	}

	return opt, nil
}
