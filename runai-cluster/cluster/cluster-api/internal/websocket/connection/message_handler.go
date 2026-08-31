/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package connection

import (
	"nhooyr.io/websocket"
)

type MessageHandler interface {
	Process(msgType websocket.MessageType, msgRead []byte, msgBuf []byte) (int, error)
	Close() error
}
