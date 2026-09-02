/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	prom_common "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/common/prometheus"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var _ = Describe("NMC Prometheus HTTPRoute", func() {

	var runaiConfig *runv1.RunaiConfig

	BeforeEach(func() {
		runaiConfig = &runv1.RunaiConfig{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-config",
				Namespace: "test-namespace",
			},
			Spec: runv1.RunaiConfigSpec{
				Global: &global.Global{},
				Internal: &internal.Internal{
					Global: &internal.Global{},
				},
			},
		}
	})

	Context("Enabled", func() {
		It("should be disabled when gatewayAPI is nil", func() {
			runaiConfig.Spec.Global.GatewayAPI = nil
			r := NewHTTPRouteResource("nmc-mgmt-prometheus", nil, runaiConfig)
			Expect(r.Enabled()).To(BeFalse())
		})

		It("should be disabled when gatewayAPI.enabled is false", func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled: ptr.To(false),
			}
			r := NewHTTPRouteResource("nmc-mgmt-prometheus", nil, runaiConfig)
			Expect(r.Enabled()).To(BeFalse())
		})

		It("should be enabled when gatewayAPI.enabled is true", func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "my-gw",
				Namespace: "gw-ns",
			}
			r := NewHTTPRouteResource("nmc-mgmt-prometheus", nil, runaiConfig)
			Expect(r.Enabled()).To(BeTrue())
		})
	})

	Context("initNmcHTTPRoute", func() {
		It("should create HTTPRoute with prometheus domain and root path", func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "test-gateway",
				Namespace: "gateway-system",
			}
			route := &gwv1.HTTPRoute{}
			initNmcHTTPRoute(runaiConfig, route)

			Expect(route.Name).To(Equal(nmcHTTPRouteName))
			Expect(route.Spec.Hostnames).To(HaveLen(1))
			Expect(string(route.Spec.Hostnames[0])).To(Equal(prom_common.MgmtPrometheusDomain))
			Expect(route.Spec.ParentRefs).To(HaveLen(1))
			Expect(string(route.Spec.ParentRefs[0].Name)).To(Equal("test-gateway"))

			Expect(route.Spec.Rules).To(HaveLen(1))
			Expect(route.Spec.Rules[0].Matches).To(HaveLen(1))
			Expect(*route.Spec.Rules[0].Matches[0].Path.Value).To(Equal("/"))
		})

		It("should route to prometheus backend on port 9090", func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "test-gateway",
				Namespace: "gateway-system",
			}
			route := &gwv1.HTTPRoute{}
			initNmcHTTPRoute(runaiConfig, route)

			backendRefs := route.Spec.Rules[0].BackendRefs
			Expect(backendRefs).To(HaveLen(1))
			Expect(string(backendRefs[0].Name)).To(Equal(prom_common.MgmtPrometheusServiceName))
			Expect(int(*backendRefs[0].Port)).To(Equal(9090))
		})

		It("should include security header filter", func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "test-gateway",
				Namespace: "gateway-system",
			}
			route := &gwv1.HTTPRoute{}
			initNmcHTTPRoute(runaiConfig, route)

			Expect(route.Spec.Rules[0].Filters).To(HaveLen(1))
			Expect(route.Spec.Rules[0].Filters[0].Type).To(Equal(gwv1.HTTPRouteFilterResponseHeaderModifier))
		})

		It("should use per-operand gateway override when set", func() {
			runaiConfig.Spec.Global.GatewayAPI = &global.GatewayAPI{
				Enabled:   ptr.To(true),
				Name:      "global-gateway",
				Namespace: "global-ns",
			}
			runaiConfig.Spec.Internal.Global.NMC.MgmtPrometheusGatewayAPI = &common.GatewayRef{
				Name:      "override-gateway",
				Namespace: "override-ns",
			}
			route := &gwv1.HTTPRoute{}
			initNmcHTTPRoute(runaiConfig, route)

			Expect(route.Spec.ParentRefs).To(HaveLen(1))
			Expect(string(route.Spec.ParentRefs[0].Name)).To(Equal("override-gateway"))
			Expect(route.Spec.ParentRefs[0].Namespace).NotTo(BeNil())
			Expect(string(*route.Spec.ParentRefs[0].Namespace)).To(Equal("override-ns"))
		})
	})
})
