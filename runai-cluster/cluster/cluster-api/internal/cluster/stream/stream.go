/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package stream

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/cluster/meta"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/cluster/stream/resize"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket/connection"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/websocket/keepalive"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"
)

type Type string

const (
	Interactive Type = "interactive"
	ReadOnly    Type = "readonly"
)

type Handler struct {
	options          Options
	conn             connection.WebSocket
	ht               Type
	pingDurationTime time.Duration
}

func NewHandler(conn connection.WebSocket, options Options) *Handler {
	return &Handler{
		conn:             conn,
		options:          options,
		ht:               setType(options),
		pingDurationTime: GetInstance().GetPingTime(),
	}
}

func (h *Handler) GetStreamOptions() remotecommand.StreamOptions {
	if h.GetType() == ReadOnly {
		return remotecommand.StreamOptions{
			Stdout: h.conn,
			Stderr: h.conn,
		}
	}

	streamOptions := remotecommand.StreamOptions{
		Stdout: h.conn,
		Stderr: h.conn,
		Stdin:  h.conn,
		Tty:    h.options.TTY,
	}

	// Add if stream is not interactive, no need to add interactive message handler
	if !h.options.TTY {
		return streamOptions
	}

	terminalSizeQueue := resize.NewTerminalSizeQueue()
	if h.options.TTYWidth > 0 && h.options.TTYHeight > 0 {
		terminalSizeQueue.Resize(h.options.TTYWidth, h.options.TTYHeight)
	} else {
		log.Warn().Msgf("options TTY size must be positive: %d/%d", h.options.TTYWidth, h.options.TTYHeight)
	}

	if meta.GetFeatures().IsDynamicTerminalResizeEnabled() {
		interactiveHandler := connection.NewInteractiveHandler(h.conn, terminalSizeQueue)
		h.conn.SetMessageHandler(interactiveHandler)
		streamOptions.TerminalSizeQueue = interactiveHandler
	} else {
		if meta.GetFeatures().IsMessageEnabled() {
			websocket.SetClientMessageHandler(h.conn)
		}

		streamOptions.TerminalSizeQueue = terminalSizeQueue
	}

	return streamOptions
}

func (h *Handler) GetKeepAlive() (keepalive.Keepalive, error) {
	switch h.GetType() {
	case Interactive:
		return keepalive.NewPing(h.conn, h.conn, h.pingDurationTime), nil
	case ReadOnly:
		return keepalive.NewIdle(h.conn, h.conn, h.pingDurationTime, h.pingDurationTime)
	}

	return nil, fmt.Errorf("keepalive type was not defined")
}

func (h *Handler) GetType() Type {
	return h.ht
}

func (h *Handler) Close() {
	_ = h.conn.Close()
}

func (h *Handler) WriteStatusMsg(status v1.Status) error {
	p, err := json.Marshal(status)
	if err != nil {
		return err
	}

	return h.conn.WriteText(p)
}

func setType(o Options) Type {
	if o.Stdin || o.TTY {
		return Interactive
	}

	return ReadOnly
}
