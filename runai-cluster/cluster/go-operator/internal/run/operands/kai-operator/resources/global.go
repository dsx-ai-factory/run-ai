/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	kaiv1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1"
	kaicommon "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/common"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
)

func convertKAIGlobal(runaiConfig *runaiv1.RunaiConfig) *kaiv1.GlobalConfig {
	// Image pull secrets (the control-plane registry secret runai-reg-creds plus any
	// additional secrets) are chart-configured and delivered via AdditionalImagePullSecrets;
	// they go in the ImagePullSecrets slice (the singular ImagePullSecret field is deprecated).
	imagePullSecrets := runaiConfig.Spec.Internal.Global.AdditionalImagePullSecrets

	globalConfig := &kaiv1.GlobalConfig{
		Affinity:                          utils.GetAffinity(runaiConfig),
		Tolerations:                       utils.GetTolerations(runaiConfig),
		DaemonsetsTolerations:             runaiConfig.Spec.Global.DaemonsetsTolerations,
		SecurityContext:                   (*v1.SecurityContext)(runaiConfig.Spec.Global.SecurityContext),
		ImagePullSecrets:                  imagePullSecrets,
		ReplicaCount:                      runaiConfig.Spec.Global.ReplicaCount,
		SchedulerName:                     ptr.To(clustercommonconstants.RunaiScheduler),
		QueueLabelKey:                     ptr.To(clustercommonconstants.RunaiQueueLabel),
		NodePoolLabelKey:                  ptr.To(clustercommonconstants.NodePoolNameLabel),
		RequireDefaultPodAntiAffinityTerm: ptr.To(runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity),
		VPA:                               convertVPA(runaiConfig.Spec.Global.VPA),
		FIPSOnly:                          ptr.To(runaiConfig.Spec.Internal.Global.FipsMode == "only"),
		// Deliberately not configurable: a node that cannot run CUDA-native fractions keeps
		// the workload Pending rather than falling back to the unenforced legacy ones.
		GpuSharingMode: ptr.To(kaicommon.GpuSharingModeNvFractions),
	}

	if runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor == runv1internal.KubernetesFlavorOpenShift {
		globalConfig.Openshift = ptr.To(true)
	}

	return globalConfig
}
