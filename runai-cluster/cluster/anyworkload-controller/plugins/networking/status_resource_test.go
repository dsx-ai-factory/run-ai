package networking

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openshiftv1 "github.com/openshift/api/route/v1"
	endpoints "github.com/run-ai/runai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/dromos/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// newPluginStatus mirrors how the framework initializes plugin status before calling Status().
func newPluginStatus() *dromosv1alpha1.PluginStatus {
	return &dromosv1alpha1.PluginStatus{
		Phase:      dromosv1alpha1.PluginPhaseUnknown,
		Conditions: []metav1.Condition{},
	}
}

func parsedNetworkingDetails(status *dromosv1alpha1.PluginStatus) *endpoints.NetworkingStatus {
	details, err := endpoints.NetworkingStatusFromRawExtension(status.AdditionalInfo)
	Expect(err).NotTo(HaveOccurred())
	return details
}

func newAuthConfig() *dromosv1alpha1.NetworkingConfig {
	return &dromosv1alpha1.NetworkingConfig{
		Services: []dromosv1alpha1.NetworkingServiceItem{{
			Create: newTestService("my-svc", 8080),
			Auth: &dromosv1alpha1.NetworkingAuthConfig{
				AuthorizationType: dromosv1alpha1.AuthorizationTypeAuthenticatedUsers,
			},
		}},
	}
}

func newAuthConfigWithIngress(ingressName string) *dromosv1alpha1.NetworkingConfig {
	config := newAuthConfig()
	config.ExternalAccess = &dromosv1alpha1.NetworkingExternalAccessConfig{
		Ingress: newTestIngress(ingressName, "test.example.com", "/"),
	}
	return config
}

func newAuthConfigWithRoute(routeName string) *dromosv1alpha1.NetworkingConfig {
	config := newAuthConfig()
	config.ExternalAccess = &dromosv1alpha1.NetworkingExternalAccessConfig{
		Route: newTestRoute(routeName, "test.example.com", "/"),
	}
	return config
}

func newAuthConfigWithHTTPRoute(httpRouteName string) *dromosv1alpha1.NetworkingConfig {
	config := newAuthConfig()
	config.ExternalAccess = &dromosv1alpha1.NetworkingExternalAccessConfig{
		HTTPRoute: statusHTTPRoute("", httpRouteName, "test.example.com", "my-svc", 8080, nil),
	}
	return config
}

func newNoAuthConfigWithIngress(ingressName string) *dromosv1alpha1.NetworkingConfig {
	return &dromosv1alpha1.NetworkingConfig{
		Services: []dromosv1alpha1.NetworkingServiceItem{{
			Create: newTestService("my-svc", 8080),
		}},
		ExternalAccess: &dromosv1alpha1.NetworkingExternalAccessConfig{
			Ingress: newTestIngress(ingressName, "test.example.com", "/"),
		},
	}
}

func statusHTTPRoute(namespace, name, host, serviceName string, port int32, accepted *metav1.Condition) *gatewayv1.HTTPRoute {
	routePort := gatewayv1.PortNumber(port)
	httpRoute := &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: gatewayv1.HTTPRouteSpec{
			Hostnames: []gatewayv1.Hostname{gatewayv1.Hostname(host)},
			Rules: []gatewayv1.HTTPRouteRule{{
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
	if accepted != nil {
		httpRoute.Status.Parents = []gatewayv1.RouteParentStatus{{
			Conditions: []metav1.Condition{*accepted},
		}}
	}
	return httpRoute
}

func routeAcceptedCondition(status metav1.ConditionStatus, message string) *metav1.Condition {
	return &metav1.Condition{
		Type:    string(gatewayv1.RouteConditionAccepted),
		Status:  status,
		Message: message,
	}
}

func newAuthProxyDeployment(runner *dromosv1alpha1.WorkloadRunner, availableReplicas, totalReplicas int32) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getAuthProxyName(runner),
			Namespace: runner.Namespace,
		},
		Status: appsv1.DeploymentStatus{
			Replicas:          totalReplicas,
			AvailableReplicas: availableReplicas,
		},
	}
}

