/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package base

import (
	"io"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/cluster/status"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/websocket"
)

type StreamAttachExecute struct {
	*StreamBase
}

func NewStreamAttachExecute(streamBase *StreamBase) *StreamAttachExecute {
	return &StreamAttachExecute{StreamBase: streamBase}
}

func (e *StreamAttachExecute) SetupMessageHandlerAttachExecute(conn io.ReadWriteCloser, clusterMetadata *status.Metadata, tty bool, height, width int) (io.ReadWriteCloser, error) {
	var handler websocket.MessageHandler

	enc, ok := conn.(*websocket.ExtendedNetConnection)
	if !ok {
		// better practice will be return an error, but that will break our unit tests.
		return conn, nil
	}

	// TTY with resized enabled
	if tty && clusterMetadata.Features.TerminalResizeEnabled {
		handler = websocket.NewTerminalHandler(enc, height, width)
	}
	return e.SetupMessageHandler(enc, clusterMetadata, handler)
}
