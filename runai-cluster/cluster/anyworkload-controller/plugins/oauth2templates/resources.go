/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package oauth2templates

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants/oauth2proxy"
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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (p *Plugin) HandleAdditionalResources(
	ctx context.Context, runner *dromosv1alpha1.WorkloadRunner,
	_ *unstructured.Unstructured, _ *kartav1alpha1.Karta,
	k8sClient client.Client, _ *reconcile.State,
) error {
	templates := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: p.configMapName, Namespace: runner.Namespace}}
	_, err := controllerutil.CreateOrUpdate(ctx, k8sClient, templates, func() error {
		// Keep the shared file until namespace deletion, independently of any
		// individual workload, and never take over another controller's object.
		if owner := metav1.GetControllerOf(templates); owner != nil {
			return fmt.Errorf("OAuth2 proxy templates are owned by %s %s", owner.Kind, owner.Name)
		}
		templates.Data = map[string]string{oauth2proxy.TemplateKey: oauth2proxy.ErrorPageHTML}
		templates.BinaryData = nil
		return nil
	})
	if err != nil {
		return fmt.Errorf("reconcile OAuth2 proxy templates: %w", err)
	}
	return nil
}

func (p *Plugin) Watchers(bldr *builder.Builder) (*builder.Builder, error) {
	return bldr.Watches(&corev1.ConfigMap{}, handler.EnqueueRequestsFromMapFunc(p.mapConfigMapToRunners)), nil
}

// Reconcile after edits or deletion so restarted pods can still mount the page.
func (p *Plugin) mapConfigMapToRunners(ctx context.Context, obj client.Object) []ctrl.Request {
	if obj.GetName() != p.configMapName {
		return nil
	}
	runners := &dromosv1alpha1.WorkloadRunnerList{}
	if err := p.reader.List(ctx, runners, client.InNamespace(obj.GetNamespace())); err != nil {
		log.FromContext(ctx).Error(err, "list workloads using OAuth2 proxy templates")
		return nil
	}
	requests := make([]ctrl.Request, 0, len(runners.Items))
	for _, runner := range runners.Items {
		requests = append(requests, ctrl.Request{NamespacedName: client.ObjectKeyFromObject(&runner)})
	}
	return requests
}

var _ framework.AdditionalResourcePlugin = &Plugin{}
var _ framework.WatcherPlugin = &Plugin{}
