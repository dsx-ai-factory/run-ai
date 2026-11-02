/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	endpoints "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile/issue"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/testutil"
	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	testInferenceHost       = "llama-chat-team-a.runai.example.com"
	testInferenceEndpointID = "inference-test-uid-12345"
	testInferenceBlock      = `{"epp":{"profile":"gie-v1.5"},` +
		`"objectives":[{"model":"llama-chat","priority":5}],` +
		`"rewrites":[{"model":"chat","targets":[{"model":"meta-llama/Llama-3.1-70B-Instruct","weight":1}]}]}`
)

func inferenceTestScheme() *runtime.Scheme {
	s := newTestScheme()
	s.AddKnownTypeWithName(endpoints.InferenceEndpointGVK, &unstructured.Unstructured{})
	s.AddKnownTypeWithName(
		endpoints.InferenceEndpointGVK.GroupVersion().WithKind("InferenceEndpointList"),
		&unstructured.UnstructuredList{},
	)
	return s
}

func inferenceTestMapper(installed bool) meta.RESTMapper {
	mapper := meta.NewDefaultRESTMapper(nil)
	if installed {
		mapper.Add(endpoints.InferenceEndpointGVK, meta.RESTScopeNamespace)
	}
	return mapper
}

func inferenceTestBlock() *endpoints.InferenceConfig {
	cfg := &endpoints.InferenceConfig{}
	Expect(json.Unmarshal([]byte(testInferenceBlock), cfg)).To(Succeed())
	return cfg
}

func llamaLeaderService(targetPort int32) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "llama-chat-leader"},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"leaderworkerset.sigs.k8s.io/name":         "llama-chat",
				"leaderworkerset.sigs.k8s.io/worker-index": "0",
			},
			Ports: []corev1.ServicePort{{Port: 8000, TargetPort: intstr.FromInt32(targetPort)}},
		},
	}
}

func llamaHTTPRoute() *gatewayv1.HTTPRoute {
	route := newTestHTTPRouteWithPath("llama-chat", testInferenceHost, "/", "llama-chat-leader", 8000)
	namespace := gatewayv1.Namespace("runai")
	route.Spec.ParentRefs = []gatewayv1.ParentReference{{Name: "runai-gateway", Namespace: &namespace}}
	return route
}

func inferenceEndpointConfig(auth bool, targetPort int32, inference *endpoints.InferenceConfig) *endpoints.Config {
	service := endpoints.Service{Create: llamaLeaderService(targetPort)}
	if auth {
		service.Auth = &endpoints.AuthConfig{AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers}
	}
	return &endpoints.Config{Endpoints: []endpoints.Endpoint{{
		Services: []endpoints.Service{service},
		ExternalAccess: &endpoints.ExternalAccess{
			HTTPRoute: llamaHTTPRoute(),
			Inference: inference,
		},
	}}}
}

func setRunnerEndpoints(runner *dromosv1alpha1.WorkloadRunner, cfg *endpoints.Config) {
	cfg.SchemaVersion = endpoints.SchemaVersionV1
	raw, err := json.Marshal(cfg)
	Expect(err).NotTo(HaveOccurred())
	runner.Annotations[endpoints.Annotation] = string(raw)
}

func specJSON(obj *unstructured.Unstructured) string {
	raw, err := json.Marshal(obj.Object["spec"])
	Expect(err).NotTo(HaveOccurred())
	return string(raw)
}

