/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	kaicommon "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/common"
	kaipodgroupcontroller "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/pod_group_controller"
	"k8s.io/utils/ptr"

	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

func convertPodGroupControllerConfig(runaiConfig *runaiv1.RunaiConfig) *kaipodgroupcontroller.PodGroupController {
	config := runaiConfig.Spec.PodGroupController
	return &kaipodgroupcontroller.PodGroupController{
		Service: &kaicommon.Service{
			Enabled:   config.Enabled,
			Image:     convertImage(config.Image),
			Resources: (*kaicommon.Resources)(config.Resources),
			Affinity:  config.Affinity,
		},
		Webhooks: &kaipodgroupcontroller.PodGroupControllerWebhooks{
			EnableValidation:               ptr.To(true),
			WebhookConfigurationNamePrefix: ptr.To("runai-podgroup-validation-"),
		},
		VPA: convertVPA(config.VPA),
	}
}
