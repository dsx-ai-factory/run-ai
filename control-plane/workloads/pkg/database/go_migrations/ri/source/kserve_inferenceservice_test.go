/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package source

import (
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	karta "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/karta/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var _ = Describe("KServe InferenceService status", func() {
	DescribeTable("maps Standard and RawDeployment status without Knative conditions",
		func(statusJSON string, expected karta.ResourceStatus) {
			for _, mode := range []string{"Standard", "RawDeployment"} {
				var status map[string]any
				Expect(json.Unmarshal([]byte(statusJSON), &status)).To(Succeed())
				status["deploymentMode"] = mode
				Expect(kserveMatchedStatuses(status)).To(ConsistOf(expected), "deployment mode %s", mode)
			}
		},
		Entry("ready without model status (custom container)",
			`{"conditions":[{"type":"Ready","status":"True"}]}`, karta.RunningStatus),
		Entry("ready after loading the model",
			`{"conditions":[{"type":"Ready","status":"True"}],"modelStatus":{"states":{"targetModelState":"Loaded"}}}`, karta.RunningStatus),
		Entry("pending during normal deployment startup",
			`{"conditions":[{"type":"Ready","status":"False","reason":"MinimumReplicasUnavailable"}],"modelStatus":{"states":{"targetModelState":"Pending"}}}`, karta.UndefinedStatus),
		Entry("loading the model",
			`{"conditions":[{"type":"Ready","status":"False"}],"modelStatus":{"states":{"targetModelState":"Loading"}}}`, karta.UndefinedStatus),
		Entry("model download failed",
			`{"conditions":[{"type":"Ready","status":"False"}],"modelStatus":{"states":{"targetModelState":"FailedToLoad"},"transitionStatus":"BlockedByFailedLoad"}}`, karta.FailedStatus),
		Entry("invalid runtime before readiness is known",
			`{"conditions":[{"type":"Ready","status":"Unknown"}],"modelStatus":{"transitionStatus":"InvalidSpec"}}`, karta.FailedStatus),
		Entry("failure before conditions are populated",
			`{"modelStatus":{"states":{"targetModelState":"FailedToLoad"}}}`, karta.FailedStatus),
		Entry("recovered despite retained failure information",
			`{"conditions":[{"type":"Ready","status":"True"}],"modelStatus":{"states":{"targetModelState":"Loaded"},"lastFailureInfo":{"reason":"RuntimeNotRecognized"}}}`, karta.RunningStatus),
		Entry("ready takes precedence over stale model failure state",
			`{"conditions":[{"type":"Ready","status":"True"}],"modelStatus":{"states":{"targetModelState":"FailedToLoad"},"transitionStatus":"InvalidSpec"}}`, karta.RunningStatus),
		Entry("historical failure does not fail a new load",
			`{"conditions":[{"type":"Ready","status":"False"}],"modelStatus":{"states":{"targetModelState":"Loading"},"lastFailureInfo":{"reason":"ModelLoadFailed"}}}`, karta.UndefinedStatus),
		Entry("stopped with a retained failure",
			`{"conditions":[{"type":"Ready","status":"False"},{"type":"Stopped","status":"True"}],"modelStatus":{"states":{"targetModelState":"FailedToLoad"}}}`, karta.UndefinedStatus),
		Entry("stopped with a retained ready condition",
			`{"conditions":[{"type":"Ready","status":"True"},{"type":"Stopped","status":"True"}]}`, karta.UndefinedStatus),
		Entry("resumed service is ready",
			`{"conditions":[{"type":"Ready","status":"True"},{"type":"Stopped","status":"False"}]}`, karta.RunningStatus),
		Entry("missing conditions and model status", `{}`, karta.UndefinedStatus),
	)

	DescribeTable("preserves Knative condition matching",
		func(conditionsJSON string, expected karta.ResourceStatus) {
			for _, mode := range []string{"", "Knative"} {
				var conditions []any
				Expect(json.Unmarshal([]byte(conditionsJSON), &conditions)).To(Succeed())
				status := map[string]any{"conditions": conditions}
				if mode != "" {
					status["deploymentMode"] = mode
				}
				Expect(kserveMatchedStatuses(status)).To(ConsistOf(expected), "deployment mode %q", mode)
			}
		},
		Entry("ready predictor, route, and deployment",
			`[{"type":"PredictorReady","status":"True"},{"type":"RoutesReady","status":"True"},{"type":"LatestDeploymentReady","status":"True"}]`, karta.RunningStatus),
		Entry("failed predictor, configuration, and route",
			`[{"type":"PredictorReady","status":"False"},{"type":"PredictorConfigurationReady","status":"False"},{"type":"RoutesReady","status":"False"}]`, karta.FailedStatus),
		Entry("incomplete conditions", `[{"type":"PredictorReady","status":"True"}]`, karta.UndefinedStatus),
	)

	DescribeTable("restricts the new matchers to Standard and RawDeployment",
		func(mode string) {
			for _, status := range []map[string]any{
				{"conditions": []any{map[string]any{"type": "Ready", "status": "True"}}},
				{"modelStatus": map[string]any{"transitionStatus": "InvalidSpec"}},
			} {
				if mode != "" {
					status["deploymentMode"] = mode
				}
				Expect(kserveMatchedStatuses(status)).To(ConsistOf(karta.UndefinedStatus))
			}
		},
		Entry("missing mode", ""),
		Entry("Knative", "Knative"),
		Entry("ModelMesh", "ModelMesh"),
	)

	It("leaves a newly created service without status undefined", func() {
		Expect(kserveMatchedStatuses(nil)).To(ConsistOf(karta.UndefinedStatus))
	})
})

func kserveMatchedStatuses(status map[string]any) []karta.ResourceStatus {
	GinkgoHelper()
	object := &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "serving.kserve.io/v1beta1",
		"kind":       "InferenceService",
		"metadata":   map[string]any{"name": "model", "namespace": "test"},
	}}
	if status != nil {
		object.Object["status"] = status
	}
	definition := karta.Karta(BuildKServeInferenceServiceV1beta1())
	root, err := resource.NewComponentFactoryFromObject(&definition, object).GetRootComponent()
	Expect(err).NotTo(HaveOccurred())
	result, err := root.GetStatus(context.Background())
	Expect(err).NotTo(HaveOccurred())
	Expect(result).NotTo(BeNil())
	return result.MatchedStatuses
}
