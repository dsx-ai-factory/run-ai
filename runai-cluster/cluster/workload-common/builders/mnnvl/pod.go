/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	"context"
	"fmt"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	drautils "github.com/run-ai/runai/runai-cluster/cluster/workload-common/dra-utils"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ConfigurePodForMNNVL configures a pod spec to support MNNVL with the provided parameters.
// It takes care of setting up node affinity, pod affinity, and resource claim.
// Note that when this function is being used, node affinity is taken care of and thus should not be added again externally.
func ConfigurePodForMNNVL(ctx context.Context, workloadInitials string, workloadName string, workloadId string, mnnvlMode mnnvl.Mode,
	nodepoolList []kaiv1alpha1.NodePool, podSpec *corev1.PodSpec, container *corev1.Container, topologyEnabled bool) error {
	logger := log.FromContext(ctx)

	logger.Info("adding mnnvl resource claim")
	AddResourceClaimTemplateToPod(workloadInitials, workloadName, &podSpec.ResourceClaims, container)

	if topologyEnabled {
		logger.Info(fmt.Sprintf("topology enabled with %s", mnnvl.UserFacingFeatureName))
		return nil
	}

	logger.Info(fmt.Sprintf("modifying pod spec for %s", mnnvl.UserFacingFeatureName))

	logger.Info("adding mnnvl node affinity")
	err := NodeAffinityFromNodePools(
		nodepoolList,
		&podSpec.Affinity,
		mnnvlMode,
	)
	if err != nil {
		return fmt.Errorf("failed to add node affinity for %s support, workload %s: %w",
			mnnvl.UserFacingFeatureName, workloadName, err)
	}

	logger.Info("adding mnnvl pod affinity")
	if podSpec.Affinity.PodAffinity == nil {
		podSpec.Affinity.PodAffinity = &corev1.PodAffinity{}
	}

	gpuAccelerationLabels := lo.Map(nodepoolList, func(nodePool kaiv1alpha1.NodePool, _ int) string {
		return getGPUNetworkAccelerationLabel(&nodePool)
	})
	AddPodAffinity(podSpec.Affinity.PodAffinity, gpuAccelerationLabels, workloadId)

	return nil
}

// AddResourceClaimTemplateToPod adds MNNVL resource claim template and claim to the pod and container.
func AddResourceClaimTemplateToPod(workloadInitials string, workloadName string, resourceClaims *[]corev1.PodResourceClaim, container *corev1.Container) {
	templateName := ComposeResourceClaimTemplateName(workloadName, workloadInitials)
	claimName := GetResourceClaimName()
	drautils.AddResourceClaimTemplateToPodClaims(resourceClaims, claimName, templateName)

	if container != nil {
		drautils.AddResourceClaimToContainer(container, claimName)
	}
}

func ConfigureContainerForMNNVL(container *corev1.Container) {
	claimName := GetResourceClaimName()
	drautils.AddResourceClaimToContainer(container, claimName)
}
