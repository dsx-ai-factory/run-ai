/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resource_actions

import (
	"context"
	ra "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/cluster-actions"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetRunaiConfigDynamicResourceInterface(dynamicClient dynamic.Interface) dynamic.ResourceInterface {
	return dynamicClient.Resource(ra.RunaiConfigResource).Namespace(runaiv1.DefaultNamespace)
}

func RunaiConfigExists(ctx context.Context, k8s dynamic.Interface) (bool, error) {
	_, err := GetRunaiConfigDynamicResourceInterface(k8s).Get(ctx, "runai", metav1.GetOptions{})
	if err == nil {
		return true, nil
	}
	if errors.IsNotFound(err) {
		return false, nil
	}
	return false, err
}

func GetExistingRunaiConfig(k8s client.Client) (*runaiv1.RunaiConfig, error) {
	runaiConfig := DefaultRunaiConfig()

	err := k8s.Get(context.Background(), client.ObjectKeyFromObject(runaiConfig), runaiConfig)
	if err != nil {
		return nil, err
	}

	return runaiConfig, nil
}

func DefaultRunaiConfig() *runaiv1.RunaiConfig {
	return &runaiv1.RunaiConfig{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: runaiv1.DefaultNamespace,
			Name:      runaiv1.DefaultName,
		},
	}
}
