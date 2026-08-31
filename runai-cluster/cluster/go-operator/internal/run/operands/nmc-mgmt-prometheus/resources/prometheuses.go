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
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/consts"
	prom_common "github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/operands/common/prometheus"
	controller_utils "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
)

const prometheusInstanceName = "nmc-mgmt-prometheus"

type PrometheusResource struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runaiv1.RunaiConfig
}

func NewPrometheusResource(operandName string, k8sReader client.Reader, runaiConfig *runaiv1.RunaiConfig) *PrometheusResource {
	return &PrometheusResource{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (r *PrometheusResource) OperandName() string {
	return r.operandName
}

func (r *PrometheusResource) Object(ctx context.Context) (client.Object, error) {
	var (
		name      = prometheusInstanceName
		namespace = r.runaiConfig.Namespace
	)

	prom := &monitoringv1.Prometheus{}

	err := r.k8sReader.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, prom)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	prom.Namespace = namespace
	prom.Name = name

	monitoredNamespaces, err := getMonitoredNamespaces(ctx, r.runaiConfig, r.k8sReader)
	if err != nil {
		return nil, err
	}

	prom_common.SetCommonPrometheusFields(prom, r.runaiConfig, monitoredNamespaces, true, false)

	setNmcMgmtSpecificFields(prom, r)

	return prom, nil
}

func setNmcMgmtSpecificFields(prom *monitoringv1.Prometheus, r *PrometheusResource) {
	prom.Spec.ServiceAccountName = utils.GetNMCServiceAccountNameOrDefault(r.runaiConfig, "runai-prometheus")
	prom.Spec.AdditionalArgs = []monitoringv1.Argument{
		{
			Name: "web.enable-remote-write-receiver",
		},
	}
	for i := range prom.Spec.RemoteWrite {
		if prom.Spec.RemoteWrite[i].BasicAuth != nil {
			prom.Spec.RemoteWrite[i].BasicAuth.Password.Name = rwAuthSecretName
			prom.Spec.RemoteWrite[i].BasicAuth.Username.Name = rwAuthSecretName
		}
	}
}

func (r *PrometheusResource) Status(ctx context.Context) error {
	resource, err := r.Object(ctx)
	if err != nil {
		return err
	}

	prometheus := resource.(*monitoringv1.Prometheus)

	prometheusStatus := prometheus.Status

	if prometheusStatus.AvailableReplicas < *prometheus.Spec.Replicas {
		return fmt.Errorf("prometheus [%s] is not rolled out", prometheus.GetName())
	}

	return nil
}

func (r *PrometheusResource) Enabled() bool {
	return true
}

func getMonitoredNamespaces(ctx context.Context, runaiConfig *runaiv1.RunaiConfig, k8sReader client.Reader) ([]string, error) {
	monitoredNamespaces := []string{
		runaiv1.DefaultNamespace,
		consts.OpenshiftMonitoringNamespace,
		consts.MonitoringNamespace,
		consts.KubeSystemForRKEMonitoringNamespace,
		consts.DGXCSystemNamespace,
	}

	if !runaiConfig.ShouldSkipDependencyChecks() {
		prometheusOperatorNamespace, err := controller_utils.GetPrometheusOperatorNamespace(ctx, k8sReader)
		if err != nil {
			return nil, fmt.Errorf("failed to get prometheus operator namespace: %w", err)
		}

		monitoredNamespaces = append(monitoredNamespaces, prometheusOperatorNamespace)
	}

	return monitoredNamespaces, nil
}
