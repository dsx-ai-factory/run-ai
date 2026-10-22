/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"encoding/json"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
)

var _ = Describe("Inference endpoint capability transport", func() {
	It("preserves the operator-declared version in the control-plane report", func() {
		var config runv1.PublicConfiguration
		Expect(yaml.Unmarshal([]byte("inferenceEndpointSchemaVersion: v1\ngatewayName: inference\ngatewayNamespace: gateways\ngatewayAPIEnabled: true\n"), &config)).To(Succeed())
		report := parseClusterInfoConfigToServerScheme(&config)
		payload, err := json.Marshal(report)
		Expect(err).NotTo(HaveOccurred())
		var fields map[string]interface{}
		Expect(json.Unmarshal(payload, &fields)).To(Succeed())
		Expect(fields).To(HaveKeyWithValue("inferenceEndpointSchemaVersion", "v1"))
		Expect(fields).To(HaveKeyWithValue("gatewayName", "inference"))
		Expect(fields).To(HaveKeyWithValue("gatewayNamespace", "gateways"))
	})

	It("does not claim inference endpoint support for an older operator report", func() {
		report := parseClusterInfoConfigToServerScheme(&runv1.PublicConfiguration{})
		payload, err := json.Marshal(report)
		Expect(err).NotTo(HaveOccurred())
		var fields map[string]interface{}
		Expect(json.Unmarshal(payload, &fields)).To(Succeed())
		Expect(fields["inferenceEndpointSchemaVersion"]).To(BeNil())
	})
})
