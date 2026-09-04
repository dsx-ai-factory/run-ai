/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"fmt"
	"log"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/metrics_exporter/pkg/exporters/pod_exporter"
)

var (
	podGpuInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: runaiNamespace,
			Name:      "pod_gpu_info",
			Help:      "Pod phase",
		}, []string{"pod_uuid", "pod_namespace", "pod_name", "gpu_index", "gpu", "gpu_instance_id", "gpu_num_devices"},
	)
)

func SetPodGpuInfoMetrics(podExporter *pod_exporter.PodExporter) error {
	name := podExporter.GetName()
	namespace := podExporter.GetNamespace()
	uuid := podExporter.GetUUID()

	gpuIndexes, err := podExporter.GetGpuIndex()
	if err != nil {
		return fmt.Errorf("could not get GpuIndex for pod %s/%s, error: %w", namespace, name, err)
	}
	gpuInstanceId, err := podExporter.GetGpuInstanceId()
	if err != nil {
		return fmt.Errorf("could not get GpuInstanceId for pod %s/%s, error: %w", namespace, name, err)
	}
	gpuNumDevices, err := podExporter.GetGpuNumDevices()
	if err != nil {
		return fmt.Errorf("could not get GpuNumDevices for pod %s/%s, error: %w", namespace, name, err)
	}

	ResetPodGpuInfoMetrics(name, namespace)
	for _, gpuIndex := range strings.Split(gpuIndexes, ",") {
		podGpuInfo.WithLabelValues(uuid, namespace, name, gpuIndex, gpuIndex, gpuInstanceId, gpuNumDevices).Set(1)
	}

	log.Printf("Updated Pod %s/%s gpu info metrics successfully", namespace, name)
	return nil
}

func ResetPodGpuInfoMetrics(podName, podNamespace string) {
	podIdentifier := prometheus.Labels{"pod_namespace": podNamespace, "pod_name": podName}
	podGpuInfo.DeletePartialMatch(podIdentifier)
}
