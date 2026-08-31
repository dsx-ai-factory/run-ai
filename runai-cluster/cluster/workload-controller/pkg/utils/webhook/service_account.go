/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package webhook

import (
	"context"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

var logUtils = logger.Log.WithName("utils.security")

func GetServiceAccount(ctx context.Context, k8sClient client.Client, name, namespace string) (*v1.ServiceAccount, error) {
	serviceAccount := &v1.ServiceAccount{}
	serviceAccountKey := types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}
	err := k8sClient.Get(ctx, serviceAccountKey, serviceAccount)
	if err != nil {
		logUtils.Error(err, "Failed getting ServiceAccount", "namespace", namespace, "name", name)
		return nil, err
	}

	return serviceAccount, nil
}
