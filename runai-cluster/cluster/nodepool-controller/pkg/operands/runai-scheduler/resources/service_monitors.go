/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/nodepool-controller/pkg/operands"
	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"

	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ServiceMonitorForNodePool(ctx context.Context, k8sReader client.Reader,
	nodePool *v1alpha1.NodePool, params *common.NodePoolControllerParams, operandName string) (client.Object, error) {
	schedulerName := SchedulerBaseOperandName()
	var (
		name      = fmt.Sprintf("%s-%s", schedulerName, operandName)
		namespace = config.Get().SchedulerNamespace
		appName   = name
	)

	serviceMonitor := &monitorv1.ServiceMonitor{}
	// Get the existing serviceMonitor if it exists to consume any cluster-set values
	err := k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, serviceMonitor)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	serviceMonitor.ObjectMeta.Name = name
	serviceMonitor.ObjectMeta.Namespace = namespace
	serviceMonitor.ObjectMeta.Labels = map[string]string{
		"app": schedulerName,
	}
	serviceMonitor.Spec.JobLabel = schedulerName
	serviceMonitor.Spec.NamespaceSelector = monitorv1.NamespaceSelector{
		MatchNames: []string{namespace},
	}
	serviceMonitor.Spec.Selector = v1.LabelSelector{
		MatchLabels: map[string]string{
			"app": appName,
		},
	}
	serviceMonitor.Spec.Endpoints = []monitorv1.Endpoint{
		{
			Port:            "http-metrics",
			BearerTokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
		},
	}
	return serviceMonitor, nil
}

func ServiceMonitorStatus(ctx context.Context, k8sReader client.Reader,
	nodePool *v1alpha1.NodePool, params *common.NodePoolControllerParams, operandName string) (operands.Status, error) {
	serviceMonitor, err := ServiceMonitorForNodePool(ctx, k8sReader, nodePool, params, operandName)
	if err != nil {
		return operands.NotReadyStatus(), err
	}

	if serviceMonitor.GetUID() == "" {
		return operands.NotReadyStatus(fmt.Sprintf("service monitor [%s] is missing",
			serviceMonitor.GetName())), nil
	}

	return operands.ReadyStatus(), nil
}
