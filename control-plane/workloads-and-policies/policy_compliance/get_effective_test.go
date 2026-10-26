/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy_compliance

import (
	"context"
	"encoding/json"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("replica dependency ordering", func() {
	DescribeTable("normalizes optional replica maps without losing existing dependencies", func(body string) {
		var response policyApi.DistributedPolicyV2
		Expect(json.Unmarshal([]byte(body), &response)).To(Succeed())
		effective := response.Effective.Get()
		expectedWorker := effective.DefaultFromOrder.Get().GetWorker()
		expectedMaster := effective.DefaultFromOrder.Get().GetMaster()
		client := &policies.SimpleMockClient{EffectiveDistributedPolicyV2: effective}
		result, err := GetEffectivePolicyOfProject(context.Background(), client, "4500210", workloadsApi.WORKLOADTYPE_DISTRIBUTED)
		Expect(err).To(BeNil())
		Expect(result).NotTo(BeNil())
		Expect(result.DefaultFromOrder.Worker).NotTo(BeNil())
		Expect(result.DefaultFromOrder.Master).NotTo(BeNil())
		if expectedWorker == nil {
			Expect(*result.DefaultFromOrder.Worker).To(BeEmpty())
		} else {
			Expect(*result.DefaultFromOrder.Worker).To(Equal(expectedWorker))
		}
		if expectedMaster == nil {
			Expect(*result.DefaultFromOrder.Master).To(BeEmpty())
		} else {
			Expect(*result.DefaultFromOrder.Master).To(Equal(expectedMaster))
		}
	},
		Entry("master only", `{"effective":{"defaultFromOrder":{"master":{"": ["compute.cpuCoreRequest"]}}}}`),
		Entry("worker only", `{"effective":{"defaultFromOrder":{"worker":{"": ["compute.cpuCoreRequest"]}}}}`),
	)
})
