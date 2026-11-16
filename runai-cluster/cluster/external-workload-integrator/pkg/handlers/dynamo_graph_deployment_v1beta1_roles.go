/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	dynamov1beta1 "github.com/ai-dynamo/dynamo/deploy/operator/api/v1beta1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	riresource "github.com/run-ai/karta/pkg/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// The pinned Dynamo Go types omit roles, so extraction needs the original manifest.
func (h *DynamoGraphDeploymentTopOwnerHandler) getDynamoV1beta1RoleInstances(ctx context.Context, object client.Object) (map[string]riresource.ExtractedInstance, []string, error) {
	raw, ok := object.(*unstructured.Unstructured)
	if !ok || raw.GetAPIVersion() != dynamov1beta1.GroupVersion.String() || raw.GetKind() != "DynamoGraphDeployment" {
		return nil, nil, nil
	}
	components, _, err := unstructured.NestedSlice(raw.Object, "spec", "components")
	if err != nil {
		return nil, nil, fmt.Errorf("read Dynamo v1beta1 components: %w", err)
	}
	hasRoleTemplates := false
	for _, rawComponent := range components {
		component, _ := rawComponent.(map[string]any)
		roles, _ := component["roles"].([]any)
		for _, rawRole := range roles {
			role, _ := rawRole.(map[string]any)
			if role["podTemplate"] != nil {
				hasRoleTemplates = true
			}
		}
	}
	// Workloads without role templates must still parse when Karta is missing or outdated.
	if !hasRoleTemplates {
		return nil, nil, nil
	}
	karta, err := utils.GetKartaOfGvk(ctx, h.cachedClient, dynamov1beta1.GroupVersion.WithKind("DynamoGraphDeployment"))
	if err != nil {
		return nil, nil, fmt.Errorf("get Dynamo v1beta1 Karta: %w", err)
	}
	roles, err := riresource.NewComponentFactoryFromObject(karta, raw).GetComponent("role")
	if err != nil {
		return nil, nil, fmt.Errorf("get Dynamo v1beta1 roles: %w", err)
	}
	instances, err := roles.GetExtractedInstances(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("extract Dynamo v1beta1 roles: %w", err)
	}
	ids, err := roles.GetInstanceIds(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("get Dynamo v1beta1 role instances: %w", err)
	}
	return instances, ids, nil
}
