/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resource_upgrades

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

const (
	configMapName = "cluster-monitoring-config"
	namespace     = "openshift-monitoring"
)

var (
	cmGVK = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "configmaps",
	}
)

type ocpMonitoringConfigmapUpgrade struct{}

func NewOcpMonitoringConfigmapUpgrade() *ocpMonitoringConfigmapUpgrade {
	return &ocpMonitoringConfigmapUpgrade{}
}

func (o *ocpMonitoringConfigmapUpgrade) Upgrade(ctx context.Context, clients *clients.Clients) error {
	unstructuredCM, err := clients.RemoteClusterDynamicClient.Resource(cmGVK).
		Namespace(namespace).
		Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		if k8serr.IsNotFound(err) {
			// configmap not found, nothing to do
			return nil
		} else {
			return fmt.Errorf("failed to get configmap %s, error: %w", configMapName, err)
		}
	}

	configMap := &corev1.ConfigMap{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredCM.Object, configMap)
	if err != nil {
		return fmt.Errorf("failed to convert configmap %s, error: %w", configMapName, err)
	}

	if configMap.Data["config.yaml"] == "" {
		// configmap is empty, nothing to do
		return nil
	}

	clusterMonitoringConfig := &map[string]interface{}{}
	err = yaml.Unmarshal([]byte(configMap.Data["config.yaml"]), clusterMonitoringConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal configmap %s, error: %w", configMapName, err)
	}

	promK8S := (*clusterMonitoringConfig)["prometheusK8s"]
	if promK8S == nil {
		// no prometheus config, nothing to do
		return nil
	}

	promK8SMap := promK8S.(map[string]interface{})
	externalLabels := promK8SMap["externalLabels"]
	if externalLabels == nil {
		// no external labels, nothing to do
		return nil
	}

	externalLabelsMap := externalLabels.(map[string]interface{})

	// The following labels are reserved by the cluster-monitoring-operator and should not be applied
	// by users.
	// OCP added client-side validation for these labels, and will reject the configmap if they are present.
	// https://github.com/openshift/cluster-monitoring-operator/pull/2604
	labelsToDelete := []string{"prometheus", "prometheus_replica", "cluster"}
	shouldUpdate := false
	for _, label := range labelsToDelete {
		if _, ok := externalLabelsMap[label]; ok {
			shouldUpdate = true
			delete(externalLabelsMap, label)
		}
	}

	if !shouldUpdate {
		// no changes were made, update not needed
		return nil
	}

	updatedMonitoringConfig, err := yaml.Marshal(clusterMonitoringConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal configmap %s, error: %w", configMapName, err)
	}

	configMap.Data["config.yaml"] = string(updatedMonitoringConfig)

	updatedUnstructuredCM, err := runtime.DefaultUnstructuredConverter.ToUnstructured(configMap)
	if err != nil {
		return fmt.Errorf("failed to convert configmap to unstructured: %w", err)
	}

	_, err = clients.RemoteClusterDynamicClient.Resource(cmGVK).
		Namespace(namespace).
		Update(ctx, &unstructured.Unstructured{Object: updatedUnstructuredCM},
			metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update configmap %s, error: %w", configMapName, err)
	}

	return nil
}
