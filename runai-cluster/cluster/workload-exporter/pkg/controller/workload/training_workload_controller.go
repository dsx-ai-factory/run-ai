/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"context"

	twv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/controller/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/metrics"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type TrainingWorkloadController struct {
	client client.Client
}

func NewTrainingWorkloadController(client client.Client) *TrainingWorkloadController {
	return &TrainingWorkloadController{
		client: client,
	}
}

func (c *TrainingWorkloadController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&twv2a1.TrainingWorkload{}).
		WithEventFilter(workloadsEventFilter()).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
			utils.RateLimiterBaseDelay, utils.RateLimiterMaxDelay)}).
		Complete(c)
}

func (c *TrainingWorkloadController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = log.Logger.WithContext(ctx)
	log.Debug().Msgf("Reconciling trainingworkload %s/%s", req.Name, req.Namespace)

	trainingWorkload := &twv2a1.TrainingWorkload{}
	isObjectFound, err := utils.GetObjectForRequest(ctx, req, "trainingworkload", trainingWorkload, c.client)
	if err != nil {
		return ctrl.Result{}, err
	} else if !isObjectFound || !trainingWorkload.DeletionTimestamp.IsZero() {
		metrics.ResetWorkloadMetrics(req.Name, req.Namespace, metrics.TrainingSource)
		return ctrl.Result{Requeue: false}, nil
	}

	err = metrics.SetTrainingWorkloadMetrics(ctx, trainingWorkload, c.client)
	if err != nil {
		log.Error().Msgf("Failed to set metrics for training workload %s/%s, error: %s",
			trainingWorkload.Namespace, trainingWorkload.Name, err.Error())
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
