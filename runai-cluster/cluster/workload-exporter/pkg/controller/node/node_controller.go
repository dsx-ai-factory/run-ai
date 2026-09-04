/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package node

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/controller/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/metrics"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/metrics/node_mig_exporter"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type NodeController struct {
	client client.Client
}

func NewNodeController(client client.Client) *NodeController {
	return &NodeController{
		client: client,
	}
}

func (c *NodeController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		WithEventFilter(MigNodesEventFilter()).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
			utils.RateLimiterBaseDelay, utils.RateLimiterMaxDelay)}).
		Complete(c)
}

func (c *NodeController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = log.Logger.WithContext(ctx)
	log.Debug().Msgf("Reconciling node %s", req.Name)

	node := &corev1.Node{}
	isObjectFound, err := utils.GetObjectForRequest(ctx, req, "node", node, c.client)
	if err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: utils.RequeueTimeout}, nil
	} else if !isObjectFound {
		metrics.ResetNodeMetrics(req.Name)
		return ctrl.Result{Requeue: false}, nil
	}

	err = metrics.SetNodeMetrics(node)
	if err != nil {
		log.Error().Msgf("Failed to set metrics for node %s, error: %s", req.Name, err.Error())
		return ctrl.Result{Requeue: true, RequeueAfter: utils.RequeueTimeout}, nil
	}
	return ctrl.Result{}, nil
}

func MigNodesEventFilter() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isMigNodeReady(e.Object.(*corev1.Node))
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isMigNodeReady(e.ObjectNew.(*corev1.Node))
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isMigNodeReady(e.Object.(*corev1.Node))
		},
	}
}

func isMigNodeReady(node *corev1.Node) bool {
	if !metrics.IsNodeReady(node) {
		return false
	}
	nodeMigExporter := node_mig_exporter.NewMigStrategyExporter(node)
	return nodeMigExporter.IsMigNode()
}
