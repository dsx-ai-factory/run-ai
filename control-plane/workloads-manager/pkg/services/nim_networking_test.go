/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-manager/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("validateNimServingPortUrlRequiresExposure", func() {
	It("returns an error when exposeExternally=false but a URL is provided", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(false),
			ExposedUrl:       convutils.Ref("https://my-nim.example.com"),
		}
		Expect(validateNimServingPortUrlRequiresExposure(sp)).NotTo(BeNil())
	})

	It("returns nil when exposeExternally=true and URL is set", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(true),
			ExposedUrl:       convutils.Ref("https://my-nim.example.com"),
		}
		Expect(validateNimServingPortUrlRequiresExposure(sp)).To(BeNil())
	})
})

var _ = Describe("validateNimServingPortAuthRequiresExposure", func() {
	It("returns an error when auth is set but exposeExternally=false", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			ExposeExternally:  convutils.Ref(false),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHENTICATED_USERS),
		}
		Expect(validateNimServingPortAuthRequiresExposure(sp)).NotTo(BeNil())
	})

	It("returns nil when auth is set and exposeExternally=true", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			ExposeExternally:  convutils.Ref(true),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHENTICATED_USERS),
		}
		Expect(validateNimServingPortAuthRequiresExposure(sp)).To(BeNil())
	})

	It("returns nil when no auth type is set", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(false),
		}
		Expect(validateNimServingPortAuthRequiresExposure(sp)).To(BeNil())
	})
})

var _ = Describe("validateNimServingPortAuthFieldsConsistency", func() {
	It("returns error when authorizedUsers is set without authorizationType", func() {
		sp := &api.NimServiceServingPort{
			Port:            convutils.Ref(int32(8000)),
			AuthorizedUsers: &[]string{"user@example.com"},
		}
		err := validateNimServingPortAuthFieldsConsistency(sp)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("authorizationType is required"))
	})

	It("returns error when authorizedGroups is set without authorizationType", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			AuthorizedGroups: &[]string{"admins"},
		}
		err := validateNimServingPortAuthFieldsConsistency(sp)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("authorizationType is required"))
	})

	It("returns error when empty authorizedUsers is set without authorizationType", func() {
		sp := &api.NimServiceServingPort{
			Port:            convutils.Ref(int32(8000)),
			AuthorizedUsers: &[]string{},
		}
		err := validateNimServingPortAuthFieldsConsistency(sp)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("authorizationType is required"))
	})

	It("returns error when empty authorizedGroups is set without authorizationType", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			AuthorizedGroups: &[]string{},
		}
		err := validateNimServingPortAuthFieldsConsistency(sp)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("authorizationType is required"))
	})

	It("returns nil when authorizationType is set with authorizedUsers", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHORIZED_USERS),
			AuthorizedUsers:   &[]string{"user@example.com"},
		}
		Expect(validateNimServingPortAuthFieldsConsistency(sp)).To(BeNil())
	})

	It("returns nil when no auth fields are set", func() {
		sp := &api.NimServiceServingPort{
			Port: convutils.Ref(int32(8000)),
		}
		Expect(validateNimServingPortAuthFieldsConsistency(sp)).To(BeNil())
	})

	It("returns nil when servingPort is nil", func() {
		Expect(validateNimServingPortAuthFieldsConsistency(nil)).To(BeNil())
	})

	It("returns error when both authorizedUsers and authorizedGroups are set", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHORIZED_USERS),
			AuthorizedUsers:   &[]string{"user@example.com"},
			AuthorizedGroups:  &[]string{"admins"},
		}
		err := validateNimServingPortAuthFieldsConsistency(sp)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("mutually exclusive"))
	})

	It("returns error when authorizedUsers is set for public type", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_PUBLIC),
			AuthorizedUsers:   &[]string{"user@example.com"},
		}
		err := validateNimServingPortAuthFieldsConsistency(sp)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("cannot be set for public"))
	})

	It("returns error when authorizedGroups is set for public type", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_PUBLIC),
			AuthorizedGroups:  &[]string{"admins"},
		}
		err := validateNimServingPortAuthFieldsConsistency(sp)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("cannot be set for public"))
	})

	It("returns error when authorizedUsers is set for authenticatedUsers type", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHENTICATED_USERS),
			AuthorizedUsers:   &[]string{"user@example.com"},
		}
		err := validateNimServingPortAuthFieldsConsistency(sp)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("cannot be set for authenticatedUsers"))
	})

	It("returns error when authorizedGroups is set for authenticatedUsers type", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHENTICATED_USERS),
			AuthorizedGroups:  &[]string{"admins"},
		}
		err := validateNimServingPortAuthFieldsConsistency(sp)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("cannot be set for authenticatedUsers"))
	})

	It("returns error when authorizedUsers is empty for authorizedUsers type", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHORIZED_USERS),
			AuthorizedUsers:   &[]string{},
		}
		err := validateNimServingPortAuthFieldsConsistency(sp)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("must be provided and non-empty"))
	})

	It("returns error when authorizedUsers is nil for authorizedUsers type", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHORIZED_USERS),
		}
		err := validateNimServingPortAuthFieldsConsistency(sp)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("must be provided and non-empty"))
	})

	It("returns error when authorizedGroups is set for authorizedUsers type", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHORIZED_USERS),
			AuthorizedGroups:  &[]string{"admins"},
		}
		err := validateNimServingPortAuthFieldsConsistency(sp)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("authorizedGroups cannot be set for authorizedUsers"))
	})

	It("returns error when authorizedGroups is empty for authorizedGroups type", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHORIZED_GROUPS),
			AuthorizedGroups:  &[]string{},
		}
		err := validateNimServingPortAuthFieldsConsistency(sp)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("must be provided and non-empty"))
	})

	It("returns error when authorizedUsers is set for authorizedGroups type", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHORIZED_GROUPS),
			AuthorizedUsers:   &[]string{"user@example.com"},
		}
		err := validateNimServingPortAuthFieldsConsistency(sp)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("authorizedUsers cannot be set for authorizedGroups"))
	})

	It("returns nil for valid authorizedGroups type", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHORIZED_GROUPS),
			AuthorizedGroups:  &[]string{"admins", "devs"},
		}
		Expect(validateNimServingPortAuthFieldsConsistency(sp)).To(BeNil())
	})

	It("returns nil for valid public type", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_PUBLIC),
		}
		Expect(validateNimServingPortAuthFieldsConsistency(sp)).To(BeNil())
	})
})

