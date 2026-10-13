/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GPU compute mode template compliance", func() {
	for _, replica := range []string{"", "Worker", "Master"} {
		DescribeTable("policy resolution for "+replica,
			func(mode, defaultMode, ruleJSON, wantRule string, hasGPU bool) {
				compute := map[string]interface{}{"gpuMemoryPortionRequest": 0.25}
				if !hasGPU {
					compute = map[string]interface{}{"cpuCoreRequest": 1.0}
				}
				if mode != "" {
					compute["gpuComputeMode"] = mode
				}
				values := map[string]interface{}{"compute": compute}
				specJSON, err := json.Marshal(values)
				Expect(err).NotTo(HaveOccurred())
				template := api.SupersetTemplate{Meta: api.TemplateMetaFields{Id: uuid.NewString()}}
				Expect(json.Unmarshal(specJSON, &template.CombinedSpec)).To(Succeed())
				if replica == "Master" {
					Expect(json.Unmarshal(specJSON, &template.CombinedMasterSpec)).To(Succeed())
				}
				before, err := json.Marshal(template)
				Expect(err).NotTo(HaveOccurred())
				var modeRules map[string]interface{}
				Expect(json.Unmarshal([]byte(ruleJSON), &modeRules)).To(Succeed())
				rules := map[string]interface{}{"compute": map[string]interface{}{"gpuComputeMode": modeRules}}
				defaults := map[string]interface{}{}
				if defaultMode != "" {
					defaults["compute"] = map[string]interface{}{"gpuComputeMode": defaultMode}
				}
				mutated, mutateErr := MutateWorkload(values, defaults, false, nil, nil)
				Expect(mutateErr).To(BeNil())
				submissionErr := validateSpecAgainstPolicy(Workspace, rules, defaults, mutated)
				Expect(submissionErr == nil).To(Equal(wantRule == ""))
				order, err := policy_compliance.NewDefaultsFromOrder(rules)
				Expect(err).NotTo(HaveOccurred())
				orderMap := map[string][]string(order)
				if replica != "" {
					rules = map[string]interface{}{strings.ToLower(replica): rules}
					defaults = map[string]interface{}{strings.ToLower(replica): defaults}
				}
				policyInfo := dao_crud.PolicyInfo{TenantId: 1, ProjectId: 1, WorkloadType: Workspace, ReplicaType: replica, PolicyTime: time.Now()}
				validateComplianceOfTemplate(context.Background(), patchingTaskInfo{
					template: &template, policyInfo: policyInfo,
					effectivePolicy: &policy_compliance.EffectivePolicyOfProject{
						RulesV2: rules, DefaultsV2: defaults,
						DefaultFromOrder: policyApi.DefaultFromOrder{Worker: &orderMap, Master: &orderMap},
					},
				})
				if wantRule == "" {
					Expect(template.ComplianceIssues).To(BeEmpty())
				} else {
					Expect(template.ComplianceIssues).To(HaveLen(1))
					Expect(template.ComplianceIssues[0].Field).To(Equal(convutils.Ref("compute.gpuComputeMode")))
					Expect(string(*template.ComplianceIssues[0].Rule)).To(Equal(wantRule))
				}
				saved, dbErr := dao_crud.ListPolicyComplianceOfTemplates(context.Background(), []string{template.Meta.Id}, policyInfo)
				Expect(dbErr).To(BeNil())
				Expect(saved).To(HaveLen(1))
				Expect(saved[0].Aligned).To(Equal(wantRule == ""))
				template.ComplianceIssues = nil
				after, err := json.Marshal(template)
				Expect(err).NotTo(HaveOccurred())
				Expect(after).To(Equal(before))
			},

			Entry("omitted mode with allowed time-slicing", "", "", `{"options":[{"value":"time-slicing"}]}`, "", true),
			Entry("omitted mode with allowed SM-sharing", "", "", `{"options":[{"value":"sm-sharing"}]}`, "", true),
			Entry("policy mode default", "", "sm-sharing", `{"options":[{"value":"sm-sharing"}]}`, "", true),
			Entry("editable explicit mode wins", "time-slicing", "sm-sharing", `{}`, "", true),
			Entry("locked policy default permits omission", "", "sm-sharing", `{"canEdit":false}`, "", true),
			Entry("locked equal mode passes", "sm-sharing", "sm-sharing", `{"canEdit":false}`, "", true),
			Entry("locked conflicting mode fails", "time-slicing", "sm-sharing", `{"canEdit":false}`, "canEdit", true),
			Entry("omitted required mode fails", "", "", `{"required":true}`, "required", true),
			Entry("policy default satisfies required", "", "sm-sharing", `{"required":true}`, "", true),
			Entry("omitted blocked mode passes", "", "", `{"blocked":true}`, "", true),
			Entry("explicit blocked mode fails", "sm-sharing", "", `{"blocked":true}`, "blocked", true),
			Entry("explicit mode outside options fails", "time-slicing", "", `{"options":[{"value":"sm-sharing"}]}`, "options", true),
			Entry("CPU-only omitted mode", "", "", `{"options":[{"value":"sm-sharing"}]}`, "", false),
			Entry("CPU-only required mode", "", "", `{"required":true}`, "required", false),
		)
	}
})

