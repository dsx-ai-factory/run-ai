/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	v1 "k8s.io/api/core/v1"
)

func CheckConditionIsTrue(conditions []common.WorkloadCondition, conditionType common.WorkloadConditionType) {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			if condition.Status == v1.ConditionTrue {
				return
			}
		}
	}
	ginkgo.Fail(fmt.Sprintf("Condition %s is not True", conditionType))
}

func CheckConditionIsFalseOrUnset(conditions []common.WorkloadCondition, conditionType common.WorkloadConditionType, message string) {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			if condition.Status == v1.ConditionFalse {
				gomega.Expect(condition.Message).To(gomega.ContainSubstring(message))
				return
			}
			ginkgo.Fail(fmt.Sprintf("Condition %s is not False or Unset", conditionType))
		}
	}
}

func CheckConditionIsFalseOrUnsetWithMessageReason(conditions []common.WorkloadCondition, conditionType common.WorkloadConditionType, message, reason string) {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			if condition.Status == v1.ConditionFalse {
				gomega.Expect(condition.Message).To(gomega.ContainSubstring(message))
				gomega.Expect(condition.Reason).To(gomega.Equal(reason))
				return
			}
			ginkgo.Fail(fmt.Sprintf("Condition %s is not False or Unset", conditionType))
		}
	}
}

func GetCondition(conditions []common.WorkloadCondition, conditionType common.WorkloadConditionType) *common.WorkloadCondition {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return &condition
		}
	}
	return nil
}
