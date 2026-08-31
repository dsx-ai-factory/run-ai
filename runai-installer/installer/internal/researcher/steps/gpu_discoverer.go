/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package steps

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-installer/installer/internal/logging"
	corev1 "k8s.io/api/core/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	NvidiaGPUResource = "nvidia.com/gpu"
)

type K8sGPUDiscoverer struct {
	k8sClient k8sclient.Client
}

func NewK8sGPUDiscoverer(k8sClient k8sclient.Client) *K8sGPUDiscoverer {
	return &K8sGPUDiscoverer{k8sClient: k8sClient}
}

func (d *K8sGPUDiscoverer) GetClusterGPUCount(ctx context.Context) (float32, error) {
	logging.Infof("[RunAI Assets] Getting cluster GPU count from nodes...")

	nodeList := &corev1.NodeList{}
	if err := d.k8sClient.List(ctx, nodeList); err != nil {
		return 0, fmt.Errorf("failed to list nodes: %w", err)
	}

	var totalGPUs int64
	for _, node := range nodeList.Items {
		if gpuQuantity, exists := node.Status.Allocatable[corev1.ResourceName(NvidiaGPUResource)]; exists {
			totalGPUs += gpuQuantity.Value()
			logging.Infof("[RunAI Assets] Node '%s' has %d allocatable GPUs", node.Name, gpuQuantity.Value())
		}
	}

	logging.Infof("[RunAI Assets] Total cluster GPU count: %d", totalGPUs)
	return float32(totalGPUs), nil
}
