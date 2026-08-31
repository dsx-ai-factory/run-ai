/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package legacycleanup

import (
	"fmt"

	semver "github.com/Masterminds/semver/v3"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SourceRequiresCleanup reports whether the running 3.0+ cleanup implementation
// received upgrade provenance from a release that still used legacy controllers.
func SourceRequiresCleanup(sourceVersion string) (bool, error) {
	if sourceVersion == "" {
		return false, fmt.Errorf("source version is missing")
	}

	source, err := semver.NewVersion(sourceVersion)
	if err != nil {
		return false, fmt.Errorf("parse source version %q: %w", sourceVersion, err)
	}
	return source.Major() > 0 && source.Major() < 3, nil
}

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
