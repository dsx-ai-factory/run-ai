/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package kartacontroller

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/utils"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	FinalizerName = "karta.run.ai/finalize"
)

func (rc *KartaController) finalize(ctx context.Context, karta *kartav1alpha1.Karta) error {
	log.Info().Msgf("Finalizing deletion of Karta <%v>", karta.Name)

	if !utils.IsItemInList(karta.Finalizers, FinalizerName) {
		log.Info().Msgf("Got a finalization event for Karta <%s>, "+
			"but this controller is not in its finalizers list. Skipping finalization.", karta.Name)
		return nil
	}

	originalStatus := karta.Status.DeepCopy()
	err := rc.deletionHandler.Handle(ctx, karta)
	if err != nil {
		statusUpdateErr := rc.patchStatusWithNewConditions(ctx, karta, originalStatus)
		if statusUpdateErr != nil {
			err = multierror.Append(err, statusUpdateErr)
		}
		return err
	}

	err = rc.deleteFinalizer(ctx, karta)
	if err != nil {
		return err
	}

	gvk := utils.GetGvkOfKarta(karta)
	if gvk != nil {
		delete(rc.registeredAdditionalGVKs, *gvk)
	}

	log.Info().Msgf("Done handling deletion of Karta <%s>", karta.Name)
	return err
}

func (rc *KartaController) addControllerAsFinalizerIfNeeded(ctx context.Context, karta *kartav1alpha1.Karta) error {
	if utils.IsItemInList(karta.Finalizers, FinalizerName) {
		return nil
	}

	log.Info().Msgf("Adding controller to finalizers list of Karta <%s>", karta.Name)
	karta.Finalizers = append(karta.Finalizers, FinalizerName)

	if err := rc.updateFinalizers(ctx, karta); err != nil {
		log.Error().Msgf("Error adding controller to finalizers list of Karta <%s>, error: %s",
			karta.Name, err.Error())
		return err
	}
	return nil
}

func (rc *KartaController) deleteFinalizer(ctx context.Context, karta *kartav1alpha1.Karta) error {
	log.Info().Msgf("Removing finalizer <%v> from Karta's <%v> finalizers list",
		FinalizerName, karta.Name)
	karta.Finalizers = utils.DeleteFirstFromList(karta.Finalizers, FinalizerName)

	if err := rc.updateFinalizers(ctx, karta); err != nil {
		log.Error().Msgf("Error removing controller from finalizer list for Karta <%s>; err: %s",
			karta.Name, err.Error())
		return err
	}
	return nil
}

func (rc *KartaController) updateFinalizers(ctx context.Context, karta *kartav1alpha1.Karta) (err error) {
	patchBytes, err := updateFinalizersPatchBytes(karta.Finalizers)
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - update finalizers of karta <%v>, err: %v",
			karta.Name, err)
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = rc.Client.Patch(ctx, karta, patch)
	if err != nil {
		return err
	}
	return nil
}

func updateFinalizersPatchBytes(finalizers []string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"metadata": map[string]interface{}{
		"finalizers": finalizers}})
}
