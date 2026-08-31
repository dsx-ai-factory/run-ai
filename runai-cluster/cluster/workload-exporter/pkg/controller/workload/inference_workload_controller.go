/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"context"

	"github.com/rs/zerolog/log"
	infwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-exporter/pkg/controller/utils"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-exporter/pkg/metrics"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type InferenceWorkloadController struct {
	client client.Client
}

func NewInferenceWorkloadController(client client.Client) *InferenceWorkloadController {
	return &InferenceWorkloadController{
		client: client,
	}
}

func (c *InferenceWorkloadController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infwv2a1.InferenceWorkload{}).
		WithEventFilter(workloadsEventFilter()).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
			utils.RateLimiterBaseDelay, utils.RateLimiterMaxDelay)}).
		Complete(c)
}

func (c *InferenceWorkloadController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = log.Logger.WithContext(ctx)
	log.Debug().Msgf("Reconciling inferenceworkload %s/%s", req.Name, req.Namespace)

	inferenceWorkload := &infwv2a1.InferenceWorkload{}
	isObjectFound, err := utils.GetObjectForRequest(ctx, req, "inferenceworkload", inferenceWorkload, c.client)
	if err != nil {
		return ctrl.Result{Requeue: true, RequeueAfter: utils.RequeueTimeout}, nil
	} else if !isObjectFound || !inferenceWorkload.DeletionTimestamp.IsZero() {
		metrics.ResetWorkloadMetrics(req.Name, req.Namespace, metrics.InferenceSource)
		return ctrl.Result{Requeue: false}, nil
	}

	err = metrics.SetInferenceWorkloadMetrics(ctx, inferenceWorkload, c.client)
	if err != nil {
		log.Error().Msgf("Failed to set metrics for inference workload %s/%s, error: %s",
			inferenceWorkload.Namespace, inferenceWorkload.Name, err.Error())
		return ctrl.Result{Requeue: true, RequeueAfter: utils.RequeueTimeout}, nil
	}
	return ctrl.Result{}, nil
}
