/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inferenceendpoint

import (
	"testing"

	clustersapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

func TestInferenceEndpoint(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Inference endpoint defaults Suite")
}

var _ = Describe("Administrator inference endpoint defaults", func() {
	var status *clustersapi.ClusterReportedStatus
	BeforeEach(func() {
		status = &clustersapi.ClusterReportedStatus{
			Config: *clustersapi.NewNullableClusterReportedConfig(&clustersapi.ClusterReportedConfig{
				InferenceEndpointSchemaVersion: *clustersapi.NewNullableString(ptr.To("v1")),
				GatewayAPIEnabled:              *clustersapi.NewNullableBool(ptr.To(true)),
				GatewayName:                    *clustersapi.NewNullableString(ptr.To("inference")),
				GatewayNamespace:               *clustersapi.NewNullableString(ptr.To("gateways")),
			}),
			Operands: &map[string]clustersapi.ClusterReportedStatusOperandsValue{
				"runai-agent": {Ready: true}, "anyworkload-controller": {Ready: true}, "inference-endpoint-controller": {Ready: true},
			},
		}
	})

	It("resolves the administrator Gateway and the approved EPP profile", func() {
		defaults, err := ResolveDefaults(status)
		Expect(err).NotTo(HaveOccurred())
		Expect(defaults).To(Equal(&Defaults{GatewayName: "inference", GatewayNamespace: "gateways", EPPProfile: "gie-v1.5"}))
	})

	It("leaves clusters that do not advertise the new contract unchanged", func() {
		status.Config.Get().InferenceEndpointSchemaVersion.Unset()
		defaults, err := ResolveDefaults(status)
		Expect(err).NotTo(HaveOccurred())
		Expect(defaults).To(BeNil())
		defaults, err = ResolveDefaults(nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(defaults).To(BeNil())
	})

	It("rejects a future contract instead of interpreting it as v1", func() {
		status.Config.Get().InferenceEndpointSchemaVersion.Set(ptr.To("v2"))
		defaults, err := ResolveDefaults(status)
		Expect(err).To(MatchError(ContainSubstring("unsupported inference endpoint integration")))
		Expect(defaults).To(BeNil())
	})

	DescribeTable("rejects a participating operand that is not ready", func(operand string) {
		(*status.Operands)[operand] = clustersapi.ClusterReportedStatusOperandsValue{Ready: false}
		defaults, err := ResolveDefaults(status)
		Expect(err).To(MatchError(ContainSubstring(operand)))
		Expect(defaults).To(BeNil())
	}, Entry("Agent", "runai-agent"), Entry("Anyworkload", "anyworkload-controller"), Entry("endpoint controller", "inference-endpoint-controller"))

	It("rejects missing operand evidence during an upgrade", func() {
		status.Operands = nil
		defaults, err := ResolveDefaults(status)
		Expect(err).To(HaveOccurred())
		Expect(defaults).To(BeNil())
	})

	DescribeTable("requires explicit complete Gateway defaults", func(mutate func(*clustersapi.ClusterReportedConfig)) {
		mutate(status.Config.Get())
		defaults, err := ResolveDefaults(status)
		Expect(err).To(MatchError(ContainSubstring("Gateway")))
		Expect(defaults).To(BeNil())
	},
		Entry("disabled", func(c *clustersapi.ClusterReportedConfig) { c.GatewayAPIEnabled.Set(ptr.To(false)) }),
		Entry("enablement missing", func(c *clustersapi.ClusterReportedConfig) { c.GatewayAPIEnabled.Unset() }),
		Entry("name missing", func(c *clustersapi.ClusterReportedConfig) { c.GatewayName.Unset() }),
		Entry("namespace missing", func(c *clustersapi.ClusterReportedConfig) { c.GatewayNamespace.Unset() }),
		Entry("invalid name", func(c *clustersapi.ClusterReportedConfig) { c.GatewayName.Set(ptr.To("not a name")) }),
	)
})
