/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	usagedbapi "github.com/kai-scheduler/KAI-scheduler/pkg/scheduler/cache/usagedb/api"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("getTimeBasedFairShareFields", func() {
	var (
		nodePool *v1alpha1.NodePool
	)

	BeforeEach(func() {
		nodePool = &v1alpha1.NodePool{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-nodepool",
			},
			Spec: v1alpha1.NodePoolSpec{
				SchedulingShardConfig: &v1alpha1.SchedulingShardConfig{
					TimeBasedFairShare: &v1alpha1.TimeBasedFairShare{},
				},
			},
		}
	})

	Context("when TimeBasedFairShare is disabled", func() {
		BeforeEach(func() {
			nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.Enabled = ptr.To(false)
		})

		It("should return nil", func() {
			result := getTimeBasedFairShareFields(nodePool)

			Expect(result).To(BeNil())
		})
	})

	Context("when TimeBasedFairShare is enabled", func() {
		BeforeEach(func() {
			nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.Enabled = ptr.To(true)
		})

		Context("with default values", func() {
			It("should initialize UsageDBConfig", func() {
				result := getTimeBasedFairShareFields(nodePool)

				Expect(result).NotTo(BeNil())
				Expect(result.UsageParams).NotTo(BeNil())
			})

			It("should set HalfLifePeriod to default (0s)", func() {
				result := getTimeBasedFairShareFields(nodePool)

				Expect(result.UsageParams.HalfLifePeriod).NotTo(BeNil())
				Expect(result.UsageParams.HalfLifePeriod.Duration).To(Equal(0 * time.Second))
			})

			It("should set WindowSize to default (7 days)", func() {
				result := getTimeBasedFairShareFields(nodePool)

				Expect(result.UsageParams.WindowSize).NotTo(BeNil())
				Expect(*result.UsageParams.WindowSize).To(Equal(monitoringv1.Duration("168h0m0s")))
			})

			It("should set WindowType to default (sliding)", func() {
				result := getTimeBasedFairShareFields(nodePool)

				Expect(result.UsageParams.WindowType).NotTo(BeNil())
				Expect(*result.UsageParams.WindowType).To(Equal(usagedbapi.SlidingWindow))
			})

			It("should set TumblingWindowStartTime to nil", func() {
				result := getTimeBasedFairShareFields(nodePool)

				Expect(result.UsageParams.TumblingWindowStartTime).To(BeNil())
			})
		})

		Context("with custom HalfLifePeriod", func() {
			It("should parse valid duration string", func() {
				nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.HalfLifePeriod = ptr.To("24h")

				result := getTimeBasedFairShareFields(nodePool)

				Expect(result.UsageParams.HalfLifePeriod).NotTo(BeNil())
				Expect(result.UsageParams.HalfLifePeriod.Duration).To(Equal(24 * time.Hour))
			})

			It("should parse complex duration string", func() {
				nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.HalfLifePeriod = ptr.To("2d12h30m")

				result := getTimeBasedFairShareFields(nodePool)

				expectedDuration := 2*24*time.Hour + 12*time.Hour + 30*time.Minute
				Expect(result.UsageParams.HalfLifePeriod).NotTo(BeNil())
				Expect(result.UsageParams.HalfLifePeriod.Duration).To(Equal(expectedDuration))
			})

			It("should use default for invalid duration string", func() {
				nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.HalfLifePeriod = ptr.To("invalid")

				result := getTimeBasedFairShareFields(nodePool)

				Expect(result.UsageParams.HalfLifePeriod).NotTo(BeNil())
				Expect(result.UsageParams.HalfLifePeriod.Duration).To(Equal(0 * time.Second))
			})
		})

		Context("with custom Window.Size", func() {
			It("should parse valid duration string", func() {
				nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.Window = &v1alpha1.FairShareWindow{
					Size: ptr.To("14d"),
				}

				result := getTimeBasedFairShareFields(nodePool)

				Expect(result.UsageParams.WindowSize).NotTo(BeNil())
				Expect(*result.UsageParams.WindowSize).To(Equal(monitoringv1.Duration("14d")))
			})

			It("should parse hours duration", func() {
				nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.Window = &v1alpha1.FairShareWindow{
					Size: ptr.To("168h"),
				}

				result := getTimeBasedFairShareFields(nodePool)

				Expect(result.UsageParams.WindowSize).NotTo(BeNil())
				Expect(*result.UsageParams.WindowSize).To(Equal(monitoringv1.Duration("168h")))
			})

			It("should use default for invalid duration string", func() {
				nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.Window = &v1alpha1.FairShareWindow{
					Size: ptr.To("not-a-duration"),
				}

				result := getTimeBasedFairShareFields(nodePool)

				Expect(result.UsageParams.WindowSize).NotTo(BeNil())
				Expect(*result.UsageParams.WindowSize).To(Equal(monitoringv1.Duration("168h0m0s")))
			})
		})

		Context("with custom Window.Type", func() {
			It("should set tumbling window type", func() {
				nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.Window = &v1alpha1.FairShareWindow{
					Type: ptr.To(string(usagedbapi.TumblingWindow)),
				}

				result := getTimeBasedFairShareFields(nodePool)

				Expect(result.UsageParams.WindowType).NotTo(BeNil())
				Expect(*result.UsageParams.WindowType).To(Equal(usagedbapi.TumblingWindow))
			})

			It("should set sliding window type", func() {
				nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.Window = &v1alpha1.FairShareWindow{
					Type: ptr.To(string(usagedbapi.SlidingWindow)),
				}

				result := getTimeBasedFairShareFields(nodePool)

				Expect(result.UsageParams.WindowType).NotTo(BeNil())
				Expect(*result.UsageParams.WindowType).To(Equal(usagedbapi.SlidingWindow))
			})

			It("should use default for invalid window type", func() {
				nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.Window = &v1alpha1.FairShareWindow{
					Type: ptr.To("invalid"),
				}

				result := getTimeBasedFairShareFields(nodePool)

				Expect(result.UsageParams.WindowType).NotTo(BeNil())
				Expect(*result.UsageParams.WindowType).To(Equal(usagedbapi.SlidingWindow))
			})
		})

		Context("with Window.TumblingStartTime", func() {
			It("should set the start time when provided", func() {
				startTime := metav1.NewTime(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
				nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.Window = &v1alpha1.FairShareWindow{
					TumblingStartTime: &startTime,
				}

				result := getTimeBasedFairShareFields(nodePool)

				Expect(result.UsageParams.TumblingWindowStartTime).NotTo(BeNil())
				Expect(result.UsageParams.TumblingWindowStartTime.Time).To(Equal(startTime.Time))
			})

			It("should set nil when not provided", func() {
				nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.Window = &v1alpha1.FairShareWindow{
					TumblingStartTime: nil,
				}

				result := getTimeBasedFairShareFields(nodePool)

				Expect(result.UsageParams.TumblingWindowStartTime).To(BeNil())
			})
		})

		Context("with all custom values", func() {
			It("should set all fields correctly", func() {
				weight := float32(0.75)
				startTime := metav1.NewTime(time.Date(2024, 6, 1, 8, 0, 0, 0, time.UTC))

				tbfs := nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare
				tbfs.HistoricalUsageWeight = &weight
				tbfs.HalfLifePeriod = ptr.To("48h")
				tbfs.Window = &v1alpha1.FairShareWindow{
					Type:              ptr.To(string(usagedbapi.TumblingWindow)),
					Size:              ptr.To("30d"),
					TumblingStartTime: &startTime,
				}

				result := getTimeBasedFairShareFields(nodePool)

				Expect(result.UsageParams.HalfLifePeriod.Duration).To(Equal(48 * time.Hour))
				Expect(*result.UsageParams.WindowSize).To(Equal(monitoringv1.Duration("30d")))
				Expect(*result.UsageParams.WindowType).To(Equal(usagedbapi.TumblingWindow))
				Expect(result.UsageParams.TumblingWindowStartTime.Time).To(Equal(startTime.Time))
			})
		})
	})
})

