/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package connection

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	"nhooyr.io/websocket"
)

// BaseMessageHandler provides common websocket message processing functionality
type BaseMessageHandler struct{}

// ProcessCommon handles the common message processing logic that both keepalive
// and terminal handlers share
func (b *BaseMessageHandler) ProcessCommon(
	msgType websocket.MessageType, msgRead []byte, msgBuffer []byte,
) (*Message, int, error) {
	if msgType == websocket.MessageBinary {
		// if the message is binary we just copy what we can.
		return nil, copy(msgBuffer, msgRead), nil
	}

	if msgType != websocket.MessageText {
		return nil, 0, fmt.Errorf("unrecognized message type: %d", msgType)
	}

	statusMsg := Message{}

	err := json.Unmarshal(msgRead, &statusMsg)
	if err != nil {
		log.Warn().Msgf("failed to unmarshall message, message: %s", msgRead)
		return nil, 0, err
	}

	log.Debug().Msgf("received message, type: %s, message: %s", statusMsg.Type, statusMsg.Message)

	return &statusMsg, len(msgRead), nil
}