var _ = Describe("Status", func() {
	var (
		ctx       context.Context
		plugin    *Plugin
		k8sClient client.Client
	)

	BeforeEach(func() {
		ctx = context.Background()
		plugin = NewPlugin(nil)
		k8sClient = fake.NewClientBuilder().WithScheme(newTestScheme()).Build()
	})

	It("returns error if status is nil", func() {
		runner := newTestRunner(testWorkloadName, testNamespace, newAuthConfig())

		err := plugin.Status(ctx, runner, k8sClient, nil, nil)

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("nil status for plugin"))
	})

	It("does not set conditions when no networking config", func() {
		runner := newTestRunner(testWorkloadName, testNamespace, nil)
		status := newPluginStatus()
		raw, err := endpoints.NetworkingStatusToRawExtension(endpoints.NewNetworkingStatus(1))
		Expect(err).NotTo(HaveOccurred())
		status.AdditionalInfo = raw

		err = plugin.Status(ctx, runner, k8sClient, status, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseDisabled))
		Expect(status.Conditions).To(BeEmpty())
		Expect(status.AdditionalInfo).To(BeNil())
	})

	It("does not set conditions when networking is empty", func() {
		runner := newTestRunner(testWorkloadName, testNamespace, &dromosv1alpha1.NetworkingConfig{})
		status := newPluginStatus()

		err := plugin.Status(ctx, runner, k8sClient, status, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseDisabled))
		Expect(status.Conditions).To(BeEmpty())
	})

	It("sets phase to Disabled when auth is not enabled", func() {
		config := &dromosv1alpha1.NetworkingConfig{
			Services: []dromosv1alpha1.NetworkingServiceItem{{
				Create: newTestService("my-svc", 8080),
			}},
		}
		runner := newTestRunner(testWorkloadName, testNamespace, config)
		status := newPluginStatus()

		err := plugin.Status(ctx, runner, k8sClient, status, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseDisabled))
		Expect(status.Conditions).To(HaveLen(1))
		authCond := findCondition(status.Conditions, AuthProxyReadyConditionName)
		Expect(authCond).NotTo(BeNil())
		Expect(authCond.Status).To(Equal(metav1.ConditionFalse))
		Expect(authCond.Reason).To(Equal(AuthProxyReasonDisabled))
	})

	It("sets phase to Creating when auth proxy Deployment is not found", func() {
		runner := newTestRunner(testWorkloadName, testNamespace, newAuthConfig())
		status := newPluginStatus()

		err := plugin.Status(ctx, runner, k8sClient, status, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseCreating))
		Expect(status.Conditions).To(HaveLen(1))
		Expect(status.Conditions[0].Type).To(Equal(AuthProxyReadyConditionName))
		Expect(status.Conditions[0].Reason).To(Equal(AuthProxyReasonNotFound))
	})

	It("sets phase to Initializing when auth proxy Deployment has no available replicas", func() {
		runner := newTestRunner(testWorkloadName, testNamespace, newAuthConfig())
		Expect(k8sClient.Create(ctx, newAuthProxyDeployment(runner, 0, 1))).To(Succeed())

		status := newPluginStatus()
		err := plugin.Status(ctx, runner, k8sClient, status, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseInitializing))
		Expect(status.Conditions[0].Reason).To(Equal(AuthProxyReasonNotReady))
	})

	It("sets phase to Ready when auth proxy Deployment is fully available and no external access", func() {
		runner := newTestRunner(testWorkloadName, testNamespace, newAuthConfig())
		Expect(k8sClient.Create(ctx, newAuthProxyDeployment(runner, 1, 1))).To(Succeed())

		status := newPluginStatus()
		err := plugin.Status(ctx, runner, k8sClient, status, nil)

		Expect(err).NotTo(HaveOccurred())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))
		Expect(status.Conditions).To(HaveLen(1))
		Expect(status.Conditions[0].Type).To(Equal(AuthProxyReadyConditionName))
		Expect(status.Conditions[0].Status).To(Equal(metav1.ConditionTrue))
	})

	It("updates status from Creating to Ready when Deployment becomes available", func() {
		runner := newTestRunner(testWorkloadName, testNamespace, newAuthConfig())
		status := newPluginStatus()

		err := plugin.Status(ctx, runner, k8sClient, status, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseCreating))

		Expect(k8sClient.Create(ctx, newAuthProxyDeployment(runner, 1, 1))).To(Succeed())

		err = plugin.Status(ctx, runner, k8sClient, status, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))
	})

	Context("with external access (Ingress)", func() {
		const ingressName = "my-ingress"

		It("sets phase to Creating when Deployment is ready but Ingress is not found", func() {
			runner := newTestRunner(testWorkloadName, testNamespace, newAuthConfigWithIngress(ingressName))
			Expect(k8sClient.Create(ctx, newAuthProxyDeployment(runner, 1, 1))).To(Succeed())

			status := newPluginStatus()
			err := plugin.Status(ctx, runner, k8sClient, status, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseCreating))
			Expect(status.Conditions).To(HaveLen(2))

			authCond := findCondition(status.Conditions, AuthProxyReadyConditionName)
			Expect(authCond).NotTo(BeNil())
			Expect(authCond.Status).To(Equal(metav1.ConditionTrue))

			accessCond := findCondition(status.Conditions, ExternalAccessReadyConditionName)
			Expect(accessCond).NotTo(BeNil())
			Expect(accessCond.Status).To(Equal(metav1.ConditionFalse))
			Expect(accessCond.Reason).To(Equal(ExternalAccessReasonNotFound))
		})

		It("sets phase to Ready when Deployment and Ingress both exist", func() {
			runner := newTestRunner(testWorkloadName, testNamespace, newAuthConfigWithIngress(ingressName))
			Expect(k8sClient.Create(ctx, newAuthProxyDeployment(runner, 1, 1))).To(Succeed())
			Expect(k8sClient.Create(ctx, &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{Name: ingressName, Namespace: testNamespace},
			})).To(Succeed())

			status := newPluginStatus()
			err := plugin.Status(ctx, runner, k8sClient, status, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))
			Expect(status.Conditions).To(HaveLen(2))

			accessCond := findCondition(status.Conditions, ExternalAccessReadyConditionName)
			Expect(accessCond).NotTo(BeNil())
			Expect(accessCond.Status).To(Equal(metav1.ConditionTrue))
		})

		It("sets phase to Initializing when Deployment is not ready and Ingress exists (worst case wins)", func() {
			runner := newTestRunner(testWorkloadName, testNamespace, newAuthConfigWithIngress(ingressName))
			Expect(k8sClient.Create(ctx, newAuthProxyDeployment(runner, 0, 1))).To(Succeed())
			Expect(k8sClient.Create(ctx, &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{Name: ingressName, Namespace: testNamespace},
			})).To(Succeed())

			status := newPluginStatus()
			err := plugin.Status(ctx, runner, k8sClient, status, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseInitializing))
		})

		It("preserves primary Ingress readiness while reporting HTTPRoute details", func() {
			const httpRouteName = "my-httproute"
			config := newAuthConfigWithIngress(ingressName)
			config.ExternalAccess.HTTPRoute = statusHTTPRoute("", httpRouteName, "gateway.example.com", "my-svc", 8080, nil)
			runner := newTestRunner(testWorkloadName, testNamespace, config)
			Expect(k8sClient.Create(ctx, newAuthProxyDeployment(runner, 1, 1))).To(Succeed())
			svc := newTestService("my-svc", 8080)
			svc.Namespace = testNamespace
			Expect(k8sClient.Create(ctx, svc)).To(Succeed())
			ingress := newTestIngress(ingressName, "test.example.com", "/")
			ingress.Namespace = testNamespace
			Expect(k8sClient.Create(ctx, ingress)).To(Succeed())
			Expect(k8sClient.Create(ctx, statusHTTPRoute(testNamespace, httpRouteName, "gateway.example.com", "my-svc", 8080, nil))).To(Succeed())

			status := newPluginStatus()
			err := plugin.Status(ctx, runner, k8sClient, status, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))

			accessCond := findCondition(status.Conditions, ExternalAccessReadyConditionName)
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
					URL:        "https://test.example.com/",
					SourceKind: endpoints.ExternalAccessKindIngress,
					SourceName: ingressName,
				},
				endpoints.ServiceURLStatus{
					URL:        "https://gateway.example.com/",
					SourceKind: endpoints.ExternalAccessKindHTTPRoute,
					SourceName: httpRouteName,
				},
			))
		})
	})

	Context("with external access (Route)", func() {
		const routeName = "my-route"

		It("sets phase to Creating when Deployment is ready but Route is not found", func() {
			runner := newTestRunner(testWorkloadName, testNamespace, newAuthConfigWithRoute(routeName))
			Expect(k8sClient.Create(ctx, newAuthProxyDeployment(runner, 1, 1))).To(Succeed())

			status := newPluginStatus()
			err := plugin.Status(ctx, runner, k8sClient, status, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseCreating))

			accessCond := findCondition(status.Conditions, ExternalAccessReadyConditionName)
			Expect(accessCond).NotTo(BeNil())
			Expect(accessCond.Reason).To(Equal(ExternalAccessReasonNotFound))
		})

		It("sets phase to Ready when Deployment and Route both exist", func() {
			runner := newTestRunner(testWorkloadName, testNamespace, newAuthConfigWithRoute(routeName))
			Expect(k8sClient.Create(ctx, newAuthProxyDeployment(runner, 1, 1))).To(Succeed())
			Expect(k8sClient.Create(ctx, &openshiftv1.Route{
				ObjectMeta: metav1.ObjectMeta{Name: routeName, Namespace: testNamespace},
			})).To(Succeed())

			status := newPluginStatus()
			err := plugin.Status(ctx, runner, k8sClient, status, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))
		})
	})

	Context("with external access (HTTPRoute)", func() {
		const httpRouteName = "my-httproute"

		It("sets phase to Ready when HTTPRoute acceptance status is not reported", func() {
			runner := newTestRunner(testWorkloadName, testNamespace, newAuthConfigWithHTTPRoute(httpRouteName))
			Expect(k8sClient.Create(ctx, newAuthProxyDeployment(runner, 1, 1))).To(Succeed())
			svc := newTestService("my-svc", 8080)
			svc.Namespace = testNamespace
			Expect(k8sClient.Create(ctx, svc)).To(Succeed())
			Expect(k8sClient.Create(ctx, statusHTTPRoute(testNamespace, httpRouteName, "test.example.com", "my-svc", 8080, nil))).To(Succeed())

			status := newPluginStatus()
			err := plugin.Status(ctx, runner, k8sClient, status, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))

			accessCond := findCondition(status.Conditions, ExternalAccessReadyConditionName)
			Expect(accessCond).NotTo(BeNil())
			Expect(accessCond.Status).To(Equal(metav1.ConditionTrue))
			Expect(accessCond.Reason).To(Equal(ExternalAccessReadyConditionName))

			details := parsedNetworkingDetails(status)
			resource := details.Endpoints[0].ExternalAccess.Resources[0]
			Expect(resource.Kind).To(Equal(endpoints.ExternalAccessKindHTTPRoute))
			Expect(resource.Phase).To(Equal(endpoints.PhaseReady))
			Expect(resource.Reason).To(Equal(ExternalAccessReadyConditionName))
			Expect(resource.Accepted).To(BeNil())
			Expect(details.Endpoints[0].Services[0].URLs).To(ConsistOf(endpoints.ServiceURLStatus{
				URL:        "https://test.example.com/",
				SourceKind: endpoints.ExternalAccessKindHTTPRoute,
				SourceName: httpRouteName,
			}))
		})

		It("sets phase to Ready when HTTPRoute is accepted", func() {
			runner := newTestRunner(testWorkloadName, testNamespace, newAuthConfigWithHTTPRoute(httpRouteName))
			Expect(k8sClient.Create(ctx, newAuthProxyDeployment(runner, 1, 1))).To(Succeed())
			svc := newTestService("my-svc", 8080)
			svc.Namespace = testNamespace
			Expect(k8sClient.Create(ctx, svc)).To(Succeed())
			Expect(k8sClient.Create(ctx, statusHTTPRoute(testNamespace, httpRouteName, "test.example.com", "my-svc", 8080,
				routeAcceptedCondition(metav1.ConditionTrue, "")))).To(Succeed())

			status := newPluginStatus()
			err := plugin.Status(ctx, runner, k8sClient, status, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))

			accessCond := findCondition(status.Conditions, ExternalAccessReadyConditionName)
			Expect(accessCond).NotTo(BeNil())
			Expect(accessCond.Status).To(Equal(metav1.ConditionTrue))
			Expect(accessCond.Reason).To(Equal(ExternalAccessReadyConditionName))

			details := parsedNetworkingDetails(status)
			resource := details.Endpoints[0].ExternalAccess.Resources[0]
			Expect(resource.Phase).To(Equal(endpoints.PhaseReady))
			Expect(resource.Reason).To(Equal(ExternalAccessReadyConditionName))
			Expect(resource.Accepted).To(HaveValue(Equal(true)))
			Expect(details.Endpoints[0].Services[0].URLs).To(ConsistOf(endpoints.ServiceURLStatus{
				URL:        "https://test.example.com/",
				SourceKind: endpoints.ExternalAccessKindHTTPRoute,
				SourceName: httpRouteName,
			}))
		})

		It("summarizes HTTPRoute acceptance deterministically across parents", func() {
			rejectedParent := gatewayv1.RouteParentStatus{
				Conditions: []metav1.Condition{*routeAcceptedCondition(metav1.ConditionFalse, "listener rejected")},
			}
			acceptedParent := gatewayv1.RouteParentStatus{
				Conditions: []metav1.Condition{*routeAcceptedCondition(metav1.ConditionTrue, "")},
			}

			for _, parents := range [][]gatewayv1.RouteParentStatus{
				{rejectedParent, acceptedParent},
				{acceptedParent, rejectedParent},
			} {
				route := statusHTTPRoute(testNamespace, httpRouteName, "test.example.com", "my-svc", 8080, nil)
				route.Status.Parents = parents

				acceptance := httpRouteAcceptanceResult(route)

				Expect(acceptance.phase).To(Equal(dromosv1alpha1.PluginPhaseReady))
				Expect(acceptance.status).To(Equal(metav1.ConditionTrue))
				Expect(acceptance.reason).To(Equal(ExternalAccessReadyConditionName))
				Expect(acceptance.accepted).To(HaveValue(Equal(true)))
			}
		})

		It("reports HTTPRoute rejection when no parent accepted it", func() {
			route := statusHTTPRoute(testNamespace, httpRouteName, "test.example.com", "my-svc", 8080, nil)
			route.Status.Parents = []gatewayv1.RouteParentStatus{
				{Conditions: []metav1.Condition{*routeAcceptedCondition(metav1.ConditionUnknown, "waiting for gateway")}},
				{Conditions: []metav1.Condition{*routeAcceptedCondition(metav1.ConditionFalse, "listener rejected")}},
			}

			acceptance := httpRouteAcceptanceResult(route)

			Expect(acceptance.phase).To(Equal(dromosv1alpha1.PluginPhaseInitializing))
			Expect(acceptance.status).To(Equal(metav1.ConditionFalse))
			Expect(acceptance.reason).To(Equal(HTTPRouteNotAcceptedReason))
			Expect(acceptance.message).To(Equal("listener rejected"))
			Expect(acceptance.accepted).To(HaveValue(Equal(false)))
		})
	})

	Context("with external access (Ingress) and auth disabled", func() {
		const ingressName = "no-auth-ingress"

		It("sets phase to Creating when Ingress is not found", func() {
			runner := newTestRunner(testWorkloadName, testNamespace, newNoAuthConfigWithIngress(ingressName))
			status := newPluginStatus()
			err := plugin.Status(ctx, runner, k8sClient, status, nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseCreating))
			authCond := findCondition(status.Conditions, AuthProxyReadyConditionName)
			Expect(authCond).NotTo(BeNil())
			Expect(authCond.Reason).To(Equal(AuthProxyReasonDisabled))
			accessCond := findCondition(status.Conditions, ExternalAccessReadyConditionName)
			Expect(accessCond).NotTo(BeNil())
			Expect(accessCond.Reason).To(Equal(ExternalAccessReasonNotFound))
		})

		It("reports full PluginStatus when auth is disabled and ingress exists", func() {
			runner := newTestRunner(testWorkloadName, testNamespace, newNoAuthConfigWithIngress(ingressName))
			svc := newTestService("my-svc", 8080)
			svc.Namespace = testNamespace
			svc.Spec.Type = corev1.ServiceTypeNodePort
			svc.Spec.Ports[0].Name = "http"
			svc.Spec.Ports[0].NodePort = 32123
			svc.Spec.Ports[0].TargetPort = intstr.FromInt32(8080)
			svc.Spec.Ports[0].Protocol = corev1.ProtocolTCP
			Expect(k8sClient.Create(ctx, svc)).To(Succeed())
			Expect(k8sClient.Create(ctx, &networkingv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{Name: ingressName, Namespace: testNamespace},
				Spec: networkingv1.IngressSpec{
					Rules: []networkingv1.IngressRule{{Host: "test.example.com"}},
				},
			})).To(Succeed())

			status := newPluginStatus()
			err := plugin.Status(ctx, runner, k8sClient, status, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))
			Expect(status.Conditions).To(HaveLen(2))

			authCond := findCondition(status.Conditions, AuthProxyReadyConditionName)
			Expect(authCond).NotTo(BeNil())
			Expect(authCond.Status).To(Equal(metav1.ConditionFalse))
			Expect(authCond.Reason).To(Equal(AuthProxyReasonDisabled))
			Expect(authCond.Message).To(Equal("Auth proxy is disabled"))

			accessCond := findCondition(status.Conditions, ExternalAccessReadyConditionName)
			Expect(accessCond).NotTo(BeNil())
			Expect(accessCond.Status).To(Equal(metav1.ConditionTrue))
			Expect(accessCond.Reason).To(Equal(ExternalAccessReadyConditionName))
			Expect(accessCond.Message).To(Equal(fmt.Sprintf("Ingress %s exists, url: https://test.example.com/", ingressName)))

			details := parsedNetworkingDetails(status)
			Expect(details.SchemaVersion).To(Equal(endpoints.StatusSchemaVersionV1))
			Expect(details.Endpoints).To(HaveLen(1))
			Expect(details.Endpoints[0].Services).To(HaveLen(1))
			service := details.Endpoints[0].Services[0]
			Expect(service.Name).To(Equal("my-svc"))
			Expect(service.Phase).To(Equal(endpoints.PhaseReady))
			Expect(service.Type).To(Equal(string(corev1.ServiceTypeNodePort)))
			Expect(service.Ports).To(HaveLen(1))
			Expect(service.Ports[0].Name).To(Equal("http"))
			Expect(service.Ports[0].Port).To(Equal(int32(32123)))
			Expect(service.Ports[0].TargetPort).To(Equal(intstr.FromInt32(8080)))
			Expect(service.AuthProxyReady).To(BeNil())
			Expect(service.URLs).To(ConsistOf(endpoints.ServiceURLStatus{
				URL:        "https://test.example.com/",
				SourceKind: endpoints.ExternalAccessKindIngress,
				SourceName: ingressName,
			}))
		})
	})
})

func findCondition(conditions []metav1.Condition, condType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == condType {
			return &conditions[i]
		}
	}
	return nil
}
