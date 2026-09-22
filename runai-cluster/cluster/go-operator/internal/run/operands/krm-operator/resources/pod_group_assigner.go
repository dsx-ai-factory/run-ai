/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	kaicommon "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/common"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"k8s.io/utils/ptr"

	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	pod_group_assigner "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/pod-group-assigner"
	runv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
)

func convertPodGroupAssigner(runaiConfig *runaiv1.RunaiConfig) *kaiv1alpha1.PodGroupAssigner {
	config := runaiConfig.Spec.PodGroupAssigner
	if config == nil {
		return nil
	}

	return &kaiv1alpha1.PodGroupAssigner{
		Service: &kaicommon.Service{
			Enabled:         config.Enabled,
			Resources:       (*kaicommon.Resources)(config.Resources),
			K8sClientConfig: (*kaicommon.K8sClientConfig)(config.K8sClientConfig),
			Affinity:        config.Affinity,
			// Image is KRM's own, resolved from its chart. PodDisruptionBudget is
			// left to KRM's default of disabled, matching runai, which creates none.
		},
		ControllerService: &kaiv1alpha1.PodGroupAssignerService{
			Webhook: podGroupAssignerWebhookPorts(config),
		},
		Webhooks: &kaiv1alpha1.PodGroupAssignerWebhooks{
			// The Pod mutation runai needs is served by its admission-controller.
			// The PodGroup webhook has no toggle and is always served.
			EnablePodWebhook: ptr.To(false),
		},
		Args: &kaiv1alpha1.PodGroupAssignerArgs{
			Debug:                      podGroupAssignerDebug(config),
			UnexistingNodepoolSentinel: ptr.To(runv1alpha1.UnexistingRunaiNodePool),
			AnnotationNodepoolsKey:     ptr.To(commonconstants.AnnotationNodepools),
		},
		Replicas: config.Replicas,
		VPA:      convertVPA(config.VPA),
	}
}

func podGroupAssignerWebhookPorts(config *pod_group_assigner.PodGroupAssigner) *kaiv1alpha1.PortMapping {
	if config.Service == nil {
		return nil
	}
	return portMapping(config.Service.Webhook.Port, config.Service.Webhook.TargetPort)
}

func portMapping(port, targetPort *int) *kaiv1alpha1.PortMapping {
	if port == nil && targetPort == nil {
		return nil
	}

	mapping := &kaiv1alpha1.PortMapping{}
	if port != nil {
		mapping.Port = ptr.To(int32(*port)) // #nosec G115 -- port value is bounded by config defaults (443)
	}
	if targetPort != nil {
		mapping.TargetPort = ptr.To(int32(*targetPort)) // #nosec G115 -- port value is bounded by config defaults (8443)
	}
	return mapping
}

func podGroupAssignerDebug(config *pod_group_assigner.PodGroupAssigner) *bool {
	if config.Args == nil {
		return nil
	}
	return ptr.To(config.Args.DebugLog)
}
