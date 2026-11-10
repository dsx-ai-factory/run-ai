/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package system_mapping_test

import (
	"context"
	"encoding/json"

	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Run serially: runLayer3 temporarily replaces the global broken-rule hook.
var _ = Describe("DynamoGraphDeployment/v1beta1 policies", Serial, func() {
	cfg := dynamoGraphDeploymentV1beta1Config

	It("keeps explicit values when applying defaults", func() {
		skeleton := loadReferenceSpecGinkgo(cfg.inputDir + "/skeleton.yaml")
		components := skeleton["spec"].(map[string]any)["components"].([]any)
		first := components[0].(map[string]any)
		first["replicas"] = 0
		first["globalDynamoNamespace"] = false
		secondPod := components[1].(map[string]any)["podTemplate"].(map[string]any)["spec"].(map[string]any)
		secondPod["serviceAccountName"] = "custom-sa"

		rules, err := loadLayer3PolicyFile(cfg.inputDir + "/cases/defaults/policy.yaml")
		Expect(err).NotTo(HaveOccurred())
		policy, err := composeLayer3Policy(cfg.gvk, rules)
		Expect(err).NotTo(HaveOccurred())
		result := runLayer3(context.Background(), policy, skeleton, len(rules)+10)
		Expect(result.Errors).To(BeEmpty())
		Expect(result.BrokenRules).To(BeEmpty())
		rendered := result.Workload["spec"].(map[string]any)["components"].([]any)
		Expect(rendered[0].(map[string]any)).To(HaveKeyWithValue("replicas", json.Number("0")))
		Expect(rendered[0].(map[string]any)).To(HaveKeyWithValue("globalDynamoNamespace", false))
		pod := rendered[1].(map[string]any)["podTemplate"].(map[string]any)["spec"].(map[string]any)
		Expect(pod).To(HaveKeyWithValue("serviceAccountName", "custom-sa"))
	})

	for _, selector := range []string{"allPodSpecs", "componentPodSpec"} {
		for _, templateState := range []string{"omitted", "null", "metadata-only"} {
			It(selector+" rejects pod policies for a "+templateState+" template", func() {
				skeleton := loadReferenceSpecGinkgo(cfg.inputDir + "/skeleton.yaml")
				components := skeleton["spec"].(map[string]any)["components"].([]any)
				for _, component := range components[1:] {
					c := component.(map[string]any)
					switch templateState {
					case "omitted":
						delete(c, "podTemplate")
					case "null":
						c["podTemplate"] = nil
					case "metadata-only":
						c["podTemplate"] = map[string]any{"metadata": map[string]any{"labels": map[string]any{"existing": "label"}}}
					}
				}
				schema, err := loadOpenAPISchema(cfg.schemaPath, cfg.schemaName)
				Expect(err).NotTo(HaveOccurred())
				Expect(formatOpenAPIErrors(validateAgainstOpenAPI(schema, stripNilLeavesMap(skeleton)))).To(BeEmpty())
				rules, err := loadLayer3PolicyFile(cfg.inputDir + "/cases/parameterized-dependency/policy.yaml")
				Expect(err).NotTo(HaveOccurred())
				rules = rules[:1]
				rules[0].SpecSelector = selector
				if selector == "allPodSpecs" {
					rules[0].SelectorParams = nil
				}
				policy, err := composeLayer3Policy(cfg.gvk, rules)
				Expect(err).NotTo(HaveOccurred())
				expectedTarget := deepCopyMap(components[2].(map[string]any))
				result := runLayer3(context.Background(), policy, skeleton, 10)
				Expect(result.Errors).NotTo(BeEmpty())
				Expect(result.Errors[0].Message).To(ContainSubstring("Dynamo pod policies require an explicit podTemplate.spec"))
				Expect(result.BrokenRules).To(BeEmpty())
				Expect(formatOpenAPIErrors(validateAgainstOpenAPI(schema, stripNilLeavesMap(result.Workload)))).To(BeEmpty())
				rendered := result.Workload["spec"].(map[string]any)["components"].([]any)
				Expect(rendered[2]).To(Equal(expectedTarget))
			})
		}
	}

	for _, selector := range []string{"allContainers", "allAppContainers", "allMainContainers", "componentContainers", "componentAppContainers", "componentMainContainer", "appContainerByName", "componentAppContainerByName"} {
		for _, templateState := range []string{"omitted", "sidecar-only"} {
			It(selector+" rejects container policies when main is "+templateState, func() {
				rules, err := loadLayer3PolicyFile(cfg.inputDir + "/cases/defaults/policy.yaml")
				Expect(err).NotTo(HaveOccurred())
				var imageRules []policyApi.EffectiveRulesEffectiveRulesInner
				for _, rule := range rules {
					if rule.Field == "image" {
						rule.SpecSelector = selector
						rule.SelectorParams = map[string]any{"componentName": "worker", "containerName": "main"}
						imageRules = append(imageRules, rule)
					}
				}
				Expect(imageRules).To(HaveLen(1))
				policy, err := composeLayer3Policy(cfg.gvk, imageRules)
				Expect(err).NotTo(HaveOccurred())
				component := map[string]any{"name": "worker", "type": "worker"}
				if templateState == "sidecar-only" {
					component["podTemplate"] = map[string]any{"spec": map[string]any{"containers": []any{map[string]any{"name": "sidecar", "image": "example/sidecar:test"}}}}
				}
				workload := map[string]any{"apiVersion": "nvidia.com/v1beta1", "kind": "DynamoGraphDeployment", "metadata": map[string]any{"name": "test"}, "spec": map[string]any{"components": []any{component}}}
				result := runLayer3(context.Background(), policy, workload, 10)
				Expect(result.Errors).NotTo(BeEmpty())
				Expect(result.Errors[0].Message).To(ContainSubstring("Dynamo container policies require an explicit main container"))
				Expect(result.BrokenRules).To(BeEmpty())
			})
		}
	}

	for _, caseName := range cfg.caseNames {
		It("applying the "+caseName+" policy twice gives the same result", func() {
			rules, err := loadLayer3PolicyFile(cfg.inputDir + "/cases/" + caseName + "/policy.yaml")
			Expect(err).NotTo(HaveOccurred())
			policy, err := composeLayer3Policy(cfg.gvk, rules)
			Expect(err).NotTo(HaveOccurred())
			skeleton := loadReferenceSpecGinkgo(cfg.inputDir + "/skeleton.yaml")
			first := runLayer3(context.Background(), policy, skeleton, len(rules)+10)
			Expect(first.Errors).To(BeEmpty())
			Expect(first.BrokenRules).To(BeEmpty())
			second := runLayer3(context.Background(), policy, first.Workload, len(rules)+10)
			Expect(second.Errors).To(BeEmpty())
			Expect(second.BrokenRules).To(BeEmpty())
			expected, err := yaml.Marshal(stripNilLeavesMap(first.Workload))
			Expect(err).NotTo(HaveOccurred())
			actual, err := yaml.Marshal(stripNilLeavesMap(second.Workload))
			Expect(err).NotTo(HaveOccurred())
			assert.Equal(GinkgoT(), string(expected), string(actual))
		})
	}
})
