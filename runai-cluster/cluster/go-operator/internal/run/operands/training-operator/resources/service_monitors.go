/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ServiceMonitorForRunaiConfig(ctx context.Context, k8sReader client.Reader,
	runaiConfig *runv1.RunaiConfig) (client.Object, error) {
	var (
		namespace   = runaiConfig.Namespace
		serviceName = name
	)

	serviceMonitor := &monitoringv1.ServiceMonitor{}

	// Get the existing service monitor if it exists to consume any cluster-set values
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      serviceName,
		Namespace: namespace,
	}, serviceMonitor)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	serviceMonitor.ObjectMeta.Namespace = namespace
	serviceMonitor.ObjectMeta.Name = serviceName
	serviceMonitor.ObjectMeta.Labels = map[string]string{
		"app": serviceName,
	}

	serviceMonitor.Spec.JobLabel = serviceName
	serviceMonitor.Spec.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": serviceName,
		},
	}
	serviceMonitor.Spec.NamespaceSelector = monitoringv1.NamespaceSelector{
		MatchNames: []string{
			serviceNamespace,
		},
	}
	serviceMonitor.Spec.Endpoints = []monitoringv1.Endpoint{
		{
			Port:            metricPortName,
			BearerTokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
			MetricRelabelConfigs: []monitoringv1.RelabelConfig{
				{
					Replacement: ptr.To("stats"),
					TargetLabel: "type",
					Action:      "replace",
				},
			},
		},
	}

	return serviceMonitor, nil
}

func ServiceMonitorStatus(ctx context.Context,
	k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) (operands.Status, error) {
	serviceMonitor, err := ServiceMonitorForRunaiConfig(ctx, k8sReader, runaiConfig)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	if serviceMonitor.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("service monitor [%s] is missing",
			serviceMonitor.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}
