/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inferenceendpoint

import (
	"context"
	"encoding/json"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer/composer_v2/dinf_to_lws"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	workloadsapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	lwsv1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

var _ = Describe("Distributed inference endpoint intent", func() {
	var lws *lwsv1.LeaderWorkerSet
	var endpoint api.NetworkingEndpoint
	var defaults *Defaults
	var route gatewayv1.HTTPRoute

	BeforeEach(func() {
		lws = &lwsv1.LeaderWorkerSet{
			ObjectMeta: metav1.ObjectMeta{Name: "model-1234", Namespace: "project"},
			Spec: lwsv1.LeaderWorkerSetSpec{
				Replicas: ptr.To(int32(2)),
				LeaderWorkerTemplate: lwsv1.LeaderWorkerTemplate{
					Size: ptr.To(int32(4)),
					LeaderTemplate: &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{{
						Name: "model", Image: "example.com/model:1", Ports: []corev1.ContainerPort{{ContainerPort: 8000}},
					}}}},
				},
			},
		}
		endpoint = api.NetworkingEndpoint{Services: &[]api.NetworkingService{{Spec: &api.ServiceSpec{
			Name:     "dinfw-model-sp-0",
			Selector: &map[string]string{lwsv1.SetNameLabelKey: lws.Name, lwsv1.WorkerIndexLabelKey: "0"},
			Ports:    []api.ServicePort{{Port: 80, TargetPort: ptr.To(int32(8000)), Protocol: ptr.To(api.SERVICEPROTOCOL_TCP)}},
		}}}}
		defaults = &Defaults{GatewayName: "inference", GatewayNamespace: "gateways", EPPProfile: "gie-v1.5"}
		route = gatewayv1.HTTPRoute{
			TypeMeta: metav1.TypeMeta{APIVersion: gatewayv1.GroupVersion.String(), Kind: "HTTPRoute"},
			Spec: gatewayv1.HTTPRouteSpec{
				CommonRouteSpec: gatewayv1.CommonRouteSpec{ParentRefs: []gatewayv1.ParentReference{{
					Name: "inference", Namespace: ptr.To(gatewayv1.Namespace("gateways")),
				}}},
				Hostnames: []gatewayv1.Hostname{"model.example.com"},
				Rules: []gatewayv1.HTTPRouteRule{{
					Matches: []gatewayv1.HTTPRouteMatch{{Path: &gatewayv1.HTTPPathMatch{Type: ptr.To(gatewayv1.PathMatchPathPrefix), Value: ptr.To("/")}}},
					BackendRefs: []gatewayv1.HTTPBackendRef{{BackendRef: gatewayv1.BackendRef{BackendObjectReference: gatewayv1.BackendObjectReference{
						Name: "dinfw-model-sp-0", Port: ptr.To(gatewayv1.PortNumber(80)),
					}}}},
				}},
			},
		}
	})

	setRoute := func() {
		raw, err := json.Marshal(route)
		Expect(err).NotTo(HaveOccurred())
		var object map[string]interface{}
		Expect(json.Unmarshal(raw, &object)).To(Succeed())
		endpoint.ExternalAccess = &api.ExternalAccess{HttpRoute: &object}
	}

	It("returns an epp-only inference block for a validated topology without a path prefix", func() {
		setRoute()
		intent, err := BuildDistributedInferenceIntent(lws, endpoint, "", defaults)
		Expect(err).NotTo(HaveOccurred())
		Expect(intent).NotTo(BeNil())
		raw, err := json.Marshal(intent)
		Expect(err).NotTo(HaveOccurred())
		// Selector, port, Gateway and hostnames are derived on the cluster.
		Expect(string(raw)).To(MatchJSON(`{"epp":{"profile":"gie-v1.5"}}`))
	})

	It("carries the path prefix of the workload", func() {
		setRoute()
		intent, err := BuildDistributedInferenceIntent(lws, endpoint, "/openai/v1", defaults)
		Expect(err).NotTo(HaveOccurred())
		raw, err := json.Marshal(intent)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(raw)).To(MatchJSON(`{"epp":{"profile":"gie-v1.5"},"pathPrefix":"/openai/v1"}`))
	})

	composeNative := func(authorizationType *workloadsapi.DistributedInferenceServingPortAccessAuthorizationTypeEnum) (*lwsv1.LeaderWorkerSet, api.NetworkingEndpoint) {
		config := &composer.StaticClusterStorageConfig{}
		config.SetIngressEnabled(false)
		config.SetGatewayAPIEnabled(true)
		config.SetGatewayName(defaults.GatewayName)
		config.SetGatewayNamespace(defaults.GatewayNamespace)
		config.SetClusterDomain("example.com")
		spec := workloadsapi.NewDistributedInferenceSpecSpec()
		leader := workloadsapi.NewDistributedInferenceLeaderWorkerSpec()
		leader.SetImage("example.com/model:1")
		spec.SetLeader(*leader)
		port := workloadsapi.NewDistributedInferenceServingPort()
		port.SetPort(8000)
		port.SetExposeExternally(true)
		port.SetExposedUrl("https://model.example.com")
		if authorizationType != nil {
			port.SetAuthorizationType(*authorizationType)
			config.SetOauth2ProxyImage("example.com/oauth2-proxy:1")
			config.SetOauth2ProxyOidcIssuerUrl("https://auth.example.com")
		}
		spec.SetServingPort(*port)
		builder := dinf_to_lws.NewDistributedInferenceToLWSComposer("model", "project", "project", "user", nil, nil, composer.FieldsComposerOptions{}, nil).
			WithLeaderWorkerSetName("model-1234").WithClusterStorageConfig(config)
		target := dinf_to_lws.NewMinimalLeaderWorkerSetTarget()
		Expect(builder.Compose(context.Background(), spec, target)()).To(BeNil())
		Expect(target.Config.Networking).To(HaveLen(1))
		raw, err := json.Marshal(target.Config.Networking[0])
		Expect(err).NotTo(HaveOccurred())
		var resolved api.NetworkingEndpoint
		Expect(json.Unmarshal(raw, &resolved)).To(Succeed())
		return &target.LeaderWorkerSet, resolved
	}

	It("accepts the resolved output of the native distributed-inference composer", func() {
		composed, resolved := composeNative(nil)
		intent, err := BuildDistributedInferenceIntent(composed, resolved, "", defaults)
		Expect(err).NotTo(HaveOccurred())
		setRoute()
		expected, err := BuildDistributedInferenceIntent(lws, endpoint, "", defaults)
		Expect(err).NotTo(HaveOccurred())
		Expect(intent).To(Equal(expected))
	})

	It("rejects the protected output of the native composer", func() {
		authenticated := workloadsapi.DISTRIBUTEDINFERENCESERVINGPORTACCESSAUTHORIZATIONTYPEENUM_AUTHENTICATED_USERS
		composed, resolved := composeNative(&authenticated)
		service := (*resolved.Services)[0]
		Expect(service.Authorization).To(BeNil())
		Expect(*service.Spec.Ports[0].TargetPort).To(Equal(composer.Oauth2ContainerDefaultPort))

		_, err := BuildDistributedInferenceIntent(composed, resolved, "", defaults)
		Expect(err).To(MatchError(ContainSubstring("public distributed inference must not target the auth proxy port")))
	})

	It("accepts additional bound selectors that the leader template carries", func() {
		const ownerLabel = "run.ai/networking-workload-id"
		lws.Spec.LeaderWorkerTemplate.LeaderTemplate.Labels = map[string]string{ownerLabel: "workload-123"}
		(*(*endpoint.Services)[0].Spec.Selector)[ownerLabel] = "workload-123"
		setRoute()
		_, err := BuildDistributedInferenceIntent(lws, endpoint, "", defaults)
		Expect(err).NotTo(HaveOccurred())
	})

	DescribeTable("rejects unsupported serving topology", func(mutate func()) {
		setRoute()
		mutate()
		intent, err := BuildDistributedInferenceIntent(lws, endpoint, "", defaults)
		Expect(err).To(HaveOccurred())
		Expect(intent).To(BeNil())
	},
		Entry("no administrator defaults", func() { defaults = nil }),
		Entry("missing LWS", func() { lws = nil }),
		Entry("unsupported EPP profile", func() { defaults.EPPProfile = "unknown" }),
		Entry("internal-only endpoint", func() { endpoint.ExternalAccess = nil }),
		Entry("parallel Ingress", func() { endpoint.ExternalAccess.Ingress = &map[string]interface{}{} }),
		Entry("referenced Service", func() { (*endpoint.Services)[0].Spec = nil; (*endpoint.Services)[0].Ref = &api.ServiceRef{} }),
		Entry("all workers selected", func() { delete(*(*endpoint.Services)[0].Spec.Selector, lwsv1.WorkerIndexLabelKey) }),
		Entry("another LWS selected", func() { (*(*endpoint.Services)[0].Spec.Selector)[lwsv1.SetNameLabelKey] = "other" }),
		Entry("nonmatching extra selector", func() { (*(*endpoint.Services)[0].Spec.Selector)["app"] = "other" }),
		Entry("missing target port", func() { (*endpoint.Services)[0].Spec.Ports[0].TargetPort = nil }),
		Entry("auth proxy target", func() { (*endpoint.Services)[0].Spec.Ports[0].TargetPort = ptr.To(int32(9000)) }),
		Entry("UDP serving port", func() { (*endpoint.Services)[0].Spec.Ports[0].Protocol = ptr.To(api.SERVICEPROTOCOL_UDP) }),
		Entry("missing leader template", func() { lws.Spec.LeaderWorkerTemplate.LeaderTemplate = nil }),
	)

	DescribeTable("rejects a serving port that targets the auth proxy", func(mutate func()) {
		mutate()
		setRoute()
		intent, err := BuildDistributedInferenceIntent(lws, endpoint, "", defaults)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("port"))
		Expect(intent).To(BeNil())
	},
		Entry("on the auth proxy port declared by a sidecar", func() {
			(*endpoint.Services)[0].Spec.Ports[0].TargetPort = ptr.To(composer.Oauth2ContainerDefaultPort)
			template := lws.Spec.LeaderWorkerTemplate.LeaderTemplate
			template.Spec.Containers = append(template.Spec.Containers, corev1.Container{
				Name:  "oauth2-proxy",
				Ports: []corev1.ContainerPort{{ContainerPort: composer.Oauth2ContainerDefaultPort, Protocol: corev1.ProtocolTCP}},
			})
		}),
		Entry("on the auth proxy port declared by the leader container", func() {
			(*endpoint.Services)[0].Spec.Ports[0].TargetPort = ptr.To(composer.Oauth2ContainerDefaultPort)
			leader := lws.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0]
			leader.Ports = append(leader.Ports, corev1.ContainerPort{ContainerPort: composer.Oauth2ContainerDefaultPort, Protocol: corev1.ProtocolTCP})
			lws.Spec.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0] = leader
		}),
	)

	DescribeTable("rejects routing constraints that the endpoint CR cannot preserve", func(mutate func()) {
		mutate()
		setRoute()
		intent, err := BuildDistributedInferenceIntent(lws, endpoint, "", defaults)
		Expect(err).To(HaveOccurred())
		Expect(intent).To(BeNil())
	},
		Entry("another Gateway", func() { route.Spec.ParentRefs[0].Name = "other" }),
		Entry("another Gateway namespace", func() { route.Spec.ParentRefs[0].Namespace = ptr.To(gatewayv1.Namespace("other")) }),
		Entry("specific listener", func() { route.Spec.ParentRefs[0].SectionName = ptr.To(gatewayv1.SectionName("private")) }),
		Entry("path-based URL", func() { route.Spec.Rules[0].Matches[0].Path.Value = ptr.To("/project/model") }),
		Entry("header match", func() {
			route.Spec.Rules[0].Matches[0].Headers = []gatewayv1.HTTPHeaderMatch{{Name: "tenant", Value: "a"}}
		}),
		Entry("filter", func() {
			route.Spec.Rules[0].Filters = []gatewayv1.HTTPRouteFilter{{Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier}}
		}),
		Entry("different Service", func() { route.Spec.Rules[0].BackendRefs[0].Name = "other" }),
		Entry("wildcard hostname", func() { route.Spec.Hostnames = []gatewayv1.Hostname{"*.example.com"} }),
	)
})
