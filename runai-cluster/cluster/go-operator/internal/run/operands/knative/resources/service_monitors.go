/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/consts"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	serviceNamespace = consts.KnativeNamespace
	runaiNamespace   = "runai"
	httpMetricsPort  = "http-metrics"
	metricsPath      = "/metrics"
	scrapeInterval   = "30s"
)

// ServiceMonitorResource represents a ServiceMonitor for a Knative component
type ServiceMonitorResource struct {
	operandName   string
	k8sReader     client.Reader
	runaiConfig   *runv1.RunaiConfig
	componentName string
}

// NewServiceMonitorResource creates a new ServiceMonitor resource for a specific Knative component
func NewServiceMonitorResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig, componentName string) *ServiceMonitorResource {
	return &ServiceMonitorResource{
		operandName:   operandName,
		k8sReader:     k8sReader,
		runaiConfig:   runaiConfig,
		componentName: componentName,
	}
}

// OperandName returns the name of the operand
func (r *ServiceMonitorResource) OperandName() string {
	return r.operandName
}

// Object returns the ServiceMonitor object
func (r *ServiceMonitorResource) Object(ctx context.Context) (client.Object, error) {
	serviceMonitor := &monitoringv1.ServiceMonitor{}

	// Get the existing service monitor if it exists to consume any cluster-set values
	err := r.k8sReader.Get(ctx, types.NamespacedName{
		Name:      fmt.Sprintf("%s-%s", serviceNamespace, r.componentName),
		Namespace: runaiNamespace,
	}, serviceMonitor)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	serviceMonitor.ObjectMeta.Namespace = runaiNamespace
	serviceMonitor.ObjectMeta.Name = fmt.Sprintf("%s-%s", serviceNamespace, r.componentName)
	serviceMonitor.ObjectMeta.Labels = map[string]string{
		"app": r.componentName,
	}

	serviceMonitor.Spec.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": r.componentName,
		},
	}
	serviceMonitor.Spec.NamespaceSelector = monitoringv1.NamespaceSelector{
		MatchNames: []string{
			serviceNamespace,
		},
	}
	serviceMonitor.Spec.Endpoints = []monitoringv1.Endpoint{
		{
			Port:            httpMetricsPort,
			Path:            metricsPath,
			Interval:        scrapeInterval,
			HonorLabels:     true,
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

// Status checks the status of the ServiceMonitor
func (r *ServiceMonitorResource) Status(ctx context.Context) error {
	resource, err := r.Object(ctx)
	if err != nil {
		return err
	}

	if resource.GetUID() == "" {
		return fmt.Errorf("service monitor [%s] is missing", resource.GetName())
	}

	return nil
}

// Enabled returns whether this ServiceMonitor should be created
func (r *ServiceMonitorResource) Enabled() bool {
	return shouldCreateServiceMonitors(r.runaiConfig)
}
