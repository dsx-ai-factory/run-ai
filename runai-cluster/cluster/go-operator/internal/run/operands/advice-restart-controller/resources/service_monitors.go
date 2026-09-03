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

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

type ServiceMonitorResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewServiceMonitorResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *ServiceMonitorResource {
	return &ServiceMonitorResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (sm *ServiceMonitorResource) OperandName() string {
	return sm.operandName
}

func (sm *ServiceMonitorResource) Object(ctx context.Context) (client.Object, error) {
	namespace := sm.runaiConfig.Namespace

	serviceMonitor := &monitoringv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sm.operandName,
			Namespace: namespace,
		},
	}

	err := sm.k8sReader.Get(ctx, types.NamespacedName{
		Name:      sm.operandName,
		Namespace: namespace,
	}, serviceMonitor)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	serviceMonitor.SetLabels(map[string]string{
		"app": sm.operandName,
	})

	serviceMonitor.Spec.JobLabel = sm.operandName
	serviceMonitor.Spec.NamespaceSelector = monitoringv1.NamespaceSelector{
		MatchNames: []string{namespace},
	}
	serviceMonitor.Spec.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": sm.operandName,
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

func (sm *ServiceMonitorResource) Status(ctx context.Context) error {
	resource, err := sm.Object(ctx)
	if err != nil {
		return err
	}

	serviceMonitor := resource.(*monitoringv1.ServiceMonitor)
	if serviceMonitor.GetUID() == "" {
		return fmt.Errorf("service monitor [%s] is missing", serviceMonitor.GetName())
	}

	return nil
}

func (sm *ServiceMonitorResource) Enabled() bool {
	return true
}
