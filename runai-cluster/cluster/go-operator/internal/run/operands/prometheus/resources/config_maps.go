/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"fmt"

	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/consts"
	"github.com/samber/lo"
	"sigs.k8s.io/yaml"

	"github.com/openshift/cluster-monitoring-operator/pkg/manifests"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/prometheus"
	runv1_internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ConfigMap struct {
	operandName string
	k8sReader   client.Reader
	runaiConfig *runv1.RunaiConfig
}

func NewConfigMap(operandName string, k8sReader client.Reader, runaiConfig *runv1.RunaiConfig) *ConfigMap {
	return &ConfigMap{
		operandName: operandName,
		k8sReader:   k8sReader,
		runaiConfig: runaiConfig,
	}
}

func (cm *ConfigMap) OperandName() string {
	return cm.operandName
}

func (cm *ConfigMap) Object(ctx context.Context) (client.Object, error) {
	if cm.runaiConfig.Spec.Internal.Global.Platform.KubernetesFlavor != runv1_internal.KubernetesFlavorOpenShift {
		return nil, nil
	}

	var (
		name      = "cluster-monitoring-config"
		namespace = consts.OpenshiftMonitoringNamespace
		config    = cm.runaiConfig.Spec.Prometheus
	)

	configMap := &corev1.ConfigMap{}

	// Get the existing configMap if it exists to consume any cluster-set values
	err := cm.k8sReader.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, configMap)
	if err != nil && !errors.IsNotFound(err) {
		return nil, err
	}

	configMap.ObjectMeta.Name = name
	configMap.ObjectMeta.Namespace = namespace

	if configMap.Annotations == nil {
		configMap.Annotations = map[string]string{}
	}

	// Prevent the operator from deleting this configmap on finalization
	configMap.Annotations[runv1.DeletionPolicyAnnotationKey] = string(runv1.DeletionPolicyRetain)

	clusterMonitoringConfig := &manifests.ClusterMonitoringConfiguration{}

	if configMap.Data == nil {
		configMap.Data = map[string]string{}
	}

	if configMap.Data["config.yaml"] != "" {
		err := yaml.Unmarshal([]byte(configMap.Data["config.yaml"]), clusterMonitoringConfig)
		if err != nil {
			return nil, err
		}
	}

	// If not using openshift built-in prometheus, revert configuration to default
	if *cm.runaiConfig.Spec.Prometheus.Openshift.UseBuiltInPrometheus {
		applyPrometheusConfig(clusterMonitoringConfig, config)
	} else {
		resetPrometheusConfig(clusterMonitoringConfig, config)
	}

	yamlConfig, err := yaml.Marshal(clusterMonitoringConfig)
	if err != nil {
		return nil, err
	}

	configMap.Data["config.yaml"] = string(yamlConfig)

	return configMap, nil
}

func (cm *ConfigMap) Status(ctx context.Context) error {
	if !shouldUseOpenshiftBuiltInPrometheus(cm.runaiConfig) {
		return nil
	}

	configMap, err := cm.Object(ctx)
	if err != nil {
		return err
	}

	if configMap.GetUID() == "" {
		return fmt.Errorf("configmap [%s] is missing", configMap.GetName())
	}

	return nil
}

func (cm *ConfigMap) Enabled() bool {
	return true
}

func applyPrometheusConfig(clusterMonitoringConfig *manifests.ClusterMonitoringConfiguration, config *prometheus.Prometheus) {
	if clusterMonitoringConfig.PrometheusK8sConfig == nil {
		clusterMonitoringConfig.PrometheusK8sConfig = &manifests.PrometheusK8sConfig{}
	}

	if len(clusterMonitoringConfig.PrometheusK8sConfig.ExternalLabels) == 0 {
		clusterMonitoringConfig.PrometheusK8sConfig.ExternalLabels = map[string]string{}
	}

	for _, spec := range config.Spec.RemoteWrite {
		var insecureSkipVerify *bool
		ca := promv1.SecretOrConfigMap{}
		if spec.TLSConfig != nil {
			insecureSkipVerify = spec.TLSConfig.InsecureSkipVerify
			ca = spec.TLSConfig.CA
		}
		remoteWrite := manifests.RemoteWriteSpec{
			BasicAuth:     spec.BasicAuth,
			RemoteTimeout: string(utils.Deref(spec.RemoteTimeout)),
			Name:          utils.Deref(spec.Name),
			TLSConfig: &promv1.SafeTLSConfig{
				InsecureSkipVerify: insecureSkipVerify,
				CA:                 ca,
			},
			URL:                 spec.URL,
			WriteRelabelConfigs: spec.WriteRelabelConfigs,
		}
		exist, idx := findRemoteWrite(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite, remoteWrite)
		if !exist {
			clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite = append(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite, remoteWrite)
		} else {
			clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite[idx] = remoteWrite
		}
	}
}

// Reset the changes made to the configmap by the operator.
// Not resetting the external labels as they might be configured by the user and are not RunAI specific.
func resetPrometheusConfig(clusterMonitoringConfig *manifests.ClusterMonitoringConfiguration, config *prometheus.Prometheus) {
	remoteWriteNames := lo.Map(config.Spec.RemoteWrite, func(rws promv1.RemoteWriteSpec, _ int) string { return utils.Deref(rws.Name) })
	clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite = lo.Reject(clusterMonitoringConfig.PrometheusK8sConfig.RemoteWrite,
		func(rw manifests.RemoteWriteSpec, _ int) bool {
			return lo.Contains(remoteWriteNames, rw.Name)
		},
	)
}
