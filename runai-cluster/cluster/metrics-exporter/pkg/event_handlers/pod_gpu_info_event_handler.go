/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package event_handlers

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/metrics_exporter/pkg/exporters/pod_exporter"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/metrics_exporter/pkg/metrics"
)

type PodGpuInfoEventHandler struct {
	k8sClient  client.Client
	draCounter pod_exporter.DRAGpuCounter
}

// NewPodGpuInfoEventHandler's draCounter may be nil, in which case GPUs requested through DRA claims are not counted.
func NewPodGpuInfoEventHandler(k8sClient client.Client, draCounter pod_exporter.DRAGpuCounter) *PodGpuInfoEventHandler {
	return &PodGpuInfoEventHandler{k8sClient: k8sClient, draCounter: draCounter}
}

func (pe *PodGpuInfoEventHandler) ResetMetrics(identifier ObjectIdentifier) {
	metrics.ResetPodGpuInfoMetrics(identifier.Name, identifier.Namespace)
}

func (pe *PodGpuInfoEventHandler) SetMetrics(ctx context.Context, eventObject interface{}) error {
	pod, ok := eventObject.(*v1.Pod)
	if !ok {
		return fmt.Errorf("PodGpuInfoEventHandler: eventObject is not a pod")
	}

	podExporter := pod_exporter.NewPodExporter(ctx, pe.k8sClient, pod).WithDRAGpuCounter(pe.draCounter)
	return metrics.SetPodGpuInfoMetrics(podExporter)
}
