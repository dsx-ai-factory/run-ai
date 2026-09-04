/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	kaicommon "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/common"
	kainsa "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/node_scale_adjuster"
	"k8s.io/utils/ptr"

	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	nsa "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/node-scale-adjuster"
)

func convertNodeScaleAdjusterConfig(runaiConfig *runaiv1.RunaiConfig) *kainsa.NodeScaleAdjuster {
	nsaConfig := runaiConfig.Spec.NodeScaleAdjuster
	return &kainsa.NodeScaleAdjuster{
		Service: &kaicommon.Service{
			Enabled:   ptr.To(ptr.Deref(nsaConfig.Enabled, false)),
			Image:     convertImage(nsaConfig.Image),
			Resources: (*kaicommon.Resources)(nsaConfig.Resources),
			Affinity:  nsaConfig.Affinity,
		},
		Args: convertNodeScaleAdjusterArgs(runaiConfig, nsaConfig),
	}
}

func convertNodeScaleAdjusterArgs(runaiConfig *runaiv1.RunaiConfig, nsaConfig *nsa.NodeScaleAdjuster) *kainsa.Args {
	args := &kainsa.Args{
		NodeScaleNamespace: ptr.To(runaiv1.ScaleAdjustNamespace),
		ScalingPodImage:    convertImage(nsaConfig.ScalingPodImage),
	}
	if nsaConfig.Args != nil {
		args.GPUMemoryToFractionRatio = nsaConfig.Args.GPUMemoryToFractionRatio
	}
	return args
}