var _ = Describe("Inference routing (endpoints path)", func() {
	var (
		ctx       context.Context
		k8sClient client.Client
		plugin    *Plugin
		state     *reconcile.State
		runner    *dromosv1alpha1.WorkloadRunner
	)

	BeforeEach(func() {
		ctx = context.Background()
		for key, value := range map[string]string{
			workloadconfig.EnvVarGatewayName:      "runai-gateway",
			workloadconfig.EnvVarGatewayNamespace: "runai",
		} {
			previous, found := os.LookupEnv(key)
			Expect(os.Setenv(key, value)).To(Succeed())
			DeferCleanup(func() {
				if found {
					Expect(os.Setenv(key, previous)).To(Succeed())
				} else {
					Expect(os.Unsetenv(key)).To(Succeed())
				}
			})
		}
		k8sClient = testutil.WithSSASupport(fake.NewClientBuilder().
			WithScheme(inferenceTestScheme()).
			WithRESTMapper(inferenceTestMapper(true)).
			Build())
		plugin = newOwnershipTestPlugin(k8sClient)
		state = reconcile.New()
		runner = runnerFromEndpoints(inferenceEndpointConfig(false, 8000, inferenceTestBlock()))
		runner.Generation = 2
	})

	reconcileRunner := func() error {
		return plugin.HandleAdditionalResources(ctx, runner, nil, nil, k8sClient, state)
	}
	getInferenceEndpoint := func() (*unstructured.Unstructured, error) {
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(endpoints.InferenceEndpointGVK)
		err := k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: testInferenceEndpointID}, obj)
		return obj, err
	}
	httpRoutes := func() []gatewayv1.HTTPRoute {
		list := &gatewayv1.HTTPRouteList{}
		Expect(k8sClient.List(ctx, list)).To(Succeed())
		return list.Items
	}
	markReady := func(status metav1.ConditionStatus) {
		obj, err := getInferenceEndpoint()
		Expect(err).NotTo(HaveOccurred())
		obj.SetGeneration(1)
		obj.Object["status"] = map[string]any{
			"observedGeneration": int64(1),
			"conditions": []any{map[string]any{
				"type": "Ready", "status": string(status), "observedGeneration": int64(1),
				"reason": "RouteAccepted", "message": "routing ready", "lastTransitionTime": "2026-11-02T00:00:00Z",
			}},
		}
		Expect(k8sClient.Update(ctx, obj)).To(Succeed())
	}

	It("creates an InferenceEndpoint with the derived spec instead of the HTTPRoute", func() {
		Expect(reconcileRunner()).To(Succeed())

		obj, err := getInferenceEndpoint()
		Expect(err).NotTo(HaveOccurred())
		Expect(metav1.IsControlledBy(obj, runner)).To(BeTrue())
		Expect(specJSON(obj)).To(MatchJSON(`{
			"workload":{"selector":{"matchLabels":{"leaderworkerset.sigs.k8s.io/name":"llama-chat","leaderworkerset.sigs.k8s.io/worker-index":"0"}},"port":8000},
			"routing":{"gatewayRef":{"name":"runai-gateway","namespace":"runai"},"hostnames":["` + testInferenceHost + `"],
				"defaultBackendRef":{"name":"llama-chat-leader","port":8000}},
			"epp":{"profile":"gie-v1.5"},
			"objectives":[{"model":"llama-chat","priority":5}],
			"rewrites":[{"model":"chat","targets":[{"model":"meta-llama/Llama-3.1-70B-Instruct","weight":1}]}]}`))
		Expect(httpRoutes()).To(BeEmpty())
		Expect(k8sClient.Get(ctx, client.ObjectKey{Namespace: testNamespace, Name: "llama-chat-leader"}, &corev1.Service{})).To(Succeed())

		rv := obj.GetResourceVersion()
		Expect(reconcileRunner()).To(Succeed())
		again, err := getInferenceEndpoint()
		Expect(err).NotTo(HaveOccurred())
		Expect(again.GetResourceVersion()).To(Equal(rv))
	})

	It("creates an InferenceEndpoint from an epp-only inference block", func() {
		setRunnerEndpoints(runner, inferenceEndpointConfig(false, 8000, &endpoints.InferenceConfig{
			Epp: json.RawMessage(`{"profile":"gie-v1.5"}`),
		}))
		Expect(reconcileRunner()).To(Succeed())

		obj, err := getInferenceEndpoint()
		Expect(err).NotTo(HaveOccurred())
		Expect(obj.Object["spec"]).To(SatisfyAll(
			HaveKeyWithValue("epp", map[string]any{"profile": "gie-v1.5"}),
			Not(HaveKey("objectives")),
			Not(HaveKey("rewrites")),
		))
		Expect(httpRoutes()).To(BeEmpty())
		_, errs := plugin.WebhookValidate(ctx, nil, runner)
		Expect(errs).To(BeEmpty())
	})

	It("targets the leader's auth proxy sidecar for protected distributed inference", func() {
		// The composer protects distributed inference with a sidecar on 4180 and sets no Service auth.
		setRunnerEndpoints(runner, inferenceEndpointConfig(false, 4180, inferenceTestBlock()))
		Expect(reconcileRunner()).To(Succeed())

		obj, err := getInferenceEndpoint()
		Expect(err).NotTo(HaveOccurred())
		port, _, err := unstructured.NestedInt64(obj.Object, "spec", "workload", "port")
		Expect(err).NotTo(HaveOccurred())
		Expect(port).To(Equal(int64(4180)))
		backend, _, err := unstructured.NestedMap(obj.Object, "spec", "routing", "defaultBackendRef")
		Expect(err).NotTo(HaveOccurred())
		Expect(backend).To(Equal(map[string]any{"name": "llama-chat-leader", "port": int64(8000)}))
		deployments := &appsv1.DeploymentList{}
		Expect(k8sClient.List(ctx, deployments)).To(Succeed())
		Expect(deployments.Items).To(BeEmpty())
	})

	It("routes non-inference paths through the auth proxy when the Service has auth", func() {
		runner = runnerFromEndpoints(inferenceEndpointConfig(true, 4180, inferenceTestBlock()))
		Expect(reconcileRunner()).To(Succeed())

		obj, err := getInferenceEndpoint()
		Expect(err).NotTo(HaveOccurred())
		port, _, err := unstructured.NestedInt64(obj.Object, "spec", "workload", "port")
		Expect(err).NotTo(HaveOccurred())
		Expect(port).To(Equal(int64(4180)))
		backend, _, err := unstructured.NestedMap(obj.Object, "spec", "routing", "defaultBackendRef")
		Expect(err).NotTo(HaveOccurred())
		Expect(backend).To(Equal(map[string]any{
			"name": getEndpointAuthProxyName("llama-chat-leader"), "port": int64(AuthProxyServicePort),
		}))
		Expect(httpRoutes()).To(BeEmpty())
		Expect(k8sClient.Get(ctx, client.ObjectKey{
			Namespace: testNamespace, Name: getEndpointAuthProxyName("llama-chat-leader"),
		}, &appsv1.Deployment{})).To(Succeed())
		Expect(k8sClient.Get(ctx, client.ObjectKey{
			Namespace: testNamespace, Name: getEndpointAuthProxyName("llama-chat-leader"),
		}, &corev1.Service{})).To(Succeed())
	})

	It("switches between the composed HTTPRoute and the InferenceEndpoint when inference is added or removed", func() {
		setRunnerEndpoints(runner, inferenceEndpointConfig(false, 8000, nil))
		Expect(reconcileRunner()).To(Succeed())
		Expect(httpRoutes()).To(HaveLen(1))
		_, err := getInferenceEndpoint()
		Expect(apierrors.IsNotFound(err)).To(BeTrue())

		setRunnerEndpoints(runner, inferenceEndpointConfig(false, 8000, inferenceTestBlock()))
		Expect(reconcileRunner()).To(Succeed())
		_, err = getInferenceEndpoint()
		Expect(err).NotTo(HaveOccurred())
		Expect(httpRoutes()).To(BeEmpty())

		setRunnerEndpoints(runner, inferenceEndpointConfig(false, 8000, nil))
		Expect(reconcileRunner()).To(Succeed())
		_, err = getInferenceEndpoint()
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
		Expect(httpRoutes()).To(HaveLen(1))
		Expect(string(httpRoutes()[0].Spec.Hostnames[0])).To(Equal(testInferenceHost))
	})

	It("deletes the InferenceEndpoint when the endpoints annotation is removed", func() {
		Expect(reconcileRunner()).To(Succeed())
		delete(runner.Annotations, endpoints.Annotation)
		Expect(reconcileRunner()).To(Succeed())
		_, err := getInferenceEndpoint()
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})

	It("updates the spec while preserving unrelated metadata and status", func() {
		Expect(reconcileRunner()).To(Succeed())
		markReady(metav1.ConditionTrue)
		obj, err := getInferenceEndpoint()
		Expect(err).NotTo(HaveOccurred())
		obj.SetAnnotations(map[string]string{"keep": "value"})
		Expect(k8sClient.Update(ctx, obj)).To(Succeed())

		setRunnerEndpoints(runner, inferenceEndpointConfig(false, 9000, inferenceTestBlock()))
		Expect(reconcileRunner()).To(Succeed())
		obj, err = getInferenceEndpoint()
		Expect(err).NotTo(HaveOccurred())
		Expect(obj.GetAnnotations()).To(HaveKeyWithValue("keep", "value"))
		Expect(obj.Object).To(HaveKey("status"))
		port, _, err := unstructured.NestedInt64(obj.Object, "spec", "workload", "port")
		Expect(err).NotTo(HaveOccurred())
		Expect(port).To(Equal(int64(9000)))
	})

	It("refuses to adopt or delete an InferenceEndpoint controlled by another owner", func() {
		Expect(reconcileRunner()).To(Succeed())
		obj, err := getInferenceEndpoint()
		Expect(err).NotTo(HaveOccurred())
		obj.SetOwnerReferences(nil)
		Expect(k8sClient.Update(ctx, obj)).To(Succeed())

		Expect(reconcileRunner()).To(MatchError(ContainSubstring("not controlled")))

		setRunnerEndpoints(runner, inferenceEndpointConfig(false, 8000, nil))
		Expect(reconcileRunner()).To(Succeed())
		_, err = getInferenceEndpoint()
		Expect(err).NotTo(HaveOccurred())
	})

	It("records API failures as critical issues and reports the condition as failed", func() {
		k8sClient = interceptor.NewClient(k8sClient.(client.WithWatch), interceptor.Funcs{
			Create: func(ctx context.Context, c client.WithWatch, obj client.Object, opts ...client.CreateOption) error {
				if obj.GetObjectKind().GroupVersionKind() == endpoints.InferenceEndpointGVK {
					return fmt.Errorf("write denied")
				}
				return c.Create(ctx, obj, opts...)
			},
		})
		Expect(reconcileRunner()).To(MatchError(ContainSubstring("write denied")))
		status := newPluginStatus()
		Expect(plugin.Status(ctx, runner, k8sClient, status, state)).To(Succeed())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseFailed))
		condition := findEndpointsCondition(status, endpoints.InferenceEndpointReadyCondition)
		Expect(condition).NotTo(BeNil())
		Expect(condition.Status).To(Equal(metav1.ConditionFalse))
		Expect(condition.Reason).To(Equal("ReconcileFailed"))
		Expect(condition.ObservedGeneration).To(Equal(runner.Generation))
	})

	It("mirrors the InferenceEndpoint Ready condition and reports the httpRoute URL", func() {
		Expect(reconcileRunner()).To(Succeed())

		status := newPluginStatus()
		Expect(plugin.Status(ctx, runner, k8sClient, status, state)).To(Succeed())
		condition := findEndpointsCondition(status, endpoints.InferenceEndpointReadyCondition)
		Expect(condition).NotTo(BeNil())
		Expect(condition.Status).To(Equal(metav1.ConditionFalse))
		Expect(condition.Reason).To(Equal("Pending"))
		Expect(status.Phase).NotTo(Equal(dromosv1alpha1.PluginPhaseReady))

		markReady(metav1.ConditionFalse)
		Expect(plugin.Status(ctx, runner, k8sClient, status, state)).To(Succeed())
		condition = findEndpointsCondition(status, endpoints.InferenceEndpointReadyCondition)
		Expect(condition.Reason).To(Equal("EndpointNotReady"))
		Expect(findEndpointsCondition(status, ExternalAccessReadyConditionName).Reason).To(Equal(InferenceEndpointNotReadyReason))

		markReady(metav1.ConditionTrue)
		Expect(plugin.Status(ctx, runner, k8sClient, status, state)).To(Succeed())
		condition = findEndpointsCondition(status, endpoints.InferenceEndpointReadyCondition)
		Expect(condition.Status).To(Equal(metav1.ConditionTrue))
		Expect(condition.ObservedGeneration).To(Equal(runner.Generation))
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseReady))
		access := findEndpointsCondition(status, ExternalAccessReadyConditionName)
		Expect(access.Status).To(Equal(metav1.ConditionTrue))
		Expect(access.Message).To(HaveSuffix(", url: https://" + testInferenceHost + "/"))

		details, err := endpoints.NetworkingStatusFromRawExtension(status.AdditionalInfo)
		Expect(err).NotTo(HaveOccurred())
		Expect(details.Endpoints).To(HaveLen(1))
		Expect(details.Endpoints[0].ExternalAccess.Resources).To(ConsistOf(HaveField("Kind", endpoints.ExternalAccessKindInferenceEndpoint)))
		Expect(details.Endpoints[0].Services[0].URLs).To(ConsistOf(endpoints.ServiceURLStatus{
			URL:        "https://" + testInferenceHost + "/",
			SourceKind: endpoints.ExternalAccessKindInferenceEndpoint,
			SourceName: testInferenceEndpointID,
		}))

		setRunnerEndpoints(runner, inferenceEndpointConfig(false, 9000, inferenceTestBlock()))
		Expect(plugin.Status(ctx, runner, k8sClient, status, state)).To(Succeed())
		Expect(findEndpointsCondition(status, endpoints.InferenceEndpointReadyCondition).Status).To(Equal(metav1.ConditionFalse))
	})

	It("does not trust Ready on an InferenceEndpoint it does not control", func() {
		Expect(reconcileRunner()).To(Succeed())
		markReady(metav1.ConditionTrue)
		obj, err := getInferenceEndpoint()
		Expect(err).NotTo(HaveOccurred())
		obj.SetOwnerReferences(nil)
		Expect(k8sClient.Update(ctx, obj)).To(Succeed())

		status := newPluginStatus()
		Expect(plugin.Status(ctx, runner, k8sClient, status, state)).To(Succeed())
		Expect(status.Phase).To(Equal(dromosv1alpha1.PluginPhaseFailed))
		Expect(findEndpointsCondition(status, endpoints.InferenceEndpointReadyCondition).Status).To(Equal(metav1.ConditionFalse))
	})

	It("disables a previously reported condition once inference is removed", func() {
		status := newPluginStatus()
		status.SetCondition(metav1.Condition{
			Type: endpoints.InferenceEndpointReadyCondition, Status: metav1.ConditionTrue, Reason: "Ready",
		})
		runner.Status.Plugins = map[string]*dromosv1alpha1.PluginStatus{PluginName: status}
		setRunnerEndpoints(runner, inferenceEndpointConfig(false, 8000, nil))
		Expect(plugin.Status(ctx, runner, k8sClient, status, state)).To(Succeed())
		condition := findEndpointsCondition(status, endpoints.InferenceEndpointReadyCondition)
		Expect(condition.Status).To(Equal(metav1.ConditionFalse))
		Expect(condition.Reason).To(Equal("Disabled"))
	})

	It("does not report the condition for endpoints without inference", func() {
		setRunnerEndpoints(runner, inferenceEndpointConfig(false, 8000, nil))
		Expect(reconcileRunner()).To(Succeed())
		status := newPluginStatus()
		Expect(plugin.Status(ctx, runner, k8sClient, status, state)).To(Succeed())
		Expect(findEndpointsCondition(status, endpoints.InferenceEndpointReadyCondition)).To(BeNil())
	})

	It("requires the configured Gateway before creating an InferenceEndpoint", func() {
		cfg := inferenceEndpointConfig(false, 8000, inferenceTestBlock())
		cfg.Endpoints[0].ExternalAccess.HTTPRoute.Spec.ParentRefs[0].Name = "other-gateway"
		setRunnerEndpoints(runner, cfg)
		Expect(reconcileRunner()).To(Succeed())
		Expect(state.PluginIssues(plugin)).To(ContainElement(WithTransform(
			func(i issue.Issue) string { return i.Error() },
			ContainSubstring("must reference the configured Gateway"),
		)))
		_, err := getInferenceEndpoint()
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})

	It("does not recreate the InferenceEndpoint for a deleting runner", func() {
		runner.DeletionTimestamp = &metav1.Time{Time: metav1.Now().Time}
		Expect(reconcileRunner()).To(Succeed())
		_, err := getInferenceEndpoint()
		Expect(apierrors.IsNotFound(err)).To(BeTrue())
	})

	It("reports a missing InferenceEndpoint API only for workloads that request inference", func() {
		k8sClient = testutil.WithSSASupport(fake.NewClientBuilder().
			WithScheme(inferenceTestScheme()).
			WithRESTMapper(inferenceTestMapper(false)).
			Build())
		plugin = newOwnershipTestPlugin(k8sClient)

		Expect(reconcileRunner()).To(MatchError(ContainSubstring("InferenceEndpoint")))
		_, errs := plugin.WebhookValidate(ctx, nil, runner)
		Expect(errs.ToAggregate()).To(MatchError(ContainSubstring("InferenceEndpoint")))

		bldr := &builder.Builder{}
		watched, err := plugin.Watchers(bldr)
		Expect(err).NotTo(HaveOccurred())
		Expect(watched).NotTo(BeNil())

		setRunnerEndpoints(runner, inferenceEndpointConfig(false, 8000, nil))
		state = reconcile.New()
		Expect(reconcileRunner()).To(Succeed())
		_, errs = plugin.WebhookValidate(ctx, nil, runner)
		Expect(errs).To(BeEmpty())
	})

	It("admits an unchanged configuration after the InferenceEndpoint API is removed", func() {
		plugin.mapper = inferenceTestMapper(false)
		existing := runner.DeepCopy()
		_, errs := plugin.WebhookValidate(ctx, existing, runner)
		Expect(errs).To(BeEmpty())
	})

	DescribeTable("rejects inference routing that cannot be derived",
		func(mutate func(*endpoints.Config), expected string) {
			cfg := inferenceEndpointConfig(false, 8000, inferenceTestBlock())
			mutate(cfg)
			setRunnerEndpoints(runner, cfg)
			_, errs := plugin.WebhookValidate(ctx, nil, runner)
			Expect(errs.ToAggregate()).To(MatchError(ContainSubstring(expected)))
		},
		Entry("missing httpRoute", func(cfg *endpoints.Config) {
			cfg.Endpoints[0].ExternalAccess.HTTPRoute = nil
			cfg.Endpoints[0].ExternalAccess.Ingress = newTestIngress("llama", testInferenceHost, "/")
		}, "requires externalAccess.httpRoute"),
		Entry("referenced Service", func(cfg *endpoints.Config) {
			cfg.Endpoints[0].Services = []endpoints.Service{{
				Ref:  &endpoints.ServiceRef{Name: "existing", Port: 8000},
				Auth: &endpoints.AuthConfig{AuthorizationType: endpoints.AuthorizationTypeAuthenticatedUsers},
			}}
		}, "requires a created Service"),
		Entry("named target port", func(cfg *endpoints.Config) {
			cfg.Endpoints[0].Services[0].Create.Spec.Ports[0].TargetPort = intstr.FromString("http")
		}, "numeric Service targetPort"),
		Entry("no parentRef", func(cfg *endpoints.Config) {
			cfg.Endpoints[0].ExternalAccess.HTTPRoute.Spec.ParentRefs = nil
		}, "exactly one httpRoute.spec.parentRefs"),
		Entry("listener section", func(cfg *endpoints.Config) {
			section := gatewayv1.SectionName("https")
			cfg.Endpoints[0].ExternalAccess.HTTPRoute.Spec.ParentRefs[0].SectionName = &section
		}, "without sectionName or port"),
		Entry("no hostnames", func(cfg *endpoints.Config) {
			cfg.Endpoints[0].ExternalAccess.HTTPRoute.Spec.Hostnames = nil
		}, "httpRoute.spec.hostnames"),
		Entry("empty inference block, because the CRD has no default EPP profile", func(cfg *endpoints.Config) {
			cfg.Endpoints[0].ExternalAccess.Inference = &endpoints.InferenceConfig{}
		}, "inference.epp"),
		Entry("empty epp", func(cfg *endpoints.Config) {
			cfg.Endpoints[0].ExternalAccess.Inference = &endpoints.InferenceConfig{Epp: json.RawMessage(`{}`)}
		}, "inference.epp"),
		Entry("two inference endpoints", func(cfg *endpoints.Config) {
			second := inferenceEndpointConfig(false, 8000, inferenceTestBlock()).Endpoints[0]
			second.Services[0].Create.Name = "second"
			second.ExternalAccess.HTTPRoute.Name = "second"
			cfg.Endpoints = append(cfg.Endpoints, second)
		}, "at most one endpoint"),
	)

	It("admits a derivable inference endpoint", func() {
		_, errs := plugin.WebhookValidate(ctx, nil, runner)
		Expect(errs).To(BeEmpty())
	})
})
