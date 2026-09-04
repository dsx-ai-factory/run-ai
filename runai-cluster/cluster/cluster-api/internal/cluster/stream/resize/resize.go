/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resize

import (
	"encoding/json"
	"sync"

	"github.com/rs/zerolog/log"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	resizeChannelSize = 10
)

// TerminalSize represents the width and height of a terminal
type TerminalSize struct {
	Width  uint16 `json:"width"`
	Height uint16 `json:"height"`
}

// TerminalSizeQueue implements remotecommand.TerminalSizeQueue
// It receives terminal resize events and provides them to the remote command executor
type TerminalSizeQueue struct {
	resizeChan chan *remotecommand.TerminalSize
	stopChan   chan struct{}
	mutex      sync.Mutex
	stopped    bool
}

// NewTerminalSizeQueue creates a new terminal size queue
func NewTerminalSizeQueue() *TerminalSizeQueue {
	return &TerminalSizeQueue{
		resizeChan: make(chan *remotecommand.TerminalSize, resizeChannelSize), // Buffer to prevent blocking
		stopChan:   make(chan struct{}),
	}
}

// Next returns the new terminal size after the terminal has been resized.
// It returns nil when monitoring has been stopped.
func (t *TerminalSizeQueue) Next() *remotecommand.TerminalSize {
	select {
	case size := <-t.resizeChan:
		log.Debug().Uint16("width", size.Width).Uint16("height", size.Height).Msg("Terminal resize event received")
		return size
	case <-t.stopChan:
		log.Debug().Msg("Terminal size queue stopped")
		return nil
	}
}

// Resize adds a new terminal size to the queue
func (t *TerminalSizeQueue) Resize(width, height uint16) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if t.stopped {
		log.Debug().Msg("Terminal size queue is stopped, ignoring resize event")
		return
	}

	size := &remotecommand.TerminalSize{
		Width:  width,
		Height: height,
	}

	select {
	case t.resizeChan <- size:
		log.Debug().Uint16("width", width).Uint16("height", height).Msg("Terminal resize event queued")
	default:
		// Channel is full, drop the oldest resize event and add the new one
		t.dropOldestResizeEvent()

		t.resizeChan <- size

		log.Debug().Uint16("width", width).Uint16("height", height).Msg("Terminal resize event queued (dropped oldest)")
	}
}

func (t *TerminalSizeQueue) dropOldestResizeEvent() {
	select {
	case <-t.resizeChan:
	default:
	}
}

// Stop stops the terminal size queue
func (t *TerminalSizeQueue) Stop() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.stopped {
		t.stopped = true
		close(t.stopChan)
		log.Debug().Msg("Terminal size queue stopped")
	}
}

// HandleResizeMessage processes a resize message from the websocket
func (t *TerminalSizeQueue) HandleResizeMessage(data []byte) error {
	var size TerminalSize
	if err := json.Unmarshal(data, &size); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal terminal resize message")
		return err
	}

	log.Debug().Uint16("width", size.Width).Uint16("height", size.Height).Msg("Handling terminal resize message")

	t.Resize(size.Width, size.Height)

	return nil
}
