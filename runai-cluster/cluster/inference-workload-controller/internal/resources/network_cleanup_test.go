/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources_test

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/inference-workload-controller/internal/resources"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-common/config"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var _ = Describe("CleanupDisabledNetworkingResources", func() {
	var (
		scheme     *runtime.Scheme
		fakeClient client.Client
		ctx        context.Context
		namespace  string
		labels     map[string]string
	)

	const workloadName = "test-workload"

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		Expect(netv1.AddToScheme(scheme)).To(Succeed())
		Expect(routev1.Install(scheme)).To(Succeed())
		Expect(gatewayv1.Install(scheme)).To(Succeed())

		ctx = context.Background()
		namespace = "test-namespace"
		labels = map[string]string{constants.LabelRelease: workloadName}
	})

	AfterEach(func() {
		_ = os.Unsetenv(config.EnvVarIngressEnabled)
		_ = os.Unsetenv(config.EnvVarGatewayAPIEnabled)
	})

	Context("switching from Ingress to Gateway-only mode (non-OpenShift)", func() {
		It("should delete orphaned Ingresses when Ingress mode is disabled", func() {
			ingress := &netv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ingress-0",
					Namespace: namespace,
					Labels:    labels,
				},
			}
			fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(ingress).Build()

			_ = os.Setenv(config.EnvVarIngressEnabled, "false")
			_ = os.Setenv(config.EnvVarGatewayAPIEnabled, "true")

			err := resources.CleanupDisabledNetworkingResources(ctx, fakeClient, namespace, labels, false)
			Expect(err).ToNot(HaveOccurred())

			list := &netv1.IngressList{}
			Expect(fakeClient.List(ctx, list, client.InNamespace(namespace))).To(Succeed())
			Expect(list.Items).To(BeEmpty())
		})
	})

	Context("switching from Ingress to Gateway-only mode (OpenShift)", func() {
		It("should delete orphaned Routes when Ingress mode is disabled", func() {
			route := &routev1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-route-0",
					Namespace: namespace,
					Labels:    labels,
				},
			}
			fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(route).Build()

			_ = os.Setenv(config.EnvVarIngressEnabled, "false")
			_ = os.Setenv(config.EnvVarGatewayAPIEnabled, "true")

			err := resources.CleanupDisabledNetworkingResources(ctx, fakeClient, namespace, labels, true)
			Expect(err).ToNot(HaveOccurred())

			list := &routev1.RouteList{}
			Expect(fakeClient.List(ctx, list, client.InNamespace(namespace))).To(Succeed())
			Expect(list.Items).To(BeEmpty())
		})
	})

	Context("switching from Gateway to Ingress-only mode", func() {
		It("should delete orphaned HTTPRoutes when Gateway mode is disabled", func() {
			httpRoute := &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-httproute-0",
					Namespace: namespace,
					Labels:    labels,
				},
			}
			fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(httpRoute).Build()

			_ = os.Setenv(config.EnvVarIngressEnabled, "true")
			_ = os.Setenv(config.EnvVarGatewayAPIEnabled, "false")

			err := resources.CleanupDisabledNetworkingResources(ctx, fakeClient, namespace, labels, false)
			Expect(err).ToNot(HaveOccurred())

			list := &gatewayv1.HTTPRouteList{}
			Expect(fakeClient.List(ctx, list, client.InNamespace(namespace))).To(Succeed())
			Expect(list.Items).To(BeEmpty())
		})
	})

	Context("dual mode - both enabled", func() {
		It("should not delete any resources when both modes are enabled", func() {
			ingress := &netv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-ingress-0",
					Namespace: namespace,
					Labels:    labels,
				},
			}
			httpRoute := &gatewayv1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-httproute-0",
					Namespace: namespace,
					Labels:    labels,
				},
			}
			fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(ingress, httpRoute).Build()

			_ = os.Setenv(config.EnvVarIngressEnabled, "true")
			_ = os.Setenv(config.EnvVarGatewayAPIEnabled, "true")

			err := resources.CleanupDisabledNetworkingResources(ctx, fakeClient, namespace, labels, false)
			Expect(err).ToNot(HaveOccurred())

			ingressList := &netv1.IngressList{}
			Expect(fakeClient.List(ctx, ingressList, client.InNamespace(namespace))).To(Succeed())
			Expect(ingressList.Items).To(HaveLen(1))

			httpRouteList := &gatewayv1.HTTPRouteList{}
			Expect(fakeClient.List(ctx, httpRouteList, client.InNamespace(namespace))).To(Succeed())
			Expect(httpRouteList.Items).To(HaveLen(1))
		})
	})

	Context("label isolation", func() {
		It("should not delete resources belonging to a different workload", func() {
			ownIngress := &netv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "own-ingress",
					Namespace: namespace,
					Labels:    labels,
				},
			}
			otherIngress := &netv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "other-ingress",
					Namespace: namespace,
					Labels:    map[string]string{constants.LabelRelease: "other-workload"},
				},
			}
			fakeClient = fake.NewClientBuilder().WithScheme(scheme).WithObjects(ownIngress, otherIngress).Build()

			_ = os.Setenv(config.EnvVarIngressEnabled, "false")
			_ = os.Setenv(config.EnvVarGatewayAPIEnabled, "true")

			err := resources.CleanupDisabledNetworkingResources(ctx, fakeClient, namespace, labels, false)
			Expect(err).ToNot(HaveOccurred())

			list := &netv1.IngressList{}
			Expect(fakeClient.List(ctx, list, client.InNamespace(namespace))).To(Succeed())
			Expect(list.Items).To(HaveLen(1))
			Expect(list.Items[0].Name).To(Equal("other-ingress"))
		})
	})
})
