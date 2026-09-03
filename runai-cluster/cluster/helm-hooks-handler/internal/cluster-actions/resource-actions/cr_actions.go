/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resource_actions

import (
	"context"

	controller_runtime_utils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/utils/controller-runtime"

	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HandleClusterCR handles creating/updating RemoteConfig resources using the provided context.
func HandleClusterCR(ctx context.Context, k8s client.Client, headless bool) (*runaiv1.Cluster, error) {
	cluster, err := GetClusterCr(ctx, k8s)
	if err != nil {
		return nil, err
	}

	if !headless {
		err = applyRemoteConfigSyncCR(ctx, k8s, cluster)
		if err != nil {
			return nil, err
		}
		err = applyRemoteConfigCR(ctx, k8s, cluster)
		if err != nil {
			return nil, err
		}
	}

	return cluster, nil
}

// (old non-ctx wrapper removed)

func GetClusterCr(ctx context.Context, k8sClient client.Client) (*runaiv1.Cluster, error) {
	cluster := &runaiv1.Cluster{}
	clusterKey := client.ObjectKey{
		Name: runaiv1.ClusterName,
	}
	err := k8sClient.Get(ctx, clusterKey, cluster)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func applyRemoteConfigSyncCR(ctx context.Context, k8sClient client.Client, cluster *runaiv1.Cluster) error {
	return controller_runtime_utils.CreateOrUpdate(ctx, k8sClient, &runaiv1.RemoteConfigSync{
		ObjectMeta: metav1.ObjectMeta{
			Name: runaiv1.RemoteConfigSyncName,
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind:       runaiv1.ClusterKind,
					Name:       runaiv1.ClusterName,
					UID:        cluster.UID,
					APIVersion: runaiv1.GroupVersion.String(),
				},
			},
		},
	})
}

func applyRemoteConfigCR(ctx context.Context, k8sClient client.Client, cluster *runaiv1.Cluster) error {
	return controller_runtime_utils.CreateOrUpdate(ctx, k8sClient, &runaiv1.RemoteConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name: runaiv1.RemoteConfigName,
			OwnerReferences: []metav1.OwnerReference{
				{
					Kind:       runaiv1.ClusterKind,
					Name:       runaiv1.ClusterName,
					UID:        cluster.UID,
					APIVersion: runaiv1.GroupVersion.String(),
				},
			},
		},
	})
}
