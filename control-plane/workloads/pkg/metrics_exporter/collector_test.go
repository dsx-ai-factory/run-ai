/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics_exporter

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
	openapi "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/workloads/pkg/metrics_exporter/mocks"
	"github.com/run-ai/runai/control-plane/workloads/pkg/metrics_exporter/models"
	"github.com/run-ai/runai/control-plane/workloads/pkg/runaimetrics"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/mock"
)

var (
	mockRunaiExternalPodInfoMetricQueryResponse = openapi.GetMetrics200Response{
		Status: "success",
		Data: []openapi.Result{
			{
				MetricLabels: map[string]string{
					"clusterId": "clusterId_value",
					"pod_uuid":  "podId_value",
				},
				Measurements: []openapi.Measurement{
					{
						Value: "1",
					},
				},
			},
		},
	}
	mockMetricServerResponseMap = map[string]openapi.GetMetrics200Response{
		runaimetrics.RunaiExternalPodInfoMetric{}.GetName(): mockRunaiExternalPodInfoMetricQueryResponse,
	}
)

var _ = Describe("MetricsCollector", Ordered, func() {
	BeforeAll(func() {
		metrics_client.SetMockMetricsServer(mockMetricServerResponseMap)
	})

	Describe("NewMetricsCollector", func() {
		It("should create a new MetricsCollector", func() {
			exportableMetric := &models.ExportableMetric{
				Name:          "test_metric",
				Description:   "test_metric_description",
				LabelsMapping: map[string]string{},
				Metric:        runaimetrics.RunaiExternalPodInfoMetric{},
			}

			collector := NewMetricsCollector(exportableMetric)
			Expect(collector).NotTo(BeNil())
			Expect(collector.exportableMetric).To(Equal(exportableMetric))
		})
	})

	Describe("Collect", func() {
		var (
			collector            *MetricsCollector
			exportableMetric     *models.ExportableMetric
			clustersClienterMock *mocks.ClustersClienter
		)

		BeforeEach(func() {
			exportableMetric = &models.ExportableMetric{
				Name:        "test_metric",
				Description: "test_metric_description",
				LabelsMapping: map[string]string{
					"clusterId": "mapped_clusterId",
					"pod_uuid":  "mapped_podId",
				},
				Metric: runaimetrics.RunaiExternalPodInfoMetric{
					ResponseLabels: runaimetrics.RunaiExternalPodInfoResponseLabels{
						ClusterId: true,
						PodId:     true,
					},
				},
			}
			clustersClienterMock = mocks.NewClustersClienter(GinkgoT())
			collector = &MetricsCollector{
				exportableMetric:  exportableMetric,
				clustersClienter:  clustersClienterMock,
				concurrentLimitCh: make(chan struct{}, 50),
			}
		})

		It("should collect metric records for tenants", func() {
			ctx := context.Background()
			tenantIds := []int32{1001, 1002}
			clustersClienterMock.On("GetClusterIds", context.WithValue(ctx, runaictx.CfTenantId, "1001"), tenantIds[0]).Return([]string{"clusterId_value"}, nil)
			clustersClienterMock.On("GetClusterIds", context.WithValue(ctx, runaictx.CfTenantId, "1002"), tenantIds[1]).Return([]string{"clusterId_value"}, nil)
			result := collector.Collect(ctx, tenantIds)
			Expect(result).To(ConsistOf(
				models.MetricRecord{
					Labels: map[string]string{
						"tenant_id":        "1001",
						"mapped_clusterId": "clusterId_value",
						"mapped_podId":     "podId_value",
					},
					Value: 1,
				},
				models.MetricRecord{
					Labels: map[string]string{
						"tenant_id":        "1002",
						"mapped_clusterId": "clusterId_value",
						"mapped_podId":     "podId_value",
					},
					Value: 1,
				},
			))
		})

		It("should stop collection and return early when context is done", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // cancel immediately
			clustersClienterMock.On("GetClusterIds", mock.Anything, int32(1001)).Return([]string{"clusterId_value"}, nil).Maybe()
			result := collector.Collect(ctx, []int32{1001})
			Expect(result).To(BeEmpty())
		})
	})

	Describe("queryMetric", func() {
		It("should query metric for tenant", func() {
			ctx := context.Background()
			metric := runaimetrics.RunaiExternalPodInfoMetric{}
			tenantId := int32(1001)
			clusterId := "clusterId_value"
			result, err := queryMetric(ctx, metric, tenantId, clusterId)
			Expect(result).To(Equal(&mockRunaiExternalPodInfoMetricQueryResponse))
			Expect(err).To(BeNil())
		})
	})

	Describe("mapLabels", func() {
		It("should map labels correctly", func() {
			mapping := map[string]string{
				"source_label_1": "target_label_1",
				"source_label_2": "target_label_2",
			}

			labels := map[string]string{
				"source_label_1": "value1",
				"source_label_2": "value2",
				"other_label":    "value3",
			}

			result := mapLabels(mapping, labels, "test_metric")

			Expect(result).To(Equal(map[string]string{
				"target_label_1": "value1",
				"target_label_2": "value2",
			}))
		})

		It("should return empty map when mapping is empty", func() {
			mapping := map[string]string{}

			labels := map[string]string{
				"label1": "value1",
			}

			result := mapLabels(mapping, labels, "test_metric")

			Expect(result).To(BeEmpty())
		})
	})
})
