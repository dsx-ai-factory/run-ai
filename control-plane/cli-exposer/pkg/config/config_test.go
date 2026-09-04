/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"os"
	"testing"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Config", func() {
	BeforeEach(func() {
		os.Clearenv()
		config.FlushAll()
	})

	Describe("GetConfig", func() {
		Context("with defaults", func() {
			It("should create valid configuration", func() {
				Expect(InitEnvironment()).To(BeNil())

				serverConf := config.Get[*ServerConfig]()

				Expect(serverConf.ServicePort).To(Equal(8080))
				Expect(serverConf.RetryMaxIntervalSec).To(Equal(10))
				Expect(serverConf.RetryMaxElapsedSec).To(Equal(60))
			})
		})

		Context("with specific values", func() {
			It("should create valid configuration", func() {
				Expect(os.Setenv("SERVICE_PORT", "9001")).To(BeNil())
				Expect(os.Setenv("RETRY_MAX_INTERVAL_SEC", "5")).To(BeNil())
				Expect(os.Setenv("RETRY_MAX_ELAPSED_SEC", "10")).To(BeNil())
				Expect(os.Setenv("ROUTER_WRITE_TIME_SEC", "45")).To(BeNil())
				Expect(os.Setenv("ROUTER_READ_TIME_SEC", "46")).To(BeNil())
				Expect(os.Setenv("ROUTER_IDLE_TIME_SEC", "47")).To(BeNil())

				Expect(InitEnvironment()).To(BeNil())

				serverConf := config.Get[*ServerConfig]()

				Expect(serverConf.ServicePort).To(Equal(9001))
				Expect(serverConf.RetryMaxIntervalSec).To(Equal(5))
				Expect(serverConf.RetryMaxElapsedSec).To(Equal(10))
			})
		})

	})

})
