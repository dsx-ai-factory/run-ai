/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kartacontroller

import (
	"context"
	"os"

	"github.com/rs/zerolog/log"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/kartacontroller/handlers"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/supportedhandlers"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/utils"
	runv1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// exitFn is a package-level indirection for process exit to enable testing
var exitFn = os.Exit

func (rc *KartaController) reconcile(ctx context.Context, karta *kartav1alpha1.Karta) error {
	originalStatus := karta.Status.DeepCopy()

	restartNeeded, err := rc.reconcileSpec(ctx, karta)
	// if error - logged inside
	//continue to update status

	statusUpdateErr := rc.patchStatusWithNewConditions(ctx, karta, originalStatus)
	if statusUpdateErr != nil {
		if err == nil {
			err = statusUpdateErr
		}
		log.Error().Msgf("Failed to update status for Karta <%s>, error: %v", karta.Name, statusUpdateErr)
	}

	if restartNeeded {
		log.Info().Msgf("Restarting external workload integrator to apply changes for Karta <%s>", karta.Name)
		exitFn(0)
	}

	return err
}

// reconcileSpec
// return params:
// - restartNeeded - true if the controller needs to restart to apply changes
// - error         - if any error occurred during reconciliation
func (rc *KartaController) reconcileSpec(ctx context.Context, karta *kartav1alpha1.Karta) (bool, error) {
	conditionStatuses := handlers.KartaConditionStatuses{
		RBACReady:   metav1.ConditionFalse,
		SyncEnabled: metav1.ConditionFalse,
	}
	defer func() {
		handlers.SetConditions(&karta.Status, conditionStatuses)
	}()

	gvk := utils.GetGvkOfKarta(karta)
	if gvk == nil {
		log.Error().Msgf("Karta <%s> has no root component kind defined, skipping reconciliation", karta.Name)
		return false, nil
	}

	if !kartaHasExistingCRD(karta) {
		return false, nil
	}

	_, isAlreadySupported := rc.isAlreadySupportedAsBuiltInGvk(*gvk)

	err := rc.rbacHandler.Handle(ctx, *gvk, utils.GetAdditionalGvksOfKarta(karta))
	if err != nil {
		log.Error().Msgf("Failed to handle RBAC for Karta <%s> with GVK <%s>, error: %s", karta.Name, gvk.String(), err.Error())
		return false, err
	}
	conditionStatuses.RBACReady = metav1.ConditionTrue

	if isAlreadySupported {
		conditionStatuses.SyncEnabled = metav1.ConditionTrue
		log.Debug().Msgf("Karta <%s> is already supported, skipping rest of reconciliation", karta.Name)
		return false, nil
	}

	isRegistered := rc.isGvkRegisteredAsAdditional(*gvk)
	if !isRegistered {
		// restartNeeded = true
		// SyncEnabled condition should stay False
		return true, nil
	}
	conditionStatuses.SyncEnabled = metav1.ConditionTrue

	return false, nil
}

// kartaHasExistingCRD reports whether the OSS controller has set CRDExists=True on the Karta.
func kartaHasExistingCRD(karta *kartav1alpha1.Karta) bool {
	for _, c := range karta.Status.Conditions {
		if c.Type == string(runv1alpha1.CRDExists) {
			return c.Status == metav1.ConditionTrue
		}
	}
	return false
}

// isAlreadySupportedAsBuiltInGvk
// return params:
// - isCrd       - true if the GVK is a CRD
// - isSupported - true if the GVK is supported by the handlers
func (rc *KartaController) isAlreadySupportedAsBuiltInGvk(kartaGvk schema.GroupVersionKind) (bool, bool) {
	supportedGVKsHandlers := supportedhandlers.GetSupportedGVKsHandlersWithConfig(nil, nil)
	if supported, ok := supportedGVKsHandlers[kartaGvk]; ok {
		return supported.IsCrd, true
	}

	return false, false
}

func (rc *KartaController) isGvkRegisteredAsAdditional(gvk schema.GroupVersionKind) bool {
	_, exists := rc.registeredAdditionalGVKs[gvk]
	return exists
}
