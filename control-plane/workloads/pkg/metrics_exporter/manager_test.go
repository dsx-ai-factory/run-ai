/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics_exporter

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads/pkg/config"
	"github.com/run-ai/runai/control-plane/workloads/pkg/metrics_exporter/mocks"
	"github.com/run-ai/runai/control-plane/workloads/pkg/metrics_exporter/models"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	metrics_mock "github.com/run-ai/runai/runai-common-packages/metrics/runaimetrics/mocks"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Manager", func() {
	var (
		tenantsClienterMock        *mocks.TenantsClienter
		metricsExporterMock        *mocks.MetricsExporterer
		metricsExporterFactoryMock *MockExporterFactoryer
		manager                    *Manager
		metricsScrapeDurationMock  *metrics_mock.MockMetricHistogram
	)

	BeforeEach(func() {
		tenantsClienterMock = mocks.NewTenantsClienter(GinkgoT())
		metricsExporterMock = mocks.NewMetricsExporterer(GinkgoT())
		metricsExporterFactoryMock = NewMockExporterFactoryer(GinkgoT())

		metricsScrapeDurationMock = metrics_mock.NewMockMetricHistogram(GinkgoT())
		manager = &Manager{
			tenantsClienter:        tenantsClienterMock,
			metricsExporterFactory: metricsExporterFactoryMock,
			metricsScrapeDuration:  metricsScrapeDurationMock,
		}
	})

	Describe("exportMetrics", func() {
		var (
			ctx       context.Context
			metrics   []*models.ExportableMetric
			tenantIds []int32
			done      chan bool
		)

		doneFunc := func(args mock.Arguments) {
			done <- true
		}

		BeforeEach(func() {
			ctx = context.Background()
			metrics = []*models.ExportableMetric{
				{
					Name:          "test_metric_1",
					LabelsMapping: map[string]string{},
				},
			}
			tenantIds = []int32{1001, 1002}
			done = make(chan bool, 10)
		})

		It("should get tenant IDs and export metrics successfully", func() {
			runai_config.Set(&config.MetricsExporter{IntervalSeconds: 1})
			ctx, cancel := context.WithCancel(context.Background())
			tenantsClienterMock.On("GetTenantIds", mock.Anything).Return(tenantIds, nil)
			metricsExporterFactoryMock.On("CreateExporter", mock.Anything).Return(metricsExporterMock)
			metricsExporterMock.On("Export", mock.Anything, tenantIds).Run(doneFunc)
			metricsScrapeDurationMock.On("Observe", mock.Anything)
			go manager.start(ctx, metrics)

			<-done
			cancel()
			tenantsClienterMock.AssertExpectations(GinkgoT())
			metricsExporterMock.AssertExpectations(GinkgoT())
		})

		It("should return error when getting tenant IDs fails", func() {
			tenantsClienterMock.On("GetTenantIds", ctx).Return(nil, errors.New("connection refused"))

			err := manager.exportMetrics(ctx, metrics)

			Expect(err.Error()).To(Equal("failed to get tenant IDs: connection refused"))
			tenantsClienterMock.AssertExpectations(GinkgoT())
		})

		It("should export multiple metrics", func() {
			metrics = []*models.ExportableMetric{
				{Name: "test_metric_1", LabelsMapping: map[string]string{}},
				{Name: "test_metric_2", LabelsMapping: map[string]string{}},
				{Name: "test_metric_3", LabelsMapping: map[string]string{}},
			}

			tenantsClienterMock.On("GetTenantIds", ctx).Return(tenantIds, nil)
			metricsExporterFactoryMock.On("CreateExporter", mock.Anything).Return(metricsExporterMock)
			metricsExporterMock.On("Export", ctx, tenantIds).Run(doneFunc)

			err := manager.exportMetrics(ctx, metrics)

			<-done
			<-done
			<-done
			Expect(err).To(BeNil())
			tenantsClienterMock.AssertExpectations(GinkgoT())
			metricsExporterMock.AssertExpectations(GinkgoT())
		})

		It("should respect context cancellation", func() {
			runai_config.Set(&config.MetricsExporter{IntervalSeconds: 1})
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				<-time.After(50 * time.Millisecond)
				cancel()
			}()

			manager.start(ctx, metrics)

			tenantsClienterMock.AssertNotCalled(GinkgoT(), "GetTenantIds")
			metricsExporterMock.AssertNotCalled(GinkgoT(), "Export")
		})
	})
})
