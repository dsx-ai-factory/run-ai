/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	swagger "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	global_consts "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	schedulingv1 "k8s.io/api/scheduling/v1"
	"k8s.io/utils/ptr"
)

func ParseBackendPriority(pc *schedulingv1.PriorityClass, source string, phase string) swagger.PrioritiesForStatusPrioritiesInner {
	backendPriorityClass := swagger.PrioritiesForStatusPrioritiesInner{
		Name:   &pc.Name,
		Value:  &pc.Value,
		Source: ptr.To(source),
		Status: &swagger.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: ptr.To(phase)},
	}
	id := pc.Labels[global_consts.PriorityClassIdLabel]
	if id != "" {
		backendPriorityClass.Id = ptr.To(id)
	}
	return backendPriorityClass
}
