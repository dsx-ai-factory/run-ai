/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"fmt"

	dynamov1beta1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Our pinned Dynamo types omit roles. Read role templates from the raw beta object
// and turn them into components so the existing parser counts their pods and resources.
// Leave typed objects and other API versions unchanged.
func expandDynamoV1beta1RoleTemplates(object client.Object) (client.Object, error) {
	raw, ok := object.(*unstructured.Unstructured)
	if !ok || raw.GetAPIVersion() != dynamov1beta1.GroupVersion.String() || raw.GetKind() != "DynamoGraphDeployment" {
		return object, nil
	}
	dgd := &dynamov1beta1.DynamoGraphDeployment{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, dgd); err != nil {
		return nil, fmt.Errorf("decode Dynamo v1beta1 deployment: %w", err)
	}
	var roleSpecs struct {
		Spec struct {
			Components []struct {
				Roles []struct {
					Name        string                  `json:"name"`
					Replicas    *int32                  `json:"replicas,omitempty"`
					PodTemplate *corev1.PodTemplateSpec `json:"podTemplate,omitempty"`
				} `json:"roles,omitempty"`
			} `json:"components"`
		} `json:"spec"`
	}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &roleSpecs); err != nil {
		return nil, fmt.Errorf("decode Dynamo role templates: %w", err)
	}
	expanded := make([]dynamov1beta1.DynamoComponentDeploymentSharedSpec, 0, len(dgd.Spec.Components))
	for i, component := range dgd.Spec.Components {
		hasTemplates := false
		for _, role := range roleSpecs.Spec.Components[i].Roles {
			if role.PodTemplate == nil {
				continue
			}
			hasTemplates = true
			roleComponent := component
			roleComponent.ComponentName = component.ComponentName + "/" + role.Name
			roleCount := int32(1)
			if role.Name == "worker" && component.Multinode != nil {
				nodeCount := component.Multinode.NodeCount
				if nodeCount == 0 {
					nodeCount = 2
				}
				roleCount = nodeCount - 1
			}
			roleCount = ptr.Deref(role.Replicas, roleCount)
			roleComponent.Replicas = ptr.To(ptr.Deref(component.Replicas, 1) * roleCount)
			roleComponent.Multinode = nil
			roleComponent.PodTemplate = role.PodTemplate
			expanded = append(expanded, roleComponent)
		}
		if !hasTemplates {
			expanded = append(expanded, component)
		}
	}
	dgd.Spec.Components = expanded
	return dgd, nil
}
