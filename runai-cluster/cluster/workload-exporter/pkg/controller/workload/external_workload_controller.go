/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/controller/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-exporter/pkg/metrics"
	"github.com/rs/zerolog/log"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

type ExternalWorkloadController struct {
	client client.Client
}

func NewExternalWorkloadController(client client.Client) *ExternalWorkloadController {
	return &ExternalWorkloadController{
		client: client,
	}
}

func (c *ExternalWorkloadController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.ExternalWorkload{}).
		WithEventFilter(workloadsEventFilter()).
		WithOptions(controller.Options{RateLimiter: workqueue.NewTypedItemExponentialFailureRateLimiter[ctrl.Request](
			utils.RateLimiterBaseDelay, utils.RateLimiterMaxDelay)}).
		Complete(c)
}

func (c *ExternalWorkloadController) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	ctx = log.Logger.WithContext(ctx)
	log.Debug().Msgf("Reconciling externalworkload %s/%s", req.Name, req.Namespace)

	externalWorkload := &v1alpha1.ExternalWorkload{}
	isObjectFound, err := utils.GetObjectForRequest(ctx, req, "externalworkload", externalWorkload, c.client)
	if err != nil {
		return ctrl.Result{}, err
	} else if !isObjectFound || !externalWorkload.DeletionTimestamp.IsZero() {
		metrics.ResetWorkloadMetrics(req.Name, req.Namespace, metrics.ExternalSource)
		return ctrl.Result{Requeue: false}, nil
	}

	err = metrics.SetExternalWorkloadMetrics(ctx, externalWorkload, c.client)
	if err != nil {
		log.Error().Msgf("Failed to set metrics for external workload %s/%s, error: %s",
			externalWorkload.Namespace, externalWorkload.Name, err.Error())
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
