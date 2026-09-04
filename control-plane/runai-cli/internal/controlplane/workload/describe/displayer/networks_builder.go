/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

// Builder
type NetworkBuilder struct {
	network Network
}

func NewNetworkBuilder() *NetworkBuilder {
	return &NetworkBuilder{network: Network{}}
}

func (b *NetworkBuilder) Name(name string) *NetworkBuilder {
	b.network.Name = name
	return b
}

func (b *NetworkBuilder) ToolType(toolType string) *NetworkBuilder {
	b.network.ToolType = toolType
	return b
}

func (b *NetworkBuilder) ConnectionType(connectionType string) *NetworkBuilder {
	b.network.ConnectionType = connectionType
	return b
}

func (b *NetworkBuilder) Url(url string) *NetworkBuilder {
	b.network.Url = &url
	return b
}

func (b *NetworkBuilder) AuthorizationType(authType string) *NetworkBuilder {
	b.network.AuthorizationType = &authType
	return b
}

func (b *NetworkBuilder) AuthorizedUsers(users []string) *NetworkBuilder {
	b.network.AuthorizedUsers = users
	return b
}

func (b *NetworkBuilder) AuthorizedGroups(groups []string) *NetworkBuilder {
	b.network.AuthorizedGroups = groups
	return b
}

func (b *NetworkBuilder) Port(port int32) *NetworkBuilder {
	b.network.Port = &port
	return b
}

func (b *NetworkBuilder) Build() Network {
	return b.network
}
