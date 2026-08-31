/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"k8s.io/utils/ptr"

	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	corev1 "k8s.io/api/core/v1"
)

const (
	defaultOperator = string(corev1.TolerationOpExists)
	defaultEffect   = string(corev1.TaintEffectNoExecute)
)

// GetTolerations returns the list of tolerations associated with the pods of the Workload
func GetTolerations(_ context.Context, containerFields *field_groups.ContainerFields) ([]corev1.Toleration, error) {

	var result []corev1.Toleration

	tolerations := containerFields.Tolerations.Deref()

	for _, nick := range flds.SortedTolerationMapKeys(tolerations) {

		rule := tolerations[nick]
		if rule.Value == nil {
			continue
		}

		operator := ptr.Deref(rule.Value.Operator, defaultOperator)
		value := ptr.Deref(rule.Value.Value, "")

		if operator == string(corev1.TolerationOpExists) && value != "" {
			return nil, fmt.Errorf("tolerationValue must be empty when operator is %s", operator)
		}

		var seconds *int64
		if rule.Value.Seconds != nil {
			seconds = ptr.To(int64(*rule.Value.Seconds))
		}

		result = append(result, corev1.Toleration{
			Operator:          corev1.TolerationOperator(operator),
			Key:               ptr.Deref(rule.Value.Key, ""),
			Value:             value,
			Effect:            corev1.TaintEffect(ptr.Deref(rule.Value.Effect, "")),
			TolerationSeconds: seconds,
		})
	}

	return result, nil
}
