/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"context"

	"github.com/rs/zerolog/log"
	iwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-exporter/pkg/controller/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-exporter/pkg/metrics"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type InteractiveWorkloadController struct {
	client client.Client
}

func NewInteractiveWorkloadController(client client.Client) *InteractiveWorkloadController {
	return &InteractiveWorkloadController{
		client: client,
	}
}

func (c *InteractiveWorkloadController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&iwv2a1.InteractiveWorkload{}).
		WithEventFilter(workloadsEventFilter()).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
			utils.RateLimiterBaseDelay, utils.RateLimiterMaxDelay)}).
		Complete(c)
}

func (c *InteractiveWorkloadController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = log.Logger.WithContext(ctx)
	log.Debug().Msgf("Reconciling interactiveworkload %s/%s", req.Name, req.Namespace)

	interactiveWorkload := &iwv2a1.InteractiveWorkload{}
	isObjectFound, err := utils.GetObjectForRequest(ctx, req, "interactiveworkload", interactiveWorkload, c.client)
	if err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: utils.RequeueTimeout}, nil
	} else if !isObjectFound || !interactiveWorkload.DeletionTimestamp.IsZero() {
		metrics.ResetWorkloadMetrics(req.Name, req.Namespace, metrics.InteractiveSource)
		return ctrl.Result{Requeue: false}, nil
	}

	err = metrics.SetInteractiveWorkloadMetrics(ctx, interactiveWorkload, c.client)
	if err != nil {
		log.Error().Msgf("Failed to set metrics for interactive workload %s/%s, error: %s",
			interactiveWorkload.Namespace, interactiveWorkload.Name, err.Error())
		return ctrl.Result{Requeue: true, RequeueAfter: utils.RequeueTimeout}, nil
	}
	return ctrl.Result{}, nil
}