var _ = Describe("validateNimServingPortAuthSupportedForClusterVersion", func() {
	It("returns error when cluster < 2.25 and auth is set", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			ExposeExternally:  convutils.Ref(true),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHENTICATED_USERS),
		}
		err := validateNimServingPortAuthSupportedForClusterVersion(context.Background(), "2.24.0", sp)
		Expect(err).NotTo(BeNil())
		Expect(err.Message).To(ContainSubstring("supported for cluster version 2.25 and above"))
	})

	It("returns nil when cluster >= 2.25 and auth is set", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			ExposeExternally:  convutils.Ref(true),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHENTICATED_USERS),
		}
		Expect(validateNimServingPortAuthSupportedForClusterVersion(context.Background(), "2.25.0", sp)).To(BeNil())
	})

	It("returns nil when auth is not set regardless of version", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(true),
		}
		Expect(validateNimServingPortAuthSupportedForClusterVersion(context.Background(), "2.24.0", sp)).To(BeNil())
	})
})

var _ = Describe("nimServingPortToNetworkingConfig", func() {
	const workloadName = "my-nim"

	networkingConfig := func(sp *api.NimServiceServingPort, info nimServiceNetworkingInfo) *[]api.NetworkingEndpoint {
		cfg, err := nimServingPortToNetworkingConfig(workloadName, sp, info)
		Expect(err).To(BeNil())
		return cfg
	}

	singleNetworkingEndpoint := func(cfg *[]api.NetworkingEndpoint) api.NetworkingEndpoint {
		Expect(cfg).NotTo(BeNil())
		Expect(*cfg).To(HaveLen(1))
		return (*cfg)[0]
	}

	It("returns nil when servingPort is nil", func() {
		Expect(networkingConfig(nil, nimServiceNetworkingInfo{})).To(BeNil())
	})

	It("returns nil when port is nil", func() {
		Expect(networkingConfig(&api.NimServiceServingPort{}, nimServiceNetworkingInfo{})).To(BeNil())
	})

	It("returns nil when explicitly not exposed and no auth", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(false),
		}
		Expect(networkingConfig(sp, nimServiceNetworkingInfo{})).To(BeNil())
	})

	It("exposes by default when exposeExternally is nil", func() {
		sp := &api.NimServiceServingPort{
			Port:       convutils.Ref(int32(8000)),
			ExposedUrl: convutils.Ref("https://my-nim.example.com"),
		}
		cfg := networkingConfig(sp, nimServiceNetworkingInfo{})

		endpoint := singleNetworkingEndpoint(cfg)
		Expect(endpoint.ExternalAccess).NotTo(BeNil())
	})

	It("returns an error when exposedUrl has no host", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(true),
			ExposedUrl:       convutils.Ref("/missing-host"),
		}
		cfg, err := nimServingPortToNetworkingConfig(workloadName, sp, nimServiceNetworkingInfo{})

		Expect(cfg).To(BeNil())
		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(int32(400)))
		Expect(err.Message).To(Equal("invalid URL \"/missing-host\": missing host"))
	})

	It("sets ServiceRef with workload name and port when auth is requested", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			ExposeExternally:  convutils.Ref(true),
			ExposedUrl:        convutils.Ref("https://my-nim.example.com"),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHENTICATED_USERS),
		}
		cfg := networkingConfig(sp, nimServiceNetworkingInfo{})

		endpoint := singleNetworkingEndpoint(cfg)
		svc := (*endpoint.Services)[0]
		Expect(svc.Ref).NotTo(BeNil())
		Expect(svc.Ref.Name).To(Equal(workloadName))
		Expect(svc.Ref.Port).To(Equal(int32(8000)))
	})

	It("uses autoGeneratedUrl from info when no exposedUrl is set", func() {
		autoURL := "https://auto.example.com"
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(true),
		}
		cfg := networkingConfig(sp, nimServiceNetworkingInfo{AutoExposureUrl: &autoURL, IngressClass: "nginx"})

		endpoint := singleNetworkingEndpoint(cfg)
		Expect(endpoint.ExternalAccess).NotTo(BeNil())
		Expect(endpoint.ExternalAccess.Ingress).NotTo(BeNil())
		ingress := *endpoint.ExternalAccess.Ingress
		spec := ingress["spec"].(map[string]interface{})
		rules := spec["rules"].([]interface{})
		rule := rules[0].(map[string]interface{})
		Expect(rule["host"]).To(Equal("auto.example.com"))
	})

	It("does not include Services when only external exposure is requested (no auth)", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(true),
			ExposedUrl:       convutils.Ref("https://my-nim.example.com"),
		}
		cfg := networkingConfig(sp, nimServiceNetworkingInfo{IngressClass: "nginx"})

		endpoint := singleNetworkingEndpoint(cfg)
		Expect(endpoint.Services).To(BeNil())
		Expect(endpoint.ExternalAccess).NotTo(BeNil())
		Expect(endpoint.ExternalAccess.Ingress).NotTo(BeNil())
		spec := (*endpoint.ExternalAccess.Ingress)["spec"].(map[string]interface{})
		rules := spec["rules"].([]interface{})
		Expect(rules).To(HaveLen(1))
		Expect(rules[0].(map[string]interface{})["host"]).To(Equal("my-nim.example.com"))
	})

	It("sets Authorization when auth is explicitly requested alongside exposure", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			ExposeExternally:  convutils.Ref(true),
			ExposedUrl:        convutils.Ref("https://my-nim.example.com"),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHENTICATED_USERS),
		}
		cfg := networkingConfig(sp, nimServiceNetworkingInfo{})

		endpoint := singleNetworkingEndpoint(cfg)
		auth := (*endpoint.Services)[0].Authorization
		Expect(auth).NotTo(BeNil())
		Expect(*auth.AuthorizationType).To(Equal(api.AUTHORIZATIONTYPE_AUTHENTICATED_USERS))
		Expect(endpoint.ExternalAccess).NotTo(BeNil())
	})

	It("sets full Ingress with pathType ImplementationSpecific and regex path when URL has non-root path", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(true),
			ExposedUrl:       convutils.Ref("https://my-nim.example.com/api/v1"),
		}
		cfg := networkingConfig(sp, nimServiceNetworkingInfo{IngressClass: "nginx"})

		endpoint := singleNetworkingEndpoint(cfg)
		Expect(endpoint.ExternalAccess).NotTo(BeNil())
		Expect(endpoint.ExternalAccess.Ingress).NotTo(BeNil())
		spec := (*endpoint.ExternalAccess.Ingress)["spec"].(map[string]interface{})
		rules := spec["rules"].([]interface{})
		rule := rules[0].(map[string]interface{})
		http := rule["http"].(map[string]interface{})
		paths := http["paths"].([]interface{})
		pathObj := paths[0].(map[string]interface{})
		Expect(pathObj["path"]).To(Equal("/api/v1(/|$)(.*)"))
		Expect(pathObj["pathType"]).To(Equal("ImplementationSpecific"))
	})

	DescribeTable("sets the NIM Service backend on an OpenShift Route", func(authType *api.NimServiceServingPortAuthorizationType, hasAuth bool) {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			ExposeExternally:  convutils.Ref(true),
			ExposedUrl:        convutils.Ref("https://my-nim.example.com/foo"),
			AuthorizationType: authType,
		}
		cfg := networkingConfig(sp, nimServiceNetworkingInfo{UseOpenShiftRoute: true})

		endpoint := singleNetworkingEndpoint(cfg)
		Expect(endpoint.ExternalAccess).NotTo(BeNil())
		Expect(endpoint.ExternalAccess.Route).NotTo(BeNil())
		Expect(endpoint.ExternalAccess.Ingress).To(BeNil())
		Expect(endpoint.ExternalAccess.HttpRoute).To(BeNil())
		spec := (*endpoint.ExternalAccess.Route)["spec"].(map[string]interface{})
		Expect(spec["host"]).To(Equal("my-nim.example.com"))
		Expect(spec["path"]).To(Equal("/foo"))
		Expect(spec["to"]).To(Equal(map[string]interface{}{"kind": "Service", "name": workloadName}))
		Expect(spec).NotTo(HaveKey("port"))
		Expect((*endpoint.ExternalAccess.Route)["kind"]).To(Equal("Route"))
		Expect((*endpoint.ExternalAccess.Route)["apiVersion"]).To(Equal("route.openshift.io/v1"))
		if hasAuth {
			Expect(endpoint.Services).NotTo(BeNil())
			Expect(*endpoint.Services).To(HaveLen(1))
			svc := (*endpoint.Services)[0]
			Expect(svc.Ref).To(Equal(&api.ServiceRef{Name: workloadName, Port: 8000}))
			Expect(svc.Authorization).NotTo(BeNil())
			Expect(svc.Authorization.AuthorizationType).To(Equal(convutils.Ref(api.AUTHORIZATIONTYPE_AUTHENTICATED_USERS)))
		} else {
			Expect(endpoint.Services).To(BeNil())
		}
	},
		Entry("without authorization", nil, false),
		Entry("with explicit public access", convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_PUBLIC), false),
		Entry("with authenticated access", convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHENTICATED_USERS), true),
	)

	It("sets authorizedGroups on the auth config", func() {
		sp := &api.NimServiceServingPort{
			Port:              convutils.Ref(int32(8000)),
			ExposeExternally:  convutils.Ref(true),
			ExposedUrl:        convutils.Ref("https://my-nim.example.com"),
			AuthorizationType: convutils.Ref(api.NIMSERVICESERVINGPORTAUTHORIZATIONTYPE_AUTHORIZED_GROUPS),
			AuthorizedGroups:  &[]string{"group1"},
		}
		cfg := networkingConfig(sp, nimServiceNetworkingInfo{})

		endpoint := singleNetworkingEndpoint(cfg)
		auth := (*endpoint.Services)[0].Authorization
		Expect(auth).NotTo(BeNil())
		Expect(*auth.AuthorizationType).To(Equal(api.AUTHORIZATIONTYPE_AUTHORIZED_GROUPS))
		Expect(auth.AuthorizedUsers).To(BeNil())
		Expect(*auth.AuthorizedGroups).To(Equal([]string{"group1"}))
	})

	It("sets nginx GRPC backend annotation when grpcPort is set and ingressClass is nginx", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(true),
			ExposedUrl:       convutils.Ref("https://my-nim.example.com"),
			GrpcPort:         convutils.Ref(int32(8001)),
		}
		cfg := networkingConfig(sp, nimServiceNetworkingInfo{IngressClass: "nginx"})

		endpoint := singleNetworkingEndpoint(cfg)
		ingress := *endpoint.ExternalAccess.Ingress
		meta := ingress["metadata"].(map[string]interface{})
		annotations := meta["annotations"].(map[string]interface{})
		Expect(annotations[annotationNginxBackendProtocol]).To(Equal(backendProtocolGRPCNginx))
		Expect(annotations).NotTo(HaveKey(annotationHAProxyBackendProtocol))
	})

	It("sets haproxy GRPC backend annotation when grpcPort is set and ingressClass is haproxy", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(true),
			ExposedUrl:       convutils.Ref("https://my-nim.example.com"),
			GrpcPort:         convutils.Ref(int32(8001)),
		}
		cfg := networkingConfig(sp, nimServiceNetworkingInfo{IngressClass: ingressClassHAProxy})

		endpoint := singleNetworkingEndpoint(cfg)
		ingress := *endpoint.ExternalAccess.Ingress
		meta := ingress["metadata"].(map[string]interface{})
		annotations := meta["annotations"].(map[string]interface{})
		Expect(annotations[annotationHAProxyBackendProtocol]).To(Equal(backendProtocolGRPCHAProxy))
		Expect(annotations).NotTo(HaveKey(annotationNginxBackendProtocol))
	})

	It("uses haproxy rewrite annotation when ingressClass is haproxy", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(true),
			ExposedUrl:       convutils.Ref("https://my-nim.example.com/api/v1"),
		}
		cfg := networkingConfig(sp, nimServiceNetworkingInfo{IngressClass: ingressClassHAProxy})

		endpoint := singleNetworkingEndpoint(cfg)
		ingress := *endpoint.ExternalAccess.Ingress
		meta := ingress["metadata"].(map[string]interface{})
		annotations := meta["annotations"].(map[string]interface{})
		Expect(annotations).To(HaveKey(annotationHAProxyRewriteTarget))
		Expect(annotations).NotTo(HaveKey(annotationNginxRewriteTarget))
	})

	It("sets Ingress and HTTPRoute when Gateway API is enabled on a non-OpenShift cluster", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(true),
			ExposedUrl:       convutils.Ref("https://my-nim.example.com"),
		}
		cfg := networkingConfig(sp, nimServiceNetworkingInfo{
			UseGatewayAPI:    true,
			GatewayName:      "my-gateway",
			GatewayNamespace: "gateway-ns",
		})

		endpoint := singleNetworkingEndpoint(cfg)
		Expect(endpoint.ExternalAccess).NotTo(BeNil())
		Expect(endpoint.ExternalAccess.HttpRoute).NotTo(BeNil())
		Expect(endpoint.ExternalAccess.Ingress).NotTo(BeNil())
		Expect(endpoint.ExternalAccess.Route).To(BeNil())

		httpRoute := *endpoint.ExternalAccess.HttpRoute
		Expect(httpRoute["apiVersion"]).To(Equal("gateway.networking.k8s.io/v1"))
		Expect(httpRoute["kind"]).To(Equal("HTTPRoute"))

		spec := httpRoute["spec"].(map[string]interface{})
		parentRefs := spec["parentRefs"].([]map[string]interface{})
		Expect(parentRefs).To(HaveLen(1))
		Expect(parentRefs[0]["name"]).To(Equal("my-gateway"))
		Expect(parentRefs[0]["namespace"]).To(Equal("gateway-ns"))

		hostnames := spec["hostnames"].([]string)
		Expect(hostnames).To(Equal([]string{"my-nim.example.com"}))

		rules := spec["rules"].([]map[string]interface{})
		Expect(rules).To(HaveLen(1))
		backendRefs := rules[0]["backendRefs"].([]map[string]interface{})
		Expect(backendRefs[0]["name"]).To(Equal(workloadName))
		Expect(backendRefs[0]["port"]).To(BeNumerically("==", 8000))
	})

	It("sets HTTPRoute with path prefix and URLRewrite filter when URL has non-root path", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(true),
			ExposedUrl:       convutils.Ref("https://my-nim.example.com/api/v1"),
		}
		cfg := networkingConfig(sp, nimServiceNetworkingInfo{
			UseGatewayAPI:    true,
			GatewayName:      "my-gateway",
			GatewayNamespace: "gateway-ns",
		})

		endpoint := singleNetworkingEndpoint(cfg)
		Expect(endpoint.ExternalAccess.HttpRoute).NotTo(BeNil())

		httpRoute := *endpoint.ExternalAccess.HttpRoute
		spec := httpRoute["spec"].(map[string]interface{})
		rules := spec["rules"].([]map[string]interface{})
		matches := rules[0]["matches"].([]map[string]interface{})
		pathMatch := matches[0]["path"].(map[string]interface{})
		Expect(pathMatch["type"]).To(Equal("PathPrefix"))
		Expect(pathMatch["value"]).To(Equal("/api/v1"))

		filters := rules[0]["filters"].([]map[string]interface{})
		Expect(filters).To(HaveLen(1))
		Expect(filters[0]["type"]).To(Equal("URLRewrite"))
		urlRewrite := filters[0]["urlRewrite"].(map[string]interface{})
		rewritePath := urlRewrite["path"].(map[string]interface{})
		Expect(rewritePath["type"]).To(Equal("ReplacePrefixMatch"))
		Expect(rewritePath["replacePrefixMatch"]).To(Equal("/"))
	})

	It("sets Route and HTTPRoute without Ingress when OpenShift Route, Gateway API, and IngressClass are all enabled", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(true),
			ExposedUrl:       convutils.Ref("https://my-nim.example.com"),
		}
		cfg := networkingConfig(sp, nimServiceNetworkingInfo{
			IngressClass:      "nginx",
			UseOpenShiftRoute: true,
			UseGatewayAPI:     true,
			GatewayName:       "my-gateway",
			GatewayNamespace:  "gateway-ns",
		})

		endpoint := singleNetworkingEndpoint(cfg)
		Expect(endpoint.ExternalAccess.Route).NotTo(BeNil())
		Expect(endpoint.ExternalAccess.HttpRoute).NotTo(BeNil())
		Expect(endpoint.ExternalAccess.Ingress).To(BeNil())
		routeSpec := (*endpoint.ExternalAccess.Route)["spec"].(map[string]interface{})
		Expect(routeSpec["to"]).To(Equal(map[string]interface{}{"kind": "Service", "name": workloadName}))
	})

	It("returns an error when ingress is disabled and Gateway API is not enabled", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(true),
			ExposedUrl:       convutils.Ref("https://my-nim.example.com"),
		}
		cfg, err := nimServingPortToNetworkingConfig(workloadName, sp, nimServiceNetworkingInfo{
			IngressClass:      "nginx",
			IngressEnabled:    convutils.Ref(false),
			UseOpenShiftRoute: true,
		})

		Expect(cfg).To(BeNil())
		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(int32(503)))
		Expect(err.Message).To(Equal("cluster has no enabled external URL routing mechanism"))
	})

	It("sets only HTTPRoute when ingress is disabled and Gateway API is enabled", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(true),
			ExposedUrl:       convutils.Ref("https://my-nim.example.com"),
		}
		cfg := networkingConfig(sp, nimServiceNetworkingInfo{
			IngressClass:      "nginx",
			IngressEnabled:    convutils.Ref(false),
			UseOpenShiftRoute: true,
			UseGatewayAPI:     true,
			GatewayName:       "my-gateway",
			GatewayNamespace:  "gateway-ns",
		})

		endpoint := singleNetworkingEndpoint(cfg)
		Expect(endpoint.ExternalAccess).NotTo(BeNil())
		Expect(endpoint.ExternalAccess.Ingress).To(BeNil())
		Expect(endpoint.ExternalAccess.Route).To(BeNil())
		Expect(endpoint.ExternalAccess.HttpRoute).NotTo(BeNil())
	})

	It("returns an error when Gateway API is enabled but gateway name is missing", func() {
		sp := &api.NimServiceServingPort{
			Port:             convutils.Ref(int32(8000)),
			ExposeExternally: convutils.Ref(true),
			ExposedUrl:       convutils.Ref("https://my-nim.example.com"),
		}
		cfg, err := nimServingPortToNetworkingConfig(workloadName, sp, nimServiceNetworkingInfo{
			IngressEnabled: convutils.Ref(false),
			UseGatewayAPI:  true,
		})

		Expect(cfg).To(BeNil())
		Expect(err).NotTo(BeNil())
		Expect(err.Code).To(Equal(int32(503)))
		Expect(err.Message).To(Equal("cluster Gateway API is enabled but gateway name is not reported"))
	})
})

