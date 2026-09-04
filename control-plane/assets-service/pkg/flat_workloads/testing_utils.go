/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package flat_workloads

import (
	"encoding/json"

	. "github.com/onsi/gomega" //nolint:staticcheck // dot-import is the idiomatic gomega usage in test helpers
	policyService "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
)

func ConvertToWorkspacePolicy(rules *policyService.WorkspacePolicyDefaultsAndRulesV2) *policyService.WorkspacePolicyDefaultsAndRulesV2 {
	asJson, err := json.Marshal(rules)
	Expect(err).To(BeNil())

	var converted policyService.WorkspacePolicyDefaultsAndRulesV2
	err = json.Unmarshal(asJson, &converted)
	Expect(err).To(BeNil())

	return &converted
}

func ConvertToTrainingPolicy(rules *policyService.TrainingPolicyDefaultsAndRulesV2) *policyService.TrainingPolicyDefaultsAndRulesV2 {
	asJson, err := json.Marshal(rules)
	Expect(err).To(BeNil())

	var converted policyService.TrainingPolicyDefaultsAndRulesV2
	err = json.Unmarshal(asJson, &converted)
	Expect(err).To(BeNil())

	return &converted
}

func ConvertToDistributedPolicy(rules *policyService.DistributedPolicyDefaultsAndRulesV2) *policyService.DistributedPolicyDefaultsAndRulesV2 {
	asJson, err := json.Marshal(rules)
	Expect(err).To(BeNil())

	var converted policyService.DistributedPolicyDefaultsAndRulesV2
	err = json.Unmarshal(asJson, &converted)
	Expect(err).To(BeNil())

	return &converted
}

func ConvertToDistributedInferencePolicy(rules *policyService.DistributedInferencePolicyDefaultsAndRulesV2) *policyService.DistributedInferencePolicyDefaultsAndRulesV2 {
	asJson, err := json.Marshal(rules)
	Expect(err).To(BeNil())

	var converted policyService.DistributedInferencePolicyDefaultsAndRulesV2
	err = json.Unmarshal(asJson, &converted)
	Expect(err).To(BeNil())

	return &converted
}
