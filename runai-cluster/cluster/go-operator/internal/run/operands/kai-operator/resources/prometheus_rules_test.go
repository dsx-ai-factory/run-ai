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
