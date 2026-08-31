/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package extensions

import (
	"context"
	"maps"
)

type CapabilityType string

type EventType string

type Type string

func (t Type) String() string {
	return string(t)
}

type Extension struct {
	Type   Type   `json:"type"`
	Status Status `json:"status"`
}

func NewExtension(t Type, enabled bool, capabilities map[CapabilityType]bool) Extension {
	return Extension{
		Type:   t,
		Status: Status{Enabled: enabled, Capabilities: capabilities},
	}
}

func (e Extension) GetType() Type {
	return e.Type
}

func (e Extension) Equals(other Extension) bool {
	if e.Type != other.Type {
		return false
	}
	if e.Status.Enabled != other.Status.Enabled {
		return false
	}
	return maps.Equal(e.Status.Capabilities, other.Status.Capabilities)
}

func (e Extension) IsEnabled() bool {
	return e.Status.Enabled
}

func (e Extension) IsCapable(capability CapabilityType) bool {
	return e.Status.Capabilities[capability]
}

type Status struct {
	Enabled      bool                    `json:"enabled"`
	Capabilities map[CapabilityType]bool `json:"capabilities"`
}

type Group map[Type]Extension

func (g Group) Equals(other Group) bool {
	if len(g) != len(other) {
		return false
	}
	for extType, ext := range g {
		otherExt, exists := other[extType]
		if !exists {
			return false
		}
		if !ext.Equals(otherExt) {
			return false
		}
	}
	return true
}

type Client interface {
	Get(ctx context.Context, extension Type) (*Extension, error)
	List(ctx context.Context, extensions []Type) (Group, error)
}
