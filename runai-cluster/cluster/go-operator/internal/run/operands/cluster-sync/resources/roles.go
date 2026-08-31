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

	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

type RoleResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewRoleResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *RoleResource {
	return &RoleResource{operandName: operandName, k8sReader: k8sReader, runaiConfig: runaiConfig}
}

func (r *RoleResource) OperandName() string { return r.operandName }

func (r *RoleResource) Object(ctx context.Context) (client.Object, error) {
	namespace := r.runaiConfig.Namespace
	role := &rbacv1.Role{}
	err := r.k8sReader.Get(ctx, types.NamespacedName{Name: DeploymentName, Namespace: namespace}, role)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}
	role.ObjectMeta.Name = DeploymentName
	role.ObjectMeta.Namespace = namespace
	role.Rules = []rbacv1.PolicyRule{
		{APIGroups: []string{""}, Resources: []string{"configmaps"}, Verbs: []string{"get", "list", "watch"}},
	}
	return role, nil
}

func (r *RoleResource) Status(ctx context.Context) error {
	resource, err := r.Object(ctx)
	if err != nil {
		return err
	}
	if resource.GetUID() == "" {
		return fmt.Errorf("role [%s] is missing", resource.GetName())
	}
	return nil
}

func (r *RoleResource) Enabled() bool { return true }
