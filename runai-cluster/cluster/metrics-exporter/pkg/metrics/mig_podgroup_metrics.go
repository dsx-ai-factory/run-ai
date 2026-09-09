/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"fmt"
	"log"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/metrics_exporter/pkg/exporters/podgroup_exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	podGroupRequestedMigDevices = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: runaiNamespace,
			Name:      "requested_mig_devices",
			Help:      "pod group requested mig devices",
		}, []string{"queue_name", "pod_group_name", "pod_group_uuid", "workload_name", "workload_type",
			"mig_device_type", "project"},
	)
)

func SetMigPodGroupMetrics(podGroupExporter *podgroup_exporter.PodGroupExporter) error {
	name := podGroupExporter.GetName()
	namespace := podGroupExporter.GetNamespace()

	requestedMigDevices, found, err := podGroupExporter.GetRequestedMigDevices()
	if err != nil {
		return fmt.Errorf("could not get requested mig devices for pod group %s/%s, error: %w",
			namespace, name, err)
	}
	if !found {
		return nil
	}

	topOwnerMetadata, err := podGroupExporter.GetTopOwnerMetadata()
	if err != nil {
		return fmt.Errorf("could not find top owner for pod group %s/%s, error: %w", namespace, name, err)
	}
	projectName, err := podGroupExporter.GetProjectName()
	if err != nil {
		return fmt.Errorf("could not find project name for pod group %s/%s, error: %w", namespace, name, err)
	}
	queueName := podGroupExporter.GetQueueName()
	workloadType := podGroupExporter.GetWorkloadType()
	// PodGroupUID is expected to be the jobUID
	uid := topOwnerMetadata.UID

	for deviceName, count := range requestedMigDevices {
		podGroupRequestedMigDevices.WithLabelValues(queueName, name, uid, topOwnerMetadata.Name, workloadType,
			deviceName, projectName).Set(float64(count))
	}

	log.Printf("Updated Mig PodGroup %s/%s metrics successfully", namespace, name)
	return nil
}

func ResetMigPodGroupMetrics(podGroupName string) {
	labels := prometheus.Labels{"pod_group_name": podGroupName}
	podGroupRequestedMigDevices.DeletePartialMatch(labels)
}
