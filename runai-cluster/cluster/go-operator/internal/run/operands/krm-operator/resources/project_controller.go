/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"fmt"
	"strconv"

	kaicommon "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1/common"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"k8s.io/utils/ptr"

	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	project_controller "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/project-controller"
)

func convertProjectController(runaiConfig *runaiv1.RunaiConfig) *kaiv1alpha1.ProjectController {
	config := runaiConfig.Spec.ProjectController
	if config == nil {
		return nil
	}

	return &kaiv1alpha1.ProjectController{
		Service: &kaicommon.Service{
			Enabled:   config.Enabled,
			Resources: (*kaicommon.Resources)(config.Resources),
			// KAI's client defaults are lower than runai's, so leaving this unset
			// would quietly throttle the controller.
			K8sClientConfig: (*kaicommon.K8sClientConfig)(config.K8sClientConfig),
			Affinity:        config.Affinity,
			// Image is KRM's own, resolved from its chart. PodDisruptionBudget is
			// left to KRM's default of disabled, matching runai, which creates none.
		},
		ControllerService: &kaiv1alpha1.ProjectControllerService{
			Metrics: projectControllerMetricsPorts(config),
			// No webhook port: both validating webhooks are off, see Webhooks below.
		},
		Webhooks: &kaiv1alpha1.ProjectControllerWebhooks{
			// Project and department validation are served by the agent, so the
			// chart renders no webhook configuration and mints no certificate.
			EnableProjectValidation:    ptr.To(false),
			EnableDepartmentValidation: ptr.To(false),
		},
		Features: &kaiv1alpha1.ProjectControllerFeatures{
			CreateNamespaces:   config.CreateNamespaces,
			CreateRoleBindings: config.CreateRoleBindings,
			LimitRange:         config.LimitRange,
		},
		Profiling: projectControllerProfiling(config),
		Args: &kaiv1alpha1.ProjectControllerArgs{
			Debug:                          projectControllerDebug(config),
			ProjectNamePrefix:              ptr.To(clustercommonconstants.NamespaceRunAi),
			ProjectIDLabelKey:              ptr.To(clustercommonconstants.ProjectIdLabel),
			QueueDepartmentNameLabelKey:    ptr.To(clustercommonconstants.QueueDepartmentNameLabel),
			NamespaceVersionLabelKey:       ptr.To(clustercommonconstants.RunaiNamespaceVersionLabel),
			ResourceManualOverrideLabelKey: ptr.To(clustercommonconstants.RunaiResourceManualOverrideLabel),
			LimitRangeName:                 ptr.To(clustercommonconstants.RunaiLimitRangeName),
		},
		// Neither is modelled on KRMConfig. The blockers namespace is where the
		// controller reads the ConfigMap the operator writes for it.
		ExtraArgs: []string{
			fmt.Sprintf("--install-namespace=%s", runaiConfig.GetNamespace()),
			fmt.Sprintf("--project-delete-blockers-namespace=%s", clustercommonconstants.NamespaceRunAi),
		},
		RoleBindingsConfigMapName: ptr.To(clustercommonconstants.ProjectControllerRoleBindingCm),
		Replicas:                  config.Replicas,
		// RunaiConfig carries no VPA or extra project role bindings for this
		// controller, so both stay at KRM's defaults.
	}
}

func projectControllerMetricsPorts(config *project_controller.ProjectController) *kaiv1alpha1.PortMapping {
	if config.Service == nil {
		return nil
	}
	return portMapping(config.Service.Metrics.Port, config.Service.Metrics.TargetPort)
}

// projectControllerProfiling carries the port as a number; RunaiConfig stores it
// as a string, and an unparseable value is left to KRM's own default.
func projectControllerProfiling(config *project_controller.ProjectController) *kaiv1alpha1.Profiling {
	if config.Profiling == nil {
		return nil
	}

	profiling := &kaiv1alpha1.Profiling{Enabled: config.Profiling.Enabled}
	if apiPort, err := strconv.ParseInt(config.Profiling.ApiPort, 10, 32); err == nil {
		profiling.APIPort = ptr.To(int32(apiPort))
	}
	return profiling
}

// projectControllerDebug mirrors the operand: this controller takes a verbosity
// number rather than a boolean, and anything below zero means debug.
func projectControllerDebug(config *project_controller.ProjectController) *bool {
	if config.Args == nil || config.Args.Verbosity == nil {
		return nil
	}
	return ptr.To(*config.Args.Verbosity < 0)
}
