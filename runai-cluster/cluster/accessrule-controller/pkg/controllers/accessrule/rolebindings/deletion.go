/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package rolebindings

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/accessrule-controller/pkg/controllers/accessrule/rolebindingsutils"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	rbacv1 "k8s.io/api/rbac/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
)

func (r *RoleBindingsHandler) DeleteFromExistingProjectRoleBinding(
	ctx context.Context,
	accessRule *v1alpha1.AccessRule,
	roleBinding *rbacv1.RoleBinding) error {
	roleBindingAfterUpdate, err := r.deleteFromRoleBindingSubjects(ctx, accessRule, roleBinding, true)
	if err != nil {
		return err
	}

	if len(roleBindingAfterUpdate.Subjects) == 0 {
		err = r.deleteRoleBinding(ctx, roleBindingAfterUpdate, accessRule.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *RoleBindingsHandler) DeleteFromExtraRoleBindingsForAccessRule(
	ctx context.Context, accessRule *v1alpha1.AccessRule) (string, error) {
	var err error
	statusMessages := []string{}

	innerErr := r.deleteFromRunaiNamespaceRoleBinding(ctx, accessRule)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
		statusMessages = append(statusMessages, common.FailedDeleteRunaiMessage)
	}

	innerErr = r.deleteFromClusterRoleBinding(ctx, accessRule)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
		statusMessages = append(statusMessages, common.FailedDeleteClusterRoleBindingMessage)
	}

	innerErr = r.deleteFromClusterScopedRunaiNamespaceRoleBinding(ctx, accessRule)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
		statusMessages = append(statusMessages, common.FailedDeleteClusterScopedRunaiMessage)
	}

	innerErr = r.deleteExtraRoleBindingForSystemAdmin(ctx, accessRule)
	if innerErr != nil {
		err = multierror.Append(err, innerErr)
		statusMessages = append(statusMessages, common.FailedDeleteClusterScopedRunaiMessage)
	}

	return strings.Join(statusMessages, ", "), err
}

func (r *RoleBindingsHandler) deleteFromRunaiNamespaceRoleBinding(
	ctx context.Context, accessRule *v1alpha1.AccessRule) error {
	runaiNamespaceRoleBindingName := rolebindingsutils.GetRunaiNamespaceRoleBindingName(accessRule.Spec.RoleName)

	return r.deleteFromRoleBindingIfNeeded(ctx, accessRule, runaiNamespaceRoleBindingName,
		runv1.DefaultNamespace)
}

func (r *RoleBindingsHandler) deleteFromClusterRoleBinding(ctx context.Context, accessRule *v1alpha1.AccessRule,
) error {
	clusterRoleBindingName := rolebindingsutils.GetClusterRoleBindingName(accessRule.Spec.RoleName)

	return r.deleteFromClusterRoleBindingIfNeeded(ctx, accessRule, clusterRoleBindingName)
}

func (r *RoleBindingsHandler) deleteFromClusterScopedRunaiNamespaceRoleBinding(
	ctx context.Context, accessRule *v1alpha1.AccessRule) error {
	clusterScopedRunaiNamespaceRoleBindingName :=
		rolebindingsutils.GetClusterScopedRunaiNamespaceRoleBindingName(accessRule.Spec.RoleName)

	return r.deleteFromRoleBindingIfNeeded(ctx, accessRule, clusterScopedRunaiNamespaceRoleBindingName,
		runv1.DefaultNamespace)
}

func (r *RoleBindingsHandler) deleteExtraRoleBindingForSystemAdmin(ctx context.Context, accessRule *v1alpha1.AccessRule) error {
	if accessRule.Spec.RoleName != systemAdministratorRoleName {
		return nil
	}

	var err error
	for _, namespace := range namespacesForAccessSystemAdmin {
		if innerErr := r.deleteFromRoleBindingIfNeeded(ctx, accessRule, systemAdminAccessToLogsRoleBindingName, namespace); innerErr != nil {
			log.Ctx(ctx).Debug().Msgf("RoleBindingsHandler.DeleteFromExtraRoleBindingsForAccessRule failed to delete RoleBinding in %s namespace, err: %+v", namespace, innerErr)
			err = multierror.Append(err, innerErr)
			continue
		}

		log.Ctx(ctx).Debug().Msgf("RoleBindingsHandler.DeleteFromExtraRoleBindingsForAccessRule deleted RoleBinding in %s namespace", namespace)
	}

	return err
}

func (r *RoleBindingsHandler) deleteFromRoleBindingIfNeeded(
	ctx context.Context, accessRule *v1alpha1.AccessRule,
	roleBindingName string,
	namespace string) error {
	roleBinding, err := r.getRoleBindingIfExists(ctx, roleBindingName, namespace, false)
	if err != nil {
		return err
	}

	if roleBinding == nil {
		log.Ctx(ctx).Debug().Msgf("RoleBinding <%s/%s> doesn't exist -"+
			" nothing to do for permissions removal for accessrule <%s>", namespace, roleBindingName, accessRule.Name)
		return nil
	}

	_, err = r.deleteFromRoleBindingSubjects(ctx, accessRule, roleBinding, false)
	return err
}

func (r *RoleBindingsHandler) deleteFromClusterRoleBindingIfNeeded(
	ctx context.Context, accessRule *v1alpha1.AccessRule,
	clusterRoleBindingName string) error {
	clusterRoleBinding, err := r.getClusterRoleBinding(ctx, clusterRoleBindingName)
	if err != nil {
		if !k8s_errors.IsNotFound(err) {
			log.Ctx(ctx).Error().Msgf("Error while getting ClusterRoleBinding <%s>, error: %s",
				clusterRoleBindingName, err.Error())
			return err
		}
		log.Ctx(ctx).Debug().Msgf("ClusterRoleBinding <%s> doesn't exist -"+
			" nothing to do for deletion of accessrule <%s>", clusterRoleBindingName, accessRule.Name)
		return nil
	}

	err = r.deleteFromClusterRoleBindingSubjects(ctx, accessRule, clusterRoleBinding)
	if err != nil {
		return err
	}
	return nil
}

func (r *RoleBindingsHandler) deleteRoleBinding(ctx context.Context, roleBinding *rbacv1.RoleBinding, accessRuleName string) error {
	log.Ctx(ctx).Info().Msgf("Deleting RoleBinding <%s> in namespace <%s> for accessrule <%s> -"+
		" it's subjects are empty", roleBinding.Name, roleBinding.Namespace, accessRuleName)
	err := r.cachedClient.Delete(ctx, roleBinding)
	if err != nil {
		log.Ctx(ctx).Error().Msgf("For access rule <%s>, failed to delete unnecessary RoleBinding <%s> "+
			"in namespace <%s>, error: %s", accessRuleName, roleBinding.Name, roleBinding.Namespace, err.Error())
		return err
	}
	return nil
}
