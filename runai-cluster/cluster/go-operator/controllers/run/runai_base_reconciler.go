/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package run

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/runai-cluster/cluster/go-operator/internal/run/utils"
	cu "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	kube_client "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/kube-client"
	me "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/multi-error"
	runaiv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type RunaiBaseReconcilerI interface {
	controllerDeferFunction(ctx context.Context, runaiCustomType runaiv1.RunaiCustomTypeI, multiErr *me.MultiError, successRes ctrl.Result, errRes ctrl.Result) (ctrl.Result, error)
	updateStatus(ctx context.Context, runaiCustomType runaiv1.RunaiCustomTypeI, multiErr *me.MultiError) error
}

type runaiReconciler struct {
	localClusterClient kube_client.HybridClient
	scheme             *runtime.Scheme
	verbosity          int
}

func (r *runaiReconciler) controllerDeferFunction(ctx context.Context, runaiCustomType runaiv1.RunaiCustomTypeI, multiErr *me.MultiError, res ctrl.Result) (ctrl.Result, error) {
	logger := log.FromContext(ctx).V(r.verbosity)
	if !runaiCustomType.GetDeletionTimestamp().IsZero() {
		//Stop reconcile if the resource is deleted
		return ctrl.Result{}, nil
	}
	if !multiErr.Empty() {
		cu.SetCondition(runaiCustomType, runaiv1.ConditionTypeReconciled,
			metav1.ConditionFalse,
			runaiv1.NotReconciled,
			multiErr.Error(), runaiCustomType.GetGeneration())
	} else {
		cu.SetCondition(runaiCustomType, runaiv1.ConditionTypeReconciled,
			metav1.ConditionTrue,
			runaiv1.Reconciled,
			runaiv1.ConditionTypeReconciledSuccessMessage, runaiCustomType.GetGeneration())
	}

	runaiCustomType.GetRunaiCustomStatus().ObservedGeneration = runaiCustomType.GetGeneration()
	err := r.updateStatus(ctx, runaiCustomType, multiErr)
	if err != nil {
		if multiErr.Empty() {
			err = fmt.Errorf("status update error: [%v]", err)
		} else {
			err = fmt.Errorf("status update error: [%v] reconciliation error: %v", err, multiErr)
		}
	} else {
		logger.Info(fmt.Sprintf("Updated status conditions for %s",
			utils.ResourceUniqueID(runaiCustomType)))
	}
	return res, err
}

func (r *runaiReconciler) updateStatus(ctx context.Context, runaiCustomType runaiv1.RunaiCustomTypeI, multiErr *me.MultiError) error {
	err := r.localClusterClient.Status().Update(ctx, runaiCustomType)
	if err != nil {
		if multiErr.Empty() {
			return fmt.Errorf("status update error: [%v]", err)
		} else {
			return fmt.Errorf("status update error: [%v] reconciliation error: %v", err, multiErr)
		}
	}
	return nil
}
