/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resource_upgrades

import (
	"context"
	"fmt"
	"testing"

	"github.com/openshift/cluster-monitoring-operator/pkg/manifests"
	scheme_init "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/clients"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	faked "k8s.io/client-go/dynamic/fake"
	"sigs.k8s.io/yaml"
)

var (
	badExternalLabels = map[string]string{
		"prometheus":         "test",
		"prometheus_replica": "test",
		"cluster":            "test",
	}

	goodExternalLabels = map[string]string{
		"something_good": "test",
	}
)

func createClients() *clients.Clients {
	return &clients.Clients{
		LocalClusterDynamicClient:  faked.NewSimpleDynamicClient(scheme_init.Scheme()),
		RemoteClusterDynamicClient: faked.NewSimpleDynamicClient(scheme_init.Scheme()),
	}
}

func ocpMonitoringConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      configMapName,
		},
	}
}

func getConfigMap(clients *clients.Clients) (*corev1.ConfigMap, error) {
	unstCM, err := clients.RemoteClusterDynamicClient.Resource(cmGVK).
		Namespace(namespace).
		Get(context.Background(), configMapName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap: %w", err)
	}

	cm := &corev1.ConfigMap{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstCM.Object, cm)
	if err != nil {
		return nil, fmt.Errorf("failed to convert configmap to typed: %w", err)
	}

	return cm, nil
}

func createUnstructuredConfigMap(clients *clients.Clients, cm *corev1.ConfigMap) error {
	unstructuredCM, err := runtime.DefaultUnstructuredConverter.ToUnstructured(cm)
	if err != nil {
		return fmt.Errorf("failed to convert configmap to unstructured: %w", err)
	}

	_, err = clients.RemoteClusterDynamicClient.Resource(cmGVK).
		Namespace(namespace).
		Create(context.Background(),
			&unstructured.Unstructured{Object: unstructuredCM}, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create configmap using the dynamic client: %w", err)
	}

	return nil
}

func createConfigMap(clients *clients.Clients, externalLabels map[string]string) error {
	cm := ocpMonitoringConfigMap()
	conf := manifests.ClusterMonitoringConfiguration{}
	conf.PrometheusK8sConfig = &manifests.PrometheusK8sConfig{}
	conf.PrometheusK8sConfig.ExternalLabels = externalLabels

	confYAML, err := yaml.Marshal(conf)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	cm.Data = map[string]string{
		"config.yaml": string(confYAML),
	}

	return createUnstructuredConfigMap(clients, cm)
}

func createConfigMapWithBadExternalLabels(clients *clients.Clients) error {
	return createConfigMap(clients, badExternalLabels)
}

func createConfigMapWithGoodExternalLabels(clients *clients.Clients) error {
	return createConfigMap(clients, goodExternalLabels)
}

func createConfigMapWithNoExternalLabels(clients *clients.Clients) error {
	return createConfigMap(clients, nil)
}

func createConfigMapWithNoK8sPrometheusConfig(clients *clients.Clients) error {
	cm := ocpMonitoringConfigMap()
	conf := manifests.ClusterMonitoringConfiguration{}
	conf.PrometheusK8sConfig = nil

	confYAML, err := yaml.Marshal(conf)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	cm.Data = map[string]string{
		"config.yaml": string(confYAML),
	}

	return createUnstructuredConfigMap(clients, cm)
}

func TestOCPMonitoringConfigMapUpgrade_ConfigMapDoesNotExist(t *testing.T) {
	clients := createClients()
	upgrade := NewOcpMonitoringConfigmapUpgrade()
	err := upgrade.Upgrade(context.Background(), clients)
	if err != nil {
		t.Fatalf("failed to upgrade configmap: %v", err)
	}

	_, err = getConfigMap(clients)
	assert.Error(t, err)
	assert.True(t, k8serr.IsNotFound(err))
}

func TestOCPMonitoringConfigMapUpgrade_ConfigMapExistsWithGoodExternalLabels(t *testing.T) {
	clients := createClients()
	err := createConfigMapWithGoodExternalLabels(clients)
	if err != nil {
		t.Fatalf("failed to create configmap: %v", err)
	}

	originalCM, err := getConfigMap(clients)
	assert.NoError(t, err)

	upgrade := NewOcpMonitoringConfigmapUpgrade()
	err = upgrade.Upgrade(context.Background(), clients)
	if err != nil {
		t.Fatalf("failed to upgrade configmap: %v", err)
	}

	newCM, err := getConfigMap(clients)
	assert.NoError(t, err)

	assert.Equal(t, originalCM.Data["config.yaml"], newCM.Data["config.yaml"])
}

func TestOCPMonitoringConfigMapUpgrade_ConfigMapExistsWithBadExternalLabels(t *testing.T) {
	clients := createClients()
	err := createConfigMapWithBadExternalLabels(clients)
	if err != nil {
		t.Fatalf("failed to create configmap: %v", err)
	}

	originalCM, err := getConfigMap(clients)
	assert.NoError(t, err)

	upgrade := NewOcpMonitoringConfigmapUpgrade()
	err = upgrade.Upgrade(context.Background(), clients)
	if err != nil {
		t.Fatalf("failed to upgrade configmap: %v", err)
	}

	newCM, err := getConfigMap(clients)
	assert.NoError(t, err)

	assert.NotEqual(t, originalCM.Data["config.yaml"], newCM.Data["config.yaml"])

	clusterMonitoringConfig := &manifests.ClusterMonitoringConfiguration{}
	err = yaml.Unmarshal([]byte(newCM.Data["config.yaml"]), clusterMonitoringConfig)
	if err != nil {
		t.Fatalf("failed to unmarshal configmap: %v", err)
	}

	assert.False(t, haveCommonKeys(clusterMonitoringConfig.PrometheusK8sConfig.ExternalLabels, badExternalLabels))
}

func TestOCPMonitoringConfigMapUpgrade_ConfigMapExistsWithNoExternalLabels(t *testing.T) {
	clients := createClients()
	err := createConfigMapWithNoExternalLabels(clients)
	if err != nil {
		t.Fatalf("failed to create configmap: %v", err)
	}

	originalCM, err := getConfigMap(clients)
	assert.NoError(t, err)

	upgrade := NewOcpMonitoringConfigmapUpgrade()
	err = upgrade.Upgrade(context.Background(), clients)
	if err != nil {
		t.Fatalf("failed to upgrade configmap: %v", err)
	}

	newCM, err := getConfigMap(clients)
	assert.NoError(t, err)

	assert.Equal(t, originalCM.Data["config.yaml"], newCM.Data["config.yaml"])
}

func TestOCPMonitoringConfigMapUpgrade_ConfigMapExistsWithNoK8sPrometheusConfig(t *testing.T) {
	clients := createClients()
	err := createConfigMapWithNoK8sPrometheusConfig(clients)
	if err != nil {
		t.Fatalf("failed to create configmap: %v", err)
	}

	originalCM, err := getConfigMap(clients)
	assert.NoError(t, err)

	upgrade := NewOcpMonitoringConfigmapUpgrade()
	err = upgrade.Upgrade(context.Background(), clients)
	if err != nil {
		t.Fatalf("failed to upgrade configmap: %v", err)
	}

	newCM, err := getConfigMap(clients)
	assert.NoError(t, err)

	assert.Equal(t, originalCM.Data["config.yaml"], newCM.Data["config.yaml"])
}

func haveCommonKeys(m1, m2 map[string]string) bool {
	for k := range m1 {
		if _, ok := m2[k]; ok {
			return true
		}
	}
	return false
}
