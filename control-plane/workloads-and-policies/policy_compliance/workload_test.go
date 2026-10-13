/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy_compliance

import (
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("workload policy preparation", func() {
	It("resolves mixed-alias dependencies and locks without changing policy sources", func() {
		rules := map[string]interface{}{"compute": map[string]interface{}{
			"gpuPortionRequest": map[string]interface{}{"canEdit": false},
			"cpuCoreRequest":    map[string]interface{}{"defaultFrom": map[string]interface{}{"field": "compute.gpuPortionRequest", "factor": 8.0}},
			"cpuCoreLimit":      map[string]interface{}{"defaultFrom": map[string]interface{}{"field": "compute.cpuCoreRequest", "factor": 2.0}},
		}}
		defaults := map[string]interface{}{"compute": map[string]interface{}{"gpuPortionRequest": 0.25}}
		values := map[string]interface{}{"compute": map[string]interface{}{"gpuMemoryPortionRequest": 0.25}}
		normalizedDefaults, issues, err := ValidateWorkloadCompliance(rules, defaults, values, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(issues).To(BeEmpty())
		compute := convutils.MapMe(values["compute"])
		Expect(compute["cpuCoreRequest"]).To(BeNumerically("==", 2))
		Expect(compute["cpuCoreLimit"]).To(BeNumerically("==", 4))
		Expect(compute).NotTo(HaveKey("gpuComputeMode"))
		Expect(convutils.MapMe(normalizedDefaults["compute"])).To(HaveKeyWithValue("gpuMemoryPortionRequest", 0.25))
		Expect(convutils.MapMe(defaults["compute"])).To(HaveKey("gpuPortionRequest"))
		Expect(convutils.MapMe(rules["compute"])).To(HaveKey("gpuPortionRequest"))

		compute["gpuMemoryPortionRequest"] = 0.5
		_, issues, err = ValidateWorkloadCompliance(rules, defaults, values, false)
		Expect(err).NotTo(HaveOccurred())
		Expect(issues).To(HaveLen(1))
		Expect(issues[0].WorkloadField).To(Equal("compute.gpuMemoryPortionRequest"))
	})

	DescribeTable("rejects cycles revealed by alias normalization",
		func(rules map[string]interface{}) {
			_, _, err := ValidateWorkloadCompliance(map[string]interface{}{"compute": rules}, nil, map[string]interface{}{}, false)
			Expect(err).To(HaveOccurred())
		},
		Entry("self reference through the other alias", map[string]interface{}{
			"gpuPortionRequest": map[string]interface{}{"defaultFrom": map[string]interface{}{"field": "compute.gpuMemoryPortionRequest"}},
		}),
		Entry("cycle through a non-GPU field", map[string]interface{}{
			"gpuPortionRequest": map[string]interface{}{"defaultFrom": map[string]interface{}{"field": "compute.cpuCoreRequest"}},
			"cpuCoreRequest":    map[string]interface{}{"defaultFrom": map[string]interface{}{"field": "compute.gpuMemoryPortionRequest"}},
		}),
	)

	It("keeps worker and master defaults and error paths separate", func() {
		rules := map[string]interface{}{}
		defaults := map[string]interface{}{}
		values := map[string]interface{}{}
		for replica, portion := range map[string]float64{"worker": 0.25, "master": 0.5} {
			rules[replica] = map[string]interface{}{"compute": map[string]interface{}{
				"gpuPortionRequest": map[string]interface{}{"canEdit": false},
				"gpuComputeMode":    map[string]interface{}{"options": []interface{}{map[string]interface{}{"value": "sm-sharing"}}},
			}}
			defaults[replica] = map[string]interface{}{"compute": map[string]interface{}{"gpuPortionRequest": portion}}
			values[replica] = map[string]interface{}{"compute": map[string]interface{}{"gpuMemoryPortionRequest": portion, "gpuComputeMode": "time-slicing"}}
		}
		_, issues, err := ValidateWorkloadCompliance(rules, defaults, values, true)
		Expect(err).NotTo(HaveOccurred())
		Expect(issues).To(HaveLen(2))
		Expect(issues[0].WorkloadField).To(Equal("compute.gpuComputeMode"))
		Expect(issues[1].WorkloadField).To(Equal("masterSpec.compute.gpuComputeMode"))
		Expect(issues[1].PolicyField).To(Equal("masterSpec.compute.gpuComputeMode"))
		for _, replica := range []string{"worker", "master"} {
			Expect(convutils.MapMe(convutils.MapMe(values[replica])["compute"])).To(HaveKeyWithValue("gpuComputeMode", "time-slicing"))
		}
	})
})

var _ = Describe("GPU mode uses ordinary optional enum policy rules", func() {
	DescribeTable("matches imagePullPolicy",
		func(value, defaultValue, rule, wantIssue string) {
			for _, field := range []struct {
				name, allowed, other string
			}{
				{"gpuComputeMode", "sm-sharing", "time-slicing"},
				{"imagePullPolicy", "Always", "IfNotPresent"},
			} {
				values := map[string]interface{}{"compute": map[string]interface{}{"gpuDevicesRequest": 1}}
				defaults := map[string]interface{}{}
				rules := map[string]interface{}{}
				fieldValues, fieldDefaults, fieldRules := values, defaults, rules
				if field.name == "gpuComputeMode" {
					fieldValues = convutils.MapMe(values["compute"])
					fieldDefaults, fieldRules = map[string]interface{}{}, map[string]interface{}{}
					defaults["compute"], rules["compute"] = fieldDefaults, fieldRules
				}
				if defaultValue != "" {
					fieldDefaults[field.name] = field.allowed
				}
				if value != "" {
					fieldValues[field.name] = field.allowed
					if value == "other" {
						fieldValues[field.name] = field.other
					}
				}
				constraint := map[string]interface{}{}
				switch rule {
				case "options":
					constraint[rule] = []interface{}{map[string]interface{}{"value": field.allowed}}
				case "canEdit":
					constraint[rule] = false
				case "required", "blocked":
					constraint[rule] = true
				}
				fieldRules[field.name] = constraint
				merged := map[string]interface{}{}
				Expect(convutils.CopySpecs(&merged, defaults)).To(Succeed())
				Expect(ApplyDefaults(values, merged, nil)).To(BeNil())
				values = merged
				_, issues, err := ValidateWorkloadCompliance(rules, defaults, values, false)
				Expect(err).NotTo(HaveOccurred())
				if wantIssue == "" {
					Expect(issues).To(BeEmpty(), field.name)
				} else {
					Expect(issues).To(HaveLen(1), field.name)
					Expect(string(*issues[0].Rule)).To(Equal(wantIssue), field.name)
				}
				resolved := values
				if field.name == "gpuComputeMode" {
					resolved = convutils.MapMe(values["compute"])
				}
				if value == "" && defaultValue == "" {
					Expect(resolved).NotTo(HaveKey(field.name), field.name)
				} else {
					Expect(resolved).To(HaveKey(field.name), field.name)
				}
			}
		},
		Entry("optional omission", "", "", "", ""),
		Entry("options do not require a value", "", "", "options", ""),
		Entry("required omission", "", "", "required", "required"),
		Entry("blocked omission", "", "", "blocked", ""),
		Entry("allowed explicit value", "allowed", "", "options", ""),
		Entry("forbidden explicit value", "other", "", "options", "options"),
		Entry("blocked explicit value", "allowed", "", "blocked", "blocked"),
		Entry("policy default satisfies required", "", "allowed", "required", ""),
		Entry("policy default is validated", "", "allowed", "blocked", "blocked"),
		Entry("locked default fills omission", "", "allowed", "canEdit", ""),
		Entry("locked matching value", "allowed", "allowed", "canEdit", ""),
		Entry("locked conflicting value", "other", "allowed", "canEdit", "canEdit"),
	)
})
