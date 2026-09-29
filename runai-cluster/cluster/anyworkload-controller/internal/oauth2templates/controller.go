/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package oauth2templates

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Reconciler provides the Run:ai OAuth2 proxy error page in each project namespace.
type Reconciler struct {
	client.Client
}

func NewReconciler(c client.Client) *Reconciler {
	return &Reconciler{Client: c}
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("oauth2-proxy-templates").
		For(&kaiorg.Project{}).
		Owns(&corev1.ConfigMap{}, builder.WithPredicates(predicate.NewPredicateFuncs(func(obj client.Object) bool {
			return obj.GetName() == oauth2proxy.ConfigMapName
		}))).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	project := &kaiorg.Project{}
	if err := r.Get(ctx, req.NamespacedName, project); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	return ctrl.Result{}, r.reconcileProject(ctx, project)
}

func (r *Reconciler) reconcileProject(ctx context.Context, project *kaiorg.Project) error {
	namespace := project.Status.Namespace
	if namespace == "" || !project.DeletionTimestamp.IsZero() {
		return nil
	}
	ns := &corev1.Namespace{}
	if err := r.Get(ctx, client.ObjectKey{Name: namespace}, ns); err != nil {
		return fmt.Errorf("get namespace for OAuth2 proxy templates: %w", err)
	}
	if !ns.DeletionTimestamp.IsZero() {
		return nil
	}

	templates := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      oauth2proxy.ConfigMapName,
			Namespace: namespace,
		},
	}
	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, templates, func() error {
		if err := controllerutil.SetControllerReference(project, templates, r.Scheme(), controllerutil.WithBlockOwnerDeletion(false)); err != nil {
			return err
		}
		templates.Data = map[string]string{oauth2proxy.TemplateKey: oauth2proxy.ErrorPageHTML}
		templates.BinaryData = nil
		return nil
	})
	if err != nil {
		return fmt.Errorf("reconcile OAuth2 proxy templates for project %s: %w", project.Name, err)
	}
	return nil
}
