/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pevent

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	config_models "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/config"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	namePattern = "%sDlq"
)

var (
	topicName     string
	topicNameOnce sync.Once
)

type DlqMessage struct {
	Error       string    `json:"error"`
	OriginTopic string    `json:"origin_topic"`
	Consumer    string    `json:"consumer"`
	Body        []byte    `json:"body"`
	Timestamp   time.Time `json:"timestamp"`
}

func NewDlq(errMsg string, topic string, body []byte) *Event[DlqMessage] {
	msg := &DlqMessage{
		Error:       errMsg,
		OriginTopic: topic,
		Consumer:    config.Get[*config_models.Pubsub]().ServiceName,
		Body:        body,
		Timestamp:   time.Now(),
	}
	return New[DlqMessage]().WithPayload(msg)
}

func (m *DlqMessage) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func GetDlqTopicName() string {
	topicNameOnce.Do(func() {
		topicName = fmt.Sprintf(namePattern, toPascalCase(config.Get[*config_models.Pubsub]().ServiceName))
	})
	return topicName
}

func toPascalCase(input string) string {
	if input == "" {
		return ""
	}

	replacer := strings.NewReplacer("-", " ", "_", " ")
	normalized := replacer.Replace(input)

	words := strings.Fields(normalized)

	caser := cases.Title(language.English)
	for i, word := range words {
		words[i] = caser.String(word[:1]) + word[1:]
	}

	return strings.Join(words, "")
}
