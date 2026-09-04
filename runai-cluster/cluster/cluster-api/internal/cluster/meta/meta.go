/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package meta

import (
	rcpconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

const (
	version = "1.0"
)

type Features struct {
	MessageEnabled        bool `mapstructure:"MESSAGE_ENABLED" config:"message_enabled" json:"message_enabled"`
	TerminalResizeEnabled bool `mapstructure:"TERMINAL_RESIZE_ENABLED" json:"message_resize_enabled"`
}

func (c *Features) SetDefaults() {
	c.MessageEnabled = true
	c.TerminalResizeEnabled = true
}

func (c *Features) IsMessageEnabled() bool {
	return c.MessageEnabled
}

func (c *Features) IsDynamicTerminalResizeEnabled() bool {
	if c.MessageEnabled && c.TerminalResizeEnabled {
		return true
	}

	return false
}

func GetFeatures() *Features {
	return rcpconfig.Get[*Features]()
}

type Metadata struct {
	Version  string    `json:"version"`
	Features *Features `json:"features"`
}

func GetVersion() string {
	return version
}

func Get() *Metadata {
	return &Metadata{
		Version:  GetVersion(),
		Features: GetFeatures(),
	}
}
