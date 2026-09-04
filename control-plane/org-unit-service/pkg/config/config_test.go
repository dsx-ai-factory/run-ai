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
	})

	Describe("Init", func() {
		Context("General config", func() {
			It("should return default values when no config is provided", func() {
				Expect(Init()).To(BeNil())

				generalConf := config.Get[*General]()

				Expect(generalConf.LogLevel).To(Equal("info"))
			})

			It("should return custom values from the provided config", func() {
				os.Setenv("LOG_LEVEL", "debug")

				Expect(Init()).To(BeNil())

				generalConf := config.Get[*General]()

				Expect(generalConf.LogLevel).To(Equal("debug"))
			})
		})

		Context("Server config", func() {
			It("should return default values when no config is provided", func() {
				Expect(Init()).To(BeNil())

				serverConf := config.Get[*Server]()

				Expect(serverConf.ServicePort).To(Equal(8080))
				Expect(serverConf.TempBackendSettingsUrl).To(Equal("http://runai-backend-temp-backend-settings:8080"))
				Expect(serverConf.TenantsManagerUrl).To(Equal("http://runai-backend-tenants-manager:8080"))
				Expect(serverConf.ClustersServiceUrl).To(Equal("http://runai-backend-cluster-service:8080"))
				Expect(serverConf.WorkloadsServiceUrl).To(Equal("http://runai-backend-workloads:8080"))
				Expect(serverConf.DatavolumesServiceUrl).To(Equal("http://runai-backend-datavolumes:8080"))
			})

			It("should return custom values from the provided config", func() {
				os.Setenv("SERVICE_PORT", "1234")
				os.Setenv("TEMP_BACKEND_SETTINGS_URL", "temp-backend-settings-url")
				os.Setenv("TENANTS_MANAGER_URL", "tenants-manager-url")
				os.Setenv("CLUSTER_SERVICE_URL", "clusters-service-url")
				os.Setenv("WORKLOADS_SERVICE_URL", "workloads-service-url")
				os.Setenv("DATAVOLUMES_SERVICE_URL", "datavolumes-service-url")

				Expect(Init()).To(BeNil())

				serverConf := config.Get[*Server]()

				Expect(serverConf.ServicePort).To(Equal(1234))
				Expect(serverConf.TempBackendSettingsUrl).To(Equal("temp-backend-settings-url"))
				Expect(serverConf.TenantsManagerUrl).To(Equal("tenants-manager-url"))
				Expect(serverConf.ClustersServiceUrl).To(Equal("clusters-service-url"))
				Expect(serverConf.WorkloadsServiceUrl).To(Equal("workloads-service-url"))
				Expect(serverConf.DatavolumesServiceUrl).To(Equal("datavolumes-service-url"))
			})
		})

		Context("Reports config", func() {
			It("should return default values when no config is provided", func() {
				Expect(Init()).To(BeNil())

				reportsConf := config.Get[*Reports]()

				Expect(reportsConf.BucketName).To(Equal(""))
			})

			It("should return custom values from the provided config", func() {
				os.Setenv("S3_BUCKET", "bucket-name")

				Expect(Init()).To(BeNil())

				reportsConf := config.Get[*Reports]()

				Expect(reportsConf.BucketName).To(Equal("bucket-name"))
			})
		})

		Context("OvertimeMetrics config", func() {
			It("should return default values when no config is provided", func() {
				Expect(Init()).To(BeNil())

				overtimeMetricsConf := config.Get[*OvertimeMetrics]()

				Expect(overtimeMetricsConf.Query24hRangeIntervalSeconds).To(Equal(3600))
				Expect(overtimeMetricsConf.Query7dRangeIntervalSeconds).To(Equal(21600))
				Expect(overtimeMetricsConf.Query30dRangeIntervalSeconds).To(Equal(86400))
				Expect(overtimeMetricsConf.Query24hRangeInitialSleepSeconds).To(Equal(10))
				Expect(overtimeMetricsConf.Query7dRangeInitialSleepSeconds).To(Equal(21))
				Expect(overtimeMetricsConf.Query30dRangeInitialSleepSeconds).To(Equal(33))
				Expect(overtimeMetricsConf.RateLimiterLimit).To(Equal(1))
				Expect(overtimeMetricsConf.RateLimiterBurst).To(Equal(10))
			})

			It("should return custom values from the provided config", func() {
				os.Setenv("QUERY_24H_RANGE_INTERVAL_SECONDS", "1")
				os.Setenv("QUERY_7D_RANGE_INTERVAL_SECONDS", "2")
				os.Setenv("QUERY_30D_RANGE_INTERVAL_SECONDS", "3")
				os.Setenv("QUERY_24H_RANGE_INITIAL_SLEEP_SECONDS", "4")
				os.Setenv("QUERY_7D_RANGE_INITIAL_SLEEP_SECONDS", "5")
				os.Setenv("QUERY_30D_RANGE_INITIAL_SLEEP_SECONDS", "6")
				os.Setenv("RATE_LIMITER_LIMIT", "7")
				os.Setenv("RATE_LIMITER_BURST", "8")

				Expect(Init()).To(BeNil())

				overtimeMetricsConf := config.Get[*OvertimeMetrics]()

				Expect(overtimeMetricsConf.Query24hRangeIntervalSeconds).To(Equal(1))
				Expect(overtimeMetricsConf.Query7dRangeIntervalSeconds).To(Equal(2))
				Expect(overtimeMetricsConf.Query30dRangeIntervalSeconds).To(Equal(3))
				Expect(overtimeMetricsConf.Query24hRangeInitialSleepSeconds).To(Equal(4))
				Expect(overtimeMetricsConf.Query7dRangeInitialSleepSeconds).To(Equal(5))
				Expect(overtimeMetricsConf.Query30dRangeInitialSleepSeconds).To(Equal(6))
				Expect(overtimeMetricsConf.RateLimiterLimit).To(Equal(7))
				Expect(overtimeMetricsConf.RateLimiterBurst).To(Equal(8))
			})
		})

		Context("SyncJob config", func() {
			It("should return default values when no config is provided", func() {
				Expect(Init()).To(BeNil())

				syncJobConf := config.Get[*SyncJob]()

				Expect(syncJobConf.SyncJobIntervalHours).To(Equal(24))
				Expect(syncJobConf.SyncJobTenantIntervalSeconds).To(Equal(5))
			})

			It("should return custom values from the provided config", func() {
				os.Setenv("SYNC_JOB_INTERVAL_HOURS", "1")
				os.Setenv("SYNC_JOB_INTERVAL_TENANT_SECONDS", "2")

				Expect(Init()).To(BeNil())

				syncJobConf := config.Get[*SyncJob]()

				Expect(syncJobConf.SyncJobIntervalHours).To(Equal(1))
				Expect(syncJobConf.SyncJobTenantIntervalSeconds).To(Equal(2))
			})
		})
	})

})
