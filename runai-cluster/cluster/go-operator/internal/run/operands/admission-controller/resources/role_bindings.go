/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func RoleBindingForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (client.Object, error) {
	var (
		roleBindingName = "runai-admission-controller"
		namespace       = runaiConfig.Namespace
	)

	roleBinding := &rbacv1.RoleBinding{}

	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      roleBindingName,
		Namespace: namespace,
	}, roleBinding)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	roleBinding.ObjectMeta.Name = roleBindingName
	roleBinding.ObjectMeta.Namespace = namespace

	roleBinding.RoleRef = rbacv1.RoleRef{
		Kind:     "Role",
		APIGroup: rbacv1.SchemeGroupVersion.Group,
		Name:     "runai-admission-controller",
	}

	roleBinding.Subjects = []rbacv1.Subject{
		{
			Kind:      "ServiceAccount",
			Name:      "runai-admission-controller",
			Namespace: "runai",
		},
	}

	return roleBinding, nil
}

func RoleBindingStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	roleBinding, err := RoleBindingForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	if roleBinding.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("role binding [%s] is missing", roleBinding.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}
