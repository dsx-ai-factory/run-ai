/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	rcpconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("RestartEnabled", func() {
	It("defaults to enabled when the env is unset", func() {
		Expect(rcpconfig.Register[*Config]()).To(Succeed())
		Expect(rcpconfig.Get[*Config]().RestartEnabled).To(BeTrue(), "restart reconcilers must default to enabled")
	})

	It("is disabled by RESTART_ENABLED=false", func() {
		GinkgoT().Setenv("RESTART_ENABLED", "false")

		Expect(rcpconfig.Register[*Config]()).To(Succeed())
		Expect(rcpconfig.Get[*Config]().RestartEnabled).To(BeFalse(), "RESTART_ENABLED=false must disable the restart reconcilers")
	})
})

var _ = Describe("SnapshotEnabled", func() {
	// The operator has already ANDed cluster availability with its development
	// gate, so the controller has one knob and no decision of its own to make.
	It("defaults to disabled when the env is unset", func() {
		Expect(rcpconfig.Register[*Config]()).To(Succeed())
		Expect(rcpconfig.Get[*Config]().SnapshotEnabled).To(BeFalse(),
			"a cluster without Snapshot must behave exactly as it does today")
	})

	It("is enabled by SNAPSHOT_ENABLED=true", func() {
		GinkgoT().Setenv("SNAPSHOT_ENABLED", "true")

		Expect(rcpconfig.Register[*Config]()).To(Succeed())
		Expect(rcpconfig.Get[*Config]().SnapshotEnabled).To(BeTrue())
	})
})

var _ = Describe("AgentNodeSelector", func() {
	DescribeTable("parses the operator's comma-separated list",
		func(raw string, expected map[string]string) {
			cfg := &Config{SnapshotAgentNodeSelector: raw}
			Expect(cfg.AgentNodeSelector()).To(Equal(expected))
		},
		Entry("empty means no constraint", "", map[string]string(nil)),
		Entry("a single pair", "nvidia.com/gpu.present=true",
			map[string]string{"nvidia.com/gpu.present": "true"}),
		Entry("several pairs", "a=1,b=2", map[string]string{"a": "1", "b": "2"}),
		Entry("surrounding space is ignored", " a=1 , b=2 ",
			map[string]string{"a": "1", "b": "2"}),
		Entry("an empty value is still a constraint", "a=", map[string]string{"a": ""}),
		// A malformed entry drops rather than failing: a cosmetic mistake in a
		// node selector should not take the controller down.
		Entry("entries without a separator are dropped", "a=1,garbage",
			map[string]string{"a": "1"}),
		Entry("nothing usable means no constraint", "garbage", map[string]string(nil)),
	)
})
