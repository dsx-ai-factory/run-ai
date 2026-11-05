/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	"context"
	"fmt"
	"reflect"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/rs/zerolog/log"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	clientgocache "k8s.io/client-go/tools/cache"
)

func setupKartaStatusWatcher(ctx context.Context, cachedClient cachedclient.CachedClient, handler *ExternalWorkloadHandler) {
	informer, err := cachedClient.SetInformerForKind(ctx, kartav1alpha1.GroupVersion.WithKind("Karta"))
	if err != nil {
		log.Error().Err(err).Msg("Failed to watch Karta status mappings")
		return
	}
	if _, err := informer.AddEventHandler(kartaStatusEventHandler(handler)); err != nil {
		log.Error().Err(err).Msg("Failed to register Karta status mapping event handler")
	}
}

func kartaStatusEventHandler(handler *ExternalWorkloadHandler) clientgocache.ResourceEventHandler {
	return clientgocache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj any) {
			oldKarta, oldOK := oldObj.(*kartav1alpha1.Karta)
			newKarta, newOK := newObj.(*kartav1alpha1.Karta)
			if !oldOK || !newOK || !newKarta.DeletionTimestamp.IsZero() || oldKarta.UID != newKarta.UID {
				return
			}
			oldRoot := oldKarta.Spec.StructureDefinition.RootComponent
			newRoot := newKarta.Spec.StructureDefinition.RootComponent
			if oldRoot.StatusDefinition == nil || newRoot.StatusDefinition == nil ||
				!reflect.DeepEqual(oldRoot.Kind, newRoot.Kind) ||
				reflect.DeepEqual(oldRoot.StatusDefinition.StatusMappings, newRoot.StatusDefinition.StatusMappings) {
				return
			}
			// Queue the refresh itself so a transient list failure is retried, even
			// when no workload events arrive after the mapping changes.
			task := utils.NewExternalWorkloadTask(utils.RefreshKartaOperation,
				types.NamespacedName{Name: newKarta.Name}, kartav1alpha1.GroupVersion.WithKind("Karta"))
			handler.tasksQueue.Add(*task)
		},
	}
}

func (h *ExternalWorkloadHandler) refreshKartaWorkloads(ctx context.Context, task utils.ExternalWorkloadTask) error {
	karta := &kartav1alpha1.Karta{}
	if err := h.cachedClient.Get(ctx, task.Key, karta); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("get Karta %s for status refresh: %w", task.Key.Name, err)
	}
	gvk := utils.GetGvkOfKarta(karta)
	if gvk == nil || !karta.DeletionTimestamp.IsZero() {
		return nil
	}
	if _, registered := h.supportedGVKsHandlers[*gvk]; !registered {
		return nil
	}
	workloads := &v1alpha1.ExternalWorkloadList{}
	if err := h.cachedClient.List(ctx, workloads); err != nil {
		return fmt.Errorf("list ExternalWorkloads for Karta %s: %w", karta.Name, err)
	}
	for _, workload := range workloads.Items {
		kind := workload.Spec.OriginalWorkloadKind
		if kind == "" {
			kind = workload.Spec.Type
		}
		if workload.Spec.Group != gvk.Group || kind != gvk.Kind ||
			(workload.Spec.Version != "" && workload.Spec.Version != gvk.Version) ||
			workload.Spec.OriginalWorkloadName == "" || !workload.DeletionTimestamp.IsZero() {
			continue
		}
		key := types.NamespacedName{Name: workload.Spec.OriginalWorkloadName, Namespace: workload.Namespace}
		refresh := utils.NewExternalWorkloadTask(utils.UpdateOperation, key, *gvk).
			SetExternalWorkloadName(workload.Name)
		h.tasksQueue.Add(*refresh)
	}
	return nil
}
