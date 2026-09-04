/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package run

import (
	cutils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/go-operator/pkg/controller-utils"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type reconciliationPredicate struct{}

func (*reconciliationPredicate) Create(e event.CreateEvent) bool {
	return true
}

func (*reconciliationPredicate) Delete(e event.DeleteEvent) bool {
	return true
}

func (*reconciliationPredicate) Update(e event.UpdateEvent) bool {
	if _, isRunaiConfig := e.ObjectNew.(*runaiv1.RunaiConfig); isRunaiConfig {
		o := e.ObjectOld.(*runaiv1.RunaiConfig)
		n := e.ObjectNew.(*runaiv1.RunaiConfig)

		specChanged := o.GetGeneration() != n.GetGeneration()
		missingFinalizer := !cutils.HasFinalizer(n, runaiv1.Finalizer)

		return specChanged || missingFinalizer
	}
	if isOwnedByDefaultRunaiConfig(e.ObjectNew) {
		return true
	}
	if e.ObjectNew.GetNamespace() == runaiv1.DefaultNamespace &&
		cutils.IsOwnedByMe(e.ObjectNew, runaiv1.GroupVersion.String(), runaiv1.RunaiConfigKind) {
		return true
	}
	return false
}

func (*reconciliationPredicate) Generic(e event.GenericEvent) bool {
	return true
}
