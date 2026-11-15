/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics_exporter

import (
	"context"

	metrics_client "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/metrics_exporter/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/metrics_exporter/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/runaimetrics"
	metrics_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

		DescribeTable("fills clusterId from the cluster used for each query",
			func(responseLabels map[string]string) {
				stubPodMetricResponse(exportableMetric.Metric.GetName(), openapi.Result{
					MetricLabels: responseLabels,
					Measurements: []openapi.Measurement{{Value: "1"}},
				})
				clustersClienterMock.On("GetClusterIds", mock.Anything, int32(1001)).
					Return([]string{"cluster-1", "cluster-2"}, nil)

				records := collector.Collect(context.Background(), []int32{1001})

				Expect(records).To(ConsistOf(
					models.MetricRecord{
						Labels: map[string]string{
							"tenant_id":        "1001",
							"mapped_clusterId": "cluster-1",
							"mapped_podId":     "pod-1",
						},
						Value: 1,
					},
					models.MetricRecord{
						Labels: map[string]string{
							"tenant_id":        "1001",
							"mapped_clusterId": "cluster-2",
							"mapped_podId":     "pod-1",
						},
						Value: 1,
					},
				))
			},
			Entry("when clusterId is absent", map[string]string{"pod_uuid": "pod-1"}),
			Entry("when clusterId is empty", map[string]string{"pod_uuid": "pod-1", "clusterId": ""}),
		)

		It("preserves a clusterId supplied by the metrics service", func() {
			stubPodMetricResponse(exportableMetric.Metric.GetName(), openapi.Result{
				MetricLabels: map[string]string{"pod_uuid": "pod-1", "clusterId": "reported-cluster"},
				Measurements: []openapi.Measurement{{Value: "1"}},
			})
			clustersClienterMock.On("GetClusterIds", mock.Anything, int32(1001)).
				Return([]string{"queried-cluster"}, nil)

			records := collector.Collect(context.Background(), []int32{1001})

			Expect(records).To(Equal([]models.MetricRecord{{
				Labels: map[string]string{
					"tenant_id":        "1001",
					"mapped_clusterId": "reported-cluster",
					"mapped_podId":     "pod-1",
				},
				Value: 1,
			}}))
		})

		It("should stop collection and return early when context is done", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // cancel immediately
			clustersClienterMock.On("GetClusterIds", mock.Anything, int32(1001)).Return([]string{"clusterId_value"}, nil).Maybe()
			result := collector.Collect(ctx, []int32{1001})
			Expect(result).To(BeEmpty())
		})
	})

	Describe("pod metric export", func() {
		var (
			metric                 *models.ExportableMetric
			exporter               *GaugeExporter
			registry               metrics_pkg.MetricsRegistrar
			crashingPod            openapi.Result
			otherPod               openapi.Result
			expectedOtherPodRecord models.MetricRecord
		)

		BeforeEach(func() {
			metric = exportedMetrics[0]
			crashingPod = podMetricResult("crashing-pod", "CrashLoopBackOff")
			otherPod = podMetricResult("other-pod", "Error")
			expectedOtherPodRecord = models.MetricRecord{
				Labels: map[string]string{
					"tenant_id":   "1011",
					"cluster_id":  "test-cluster",
					"node":        "test-node",
					"pod_id":      "other-pod",
					"workload_id": "test-workload",
					"status":      "Error",
				},
				Value: 1,
			}
			clusters := mocks.NewClustersClienter(GinkgoT())
			clusters.On("GetClusterIds", mock.Anything, int32(1011)).Return([]string{"test-cluster"}, nil)
			collector := &MetricsCollector{
				exportableMetric: metric, clustersClienter: clusters,
				concurrentLimitCh: make(chan struct{}, 1),
			}
			registry = metrics_pkg.Registry("pod-metric-export-test-" + CurrentSpecReport().LeafNodeText)
			registry.GaugeVec(metric.Name).Register(metric.Description, getLabels(metric))
			exporter = &GaugeExporter{registry: registry, exportableMetric: metric, collector: collector}
		})

		exportAndGatherPodMetrics := func(results ...openapi.Result) []models.MetricRecord {
			GinkgoHelper()
			stubPodMetricResponse(metric.Metric.GetName(), results...)

			Expect(func() { exporter.Export(context.Background(), []int32{1011}) }).NotTo(Panic())

			families, err := registry.GetPrometheusRegistry().Gather()
			Expect(err).NotTo(HaveOccurred())
			Expect(families).To(HaveLen(1))
			var records []models.MetricRecord
			for _, sample := range families[0].GetMetric() {
				labels := map[string]string{}
				for _, label := range sample.GetLabel() {
					labels[label.GetName()] = label.GetValue()
				}
				records = append(records, models.MetricRecord{Labels: labels, Value: sample.GetGauge().GetValue()})
			}
			return records
		}

		It("exports both pods after restoring the missing clusterId", func() {
			delete(crashingPod.MetricLabels, "clusterId")

			records := exportAndGatherPodMetrics(crashingPod, otherPod)

			Expect(records).To(ConsistOf(expectedOtherPodRecord, models.MetricRecord{
				Labels: map[string]string{
					"tenant_id":   "1011",
					"cluster_id":  "test-cluster",
					"node":        "test-node",
					"pod_id":      "crashing-pod",
					"workload_id": "test-workload",
					"status":      "CrashLoopBackOff",
				},
				Value: 1,
			}))
		})

		DescribeTable("skips the incomplete pod and exports the next pod",
			func(missingLabel string) {
				delete(crashingPod.MetricLabels, missingLabel)

				records := exportAndGatherPodMetrics(crashingPod, otherPod)

				Expect(records).To(Equal([]models.MetricRecord{expectedOtherPodRecord}))
			},
			Entry("when node is absent", "node"),
			Entry("when pod_uuid is absent", "pod_uuid"),
			Entry("when workload_id is absent", "workload_id"),
			Entry("when detailed_status is absent", "detailed_status"),
		)
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

func stubPodMetricResponse(metricName string, results ...openapi.Result) {
	GinkgoHelper()
	previous := mockMetricServerResponseMap[metricName]
	DeferCleanup(func() { mockMetricServerResponseMap[metricName] = previous })
	mockMetricServerResponseMap[metricName] = openapi.GetMetrics200Response{Status: "success", Data: results}
}

func podMetricResult(podId, status string) openapi.Result {
	return openapi.Result{
		MetricLabels: map[string]string{
			"clusterId":       "test-cluster",
			"node":            "test-node",
			"pod_uuid":        podId,
			"workload_id":     "test-workload",
			"detailed_status": status,
		},
		Measurements: []openapi.Measurement{{Value: "1"}},
	}
}
