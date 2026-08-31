/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference_workloads

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	openapi "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
)

var _ = Describe("Inference engine metrics", func() {
	DescribeTable("detects the serving engine",
		func(labels []workloadsApi.Label, image string, expected inferenceEngine, expectConflict bool) {
			engine, err := detectInferenceEngine(labels, image)
			if expectConflict {
				Expect(err).NotTo(BeNil())
				Expect(err.Message).To(ContainSubstring("conflicting inference engine labels"))
				return
			}
			Expect(err).To(BeNil())
			Expect(engine).To(Equal(expected))
		},
		Entry("canonical NIM label", []workloadsApi.Label{engineLabel(constants.LabelNIMWorkload, "true", false)}, "custom/image", inferenceEngineNIM, false),
		Entry("legacy NIM label", []workloadsApi.Label{engineLabel(legacyNIMWorkloadLabel, "true", false)}, "custom/image", inferenceEngineNIM, false),
		Entry("vLLM label", []workloadsApi.Label{engineLabel(constants.LabelVLLMWorkload, "true", false)}, "custom/image", inferenceEngineVLLM, false),
		Entry("NIM registry image", nil, "nvcr.io/nim/meta/llama3:latest", inferenceEngineNIM, false),
		Entry("NIM repository image", nil, "nvcr.io/nvidia/nim:latest", inferenceEngineNIM, false),
		Entry("vLLM OpenAI image", nil, "vllm/vllm-openai:latest", inferenceEngineVLLM, false),
		Entry("qualified vLLM image", nil, "registry.example.com/team/vllm:1.0", inferenceEngineVLLM, false),
		Entry("explicit NIM wins over vLLM image", []workloadsApi.Label{engineLabel(constants.LabelNIMWorkload, "true", false)}, "vllm/vllm-openai:latest", inferenceEngineNIM, false),
		Entry("explicit vLLM wins over NIM image", []workloadsApi.Label{engineLabel(constants.LabelVLLMWorkload, "true", false)}, "nvcr.io/nim/meta/llama3:latest", inferenceEngineVLLM, false),
		Entry("false label falls back to image", []workloadsApi.Label{engineLabel(constants.LabelNIMWorkload, "false", false)}, "vllm/vllm-openai:latest", inferenceEngineVLLM, false),
		Entry("false vLLM label suppresses vLLM image fallback", []workloadsApi.Label{engineLabel(constants.LabelVLLMWorkload, "false", false)}, "vllm/vllm-openai:latest", inferenceEngineUnknown, false),
		Entry("false NIM label suppresses NIM image fallback", []workloadsApi.Label{engineLabel(constants.LabelNIMWorkload, "false", false)}, "nvcr.io/nim/meta/llama3:latest", inferenceEngineUnknown, false),
		Entry("excluded label falls back to image", []workloadsApi.Label{engineLabel(constants.LabelNIMWorkload, "true", true)}, "vllm/vllm-openai:latest", inferenceEngineVLLM, false),
		Entry("unknown custom image", nil, "registry.example.com/team/server:latest", inferenceEngineUnknown, false),
		Entry("conflicting explicit labels", []workloadsApi.Label{
			engineLabel(constants.LabelNIMWorkload, "true", false),
			engineLabel(constants.LabelVLLMWorkload, "true", false),
		}, "custom/image", inferenceEngineUnknown, true),
	)

	It("builds the exact private ClusterIP metrics Service shape", func() {
		servingPort := workloadsApi.NewServingPort()
		servingPort.SetContainer(9000)

		endpoint := buildMetricsEndpoint(inferenceEngineVLLM, "my-llm", servingPort)

		Expect(endpoint).NotTo(BeNil())
		Expect(endpoint.HasExternalAccess()).To(BeFalse())
		Expect(endpoint.Services).To(HaveLen(1))
		service := endpoint.Services[0]
		Expect(service.HasAuthorization()).To(BeFalse())
		Expect(service.HasRef()).To(BeFalse())
		spec := service.GetSpec()
		Expect(spec.Name).To(Equal("my-llm-metrics"))
		Expect(spec.GetType()).To(Equal(openapi.SERVICETYPE_CLUSTER_IP))
		Expect(spec.Selector).To(Equal(map[string]string{constants.LabelWorkloadName: "my-llm"}))
		Expect(*spec.Labels).To(Equal(map[string]string{
			constants.LabelWorkloadMetricsService:     "true",
			constants.LabelVLLMWorkload:               "true",
			constants.LabelVLLMWorkloadMetricsService: "true",
			constants.LabelWorkloadKind:               constants.KindInferenceWorkload,
		}))
		Expect(spec.Ports).To(HaveLen(1))
		Expect(spec.Ports[0].GetName()).To(Equal("metrics"))
		Expect(spec.Ports[0].Port).To(Equal(int32(9000)))
		Expect(spec.Ports[0].GetTargetPort()).To(Equal(int32(9000)))
		Expect(string(spec.Ports[0].GetProtocol())).To(Equal("TCP"))
	})

	It("falls back to port 8000", func() {
		endpoint := buildMetricsEndpoint(inferenceEngineNIM, "my-llm", nil)
		port := endpoint.Services[0].GetSpec().Ports[0]
		Expect(port.Port).To(Equal(int32(8000)))
		Expect(port.GetTargetPort()).To(Equal(int32(8000)))
	})
})

func engineLabel(name, value string, exclude bool) workloadsApi.Label {
	label := workloadsApi.NewLabel()
	label.SetName(name)
	label.SetValue(value)
	label.SetExclude(exclude)
	return *label
}
