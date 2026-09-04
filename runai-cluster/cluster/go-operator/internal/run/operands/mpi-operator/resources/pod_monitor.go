/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	name             = "mpi-operator"
	serviceNamespace = "mpi-operator"
)

func PodMonitorForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (client.Object, error) {
	var (
		namespace  = runaiConfig.Namespace
		podName    = name
		portNumber = runaiConfig.Spec.MpiOperator.MonitorPortNumber
	)

	podMonitor := &monitoringv1.PodMonitor{}

	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, podMonitor)

	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	podMonitor.ObjectMeta.Namespace = namespace
	podMonitor.ObjectMeta.Name = podName
	podMonitor.ObjectMeta.Labels = map[string]string{
		"app": podName,
	}

	podMonitor.Spec.JobLabel = "app"
	podMonitor.Spec.NamespaceSelector = monitoringv1.NamespaceSelector{
		MatchNames: []string{
			serviceNamespace,
		},
	}
	podMonitor.Spec.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": podName,
		},
	}

	podMonitor.Spec.PodMetricsEndpoints = []monitoringv1.PodMetricsEndpoint{
		{

			MetricRelabelConfigs: []monitoringv1.RelabelConfig{
				{
					SourceLabels: []monitoringv1.LabelName{"__name__"},
					Replacement:  ptr.To("stats"),
					TargetLabel:  "type",
					Action:       "replace",
				},
			},
			RelabelConfigs: []monitoringv1.RelabelConfig{
				{
					SourceLabels: []monitoringv1.LabelName{"__address__"},
					Regex:        "([^:]+)(?:\\d+)?",
					TargetLabel:  "__address__",
					Replacement:  ptr.To(fmt.Sprintf("${1}:%d", *portNumber)),
					Action:       "replace",
				},
			},
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
