/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controller

import (
	"context"
	"sync"

	reconsilestate "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/reconcile"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	dromospredicates "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/predicates"
	kartautils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/utils/karta"
	templateutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/utils/template"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlbuilder "sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	controllerName = "workloadrunner-controller"
)

type WorkloadRunnerController struct {
	client       client.Client
	recorder     record.EventRecorder
	wrController controller.Controller
	manager      ctrl.Manager
	// activeWatches is a map of GVKs that are being watched.
	// should be accessed with watchMutex.
	activeWatches map[schema.GroupVersionKind]bool
	// watchMutex to protect watch gvk once.
	watchMutex      sync.Mutex
	pluginFramework *framework.Framework
	kartaUtils      *kartautils.Factory
	templateParser  *templateutils.Parser
	name            string
}

func NewWorkloadRunnerController(client client.Client, fw *framework.Framework, recorder record.EventRecorder) *WorkloadRunnerController {
	return &WorkloadRunnerController{
		client:          client,
		recorder:        recorder,
		pluginFramework: fw,
		activeWatches:   make(map[schema.GroupVersionKind]bool),
		kartaUtils:      kartautils.NewFactory(client),
		templateParser:  templateutils.NewParser(),
		name:            controllerName,
	}
}

// WithName sets the name of the controller.
// This is enables creating multiple controllers for testing purposes.
func (c *WorkloadRunnerController) WithName(name string) *WorkloadRunnerController {
	c.name = name
	return c
}

func (c *WorkloadRunnerController) SetupWithManager(ctx context.Context, mgr ctrl.Manager) error {
	c.manager = mgr

	builder := ctrl.NewControllerManagedBy(mgr).
		Named(c.name).
		For(&dromosv1alpha1.WorkloadRunner{}, ctrlbuilder.WithPredicates(
			predicate.GenerationChangedPredicate{},
			dromospredicates.ManagedByPredicate(""),
		)).
		Watches(&kartav1alpha1.Karta{},
			handler.EnqueueRequestsFromMapFunc(c.mapKartaToWorkloadRunnerEvent))

	builder, err := c.pluginFramework.RunWatchers(builder)
	if err != nil {
		return err
	}

	newCtrl, err := builder.Build(c)
	if err != nil {
		return err
	}

	c.wrController = newCtrl

	return nil
}

func (c *WorkloadRunnerController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling WorkloadRunner", "namespace", req.Namespace, "name", req.Name)

	var runner dromosv1alpha1.WorkloadRunner
	if err := c.client.Get(ctx, req.NamespacedName, &runner); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if runner.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	// Complements the watch predicate, events queued before managedBy was set still reach here.
	if runner.Spec.ManagedBy != "" {
		logger.V(1).Info("runner managed by another controller, yielding", "managedBy", runner.Spec.ManagedBy)
		return ctrl.Result{}, nil
	}

	state := reconsilestate.New()

	specResult := c.reconcileSpec(ctx, &runner, state)
	statusResult := c.reconcileStatusAndRunPostPhase(ctx, &runner, specResult, state)

	if statusResult.ShortCircuitReconcileFlow() {
		return statusResult.Result()
	}

	return specResult.Result()
}

// mapKartaToWorkloadRunnerEvent maps a Karta to a WorkloadRunner event.
// Given a Karta, it will fetch all WorkloadRunners that have a template with the same root component GVK.
func (c *WorkloadRunnerController) mapKartaToWorkloadRunnerEvent(ctx context.Context, object client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)
	karta, ok := object.(*kartav1alpha1.Karta)

	if !ok {
		logger.Error(nil, "Object is not a Karta", "type", object)
		return []reconcile.Request{}
	}

	gvk, err := kartautils.GetRootComponentGVK(karta)
	if err != nil {
		logger.Error(err, "Failed to get root component GVK for Karta", "name", karta.Name)
		return []reconcile.Request{}
	}

	runnerList := &dromosv1alpha1.WorkloadRunnerList{}
	if err := c.client.List(ctx, runnerList); err != nil {
		logger.Error(err, "Failed to list WorkloadRunners")
		return []reconcile.Request{}
	}

	var requests []reconcile.Request

	for _, runner := range runnerList.Items {
		template, err := c.templateParser.Parse(runner.Spec.Template)
		if err != nil {
			logger.Error(err, "Failed to parse template for WorkloadRunner while mapping Karta to WorkloadRunner event", "namespace", runner.Namespace, "name", runner.Name)
			continue
		}

		templateGVK := template.GroupVersionKind()
		if templateGVK == *gvk {
			requests = append(requests, reconcile.Request{
				NamespacedName: client.ObjectKey{
					Name:      runner.Name,
					Namespace: runner.Namespace,
				},
			})
		}
	}

	logger.Info("Karta changed, triggering reconciliation for WorkloadRunners", "karta", karta.Name, "count", len(requests))

	return requests
}
