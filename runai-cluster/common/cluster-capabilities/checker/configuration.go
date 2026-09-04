/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package checker

import workloadsService "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"

type DistributedFrameworkStatus struct {
	Available   bool
	Suspendable bool
}

// ClusterConfiguration holds the configuration of a cluster
type ClusterConfiguration struct {
	WorkloadOwnershipProtection bool                                                                 // the value of the workload ownership protection
	KnativeAvailable            bool                                                                 // the value of the knative availability status
	KnativeScaleToZeroAvailable bool                                                                 // the value of the knative scale to zero config availability status
	KnativeNodeTypeAvailable    bool                                                                 // the value of the knative node type config availability status
	DistributedFrameworksStatus map[workloadsService.DistributedFramework]DistributedFrameworkStatus // the distributed frameworks status map
	SubdomainRouting            bool                                                                 // whether if the cluster is configurated for subdomain routing (A.K.A Host Base Routing)
	LWSAvailable                bool                                                                 // the value of the LeaderWorkerSet (LWS) availability status
	ClusterDomain               string                                                               // the cluster ingress domain, used to compose auto-generated workload URLs
}
