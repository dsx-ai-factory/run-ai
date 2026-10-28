/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromos "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Routing readiness describes the endpoint controller's resources, not model
// traffic health. Preserve the workload's own conditions, phase and URLs.
func appendInferenceEndpointCondition(runner *dromos.WorkloadRunner, fields *openapi.WorkloadSyncFields) {
	if _, configured := runner.Annotations[endpoints.InferenceEndpointAnnotation]; !configured {
		return
	}
	condition := metav1.Condition{
		Type: endpoints.InferenceEndpointReadyCondition, Status: metav1.ConditionUnknown,
		Reason: "Pending", Message: "waiting for current inference routing status",
	}
	if plugin := runner.Status.Plugins[endpoints.InferenceEndpointPluginName]; plugin != nil {
		if observed := meta.FindStatusCondition(plugin.Conditions, endpoints.InferenceEndpointReadyCondition); observed != nil && observed.ObservedGeneration == runner.Generation {
			condition = *observed
		}
	}
	fields.Conditions = append(fields.Conditions, openapi.Condition{
		Type: condition.Type, Status: string(condition.Status),
		Reason: strPointerIfNotEmpty(condition.Reason), Message: strPointerIfNotEmpty(condition.Message),
		LastTransitionTime: getApiTime(condition.LastTransitionTime),
	})
}
