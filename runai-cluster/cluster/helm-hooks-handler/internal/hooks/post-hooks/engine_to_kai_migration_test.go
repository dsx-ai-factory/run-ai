/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package post_hooks

import (
	"context"
	"testing"

	schemeinit "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/scheme-init"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/helm-hooks-handler/internal/hooks"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type noLegacyCRDsClient struct {
	client.Client
}

func (c *noLegacyCRDsClient) Get(
	ctx context.Context,
	key client.ObjectKey,
	obj client.Object,
	opts ...client.GetOption,
) error {
	switch {
	case obj.GetObjectKind().GroupVersionKind() == legacyEngineConfigGVK:
		return &meta.NoKindMatchError{GroupKind: schema.GroupKind{Group: "engine.run.ai", Kind: "Config"}}
	case obj.GetObjectKind().GroupVersionKind().GroupKind() == monitoringv1.SchemeGroupVersion.WithKind("ServiceMonitor").GroupKind():
		return &meta.NoKindMatchError{GroupKind: schema.GroupKind{Group: "monitoring.coreos.com", Kind: "ServiceMonitor"}}
	default:
		return c.Client.Get(ctx, key, obj, opts...)
	}
}

func TestLegacyEngineCleanupIsDisabledByDefault(t *testing.T) {
	hook := NewPostUpgradeHook(&hooks.ClusterHookParams{})
	k8sClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).Build()
	require.NoError(t, hook.cleanupLegacyEngineResources(context.Background(), k8sClient))
}

func TestLegacyEngineCleanupDoesNotWaitForKAIAvailability(t *testing.T) {
	hook := NewPostUpgradeHook(&hooks.ClusterHookParams{
		EngineToKAIMigrationEnabled: true,
	})
	k8sClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).WithObjects(
		legacyEngineConfig(),
	).Build()

	require.NoError(t, hook.cleanupLegacyEngineResources(context.Background(), k8sClient))
	require.Error(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: legacyEngineConfigName}, legacyEngineConfig()))
}

func TestLegacyEngineCleanupToleratesMissingCRDs(t *testing.T) {
	hook := NewPostUpgradeHook(&hooks.ClusterHookParams{
		EngineToKAIMigrationEnabled: true,
	})
	baseClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).Build()
	k8sClient := &noLegacyCRDsClient{Client: baseClient}

	require.NoError(t, hook.cleanupLegacyEngineResources(context.Background(), k8sClient))
}

func TestLegacyEngineCleanupDeletesResourcesAndIsIdempotent(t *testing.T) {
	hook := NewPostUpgradeHook(&hooks.ClusterHookParams{
		EngineToKAIMigrationEnabled: true,
	})
	k8sClient := fake.NewClientBuilder().WithScheme(schemeinit.Scheme()).WithObjects(
		legacyEngineConfig(),
		&monitoringv1.ServiceMonitor{ObjectMeta: metav1.ObjectMeta{
			Name:      legacyAccountingServiceMonitorName,
			Namespace: runaiv1.DefaultNamespace,
		}},
	).Build()

	require.NoError(t, hook.cleanupLegacyEngineResources(context.Background(), k8sClient))
	require.NoError(t, hook.cleanupLegacyEngineResources(context.Background(), k8sClient))

	require.Error(t, k8sClient.Get(context.Background(), client.ObjectKey{Name: legacyEngineConfigName}, legacyEngineConfig()))
	require.Error(t, k8sClient.Get(context.Background(), client.ObjectKey{
		Name: legacyAccountingServiceMonitorName, Namespace: runaiv1.DefaultNamespace,
	}, &monitoringv1.ServiceMonitor{}))
}
