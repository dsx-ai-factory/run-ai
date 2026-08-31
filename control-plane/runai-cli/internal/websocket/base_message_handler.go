/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package websocket

import (
	"encoding/json"
	"fmt"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"go.uber.org/zap"
	"nhooyr.io/websocket"
)

// BaseMessageHandler provides common websocket message processing functionality
type BaseMessageHandler struct{}

// ProcessCommon handles the common message processing logic that both keepalive and terminal handlers share
func (b *BaseMessageHandler) ProcessCommon(msgType websocket.MessageType, msgRead []byte, msgBuffer []byte) (*Message, int, error) {
	if msgType == websocket.MessageBinary {
		// if the message is binary we just copy what we can.
		return nil, copy(msgBuffer, msgRead), nil
	}

	if msgType != websocket.MessageText {
		return nil, 0, fmt.Errorf("unrecognized message type: %d", msgType)
	}

	logger := config.GetInstance().GetLogger()

	statusMsg := Message{}
	err := json.Unmarshal(msgRead, &statusMsg)
	if err != nil {
		logger.Warn("failed to unmarshall message", zap.ByteString("msg", msgRead))
		return nil, 0, err
	}

	logger.Debug("received message", zap.String("type", string(statusMsg.Type)), zap.String("message", statusMsg.Message))

	return &statusMsg, len(msgRead), nil
}
