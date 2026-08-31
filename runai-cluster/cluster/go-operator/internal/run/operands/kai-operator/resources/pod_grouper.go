/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	kaicommon "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/common"
	kaipodgrouper "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/pod_grouper"
	"k8s.io/utils/ptr"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	pod_grouper "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/pod-grouper"

	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
)

func convertPodGrouperConfig(runaiConfig *runaiv1.RunaiConfig) *kaipodgrouper.PodGrouper {
	podGrouperConfig := runaiConfig.Spec.PodGrouper
	return &kaipodgrouper.PodGrouper{
		Service: &kaicommon.Service{
			Enabled:   podGrouperConfig.Enabled,
			Image:     convertImage(podGrouperConfig.Image),
			Resources: (*kaicommon.Resources)(podGrouperConfig.Resources),
			Affinity:  podGrouperConfig.Affinity,
		},
		Args:                    convertPodGrouperArgs(podGrouperConfig),
		K8sClientConfig:         convertPodGrouperK8sClientConfig(podGrouperConfig),
		MaxConcurrentReconciles: podGrouperConfig.MaxConcurrentReconciles,
		VPA:                     convertVPA(podGrouperConfig.VPA),
	}
}

func convertPodGrouperArgs(podGrouperConfig *pod_grouper.PodGrouper) *kaipodgrouper.Args {
	var gangScheduleKnative *bool
	if podGrouperConfig.Args != nil {
		gangScheduleKnative = podGrouperConfig.Args.GangScheduleKnative
	}
	return &kaipodgrouper.Args{
		GangScheduleKnative:                 gangScheduleKnative,
		DefaultPrioritiesConfigMapName:      ptr.To(constants.WorkloadsConfigMapName),
		DefaultPrioritiesConfigMapNamespace: ptr.To(constants.WorkloadsConfigMapNamespace),
	}
}

func convertPodGrouperK8sClientConfig(podGrouperConfig *pod_grouper.PodGrouper) *kaicommon.K8sClientConfig {
	if podGrouperConfig.K8sClientConfig == nil {
		return nil
	}
	return &kaicommon.K8sClientConfig{
		Burst: podGrouperConfig.K8sClientConfig.Burst,
		QPS:   podGrouperConfig.K8sClientConfig.QPS,
	}
}
