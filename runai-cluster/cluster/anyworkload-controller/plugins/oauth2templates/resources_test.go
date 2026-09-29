/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package oauth2templates

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
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
		if err := p.HandleAdditionalResources(ctx, runner, nil, nil, c, nil); err != nil {
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

func TestTemplatesRunBeforeNetworking(t *testing.T) {
	if PluginName >= "networking-plugin" {
		t.Fatal("templates must be provisioned before the networking plugin creates proxy deployments")
	}
}
