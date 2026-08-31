/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"os"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-common-packages/config"
)

var _ = ginkgo.Describe("Config", func() {
	ginkgo.BeforeEach(func() {
		os.Unsetenv("SERVICE_PORT")
	})

	ginkgo.Describe("GetConfig", func() {
		ginkgo.Context("with default values", func() {
			ginkgo.It("should return default server and services configuration", func() {
				config.FlushAll()
				Init()

				serverConfig := config.Get[*Server]()
				servicesConfig := config.Get[*Services]()

				gomega.Expect(serverConfig.ServicePort).To(gomega.Equal(8080))
				gomega.Expect(servicesConfig.AuthorizationUrl).To(gomega.Equal("http://runai-backend-authorization:8080"))
			})
		})

		ginkgo.Context("with specific values", func() {
			ginkgo.It("should return configuration with custom service port", func() {
				// Set environment variables BEFORE Init()
				gomega.Expect(os.Setenv("SERVICE_PORT", "9001")).To(gomega.Succeed())

				config.FlushAll()
				Init()

				serverConfig := config.Get[*Server]()
				servicesConfig := config.Get[*Services]()

				gomega.Expect(serverConfig.ServicePort).To(gomega.Equal(9001))
				gomega.Expect(servicesConfig.AuthorizationUrl).To(gomega.Equal("http://runai-backend-authorization:8080"))
			})
		})
	})

	ginkgo.Describe("ServerConfig", func() {
		ginkgo.Context("with defaults", func() {
			ginkgo.It("should have correct default values", func() {
				config.FlushAll()
				Init()

				serverConfig := config.Get[*Server]()

				gomega.Expect(serverConfig.ServicePort).To(gomega.Equal(8080))
			})
		})
	})

	ginkgo.Describe("ServicesConfig", func() {
		ginkgo.Context("with defaults", func() {
			ginkgo.It("should have correct default authorization URL", func() {
				config.FlushAll()
				Init()

				servicesConfig := config.Get[*Services]()

				gomega.Expect(servicesConfig.AuthorizationUrl).To(gomega.Equal("http://runai-backend-authorization:8080"))
			})
		})
	})

})
