/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tests

import (
	// import gomega
	"time"

	. "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/tests/utils"
	v1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("RunaiConfig Managed Operands Status", func() {
	const (
		operandName = "cluster-sync"
	)

	When("an operand is not ready", func() {
		It("should report the operand as not ready, including the reason", func() {
			runaiConfig.Spec.ClusterSync.Image.PullPolicy = "invalid"
			EventuallyUpdateResource(k8s, runaiConfig)
			PollForTrueCondition(k8s, v1.ConditionTypeDeployed, runaiConfig)

			eventuallyExpectOperandStatusReadiness(operandName, false)
		})
	})

	When("an operand is ready", func() {
		It("should report the operand as ready, without a reason", func() {
			runaiConfig.Spec.ClusterSync.Image.PullPolicy = "Always"
			EventuallyUpdateResource(k8s, runaiConfig)
			PollForTrueCondition(k8s, v1.ConditionTypeDeployed, runaiConfig)

			eventuallyExpectOperandStatusReadiness(operandName, true)
		})
	})

	When("an operand is disabled", func() {
		It("should not report the operand's status", func() {
			runaiConfig.Spec.ClusterSync.Enabled = ptr.To(false)
			EventuallyUpdateResource(k8s, runaiConfig)
			PollForTrueCondition(k8s, v1.ConditionTypeDeployed, runaiConfig)

			Eventually(isOperandStatusExistsFn(operandName), 20*time.Second, 500*time.Millisecond).Should(BeFalse())
			Consistently(isOperandStatusExistsFn(operandName)).Should(BeFalse())
		})
	})
})

func eventuallyExpectOperandStatusReadiness(operandName string, expectedReadiness bool) {
	GinkgoHelper()

	Eventually(func() bool {
		runaiConfig = GetTestingRunaiConfig(k8s)
		opStatus, ok := runaiConfig.Status.OperandStatuses[operandName]

		if expectedReadiness {
			return ok && opStatus.Ready && opStatus.Reasons == nil
		}
		return ok && !opStatus.Ready && opStatus.Reasons != nil
	})
}

func isOperandStatusExistsFn(operandName string) func() bool {
	return func() bool {
		GinkgoHelper()

		runaiConfig = GetTestingRunaiConfig(k8s)
		_, ok := runaiConfig.Status.OperandStatuses[operandName]
		return ok
	}
}
