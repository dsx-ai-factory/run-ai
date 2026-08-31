/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package post_hooks

import (
	"fmt"

	semver "github.com/Masterminds/semver/v3"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	accessrule_controller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/accessrule-controller"
	advice_restart_controller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/advice-restart-controller"
	inference_workload_controller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/inference-workload-controller"
	job_controller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/job-controller"
	workload_controller "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/workload-controller"
	"k8s.io/utils/ptr"
)

const legacyCleanupSourceVersionAnnotation = "run.ai/legacy-workload-cleanup-source-version"

type controllerTransitionPath string

const (
	controllerTransitionFreshInstall controllerTransitionPath = "fresh-install"
	controllerTransitionUpgrade      controllerTransitionPath = "upgrade"
)

// applyControllerTransitionDefaults keeps legacy controllers available while a
// pre-3.0 cluster rolls forward, but prevents them from ever being reconciled on
// a fresh 3.0+ installation. It runs after the clusterConfig merge patch so the
// transition state cannot be overridden before the guarded cleanup completes.
func applyControllerTransitionDefaults(
	runaiConfig *runaiv1.RunaiConfig,
	transitionPath controllerTransitionPath,
	previousVersion string,
) error {
	if runaiConfig == nil {
		return fmt.Errorf("runai config is nil")
	}

	switch transitionPath {
	case controllerTransitionFreshInstall:
		setLegacyControllersEnabled(runaiConfig, false)
		setLegacyProjectPermissionsEnabled(runaiConfig, false)
		delete(runaiConfig.Annotations, legacyCleanupSourceVersionAnnotation)
		return nil
	case controllerTransitionUpgrade:
	default:
		return fmt.Errorf("unsupported controller transition path %q", transitionPath)
	}

	if previousVersion == "" {
		return fmt.Errorf("previous Run:ai version is missing for a 3.0+ upgrade")
	}

	previous, err := semver.NewVersion(previousVersion)
	if err != nil {
		return fmt.Errorf("parse previous Run:ai version %q: %w", previousVersion, err)
	}
	if previous.Major() == 0 || previous.Major() >= 3 {
		return nil
	}

	setLegacyControllersEnabled(runaiConfig, true)
	setCleanupAnnotation(runaiConfig, legacyCleanupSourceVersionAnnotation, previousVersion)
	if runaiConfig.Spec.AdviceRestartController == nil {
		runaiConfig.Spec.AdviceRestartController = &advice_restart_controller.AdviceRestartController{}
	}
	runaiConfig.Spec.AdviceRestartController.Enabled = ptr.To(false)
	return nil
}

func setCleanupAnnotation(runaiConfig *runaiv1.RunaiConfig, key, value string) {
	if runaiConfig.Annotations == nil {
		runaiConfig.Annotations = map[string]string{}
	}
	runaiConfig.Annotations[key] = value
}

func setLegacyControllersEnabled(runaiConfig *runaiv1.RunaiConfig, enabled bool) {
	if runaiConfig.Spec.WorkloadController == nil {
		runaiConfig.Spec.WorkloadController = &workload_controller.WorkloadController{}
	}
	if runaiConfig.Spec.InferenceWorkloadController == nil {
		runaiConfig.Spec.InferenceWorkloadController = &inference_workload_controller.InferenceWorkloadController{}
	}
	if runaiConfig.Spec.JobController == nil {
		runaiConfig.Spec.JobController = &job_controller.JobController{}
	}

	runaiConfig.Spec.WorkloadController.Enabled = ptr.To(enabled)
	runaiConfig.Spec.InferenceWorkloadController.Enabled = ptr.To(enabled)
	runaiConfig.Spec.JobController.Enabled = ptr.To(enabled)
}

func setLegacyProjectPermissionsEnabled(runaiConfig *runaiv1.RunaiConfig, enabled bool) {
	if runaiConfig.Spec.AccessRuleController == nil {
		runaiConfig.Spec.AccessRuleController = &accessrule_controller.AccessRuleController{}
	}
	if runaiConfig.Spec.AccessRuleController.Args == nil {
		runaiConfig.Spec.AccessRuleController.Args = &accessrule_controller.Args{}
	}
	runaiConfig.Spec.AccessRuleController.Args.LegacyProjectPermissions = ptr.To(enabled)
}
