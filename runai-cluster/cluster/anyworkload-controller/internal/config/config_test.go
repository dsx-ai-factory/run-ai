/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	rcpconfig "github.com/run-ai/runai/runai-common-packages/config"
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
