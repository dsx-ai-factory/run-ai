/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	health = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "workload_metrics_health_total",
			Help: "Health metric",
		},
	)
)

func InitMetrics() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(health)
	metrics.Registry.MustRegister(podInfo, podRequestedGpus, podRequestedGPUMemory, podLimitGpus, podLimitGPUMemory,
		podAllocatedGPUs, podAllocatedGPUMemory, podAllocatedCPUs, podAllocatedCpuMemory, podElements)
	metrics.Registry.MustRegister(workloadInfo)
	metrics.Registry.MustRegister(podGroupRequestedMigDevices)
	metrics.Registry.MustRegister(migModeGPUs)

	health.Inc()
}
