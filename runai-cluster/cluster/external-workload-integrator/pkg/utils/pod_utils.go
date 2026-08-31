/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	commonconsts "github.com/run-ai/runai/runai-cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListWorkloadPods returns all pods belonging to a given ExternalWorkload,
// matched by the ClusterTopOwnerUidLabel selector.
func ListWorkloadPods(ctx context.Context, cachedClient client.Reader, ew *v1alpha1.ExternalWorkload) ([]corev1.Pod, error) {
	podList := &corev1.PodList{}
	err := cachedClient.List(ctx, podList,
		client.InNamespace(ew.Namespace),
		client.MatchingLabels{commonconsts.ClusterTopOwnerUidLabel: string(ew.Spec.OriginalWorkloadUID)})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}
