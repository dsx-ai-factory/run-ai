/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"context"

	dinfw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/controller/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/metrics"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type DistributedInferenceWorkloadController struct {
	client client.Client
}

func NewDistributedInferenceWorkloadController(client client.Client) *DistributedInferenceWorkloadController {
	return &DistributedInferenceWorkloadController{
		client: client,
	}
}

func (c *DistributedInferenceWorkloadController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dinfw.DistributedInferenceWorkload{}).
		WithEventFilter(workloadsEventFilter()).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
			utils.RateLimiterBaseDelay, utils.RateLimiterMaxDelay)}).
		Complete(c)
}

func (c *DistributedInferenceWorkloadController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = log.Logger.WithContext(ctx)
	log.Debug().Msgf("Reconciling DistributedInferenceWorkload %s/%s", req.Name, req.Namespace)

	DistributedInferenceWorkload := &dinfw.DistributedInferenceWorkload{}
	isObjectFound, err := utils.GetObjectForRequest(ctx, req, "DistributedInferenceWorkload", DistributedInferenceWorkload, c.client)
	if err != nil {
		return ctrl.Result{}, err
	} else if !isObjectFound || !DistributedInferenceWorkload.DeletionTimestamp.IsZero() {
		metrics.ResetWorkloadMetrics(req.Name, req.Namespace, metrics.DistributedInferenceSource)
		return ctrl.Result{Requeue: false}, nil
	}

	err = metrics.SetDistributedInferenceWorkloadMetrics(ctx, DistributedInferenceWorkload, c.client)
	if err != nil {
		log.Error().Msgf("Failed to set metrics for distributed inference workload %s/%s, error: %s",
			DistributedInferenceWorkload.Namespace, DistributedInferenceWorkload.Name, err.Error())
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
