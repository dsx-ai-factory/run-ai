/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	runaiv2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
)

func GetProject(k8sClient client.Client, name string) (*runaiv2.Project, error) {
	key := client.ObjectKey{
		Name: name,
	}
	project := runaiv2.Project{}
	err := k8sClient.Get(context.Background(), key, &project)
	if err != nil {
		return nil, err
	}
	return &project, nil
}
