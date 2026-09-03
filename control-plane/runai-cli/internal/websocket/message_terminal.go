/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package websocket

import (
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"go.uber.org/zap"
	"nhooyr.io/websocket"
)

type TerminalHandler struct {
	base                BaseMessageHandler
	terminalSizeUpdater *TerminalSizeUpdater
}

func NewTerminalHandler(enc *ExtendedNetConnection, firstHeight int, firstWidth int) *TerminalHandler {
	tsu := NewTerminalSizeUpdater(enc, firstHeight, firstWidth)
	tsu.Start(enc.ctx)
	return &TerminalHandler{terminalSizeUpdater: tsu}
}

func (th *TerminalHandler) Process(msgType websocket.MessageType, msgRead []byte, msgBuffer []byte) (int, error) {
	statusMsg, bytesRead, err := th.base.ProcessCommon(msgType, msgRead, msgBuffer)
	if err != nil {
		return 0, err
	}

	// If it's a binary message, we're done
	if statusMsg == nil {
		return bytesRead, nil
	}

	logger := config.GetInstance().GetLogger()

	// Handle terminal size updates
	if statusMsg.Type == TypeTerminalSize {
		logger.Debug("msg response terminal size received", zap.Int("status", statusMsg.StatusCode), zap.String("message", statusMsg.Message))
		return bytesRead, nil
	}

	if statusMsg.StatusCode != http.StatusOK {
		logger.Debug("msg status failed", zap.Int("status", statusMsg.StatusCode), zap.String("message", statusMsg.Message))
		return 0, fmt.Errorf("%s", statusMsg.Message) // error extracted from response
	}
	logger.Debug("msg status success", zap.String("message", statusMsg.Message))
	return bytesRead, nil
}

func (th *TerminalHandler) Close() error {
	th.terminalSizeUpdater.Stop()
	return nil
}