var _ = Describe("getKValue", func() {
	var nodePool *v1alpha1.NodePool

	BeforeEach(func() {
		nodePool = &v1alpha1.NodePool{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-nodepool",
			},
			Spec: v1alpha1.NodePoolSpec{
				SchedulingShardConfig: &v1alpha1.SchedulingShardConfig{
					TimeBasedFairShare: &v1alpha1.TimeBasedFairShare{},
				},
			},
		}
	})

	Context("when TimeBasedFairShare is disabled", func() {
		BeforeEach(func() {
			nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.Enabled = ptr.To(false)
		})

		It("should return nil", func() {
			result := getKValue(nodePool)

			Expect(result).To(BeNil())
		})
	})

	Context("when TimeBasedFairShare is enabled", func() {
		BeforeEach(func() {
			nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.Enabled = ptr.To(true)
		})

		It("should return default value (1.0) when HistoricalUsageWeight is nil", func() {
			nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.HistoricalUsageWeight = nil

			result := getKValue(nodePool)

			Expect(result).NotTo(BeNil())
			Expect(*result).To(BeNumerically("==", 1.0))
		})

		It("should return the provided HistoricalUsageWeight value", func() {
			weight := float32(0.5)
			nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.HistoricalUsageWeight = &weight

			result := getKValue(nodePool)

			Expect(result).NotTo(BeNil())
			Expect(*result).To(BeNumerically("==", 0.5))
		})

		It("should handle zero value", func() {
			weight := float32(0.0)
			nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.HistoricalUsageWeight = &weight

			result := getKValue(nodePool)

			Expect(result).NotTo(BeNil())
			Expect(*result).To(BeNumerically("==", 0.0))
		})

		It("should handle values greater than 1", func() {
			weight := float32(2.5)
			nodePool.Spec.SchedulingShardConfig.TimeBasedFairShare.HistoricalUsageWeight = &weight

			result := getKValue(nodePool)

			Expect(result).NotTo(BeNil())
			Expect(*result).To(BeNumerically("==", 2.5))
		})
	})
})

