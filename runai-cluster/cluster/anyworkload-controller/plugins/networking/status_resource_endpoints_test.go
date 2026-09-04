/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"

	endpoints "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func endpointAuthProxyDeployment(namespace, name string, available, total int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Status: appsv1.DeploymentStatus{
			Replicas: total, AvailableReplicas: available,
		},
	}
}

func endpointIngress(namespace, name string) *networkingv1.Ingress {
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{{Host: name + ".example.com"}},
		},
	}
}

func findEndpointsCondition(status *dromosv1alpha1.PluginStatus, conditionType string) *metav1.Condition {
	for i := range status.Conditions {
		if status.Conditions[i].Type == conditionType {
			return &status.Conditions[i]
		}
	}
	return nil
}

var _ = Describe("Status (endpoints path)", func() {
	var (
		ctx       context.Context
		plugin    *Plugin
		k8sClient client.Client
		status    *dromosv1alpha1.PluginStatus
	)

	BeforeEach(func() {
		ctx = context.Background()
		k8sClient = fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
		plugin = NewPlugin(k8sClient)
		status = newPluginStatus()
	})

	It("reports Failed when annotation has no endpoints (invalid input)", func() {
		runner := runnerFromEndpoints(&endpoints.Config{Endpoints: nil})
		Expect(plugin.Status(ctx, runner, k8sClient, status, nil)).To(Succeed())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseFailed))
	})

	It("reports Ready for a single endpoint when its auth proxy and ingress are healthy", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{{
				Services: []endpoints.Service{{
					Create: newTestService("jupyter", 8888),
					Auth:   &endpoints.AuthConfig{AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers},
				}},
				ExternalAccess: &endpoints.ExternalAccess{
					Ingress: newTestIngress("jupyter-ing", "jupyter.example.com", "/"),
				},
			}},
		})

		Expect(k8sClient.Create(ctx, endpointAuthProxyDeployment(testNamespace, "jupyter-auth-proxy", 1, 1))).To(Succeed())
		svc := newTestService("jupyter", 8888)
		svc.Namespace = testNamespace
		Expect(k8sClient.Create(ctx, svc)).To(Succeed())
		Expect(k8sClient.Create(ctx, endpointIngress(testNamespace, "jupyter-ing"))).To(Succeed())

		Expect(plugin.Status(ctx, runner, k8sClient, status, nil)).To(Succeed())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))

		authCond := findEndpointsCondition(status, AuthProxyReadyConditionName)
		Expect(authCond).NotTo(BeNil())
		Expect(authCond.Status).To(Equal(metav1.ConditionTrue))

		accessCond := findEndpointsCondition(status, ExternalAccessReadyConditionName)
		Expect(accessCond).NotTo(BeNil())
		Expect(accessCond.Status).To(Equal(metav1.ConditionTrue))

		details := parsedNetworkingDetails(status)
		Expect(details.Endpoints).To(HaveLen(1))
		Expect(details.Endpoints[0].Phase).To(Equal(endpoints.PhaseReady))
		Expect(details.Endpoints[0].Services).To(HaveLen(1))
		service := details.Endpoints[0].Services[0]
		Expect(service.Name).To(Equal("jupyter"))
		Expect(service.Phase).To(Equal(endpoints.PhaseReady))
		Expect(service.AuthProxyReady).To(HaveValue(Equal(true)))
		Expect(service.URLs).To(ConsistOf(endpoints.ServiceURLStatus{
			URL:        "https://jupyter-ing.example.com/",
			SourceKind: endpoints.ExternalAccessKindIngress,
			SourceName: "jupyter-ing",
		}))
		Expect(details.Endpoints[0].ExternalAccess.Resources).To(HaveLen(1))
		Expect(details.Endpoints[0].ExternalAccess.Resources[0].Kind).To(Equal(endpoints.ExternalAccessKindIngress))
	})

	It("reports Creating for a single endpoint when ingress is missing", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{{
				Services: []endpoints.Service{{
					Create: newTestService("jupyter", 8888),
					Auth:   &endpoints.AuthConfig{AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers},
				}},
				ExternalAccess: &endpoints.ExternalAccess{
					Ingress: newTestIngress("jupyter-ing", "jupyter.example.com", "/"),
				},
			}},
		})
		Expect(k8sClient.Create(ctx, endpointAuthProxyDeployment(testNamespace, "jupyter-auth-proxy", 1, 1))).To(Succeed())

		Expect(plugin.Status(ctx, runner, k8sClient, status, nil)).To(Succeed())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseCreating))

		accessCond := findEndpointsCondition(status, ExternalAccessReadyConditionName)
		Expect(accessCond).NotTo(BeNil())
		Expect(accessCond.Status).To(Equal(metav1.ConditionFalse))
	})

	It("reports Ready for a single endpoint when HTTPRoute acceptance status is not reported", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{{
				Services: []endpoints.Service{{
					Create: newTestService("jupyter", 8888),
					Auth:   &endpoints.AuthConfig{AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers},
				}},
				ExternalAccess: &endpoints.ExternalAccess{
					HTTPRoute: statusHTTPRoute("", "jupyter-http", "jupyter.example.com", "jupyter", 8888, nil),
				},
			}},
		})
		Expect(k8sClient.Create(ctx, endpointAuthProxyDeployment(testNamespace, "jupyter-auth-proxy", 1, 1))).To(Succeed())
		svc := newTestService("jupyter", 8888)
		svc.Namespace = testNamespace
		Expect(k8sClient.Create(ctx, svc)).To(Succeed())
		Expect(k8sClient.Create(ctx, statusHTTPRoute(testNamespace, "jupyter-http", "jupyter.example.com", "jupyter", 8888, nil))).To(Succeed())

		Expect(plugin.Status(ctx, runner, k8sClient, status, nil)).To(Succeed())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))

		accessCond := findEndpointsCondition(status, ExternalAccessReadyConditionName)
		Expect(accessCond).NotTo(BeNil())
		Expect(accessCond.Status).To(Equal(metav1.ConditionTrue))
		Expect(accessCond.Reason).To(Equal(ExternalAccessReadyConditionName))

		details := parsedNetworkingDetails(status)
		Expect(details.Endpoints[0].Phase).To(Equal(endpoints.PhaseReady))
		resource := details.Endpoints[0].ExternalAccess.Resources[0]
		Expect(resource.Kind).To(Equal(endpoints.ExternalAccessKindHTTPRoute))
		Expect(resource.Phase).To(Equal(endpoints.PhaseReady))
		Expect(resource.Reason).To(Equal(ExternalAccessReadyConditionName))
		Expect(resource.Accepted).To(BeNil())
		Expect(details.Endpoints[0].Services[0].URLs).To(ConsistOf(endpoints.ServiceURLStatus{
			URL:        "https://jupyter.example.com/",
			SourceKind: endpoints.ExternalAccessKindHTTPRoute,
			SourceName: "jupyter-http",
		}))
	})

	It("preserves primary Ingress readiness while reporting HTTPRoute details", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{{
				Services: []endpoints.Service{{
					Create: newTestService("jupyter", 8888),
				}},
				ExternalAccess: &endpoints.ExternalAccess{
					Ingress:   newTestIngress("jupyter-ing", "jupyter.example.com", "/"),
					HTTPRoute: statusHTTPRoute("", "jupyter-http", "jupyter.gateway.example.com", "jupyter", 8888, nil),
				},
			}},
		})
		svc := newTestService("jupyter", 8888)
		svc.Namespace = testNamespace
		Expect(k8sClient.Create(ctx, svc)).To(Succeed())
		ingress := newTestIngress("jupyter-ing", "jupyter.example.com", "/")
		ingress.Namespace = testNamespace
		Expect(k8sClient.Create(ctx, ingress)).To(Succeed())
		Expect(k8sClient.Create(ctx, statusHTTPRoute(testNamespace, "jupyter-http", "jupyter.gateway.example.com", "jupyter", 8888, nil))).To(Succeed())

		Expect(plugin.Status(ctx, runner, k8sClient, status, nil)).To(Succeed())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))

		accessCond := findEndpointsCondition(status, ExternalAccessReadyConditionName)
		Expect(accessCond).NotTo(BeNil())
		Expect(accessCond.Status).To(Equal(metav1.ConditionTrue))
		Expect(accessCond.Reason).To(Equal(ExternalAccessReadyConditionName))

		details := parsedNetworkingDetails(status)
		Expect(details.Endpoints[0].Phase).To(Equal(endpoints.PhaseReady))
		Expect(details.Endpoints[0].ExternalAccess.Phase).To(Equal(endpoints.PhaseReady))
		Expect(details.Endpoints[0].ExternalAccess.Resources).To(HaveLen(2))
		Expect(details.Endpoints[0].ExternalAccess.Resources[0].Phase).To(Equal(endpoints.PhaseReady))
		Expect(details.Endpoints[0].ExternalAccess.Resources[1].Phase).To(Equal(endpoints.PhaseReady))
		Expect(details.Endpoints[0].Services[0].URLs).To(ConsistOf(
			endpoints.ServiceURLStatus{
				URL:        "https://jupyter.example.com/",
				SourceKind: endpoints.ExternalAccessKindIngress,
				SourceName: "jupyter-ing",
			},
			endpoints.ServiceURLStatus{
				URL:        "https://jupyter.gateway.example.com/",
				SourceKind: endpoints.ExternalAccessKindHTTPRoute,
				SourceName: "jupyter-http",
			},
		))
	})

	It("folds two endpoints to worst-of phase (one ready, one initializing → Initializing)", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{
				{
					Services: []endpoints.Service{{
						Create: newTestService("jupyter", 8888),
						Auth:   &endpoints.AuthConfig{AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers},
					}},
				},
				{
					Services: []endpoints.Service{{
						Create: newTestService("tensorboard", 6006),
						Auth:   &endpoints.AuthConfig{AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers},
					}},
				},
			},
		})

		Expect(k8sClient.Create(ctx, endpointAuthProxyDeployment(testNamespace, "jupyter-auth-proxy", 1, 1))).To(Succeed())
		Expect(k8sClient.Create(ctx, endpointAuthProxyDeployment(testNamespace, "tensorboard-auth-proxy", 0, 1))).To(Succeed())

		Expect(plugin.Status(ctx, runner, k8sClient, status, nil)).To(Succeed())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseInitializing))

		authCond := findEndpointsCondition(status, AuthProxyReadyConditionName)
		Expect(authCond).NotTo(BeNil())
		Expect(authCond.Status).To(Equal(metav1.ConditionFalse))
		Expect(authCond.Message).To(ContainSubstring("endpoint[0]"))
		Expect(authCond.Message).To(ContainSubstring("endpoint[1]"))
	})

	It("reports Disabled when an endpoint has neither auth nor external access (internal-only)", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{{
				Services: []endpoints.Service{{Create: newTestService("jupyter", 8888)}},
			}},
		})

		Expect(plugin.Status(ctx, runner, k8sClient, status, nil)).To(Succeed())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseDisabled))
	})
})
