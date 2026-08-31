/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ServiceAccountForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (client.Object, error) {
	var (
		name      = "runai-mps"
		namespace = runaiConfig.Namespace
	)

	serviceAccount := &v1.ServiceAccount{}

	// Get the existing serviceAccount if it exists to consume any cluster-set values
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, serviceAccount)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	serviceAccount.ObjectMeta.Name = name
	serviceAccount.ObjectMeta.Namespace = namespace

	return serviceAccount, nil
}

func ServiceAccountStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	serviceAccount, err := ServiceAccountForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.ReadyStatus(), err
	}

	if serviceAccount.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("service account [%s] is missing", serviceAccount.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}
