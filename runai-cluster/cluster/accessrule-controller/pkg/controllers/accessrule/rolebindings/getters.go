/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package rolebindings

import (
	"context"
	"github.com/rs/zerolog/log"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *RoleBindingsHandler) getRoleBindingIfExists(ctx context.Context,
	roleBindingName, namespace string,
	mustExist bool,
) (*rbacv1.RoleBinding, error) {
	roleBinding, err := r.getRoleBinding(ctx, roleBindingName, namespace)
	if err == nil {
		return roleBinding, nil
	}

	if !errors.IsNotFound(err) {
		log.Ctx(ctx).Error().Msgf("Error while getting RoleBinding <%s> in namespace <%s>, error: %s",
			roleBindingName, namespace, err.Error())
		return nil, err
	}

	if mustExist {
		log.Ctx(ctx).Error().Msgf("Role binding <%s/%s> doesn't exist, but it should",
			namespace, roleBindingName)
		return nil, err
	}

	return nil, nil
}

func (r *RoleBindingsHandler) getRoleBinding(ctx context.Context, roleBindingName, namespace string,
) (*rbacv1.RoleBinding, error) {
	roleBinding := &rbacv1.RoleBinding{}
	err := r.cachedClient.Get(ctx, client.ObjectKey{Name: roleBindingName, Namespace: namespace}, roleBinding)
	return roleBinding, err
}

func (r *RoleBindingsHandler) getClusterRoleBinding(ctx context.Context, clusterRoleBindingName string,
) (*rbacv1.ClusterRoleBinding, error) {
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	err := r.cachedClient.Get(ctx, client.ObjectKey{Name: clusterRoleBindingName}, clusterRoleBinding)
	return clusterRoleBinding, err
}

func (r *RoleBindingsHandler) getClusterRole(ctx context.Context, clusterRoleName string,
) (*rbacv1.ClusterRole, error) {
	clusterRole := &rbacv1.ClusterRole{}
	err := r.cachedClient.Get(ctx, client.ObjectKey{Name: clusterRoleName}, clusterRole)
	return clusterRole, err
}
