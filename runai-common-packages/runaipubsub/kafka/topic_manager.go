package kafka

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/run-ai/runai/runai-common-packages/config"
	config_models "github.com/run-ai/runai/runai-common-packages/runaipubsub/config"
	"github.com/run-ai/runai/runai-common-packages/runaipubsub/logger"
	"github.com/segmentio/kafka-go"
)

type topicManager struct {
	topics     sync.Map
	logger     logger.Loggerer
	connection *kafka.Conn
}

func newTopicManager(url string, dialer *kafka.Dialer, logger logger.Loggerer) (*topicManager, error) {
	conn, err := dialer.Dial("tcp", strings.Split(url, ",")[0])
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Kafka broker: %w", err)
	}
	controller, err := conn.Controller()
	if err != nil {
		return nil, fmt.Errorf("failed to get Kafka controller: %w", err)
	}
	controllerCon, err := dialer.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Kafka controller: %w", err)
	}

	return &topicManager{topics: sync.Map{}, logger: logger, connection: controllerCon}, nil
}

func (tc *topicManager) CreateIfNotExist(topic string) error {
	var err error
	once, _ := tc.topics.LoadOrStore(topic, &sync.Once{})
	once.(*sync.Once).Do(func() {
		if err = tc.createTopic(topic); err != nil {
			err = fmt.Errorf("failed to create topic %s: %w", topic, err)
		}
	})

	return err
}

func (tc *topicManager) createTopic(topic string) error {
	cfg := config.Get[*config_models.Pubsub]()
	retention := cfg.MaxMessageAgeMinutes * time.Minute.Milliseconds()
	return tc.connection.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1, // Can be increased if needed. Once increased, it cannot be decreased.
		ReplicationFactor: cfg.Replicas,
		ConfigEntries: []kafka.ConfigEntry{{
			ConfigName: "retention.ms", ConfigValue: strconv.Itoa(int(retention)),
		}, {
			ConfigName: "cleanup.policy", ConfigValue: "delete",
		}, {
			ConfigName: "min.insync.replicas", ConfigValue: strconv.Itoa(2),
		}},
	})
}
