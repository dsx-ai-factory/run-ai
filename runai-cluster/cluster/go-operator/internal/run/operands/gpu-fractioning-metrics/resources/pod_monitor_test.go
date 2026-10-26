/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

var _ = Describe("PodMonitor", func() {
	var (
		k8s         rclient.Client
		runaiConfig *runv1.RunaiConfig
		ctx         context.Context
	)

	BeforeEach(func() {
		k8s = fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).Build()
		ctx = context.Background()
		runaiConfig = &runv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "runai",
				Namespace: "runai",
			},
		}
		runaiConfig.SetDefaultsWhereNeeded()
	})

	buildPodMonitor := func() *monitoringv1.PodMonitor {
		object, err := PodMonitorForRunaiConfig(ctx, k8s, runaiConfig)
		Expect(err).ToNot(HaveOccurred())
		podMonitor, ok := object.(*monitoringv1.PodMonitor)
		Expect(ok).To(BeTrue())
		return podMonitor
	}

	It("is created in the RunaiConfig namespace", func() {
		podMonitor := buildPodMonitor()

		Expect(podMonitor.Namespace).To(Equal("runai"))
		Expect(podMonitor.Name).To(Equal("gpu-fractioning-metrics"))
	})

	It("selects the fractiond pods metricsd runs inside", func() {
		podMonitor := buildPodMonitor()

		Expect(podMonitor.Spec.Selector.MatchLabels).To(Equal(map[string]string{
			"app.kubernetes.io/component":  "fractiond",
			"app.kubernetes.io/managed-by": "gpu-fractioning",
		}))
	})

	It("targets the namespace gpu-fractioning is installed into", func() {
		podMonitor := buildPodMonitor()

		Expect(podMonitor.Spec.NamespaceSelector.MatchNames).To(Equal([]string{"runai"}))
	})

	It("scrapes the metricsd endpoint", func() {
		podMonitor := buildPodMonitor()

		Expect(podMonitor.Spec.PodMetricsEndpoints).To(HaveLen(1))
		endpoint := podMonitor.Spec.PodMetricsEndpoints[0]
		Expect(endpoint.Port).ToNot(BeNil())
		Expect(*endpoint.Port).To(Equal("metrics"))
		Expect(endpoint.Path).To(Equal("/metrics"))
	})

	// Without this the scrape replaces pod and namespace with metricsd's own, and the
	// pod-to-GPU mapping the recording rules join on is lost.
	It("keeps the labels metricsd puts on the series", func() {
		podMonitor := buildPodMonitor()

		Expect(podMonitor.Spec.PodMetricsEndpoints[0].HonorLabels).To(BeTrue())
	})

	// The metric names are set on GpuFractioningConfig, not rewritten at scrape time.
	It("does not relabel the metrics", func() {
		podMonitor := buildPodMonitor()

		Expect(podMonitor.Spec.PodMetricsEndpoints[0].MetricRelabelConfigs).To(BeEmpty())
		Expect(podMonitor.Spec.PodMetricsEndpoints[0].RelabelConfigs).To(BeEmpty())
	})

	It("follows the RunaiConfig namespace rather than a fixed one", func() {
		runaiConfig.Namespace = "runai-tenant"

		podMonitor := buildPodMonitor()

		Expect(podMonitor.Namespace).To(Equal("runai-tenant"))
		Expect(podMonitor.Spec.NamespaceSelector.MatchNames).To(Equal([]string{"runai-tenant"}))
	})

	Describe("Status", func() {
		It("is not ready while the pod monitor is absent from the cluster", func() {
			status, err := PodMonitorStatus(ctx, k8s, runaiConfig)

			Expect(err).ToNot(HaveOccurred())
			Expect(status.Ready).To(BeFalse())
		})

		It("is ready once the pod monitor exists", func() {
			existing := buildPodMonitor()
			existing.UID = types.UID("a-uid")
			Expect(k8s.Create(ctx, existing)).To(Succeed())

			status, err := PodMonitorStatus(ctx, k8s, runaiConfig)

			Expect(err).ToNot(HaveOccurred())
			Expect(status.Ready).To(BeTrue())
		})
	})
})
