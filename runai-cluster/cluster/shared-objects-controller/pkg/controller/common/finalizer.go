/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	"github.com/run-ai/runai/runai-cluster/cluster/shared-objects-controller/pkg/controller/utils"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *CommonReconciler) AddControllerAsFinalizerIfNeeded(ctx context.Context, object client.Object) error {
	if utils.IsItemInList(object.GetFinalizers(), c.finalizerName) {
		return nil
	}

	log.Info().Msgf("Adding controller to finalizers list of %s", ObjectLogStr(object))
	object.SetFinalizers(append(object.GetFinalizers(), c.finalizerName))

	if err := c.updateFinalizers(ctx, object); err != nil {
		log.Error().Msgf("Error adding controller to finalizers list of %s, error: %s",
			ObjectLogStr(object), err.Error())
		return err
	}
	return nil
}

func (c *CommonReconciler) Finalize(ctx context.Context, object client.Object) error {
	log.Info().Msgf("Finalizing deletion of %s", ObjectLogStr(object))

	if !utils.IsItemInList(object.GetFinalizers(), c.finalizerName) {
		log.Info().Msgf("Got a finalization event for %s, "+
			"but this controller is not in its finalizers list. Skipping finalization.", ObjectLogStr(object))
		return nil
	}

	objWithUpdatedStatus, err := c.objectReconciler.HandleDeletion(ctx, object)
	if err != nil {
		innerErr := c.objectReconciler.UpdateStatus(ctx, object, objWithUpdatedStatus)
		if innerErr != nil {
			err = multierror.Append(err, innerErr)
		}
		return err
	}

	err = c.deleteFinalizer(ctx, object)
	if err != nil {
		return err
	}

	log.Info().Msgf("Done handling deletion of %s", ObjectLogStr(object))
	return err
}

func (c *CommonReconciler) deleteFinalizer(ctx context.Context, object client.Object) error {
	log.Info().Msgf("Removing finalizer <%v> from finalizers list of %s",
		c.finalizerName, ObjectLogStr(object))
	object.SetFinalizers(utils.DeleteFirstFromList(object.GetFinalizers(), c.finalizerName))

	if err := c.updateFinalizers(ctx, object); err != nil {
		log.Error().Msgf("Error removing controller from finalizer list for %s; err: %s",
			ObjectLogStr(object), err.Error())
		return err
	}
	return nil
}

func (c *CommonReconciler) updateFinalizers(ctx context.Context, object client.Object) (err error) {
	patchBytes, err := updateFinalizersPatchBytes(object.GetFinalizers())
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - update finalizers of %s, err: %v",
			ObjectLogStr(object), err)
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = c.client.Patch(ctx, object, patch)
	if err != nil {
		return err
	}
	return nil
}

func updateFinalizersPatchBytes(finalizers []string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"metadata": map[string]interface{}{
		"finalizers": finalizers}})
}
