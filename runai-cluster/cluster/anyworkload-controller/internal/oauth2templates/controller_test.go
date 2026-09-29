/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package oauth2templates

import (
	"context"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func newOAuth2TemplatesTestClient(t *testing.T, objects ...client.Object) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	if err := kaiorg.AddToScheme(scheme); err != nil {
		t.Fatal(err)
	}
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
}

func oauth2TemplatesTestProject() *kaiorg.Project {
	return &kaiorg.Project{
		ObjectMeta: metav1.ObjectMeta{Name: "project-a", UID: types.UID("project-a-uid")},
		Status:     kaiorg.ProjectStatus{Namespace: "runai-project-a"},
	}
}

func TestOAuth2ProxyTemplatesReconcile(t *testing.T) {
	ctx := context.Background()
	project := oauth2TemplatesTestProject()
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: project.Status.Namespace}}
	k8sClient := newOAuth2TemplatesTestClient(t, project, namespace)
	reconciler := NewReconciler(k8sClient)
	req := ctrl.Request{NamespacedName: client.ObjectKeyFromObject(project)}
	key := client.ObjectKey{Namespace: namespace.Name, Name: oauth2proxy.ConfigMapName}

	if _, err := reconciler.Reconcile(ctx, req); err != nil {
		t.Fatalf("create templates: %v", err)
	}
	actual := &corev1.ConfigMap{}
	if err := k8sClient.Get(ctx, key, actual); err != nil {
		t.Fatal(err)
	}
	if actual.Data[oauth2proxy.TemplateKey] != oauth2proxy.ErrorPageHTML {
		t.Fatal("ConfigMap does not contain the preserved error page")
	}
	if !metav1.IsControlledBy(actual, project) {
		t.Fatal("ConfigMap is not owned by its project")
	}
	if owner := metav1.GetControllerOf(actual); owner.BlockOwnerDeletion == nil || *owner.BlockOwnerDeletion {
		t.Fatal("ConfigMap must not require permission to block project deletion")
	}
	if _, exists := actual.Labels[constants.ClusterWideResourceLabel]; exists {
		t.Fatal("system template must not be managed by optional cluster-wide replication")
	}

	resourceVersion := actual.ResourceVersion
	if _, err := reconciler.Reconcile(ctx, req); err != nil {
		t.Fatalf("repeat reconciliation: %v", err)
	}
	if err := k8sClient.Get(ctx, key, actual); err != nil {
		t.Fatal(err)
	}
	if actual.ResourceVersion != resourceVersion {
		t.Fatal("unchanged templates caused an unnecessary update")
	}

	actual.Data[oauth2proxy.TemplateKey] = "outdated error page"
	if err := k8sClient.Update(ctx, actual); err != nil {
		t.Fatal(err)
	}
	if _, err := reconciler.Reconcile(ctx, req); err != nil {
		t.Fatalf("repair templates: %v", err)
	}
	if err := k8sClient.Get(ctx, key, actual); err != nil {
		t.Fatal(err)
	}
	if actual.Data[oauth2proxy.TemplateKey] != oauth2proxy.ErrorPageHTML {
		t.Fatal("modified error page was not restored")
	}

	if err := k8sClient.Delete(ctx, actual); err != nil {
		t.Fatal(err)
	}
	if _, err := reconciler.Reconcile(ctx, req); err != nil {
		t.Fatalf("recreate templates: %v", err)
	}
	if err := k8sClient.Get(ctx, key, actual); err != nil {
		t.Fatalf("deleted ConfigMap was not recreated: %v", err)
	}
}

func TestOAuth2ProxyTemplatesSkipUnavailableProjectNamespace(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		noNamespace bool
		terminating bool
		deleting    bool
	}{
		{name: "namespace not assigned", noNamespace: true},
		{name: "namespace terminating", terminating: true},
		{name: "project deleting", deleting: true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			project := oauth2TemplatesTestProject()
			namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: project.Status.Namespace}}
			if testCase.noNamespace {
				project.Status.Namespace = ""
			}
			if testCase.terminating {
				now := metav1.Now()
				namespace.DeletionTimestamp = &now
				namespace.Finalizers = []string{"test/cleanup"}
			}
			if testCase.deleting {
				now := metav1.Now()
				project.DeletionTimestamp = &now
			}
			k8sClient := newOAuth2TemplatesTestClient(t, namespace)
			if err := NewReconciler(k8sClient).reconcileProject(context.Background(), project); err != nil {
				t.Fatal(err)
			}
			configMaps := &corev1.ConfigMapList{}
			if err := k8sClient.List(context.Background(), configMaps); err != nil {
				t.Fatal(err)
			}
			if len(configMaps.Items) != 0 {
				t.Fatal("created templates before namespace was ready or while it was terminating")
			}
		})
	}
}

func TestOAuth2ProxyTemplatesDeletedProjectDoesNotRequeue(t *testing.T) {
	k8sClient := newOAuth2TemplatesTestClient(t)
	result, err := NewReconciler(k8sClient).Reconcile(context.Background(), ctrl.Request{
		NamespacedName: client.ObjectKey{Name: "deleted-project"},
	})
	if err != nil || result != (ctrl.Result{}) {
		t.Fatalf("deleted project should not requeue: result=%+v, err=%v", result, err)
	}
}

func TestOAuth2ProxyTemplatesMissingNamespaceReturnsError(t *testing.T) {
	project := oauth2TemplatesTestProject()
	k8sClient := newOAuth2TemplatesTestClient(t)
	err := NewReconciler(k8sClient).reconcileProject(context.Background(), project)
	if !apierrors.IsNotFound(err) {
		t.Fatalf("expected missing namespace error to retry reconciliation, got %v", err)
	}
}

func TestOAuth2ProxyTemplatesDoesNotTakeAnotherControllersConfigMap(t *testing.T) {
	project := oauth2TemplatesTestProject()
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: project.Status.Namespace}}
	otherOwner := &kaiorg.Project{ObjectMeta: metav1.ObjectMeta{Name: "other-project", UID: types.UID("other-uid")}}
	templates := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: oauth2proxy.ConfigMapName, Namespace: namespace.Name},
		Data:       map[string]string{oauth2proxy.TemplateKey: "owned by someone else"},
	}
	k8sClient := newOAuth2TemplatesTestClient(t, namespace)
	if err := controllerutil.SetControllerReference(otherOwner, templates, k8sClient.Scheme()); err != nil {
		t.Fatal(err)
	}
	if err := k8sClient.Create(context.Background(), templates); err != nil {
		t.Fatal(err)
	}
	if err := NewReconciler(k8sClient).reconcileProject(context.Background(), project); err == nil {
		t.Fatal("expected owner conflict")
	}
	actual := &corev1.ConfigMap{}
	if err := k8sClient.Get(context.Background(), client.ObjectKeyFromObject(templates), actual); err != nil {
		t.Fatal(err)
	}
	if actual.Data[oauth2proxy.TemplateKey] != templates.Data[oauth2proxy.TemplateKey] {
		t.Fatal("changed another controller's ConfigMap")
	}
}
