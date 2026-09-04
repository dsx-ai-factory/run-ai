/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"os"
	"testing"

	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Init", func() {
	Describe("Get", func() {
		BeforeEach(func() {
			runai_config.FlushAll()
			Expect(os.Unsetenv("SERVICE_PORT")).To(BeNil())
			Expect(os.Setenv("AUTH_SERVICE_URL", "localhost:8080")).To(BeNil())
		})

		Context("Get config", func() {
			It("should return default values", func() {
				Init()
				serverConf := runai_config.Get[*ServerConfig]()
				Expect(serverConf.ServicePort).To(Equal(8080))
				Expect(runai_config.Get[*OrgHierarchyManagerConfig]().BackendUrl).To(Equal("http://runai-backend-temp-backend-settings:8080"))
				Expect(runai_config.Get[*OrgHierarchyManagerConfig]().ClusterServiceUrl).To(Equal("http://runai-backend-cluster-service:8080"))
			})
		})
	})
})
