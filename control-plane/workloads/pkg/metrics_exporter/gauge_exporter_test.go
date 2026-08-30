package metrics_exporter

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/workloads/pkg/metrics_exporter/mocks"
	"github.com/run-ai/runai/control-plane/workloads/pkg/metrics_exporter/models"
	"github.com/run-ai/runai/control-plane/workloads/pkg/runaimetrics"
	metrics_mocks "github.com/run-ai/runai/runai-common-packages/metrics/mocks"
	metrics_testing "github.com/run-ai/runai/runai-common-packages/metrics/testing"
)

var _ = Describe("GaugeExporter", func() {
	Describe("NewGaugeExporter", func() {
		It("should create a new GaugeExporter", func() {
			registry := metrics_mocks.NewMockMetricsRegistrar(GinkgoT())
			exportableMetric := &models.ExportableMetric{
				Description:   "test_metric_description",
				Name:          "test_metric",
				LabelsMapping: map[string]string{},
				Metric:        runaimetrics.RunaiExternalPodInfoMetric{},
			}

			gaugeExporter := NewGaugeExporter(registry, exportableMetric)
			Expect(gaugeExporter).NotTo(BeNil())
			Expect(gaugeExporter.registry).To(Equal(registry))
			Expect(gaugeExporter.exportableMetric).To(Equal(exportableMetric))
			Expect(gaugeExporter.collector).NotTo(BeNil())
		})
	})

	Describe("Export", func() {
		It("should export the metric records", func() {
			ctx := context.Background()
			tenantIds := []int32{1001}
			exportableMetric := &models.ExportableMetric{
				Description: "test_metric_description",
				Name:        "test_metric",
				Metric: runaimetrics.RunaiExternalPodInfoMetric{
					ResponseLabels: runaimetrics.RunaiExternalPodInfoResponseLabels{
						ClusterId: true,
						PodId:     true,
					},
				},
			}

			metricsManagerMock := metrics_testing.SetMock(GinkgoT())
			metricsManagerMock.GaugeVec(exportableMetric.Name).Get().Reset()
			metricsManagerMock.GaugeVec(exportableMetric.Name).Get().WithLabels(map[string]string{
				"tenant_id": "1001",
				"clusterId": "clusterId_1001",
			}).Set(1)
			metricsManagerMock.GaugeVec(exportableMetric.Name).Get().WithLabels(map[string]string{
				"tenant_id": "1002",
				"clusterId": "clusterId_1002",
			}).Set(0)

			collectorMock := mocks.NewMetricsCollectorer(GinkgoT())
			collectorMock.On("Collect", ctx, tenantIds).Return([]models.MetricRecord{
				{
					Labels: map[string]string{
						"tenant_id": "1001",
						"clusterId": "clusterId_1001",
					},
					Value: 1,
				},
				{
					Labels: map[string]string{
						"tenant_id": "1002",
						"clusterId": "clusterId_1002",
					},
					Value: 0,
				},
			})

			gaugeExporter := &GaugeExporter{
				registry:         metricsManagerMock.GetMetricsRegistrar(),
				exportableMetric: exportableMetric,
				collector:        collectorMock,
			}
			gaugeExporter.Export(ctx, tenantIds)
		})

		It("should not export metrics if received nil metricRecords", func() {
			ctx := context.Background()
			tenantIds := []int32{1001}
			exportableMetric := &models.ExportableMetric{
				Description: "test_metric_description",
				Name:        "test_metric",
				Metric: runaimetrics.RunaiExternalPodInfoMetric{
					ResponseLabels: runaimetrics.RunaiExternalPodInfoResponseLabels{
						ClusterId: true,
						PodId:     true,
					},
				},
			}

			metricsManagerMock := metrics_testing.SetMock(GinkgoT())
			metricsManagerMock.GaugeVec(exportableMetric.Name).Get().Reset()

			collectorMock := mocks.NewMetricsCollectorer(GinkgoT())
			collectorMock.On("Collect", ctx, tenantIds).Return(nil)
			gaugeExporter := &GaugeExporter{
				registry:         metricsManagerMock.GetMetricsRegistrar(),
				exportableMetric: exportableMetric,
				collector:        collectorMock,
			}
			gaugeExporter.Export(ctx, tenantIds)
		})
	})
})
