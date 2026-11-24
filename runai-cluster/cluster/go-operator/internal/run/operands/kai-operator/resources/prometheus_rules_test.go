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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("resources_per_nodepool", func() {
	var expr string

	BeforeEach(func() {
		rule := desiredPrometheusRule(defaultRunaiConfig())
		Expect(rule.Spec.Groups).To(HaveLen(1))
		Expect(rule.Spec.Groups[0].Rules).To(HaveLen(1))
		Expect(rule.Spec.Groups[0].Rules[0].Record).To(Equal("resources_per_nodepool"))
		expr = rule.Spec.Groups[0].Rules[0].Expr.String()
	})

	It("keeps summing node capacity for device-plugin resources", func() {
		Expect(expr).To(ContainSubstring("kube_node_status_capacity * on(node) group_left(nodepool)"))
	})

	It("adds the DCGM-derived GPU count for DRA-only nodes that lack nvidia_com_gpu capacity", func() {
		Expect(expr).To(ContainSubstring(`runai_gpu_count_per_node > 0`))
		Expect(expr).To(ContainSubstring(`"resource", "nvidia_com_gpu"`))
	})

	It("ignores nodes whose device plugin still reports GPU capacity so they are not double counted", func() {
		Expect(expr).To(ContainSubstring(`unless on(node) (kube_node_status_capacity{resource="nvidia_com_gpu"} > 0)`))
	})
})

var _ = Describe("PrometheusRuleStatus", func() {
	newClient := func(objects ...client.Object) client.Client {
		testScheme := runtime.NewScheme()
		Expect(monitoringv1.AddToScheme(testScheme)).To(Succeed())
		return fake.NewClientBuilder().WithScheme(testScheme).WithObjects(objects...).Build()
	}

	It("is not ready when the live rule is missing", func() {
		status, err := PrometheusRuleStatus(context.Background(), newClient(), defaultRunaiConfig())
		Expect(err).NotTo(HaveOccurred())
		Expect(status.Ready).To(BeFalse())
		Expect(status.Reasons).To(ContainElement(ContainSubstring("is missing")))
	})

	It("is not ready when the live rule is stale", func() {
		runaiConfig := defaultRunaiConfig()
		rule := desiredPrometheusRule(runaiConfig)
		rule.UID = types.UID("existing")
		rule.Spec.Groups = nil

		status, err := PrometheusRuleStatus(context.Background(), newClient(rule), runaiConfig)
		Expect(err).NotTo(HaveOccurred())
		Expect(status.Ready).To(BeFalse())
		Expect(status.Reasons).To(ContainElement(ContainSubstring("is out of date")))
	})

	It("is ready when the live rule matches the desired state", func() {
		runaiConfig := defaultRunaiConfig()
		rule := desiredPrometheusRule(runaiConfig)
		rule.UID = types.UID("existing")

		status, err := PrometheusRuleStatus(context.Background(), newClient(rule), runaiConfig)
		Expect(err).NotTo(HaveOccurred())
		Expect(status.Ready).To(BeTrue())
		Expect(status.Reasons).To(BeEmpty())
	})
})
