/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package rolebindings

import (
	"context"

	"github.com/rs/zerolog/log"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/rolebindingsutils"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

func (r *RoleBindingsHandler) handleExistingClusterRoleBinding(
	ctx context.Context,
	accessRule *v1alpha1.AccessRule,
	clusterRoleBindingName string) error {
	clusterRoleBinding, err := r.getClusterRoleBinding(ctx, clusterRoleBindingName)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Error while getting ClusterRoleBinding <%s>, error: %s",
			clusterRoleBindingName, err.Error())
		return err
	}

	return r.addToClusterRoleBindingSubjects(ctx, accessRule, clusterRoleBinding)
}

func (r *RoleBindingsHandler) handleExistingRoleBinding(
	ctx context.Context,
	accessRule *v1alpha1.AccessRule,
	roleBindingName string,
	namespace string) error {
	return r.createOrUpdateRoleBindingForAccessRuleIfNeeded(ctx, accessRule, roleBindingName, namespace,
		true, nil)
}

func (r *RoleBindingsHandler) handleRoleBindingIfExists(
	ctx context.Context,
	accessRule *v1alpha1.AccessRule,
	roleBindingName string,
	namespace string) error {
	return r.createOrUpdateRoleBindingForAccessRuleIfNeeded(ctx, accessRule, roleBindingName, namespace,
		false, nil)
}

func (r *RoleBindingsHandler) createOrUpdateRoleBindingForAccessRuleIfNeeded(
	ctx context.Context,
	accessRule *v1alpha1.AccessRule,
	roleBindingName string,
	namespace string,
	mustExist bool,
	project *kaiorg.Project) error {
	roleBinding, err := r.getRoleBindingIfExists(ctx, roleBindingName, namespace, mustExist)
	if err != nil {
		return err
	}

	if roleBinding != nil {
		err = r.addToRoleBindingSubjects(ctx, accessRule, roleBinding, project != nil)
		if err != nil {
			return err
		}
	}

	if project == nil {
		return nil
	}

	if roleBinding == nil {
		// only create a RoleBinding if it's a projects' RoleBinding
		err = r.createRoleBindingForAccessRule(ctx, roleBindingName, namespace, accessRule, project)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RoleBindingsHandler) createRoleBindingForAccessRule(
	ctx context.Context, roleBindingName, namespace string, accessRule *v1alpha1.AccessRule, project *kaiorg.Project) error {
	log.Ctx(ctx).Info().Msgf("Creating RoleBinding <%s> for access rule <%s> in namespace <%s>",
		roleBindingName, accessRule.Name, namespace)

	labels := map[string]string{
		rolebindingsutils.GetAccessRuleAddedSubjectLabelKey(accessRule, true): "true",
	}

	requiredRoleBinding := rolebindingsutils.GetRoleBindingObjectForAccessRuleInProject(
		roleBindingName, namespace, accessRule, project, labels,
	)
	err := r.cachedClient.Create(ctx, requiredRoleBinding)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("Failed creating RoleBinding <%s> for access rule <%s> in namespace <%s>, error: %s",
			roleBindingName, accessRule.Name, namespace, err.Error())
		return err
	}
	return nil
}
