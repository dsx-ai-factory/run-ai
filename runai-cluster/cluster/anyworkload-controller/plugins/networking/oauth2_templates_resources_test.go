/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"errors"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	rcpconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func templateResourceClient(t *testing.T, objects ...client.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	for _, register := range []func(*runtime.Scheme) error{corev1.AddToScheme, dromosv1alpha1.AddToScheme} {
		if err := register(scheme); err != nil {
			t.Fatal(err)
		}
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

func TestTemplateConfigMapLifecycle(t *testing.T) {
	ctx := context.Background()
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "project-a", UID: "namespace-uid"}}
	runner := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{Name: "workload-a", Namespace: namespace.Name, UID: "runner-uid"}}
	c := templateResourceClient(t, namespace, runner)
	p := &Plugin{reader: c, configMapName: oauth2proxy.ConfigMapName}
	key := client.ObjectKey{Namespace: namespace.Name, Name: oauth2proxy.ConfigMapName}
	ensure := func() *corev1.ConfigMap {
		t.Helper()
		if err := p.HandleAdditionalResources(ctx, runner, nil, nil, c, reconcile.New()); err != nil {
			t.Fatal(err)
		}
		actual := &corev1.ConfigMap{}
		if err := c.Get(ctx, key, actual); err != nil {
			t.Fatal(err)
		}
		if actual.Data[oauth2proxy.TemplateKey] != oauth2proxy.ErrorPageHTML {
			t.Fatal("ConfigMap does not contain the preserved error page")
		}
		if len(actual.OwnerReferences) != 0 {
			t.Fatalf("shared templates must survive workload deletion: %+v", actual.OwnerReferences)
		}
		return actual
	}
	actual := ensure()
	if again := ensure(); again.ResourceVersion != actual.ResourceVersion {
		t.Fatal("unchanged templates caused an unnecessary update")
	}
	actual.Data[oauth2proxy.TemplateKey] = "outdated error page"
	if err := c.Update(ctx, actual); err != nil {
		t.Fatal(err)
	}
	actual = ensure()
	if err := c.Delete(ctx, actual); err != nil {
		t.Fatal(err)
	}
	ensure()
}

func TestTemplateConfigMapDeletionQueuesOnlyItsNamespace(t *testing.T) {
	first := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{Name: "first", Namespace: "project-a"}}
	second := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{Name: "second", Namespace: "project-a"}}
	other := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{Name: "other", Namespace: "project-b"}}
	c := templateResourceClient(t, first, second, other)
	p := &Plugin{reader: c, configMapName: oauth2proxy.ConfigMapName}
	// Deleted ConfigMaps are absent from the client; the event still carries their namespace.
	deleted := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: oauth2proxy.ConfigMapName, Namespace: "project-a"}}
	requests := p.mapConfigMapToRunners(context.Background(), deleted)
	expected := map[client.ObjectKey]bool{client.ObjectKeyFromObject(first): true, client.ObjectKeyFromObject(second): true}
	if len(requests) != len(expected) {
		t.Fatalf("expected both workloads in project-a, got %+v", requests)
	}
	for _, request := range requests {
		if !expected[request.NamespacedName] {
			t.Fatalf("unexpected or duplicate request: %+v", request)
		}
		delete(expected, request.NamespacedName)
	}
	deleted.Name = "unrelated-configmap"
	if requests := p.mapConfigMapToRunners(context.Background(), deleted); len(requests) != 0 {
		t.Fatalf("unrelated ConfigMap queued workloads: %+v", requests)
	}
}

func TestDisabledNetworkingStillProvidesTemplates(t *testing.T) {
	previous, err := rcpconfig.GetRootR[*Config]()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := rcpconfig.Set(previous); err != nil {
			t.Errorf("restore networking config: %v", err)
		}
	})
	if err := rcpconfig.Set(&Config{Enabled: false}); err != nil {
		t.Fatal(err)
	}
	t.Setenv(oauth2proxy.TemplatesConfigMapEnvVar, oauth2proxy.ConfigMapName)
	runner := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{
		Name: "inline-proxy", Namespace: "project-a", Annotations: map[string]string{endpoints.Annotation: "invalid networking input"},
	}}
	c := templateResourceClient(t, runner)
	integration := NewIntegration()
	plugin, err := integration.Build(context.Background(), framework.PluginDeps{Client: c})
	if err != nil {
		t.Fatal(err)
	}
	p := plugin.(*Plugin)
	if !p.routingDisabled || p.configMapName != oauth2proxy.ConfigMapName {
		t.Fatal("disabled networking did not retain template provisioning")
	}
	if err := integration.Setup(context.Background(), nil, nil); err != nil {
		t.Fatalf("template provisioning must not require routing setup: %v", err)
	}
	if err := p.HandleAdditionalResources(context.Background(), runner, nil, nil, c, reconcile.New()); err != nil {
		t.Fatal(err)
	}
	if err := c.Get(context.Background(), client.ObjectKey{Namespace: runner.Namespace, Name: oauth2proxy.ConfigMapName}, &corev1.ConfigMap{}); err != nil {
		t.Fatalf("inline-only workload did not receive templates: %v", err)
	}
	if warnings, errs := p.WebhookValidate(context.Background(), nil, runner); len(warnings) != 0 || len(errs) != 0 {
		t.Fatalf("disabled routing validated networking input: %v, %v", warnings, errs)
	}
	status := newPluginStatus()
	if err := p.Status(context.Background(), runner, nil, status, nil); err != nil || status.Phase != dromosv1alpha1.PluginPhaseDisabled {
		t.Fatalf("disabled routing status: phase=%s, err=%v", status.Phase, err)
	}

	t.Setenv(oauth2proxy.TemplatesConfigMapEnvVar, "")
	if _, err := NewIntegration().Build(context.Background(), framework.PluginDeps{}); !errors.Is(err, framework.ErrPluginDisabled) {
		t.Fatalf("networking without templates should remain disabled: %v", err)
	}
}
