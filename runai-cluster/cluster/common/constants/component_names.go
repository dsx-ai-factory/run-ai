/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

const (
	RunaiInferenceEndpointControllerName = "inference-endpoint-controller"
	RunaiOperatorName                    = "runai-operator"
	RunaiAdviceRestartControllerName     = "advice-restart-controller"
	RunaiAnyworkloadControllerName       = "anyworkload-controller"
	RunaiApplicationControllerName       = "application-controller"
	RunaiSharedObjectsControllerName     = "shared-objects-controller"
	RunaiClusterApiName                  = "cluster-api"
	RunaiAgentName                       = "runai-agent"
	RunaiClusterSyncName                 = "cluster-sync"
	RunaiAccessRuleControllerName        = "accessrule-controller"
	RunaiProjectControllerName           = "runai-project-controller"
	RunaiExternalWorkloadIntegratorName  = "external-workload-integrator"
	RunaiWorkloadExporterName            = "workload-exporter"

	// KRMProjectControllerName is KRM's project-controller, which replaced the
	// Run:ai-built one named by RunaiProjectControllerName. That name is still the
	// right one for the resources the now-disabled Run:ai operand owns and deletes.
	KRMProjectControllerName = "project-controller"
)
