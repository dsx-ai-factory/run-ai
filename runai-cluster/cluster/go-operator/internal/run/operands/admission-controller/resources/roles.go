/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

func RoleForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (client.Object, error) {
	var (
		roleName  = "runai-admission-controller"
		namespace = runaiConfig.Namespace
	)

	role := &rbacv1.Role{}

	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      roleName,
		Namespace: namespace,
	}, role)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	role.ObjectMeta.Name = roleName
	role.ObjectMeta.Namespace = namespace

	role.Rules = []rbacv1.PolicyRule{
		{
			APIGroups: []string{""},
			Resources: []string{"configmaps"},
			Verbs:     []string{"get", "list", "watch"},
		},
	}

	return role, nil
}

func RoleStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	role, err := RoleForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	if role.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("role [%s] is missing", role.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}
