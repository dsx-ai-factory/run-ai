/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package websocket

const (
	IntMapFieldHeight = "height"
	IntMapFieldWidth  = "width"
)

type MessageType string
type MessageStatus string
type MessageKind string

const (
	TypeTerminalSize MessageType = "terminal_size"

	KindRequest  MessageKind = "request"
	KindResponse MessageKind = "response"

	StatusSuccess MessageStatus = "Success"
	StatusFailure MessageStatus = "Failure"
)

type Message struct {
	Version    string         `json:"version"`
	Id         int            `json:"id,omitempty"`
	Message    string         `json:"message,omitempty"`
	Type       MessageType    `json:"type,omitempty"`
	Params     map[string]any `json:"params,omitempty"`
	Status     MessageStatus  `json:"status,omitempty"` // Status was used in the older cluster versions.
	StatusCode int            `json:"statusCode,omitempty"`
	Kind       MessageKind    `json:"kind,omitempty"`
}
