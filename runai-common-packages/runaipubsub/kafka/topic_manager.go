/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kafka

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	config_models "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/logger"
	"github.com/segmentio/kafka-go"
)

type topicManager struct {
	topics     sync.Map
	logger     logger.Loggerer
	connection *kafka.Conn

	// createTopic creates a single topic on the broker. It is a field so tests
	// can exercise the creation bookkeeping without a broker; newTopicManager
	// points it at the broker-backed implementation.
	createTopic func(topic string) error
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

	manager := &topicManager{topics: sync.Map{}, logger: logger, connection: controllerCon}
	manager.createTopic = manager.createTopicOnBroker
	return manager, nil
}

// topicCreation guards a single topic's creation and remembers its outcome. The
// error lives next to the Once because a sync.Once cannot report it: only the
// caller that runs the function would see it, and every concurrent caller would
// be told the topic exists.
type topicCreation struct {
	once sync.Once
	err  error
}

func (tc *topicManager) CreateIfNotExist(topic string) error {
	state, _ := tc.topics.LoadOrStore(topic, &topicCreation{})
	creation := state.(*topicCreation)

	creation.once.Do(func() {
		if err := tc.createTopic(topic); err != nil {
			creation.err = fmt.Errorf("failed to create topic %s: %w", topic, err)
			// Do not remember a failed creation: callers retry through
			// retry.RetryStreamErrors and have to attempt the topic again.
			tc.topics.Delete(topic)
		}
	})

	return creation.err
}

func (tc *topicManager) createTopicOnBroker(topic string) error {
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
