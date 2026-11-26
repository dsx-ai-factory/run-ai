/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package constants

import (
	clustercommonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
)

const (
	RunaiQueueLabel    = "runai/queue"
	RunaiSchedulerName = "runai-scheduler"

	ExternalWorkloadByOriginalUIDIndexerName = "externalWorkloadByOriginalUIDIndexer"
	SchedulerNameIndexerName                 = "spec.schedulerName"

	GpuResourceName       = clustercommonconstants.NvidiaGpuResourceName
	AmdGpuResourceName    = clustercommonconstants.AmdGpuResourceName
	GpuMemoryResourceName = "run.ai/gpu.memory"

	ManagedByLabelKey                 = "app.kubernetes.io/managed-by"
	HelmReleaseAnnotationKey          = "meta.helm.sh/release-name"
	DynamoGraphDeploymentNameLabelKey = "nvidia.com/dynamo-graph-deployment-name"

	WorkflowLabelKey          = "run.ai/workflow"
	ArgoWorkflowLabelKey      = "workflows.argoproj.io/workflow"
	ArgoNodeNameAnnotationKey = "workflows.argoproj.io/node-name"
)
