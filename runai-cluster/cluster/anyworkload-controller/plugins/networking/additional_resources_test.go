/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openshiftv1 "github.com/openshift/api/route/v1"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/testutil"
	workloadbuilders "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func newTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = dromosv1alpha1.AddToScheme(s)
	_ = corev1.AddToScheme(s)
	_ = appsv1.AddToScheme(s)
	_ = networkingv1.AddToScheme(s)
	_ = openshiftv1.AddToScheme(s)
	_ = gatewayv1.Install(s)
	return s
}

func newTestService(name string, port int32) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: corev1.ServiceSpec{
			Ports:    []corev1.ServicePort{{Port: port}},
			Selector: map[string]string{"app": "test"},
		},
	}
}

func newTestIngress(name, host, path string) *networkingv1.Ingress {
	pathType := networkingv1.PathTypePrefix
	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{{
				Host: host,
				IngressRuleValue: networkingv1.IngressRuleValue{
					HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{{
							Path:     path,
							PathType: &pathType,
						}},
					},
				},
			}},
		},
	}
}

func newTestRoute(name, host, path string) *openshiftv1.Route {
	return &openshiftv1.Route{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: openshiftv1.RouteSpec{
			Host: host,
			Path: path,
		},
	}
}

var _ = Describe("HandleAdditionalResources", func() {
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
	})

	Context("service creation", func() {
		It("creates service when services[0].create is specified", func() {
			k8sClient := testutil.WithSSASupport(fake.NewClientBuilder().WithScheme(newTestScheme()).Build())
			plugin := NewPlugin(k8sClient)
			config := &dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: newTestService("my-svc", 8080),
				}},
			}
			runner := newTestRunner(testWorkloadName, testNamespace, config)

			err := plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, reconcile.New())
			Expect(err).NotTo(HaveOccurred())

			svcList := &corev1.ServiceList{}
			Expect(k8sClient.List(ctx, svcList)).To(Succeed())
			Expect(svcList.Items).To(HaveLen(1))
			Expect(svcList.Items[0].Name).To(Equal("my-svc"))
		})
	})

	Context("auth proxy resources", func() {
		It("creates auth proxy deployment, service and ingress when auth enabled", func() {
			GinkgoT().Setenv(workloadconfig.EnvVarOauth2ProxyResources, `{"requests":{"cpu":"100m"},"limits":{"memory":"128Mi"}}`)
			GinkgoT().Setenv(workloadconfig.EnvVarUsernameLoaderResources, `{"requests":{"cpu":"25m"},"limits":{"memory":"64Mi"}}`)
			k8sClient := testutil.WithSSASupport(fake.NewClientBuilder().WithScheme(newTestScheme()).Build())
			plugin := NewPlugin(k8sClient)

			config := &dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: newTestService("my-svc", 8080),
					Auth: &dromosv1alpha1.NetworkingAuthConfig{
						AuthorizationType: dromosv1alpha1.AuthorizationTypeAuthorizedUsers,
						AuthorizedUsers:   "test@run.ai",
					},
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: newTestIngress("my-ingress", "test.example.com", "/"),
				},
			}
			runner := newTestRunner(testWorkloadName, testNamespace, config)

			err := plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, reconcile.New())
			Expect(err).NotTo(HaveOccurred())

			deployList := &appsv1.DeploymentList{}
			Expect(k8sClient.List(ctx, deployList)).To(Succeed())
			Expect(deployList.Items).To(HaveLen(1))
			Expect(deployList.Items[0].Name).To(Equal(testWorkloadName + "-auth-proxy"))
			Expect(deployList.Items[0].Spec.Template.Spec.Containers[0].Resources.Requests.Cpu().Equal(resource.MustParse("100m"))).To(BeTrue())
			Expect(deployList.Items[0].Spec.Template.Spec.Containers[0].Resources.Limits.Memory().Equal(resource.MustParse("128Mi"))).To(BeTrue())
			initContainers := deployList.Items[0].Spec.Template.Spec.InitContainers
			Expect(initContainers).To(HaveLen(1))
			Expect(initContainers[0].Name).To(Equal(workloadbuilders.UsernameLoaderInitContainerName))
			Expect(initContainers[0].Resources.Requests.Cpu().Equal(resource.MustParse("25m"))).To(BeTrue())
			Expect(initContainers[0].Resources.Limits.Memory().Equal(resource.MustParse("64Mi"))).To(BeTrue())

			svcList := &corev1.ServiceList{}
			Expect(k8sClient.List(ctx, svcList)).To(Succeed())
			Expect(svcList.Items).To(HaveLen(2))

			ingressList := &networkingv1.IngressList{}
			Expect(k8sClient.List(ctx, ingressList)).To(Succeed())
			Expect(ingressList.Items).To(HaveLen(1))
			Expect(ingressList.Items[0].Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name).To(
				Equal(testWorkloadName + "-auth-proxy"))
		})

		It("returns malformed resource configuration errors without panicking", func() {
			GinkgoT().Setenv(workloadconfig.EnvVarOauth2ProxyResources, `{`)
			k8sClient := testutil.WithSSASupport(fake.NewClientBuilder().WithScheme(newTestScheme()).Build())
			plugin := NewPlugin(k8sClient)
			config := &dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: newTestService("my-svc", 8080),
					Auth:   &dromosv1alpha1.NetworkingAuthConfig{},
				}},
			}
			runner := newTestRunner(testWorkloadName, testNamespace, config)

			_, err := plugin.buildAuthProxyDeployment(
				runner, config, "http://localhost:8080", "test.example.com", "/",
			)

			Expect(err).To(MatchError(ContainSubstring("RUNAI_OAUTH2_PROXY_RESOURCES")))
		})

		It("succeeds on second reconciliation when resources already exist", func() {
			k8sClient := testutil.WithSSASupport(fake.NewClientBuilder().WithScheme(newTestScheme()).Build())
			plugin := NewPlugin(k8sClient)

			config := &dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: newTestService("my-svc", 8080),
					Auth: &dromosv1alpha1.NetworkingAuthConfig{
						AuthorizationType: dromosv1alpha1.AuthorizationTypeAuthenticatedUsers,
					},
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: newTestIngress("my-ingress", "test.example.com", "/"),
				},
			}
			runner := newTestRunner(testWorkloadName, testNamespace, config)

			Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, reconcile.New())).To(Succeed())

			Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, reconcile.New())).To(Succeed())

			deployList := &appsv1.DeploymentList{}
			Expect(k8sClient.List(ctx, deployList)).To(Succeed())
			Expect(deployList.Items).To(HaveLen(1))
			Expect(deployList.Items[0].Name).To(Equal(testWorkloadName + "-auth-proxy"))
		})

		It("does not create auth proxy when auth section is absent", func() {
			k8sClient := testutil.WithSSASupport(fake.NewClientBuilder().WithScheme(newTestScheme()).Build())
			plugin := NewPlugin(k8sClient)

			config := &dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: newTestService("my-svc", 8080),
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: newTestIngress("my-ingress", "test.example.com", "/"),
				},
			}
			runner := newTestRunner(testWorkloadName, testNamespace, config)

			err := plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, reconcile.New())
			Expect(err).NotTo(HaveOccurred())

			deployList := &appsv1.DeploymentList{}
			Expect(k8sClient.List(ctx, deployList)).To(Succeed())
			Expect(deployList.Items).To(BeEmpty())

			svcList := &corev1.ServiceList{}
			Expect(k8sClient.List(ctx, svcList)).To(Succeed())
			Expect(svcList.Items).To(HaveLen(1))
			Expect(svcList.Items[0].Name).To(Equal("my-svc"))
		})
	})

	Context("OpenShift Route", func() {
		It("creates auth proxy and route when auth enabled with route", func() {
			k8sClient := testutil.WithSSASupport(fake.NewClientBuilder().WithScheme(newTestScheme()).Build())
			plugin := NewPlugin(k8sClient)

			config := &dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Create: newTestService("my-svc", 8080),
					Auth: &dromosv1alpha1.NetworkingAuthConfig{
						AuthorizationType: dromosv1alpha1.AuthorizationTypeAuthenticatedUsers,
					},
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Route: newTestRoute("my-route", "test.example.com", "/"),
				},
			}
			runner := newTestRunner(testWorkloadName, testNamespace, config)

			err := plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, reconcile.New())
			Expect(err).NotTo(HaveOccurred())

			deployList := &appsv1.DeploymentList{}
			Expect(k8sClient.List(ctx, deployList)).To(Succeed())
			Expect(deployList.Items).To(HaveLen(1))

			svcList := &corev1.ServiceList{}
			Expect(k8sClient.List(ctx, svcList)).To(Succeed())
			Expect(svcList.Items).To(HaveLen(2))

			routeList := &openshiftv1.RouteList{}
			Expect(k8sClient.List(ctx, routeList)).To(Succeed())
			Expect(routeList.Items).To(HaveLen(1))
			Expect(routeList.Items[0].Name).To(Equal("my-route"))
		})
	})

	Context("service reference", func() {
		It("references existing service instead of creating one", func() {
			k8sClient := testutil.WithSSASupport(fake.NewClientBuilder().WithScheme(newTestScheme()).Build())
			plugin := NewPlugin(k8sClient)

			config := &dromosv1alpha1.NetworkingConfig{
				Services: []dromosv1alpha1.NetworkingServiceItem{{
					Ref: &dromosv1alpha1.NetworkingServiceRef{
						Name: "existing-service",
						Port: 8080,
					},
					Auth: &dromosv1alpha1.NetworkingAuthConfig{
						AuthorizationType: dromosv1alpha1.AuthorizationTypeAuthenticatedUsers,
					},
				}},
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: newTestIngress("my-ingress", "test.example.com", "/"),
				},
			}
			runner := newTestRunner(testWorkloadName, testNamespace, config)

			err := plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, reconcile.New())
			Expect(err).NotTo(HaveOccurred())

			svcList := &corev1.ServiceList{}
			Expect(k8sClient.List(ctx, svcList)).To(Succeed())
			Expect(svcList.Items).To(HaveLen(1))
			Expect(svcList.Items[0].Name).To(Equal(testWorkloadName + "-auth-proxy"))

			deployList := &appsv1.DeploymentList{}
			Expect(k8sClient.List(ctx, deployList)).To(Succeed())
			Expect(deployList.Items).To(HaveLen(1))
		})
	})

	Context("ingress only without services", func() {
		It("creates ingress without backend override when no services", func() {
			k8sClient := testutil.WithSSASupport(fake.NewClientBuilder().WithScheme(newTestScheme()).Build())
			plugin := NewPlugin(k8sClient)

			pathType := networkingv1.PathTypePrefix
			config := &dromosv1alpha1.NetworkingConfig{
				// No services - just creating Ingress
				ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
					Ingress: &networkingv1.Ingress{
						ObjectMeta: metav1.ObjectMeta{
							Name: "standalone-ingress",
						},
						Spec: networkingv1.IngressSpec{
							Rules: []networkingv1.IngressRule{{
								Host: "test.example.com",
								IngressRuleValue: networkingv1.IngressRuleValue{
									HTTP: &networkingv1.HTTPIngressRuleValue{
										Paths: []networkingv1.HTTPIngressPath{{
											Path:     "/",
											PathType: &pathType,
											Backend: networkingv1.IngressBackend{
												Service: &networkingv1.IngressServiceBackend{
													Name: "external-service",
													Port: networkingv1.ServiceBackendPort{Number: 9000},
												},
											},
										}},
									},
								},
							}},
						},
					},
				},
			}
			runner := newTestRunner(testWorkloadName, testNamespace, config)

			err := plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, reconcile.New())
			Expect(err).NotTo(HaveOccurred())

			// No services should be created
			svcList := &corev1.ServiceList{}
			Expect(k8sClient.List(ctx, svcList)).To(Succeed())
			Expect(svcList.Items).To(BeEmpty())

			// No deployments should be created
			deployList := &appsv1.DeploymentList{}
			Expect(k8sClient.List(ctx, deployList)).To(Succeed())
			Expect(deployList.Items).To(BeEmpty())

			// Ingress should be created with original backend preserved
			ingressList := &networkingv1.IngressList{}
			Expect(k8sClient.List(ctx, ingressList)).To(Succeed())
			Expect(ingressList.Items).To(HaveLen(1))
			Expect(ingressList.Items[0].Name).To(Equal("standalone-ingress"))
			Expect(ingressList.Items[0].Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name).To(Equal("external-service"))
			Expect(ingressList.Items[0].Spec.Rules[0].HTTP.Paths[0].Backend.Service.Port.Number).To(Equal(int32(9000)))
		})
	})

	Context("skip scenarios", func() {
		It("skips when no networking config", func() {
			k8sClient := fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
			plugin := NewPlugin(k8sClient)
			runner := newTestRunner(testWorkloadName, testNamespace, nil)

			err := plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, reconcile.New())
			Expect(err).NotTo(HaveOccurred())

			svcList := &corev1.ServiceList{}
			Expect(k8sClient.List(ctx, svcList)).To(Succeed())
			Expect(svcList.Items).To(BeEmpty())
		})
	})
})
