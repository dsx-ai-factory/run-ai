/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metrics

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	v1 "k8s.io/api/core/v1"
	"strconv"
)

const (
	nodeNameLabel = "node_name"
)

var (
	// Export only for MIG nodes!
	migModeGPUs = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Subsystem: commonconstants.NamespaceRunAi,
			Name:      "mig_mode_gpu_count",
			Help:      "Number of GPUs on MIG nodes",
		}, []string{nodeNameLabel, nodeLabel},
	)
)

// Moved from metrics_exporter/pkg/metrics/node_metrics.go
// Soon to be deleted, once Grafana dashboards are deprecated

func SetNodeMetrics(node *v1.Node) error {
	nodeName := node.Name

	gpuCount, err := getNodeGpuCount(node)
	if err != nil {
		return fmt.Errorf("failed to get gpu count for node %s, err: %w", nodeName, err)
	}

	ResetNodeMetrics(nodeName)

	migModeGPUs.WithLabelValues(nodeName, nodeName).Set(gpuCount)

	log.Info().Msgf("Updated Node %s metrics successfully", nodeName)
	return nil
}

func ResetNodeMetrics(nodeName string) {
	nodeLabelIdentifier := prometheus.Labels{nodeLabel: nodeName, nodeNameLabel: nodeName}
	migModeGPUs.Delete(nodeLabelIdentifier)
}

func IsNodeReady(node *v1.Node) bool {
	if node.Spec.Unschedulable {
		return false
	}

	for _, c := range node.Status.Conditions {
		if c.Type == v1.NodeReady {
			if c.Status != v1.ConditionTrue {
				return false
			}
		} else {
			if c.Status != v1.ConditionFalse {
				return false
			}
		}
	}

	return true
}

func getNodeGpuCount(node *v1.Node) (float64, error) {
	gpuCountLabelValue, found := node.Labels[commonconstants.GpuCountLabel]
	if !found {
		return 0, nil
	}

	gpuCount, err := strconv.Atoi(gpuCountLabelValue)
	if err != nil {
		return 0, fmt.Errorf("failed to parse label '%s', label value: %s, err: %w",
			commonconstants.GpuCountLabel, gpuCountLabelValue, err)
	}

	amdGpus, found := node.Status.Allocatable[commonconstants.AmdGpuResourceName]
	if found {
		gpuCount += int(amdGpus.Value())
	}
	return float64(gpuCount), nil
}

func GetMigModeGPUsMetric() *prometheus.GaugeVec {
	return migModeGPUs
}
