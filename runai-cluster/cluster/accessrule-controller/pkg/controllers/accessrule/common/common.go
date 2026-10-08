/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	"k8s.io/apimachinery/pkg/fields"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	FinalizerName = "accessrule.run.ai/finalize"

	ProjectsParentNameField       = "spec.parent"
	AccessRuleProjectNameField    = "spec.projectType.scopeName"
	AccessRuleDepartmentNameField = "spec.departmentType.scopeName"
	AccessRuleClusterField        = "spec.clusterType"
	ClusterNameFieldSelectorValue = "cluster-name"

	RunaiAgentServiceAccountName                = constants.RunaiAgentName
	RunaiAccessRuleControllerServiceAccountName = constants.RunaiAccessRuleControllerName
	RunaiOperatorServiceAccountName             = constants.RunaiOperatorName

	AccessRuleAddedSubjectToProjectsRBPrefix = "runai/project-accessrule-"
	AccessRuleAddedSubjectPrefix             = "runai/accessrule-"

	ClusterRoleNotFoundMessage                  = "clusterrole <%s> not found"
	ClusterRoleFailedGettingMessage             = "failed getting clusterrole <%s> for access rule"
	ScopeObjectNotFoundMessage                  = "%s <%s> not found"
	ScopeObjectFailedGettingMessage             = "failed listing projects for access rule"
	ScopeObjectFailedGettingRoleBindingsMessage = "failed getting all rolebindings for access rule"
	ScopeObjectFailedGettingNamespacesMessage   = "failed getting projects namespaces for access rule"
	FailedReconcileProjectsMessage              = "failed reconciling permissions for projects: <%v>"
	FailedDeletingUnnecessaryMessage            = "failed deleting unnecessary permissions in namespaces: <%v>"
	FailedReconcileRunaiMessage                 = "failed reconciling permissions in runai namespace"
	FailedReconcileClusterRoleBindingMessage    = "failed reconciling cluster-scoped permissions"
	FailedReconcileClusterScopedRunaiMessage    = "failed reconciling cluster-scoped permissions in runai namespace"
	FailedDeleteRunaiMessage                    = "failed deleting permissions in runai namespace"
	FailedDeleteClusterRoleBindingMessage       = "failed deleting cluster-scoped permissions"
	FailedDeleteClusterScopedRunaiMessage       = "failed deleting cluster-scoped permissions in runai namespace"
	FailedHandlingDeletionInNamespacesMessage   = "failed deleting permissions in namespaces: <%v>"
	FailedReconcileSystemAdminNamespaceMessage  = "failed reconciling cluster-scoped permissions in %s namespace"
)

var (
	AccessRuleScopeTypeToObj = map[v1alpha1.ScopeType]string{
		v1alpha1.ProjectScope:    "project",
		v1alpha1.DepartmentScope: "department",
	}
)

func ListAccessRulesWithFieldSelector(
	ctx context.Context, cachedClient client.Client, fieldSelector fields.Selector) (*v1alpha1.AccessRuleList, error) {
	accessRules := &v1alpha1.AccessRuleList{}
	err := cachedClient.List(ctx, accessRules, &client.ListOptions{FieldSelector: fieldSelector})
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed listing access rules with field selector <%s>, error: %s",
			fieldSelector.String(), err.Error())
		return nil, err
	}
	return accessRules, nil
}

func ListProjectsWithFieldSelector(
	ctx context.Context, cachedClient client.Client, fieldSelector fields.Selector) (*kaiorg.ProjectList, error) {
	projects := &kaiorg.ProjectList{}
	err := cachedClient.List(ctx, projects, &client.ListOptions{FieldSelector: fieldSelector})
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed listing kai projects with field selector <%s>, error: %s",
			fieldSelector.String(), err.Error())
		return nil, err
	}
	return projects, nil
}
