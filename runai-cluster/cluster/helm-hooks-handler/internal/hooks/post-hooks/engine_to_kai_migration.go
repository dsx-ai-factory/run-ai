package post_hooks

import (
	"context"
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	legacyEngineConfigName             = "engine-config"
	legacyAccountingServiceMonitorName = "nodepool-controller-accounting"
)

var legacyEngineConfigGVK = schema.GroupVersionKind{
	Group: "engine.run.ai", Version: "v1alpha1", Kind: "Config",
}

func (p *PostUpgradeHook) cleanupLegacyEngineResources(ctx context.Context, k8sClient client.Client) error {
	if !p.params.EngineToKAIMigrationEnabled {
		return nil
	}

	resources := []client.Object{
		legacyEngineConfig(),
		&monitoringv1.ServiceMonitor{ObjectMeta: metav1.ObjectMeta{
			Name:      legacyAccountingServiceMonitorName,
			Namespace: runaiv1.DefaultNamespace,
		}},
	}
	for _, resource := range resources {
		if err := deleteLegacyResourceIfPresent(ctx, k8sClient, resource); err != nil {
			return err
		}
	}

	return nil
}

func legacyEngineConfig() *unstructured.Unstructured {
	engineConfig := &unstructured.Unstructured{}
	engineConfig.SetGroupVersionKind(legacyEngineConfigGVK)
	engineConfig.SetName(legacyEngineConfigName)
	return engineConfig
}

func deleteLegacyResourceIfPresent(ctx context.Context, k8sClient client.Client, resource client.Object) error {
	err := k8sClient.Get(ctx, client.ObjectKeyFromObject(resource), resource)
	if apierrors.IsNotFound(err) || meta.IsNoMatchError(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get legacy %T %q: %w", resource, resource.GetName(), err)
	}

	if err = k8sClient.Delete(ctx, resource); err != nil && !apierrors.IsNotFound(err) && !meta.IsNoMatchError(err) {
		return fmt.Errorf("failed to delete legacy %T %q: %w", resource, resource.GetName(), err)
	}
	return nil
}
