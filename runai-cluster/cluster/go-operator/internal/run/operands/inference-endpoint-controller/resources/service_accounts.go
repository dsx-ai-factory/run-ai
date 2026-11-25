/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceAccountResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewServiceAccountResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *ServiceAccountResource {
	return &ServiceAccountResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (s *ServiceAccountResource) OperandName() string {
	return s.operandName
}

func (s *ServiceAccountResource) Object(ctx context.Context) (client.Object, error) {
	namespace := s.runaiConfig.Namespace

	serviceAccount := &v1.ServiceAccount{}

	err := s.k8sReader.Get(ctx, types.NamespacedName{
		Name:      DeploymentName,
		Namespace: namespace,
	}, serviceAccount)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	serviceAccount.Name = DeploymentName
	serviceAccount.Namespace = namespace

	return serviceAccount, nil
}

func (s *ServiceAccountResource) Status(ctx context.Context) error {
	resource, err := s.Object(ctx)
	if err != nil {
		return err
	}

	serviceAccount := resource.(*v1.ServiceAccount)
	if serviceAccount.GetUID() == "" {
		return fmt.Errorf("service account [%s] is missing", serviceAccount.GetName())
	}

	return nil
}

func (s *ServiceAccountResource) Enabled() bool {
	return true
}
