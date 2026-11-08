/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inferenceendpoint

import (
	"encoding/json"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var _ = Describe("NIM service endpoint intent", func() {
	const workloadName = "my-nim"
	var endpoint api.NetworkingEndpoint
	var defaults *Defaults
	var route gatewayv1.HTTPRoute

	BeforeEach(func() {
		// The NIM API references the operator Service, named after the workload, on the serving port.
		endpoint = api.NetworkingEndpoint{Services: &[]api.NetworkingService{{Ref: &api.ServiceRef{Name: workloadName, Port: 8000}}}}
		defaults = &Defaults{GatewayName: "inference", GatewayNamespace: "gateways", EPPProfile: "gie-v1.5"}
		route = gatewayv1.HTTPRoute{
			TypeMeta: metav1.TypeMeta{APIVersion: gatewayv1.GroupVersion.String(), Kind: "HTTPRoute"},
			Spec: gatewayv1.HTTPRouteSpec{
				CommonRouteSpec: gatewayv1.CommonRouteSpec{ParentRefs: []gatewayv1.ParentReference{{
					Name: "inference", Namespace: ptr.To(gatewayv1.Namespace("gateways")),
				}}},
				Hostnames: []gatewayv1.Hostname{"my-nim.example.com"},
				Rules: []gatewayv1.HTTPRouteRule{{
					Matches: []gatewayv1.HTTPRouteMatch{{Path: &gatewayv1.HTTPPathMatch{Type: ptr.To(gatewayv1.PathMatchPathPrefix), Value: ptr.To("/")}}},
					BackendRefs: []gatewayv1.HTTPBackendRef{{BackendRef: gatewayv1.BackendRef{BackendObjectReference: gatewayv1.BackendObjectReference{
						Name: workloadName, Port: ptr.To(gatewayv1.PortNumber(8000)),
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

	It("returns an epp-only inference block for a public root-hostname endpoint", func() {
		setRoute()
		intent, err := BuildNIMServiceIntent(endpoint, workloadName, defaults)
		Expect(err).NotTo(HaveOccurred())
		raw, err := json.Marshal(intent)
		Expect(err).NotTo(HaveOccurred())
		Expect(raw).To(MatchJSON(`{"epp":{"profile":"gie-v1.5"},"pathPrefix":"/v1"}`))
	})

	DescribeTable("rejects an endpoint that the InferenceEndpoint cannot serve", func(mutate func(), message string) {
		setRoute()
		mutate()
		intent, err := BuildNIMServiceIntent(endpoint, workloadName, defaults)
		Expect(err).To(MatchError(ContainSubstring(message)))
		Expect(intent).To(BeNil())
	},
		Entry("no administrator defaults", func() { defaults = nil }, "valid administrator Gateway defaults"),
		Entry("unsupported EPP profile", func() { defaults.EPPProfile = "unknown" }, "valid administrator Gateway defaults"),
		Entry("no serving Service", func() { endpoint.Services = nil }, "exactly one referenced serving Service"),
		Entry("created Service", func() {
			(*endpoint.Services)[0] = api.NetworkingService{Spec: &api.ServiceSpec{Name: workloadName, Ports: []api.ServicePort{{Port: 8000}}}}
		}, "reference to the serving port of the operator Service"),
		Entry("another Service", func() { (*endpoint.Services)[0].Ref.Name = "other" }, "reference to the serving port of the operator Service"),
		Entry("protected serving port", func() {
			(*endpoint.Services)[0].Authorization = &api.Workloadv2Authorization{AuthorizationType: ptr.To(api.AUTHORIZATIONTYPE_AUTHENTICATED_USERS)}
		}, "protected serving port is not supported yet"),
		Entry("internal-only endpoint", func() { endpoint.ExternalAccess = nil }, "requires an external Gateway HTTPRoute"),
		Entry("parallel Ingress", func() { endpoint.ExternalAccess.Ingress = &map[string]interface{}{} }, "without Ingress or OpenShift Route"),
		Entry("parallel OpenShift Route", func() { endpoint.ExternalAccess.Route = &map[string]interface{}{} }, "without Ingress or OpenShift Route"),
		Entry("Ingress only", func() {
			endpoint.ExternalAccess = &api.ExternalAccess{Ingress: &map[string]interface{}{}}
		}, "requires an external Gateway HTTPRoute"),
	)

	DescribeTable("rejects routing constraints that the endpoint CR cannot preserve", func(mutate func(), message string) {
		mutate()
		setRoute()
		intent, err := BuildNIMServiceIntent(endpoint, workloadName, defaults)
		Expect(err).To(MatchError(ContainSubstring(message)))
		Expect(intent).To(BeNil())
	},
		Entry("path-based URL", func() {
			route.Spec.Rules[0].Matches[0].Path.Value = ptr.To("/project/my-nim")
			route.Spec.Rules[0].Filters = []gatewayv1.HTTPRouteFilter{{Type: gatewayv1.HTTPRouteFilterURLRewrite}}
		}, "path-based URLs are not supported"),
		Entry("another Gateway", func() { route.Spec.ParentRefs[0].Name = "other" }, "must use the administrator Gateway"),
		Entry("another serving port", func() {
			route.Spec.Rules[0].BackendRefs[0].Port = ptr.To(gatewayv1.PortNumber(8001))
		}, "must use the administrator Gateway"),
		Entry("wildcard hostname", func() { route.Spec.Hostnames = []gatewayv1.Hostname{"*.example.com"} }, "one concrete hostname"),
	)
})
