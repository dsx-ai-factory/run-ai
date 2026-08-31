/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package connection

import (
	"net"
)

type WebSocket interface {
	net.Conn
	WriteText(p []byte) error
	SetMessageHandler(messageHandler MessageHandler)
}
