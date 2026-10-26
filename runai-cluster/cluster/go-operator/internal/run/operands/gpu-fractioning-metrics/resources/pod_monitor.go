/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

const (
	name        = "gpu-fractioning-metrics"
	metricsPort = "metrics"
	metricsPath = "/metrics"
)

// metricsd is a sidecar of the fractiond DaemonSet, which gpu-fractioning installs as a
// KAI sub-chart and therefore lands in the Run:ai release namespace.
var fractiondPodLabels = map[string]string{
	"app.kubernetes.io/component":  "fractiond",
	"app.kubernetes.io/managed-by": "gpu-fractioning",
}

func PodMonitorForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (client.Object, error) {
	namespace := runaiConfig.Namespace

	podMonitor := &monitoringv1.PodMonitor{}
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, podMonitor)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	podMonitor.Name = name
	podMonitor.Namespace = namespace
	podMonitor.Labels = map[string]string{
		"app": name,
	}

	podMonitor.Spec.NamespaceSelector = monitoringv1.NamespaceSelector{
		MatchNames: []string{namespace},
	}
	podMonitor.Spec.Selector = metav1.LabelSelector{
		MatchLabels: fractiondPodLabels,
	}
	podMonitor.Spec.PodMetricsEndpoints = []monitoringv1.PodMetricsEndpoint{
		{
			Port: ptr.To(metricsPort),
			Path: metricsPath,
			// metricsd carries the pod-to-GPU mapping as metric labels. Without this the
			// scrape overwrites pod and namespace with metricsd's own, and the mapping is lost.
			HonorLabels: true,
		},
	}

	return podMonitor, nil
}

func PodMonitorStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	podMonitor, err := PodMonitorForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	if podMonitor.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("pod monitor [%s] is missing",
			podMonitor.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}
