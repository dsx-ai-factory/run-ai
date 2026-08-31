/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
)

var _ = Describe("NetworkBuilder", func() {
	It("should build a Network correctly", func() {
		network := displayer.NewNetworkBuilder().
			Name("network1").
			ToolType("jupyter").
			ConnectionType("external").
			Url("https://example.com").
			Port(8080).
			AuthorizationType("public").
			AuthorizedUsers([]string{"user1", "user2"}).
			AuthorizedGroups([]string{"group1"}).
			Build()

		Expect(network.Name).To(Equal("network1"))
		Expect(network.ToolType).To(Equal("jupyter"))
		Expect(network.ConnectionType).To(Equal("external"))
		Expect(*network.Url).To(Equal("https://example.com"))
		Expect(*network.Port).To(Equal(int32(8080)))
		Expect(*network.AuthorizationType).To(Equal("public"))
		Expect(network.AuthorizedUsers).To(Equal([]string{"user1", "user2"}))
		Expect(network.AuthorizedGroups).To(Equal([]string{"group1"}))
	})

	It("should handle URL and Port", func() {
		network := displayer.NewNetworkBuilder().
			Url("https://example.com").
			Port(8080).
			Build()

		Expect(*network.Url).To(Equal("https://example.com"))
		Expect(*network.Port).To(Equal(int32(8080)))
	})
})

var _ = Describe("NetworksBuilder", func() {
	It("should build a collection of networks", func() {
		network1 := displayer.NewNetworkBuilder().Name("net1").Build()
		network2 := displayer.NewNetworkBuilder().Name("net2").Build()

		networks := displayer.NewNetworksBuilder().
			Add(network1).
			Add(network2).
			Build()

		Expect(networks).To(HaveLen(2))
		Expect(networks[0].Name).To(Equal("net1"))
		Expect(networks[1].Name).To(Equal("net2"))
	})

	It("should set entire collection with Set method", func() {
		network1 := displayer.NewNetworkBuilder().Name("net1").Build()
		network2 := displayer.NewNetworkBuilder().Name("net2").Build()

		networks := displayer.NewNetworksBuilder().
			Set([]displayer.Network{network1}).
			Add(network2).
			Build()

		Expect(networks).To(HaveLen(2))
		Expect(networks[0].Name).To(Equal("net1"))
		Expect(networks[1].Name).To(Equal("net2"))
	})
})
