/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mappers

import (
	"context"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	"k8s.io/apimachinery/pkg/fields"
)

// GetAllAccessRulesForProject returns the project-, department- and cluster-scoped access rules
// that apply to the given project.
func (m *AccessRulesProjectsMapper) GetAllAccessRulesForProject(ctx context.Context, project *kaiorg.Project) ([]v1alpha1.AccessRule, error) {
	return m.GetAllAccessRulesForOrgUnit(ctx, project.Name, project.Spec.Parent)
}

// GetAllAccessRulesForOrgUnit returns the project-, department- and cluster-scoped access rules that
// apply to a project with the given name under the given parent department.
func (m *AccessRulesProjectsMapper) GetAllAccessRulesForOrgUnit(ctx context.Context,
	projectName, departmentName string) ([]v1alpha1.AccessRule, error) {
	accessRulesForProject, err := m.getProjectScopedAccessRulesForProject(ctx, projectName)
	if err != nil {
		return []v1alpha1.AccessRule{}, err
	}

	accessRulesForDepartment, err := m.getDepartmentScopedAccessRulesForDepartment(ctx, departmentName)
	if err != nil {
		return []v1alpha1.AccessRule{}, err
	}

	clusterScopedAccessRules, err := m.getClusterScopedAccessRules(ctx)
	if err != nil {
		return []v1alpha1.AccessRule{}, err
	}

	allAccessRulesForProject := []v1alpha1.AccessRule{}
	allAccessRulesForProject = append(allAccessRulesForProject, accessRulesForProject...)
	allAccessRulesForProject = append(allAccessRulesForProject, accessRulesForDepartment...)
	allAccessRulesForProject = append(allAccessRulesForProject, clusterScopedAccessRules...)

	return allAccessRulesForProject, nil
}

func (m *AccessRulesProjectsMapper) getProjectScopedAccessRulesForProject(ctx context.Context, projectName string) ([]v1alpha1.AccessRule, error) {
	fieldSelector := fields.OneTermEqualSelector(common.AccessRuleProjectNameField, projectName)
	accessRulesList, err := common.ListAccessRulesWithFieldSelector(ctx, m.cachedClient, fieldSelector)
	if err != nil {
		return []v1alpha1.AccessRule{}, err
	}
	return accessRulesList.Items, nil
}

func (m *AccessRulesProjectsMapper) getDepartmentScopedAccessRulesForDepartment(ctx context.Context, departmentName string) ([]v1alpha1.AccessRule, error) {
	fieldSelector := fields.OneTermEqualSelector(common.AccessRuleDepartmentNameField, departmentName)
	accessRulesList, err := common.ListAccessRulesWithFieldSelector(ctx, m.cachedClient, fieldSelector)
	if err != nil {
		return []v1alpha1.AccessRule{}, err
	}
	return accessRulesList.Items, nil
}

func (m *AccessRulesProjectsMapper) getClusterScopedAccessRules(ctx context.Context) ([]v1alpha1.AccessRule, error) {
	fieldSelector := fields.OneTermEqualSelector(common.AccessRuleClusterField, common.ClusterNameFieldSelectorValue)
	accessRulesList, err := common.ListAccessRulesWithFieldSelector(ctx, m.cachedClient, fieldSelector)
	if err != nil {
		return []v1alpha1.AccessRule{}, err
	}
	return accessRulesList.Items, nil
}
