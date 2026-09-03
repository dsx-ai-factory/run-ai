/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	kaiconstants "github.com/kai-scheduler/KAI-scheduler/pkg/common/constants"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
)

// accountingServiceMonitorName is a second ServiceMonitor for the same
// nodepool-controller metrics endpoint, labeled for the KAI accounting Prometheus.
//
// runai_node_nodepool (the node->nodepool mapping the controller emits) is consumed
// by BOTH Prometheuses: the runai Prometheus (runai_node_info_enriched ->
// runai_free_gpu_count_per_node, which drives the "Free nodes by node pool" chart)
// and the KAI accounting Prometheus (runai-accounting / resources_per_nodepool).
// The kai.scheduler/accounting label routes a ServiceMonitor to exactly one
// Prometheus (NotIn on runai, In on accounting), so two ServiceMonitors are needed
// to reach both. The primary ServiceMonitor stays unlabeled (runai);
// this one carries the accounting label (KAI accounting).
const accountingServiceMonitorName = DeploymentName + "-accounting"

type ServiceMonitorResource struct {
	operandName         string
	serviceMonitorName  string
	withAccountingLabel bool
	k8sReader           client.Reader
	runaiConfig         *runv1.RunaiConfig
}

// NewServiceMonitorResource creates the primary ServiceMonitor, scraped by the runai
// Prometheus (no accounting label).
func NewServiceMonitorResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *ServiceMonitorResource {
	return &ServiceMonitorResource{
		operandName:         operandName,
		serviceMonitorName:  DeploymentName,
		withAccountingLabel: false,
		k8sReader:           k8sReader,
		runaiConfig:         runaiConfig,
	}
}

// NewAccountingServiceMonitorResource creates a second ServiceMonitor labeled for the
// KAI accounting Prometheus, so runai_node_nodepool reaches it as well.
func NewAccountingServiceMonitorResource(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *ServiceMonitorResource {
	return &ServiceMonitorResource{
		operandName:         operandName,
		serviceMonitorName:  accountingServiceMonitorName,
		withAccountingLabel: true,
		k8sReader:           k8sReader,
		runaiConfig:         runaiConfig,
	}
}

func (s *ServiceMonitorResource) OperandName() string {
	return s.operandName
}

func (s *ServiceMonitorResource) Object(ctx context.Context) (client.Object, error) {
	namespace := s.runaiConfig.Namespace

	serviceMonitor := &monitoringv1.ServiceMonitor{}

	err := s.k8sReader.Get(ctx, types.NamespacedName{
		Name:      s.serviceMonitorName,
		Namespace: namespace,
	}, serviceMonitor)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	serviceMonitor.ObjectMeta.Name = s.serviceMonitorName
	serviceMonitor.ObjectMeta.Namespace = namespace
	labels := map[string]string{
		"app": DeploymentName,
	}
	if s.withAccountingLabel {
		labels[kaiconstants.DefaultAccountingLabelKey] = kaiconstants.DefaultAccountingLabelValue
	}
	serviceMonitor.ObjectMeta.Labels = labels

	serviceMonitor.Spec.JobLabel = DeploymentName
	serviceMonitor.Spec.NamespaceSelector = monitoringv1.NamespaceSelector{
		MatchNames: []string{
			"runai",
		},
	}
	serviceMonitor.Spec.Selector = metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": DeploymentName,
		},
	}
	serviceMonitor.Spec.Endpoints = []monitoringv1.Endpoint{
		{
			Port:            s.runaiConfig.Spec.NodePoolController.Metrics.PortName,
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

func (s *ServiceMonitorResource) Status(ctx context.Context) error {
	resource, err := s.Object(ctx)
	if err != nil {
		return err
	}

	if resource.GetUID() == "" {
		return fmt.Errorf("service monitor [%s] is missing", resource.GetName())
	}

	return nil
}

func (s *ServiceMonitorResource) Enabled() bool {
	return true
}
