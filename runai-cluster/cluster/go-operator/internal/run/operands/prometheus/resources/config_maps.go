/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/internal/run/consts"
	promv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/samber/lo"
	"sigs.k8s.io/yaml"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/operands/utils"
	runv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/prometheus"
	runv1_internal "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	"github.com/openshift/cluster-monitoring-operator/pkg/manifests"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

	clusterMonitoringConfig := map[string]interface{}{}

	if configMap.Data == nil {
		configMap.Data = map[string]string{}
	}

	if configMap.Data["config.yaml"] != "" {
		err := yaml.Unmarshal([]byte(configMap.Data["config.yaml"]), &clusterMonitoringConfig)
		if err != nil {
			return nil, fmt.Errorf("unmarshal OpenShift monitoring configuration: %w", err)
		}
		if clusterMonitoringConfig == nil {
			clusterMonitoringConfig = map[string]interface{}{}
		}
	}

	prometheusK8sConfig, err := getPrometheusK8sConfig(clusterMonitoringConfig,
		*cm.runaiConfig.Spec.Prometheus.Openshift.UseBuiltInPrometheus)
	if err != nil {
		return nil, err
	}

	remoteWrites, err := getRemoteWrites(prometheusK8sConfig)
	if err != nil {
		return nil, err
	}

	if *cm.runaiConfig.Spec.Prometheus.Openshift.UseBuiltInPrometheus {
		remoteWrites, err = applyPrometheusConfig(remoteWrites, config)
	} else {
		remoteWrites, err = resetPrometheusConfig(remoteWrites, config)
	}
	if err != nil {
		return nil, err
	}

	if prometheusK8sConfig != nil {
		if err := setRemoteWrites(prometheusK8sConfig, remoteWrites); err != nil {
			return nil, err
		}
		if err := unstructured.SetNestedMap(clusterMonitoringConfig, prometheusK8sConfig, "prometheusK8s"); err != nil {
			return nil, fmt.Errorf("set prometheusK8s configuration: %w", err)
		}
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

func applyPrometheusConfig(remoteWrites []interface{}, config *prometheus.Prometheus) ([]interface{}, error) {
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
		unstructuredRemoteWrite, err := remoteWriteToUnstructured(remoteWrite)
		if err != nil {
			return nil, err
		}

		exist, idx, err := findUnstructuredRemoteWrite(remoteWrites, remoteWrite.Name)
		if err != nil {
			return nil, err
		}
		if !exist {
			remoteWrites = append(remoteWrites, unstructuredRemoteWrite)
		} else {
			remoteWrites[idx] = unstructuredRemoteWrite
		}
	}

	return remoteWrites, nil
}

// Reset the changes made to the configmap by the operator.
// Not resetting the external labels as they might be configured by the user and are not RunAI specific.
func resetPrometheusConfig(remoteWrites []interface{}, config *prometheus.Prometheus) ([]interface{}, error) {
	remoteWriteNames := lo.Map(config.Spec.RemoteWrite, func(rws promv1.RemoteWriteSpec, _ int) string { return utils.Deref(rws.Name) })
	remainingRemoteWrites := make([]interface{}, 0, len(remoteWrites))
	for _, remoteWrite := range remoteWrites {
		name, err := unstructuredRemoteWriteName(remoteWrite)
		if err != nil {
			return nil, err
		}
		if !lo.Contains(remoteWriteNames, name) {
			remainingRemoteWrites = append(remainingRemoteWrites, remoteWrite)
		}
	}

	return remainingRemoteWrites, nil
}

func getPrometheusK8sConfig(clusterMonitoringConfig map[string]interface{}, create bool) (map[string]interface{}, error) {
	prometheusK8sConfig, found, err := unstructured.NestedMap(clusterMonitoringConfig, "prometheusK8s")
	if err != nil {
		return nil, fmt.Errorf("get prometheusK8s configuration: %w", err)
	}
	if !found {
		if !create {
			return nil, nil
		}

		return map[string]interface{}{}, nil
	}

	return prometheusK8sConfig, nil
}

func getRemoteWrites(prometheusK8sConfig map[string]interface{}) ([]interface{}, error) {
	if prometheusK8sConfig == nil {
		return nil, nil
	}

	remoteWrites, found, err := unstructured.NestedSlice(prometheusK8sConfig, "remoteWrite")
	if err != nil {
		return nil, fmt.Errorf("get prometheusK8s.remoteWrite configuration: %w", err)
	}
	if !found {
		return nil, nil
	}

	return remoteWrites, nil
}

func setRemoteWrites(prometheusK8sConfig map[string]interface{}, remoteWrites []interface{}) error {
	if len(remoteWrites) == 0 {
		unstructured.RemoveNestedField(prometheusK8sConfig, "remoteWrite")
		return nil
	}

	if err := unstructured.SetNestedSlice(prometheusK8sConfig, remoteWrites, "remoteWrite"); err != nil {
		return fmt.Errorf("set prometheusK8s.remoteWrite configuration: %w", err)
	}

	return nil
}

func remoteWriteToUnstructured(remoteWrite manifests.RemoteWriteSpec) (map[string]interface{}, error) {
	data, err := json.Marshal(remoteWrite)
	if err != nil {
		return nil, fmt.Errorf("marshal updated OpenShift remote write configuration: %w", err)
	}

	value := map[string]interface{}{}
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("unmarshal updated OpenShift remote write configuration: %w", err)
	}

	return value, nil
}

func findUnstructuredRemoteWrite(remoteWrites []interface{}, name string) (bool, int, error) {
	for index, remoteWrite := range remoteWrites {
		remoteWriteName, err := unstructuredRemoteWriteName(remoteWrite)
		if err != nil {
			return false, -1, err
		}
		if remoteWriteName == name {
			return true, index, nil
		}
	}

	return false, -1, nil
}

func unstructuredRemoteWriteName(remoteWrite interface{}) (string, error) {
	remoteWriteMap, ok := remoteWrite.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("prometheusK8s.remoteWrite entry must be an object")
	}

	name, found, err := unstructured.NestedString(remoteWriteMap, "name")
	if err != nil {
		return "", fmt.Errorf("get prometheusK8s.remoteWrite name: %w", err)
	}
	if !found {
		return "", nil
	}

	return name, nil
}
