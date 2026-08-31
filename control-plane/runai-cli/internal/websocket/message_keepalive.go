/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package websocket

import (
	"fmt"

	"nhooyr.io/websocket"
)

type Keepalive struct {
	base BaseMessageHandler
}

func (ka *Keepalive) Process(msgType websocket.MessageType, msgRead []byte, msgBuffer []byte) (int, error) {
	statusMsg, bytesRead, err := ka.base.ProcessCommon(msgType, msgRead, msgBuffer)
	if err != nil {
		return 0, err
	}

	// If it's a binary message, we're done
	if statusMsg == nil {
		return bytesRead, nil
	}

	err = ka.handleOldMessages(statusMsg)
	if err != nil {
		return 0, err
	}
	return bytesRead, nil
}

func (ka *Keepalive) handleOldMessages(msg *Message) error {
	// only old messages has status value
	if len(msg.Status) == 0 {
		return nil
	}

	// if failure return the message as error
	if msg.Status == StatusFailure {
		// Go 1.24 workaround - Go doesn't allow creating dynamic error message
		return fmt.Errorf("%s", msg.Message)
	}
	return nil
}

func (ka *Keepalive) Close() error {
	return nil
}
