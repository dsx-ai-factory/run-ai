/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer_v2

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
)

// SetClusterDependentOptions resolves the cluster-version-dependent composer options.
// Min supported cluster version is 2.22, so capabilities introduced at <= 2.21 are
// always supported and no longer appear here. See checker.clusterVersionToCapabilityMap.
func SetClusterDependentOptions(ctx context.Context, options *composer.FieldsComposerOptions, validator *checker.ClusterCapabilitiesChecker) error {
	capabilitiesToOption := map[int]*bool{
		checker.MinVersionOfFlexiblePriorityClasses:     &options.FlexiblePriorityClassesSupported,
		checker.MinVersionOfPreemptibility:              &options.PreemptibilitySupported,
		checker.MinVersionOfLauncherCreationPolicy:      &options.LauncherCreationPolicySupported,
		checker.MinVersionOfGitSyncWithSSHKey:           &options.GitSyncWithSshKeySupported,
		checker.MinVersionOfDefaultModeSupported:        &options.VolumeDefaultModeSupported,
		checker.MinVersionOfEmptyDirAsVolume:            &options.EmptyDirAsVolumeSupported,
		checker.MinVersionOfNodeAffinity:                &options.NodeAffinitySupported,
		checker.MinVersionWorkloadV2NetworkingEndpoints: &options.WorkloadV2NetworkingSupported,
		checker.MinVersionWorkloadV2ServiceSpecLabels:   &options.WorkloadV2ServiceSpecLabelsSupported,
		checker.MinVersionGpuMemoryLimitAnnotations:     &options.GpuMemoryLimitAnnotationsSupported,
		checker.MinVersionGpuComputeMode:                &options.GpuComputeModeSupported,
		checker.MinVersionOAuth2ProxyTemplates:          &options.OAuth2ProxyTemplatesSupported,
	}

	for capability, optionPtr := range capabilitiesToOption {
		supported, errSupported := validator.ValidateClusterCapability(ctx, capability)
		if errSupported != nil {
			return errSupported
		}

		*optionPtr = supported
	}

	return nil
}
