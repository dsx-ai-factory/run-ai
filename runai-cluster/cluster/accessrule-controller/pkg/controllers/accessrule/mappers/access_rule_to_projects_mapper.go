/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mappers

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AccessRulesProjectsMapper struct {
	cachedClient client.Client
}

func NewAccessRulesProjectsMapper(cachedClient client.Client) *AccessRulesProjectsMapper {
	return &AccessRulesProjectsMapper{
		cachedClient: cachedClient,
	}
}

func (m *AccessRulesProjectsMapper) GetProjectsOfAccessRuleScope(
	ctx context.Context, accessRule *v1alpha1.AccessRule) ([]kaiorg.Project, error) {
	switch accessRule.Spec.ScopeType {
	case v1alpha1.ProjectScope:
		return m.getProjectsForProjectScopedAccessRule(ctx, accessRule)
	case v1alpha1.DepartmentScope:
		return m.getProjectsForDepartmentScopedAccessRule(ctx, accessRule)
	case v1alpha1.ClusterScope:
		return m.getProjectsForClusterScopedAccessRule(ctx, accessRule)
	}

	return []kaiorg.Project{}, fmt.Errorf("error: access role <%s> has unrecognized scope type: %v",
		accessRule.Name, accessRule.Spec.ScopeType)
}

func (m *AccessRulesProjectsMapper) getProjectsForProjectScopedAccessRule(
	ctx context.Context, accessRule *v1alpha1.AccessRule) ([]kaiorg.Project, error) {
	project := &kaiorg.Project{}
	err := m.cachedClient.Get(ctx, types.NamespacedName{Name: accessRule.Spec.ScopeName}, project)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed getting kai project <%s> for access rule <%s>, error: %s",
			accessRule.Spec.ScopeName, accessRule.Name, err.Error())
		return []kaiorg.Project{}, err
	}

	return []kaiorg.Project{*project}, nil
}

func (m *AccessRulesProjectsMapper) getProjectsForDepartmentScopedAccessRule(
	ctx context.Context, accessRule *v1alpha1.AccessRule) ([]kaiorg.Project, error) {
	// first, validate the department exists
	department := &kaiorg.Department{}
	err := m.cachedClient.Get(ctx, types.NamespacedName{Name: accessRule.Spec.ScopeName}, department)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed getting kai department <%s> for access rule <%s>, error: %s",
			accessRule.Spec.ScopeName, accessRule.Name, err.Error())
		return []kaiorg.Project{}, err
	}

	fieldSelector := fields.OneTermEqualSelector(common.ProjectsParentNameField, accessRule.Spec.ScopeName)
	projectsList, err := common.ListProjectsWithFieldSelector(ctx, m.cachedClient, fieldSelector)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed listing kai projects with department <%s> for access rule <%s>",
			accessRule.Spec.ScopeName, accessRule.Name)
		return []kaiorg.Project{}, err
	}
	return projectsList.Items, nil
}

func (m *AccessRulesProjectsMapper) getProjectsForClusterScopedAccessRule(
	ctx context.Context, accessRule *v1alpha1.AccessRule) ([]kaiorg.Project, error) {
	projects, err := m.listProjects(ctx)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed listing all kai projects for cluster-scoped access rule <%s>",
			accessRule.Name)
		return []kaiorg.Project{}, err
	}

	return projects.Items, nil
}

func (m *AccessRulesProjectsMapper) listProjects(ctx context.Context) (*kaiorg.ProjectList, error) {
	projects := &kaiorg.ProjectList{}
	err := m.cachedClient.List(ctx, projects)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed listing kai projects, error: %s",
			err.Error())
		return nil, err
	}
	return projects, nil
}