var _ = Describe("GPU allocation defaults in template compliance", func() {
	for _, replica := range []string{"", "Worker", "Master"} {
		DescribeTable("matches submission for "+replica,
			func(specJSON, defaultsJSON, rulesJSON, wantRule string) {
				values, err := gojsonq.LoadJsonBytes([]byte(specJSON))
				Expect(err).NotTo(HaveOccurred())
				computeDefaults, err := gojsonq.LoadJsonBytes([]byte(defaultsJSON))
				Expect(err).NotTo(HaveOccurred())
				computeRules, err := gojsonq.LoadJsonBytes([]byte(rulesJSON))
				Expect(err).NotTo(HaveOccurred())
				computeRules["gpuComputeMode"] = map[string]interface{}{"options": []interface{}{map[string]interface{}{"value": "sm-sharing"}}}
				defaults := map[string]interface{}{"compute": computeDefaults}
				rules := map[string]interface{}{"compute": computeRules}
				order, err := policy_compliance.NewDefaultsFromOrder(rules)
				Expect(err).NotTo(HaveOccurred())
				orderMap := map[string][]string(order)

				mutated, mutateErr := MutateWorkload(values, defaults, false, nil, nil)
				Expect(mutateErr).To(BeNil())
				submissionErr := validateSpecAgainstPolicy(Workspace, rules, defaults, mutated)
				Expect(submissionErr == nil).To(Equal(wantRule == ""), "submission control")
				template := api.SupersetTemplate{Meta: api.TemplateMetaFields{Id: uuid.NewString()}}
				Expect(json.Unmarshal([]byte(specJSON), &template.CombinedSpec)).To(Succeed())
				if replica == "Master" {
					Expect(json.Unmarshal([]byte(specJSON), &template.CombinedMasterSpec)).To(Succeed())
				}
				before, err := json.Marshal(template)
				Expect(err).NotTo(HaveOccurred())

				if replica != "" {
					rules = map[string]interface{}{strings.ToLower(replica): rules}
					defaults = map[string]interface{}{strings.ToLower(replica): defaults}
				}

				policyInfo := dao_crud.PolicyInfo{TenantId: 1, ProjectId: 1, WorkloadType: Workspace, ReplicaType: replica, PolicyTime: time.Now()}
				validateComplianceOfTemplate(context.Background(), patchingTaskInfo{
					template: &template, policyInfo: policyInfo,
					effectivePolicy: &policy_compliance.EffectivePolicyOfProject{
						RulesV2: rules, DefaultsV2: defaults,
						DefaultFromOrder: policyApi.DefaultFromOrder{Worker: &orderMap, Master: &orderMap},
					},
				})
				if wantRule == "" {
					Expect(template.ComplianceIssues).To(BeEmpty())
				} else {
					Expect(template.ComplianceIssues).To(HaveLen(1))
					Expect(template.ComplianceIssues[0].Field).To(Equal(convutils.Ref("compute.gpuComputeMode")))
					Expect(string(*template.ComplianceIssues[0].Rule)).To(Equal(wantRule))
				}
				saved, dbErr := dao_crud.ListPolicyComplianceOfTemplates(context.Background(), []string{template.Meta.Id}, policyInfo)
				Expect(dbErr).To(BeNil())
				Expect(saved).To(HaveLen(1))
				Expect(saved[0].Aligned).To(Equal(wantRule == ""))
				template.ComplianceIssues = nil
				after, err := json.Marshal(template)
				Expect(err).NotTo(HaveOccurred())
				Expect(after).To(Equal(before))
			},

			Entry("omitted compute with policy GPU", `{}`, `{"gpuMemoryPortionRequest":0.25}`, `{}`, ""),
			Entry("CPU-only compute with policy GPU", `{"compute":{"cpuCoreRequest":1}}`, `{"gpuMemoryPortionRequest":0.25}`, `{}`, ""),
			Entry("policy devices request", `{}`, `{"gpuDevicesRequest":2}`, `{}`, ""),
			Entry("policy absolute memory request", `{}`, `{"gpuMemoryRequest":"1Gi"}`, `{}`, ""),
			Entry("policy mode is permitted", `{}`, `{"gpuMemoryPortionRequest":0.25,"gpuComputeMode":"sm-sharing"}`, `{}`, ""),
			Entry("explicit forbidden mode still fails", `{"compute":{"gpuComputeMode":"time-slicing"}}`, `{"gpuMemoryPortionRequest":0.25}`, `{}`, "options"),
			Entry("explicit mode wins", `{"compute":{"gpuComputeMode":"sm-sharing"}}`, `{"gpuMemoryPortionRequest":0.25}`, `{}`, ""),
			Entry("explicit zero overrides GPU default", `{"compute":{"gpuMemoryPortionRequest":0}}`, `{"gpuMemoryPortionRequest":0.25}`, `{}`, ""),
			Entry("derived GPU does not supply a mode", `{"compute":{"cpuCoreRequest":1}}`, `{}`, `{"gpuMemoryPortionRequest":{"defaultFrom":{"field":"compute.cpuCoreRequest","factor":0.25}}}`, ""),
			Entry("ordinary zero precedes GPU derivation", `{"compute":{"cpuCoreRequest":1}}`, `{"gpuMemoryPortionRequest":0}`, `{"gpuMemoryPortionRequest":{"defaultFrom":{"field":"compute.cpuCoreRequest","factor":0.25}}}`, ""),
		)
	}
})
