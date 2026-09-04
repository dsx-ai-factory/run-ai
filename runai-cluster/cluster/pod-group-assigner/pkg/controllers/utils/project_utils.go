/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/pod-group-assigner/pkg/config"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetProjectNameOfNamespace(ctx context.Context, k8sClient client.Client, namespace string) (string, error) {
	namespaceObj := &corev1.Namespace{}
	namespaceKey := types.NamespacedName{Name: namespace}

	if err := k8sClient.Get(ctx, namespaceKey, namespaceObj); err != nil {
		return "", fmt.Errorf("failed to get namespace <%s>, error: %s", namespace, err.Error())
	}

	projectName, found := namespaceObj.Labels[config.Config().NamespaceProjectLabelKey]
	if !found {
		return "", fmt.Errorf("can't find project label for namespace <%s>", namespace)
	}

	return projectName, nil
}
