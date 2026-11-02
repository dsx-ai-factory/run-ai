/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import (
	"encoding/json"
	"strings"

	workloadsapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Networking endpoint inference transport", func() {
	const inference = `{"epp":{"profile":"gie-v1.5"},"rewrites":[{"model":"chat","targets":[{"model":"chat-v1","weight":1}]}]}`
	const stored = `{"networking":[{"services":[{"spec":{"name":"model","selector":{"app":"model"},"ports":[{"port":80,"targetPort":8000}]}}],` +
		`"externalAccess":{"httpRoute":{"spec":{"hostnames":["model.example.test"]}},"inference":` + inference + `}}]}`

	restore := func(raw string) *WorkloadConfiguration {
		var config WorkloadConfiguration
		Expect(json.Unmarshal([]byte(raw), &config)).To(Succeed())
		return &config
	}

	It("carries the stored inference block in the networking endpoints sent to the agent", func() {
		config := restore(stored)
		roundTrip, err := json.Marshal(config)
		Expect(err).NotTo(HaveOccurred())
		record := WorkloadsQueueRecord{Name: "model", Configuration: restore(string(roundTrip))}

		payload, err := json.Marshal(record.ToClusterUpdate().Configuration)
		Expect(err).NotTo(HaveOccurred())
		var agentConfig workloadsapi.WorkloadV2UpdateConfiguration
		Expect(json.Unmarshal(payload, &agentConfig)).To(Succeed())
		Expect(agentConfig.NetworkingEndpoints).To(HaveLen(1))
		access := agentConfig.NetworkingEndpoints[0].ExternalAccess.Get()
		Expect(access).NotTo(BeNil())
		Expect(access.HttpRoute).NotTo(BeNil())
		agentInference, err := json.Marshal(access.Inference.Get())
		Expect(err).NotTo(HaveOccurred())
		Expect(agentInference).To(MatchJSON(inference))

		var wire map[string]json.RawMessage
		Expect(json.Unmarshal(payload, &wire)).To(Succeed())
		Expect(wire).NotTo(HaveKey("inferenceEndpoint"))
	})

	It("does not send the inference block through the legacy single-endpoint networking field", func() {
		// The legacy shape can represent only a Service whose target port equals its port.
		record := WorkloadsQueueRecord{Name: "model", Configuration: restore(strings.Replace(stored, `,"targetPort":8000`, "", 1))}
		legacy := record.ToClusterUpdate().Configuration.Networking
		Expect(legacy).NotTo(BeNil())
		Expect(legacy.ExternalAccess).NotTo(BeNil())
		Expect(legacy.ExternalAccess.HttpRoute).NotTo(BeNil())
		Expect(legacy.ExternalAccess.Inference).To(BeNil())
	})

	It("omits the server-set inference block from the user-facing workload", func() {
		record := WorkloadsQueueRecord{Name: "model", Configuration: restore(stored)}
		networking := record.ToWorkloadV2().Metadata.Configuration.Networking
		Expect(networking).NotTo(BeNil())
		Expect((*networking)[0].ExternalAccess.HttpRoute).NotTo(BeNil())
		Expect((*networking)[0].ExternalAccess.Inference).To(BeNil())
		Expect(record.Configuration.Networking.ToAPI("model")).To(HaveValue(ContainElement(
			HaveField("ExternalAccess.Inference", Not(BeNil())),
		)), "the projection must not change the stored configuration")
	})

	It("isolates the stored inference block from changes to the cluster update", func() {
		config := restore(stored)
		record := WorkloadsQueueRecord{Name: "model", Configuration: config}
		update := record.ToClusterUpdate()
		sent := (*update.Configuration.NetworkingEndpoints)[0].ExternalAccess.Inference
		sent.Epp["profile"] = "changed"
		(*sent.Rewrites)[0]["model"] = "changed"
		raw, err := json.Marshal((*config.Networking)[0].ExternalAccess.Inference)
		Expect(err).NotTo(HaveOccurred())
		Expect(raw).To(MatchJSON(inference))
	})

	It("does not enable inference routing for stored configuration without it", func() {
		record := WorkloadsQueueRecord{Name: "model", Configuration: restore(`{"networking":[{"services":[{"ref":{"name":"model","port":80}}],"externalAccess":{"httpRoute":{}}}]}`)}
		payload, err := json.Marshal(record.ToClusterUpdate().Configuration)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(payload)).NotTo(ContainSubstring("inference"))
	})
})
