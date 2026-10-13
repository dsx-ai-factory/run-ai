/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flat_workloads

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/services/response_patching"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/policy_compliance"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GPU compute mode asset compliance", func() {
	for _, replica := range []string{"", "Worker", "Master", "Leader"} {
		DescribeTable("policy resolution for "+replica,
			func(mode, defaultMode, ruleJSON, wantRule string, hasGPU bool) {
				spec := map[string]interface{}{"gpuMemoryPortionRequest": 0.25}
				if !hasGPU {
					spec = map[string]interface{}{"cpuCoreRequest": 1.0}
				}
				if mode != "" {
					spec["gpuComputeMode"] = mode
				}
				before, err := json.Marshal(spec)
				Expect(err).NotTo(HaveOccurred())
				var modeRules map[string]interface{}
				Expect(json.Unmarshal([]byte(ruleJSON), &modeRules)).To(Succeed())
				rules := map[string]interface{}{"compute": map[string]interface{}{"gpuComputeMode": modeRules}}
				defaults := map[string]interface{}{}
				if defaultMode != "" {
					defaults["compute"] = map[string]interface{}{"gpuComputeMode": defaultMode}
				}
				order, err := policy_compliance.NewDefaultsFromOrder(rules)
				Expect(err).NotTo(HaveOccurred())
				orderMap := map[string][]string(order)
				if replica != "" {
					rules = map[string]interface{}{strings.ToLower(replica): rules}
					defaults = map[string]interface{}{strings.ToLower(replica): defaults}
				}
				meta := api.AssetMeta{Id: uuid.NewString(), Name: "mode-compliance", Kind: api.ASSETKIND_COMPUTE}
				compliance := &api.ComplianceInfo{Compliance: true, Imposed: true}
				policyInfo := dao.PolicyInfo{TenantId: 1, ProjectId: 1, WorkloadType: auth.WorkloadTypeWorkspace, ReplicaType: replica, PolicyTime: time.Now()}
				info := assetChanInfo{
					patchingPtrs: response_patching.ResponsePatchingPtrs{Meta: &meta, Spec: &spec, ComplianceInfo: &compliance},
					rules:        rules, defaults: defaults, replicaType: replica,
					defaultFromOrder: policyApi.DefaultFromOrder{Worker: &orderMap, Master: &orderMap},
					workloadType:     auth.WorkloadTypeWorkspace, policyInfo: policyInfo, isImposed: true,
				}
				validateComplianceOfAsset(context.Background(), info)
				Expect(compliance.Compliance).To(Equal(wantRule == ""))
				Expect(compliance.Imposed).To(Equal(wantRule == ""))
				if wantRule != "" {
					Expect(compliance.Reason).To(HaveLen(1))
					Expect(compliance.Reason[0].Field).To(Equal(convutils.Ref("compute.gpuComputeMode")))
					Expect(string(*compliance.Reason[0].Rule)).To(Equal(wantRule))
				}
				after, err := json.Marshal(spec)
				Expect(err).NotTo(HaveOccurred())
				Expect(after).To(Equal(before))
				saved, dbErr := dao.ListPolicyComplianceOfAssets(context.Background(), []string{meta.Id}, policyInfo)
				Expect(dbErr).To(BeNil())
				Expect(saved).To(HaveLen(1))
				Expect(saved[0].Aligned).To(Equal(wantRule == ""))
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

	It("replaces cached compliance when memory aliases conflict", func() {
		spec := map[string]interface{}{"gpuPortionRequest": 0.25, "gpuMemoryPortionRequest": 0.5}
		meta := api.AssetMeta{Id: uuid.NewString(), Name: "conflicting-memory", Kind: api.ASSETKIND_COMPUTE}
		compliance := &api.ComplianceInfo{Compliance: true, Imposed: true}
		policyInfo := dao.PolicyInfo{TenantId: 1, ProjectId: 1, WorkloadType: auth.WorkloadTypeWorkspace, PolicyTime: time.Now()}
		Expect(dao.UpsertPolicyCompliance(context.Background(), meta.Id, policyInfo, *compliance)).To(BeNil())
		validateComplianceOfAsset(context.Background(), assetChanInfo{
			patchingPtrs: response_patching.ResponsePatchingPtrs{Meta: &meta, Spec: &spec, ComplianceInfo: &compliance},
			workloadType: auth.WorkloadTypeWorkspace, policyInfo: policyInfo, isImposed: true,
		})
		Expect(compliance.Compliance).To(BeFalse())
		Expect(compliance.Imposed).To(BeFalse())
		Expect(compliance.Reason).To(HaveLen(1))
		Expect(compliance.Reason[0].Details).To(ContainSubstring("conflict"))
		saved, dbErr := dao.ListPolicyComplianceOfAssets(context.Background(), []string{meta.Id}, policyInfo)
		Expect(dbErr).To(BeNil())
		Expect(saved).To(HaveLen(1))
		Expect(saved[0].Aligned).To(BeFalse())
		Expect(saved[0].ImposedAsset).To(BeFalse())
		Expect(saved[0].Reason).To(Equal(compliance.Reason))
	})
})

var _ = Describe("GPU allocation defaults in asset compliance", func() {
	for _, replica := range []string{"", "Worker", "Master", "Leader"} {
		DescribeTable("policy resolution for "+replica,
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

				spec, _ := values["compute"].(map[string]interface{})
				if spec == nil {
					spec = map[string]interface{}{}
				}
				before, err := json.Marshal(spec)
				Expect(err).NotTo(HaveOccurred())

				if replica != "" {
					rules = map[string]interface{}{strings.ToLower(replica): rules}
					defaults = map[string]interface{}{strings.ToLower(replica): defaults}
				}

				meta := api.AssetMeta{Id: uuid.NewString(), Name: "policy-gpu-default", Kind: api.ASSETKIND_COMPUTE}
				compliance := &api.ComplianceInfo{Compliance: true, Imposed: true}
				policyInfo := dao.PolicyInfo{TenantId: 1, ProjectId: 1, WorkloadType: auth.WorkloadTypeWorkspace, ReplicaType: replica, PolicyTime: time.Now()}
				validateComplianceOfAsset(context.Background(), assetChanInfo{
					patchingPtrs: response_patching.ResponsePatchingPtrs{Meta: &meta, Spec: &spec, ComplianceInfo: &compliance},
					rules:        rules, defaults: defaults, replicaType: replica,
					defaultFromOrder: policyApi.DefaultFromOrder{Worker: &orderMap, Master: &orderMap},
					workloadType:     auth.WorkloadTypeWorkspace, policyInfo: policyInfo, isImposed: true,
				})
				Expect(compliance.Compliance).To(Equal(wantRule == ""))
				Expect(compliance.Imposed).To(Equal(wantRule == ""))
				if wantRule != "" {
					Expect(compliance.Reason).To(HaveLen(1))
					Expect(compliance.Reason[0].Field).To(Equal(convutils.Ref("compute.gpuComputeMode")))
					Expect(string(*compliance.Reason[0].Rule)).To(Equal(wantRule))
				}
				after, err := json.Marshal(spec)
				Expect(err).NotTo(HaveOccurred())
				Expect(after).To(Equal(before))
				saved, dbErr := dao.ListPolicyComplianceOfAssets(context.Background(), []string{meta.Id}, policyInfo)
				Expect(dbErr).To(BeNil())
				Expect(saved).To(HaveLen(1))
				Expect(saved[0].Aligned).To(Equal(wantRule == ""))
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
