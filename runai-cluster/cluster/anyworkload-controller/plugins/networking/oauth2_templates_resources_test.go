/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/endpoints"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"
	rcpconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type templateWriteClient struct {
	client.Client
	failWrites          bool
	cachedConfigMapGets int
}

func (c *templateWriteClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	if _, ok := obj.(*corev1.ConfigMap); ok {
		c.cachedConfigMapGets++
		return errors.New("ConfigMap contents must not be read through the cache")
	}
	return c.Client.Get(ctx, key, obj, opts...)
}

func (c *templateWriteClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	if _, ok := obj.(*corev1.ConfigMap); ok && c.failWrites {
		return errors.New("transient ConfigMap create failure")
	}
	return c.Client.Create(ctx, obj, opts...)
}

func (c *templateWriteClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	if _, ok := obj.(*corev1.ConfigMap); ok && c.failWrites {
		return errors.New("transient ConfigMap update failure")
	}
	return c.Client.Update(ctx, obj, opts...)
}

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
	writes := &templateWriteClient{Client: c}
	p := &Plugin{reader: c, templatesReader: c, templatesEnabled: true}
	key := client.ObjectKey{Namespace: namespace.Name, Name: oauth2proxy.ConfigMapName}
	ensure := func() *corev1.ConfigMap {
		t.Helper()
		state := reconcile.New()
		if err := p.HandleAdditionalResources(ctx, runner, nil, nil, writes, state); err != nil {
			t.Fatal(err)
		}
		if len(state.PluginIssues(p)) != 0 || state.RequeueAfter() != 0 {
			t.Fatalf("successful provisioning requested a retry: %+v", state.PluginIssues(p))
		}
		actual := &corev1.ConfigMap{}
		if err := c.Get(ctx, key, actual); err != nil {
			t.Fatal(err)
		}
		if actual.Data[oauth2proxy.TemplateKey] != oauth2ErrorPageHTML {
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
	if writes.cachedConfigMapGets != 0 {
		t.Fatal("template provisioning read ConfigMap contents through the cache")
	}
}

func TestTemplateConfigMapDeletionQueuesOnlyItsNamespace(t *testing.T) {
	first := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{Name: "first", Namespace: "project-a"}}
	second := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{Name: "second", Namespace: "project-a"}}
	other := &dromosv1alpha1.WorkloadRunner{ObjectMeta: metav1.ObjectMeta{Name: "other", Namespace: "project-b"}}
	c := templateResourceClient(t, first, second, other)
	p := &Plugin{reader: c, templatesEnabled: true}
	// Deleted ConfigMaps are absent from the client; the event still carries their namespace.
	deleted := &metav1.PartialObjectMetadata{ObjectMeta: metav1.ObjectMeta{Name: oauth2proxy.ConfigMapName, Namespace: "project-a"}}
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
	t.Setenv(oauth2proxy.TemplatesEnabledEnvVar, "true")
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
	if !p.routingDisabled || !p.templatesEnabled {
		t.Fatal("disabled networking did not retain template provisioning")
	}
	if err := integration.Setup(context.Background(), nil, nil); err == nil {
		t.Fatal("template provisioning must require an uncached reader even when routing is disabled")
	}
	if err := integration.Setup(context.Background(), c, nil); err != nil {
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

	t.Setenv(oauth2proxy.TemplatesEnabledEnvVar, "false")
	if _, err := NewIntegration().Build(context.Background(), framework.PluginDeps{}); !errors.Is(err, framework.ErrPluginDisabled) {
		t.Fatalf("networking without templates should remain disabled: %v", err)
	}
}

func TestTemplateFailuresRetryWithoutBlockingRouting(t *testing.T) {
	for _, routingDisabled := range []bool{false, true} {
		for _, failure := range []string{"ownership collision", "create", "update"} {
			t.Run(fmt.Sprintf("%s/routing disabled=%t", failure, routingDisabled), func(t *testing.T) {
				ctx := context.Background()
				runner := enforcementRunner()
				pod := enforcementPod(runner, "owned-pod", runner.UID)
				apiClient := enforcementClient(t, pod)
				key := client.ObjectKey{Namespace: runner.Namespace, Name: oauth2proxy.ConfigMapName}
				var existing *corev1.ConfigMap
				if failure != "create" {
					existing = &corev1.ConfigMap{
						ObjectMeta: metav1.ObjectMeta{Name: key.Name, Namespace: key.Namespace},
						Data:       map[string]string{oauth2proxy.TemplateKey: "existing page"},
					}
					if failure == "ownership collision" {
						controller := true
						existing.OwnerReferences = []metav1.OwnerReference{{
							APIVersion: "apps/v1", Kind: "Deployment", Name: "another-controller", UID: "other-uid", Controller: &controller,
						}}
					}
					if err := apiClient.Create(ctx, existing); err != nil {
						t.Fatal(err)
					}
				}
				writes := &templateWriteClient{Client: apiClient, failWrites: failure != "ownership collision"}
				p := newOwnershipTestPlugin(apiClient)
				p.templatesEnabled = true
				p.templatesReader = apiClient
				p.routingDisabled = routingDisabled
				state := reconcile.New()
				if err := p.HandleAdditionalResources(ctx, runner, nil, nil, writes, state); err != nil {
					t.Fatalf("template failure blocked workload reconciliation: %v", err)
				}
				issues := state.PluginIssues(p)
				if len(issues) != 1 || issues[0].IsCritical || issues[0].Description != "OAuth2 proxy templates" {
					t.Fatalf("expected one noncritical template issue: %+v", issues)
				}
				if state.RequeueAfter() != 30*time.Second {
					t.Fatalf("template failure requested retry after %s", state.RequeueAfter())
				}
				if !routingDisabled {
					for _, object := range []client.Object{&corev1.Service{}, &networkingv1.Ingress{}} {
						if err := apiClient.Get(ctx, client.ObjectKey{Namespace: runner.Namespace, Name: enforcementServiceName(object)}, object); err != nil {
							t.Fatalf("template failure blocked creation of %T: %v", object, err)
						}
					}
				} else {
					services := &corev1.ServiceList{}
					if err := apiClient.List(ctx, services); err != nil {
						t.Fatal(err)
					}
					if len(services.Items) != 0 {
						t.Fatal("disabled routing created services")
					}
				}
				if existing != nil {
					unchanged := &corev1.ConfigMap{}
					if err := apiClient.Get(ctx, key, unchanged); err != nil {
						t.Fatal(err)
					}
					if unchanged.ResourceVersion != existing.ResourceVersion {
						t.Fatal("failed provisioning changed the existing ConfigMap")
					}
					if failure == "ownership collision" {
						if err := apiClient.Delete(ctx, unchanged); err != nil {
							t.Fatal(err)
						}
					}
				}
				writes.failWrites = false
				retryState := reconcile.New()
				if err := p.HandleAdditionalResources(ctx, runner, nil, nil, writes, retryState); err != nil {
					t.Fatalf("template recovery failed: %v", err)
				}
				if len(retryState.PluginIssues(p)) != 0 || retryState.RequeueAfter() != 0 {
					t.Fatalf("template issue persisted after recovery: %+v", retryState.PluginIssues(p))
				}
				recovered := &corev1.ConfigMap{}
				if err := apiClient.Get(ctx, key, recovered); err != nil {
					t.Fatal(err)
				}
				if recovered.Data[oauth2proxy.TemplateKey] != oauth2ErrorPageHTML {
					t.Fatal("retry did not restore the error page")
				}
				if writes.cachedConfigMapGets != 0 {
					t.Fatal("template provisioning read ConfigMap contents through the cache")
				}
			})
		}
	}
}
