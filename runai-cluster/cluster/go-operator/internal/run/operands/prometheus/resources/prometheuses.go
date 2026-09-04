/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runaipublic "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/runaipublic"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/consts"
	prom_common "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/operands/common/prometheus"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/utils"
	controller_utils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	outils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	runv1_internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
)

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
		name = "runai"
	)

	prom := &monitoringv1.Prometheus{}

	err := r.k8sReader.Get(ctx, types.NamespacedName{
		Namespace: name,
		Name:      name,
	}, prom)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	prom.Namespace = name
	prom.Name = name

	monitoredNamespaces, err := getMonitoredNamespaces(ctx, r.runaiConfig, r.k8sReader)
	if err != nil {
		return nil, err
	}

	setExternalLabels := !r.runaiConfig.Spec.Internal.Global.NMC.IsEnabled()
	enableOTLPReceiver, err := shouldEnableOTLPReceiver(ctx, r.runaiConfig, r.k8sReader)
	if err != nil {
		return nil, err
	}
	prom_common.SetCommonPrometheusFields(prom, r.runaiConfig, monitoredNamespaces, setExternalLabels, enableOTLPReceiver)

	prom.Spec.ServiceAccountName = serviceAccountName

	if r.runaiConfig.Spec.Internal.Global.NMC.IsEnabled() {
		r.setNmcTenantSpecificFields(prom)
	}

	prom.Spec.Affinity = outils.MergeAffinities(prom.Spec.Affinity, r.runaiConfig.Spec.Global.Affinity,
		map[string]string{"prometheus": prom.Name}, r.runaiConfig.Spec.Global.RequireDefaultPodAntiAffinity)

	return prom, nil
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
	return r.runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor != runv1_internal.KubernetesFlavorOpenShift || !*r.runaiConfig.Spec.Prometheus.Openshift.UseBuiltInPrometheus
}

func PrometheusVersion(config *runaiv1.RunaiConfig) (string, error) {
	status, err := getBuildInfo(config)
	if err != nil {
		return "", fmt.Errorf("failed to get prometheus build info: %w", err)
	}

	version := getVersion(status)
	if len(version) == 0 {
		return "", fmt.Errorf("failed to get prometheus version")
	}
	return version, nil
}

func shouldEnableOTLPReceiver(ctx context.Context, runaiConfig *runaiv1.RunaiConfig, k8sReader client.Reader) (bool, error) {
	// Enabling only if the feature flag is enabled
	if !ptr.Deref(runaiConfig.Spec.Prometheus.Config.EnableOTLPReceiver, false) {
		return false, nil
	}

	// Enabling only if Knative is installed, as this feature is required for Knative Serving request metrics
	handler := runaipublic.NewRunaiPublicHandler(ctx, k8sReader)
	knativeAvailable, _, err := handler.IsKnativeAvailable()
	if err != nil {
		return false, err
	}
	return knativeAvailable, nil
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

	projectsNamespacesNames, err := utils.ListProjectsNamespacesNames(ctx, k8sReader)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects namespaces: %w", err)
	}

	monitoredNamespaces = append(monitoredNamespaces, projectsNamespacesNames...)
	return monitoredNamespaces, nil
}

func getBuildInfo(config *runaiv1.RunaiConfig) (string, error) {
	if config.Spec.Global.PrometheusService == nil {
		return "", fmt.Errorf("no prometheus in the cluster")
	}
	address := config.Spec.Global.PrometheusService.Address
	res, err := http.Get(address + "/api/v1/status/buildinfo")
	if err != nil {
		return "", fmt.Errorf("failed to connect to prometheus in the cluster: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("prometheus returned status %d", res.StatusCode)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read prometheus response: %w", err)
	}

	return string(resBody), nil
}

func getVersion(status string) string {
	var jsonMap map[string]interface{}
	err := json.Unmarshal([]byte(status), &jsonMap)
	if err != nil || len(jsonMap) == 0 || jsonMap["status"].(string) != "success" {
		return ""
	}
	data := jsonMap["data"].(map[string]interface{})
	return data["version"].(string)
}

func (r *PrometheusResource) setNmcTenantSpecificFields(prom *monitoringv1.Prometheus) {
	prom.Spec.RemoteWrite = []monitoringv1.RemoteWriteSpec{
		{
			URL: fmt.Sprintf("https://%s/api/v1/write", prom_common.MgmtPrometheusDomain),
			TLSConfig: &monitoringv1.TLSConfig{
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					CA: promv1.SecretOrConfigMap{
						Secret: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: caSecretName,
							},
							Key: caSecretKey,
						},
					},
				},
			},
		},
	}
	prom.Spec.Tolerations = []corev1.Toleration{
		{
			Key:      "components.gke.io/gke-managed-components",
			Operator: "Exists",
		},
		{
			Key:      "dedicated",
			Operator: "Equal",
			Value:    "system-workload",
		},
	}

	prom.Spec.Affinity = &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      "dedicated",
								Operator: "In",
								Values:   []string{"system-workload"},
							},
						},
					},
				},
			},
		},
	}

	// Remove external labels to avoid unnecessary cardinality
	prom.Spec.ExternalLabels = map[string]string{}
	prom.Spec.ExternalLabels["__replica__"] = ""
	prom.Spec.ReplicaExternalLabelName = ptr.To("")
	prom.Spec.PrometheusExternalLabelName = ptr.To("")
}
