/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"encoding/json"

	workloadsapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("Inference endpoint Agent transport", func() {
	const intent = `{"schemaVersion":"v1","spec":{
		"workload":{"selector":{"matchLabels":{"app":"model"}},"port":8000},
		"routing":{"gatewayRef":{"name":"inference","namespace":"gateways"},"hostnames":["model.example.test"]},
		"epp":{"profile":"gie-v1.5"}}}`
	It("places the resolved configuration on the runner rather than its workload template", func() {
		var config workloadsapi.WorkloadV2UpdateConfiguration
		Expect(json.Unmarshal([]byte(`{"inferenceEndpoint":`+intent+`}`), &config)).To(Succeed())
		manifest := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "apps/v1", "kind": "Deployment"}}
		runner, err := wrapWithWorkloadRunner(manifest, "workload-id", &config)
		Expect(err).NotTo(HaveOccurred())
		Expect(runner.Annotations[endpoints.InferenceEndpointAnnotation]).To(MatchJSON(intent))
		Expect(manifest.GetAnnotations()).NotTo(HaveKey(endpoints.InferenceEndpointAnnotation))
		resolved, err := endpoints.GetInferenceEndpoint(runner)
		Expect(err).NotTo(HaveOccurred())
		Expect(resolved.SchemaVersion).To(Equal(endpoints.SchemaVersionV1))
	})

	It("does not enable inference routing from a user annotation in the workload manifest", func() {
		manifest := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "apps/v1", "kind": "Deployment"}}
		manifest.SetAnnotations(map[string]string{endpoints.InferenceEndpointAnnotation: intent})
		runner, err := wrapWithWorkloadRunner(manifest, "workload-id", nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(runner.Annotations).NotTo(HaveKey(endpoints.InferenceEndpointAnnotation))
	})

	DescribeTable("rejects unsupported or incomplete intent before applying a runner", func(intentJSON, expectedError string) {
		var config workloadsapi.WorkloadV2UpdateConfiguration
		Expect(json.Unmarshal([]byte(`{"inferenceEndpoint":`+intentJSON+`}`), &config)).To(Succeed())
		manifest := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "apps/v1", "kind": "Deployment"}}
		runner, err := wrapWithWorkloadRunner(manifest, "workload-id", &config)
		Expect(err).To(MatchError(ContainSubstring(expectedError)))
		Expect(runner).To(BeNil())
	},
		Entry("future version", `{"schemaVersion":"v2","spec":{}}`, "schemaVersion"),
		Entry("empty version", `{"schemaVersion":"","spec":{}}`, "schemaVersion"),
		Entry("missing required sections", `{"schemaVersion":"v1","spec":{}}`, "spec.workload"),
		Entry("explicit nested null", `{"schemaVersion":"v1","spec":{"workload":{"port":null}}}`, "must omit unset fields"),
	)

	It("updates and removes prior intent through the existing desired-state sync", func() {
		var config workloadsapi.WorkloadV2UpdateConfiguration
		Expect(json.Unmarshal([]byte(`{"inferenceEndpoint":`+intent+`}`), &config)).To(Succeed())
		manifest := &unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "apps/v1", "kind": "Deployment"}}
		runner, err := wrapWithWorkloadRunner(manifest, "workload-id", &config)
		Expect(err).NotTo(HaveOccurred())
		obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(runner)
		Expect(err).NotTo(HaveOccurred())
		desired := &unstructured.Unstructured{Object: obj}
		live := desired.DeepCopy()
		live.SetAnnotations(map[string]string{endpoints.InferenceEndpointAnnotation: "old intent"})
		UnstructuredObjectInspector{}.DeepCopySpec(desired, live)
		Expect(live.GetAnnotations()[endpoints.InferenceEndpointAnnotation]).To(MatchJSON(intent))

		runner, err = wrapWithWorkloadRunner(manifest, "workload-id", nil)
		Expect(err).NotTo(HaveOccurred())
		obj, err = runtime.DefaultUnstructuredConverter.ToUnstructured(runner)
		Expect(err).NotTo(HaveOccurred())
		UnstructuredObjectInspector{}.DeepCopySpec(&unstructured.Unstructured{Object: obj}, live)
		Expect(live.GetAnnotations()).NotTo(HaveKey(endpoints.InferenceEndpointAnnotation))
	})
})
