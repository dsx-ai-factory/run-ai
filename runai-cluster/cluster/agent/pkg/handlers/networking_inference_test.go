/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"encoding/json"

	workloadsManagerApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("Networking endpoint inference transport", func() {
	const httpRoute = `{"apiVersion":"gateway.networking.k8s.io/v1","kind":"HTTPRoute",
		"spec":{"parentRefs":[{"name":"runai-gateway","namespace":"runai"}],"hostnames":["llama-chat-team-a.runai.example.com"],
		"rules":[{"matches":[{"path":{"type":"PathPrefix","value":"/"}}],"backendRefs":[{"name":"llama-chat-leader","port":8000}]}]}}`
	const service = `{"spec":{"name":"llama-chat-leader",
		"selector":{"leaderworkerset.sigs.k8s.io/name":"llama-chat","leaderworkerset.sigs.k8s.io/worker-index":"0"},
		"ports":[{"port":8000,"targetPort":8000}]}}`

	wrap := func(inference string) (*endpoints.Config, error) {
		var config workloadsManagerApi.WorkloadV2UpdateConfiguration
		Expect(json.Unmarshal([]byte(`{"networkingEndpoints":[{"services":[`+service+`],`+
			`"externalAccess":{"httpRoute":`+httpRoute+`,"inference":`+inference+`}}]}`), &config)).To(Succeed())
		manifest := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "apps/v1", "kind": "Deployment"}}
		runner, err := wrapWithWorkloadRunner(manifest, "workload-id", &config)
		if err != nil {
			return nil, err
		}
		Expect(manifest.GetAnnotations()).NotTo(HaveKey(endpoints.Annotation))
		return endpoints.Get(runner)
	}

	It("places the inference block next to httpRoute in the run.ai/endpoints annotation", func() {
		const inference = `{"epp":{"profile":"gie-v1.5"},"objectives":[{"model":"llama-chat","priority":5}],
			"rewrites":[{"model":"chat","targets":[{"model":"meta-llama/Llama-3.1-70B-Instruct","weight":1}]}]}`
		cfg, err := wrap(inference)
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg.InferenceEndpointIndex()).To(Equal(0))
		access := cfg.Endpoints[0].ExternalAccess
		Expect(access.HTTPRoute).NotTo(BeNil())
		raw, err := json.Marshal(access.Inference)
		Expect(err).NotTo(HaveOccurred())
		Expect(raw).To(MatchJSON(inference))
	})

	It("carries an epp-only inference block without objectives or rewrites", func() {
		cfg, err := wrap(`{"epp":{"profile":"gie-v1.5"}}`)
		Expect(err).NotTo(HaveOccurred())
		raw, err := json.Marshal(cfg.Endpoints[0].ExternalAccess.Inference)
		Expect(err).NotTo(HaveOccurred())
		Expect(raw).To(MatchJSON(`{"epp":{"profile":"gie-v1.5"}}`))
	})

	It("rejects an inference block that cannot be applied before creating a runner", func() {
		_, err := wrap(`{"epp":{"profile":"gie-v1.5","image":null}}`)
		Expect(err).To(MatchError(ContainSubstring("must omit unset fields")))
	})

	It("does not set inference for endpoints without it", func() {
		var config workloadsManagerApi.WorkloadV2UpdateConfiguration
		Expect(json.Unmarshal([]byte(`{"networkingEndpoints":[{"services":[`+service+`],"externalAccess":{"httpRoute":`+httpRoute+`}}]}`), &config)).To(Succeed())
		manifest := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "apps/v1", "kind": "Deployment"}}
		runner, err := wrapWithWorkloadRunner(manifest, "workload-id", &config)
		Expect(err).NotTo(HaveOccurred())
		Expect(endpoints.HasInference(runner)).To(BeFalse())
	})

	It("does not enable inference routing from an annotation in the workload manifest", func() {
		manifest := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "apps/v1", "kind": "Deployment"}}
		manifest.SetAnnotations(map[string]string{endpoints.Annotation: `{"schemaVersion":"v1","endpoints":[{"services":[{"ref":{"name":"x","port":1}}],` +
			`"externalAccess":{"httpRoute":{},"inference":{"epp":{"profile":"gie-v1.5"}}}}]}`})
		runner, err := wrapWithWorkloadRunner(manifest, "workload-id", nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(runner.Annotations).NotTo(HaveKey(endpoints.Annotation))
	})
})
