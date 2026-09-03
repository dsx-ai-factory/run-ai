/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"context"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type KartaDeletionHandler struct {
	Client client.Client

	rbacHandler              *KartaRBACHandler
	registeredAdditionalGVKs map[schema.GroupVersionKind]bool
	informersApi             InformersApi
}

func NewKartaDeletionHandler(client client.Client, rbacHandler *KartaRBACHandler, registeredAdditionalGVKs map[schema.GroupVersionKind]bool, informersApi InformersApi) *KartaDeletionHandler {
	return &KartaDeletionHandler{
		Client:                   client,
		rbacHandler:              rbacHandler,
		registeredAdditionalGVKs: registeredAdditionalGVKs,
		informersApi:             informersApi,
	}
}

func (h *KartaDeletionHandler) Handle(ctx context.Context, karta *kartav1alpha1.Karta) error {
	gvk := utils.GetGvkOfKarta(karta)
	if gvk == nil {
		return nil
	}

	var err error
	conditionStatuses := KartaDeletionConditionStatuses{
		SyncDisabled:        metav1.ConditionFalse,
		NoWorkloadsUseKarta: metav1.ConditionFalse,
		RBACCleanupSuccess:  metav1.ConditionFalse,
	}
	defer func() {
		if err != nil {
			setDeletionConditions(&karta.Status, conditionStatuses)
		}
	}()

	registeredGvksOfGv, isGkRegistered := getGvksMapPerGroupKind(h.registeredAdditionalGVKs)[gvk.GroupKind()]

	err = h.informersApi.RemoveInformer(ctx, *gvk)
	if err != nil {
		return fmt.Errorf("failed to remove informer for GVK <%s> of Karta <%s>, error: %s",
			gvk.String(), karta.Name, err.Error())
	}
	conditionStatuses.SyncDisabled = metav1.ConditionTrue

	if isGkRegistered && len(registeredGvksOfGv) == 1 {
		// only validate workloads if this is the last GVK registered for this GK
		err = h.validateGvkNotInUseByExternalWorkloads(ctx, gvk.GroupKind(), karta.Name)
		if err != nil {
			return err
		}
	}
	conditionStatuses.NoWorkloadsUseKarta = metav1.ConditionTrue

	if isGkRegistered && len(registeredGvksOfGv) == 1 {
		// only clean RBAC if this is the last GVK registered for this GK
		err = h.rbacHandler.HandleDeletion(ctx, *gvk, utils.GetAdditionalGvksOfKarta(karta))
		if err != nil {
			return fmt.Errorf("failed to handle RBAC cleanup for GVK <%s> of Karta <%s>, error: %s",
				gvk.String(), karta.Name, err.Error())
		}
	}
	conditionStatuses.RBACCleanupSuccess = metav1.ConditionTrue

	return nil
}

func (h *KartaDeletionHandler) validateGvkNotInUseByExternalWorkloads(ctx context.Context, gk schema.GroupKind, ridName string) error {
	ews, err := h.listExternalWorkloadsOfGk(ctx, gk)
	if err != nil {
		return fmt.Errorf("failed to list workloads of GK <%s>, error: %s", gk.String(), err.Error())
	}

	if len(ews) > 0 {
		return fmt.Errorf("cannot delete Karta <%s> for GK <%s> as it is still used by <%d> ExternalWorkloads",
			ridName, gk.String(), len(ews))
	}

	return nil
}

func (h *KartaDeletionHandler) listExternalWorkloadsOfGk(ctx context.Context, gk schema.GroupKind) ([]v1alpha1.ExternalWorkload, error) {
	ews := &v1alpha1.ExternalWorkloadList{}
	err := h.Client.List(ctx, ews)
	if err != nil {
		return nil, fmt.Errorf("failed to list ExternalWorkloads, error: %s", err.Error())
	}

	result := []v1alpha1.ExternalWorkload{}
	for _, ew := range ews.Items {
		if ew.Spec.Type == gk.Kind && ew.Spec.Group == gk.Group {
			result = append(result, ew)
		}
	}
	return result, nil
}