var _ = Describe("effectiveNimGatewayAPIEnabled", func() {
	It("uses the reported flag when present", func() {
		Expect(effectiveNimGatewayAPIEnabled(convutils.Ref(true), false, "", "")).To(BeTrue())
		Expect(effectiveNimGatewayAPIEnabled(convutils.Ref(false), true, "my-gateway", "gateway-ns")).To(BeFalse())
	})

	It("falls back to the dependency signal when the reported flag is absent", func() {
		Expect(effectiveNimGatewayAPIEnabled(nil, true, "my-gateway", "gateway-ns")).To(BeTrue())
		Expect(effectiveNimGatewayAPIEnabled(nil, false, "my-gateway", "gateway-ns")).To(BeFalse())
	})

	It("requires gateway name and namespace only for the older dependency fallback", func() {
		Expect(effectiveNimGatewayAPIEnabled(nil, true, "", "gateway-ns")).To(BeFalse())
		Expect(effectiveNimGatewayAPIEnabled(nil, true, "my-gateway", "")).To(BeFalse())
	})
})

var _ = Describe("buildHTTPRouteExternalAccess", func() {
	It("sets parentRefs with gateway name and namespace", func() {
		ea := buildHTTPRouteExternalAccess("my-nim.example.com", "/", "", "my-nim", 8000, "my-gw", "gw-ns")
		httpRoute := *ea.HttpRoute
		spec := httpRoute["spec"].(map[string]interface{})
		parentRefs := spec["parentRefs"].([]map[string]interface{})
		Expect(parentRefs[0]["name"]).To(Equal("my-gw"))
		Expect(parentRefs[0]["namespace"]).To(Equal("gw-ns"))
	})

	It("returns nil when gateway name is empty", func() {
		ea := buildHTTPRouteExternalAccess("my-nim.example.com", "/", "", "my-nim", 8000, "", "gw-ns")
		Expect(ea).To(BeNil())
	})

	It("omits parentRefs namespace when gateway namespace is empty", func() {
		ea := buildHTTPRouteExternalAccess("my-nim.example.com", "/", "", "my-nim", 8000, "my-gw", "")
		httpRoute := *ea.HttpRoute
		spec := httpRoute["spec"].(map[string]interface{})
		parentRefs := spec["parentRefs"].([]map[string]interface{})
		Expect(parentRefs[0]["name"]).To(Equal("my-gw"))
		Expect(parentRefs[0]).NotTo(HaveKey("namespace"))
	})

	It("sets hostname and backend service", func() {
		ea := buildHTTPRouteExternalAccess("my-nim.example.com", "/", "", "my-nim", 8000, "my-gw", "gw-ns")
		httpRoute := *ea.HttpRoute
		spec := httpRoute["spec"].(map[string]interface{})
		Expect(spec["hostnames"]).To(Equal([]string{"my-nim.example.com"}))
		rules := spec["rules"].([]map[string]interface{})
		backendRefs := rules[0]["backendRefs"].([]map[string]interface{})
		Expect(backendRefs[0]["name"]).To(Equal("my-nim"))
		Expect(backendRefs[0]["port"]).To(BeNumerically("==", 8000))
	})

	It("does not set URLRewrite filter for root path", func() {
		ea := buildHTTPRouteExternalAccess("my-nim.example.com", "/", "", "my-nim", 8000, "my-gw", "gw-ns")
		httpRoute := *ea.HttpRoute
		spec := httpRoute["spec"].(map[string]interface{})
		rules := spec["rules"].([]map[string]interface{})
		Expect(rules[0]).NotTo(HaveKey("filters"))
	})

	It("sets URLRewrite filter for non-root path", func() {
		ea := buildHTTPRouteExternalAccess("my-nim.example.com", "/api/v1", "", "my-nim", 8000, "my-gw", "gw-ns")
		httpRoute := *ea.HttpRoute
		spec := httpRoute["spec"].(map[string]interface{})
		rules := spec["rules"].([]map[string]interface{})
		filters := rules[0]["filters"].([]map[string]interface{})
		Expect(filters).To(HaveLen(1))
		Expect(filters[0]["type"]).To(Equal("URLRewrite"))
		urlRewrite := filters[0]["urlRewrite"].(map[string]interface{})
		rewritePath := urlRewrite["path"].(map[string]interface{})
		Expect(rewritePath["type"]).To(Equal("ReplacePrefixMatch"))
		Expect(rewritePath["replacePrefixMatch"]).To(Equal("/"))
	})
})

