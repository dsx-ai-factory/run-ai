/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pre_hooks

import (
	"context"
	"log"

	"go.uber.org/multierr"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	clusterRoleBindingsToDelete = []string{
		"runai-research-manager",
	}
)

// deleteOldPermissionsPreInstallPreUpgrade deleted leftovers of old RBAC we had.
// ClusterRoles+ClusterRoleBindings+Roles are deleted automatically by helm;
// But, is happens that "runai-research-manager" ClusterRoleBinding existed in the old
// chart as well as in the new, with different RoleRef, and since RoleRef cannot be changes after the ClusterRoleBinding
// exists, we need to manually delete this ClusterRoleBinding and it will be created again with the new chart.
func deleteOldPermissionsPreInstallPreUpgrade(k8sClient client.Client) error {
	ctx := context.Background()
	var err error

	innerErr := deleteClusterRoleBindings(ctx, k8sClient, clusterRoleBindingsToDelete)
	if innerErr != nil {
		err = multierr.Append(err, innerErr)
	}

	return err
}

func deleteClusterRoleBindings(ctx context.Context, k8sClient client.Client, clusterRoleBindingNames []string) error {
	clusterRoleBindingObj := rbacv1.ClusterRoleBinding{}

	var err error
	for _, clusterRoleBindingName := range clusterRoleBindingNames {
		clusterRoleBindingObj.Name = clusterRoleBindingName
		innerErr := k8sClient.Delete(ctx, &clusterRoleBindingObj)
		if innerErr != nil && !errors.IsNotFound(innerErr) {
			log.Printf("Failed deleting ClusterRoleBinding <%s>, error: %s",
				clusterRoleBindingName, innerErr.Error())
			err = multierr.Append(err, innerErr)
		}
	}

	return err
}
