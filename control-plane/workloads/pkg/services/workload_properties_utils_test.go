/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/clients/policy"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/testutils"
	"github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/structs"
)

func validatePolicyForWorkloadType(workloadType string, config *map[string]*structs.PolicyWorkloadTypeConfig, expectedCategory, expectedPriority string) {
	workloadConfig := (*config)[workloadType]
	Expect(workloadConfig).To(Not(BeNil()))
	Expect(workloadConfig.PriorityName).NotTo(BeNil())
	Expect(workloadConfig.CategoryName).NotTo(BeNil())
	Expect(*workloadConfig.CategoryName).To(Equal(expectedCategory))
	Expect(*workloadConfig.PriorityName).To(Equal(expectedPriority))
}

var _ = Describe("workspaces service validations", func() {
	var ctx context.Context
	var randoms *testutils.RandomValues

	BeforeEach(func() {
		var policyMock = policy.NewSimpleMock()
		policy.SetMock(policyMock)
		randoms = testutils.NewRandomValues()
		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

		var configWithPartialData = map[string]*structs.PolicyWorkloadTypeConfig{
			"InteractiveWorkload": {
				CategoryName: nil,
				PriorityName: convutils.Ref("high"),
			},
			"DistributedWorkload": {
				CategoryName: convutils.Ref("Train"),
				PriorityName: convutils.Ref("low"),
			},
			"TrainingWorkload": {
				CategoryName: convutils.Ref("Train"),
				PriorityName: nil,
			},
			"InferenceWorkload": {
				CategoryName: nil,
				PriorityName: nil,
			},
			"DistributedInferenceWorkload": {
				CategoryName: convutils.Ref("Deploy"),
				PriorityName: convutils.Ref("very-high"),
			},
		}

		policyMock.SetPoliciesConfig(configWithPartialData)
	})

	Context("get workload config of tenant - category null", func() {
		It("category null in tenant policies", func() {
			config, err := GetWorkloadConfigOfTenant(ctx)
			Expect(err).To(BeNil())
			validatePolicyForWorkloadType("InteractiveWorkload", config, "Build", "high")
		})

		It("priority null in tenant policies", func() {
			config, err := GetWorkloadConfigOfTenant(ctx)
			Expect(err).To(BeNil())
			validatePolicyForWorkloadType("TrainingWorkload", config, "Train", "low")
		})

		It("both category and priority null in tenant policies", func() {
			config, err := GetWorkloadConfigOfTenant(ctx)
			Expect(err).To(BeNil())
			validatePolicyForWorkloadType("InferenceWorkload", config, "Deploy", "very-high")
		})
	})

})
