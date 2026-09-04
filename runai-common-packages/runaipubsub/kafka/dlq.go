/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kafka

import (
	"encoding/json"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaipubsub/pevent"
)

type dlqHandler struct {
	publisher publisherer
}

func newDLQHandler(publisher publisherer) *dlqHandler {
	return &dlqHandler{
		publisher: publisher,
	}
}

func (d *dlqHandler) sendToDLQ(msg *pevent.Event[pevent.DlqMessage]) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return d.publisher.Publish(pevent.GetDlqTopicName(), payload)
}
