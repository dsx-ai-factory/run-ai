/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package plugins

import (
	"context"
	"fmt"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	resourceCreateConfigMapPluginName = "resource-create-configmap-plugin"
	ConfigMapNameSuffix               = "-configmap"
)

// ResourceCreateConfigMapPlugin is a test plugin that creates a ConfigMap.
type ResourceCreateConfigMapPlugin struct{}

func NewResourceCreateConfigMapPlugin() *ResourceCreateConfigMapPlugin {
	return &ResourceCreateConfigMapPlugin{}
}

func (p *ResourceCreateConfigMapPlugin) Name() string {
	return resourceCreateConfigMapPluginName
}

// HandleAdditionalResources creates a ConfigMap with owner reference to the workload itself.
func (p *ResourceCreateConfigMapPlugin) HandleAdditionalResources(
	ctx context.Context,
	runner *dromosv1alpha1.WorkloadRunner,
	resource *unstructured.Unstructured,
	karta *kartav1alpha1.Karta,
	k8sClient client.Client,
	state *reconcile.State,
) error {
	namespace := runner.GetNamespace()
	configMapName := getConfigMapName(runner.Name)
	configMapData := getConfigMapData(runner, resource)

	configMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: namespace,
		},
		Data: configMapData,
	}

	if err := ctrl.SetControllerReference(runner, configMap, k8sClient.Scheme()); err != nil {
		return fmt.Errorf("failed to set controller reference: %w", err)
	}

	return k8sClient.Patch(ctx, configMap, client.Apply, client.ForceOwnership, client.FieldOwner(resourceCreateConfigMapPluginName))
}

func (p *ResourceCreateConfigMapPlugin) Watchers(bldr *builder.Builder) (*builder.Builder, error) {
	return bldr.Owns(&corev1.ConfigMap{}), nil
}

func getConfigMapData(runner *dromosv1alpha1.WorkloadRunner, resource *unstructured.Unstructured) map[string]string {
	return map[string]string{
		"runner-name":            runner.Name,
		"attached-resource-name": resource.GetName(),
		"attached-resource-kind": resource.GetKind(),
	}
}

func getConfigMapName(runnerName string) string {
	return runnerName + ConfigMapNameSuffix
}

var _ framework.AdditionalResourcePlugin = &ResourceCreateConfigMapPlugin{}
