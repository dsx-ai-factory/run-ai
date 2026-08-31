/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	MockConfig()
	Init()
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Config", func() {
	Describe("GetConfig", func() {
		Context("with defaults", func() {
			It("should create valid default configuration", func() {
				// Validate ServerConfig defaults
				serverConf := runai_config.Get[*ServerConfig]()
				Expect(serverConf.ServicePort).To(Equal(8080))
				Expect(serverConf.RetryMaxIntervalSec).To(Equal(10))
				Expect(serverConf.RetryMaxElapsedSec).To(Equal(60))
				Expect(serverConf.DaysLimitForFirstSync).To(Equal(7))
				Expect(serverConf.CleanupKeepAssetsDays).To(Equal(30))
				Expect(serverConf.CleanupRefreshRateHours).To(Equal(24))

				// Validate ProjectsCache defaults
				projectsCache := runai_config.Get[*ProjectsCache]()
				Expect(projectsCache.ExpirationSec).To(Equal(600))
				Expect(projectsCache.CleanupSec).To(Equal(1200))
				Expect(projectsCache.MaxRecordSizeKB).To(Equal(300))

				// Validate K8sObjectsTrackerConfig defaults
				trackerConfig := runai_config.Get[*K8sObjectsTrackerConfig]()
				Expect(trackerConfig.LogLevel).To(Equal("info"))
				Expect(trackerConfig.KeyDelimiter).To(Equal("_"))
				Expect(trackerConfig.K8sClientConfigQPS).To(Equal(600))
				Expect(trackerConfig.K8sClientConfigBurst).To(Equal(3000))
			})
		})
	})
})
