/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_monitors

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
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
)

const (
	name        = "confidential-dcgm-exporter"
	metricsPath = "/metrics"
	metricsPort = 9400
)

func NewPodMonitor(operandName string, k8sReader client.Reader, runaiConfig runv1.RunaiConfig) *PodMonitor {
	return &PodMonitor{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: &runaiConfig,
	}
}

type PodMonitor struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func (p *PodMonitor) OperandName() string {
	return p.operandName
}

func (p *PodMonitor) Enabled() bool {
	return p.runaiConfig.IsConfidentialComputing()
}

func (p *PodMonitor) Object(ctx context.Context) (client.Object, error) {
	namespace := p.runaiConfig.Namespace

	podMonitor := &monitoringv1.PodMonitor{}
	err := p.k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, podMonitor)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	podMonitor.Name = name
	podMonitor.Namespace = namespace

	if !p.Enabled() {
		return podMonitor, nil
	}

	if podMonitor.Labels == nil {
		podMonitor.Labels = map[string]string{}
	}
	podMonitor.Labels["app"] = name

	podMonitor.Spec.NamespaceSelector = monitoringv1.NamespaceSelector{Any: true}
	podMonitor.Spec.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			commonconstants.ConfidentialComputingLabel: "true",
		},
	}
	podMonitor.Spec.PodMetricsEndpoints = []monitoringv1.PodMetricsEndpoint{
		{
			Path: metricsPath,
			RelabelConfigs: []monitoringv1.RelabelConfig{
				{
					Action:       "replace",
					SourceLabels: []monitoringv1.LabelName{"__meta_kubernetes_pod_ip"},
					TargetLabel:  "__address__",
					Replacement:  ptr.To(fmt.Sprintf("$1:%d", metricsPort)),
				},
				{
					Action: "labeldrop",
					Regex:  "container",
				},
			},
		},
	}

	return podMonitor, nil
}

func (p *PodMonitor) Status(ctx context.Context) error {
	if !p.Enabled() {
		return nil
	}

	podMonitor, err := p.Object(ctx)
	if err != nil {
		return err
	}

	if podMonitor.GetUID() == "" {
		return fmt.Errorf("pod monitor [%s] is missing", podMonitor.GetName())
	}

	return nil
}
