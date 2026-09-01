/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacycleanup

import (
	"fmt"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func validateReadinessMarker(config *runaiv1.RunaiConfig) error {
	if config == nil {
		return fmt.Errorf("RunaiConfig is missing")
	}
	if config.Status.ObservedGeneration != config.Generation {
		return fmt.Errorf("RunaiConfig observed generation %d does not match generation %d",
			config.Status.ObservedGeneration, config.Generation)
	}
	if err := requireTrueCondition(config, runaiv1.ConditionTypeReconciled); err != nil {
		return err
	}
	return requireTrueCondition(config, runaiv1.ConditionTypeAvailable)
}

func ValidateReadinessMarker(config *runaiv1.RunaiConfig) error {
	return validateReadinessMarker(config)
}

func requireTrueCondition(config *runaiv1.RunaiConfig, conditionType runaiv1.ConditionType) error {
	condition := config.Status.GetCondition(conditionType)
	if condition == nil {
		return fmt.Errorf("RunaiConfig %s condition is missing", conditionType)
	}
	if condition.ObservedGeneration != config.Generation {
		return fmt.Errorf("RunaiConfig %s condition observed generation %d does not match generation %d",
			conditionType, condition.ObservedGeneration, config.Generation)
	}
	if condition.Status != metav1.ConditionTrue {
		return fmt.Errorf("RunaiConfig %s condition is %s: %s", conditionType, condition.Status, condition.Message)
	}
	return nil
}
