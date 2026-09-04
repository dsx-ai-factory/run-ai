/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	endpoints "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/testutil"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type patchFailingClient struct {
	client.Client
	name string
}

func (c *patchFailingClient) Patch(
	ctx context.Context,
	obj client.Object,
	patch client.Patch,
	opts ...client.PatchOption,
) error {
	if obj.GetName() == c.name {
		return fmt.Errorf("injected apply failure for %s", c.name)
	}
	return c.Client.Patch(ctx, obj, patch, opts...)
}

// Bypasses endpoints.Set so callers can produce structurally invalid payloads to exercise downstream validators.
func runnerFromEndpoints(cfg *endpoints.Config) *dromosv1alpha1.WorkloadRunner {
	cfg.SchemaVersion = endpoints.SchemaVersionV1
	raw, err := json.Marshal(cfg)
	Expect(err).NotTo(HaveOccurred())
	return &dromosv1alpha1.WorkloadRunner{
		ObjectMeta: metav1.ObjectMeta{
			Name:        testWorkloadName,
			Namespace:   testNamespace,
			UID:         "test-uid-12345",
			Annotations: map[string]string{endpoints.Annotation: string(raw)},
		},
	}
}

func metricsService(name, engineLabel string) endpoints.Service {
	labels := map[string]string{
		constants.LabelWorkloadMetricsService: "true",
		"workloadKind":                        constants.KindInferenceWorkload,
	}
	if engineLabel != "" {
		labels[engineLabel] = "true"
	}
	return endpoints.Service{Create: &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: map[string]string{"workloadName": testWorkloadName},
			Ports: []corev1.ServicePort{{
				Name: "metrics", Port: 8080, TargetPort: intstr.FromInt32(8080), Protocol: corev1.ProtocolTCP,
			}},
		},
	}}
}

func newTestHTTPRoute(name, host, serviceName string, port int32) *gatewayv1.HTTPRoute {
	return newTestHTTPRouteWithPath(name, host, "", serviceName, port)
}

func newTestHTTPRouteWithPath(name, host, path, serviceName string, port int32) *gatewayv1.HTTPRoute {
	routePort := gatewayv1.PortNumber(port)
	matches := []gatewayv1.HTTPRouteMatch{}
	if path != "" {
		pathType := gatewayv1.PathMatchPathPrefix
		pathValue := path
		matches = []gatewayv1.HTTPRouteMatch{{
			Path: &gatewayv1.HTTPPathMatch{
				Type:  &pathType,
				Value: &pathValue,
			},
		}}
	}
	return &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: gatewayv1.HTTPRouteSpec{
			Hostnames: []gatewayv1.Hostname{gatewayv1.Hostname(host)},
			Rules: []gatewayv1.HTTPRouteRule{{
				Matches: matches,
				BackendRefs: []gatewayv1.HTTPBackendRef{{
					BackendRef: gatewayv1.BackendRef{
						BackendObjectReference: gatewayv1.BackendObjectReference{
							Name: gatewayv1.ObjectName(serviceName),
							Port: &routePort,
						},
					},
				}},
			}},
		},
	}
}

func authProxyArgs(k8sClient client.Client) []string {
	deployList := &appsv1.DeploymentList{}
	Expect(k8sClient.List(context.Background(), deployList)).To(Succeed())
	Expect(deployList.Items).To(HaveLen(1))
	containers := deployList.Items[0].Spec.Template.Spec.Containers
	Expect(containers).To(HaveLen(1))
	return containers[0].Args
}

