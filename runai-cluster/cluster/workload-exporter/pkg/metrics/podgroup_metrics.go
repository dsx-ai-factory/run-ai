/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"context"
	"encoding/json"
	"fmt"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-exporter/pkg/metrics/data"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	podGroupRequestedMigDevices = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: commonconstants.NamespaceRunAi,
			Name:      "requested_mig_devices",
			Help:      "pod group requested mig devices",
		}, []string{"queue_name", "pod_group_name", "pod_group_uuid", "workload_name", "workload_type",
			"mig_device_type", "project"},
	)
)

func SetPodGroupMetrics(ctx context.Context, podGroup *kaiv2alpha2.PodGroup, readerClient client.Reader) error {
	name := podGroup.Name
	namespace := podGroup.Namespace

	topOwnerMetadata, err := data.GetTopOwnerMetadata(podGroup)
	if err != nil {
		return fmt.Errorf("could not find top owner for pod group %s/%s, error: %w", namespace, name, err)
	}

	projectName, err := data.GetProjectNameForPodGroup(ctx, podGroup, readerClient)
	if err != nil {
		return fmt.Errorf("could not find project name for pod group %s/%s, error: %w", namespace, name, err)
	}

	queueName := podGroup.Spec.Queue
	workloadType := data.GetWorkloadTypeFromPodGroup(ctx, podGroup, readerClient)
	// PodGroupUID is expected to be the jobUID
	uid := topOwnerMetadata.UID

	requestedMigDevices, err := getRequestedMigDevices(podGroup)
	if err != nil {
		// MIG devices annotation is optional - only set if present
		log.Debug().Msgf("No MIG devices for pod group %s/%s: %v", namespace, name, err)
		return nil
	}

	ResetPodGroupMetrics(name, namespace)

	for deviceName, count := range requestedMigDevices {
		podGroupRequestedMigDevices.WithLabelValues(queueName, name, uid, topOwnerMetadata.Name, string(workloadType),
			deviceName, projectName).Set(float64(count))
	}

	log.Info().Msgf("Updated PodGroup %s/%s mig metrics successfully", namespace, name)
	return nil
}

func ResetPodGroupMetrics(podGroupName, namespace string) {
	labels := prometheus.Labels{"pod_group_name": podGroupName}
	podGroupRequestedMigDevices.DeletePartialMatch(labels)
	log.Debug().Msgf("Reset PodGroup %s/%s metrics", namespace, podGroupName)
}

func getRequestedMigDevices(podGroup *kaiv2alpha2.PodGroup) (map[string]int64, error) {
	valString, found := podGroup.Annotations[commonconstants.PodGroupRequestedMigDevicesAnnotation]
	if !found {
		return nil, fmt.Errorf("could not find annotation '%s'", commonconstants.PodGroupRequestedMigDevicesAnnotation)
	}
	var migDevices map[string]int64
	err := json.Unmarshal([]byte(valString), &migDevices)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal value '%s' of annotation %s: %w",
			valString, commonconstants.PodGroupRequestedMigDevicesAnnotation, err)
	}
	return migDevices, nil
}

func GetPodGroupRequestedMigDevicesMetric() *prometheus.GaugeVec {
	return podGroupRequestedMigDevices
}
