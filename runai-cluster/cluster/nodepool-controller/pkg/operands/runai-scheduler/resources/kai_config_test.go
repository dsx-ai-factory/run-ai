/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	usagedbapi "github.com/kai-scheduler/KAI-scheduler/pkg/scheduler/cache/usagedb/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Non-runai flag combination test (DoD).
// Verifies that when the binary is configured with all KAI-style values
// (instead of the runai defaults the rest of the suite uses), production
// code consumers of `config.Get()` actually pick up the KAI vocabulary —
// proving the flag values flow through to the SchedulingShard /
// ServiceMonitor output rather than being pinned to compile-time SDK consts.
var _ = Describe("nodepool-controller under KAI (non-runai) config", func() {
	const (
		kaiNodePoolNameLabel   = "kai.scheduler/node-pool"
		kaiDefaultNodepoolName = "default"
		kaiSchedulerName       = "kai-scheduler"
		kaiSchedulerNamespace  = "kai-scheduler"
		kaiMetricsNamespace    = "kai"
		kaiFinalizerDomain     = "kai.scheduler"
		kaiExpectedFinalizer   = "nodepool.kai.scheduler/finalize"
	)

	BeforeEach(func() {
		DeferCleanup(config.SetForTest(&config.NodePoolControllerConfig{
			NodePoolNameLabel:   kaiNodePoolNameLabel,
			DefaultNodepoolName: kaiDefaultNodepoolName,
			SchedulerName:       kaiSchedulerName,
			SchedulerNamespace:  kaiSchedulerNamespace,
			MetricsNamespace:    kaiMetricsNamespace,
			FinalizerDomain:     kaiFinalizerDomain,
		}))
	})

	It("exposes the KAI label keys / names via config.Get()", func() {
		Expect(config.Get().NodePoolNameLabel).To(Equal(kaiNodePoolNameLabel))
		Expect(config.Get().DefaultNodepoolName).To(Equal(kaiDefaultNodepoolName))
		Expect(config.Get().SchedulerName).To(Equal(kaiSchedulerName))
		Expect(config.Get().SchedulerNamespace).To(Equal(kaiSchedulerNamespace))
		Expect(config.Get().MetricsNamespace).To(Equal(kaiMetricsNamespace))
	})

	It("composes the finalizer string from the KAI finalizer-domain", func() {
		Expect(config.FinalizerName()).To(Equal(kaiExpectedFinalizer))
	})

	It("derives the scheduler base operand name from --scheduler-name", func() {
		Expect(SchedulerBaseOperandName()).To(Equal(kaiSchedulerName))
	})

	It("returns empty PartitionLabelValue for the configured default nodepool", func() {
		Expect(getNodePoolNameLabelValueForScheduler(kaiDefaultNodepoolName)).To(Equal(""))
		Expect(getNodePoolNameLabelValueForScheduler("some-other-pool")).To(Equal("some-other-pool"))
	})

	Describe("ServiceMonitorForNodePool", func() {
		It("names the ServiceMonitor with the KAI scheduler-name prefix and writes it into the configured namespace", func() {
			scheme := runtime.NewScheme()
			Expect(monitorv1.AddToScheme(scheme)).To(Succeed())
			Expect(corev1.AddToScheme(scheme)).To(Succeed())
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			nodePool := &v1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: "np-1"}}

			obj, err := ServiceMonitorForNodePool(
				context.Background(), fakeClient, nodePool, &common.NodePoolControllerParams{}, "np-1")
			Expect(err).NotTo(HaveOccurred())

			sm, ok := obj.(*monitorv1.ServiceMonitor)
			Expect(ok).To(BeTrue())
			Expect(sm.Name).To(Equal(kaiSchedulerName + "-np-1"))
			Expect(sm.Namespace).To(Equal(kaiSchedulerNamespace))
			Expect(sm.Labels["app"]).To(Equal(kaiSchedulerName))
			Expect(sm.Spec.JobLabel).To(Equal(kaiSchedulerName))
		})
	})

	Describe("overrideCapacityMetricsIfNeeded", func() {
		It("prefixes the queue-allocated metric names with --metrics-namespace", func() {
			params := &usagedbapi.UsageParams{}
			overrideCapacityMetricsIfNeeded(params, "np-1")
			Expect(params.ExtraParams["gpuAllocationMetric"]).To(Equal("kai_queue_allocated_gpus"))
			Expect(params.ExtraParams["cpuAllocationMetric"]).To(Equal("kai_queue_allocated_cpu_cores"))
			Expect(params.ExtraParams["memoryAllocationMetric"]).To(Equal("kai_queue_allocated_memory_bytes"))
		})
	})

	Describe("SchedulingShardForNodePool", func() {
		It("uses the configured scheduler-name in the resource labels and reads the configured default-nodepool-name when computing PartitionLabelValue", func() {
			scheme := runtime.NewScheme()
			Expect(kaiv1.AddToScheme(scheme)).To(Succeed())
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			params := &common.NodePoolControllerParams{}

			defaultNP := &v1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: kaiDefaultNodepoolName}}
			obj, err := SchedulingShardForNodePool(context.Background(), fakeClient, defaultNP, params, kaiDefaultNodepoolName)
			Expect(err).NotTo(HaveOccurred())
			Expect(obj.GetLabels()["app"]).To(Equal(kaiSchedulerName))
			// PartitionLabelValue is "" for the default nodepool — derived
			// from config.Get().DefaultNodepoolName, not from a hardcoded
			// v1alpha1.DefaultNodePoolName.
			shard, ok := obj.(*kaiv1.SchedulingShard)
			Expect(ok).To(BeTrue())
			Expect(shard.Spec.PartitionLabelValue).To(Equal(""))

			// And for a non-default nodepool, PartitionLabelValue equals its name.
			namedNP := &v1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: "np-named"}}
			obj2, err := SchedulingShardForNodePool(context.Background(), fakeClient, namedNP, params, "np-named")
			Expect(err).NotTo(HaveOccurred())
			Expect(obj2.(*kaiv1.SchedulingShard).Spec.PartitionLabelValue).To(Equal("np-named"))
		})

		It("merges the cluster-wide scheduler args over the base worker-label args", func() {
			scheme := runtime.NewScheme()
			Expect(kaiv1.AddToScheme(scheme)).To(Succeed())
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			params := &common.NodePoolControllerParams{
				SchedulingShardArgs: map[string]string{
					"v":                         "6",
					"full-hierarchy-fairness":   "false",
					"cpu-worker-node-label-key": "should-be-ignored",
				},
			}

			nodePool := &v1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: "np-args"}}
			obj, err := SchedulingShardForNodePool(context.Background(), fakeClient, nodePool, params, nodePool.Name)
			Expect(err).NotTo(HaveOccurred())

			args := obj.(*kaiv1.SchedulingShard).Spec.Args
			// cluster-wide overrides land in the shard...
			Expect(args).To(HaveKeyWithValue("v", "6"))
			Expect(args).To(HaveKeyWithValue("full-hierarchy-fairness", "false"))
			// ...but the controller-owned base args are authoritative and
			// cannot be overridden by a cluster-wide arg.
			Expect(args).To(HaveKeyWithValue("cpu-worker-node-label-key", cpuWorkerNodeLabelKey))
			Expect(args).To(HaveKeyWithValue("gpu-worker-node-label-key", gpuWorkerNodeLabelKey))
			Expect(args).To(HaveKeyWithValue("mig-worker-node-label-key", migWorkerNodeLabelKey))
		})

		It("sets only the base worker-label args when no scheduler args are provided", func() {
			scheme := runtime.NewScheme()
			Expect(kaiv1.AddToScheme(scheme)).To(Succeed())
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			nodePool := &v1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: "np-empty"}}
			obj, err := SchedulingShardForNodePool(context.Background(), fakeClient, nodePool,
				&common.NodePoolControllerParams{}, nodePool.Name)
			Expect(err).NotTo(HaveOccurred())

			Expect(obj.(*kaiv1.SchedulingShard).Spec.Args).To(HaveLen(3))
		})

		It("updates the existing shard spec without dropping metadata or status", func() {
			scheme := runtime.NewScheme()
			Expect(kaiv1.AddToScheme(scheme)).To(Succeed())

			params := &common.NodePoolControllerParams{}
			nodePool := &v1alpha1.NodePool{ObjectMeta: metav1.ObjectMeta{Name: "np-existing"}}

			expectedObject, err := SchedulingShardForNodePool(
				context.Background(), fake.NewClientBuilder().WithScheme(scheme).Build(), nodePool, params, nodePool.Name)
			Expect(err).NotTo(HaveOccurred())

			existing := &kaiv1.SchedulingShard{
				ObjectMeta: metav1.ObjectMeta{
					Name:            nodePool.Name,
					UID:             types.UID("existing-shard-uid"),
					ResourceVersion: "17",
					Labels:          map[string]string{"app": "stale-scheduler", "external-label": "preserve"},
					Annotations:     map[string]string{"external-annotation": "preserve"},
					Finalizers:      []string{"external.example/finalizer"},
					OwnerReferences: []metav1.OwnerReference{{
						APIVersion: "run.ai/v1alpha1",
						Kind:       "NodePool",
						Name:       nodePool.Name,
						UID:        types.UID("nodepool-uid"),
					}},
				},
				Spec: kaiv1.SchedulingShardSpec{
					PartitionLabelValue: "stale-partition",
					Args:                map[string]string{"stale": "true"},
				},
				Status: kaiv1.SchedulingShardStatus{Conditions: []metav1.Condition{{
					Type:    string(kaiv1.ConditionTypeAvailable),
					Status:  metav1.ConditionTrue,
					Message: "preserve status",
				}}},
			}
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existing).Build()

			obj, err := SchedulingShardForNodePool(context.Background(), fakeClient, nodePool, params, nodePool.Name)
			Expect(err).NotTo(HaveOccurred())
			shard := obj.(*kaiv1.SchedulingShard)

			Expect(shard.Spec).To(Equal(expectedObject.(*kaiv1.SchedulingShard).Spec))
			Expect(shard.UID).To(Equal(existing.UID))
			Expect(shard.ResourceVersion).To(Equal(existing.ResourceVersion))
			Expect(shard.Labels).To(HaveKeyWithValue("app", kaiSchedulerName))
			Expect(shard.Labels).To(HaveKeyWithValue("external-label", "preserve"))
			Expect(shard.Annotations).To(Equal(existing.Annotations))
			Expect(shard.Finalizers).To(Equal(existing.Finalizers))
			Expect(shard.OwnerReferences).To(Equal(existing.OwnerReferences))
			Expect(shard.Status).To(Equal(existing.Status))
		})
	})
})
