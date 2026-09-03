/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package topology

const (
	KAITopologyAnnotationKey          = "kai.scheduler/topology"
	KAITopologyAnnotationPreferredKey = "kai.scheduler/topology-preferred-placement"
	KAITopologyAnnotationRequiredKey  = "kai.scheduler/topology-required-placement"

	KAISegmentSizeAnnotationKey              = "kai.scheduler/segment-size"
	KAISegmentTopologyPreferredAnnotationKey = "kai.scheduler/segment-topology-preferred-placement"
	KAISegmentTopologyRequiredAnnotationKey  = "kai.scheduler/segment-topology-required-placement"
)
