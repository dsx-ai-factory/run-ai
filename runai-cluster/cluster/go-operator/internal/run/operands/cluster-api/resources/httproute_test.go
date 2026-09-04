/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	"k8s.io/utils/ptr"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const testClusterURL = "https://cluster.example.com"

var _ = Describe("ClusterApi HTTPRoute", func() {

	var runaiConfig *runaiv1.RunaiConfig

	BeforeEach(func() {
		runaiConfig = utils.GetDefaultRunaiConfig()
	})

	Context("Enabled", func() {
		It("should be disabled when gatewayAPI is nil", func() {
			runaiConfig.Spec.Global.GatewayAPI = nil
			r := NewHTTPRouteResource("cluster-api", nil, runaiConfig)
			Expect(r.Enabled()).To(BeFalse())
		})

		It("should be disabled when gatewayAPI.name is empty", func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "",
				Namespace: "gw-ns",
			}
			runaiConfig.Spec.Internal.Global.ClusterURL = testClusterURL
			r := NewHTTPRouteResource("cluster-api", nil, runaiConfig)
			Expect(r.Enabled()).To(BeFalse())
		})

		It("should be enabled when gatewayAPI.enabled is true and clusterURL is set", func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "my-gw",
				Namespace: "gw-ns",
			}
			runaiConfig.Spec.Internal.Global.ClusterURL = testClusterURL
			r := NewHTTPRouteResource("cluster-api", nil, runaiConfig)
			Expect(r.Enabled()).To(BeTrue())
		})
	})

	Context("initClusterApiHTTPRoute", func() {
		It("should create HTTPRoute with /cluster-api path", func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "test-gateway",
				Namespace: "gateway-system",
			}
			runaiConfig.Spec.Internal.Global.ClusterURL = testClusterURL
			route := &gwv1.HTTPRoute{}
			Expect(initClusterApiHTTPRoute(runaiConfig, route)).To(Succeed())

			Expect(route.Name).To(Equal(clusterApiHTTPRouteName))
			Expect(route.Spec.Hostnames).To(HaveLen(1))
			Expect(string(route.Spec.Hostnames[0])).To(Equal("cluster.example.com"))
			Expect(route.Spec.ParentRefs).To(HaveLen(1))
			Expect(string(route.Spec.ParentRefs[0].Name)).To(Equal("test-gateway"))

			Expect(route.Spec.Rules).To(HaveLen(1))
			Expect(route.Spec.Rules[0].Matches).To(HaveLen(1))
			Expect(*route.Spec.Rules[0].Matches[0].Path.Value).To(Equal("/cluster-api"))
		})

		It("should route to cluster-api backend on correct port", func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "test-gateway",
				Namespace: "gateway-system",
			}
			runaiConfig.Spec.Internal.Global.ClusterURL = testClusterURL
			route := &gwv1.HTTPRoute{}
			Expect(initClusterApiHTTPRoute(runaiConfig, route)).To(Succeed())

			backendRefs := route.Spec.Rules[0].BackendRefs
			Expect(backendRefs).To(HaveLen(1))
			Expect(string(backendRefs[0].Name)).To(Equal(name))
			Expect(int(*backendRefs[0].Port)).To(Equal(int(serviceInPort)))
		})

		It("should include security header filter", func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "test-gateway",
				Namespace: "gateway-system",
			}
			runaiConfig.Spec.Internal.Global.ClusterURL = testClusterURL
			route := &gwv1.HTTPRoute{}
			Expect(initClusterApiHTTPRoute(runaiConfig, route)).To(Succeed())

			Expect(route.Spec.Rules[0].Filters).To(HaveLen(1))
			Expect(route.Spec.Rules[0].Filters[0].Type).To(Equal(gwv1.HTTPRouteFilterResponseHeaderModifier))
		})

		It("should use per-operand gateway override when set", func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "global-gateway",
				Namespace: "global-ns",
			}
			runaiConfig.Spec.ClusterApi.GatewayAPI = &common.GatewayRef{
				Name:      "override-gateway",
				Namespace: "override-ns",
			}
			runaiConfig.Spec.Internal.Global.ClusterURL = testClusterURL
			route := &gwv1.HTTPRoute{}
			Expect(initClusterApiHTTPRoute(runaiConfig, route)).To(Succeed())

			Expect(route.Spec.ParentRefs).To(HaveLen(1))
			Expect(string(route.Spec.ParentRefs[0].Name)).To(Equal("override-gateway"))
			Expect(route.Spec.ParentRefs[0].Namespace).NotTo(BeNil())
			Expect(string(*route.Spec.ParentRefs[0].Namespace)).To(Equal("override-ns"))
		})
	})
})
