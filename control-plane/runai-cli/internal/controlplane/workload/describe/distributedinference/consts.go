/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributedinference

const (
	// Distributed Inference field names
	FieldRestartPolicy     = "Restart Policy"
	FieldStartupPolicy     = "Startup Policy"
	FieldReplicasRequested = "Replicas (Requested)"
	FieldReplicasReady     = "Replicas (Ready)"
	FieldWorkersPerReplica = "Workers per Replica"
	FieldServingPort       = "Port"
	FieldServingURL        = "Serving URL"
	FieldAuthorizationType = "Authorization Type"
	FieldAuthorizedUsers   = "Authorized Users / Service Accounts"
	FieldAuthorizedGroups  = "Authorized Groups"
	FieldExposeExternally  = "Expose Externally"

	// Compute field names (for leader/worker)
	FieldCpuRequest       = "CPU Cores (Request)"
	FieldCpuLimit         = "CPU Cores (Limit)"
	FieldCpuMemoryRequest = "CPU Memory (Request)"
	FieldCpuMemoryLimit   = "CPU Memory (Limit)"

	// Section titles
	TitleDistributedInference = "Distributed Inference Configuration"
	TitleServingConfiguration = "Serving Configuration"
	TitleLeaderConfiguration  = "Leader Configuration"
	TitleWorkerConfiguration  = "Worker Configuration"
)
