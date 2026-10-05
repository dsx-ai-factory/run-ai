/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configmaps

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// An Available condition alone can describe old pods during a blocked upgrade.
// Advertise the contract only after all configured component images finish
// rolling out. The control plane separately checks operand/CRD readiness.
func (rp *RunaiPublic) inferenceEndpointIntegrationAvailable(ctx context.Context) bool {
	agent, aw, endpoint := rp.runaiConfig.Spec.Agent, rp.runaiConfig.Spec.AnyworkloadController, rp.runaiConfig.Spec.InferenceEndpointController
	if agent == nil || aw == nil || endpoint == nil || !ptr.Deref(agent.Enabled, false) || !ptr.Deref(aw.Enabled, false) || !ptr.Deref(endpoint.Enabled, false) {
		return false
	}
	for _, component := range []struct {
		name  string
		image *common.Image
	}{
		{operands.RunaiAgentName, agent.Image},
		{operands.RunaiAnyworkloadControllerName, aw.Image},
		{operands.RunaiInferenceEndpointControllerName, endpoint.Image},
	} {
		expectedImage := formatImageRef(component.image)
		if expectedImage == nil {
			return false
		}
		deployment := &appsv1.Deployment{}
		if err := rp.kbClient.Get(ctx, client.ObjectKey{Namespace: rp.runaiConfig.Namespace, Name: component.name}, deployment); err != nil {
			return false
		}
		replicas := ptr.Deref(deployment.Spec.Replicas, 1)
		if replicas < 1 || deployment.Status.ObservedGeneration != deployment.Generation ||
			deployment.Status.Replicas != replicas || deployment.Status.UpdatedReplicas != replicas || deployment.Status.AvailableReplicas != replicas {
			return false
		}
		matched := false
		for _, container := range deployment.Spec.Template.Spec.Containers {
			if container.Name == component.name && container.Image == *expectedImage {
				matched = true
			}
		}
		if !matched {
			return false
		}
	}
	return true
}
