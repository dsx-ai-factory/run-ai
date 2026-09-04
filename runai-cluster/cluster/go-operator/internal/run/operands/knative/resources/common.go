/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

// shouldCreateServiceMonitors checks if service monitors should be created based on the RunaiConfig
func shouldCreateServiceMonitors(runaiConfig *runv1.RunaiConfig) bool {
	return runaiConfig.Spec.Knative.CreateServiceMonitors != nil && *runaiConfig.Spec.Knative.CreateServiceMonitors
}
