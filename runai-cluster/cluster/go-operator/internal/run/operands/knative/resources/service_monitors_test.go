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
	"k8s.io/utils/ptr"

	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ServiceMonitors", func() {
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

	Describe("shouldCreateServiceMonitors", func() {
		Context("when CreateServiceMonitors is enabled", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Knative.CreateServiceMonitors = ptr.To(true)
			})

			It("should return true", func() {
				Expect(shouldCreateServiceMonitors(runaiConfig)).To(BeTrue())
			})
		})

		Context("when CreateServiceMonitors is disabled", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Knative.CreateServiceMonitors = ptr.To(false)
			})

			It("should return false", func() {
				Expect(shouldCreateServiceMonitors(runaiConfig)).To(BeFalse())
			})
		})

		Context("when CreateServiceMonitors is nil", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Knative.CreateServiceMonitors = nil
			})

			It("should return false", func() {
				Expect(shouldCreateServiceMonitors(runaiConfig)).To(BeFalse())
			})
		})
	})

	Describe("ServiceMonitorResource", func() {
		Context("when CreateServiceMonitors is enabled", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Knative.CreateServiceMonitors = ptr.To(true)
			})

			It("controller resource should be enabled and return a service monitor", func() {
				resource := NewServiceMonitorResource("knative", k8s, runaiConfig, "controller")
				Expect(resource.Enabled()).To(BeTrue())
				Expect(resource.OperandName()).To(Equal("knative"))

				obj, err := resource.Object(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(obj).ToNot(BeNil())

				sm := obj.(*monitoringv1.ServiceMonitor)
				Expect(sm.Name).To(Equal("knative-serving-controller"))
				Expect(sm.Namespace).To(Equal("runai"))
				Expect(sm.Spec.Selector.MatchLabels["app"]).To(Equal("controller"))
			})

			It("autoscaler resource should be enabled and return a service monitor", func() {
				resource := NewServiceMonitorResource("knative", k8s, runaiConfig, "autoscaler")
				Expect(resource.Enabled()).To(BeTrue())

				obj, err := resource.Object(ctx)
				Expect(err).ToNot(HaveOccurred())

				sm := obj.(*monitoringv1.ServiceMonitor)
				Expect(sm.Name).To(Equal("knative-serving-autoscaler"))
				Expect(sm.Namespace).To(Equal("runai"))
				Expect(sm.Spec.Selector.MatchLabels["app"]).To(Equal("autoscaler"))
			})

			It("activator resource should be enabled and return a service monitor", func() {
				resource := NewServiceMonitorResource("knative", k8s, runaiConfig, "activator")
				Expect(resource.Enabled()).To(BeTrue())

				obj, err := resource.Object(ctx)
				Expect(err).ToNot(HaveOccurred())

				sm := obj.(*monitoringv1.ServiceMonitor)
				Expect(sm.Name).To(Equal("knative-serving-activator"))
				Expect(sm.Namespace).To(Equal("runai"))
				Expect(sm.Spec.Selector.MatchLabels["app"]).To(Equal("activator"))
			})

			It("webhook resource should be enabled and return a service monitor", func() {
				resource := NewServiceMonitorResource("knative", k8s, runaiConfig, "webhook")
				Expect(resource.Enabled()).To(BeTrue())

				obj, err := resource.Object(ctx)
				Expect(err).ToNot(HaveOccurred())

				sm := obj.(*monitoringv1.ServiceMonitor)
				Expect(sm.Name).To(Equal("knative-serving-webhook"))
				Expect(sm.Namespace).To(Equal("runai"))
				Expect(sm.Spec.Selector.MatchLabels["app"]).To(Equal("webhook"))
			})

			It("status should return error when service monitor doesn't exist", func() {
				resource := NewServiceMonitorResource("knative", k8s, runaiConfig, "controller")
				err := resource.Status(ctx)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("is missing"))
			})
		})

		Context("when CreateServiceMonitors is disabled", func() {
			BeforeEach(func() {
				runaiConfig.Spec.Knative.CreateServiceMonitors = ptr.To(false)
			})

			It("controller resource should be disabled", func() {
				resource := NewServiceMonitorResource("knative", k8s, runaiConfig, "controller")
				Expect(resource.Enabled()).To(BeFalse())
			})

			It("autoscaler resource should be disabled", func() {
				resource := NewServiceMonitorResource("knative", k8s, runaiConfig, "autoscaler")
				Expect(resource.Enabled()).To(BeFalse())
			})

			It("activator resource should be disabled", func() {
				resource := NewServiceMonitorResource("knative", k8s, runaiConfig, "activator")
				Expect(resource.Enabled()).To(BeFalse())
			})

			It("webhook resource should be disabled", func() {
				resource := NewServiceMonitorResource("knative", k8s, runaiConfig, "webhook")
				Expect(resource.Enabled()).To(BeFalse())
			})

			It("should still return object even when disabled", func() {
				resource := NewServiceMonitorResource("knative", k8s, runaiConfig, "controller")
				obj, err := resource.Object(ctx)
				Expect(err).ToNot(HaveOccurred())
				Expect(obj).ToNot(BeNil())

				sm := obj.(*monitoringv1.ServiceMonitor)
				Expect(sm.Name).To(Equal("knative-serving-controller"))
			})
		})
	})
})