var _ = Describe("resolveExposureUrl", func() {
	It("falls back to autoGeneratedUrl when exposedUrl is not set", func() {
		sp := &api.NimServiceServingPort{}
		auto := "https://auto.example.com"
		Expect(resolveExposureUrl(sp, &auto)).To(Equal("https://auto.example.com"))
	})
})

var _ = Describe("resolveProtocol", func() {
	It("infers grpc from GrpcPort when ExposedProtocol is not set", func() {
		sp := &api.NimServiceServingPort{GrpcPort: convutils.Ref(int32(8001))}
		Expect(resolveProtocol(sp)).To(Equal(api.SERVINGPORTEXPOSEDPROTOCOL_GRPC))
	})
})

var _ = Describe("buildIngressExternalAccess", func() {
	It("sets host and backend service on the ingress rule", func() {
		ea := buildIngressExternalAccess("my-nim.example.com", "/", "", "nginx", "my-nim", 8000, false)
		spec := (*ea.Ingress)["spec"].(map[string]interface{})
		rules := spec["rules"].([]interface{})
		Expect(rules[0].(map[string]interface{})["host"]).To(Equal("my-nim.example.com"))
		paths := rules[0].(map[string]interface{})["http"].(map[string]interface{})["paths"].([]interface{})
		backend := paths[0].(map[string]interface{})["backend"].(map[string]interface{})
		svc := backend["service"].(map[string]interface{})
		Expect(svc["name"]).To(Equal("my-nim"))
		Expect(svc["port"].(map[string]interface{})["number"]).To(BeNumerically("==", 8000))
	})

	It("sets path rewrite annotations and regex path for non-root path", func() {
		ea := buildIngressExternalAccess("my-nim.example.com", "/api/v1", "", "nginx", "my-nim", 8000, false)
		meta := (*ea.Ingress)["metadata"].(map[string]interface{})
		annotations := meta["annotations"].(map[string]interface{})
		Expect(annotations).To(HaveKey(annotationNginxRewriteTarget))
		spec := (*ea.Ingress)["spec"].(map[string]interface{})
		paths := spec["rules"].([]interface{})[0].(map[string]interface{})["http"].(map[string]interface{})["paths"].([]interface{})
		Expect(paths[0].(map[string]interface{})["path"]).To(Equal("/api/v1(/|$)(.*)"))
	})

	It("sets nginx GRPC annotation for nginx class", func() {
		ea := buildIngressExternalAccess("my-nim.example.com", "/", api.SERVINGPORTEXPOSEDPROTOCOL_GRPC, "nginx", "my-nim", 8000, false)
		meta := (*ea.Ingress)["metadata"].(map[string]interface{})
		annotations := meta["annotations"].(map[string]interface{})
		Expect(annotations[annotationNginxBackendProtocol]).To(Equal(backendProtocolGRPCNginx))
	})

	It("sets haproxy GRPC annotation for haproxy class", func() {
		ea := buildIngressExternalAccess("my-nim.example.com", "/", api.SERVINGPORTEXPOSEDPROTOCOL_GRPC, ingressClassHAProxy, "my-nim", 8000, false)
		meta := (*ea.Ingress)["metadata"].(map[string]interface{})
		annotations := meta["annotations"].(map[string]interface{})
		Expect(annotations[annotationHAProxyBackendProtocol]).To(Equal(backendProtocolGRPCHAProxy))
	})

	It("sets haproxy request-set-header when path is non-root and hasAuth is true", func() {
		ea := buildIngressExternalAccess("my-nim.example.com", "/api/v1", "", ingressClassHAProxy, "my-nim", 8000, true)
		meta := (*ea.Ingress)["metadata"].(map[string]interface{})
		annotations := meta["annotations"].(map[string]interface{})
		Expect(annotations[annotationHAProxyRequestSetHeader]).To(Equal(haproxyForwardedUriHeader))
		Expect(annotations).NotTo(HaveKey("nginx.ingress.kubernetes.io/configuration-snippet"))
	})

	It("does not set auth annotations when path is root even with auth", func() {
		ea := buildIngressExternalAccess("my-nim.example.com", "/", "", "nginx", "my-nim", 8000, true)
		meta := (*ea.Ingress)["metadata"].(map[string]interface{})
		Expect(meta).NotTo(HaveKey("annotations"))
	})

	It("does not set auth annotations when hasAuth is false even with non-root path", func() {
		ea := buildIngressExternalAccess("my-nim.example.com", "/api/v1", "", "nginx", "my-nim", 8000, false)
		meta := (*ea.Ingress)["metadata"].(map[string]interface{})
		annotations := meta["annotations"].(map[string]interface{})
		Expect(annotations).NotTo(HaveKey("nginx.ingress.kubernetes.io/configuration-snippet"))
		Expect(annotations).NotTo(HaveKey(annotationHAProxyRequestSetHeader))
	})
})
