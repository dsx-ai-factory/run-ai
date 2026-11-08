/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package gpuinventory

import (
	"context"
	"fmt"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RunInCluster uses only the collector's Kubernetes identity; it never loads
// cluster-sync's control-plane credentials or starts backend synchronization.
func RunInCluster(ctx context.Context) error {
	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		return fmt.Errorf("NODE_NAME is required for GPU identity collection")
	}
	config, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("load collector Kubernetes configuration: %w", err)
	}
	config.Timeout = 20 * time.Second
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("register core Kubernetes types: %w", err)
	}
	kubeClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("create collector Kubernetes client: %w", err)
	}
	NewCollector(kubeClient, nodeName, "/host/sys", "/host/nvidia-gpus").Run(ctx)
	return nil
}
