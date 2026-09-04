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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ConfigMapForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (client.Object, error) {
	var (
		name      = runaiConfig.Name + "-" + runv1.ReconciliationTriggerConfigmapNameSuffix
		namespace = runaiConfig.Namespace
	)

	configMap := &corev1.ConfigMap{}

	// Get the existing configMap if it exists to consume any cluster-set values
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, configMap)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	configMap.ObjectMeta.Name = name
	configMap.ObjectMeta.Namespace = namespace
	configMap.ObjectMeta.Annotations = map[string]string{
		runv1.ReconciliationTriggerRequestKey: "",
	}

	return configMap, nil
}

func ConfigMapStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	configMap, err := ConfigMapForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	if configMap.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("configmap [%s] is missing",
			configMap.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}