var _ = Describe("parseDurationParam", func() {
	Context("when param is nil", func() {
		It("should return the default duration", func() {
			result := parseDurationParam(nil, 5*time.Hour, "TestParam", "test-pool")

			Expect(result).NotTo(BeNil())
			Expect(result.Duration).To(Equal(5 * time.Hour))
		})
	})

	Context("when param is a valid duration", func() {
		It("should parse hours correctly", func() {
			param := "12h"
			result := parseDurationParam(&param, 0, "TestParam", "test-pool")

			Expect(result).NotTo(BeNil())
			Expect(result.Duration).To(Equal(12 * time.Hour))
		})

		It("should parse days correctly", func() {
			param := "7d"
			result := parseDurationParam(&param, 0, "TestParam", "test-pool")

			Expect(result).NotTo(BeNil())
			Expect(result.Duration).To(Equal(7 * 24 * time.Hour))
		})

		It("should parse minutes correctly", func() {
			param := "30m"
			result := parseDurationParam(&param, 0, "TestParam", "test-pool")

			Expect(result).NotTo(BeNil())
			Expect(result.Duration).To(Equal(30 * time.Minute))
		})

		It("should parse complex durations", func() {
			param := "1d2h3m"
			result := parseDurationParam(&param, 0, "TestParam", "test-pool")

			expected := 24*time.Hour + 2*time.Hour + 3*time.Minute
			Expect(result).NotTo(BeNil())
			Expect(result.Duration).To(Equal(expected))
		})
	})

	Context("when param is an invalid duration", func() {
		It("should return the default duration", func() {
			param := "invalid-duration"
			defaultDuration := 3 * time.Hour
			result := parseDurationParam(&param, defaultDuration, "TestParam", "test-pool")

			Expect(result).NotTo(BeNil())
			Expect(result.Duration).To(Equal(defaultDuration))
		})

		It("should return the default duration for empty string", func() {
			param := ""
			defaultDuration := 6 * time.Hour
			result := parseDurationParam(&param, defaultDuration, "TestParam", "test-pool")

			Expect(result).NotTo(BeNil())
			Expect(result.Duration).To(Equal(defaultDuration))
		})
	})
})

