/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package service_monitors

import (
	"context"
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	controller_utils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

func NewServiceMonitor(operandName string, k8sReader client.Reader, runaiConfig runv1.RunaiConfig) *ServiceMonitor {
	return &ServiceMonitor{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: &runaiConfig,
	}
}

type ServiceMonitor struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func (s *ServiceMonitor) OperandName() string {
	return s.operandName
}

func (s *ServiceMonitor) Object(ctx context.Context) (client.Object, error) {
	var (
		global = s.runaiConfig.Spec.Global

		name      = "nvidia-dcgm-exporter"
		namespace = s.runaiConfig.Namespace
	)

	serviceMonitor := &monitoringv1.ServiceMonitor{}

	// Get the existing service monitor if it exists to consume any cluster-set values
	err := s.k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, serviceMonitor)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	serviceMonitor.ObjectMeta.Name = name
	serviceMonitor.ObjectMeta.Namespace = namespace

	if serviceMonitor.Labels == nil {
		serviceMonitor.Labels = map[string]string{}
	}

	serviceMonitor.Labels["app"] = name
	serviceMonitor.Labels["app.kubernetes.io/component"] = name
	serviceMonitor.Labels["app.kubernetes.io/instance"] = name
	serviceMonitor.Labels["app.kubernetes.io/name"] = name

	dcgmExporterService, err := controller_utils.GetDCGMExporterService(ctx, s.k8sReader)
	if err != nil {
		return nil, err
	}

	serviceMonitor.Spec.Selector.MatchLabels = dcgmExporterService.Labels

	serviceMonitor.Spec.NamespaceSelector = monitoringv1.NamespaceSelector{
		MatchNames: []string{
			dcgmExporterService.Namespace,
		},
	}

	serviceMonitor.Spec.Endpoints = []monitoringv1.Endpoint{
		{
			Port:     dcgmExporterService.Spec.Ports[0].Name,
			Path:     "/metrics",
			Interval: monitoringv1.Duration(global.NvidiaDcgmExporter.ServiceMonitor.Interval),
		},
	}

	return serviceMonitor, nil
}

func (s *ServiceMonitor) Status(ctx context.Context) error {
	serviceMonitor, err := s.Object(ctx)
	if err != nil {
		return err
	}

	if serviceMonitor.GetUID() == "" {
		return fmt.Errorf("service monitor [%s] is missing",
			serviceMonitor.GetName())
	}

	return nil
}

func (s *ServiceMonitor) Enabled() bool {
	return *s.runaiConfig.Spec.Global.NvidiaDcgmExporter.ServiceMonitor.Enabled
}