var _ = Describe("HandleAdditionalResources (endpoints path)", func() {
	var (
		ctx       context.Context
		k8sClient client.Client
		plugin    *Plugin
		state     *reconcile.State
	)

	BeforeEach(func() {
		ctx = context.Background()
		k8sClient = testutil.WithSSASupport(fake.NewClientBuilder().WithScheme(newTestScheme()).Build())
		plugin = NewPlugin(k8sClient)
		state = reconcile.New()
	})

	It("reconciles a single internal-only endpoint (Service only)", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{{
				Services: []endpoints.Service{{
					Create: newTestService("jupyter", 8888),
				}},
			}},
		})

		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())

		svcList := &corev1.ServiceList{}
		Expect(k8sClient.List(ctx, svcList)).To(Succeed())
		Expect(svcList.Items).To(HaveLen(1))
		Expect(svcList.Items[0].Name).To(Equal("jupyter"))

		deployList := &appsv1.DeploymentList{}
		Expect(k8sClient.List(ctx, deployList)).To(Succeed())
		Expect(deployList.Items).To(BeEmpty())

		ingressList := &networkingv1.IngressList{}
		Expect(k8sClient.List(ctx, ingressList)).To(Succeed())
		Expect(ingressList.Items).To(BeEmpty())
	})

	It("reconciles a single endpoint with auth proxy + ingress; ingress backend is rewritten to the auth proxy", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{{
				Services: []endpoints.Service{{
					Create: newTestService("jupyter", 8888),
					Auth: &endpoints.AuthConfig{
						AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers,
					},
				}},
				ExternalAccess: &endpoints.ExternalAccess{
					Ingress: newTestIngress("jupyter-ing", "jupyter.example.com", "/"),
				},
			}},
		})

		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())

		expectedAuthProxyName := "jupyter-auth-proxy"

		deployList := &appsv1.DeploymentList{}
		Expect(k8sClient.List(ctx, deployList)).To(Succeed())
		Expect(deployList.Items).To(HaveLen(1))
		Expect(deployList.Items[0].Name).To(Equal(expectedAuthProxyName))

		svcList := &corev1.ServiceList{}
		Expect(k8sClient.List(ctx, svcList)).To(Succeed())
		names := []string{}
		for _, s := range svcList.Items {
			names = append(names, s.Name)
		}
		Expect(names).To(ConsistOf("jupyter", expectedAuthProxyName))

		ingressList := &networkingv1.IngressList{}
		Expect(k8sClient.List(ctx, ingressList)).To(Succeed())
		Expect(ingressList.Items).To(HaveLen(1))
		backend := ingressList.Items[0].Spec.Rules[0].HTTP.Paths[0].Backend.Service
		Expect(backend.Name).To(Equal(expectedAuthProxyName))
	})

	It("reconciles the authenticated inference endpoint whose generated Service name is at the Kubernetes limit", func() {
		serviceName := "infw-workload-20510b80092a-b54b0b-url-authenticated--e2e17e943f"
		expectedAuthProxyName := getEndpointAuthProxyName(serviceName)
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{{
				Services: []endpoints.Service{{
					Create: newTestService(serviceName, 8888),
					Auth: &endpoints.AuthConfig{
						AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers,
					},
				}},
				ExternalAccess: &endpoints.ExternalAccess{
					Ingress: newTestIngress(serviceName, "inference.example.com", "/"),
				},
			}},
		})

		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())
		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())

		serviceList := &corev1.ServiceList{}
		Expect(k8sClient.List(ctx, serviceList)).To(Succeed())
		Expect(serviceList.Items).To(HaveLen(2))
		Expect([]string{serviceList.Items[0].Name, serviceList.Items[1].Name}).To(
			ConsistOf(serviceName, expectedAuthProxyName),
		)

		deploymentList := &appsv1.DeploymentList{}
		Expect(k8sClient.List(ctx, deploymentList)).To(Succeed())
		Expect(deploymentList.Items).To(HaveLen(1))
		Expect(deploymentList.Items[0].Name).To(Equal(expectedAuthProxyName))

		ingressList := &networkingv1.IngressList{}
		Expect(k8sClient.List(ctx, ingressList)).To(Succeed())
		Expect(ingressList.Items).To(HaveLen(1))
		Expect(ingressList.Items[0].Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name).To(
			Equal(expectedAuthProxyName),
		)
	})

	It("reconciles Ingress and HTTPRoute together for one endpoint", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{{
				Services: []endpoints.Service{{
					Create: newTestService("jupyter", 8888),
				}},
				ExternalAccess: &endpoints.ExternalAccess{
					Ingress:   newTestIngress("jupyter-ing", "jupyter.example.com", "/"),
					HTTPRoute: newTestHTTPRoute("jupyter-http", "jupyter.gateway.example.com", "jupyter", 8888),
				},
			}},
		})

		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())

		ingressList := &networkingv1.IngressList{}
		Expect(k8sClient.List(ctx, ingressList)).To(Succeed())
		Expect(ingressList.Items).To(HaveLen(1))
		Expect(ingressList.Items[0].Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name).To(Equal("jupyter"))

		httpRouteList := &gatewayv1.HTTPRouteList{}
		Expect(k8sClient.List(ctx, httpRouteList)).To(Succeed())
		Expect(httpRouteList.Items).To(HaveLen(1))
		Expect(string(httpRouteList.Items[0].Spec.Rules[0].BackendRefs[0].Name)).To(Equal("jupyter"))
	})

	It("uses HTTPRoute host/path for auth proxy when Ingress and HTTPRoute are both present", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{{
				Services: []endpoints.Service{{
					Create: newTestService("jupyter", 8888),
					Auth: &endpoints.AuthConfig{
						AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers,
					},
				}},
				ExternalAccess: &endpoints.ExternalAccess{
					Ingress:   newTestIngress("jupyter-ing", "jupyter.example.com", "/ingress"),
					HTTPRoute: newTestHTTPRouteWithPath("jupyter-http", "jupyter.gateway.example.com", "/gateway", "jupyter", 8888),
				},
			}},
		})

		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())

		args := authProxyArgs(k8sClient)
		Expect(args).To(ContainElement("--redirect-url=https://jupyter.gateway.example.com/gateway/oauth2/callback"))
		Expect(args).To(ContainElement("--whitelist-domain=jupyter.gateway.example.com"))
		Expect(args).To(ContainElement("--cookie-domain=jupyter.gateway.example.com"))
		Expect(args).NotTo(ContainElement("--redirect-url=https://jupyter.example.com/ingress/oauth2/callback"))

		ingressList := &networkingv1.IngressList{}
		Expect(k8sClient.List(ctx, ingressList)).To(Succeed())
		Expect(ingressList.Items).To(HaveLen(1))
		Expect(ingressList.Items[0].Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name).To(Equal("jupyter-auth-proxy"))

		httpRouteList := &gatewayv1.HTTPRouteList{}
		Expect(k8sClient.List(ctx, httpRouteList)).To(Succeed())
		Expect(httpRouteList.Items).To(HaveLen(1))
		Expect(string(httpRouteList.Items[0].Spec.Rules[0].BackendRefs[0].Name)).To(Equal("jupyter-auth-proxy"))
	})

	It("uses HTTPRoute host/path for auth proxy when Route and HTTPRoute are both present", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{{
				Services: []endpoints.Service{{
					Create: newTestService("jupyter", 8888),
					Auth: &endpoints.AuthConfig{
						AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers,
					},
				}},
				ExternalAccess: &endpoints.ExternalAccess{
					Route:     newTestRoute("jupyter-route", "jupyter.apps.example.com", "/route"),
					HTTPRoute: newTestHTTPRouteWithPath("jupyter-http", "jupyter.gateway.example.com", "/gateway", "jupyter", 8888),
				},
			}},
		})

		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())

		args := authProxyArgs(k8sClient)
		Expect(args).To(ContainElement("--redirect-url=https://jupyter.gateway.example.com/gateway/oauth2/callback"))
		Expect(args).To(ContainElement("--whitelist-domain=jupyter.gateway.example.com"))
		Expect(args).To(ContainElement("--cookie-domain=jupyter.gateway.example.com"))
		Expect(args).NotTo(ContainElement("--redirect-url=https://jupyter.apps.example.com/route/oauth2/callback"))

		httpRouteList := &gatewayv1.HTTPRouteList{}
		Expect(k8sClient.List(ctx, httpRouteList)).To(Succeed())
		Expect(httpRouteList.Items).To(HaveLen(1))
		Expect(string(httpRouteList.Items[0].Spec.Rules[0].BackendRefs[0].Name)).To(Equal("jupyter-auth-proxy"))
	})

	It("reconciles external access without services and preserves the declared backend", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{{
				ExternalAccess: &endpoints.ExternalAccess{
					Ingress: minimalEndpointIngress("external-service", 9000),
				},
			}},
		})

		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())

		svcList := &corev1.ServiceList{}
		Expect(k8sClient.List(ctx, svcList)).To(Succeed())
		Expect(svcList.Items).To(BeEmpty())

		deployList := &appsv1.DeploymentList{}
		Expect(k8sClient.List(ctx, deployList)).To(Succeed())
		Expect(deployList.Items).To(BeEmpty())

		ingressList := &networkingv1.IngressList{}
		Expect(k8sClient.List(ctx, ingressList)).To(Succeed())
		Expect(ingressList.Items).To(HaveLen(1))
		backend := ingressList.Items[0].Spec.Rules[0].HTTP.Paths[0].Backend.Service
		Expect(backend.Name).To(Equal("external-service"))
		Expect(backend.Port.Number).To(Equal(int32(9000)))
	})

	It("reconciles two endpoints independently, each with its own Service/auth-proxy/ingress", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{
				{
					Services: []endpoints.Service{{
						Create: newTestService("jupyter", 8888),
						Auth: &endpoints.AuthConfig{
							AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers,
						},
					}},
					ExternalAccess: &endpoints.ExternalAccess{
						Ingress: newTestIngress("jupyter-ing", "jupyter.example.com", "/"),
					},
				},
				{
					Services: []endpoints.Service{{
						Create: newTestService("tensorboard", 6006),
					}},
				},
			},
		})

		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())

		svcList := &corev1.ServiceList{}
		Expect(k8sClient.List(ctx, svcList)).To(Succeed())
		svcNames := []string{}
		for _, s := range svcList.Items {
			svcNames = append(svcNames, s.Name)
		}
		Expect(svcNames).To(ConsistOf("jupyter", "tensorboard", "jupyter-auth-proxy"))

		deployList := &appsv1.DeploymentList{}
		Expect(k8sClient.List(ctx, deployList)).To(Succeed())
		Expect(deployList.Items).To(HaveLen(1))
		Expect(deployList.Items[0].Name).To(Equal("jupyter-auth-proxy"))

		ingressList := &networkingv1.IngressList{}
		Expect(k8sClient.List(ctx, ingressList)).To(Succeed())
		Expect(ingressList.Items).To(HaveLen(1))
		Expect(ingressList.Items[0].Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name).To(
			Equal("jupyter-auth-proxy"))
	})

	It("skips endpoints with no services and reconciles the rest", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{
				{Services: []endpoints.Service{}},
				{
					Services: []endpoints.Service{{
						Create: newTestService("tensorboard", 6006),
					}},
				},
			},
		})

		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())

		svcList := &corev1.ServiceList{}
		Expect(k8sClient.List(ctx, svcList)).To(Succeed())
		Expect(svcList.Items).To(HaveLen(1))
		Expect(svcList.Items[0].Name).To(Equal("tensorboard"))
	})

	It("records endpoint reconciliation failures as critical and continues with siblings", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{
				{Services: []endpoints.Service{{}}},
				{Services: []endpoints.Service{{Create: newTestService("tensorboard", 6006)}}},
			},
		})

		err := plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)

		Expect(err).To(MatchError(ContainSubstring(ErrServiceMissingCreateAndRef.Error())))
		issues := state.PluginIssues(plugin)
		Expect(issues).To(HaveLen(1))
		Expect(issues[0].IsCritical).To(BeTrue())
		Expect(issues[0].Description).To(Equal("endpoint[0]"))

		svcList := &corev1.ServiceList{}
		Expect(k8sClient.List(ctx, svcList)).To(Succeed())
		Expect(svcList.Items).To(HaveLen(1))
		Expect(svcList.Items[0].Name).To(Equal("tensorboard"))
	})

	DescribeTable("creates label-identified metrics Services without changing composer labels",
		func(engineLabel string) {
			runner := runnerFromEndpoints(&endpoints.Config{
				Endpoints: []endpoints.Endpoint{
					{Services: []endpoints.Service{{Create: newTestService("serving", 8000)}}},
					{Services: []endpoints.Service{metricsService("model-metrics", engineLabel)}},
				},
			})

			Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())

			metrics := &corev1.Service{}
			Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "model-metrics", Namespace: testNamespace}, metrics)).To(Succeed())
			Expect(metrics.Labels).To(HaveKeyWithValue(constants.LabelWorkloadMetricsService, "true"))
			Expect(metrics.Labels).To(HaveKeyWithValue(engineLabel, "true"))
			Expect(metrics.Labels).To(HaveKeyWithValue("workloadKind", constants.KindInferenceWorkload))
			Expect(metrics.Labels).To(HaveKeyWithValue(LabelWorkloadRunner, runner.Name))
			Expect(metrics.OwnerReferences).To(HaveLen(1))
			Expect(metrics.OwnerReferences[0].Name).To(Equal(runner.Name))

			serving := &corev1.Service{}
			Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "serving", Namespace: testNamespace}, serving)).To(Succeed())
			Expect(serving.Labels).NotTo(HaveKey(LabelWorkloadRunner))
		},
		Entry("NIM labels", constants.LabelNIMWorkloadMetricsService),
		Entry("vLLM labels", constants.LabelVLLMWorkloadMetricsService),
	)

	It("reports a metrics Service through the ordinary endpoint additionalInfo", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{
				{Services: []endpoints.Service{{Create: newTestService("serving", 8000)}}},
				{Services: []endpoints.Service{metricsService("model-metrics", constants.LabelNIMWorkloadMetricsService)}},
			},
		})
		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())

		status := newPluginStatus()
		Expect(plugin.Status(ctx, runner, k8sClient, status, nil)).To(Succeed())
		details := parsedNetworkingDetails(status)
		Expect(details.Endpoints).To(HaveLen(2))
		Expect(details.Endpoints[0].Services[0].Name).To(Equal("serving"))
		Expect(details.Endpoints[1].Services[0].Name).To(Equal("model-metrics"))
		Expect(details.Endpoints[1].Services[0].Phase).To(Equal(endpoints.PhaseReady))
	})

	It("surfaces metrics apply failure like an ordinary Service and still reconciles siblings", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{
				{Services: []endpoints.Service{metricsService("model-metrics", constants.LabelNIMWorkloadMetricsService)}},
				{Services: []endpoints.Service{{Create: newTestService("serving", 8000)}}},
			},
		})
		failingClient := &patchFailingClient{Client: k8sClient, name: "model-metrics"}

		err := plugin.HandleAdditionalResources(ctx, runner, nil, nil, failingClient, state)
		Expect(err).To(MatchError(ContainSubstring("injected apply failure")))
		Expect(state.PluginIssues(plugin)).To(HaveLen(1))
		Expect(state.PluginIssues(plugin)[0].Description).To(Equal("endpoint[0]"))
		Expect(state.PluginIssues(plugin)[0].IsCritical).To(BeTrue())

		Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "serving", Namespace: testNamespace}, &corev1.Service{})).To(Succeed())

		statusWithIssue := newPluginStatus()
		Expect(plugin.Status(ctx, runner, k8sClient, statusWithIssue, state)).To(Succeed())
		Expect(statusWithIssue.Phase).To(Equal(dromosv1alpha1.PluginPhaseFailed))

		status := newPluginStatus()
		Expect(plugin.Status(ctx, runner, k8sClient, status, nil)).To(Succeed())
		details := parsedNetworkingDetails(status)
		Expect(details.Endpoints).To(HaveLen(2))
		Expect(details.Endpoints[0].Services[0].Phase).To(Equal(endpoints.PhaseCreating))
		Expect(details.Endpoints[1].Services[0].Phase).To(Equal(endpoints.PhaseReady))
	})

	It("recreates a manually deleted metrics Service", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{{
				Services: []endpoints.Service{metricsService("model-metrics", constants.LabelNIMWorkloadMetricsService)},
			}},
		})
		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())

		svc := &corev1.Service{}
		key := client.ObjectKey{Name: "model-metrics", Namespace: testNamespace}
		Expect(k8sClient.Get(ctx, key, svc)).To(Succeed())
		Expect(k8sClient.Delete(ctx, svc)).To(Succeed())

		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, reconcile.New())).To(Succeed())
		Expect(k8sClient.Get(ctx, key, &corev1.Service{})).To(Succeed())
	})

	It("prunes removed and renamed owned metrics Services but never ordinary Services", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{
				{Services: []endpoints.Service{{Create: newTestService("serving", 8000)}}},
				{Services: []endpoints.Service{metricsService("old-metrics", constants.LabelNIMWorkloadMetricsService)}},
			},
		})
		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())

		foreign := metricsService("foreign-metrics", constants.LabelNIMWorkloadMetricsService).Create
		foreign.Namespace = testNamespace
		foreign.Labels[LabelWorkloadRunner] = runner.Name
		foreign.OwnerReferences = []metav1.OwnerReference{{
			APIVersion: dromosv1alpha1.GroupVersion.String(),
			Kind:       constants.KindWorkloadRunner,
			Name:       "another-runner",
			UID:        "another-uid",
		}}
		Expect(k8sClient.Create(ctx, foreign)).To(Succeed())

		Expect(endpoints.Set(runner, &endpoints.Config{Endpoints: []endpoints.Endpoint{
			{Services: []endpoints.Service{{Create: newTestService("serving", 8000)}}},
			{Services: []endpoints.Service{metricsService("new-metrics", constants.LabelNIMWorkloadMetricsService)}},
		}})).To(Succeed())
		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, reconcile.New())).To(Succeed())

		Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "old-metrics", Namespace: testNamespace}, &corev1.Service{})).
			To(MatchError(ContainSubstring("not found")))
		Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "new-metrics", Namespace: testNamespace}, &corev1.Service{})).To(Succeed())
		Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "serving", Namespace: testNamespace}, &corev1.Service{})).To(Succeed())
		Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "foreign-metrics", Namespace: testNamespace}, &corev1.Service{})).To(Succeed())

		Expect(endpoints.Set(runner, &endpoints.Config{Endpoints: []endpoints.Endpoint{
			{Services: []endpoints.Service{{Create: newTestService("renamed-serving", 8000)}}},
		}})).To(Succeed())
		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, reconcile.New())).To(Succeed())
		Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "new-metrics", Namespace: testNamespace}, &corev1.Service{})).
			To(MatchError(ContainSubstring("not found")))
		Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "serving", Namespace: testNamespace}, &corev1.Service{})).To(Succeed())
		Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "renamed-serving", Namespace: testNamespace}, &corev1.Service{})).To(Succeed())
		Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "foreign-metrics", Namespace: testNamespace}, &corev1.Service{})).To(Succeed())
	})

	It("prunes an owned metrics Service when the final endpoint annotation is removed", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{{
				Services: []endpoints.Service{metricsService("model-metrics", constants.LabelNIMWorkloadMetricsService)},
			}},
		})
		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())

		foreign := metricsService("foreign-metrics", constants.LabelNIMWorkloadMetricsService).Create
		foreign.Namespace = testNamespace
		foreign.Labels[LabelWorkloadRunner] = runner.Name
		foreign.OwnerReferences = []metav1.OwnerReference{{
			APIVersion: dromosv1alpha1.GroupVersion.String(),
			Kind:       constants.KindWorkloadRunner,
			Name:       "another-runner",
			UID:        "another-uid",
		}}
		Expect(k8sClient.Create(ctx, foreign)).To(Succeed())

		delete(runner.Annotations, endpoints.Annotation)
		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, reconcile.New())).To(Succeed())

		Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "model-metrics", Namespace: testNamespace}, &corev1.Service{})).
			To(MatchError(ContainSubstring("not found")))
		Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "foreign-metrics", Namespace: testNamespace}, &corev1.Service{})).To(Succeed())
	})

	It("prunes an owned metrics Service when annotation networking transitions to legacy", func() {
		runner := runnerFromEndpoints(&endpoints.Config{
			Endpoints: []endpoints.Endpoint{{
				Services: []endpoints.Service{metricsService("model-metrics", constants.LabelNIMWorkloadMetricsService)},
			}},
		})
		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())

		delete(runner.Annotations, endpoints.Annotation)
		runner.Spec.Instructions = &dromosv1alpha1.Instructions{
			Networking: &dromosv1alpha1.NetworkingConfig{},
		}
		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, reconcile.New())).To(Succeed())

		Expect(k8sClient.Get(ctx, client.ObjectKey{Name: "model-metrics", Namespace: testNamespace}, &corev1.Service{})).
			To(MatchError(ContainSubstring("not found")))
	})

	It("is a no-op for a runner with neither legacy networking nor the annotation", func() {
		runner := &dromosv1alpha1.WorkloadRunner{
			ObjectMeta: metav1.ObjectMeta{Name: testWorkloadName, Namespace: testNamespace},
		}

		Expect(plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)).To(Succeed())

		svcList := &corev1.ServiceList{}
		Expect(k8sClient.List(ctx, svcList)).To(Succeed())
		Expect(svcList.Items).To(BeEmpty())

		deployList := &appsv1.DeploymentList{}
		Expect(k8sClient.List(ctx, deployList)).To(Succeed())
		Expect(deployList.Items).To(BeEmpty())

		ingressList := &networkingv1.IngressList{}
		Expect(k8sClient.List(ctx, ingressList)).To(Succeed())
		Expect(ingressList.Items).To(BeEmpty())
	})
})
