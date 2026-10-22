/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"encoding/json"

	workloadsapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Inference endpoint configuration transport", func() {
	It("preserves resolved intent through stored JSON and the cluster update", func() {
		const intent = `{"schemaVersion":"v1","spec":{"workload":{"selector":{"matchLabels":{"app":"model"}},"port":8000},"routing":{"gatewayRef":{"name":"inference","namespace":"gateways"},"hostnames":["model.example.test"]},"epp":{"profile":"gie-v1.5"}}}`
		var config WorkloadConfiguration
		Expect(json.Unmarshal([]byte(`{"inferenceEndpoint":`+intent+`}`), &config)).To(Succeed())
		stored, err := json.Marshal(config)
		Expect(err).NotTo(HaveOccurred())
		Expect(stored).To(MatchJSON(`{"inferenceEndpoint":` + intent + `}`))
		var restored WorkloadConfiguration
		Expect(json.Unmarshal(stored, &restored)).To(Succeed())
		record := WorkloadsQueueRecord{Configuration: &restored}
		payload, err := json.Marshal(record.ToClusterUpdate().Configuration)
		Expect(err).NotTo(HaveOccurred())
		var wire map[string]json.RawMessage
		Expect(json.Unmarshal(payload, &wire)).To(Succeed())
		Expect(wire["inferenceEndpoint"]).To(MatchJSON(intent))
		var agentConfig workloadsapi.WorkloadV2UpdateConfiguration
		Expect(json.Unmarshal(payload, &agentConfig)).To(Succeed())
		Expect(agentConfig.InferenceEndpoint.Get()).NotTo(BeNil())
		agentPayload, err := json.Marshal(agentConfig.InferenceEndpoint.Get())
		Expect(err).NotTo(HaveOccurred())
		Expect(agentPayload).To(MatchJSON(intent))
	})

	It("isolates stored intent from changes to the cluster update", func() {
		var config WorkloadConfiguration
		Expect(json.Unmarshal([]byte(`{"inferenceEndpoint":{"schemaVersion":"v1","spec":{"routing":{"hostnames":["original.example.test"]}}}}`), &config)).To(Succeed())
		record := WorkloadsQueueRecord{Configuration: &config}
		update := record.ToClusterUpdate()
		update.Configuration.InferenceEndpoint.SchemaVersion = "changed"
		routing := update.Configuration.InferenceEndpoint.Spec["routing"].(map[string]interface{})
		routing["hostnames"].([]interface{})[0] = "changed.example.test"
		stored, err := json.Marshal(config)
		Expect(err).NotTo(HaveOccurred())
		Expect(stored).To(MatchJSON(`{"inferenceEndpoint":{"schemaVersion":"v1","spec":{"routing":{"hostnames":["original.example.test"]}}}}`))
	})

	It("does not enable inference routing for legacy stored configuration", func() {
		var config WorkloadConfiguration
		Expect(json.Unmarshal([]byte(`{"mnnvl":{"mode":"None"}}`), &config)).To(Succeed())
		record := WorkloadsQueueRecord{Configuration: &config}
		payload, err := json.Marshal(record.ToClusterUpdate().Configuration)
		Expect(err).NotTo(HaveOccurred())
		var wire map[string]json.RawMessage
		Expect(json.Unmarshal(payload, &wire)).To(Succeed())
		Expect(wire).NotTo(HaveKey("inferenceEndpoint"))
	})
})