var _ = Describe("SchedulingShardForNodePool SchedulingShardConfig pass-through", func() {
	// numaPlugin mirrors the "numa" plugin key the nodepool controller enables NUMA-aware
	// scheduling with (it lives in another package, so it is spelled out here as a const)
	const numaPlugin = "numa"

	var params *common.NodePoolControllerParams

	BeforeEach(func() {
		params = &common.NodePoolControllerParams{}
	})

	buildShardSpec := func(cfg *v1alpha1.SchedulingShardConfig) kaiv1.SchedulingShardSpec {
		scheme := runtime.NewScheme()
		Expect(kaiv1.AddToScheme(scheme)).To(Succeed())
		fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

		nodePool := &v1alpha1.NodePool{
			ObjectMeta: metav1.ObjectMeta{Name: "np-passthrough"},
			Spec:       v1alpha1.NodePoolSpec{SchedulingShardConfig: cfg},
		}

		obj, err := SchedulingShardForNodePool(context.Background(), fakeClient, nodePool, params, nodePool.Name)
		Expect(err).NotTo(HaveOccurred())
		shard, ok := obj.(*kaiv1.SchedulingShard)
		Expect(ok).To(BeTrue())
		return shard.Spec
	}

	Context("when SchedulingShardConfig is nil", func() {
		It("leaves plugins, actions, placement strategy, min-runtime and queue depth unset", func() {
			spec := buildShardSpec(nil)

			Expect(spec.Plugins).To(BeNil())
			Expect(spec.Actions).To(BeNil())
			Expect(spec.PlacementStrategy).To(BeNil())
			Expect(spec.MinRuntime).To(BeNil())
			Expect(spec.QueueDepthPerAction).To(BeNil())
		})
	})

	Context("plugins pass-through", func() {
		It("forwards both plugins (numa enabled, gpupack disabled)", func() {
			plugins := map[string]kaiv1.PluginConfig{
				numaPlugin: {Enabled: ptr.To(true), Priority: ptr.To(300)},
				"gpupack":  {Enabled: ptr.To(false)},
			}

			spec := buildShardSpec(&v1alpha1.SchedulingShardConfig{Plugins: plugins})

			Expect(spec.Plugins).To(Equal(plugins))
			Expect(*spec.Plugins[numaPlugin].Enabled).To(BeTrue())
			Expect(*spec.Plugins[numaPlugin].Priority).To(Equal(300))
			Expect(*spec.Plugins["gpupack"].Enabled).To(BeFalse())
		})

		It("forwards a plugin's arguments and priority unchanged", func() {
			plugins := map[string]kaiv1.PluginConfig{
				"gpusharingorder": {
					Enabled:   ptr.To(true),
					Priority:  ptr.To(100),
					Arguments: map[string]string{"key": "value"},
				},
			}

			spec := buildShardSpec(&v1alpha1.SchedulingShardConfig{Plugins: plugins})

			Expect(spec.Plugins).To(Equal(plugins))
		})
	})

	Context("combination of plugins and the other shard-config fields", func() {
		It("forwards plugins together with placement strategy, min-runtime, actions and queue depth", func() {
			cfg := &v1alpha1.SchedulingShardConfig{
				Plugins: map[string]kaiv1.PluginConfig{
					numaPlugin:  {Enabled: ptr.To(true)},
					"gpuspread": {Enabled: ptr.To(true)},
				},
				PlacementStrategy:   &kaiv1.PlacementStrategy{GPU: ptr.To("binpack"), CPU: ptr.To("spread")},
				MinRuntime:          &kaiv1.MinRuntime{PreemptMinRuntime: ptr.To("5m"), ReclaimMinRuntime: ptr.To("10m")},
				Actions:             map[string]kaiv1.ActionConfig{"reclaim": {Enabled: ptr.To(false)}},
				QueueDepthPerAction: map[string]int{"allocate": 100},
			}

			spec := buildShardSpec(cfg)

			Expect(spec.Plugins).To(Equal(cfg.Plugins))
			Expect(spec.PlacementStrategy).To(Equal(cfg.PlacementStrategy))
			Expect(spec.MinRuntime).To(Equal(cfg.MinRuntime))
			Expect(spec.Actions).To(Equal(cfg.Actions))
			Expect(spec.QueueDepthPerAction).To(Equal(cfg.QueueDepthPerAction))
			// The controller-owned base worker-label args are always present
			Expect(spec.Args).To(HaveKeyWithValue("cpu-worker-node-label-key", cpuWorkerNodeLabelKey))
		})

		It("passes plugins through while TimeBasedFairShare is compiled into KValue/UsageDBConfig", func() {
			cfg := &v1alpha1.SchedulingShardConfig{
				Plugins: map[string]kaiv1.PluginConfig{numaPlugin: {Enabled: ptr.To(true)}},
				TimeBasedFairShare: &v1alpha1.TimeBasedFairShare{
					Enabled:               ptr.To(true),
					HistoricalUsageWeight: ptr.To(float32(0.5)),
				},
			}

			spec := buildShardSpec(cfg)

			// Plugins are forwarded verbatim...
			Expect(spec.Plugins).To(HaveKey(numaPlugin))
			Expect(*spec.Plugins[numaPlugin].Enabled).To(BeTrue())
			// ...independently of the fair-share fields, which are compiled elsewhere.
			Expect(spec.KValue).NotTo(BeNil())
			Expect(*spec.KValue).To(Equal(0.5))
			Expect(spec.UsageDBConfig).NotTo(BeNil())
		})
	})
})
