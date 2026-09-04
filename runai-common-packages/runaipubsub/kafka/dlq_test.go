/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kafka

import (
	"encoding/json"
	"testing"

	"errors"

	kafka_mocks "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/kafka/mocks"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/pevent"
	"github.com/stretchr/testify/require"
)

func TestSendToDLQ_Success(t *testing.T) {
	publisher := kafka_mocks.NewMockpublisherer(t)
	handler := newDLQHandler(publisher)

	ev := pevent.NewDlq("Test error", "TestTopic", []byte("test payload"))
	payload, err := json.Marshal(ev)
	require.NoError(t, err)

	publisher.On("Publish", pevent.GetDlqTopicName(), payload).Return(nil)

	err = handler.sendToDLQ(ev)
	require.NoError(t, err)
	publisher.AssertCalled(t, "Publish", pevent.GetDlqTopicName(), payload)
}

func TestSendToDLQ_FailPublish(t *testing.T) {
	publisher := kafka_mocks.NewMockpublisherer(t)
	handler := newDLQHandler(publisher)

	ev := pevent.NewDlq("Test error", "TestTopic", []byte("test payload"))
	payload, err := json.Marshal(ev)
	require.NoError(t, err)

	publisher.On("Publish", pevent.GetDlqTopicName(), payload).Return(errors.New("publish failed"))

	err = handler.sendToDLQ(ev)
	require.Equal(t, "publish failed", err.Error())

	publisher.AssertCalled(t, "Publish", pevent.GetDlqTopicName(), payload)
}
