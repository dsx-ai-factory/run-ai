/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"time"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/rs/zerolog/log"
	clustercommonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/shared-objects-controller/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/shared-objects-controller/pkg/controller/project"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const requeueDelay = 10 * time.Second

type ProjectController struct {
	client                   client.Client
	createServiceMonitors    bool
	serviceMonitorReconciler *project.ServiceMonitorReconciler
	clusterWideReconciler    *project.ClusterWideReplicatedReconciler
}

func NewProjectController(ctx context.Context, c client.Client) (*ProjectController, error) {
	projectHandlerConfig := config.Config().ProjectHandler
	pc := &ProjectController{
		client:                   c,
		createServiceMonitors:    projectHandlerConfig.CreateServiceMonitors,
		serviceMonitorReconciler: project.NewServiceMonitorReconciler(c),
	}

	if projectHandlerConfig.ClusterWideSecrets || projectHandlerConfig.ClusterWidePvcs || projectHandlerConfig.ClusterWideConfigMaps {
		cwr, err := project.NewClusterWideReplicatedReconciler(
			ctx, c, projectHandlerConfig.ClusterWideSecrets, projectHandlerConfig.ClusterWidePvcs, projectHandlerConfig.ClusterWideConfigMaps)
		if err != nil {
			return nil, err
		}
		pc.clusterWideReconciler = cwr
	}

	return pc, nil
}

func (c *ProjectController) clusterWideEnabled() bool {
	return c.clusterWideReconciler != nil
}

func (c *ProjectController) SetupWithManager(mgr ctrl.Manager) error {
	b := ctrl.NewControllerManagedBy(mgr).
		For(&kaiorg.Project{}).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
			config.Config().BaseControllerDelay, config.Config().MaxControllerDelay)})

	if c.clusterWideEnabled() {
		projectHandlerConfig := config.Config().ProjectHandler
		schedulerNamespacePredicate := predicate.NewPredicateFuncs(func(obj client.Object) bool {
			return obj.GetNamespace() == clustercommonconstants.NamespaceRunAi
		})
		if projectHandlerConfig.ClusterWideSecrets {
			b = b.Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(c.mapToAllProjects),
				builder.WithPredicates(schedulerNamespacePredicate))
		}
		if projectHandlerConfig.ClusterWidePvcs {
			b = b.Watches(&corev1.PersistentVolumeClaim{}, handler.EnqueueRequestsFromMapFunc(c.mapToAllProjects),
				builder.WithPredicates(schedulerNamespacePredicate))
		}
		if projectHandlerConfig.ClusterWideConfigMaps {
			b = b.Watches(&corev1.ConfigMap{}, handler.EnqueueRequestsFromMapFunc(c.mapToAllProjects),
				builder.WithPredicates(schedulerNamespacePredicate))
		}
	}

	return b.Complete(c)
}

// mapToAllProjects enqueues every kai Project for reconciliation when a cluster-wide source object
// (in the scheduler namespace) changes, so replicated copies are kept in sync.
func (c *ProjectController) mapToAllProjects(ctx context.Context, _ client.Object) []ctrl.Request {
	projects := &kaiorg.ProjectList{}
	if err := c.client.List(ctx, projects); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Failed listing kai projects to enqueue for cluster-wide source change")
		return nil
	}
	requests := make([]ctrl.Request, 0, len(projects.Items))
	for i := range projects.Items {
		requests = append(requests, ctrl.Request{NamespacedName: client.ObjectKey{Name: projects.Items[i].Name}})
	}
	return requests
}

func (c *ProjectController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = log.Logger.WithContext(ctx)
	log.Ctx(ctx).Info().Msgf("Reconciling kai project %s", req.Name)

	projectObj := &kaiorg.Project{}
	err := c.client.Get(ctx, req.NamespacedName, projectObj)
	if err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: requeueDelay}, err
	}

	if !projectObj.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if c.clusterWideEnabled() {
		if err = c.clusterWideReconciler.Reconcile(ctx, projectObj); err != nil {
			return ctrl.Result{}, err
		}
	}

	if c.createServiceMonitors {
		if err = c.serviceMonitorReconciler.Reconcile(ctx, projectObj); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
