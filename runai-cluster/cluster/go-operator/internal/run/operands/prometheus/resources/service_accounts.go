/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	serviceAccountName = "runai-prometheus"
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

func (r *ServiceAccountResource) OperandName() string {
	return r.operandName
}

func (r *ServiceAccountResource) Object(ctx context.Context) (client.Object, error) {
	var (
		namespace = r.runaiConfig.Namespace
	)

	serviceAccount := &v1.ServiceAccount{}

	err := r.k8sReader.Get(ctx, types.NamespacedName{
		Name:      serviceAccountName,
		Namespace: namespace,
	}, serviceAccount)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	serviceAccount.ObjectMeta.Name = serviceAccountName
	serviceAccount.ObjectMeta.Namespace = namespace

	return serviceAccount, nil
}

func (r *ServiceAccountResource) Status(ctx context.Context) error {
	serviceAccount, err := r.Object(ctx)
	if err != nil {
		return err
	}

	if serviceAccount.GetUID() == "" {
		return fmt.Errorf("service account [%s] is missing", serviceAccount.GetName())
	}

	return nil
}

func (r *ServiceAccountResource) Enabled() bool {
	return true
}
