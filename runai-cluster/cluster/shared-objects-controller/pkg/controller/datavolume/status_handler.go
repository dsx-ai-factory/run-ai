/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datavolume

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *DatavolumeReconciler) UpdateStatus(ctx context.Context,
	object, objectWithUpdatedStatus client.Object) error {
	datavolume, ok := object.(*v1alpha1.DataVolume)
	if !ok {
		log.Ctx(ctx).Error().Msgf("Expected DataVolume but got: %T", object)
		return nil
	}
	datavolumeWithUpdatedStatus, ok := objectWithUpdatedStatus.(*v1alpha1.DataVolume)
	if !ok {
		log.Ctx(ctx).Error().Msgf("Expected DataVolume but got: %T", object)
		return nil
	}

	return c.updateDatavolumeStatus(ctx, datavolume, datavolumeWithUpdatedStatus)
}

func (c *DatavolumeReconciler) updateDatavolumeStatus(ctx context.Context,
	datavolume, datavolumeWithUpdatedStatus *v1alpha1.DataVolume) error {
	log.Ctx(ctx).Debug().Msgf("Reconciling the status of datavolume <%s>", datavolume.Name)

	datavolumeOldStatusCopy := datavolume.Status.DeepCopy()
	datavolume.Status = *datavolumeWithUpdatedStatus.Status.DeepCopy()
	return c.updateStatusIfNeeded(ctx, datavolume, datavolumeOldStatusCopy)
}

func (c *DatavolumeReconciler) updateStatusIfNeeded(
	ctx context.Context, datavolume *v1alpha1.DataVolume, datavolumeStatusCopy *v1alpha1.DataVolumeStatus) error {
	if datavolume.Status.Phase == datavolumeStatusCopy.Phase &&
		datavolume.Status.Message == datavolumeStatusCopy.Message &&
		datavolume.Status.DatavolumePvcName == datavolumeStatusCopy.DatavolumePvcName &&
		datavolume.Status.DatavolumePvName == datavolumeStatusCopy.DatavolumePvName {
		// nothing to update
		return nil
	}

	return c.updateStatus(ctx, datavolume)
}

func (c *DatavolumeReconciler) updateStatus(ctx context.Context, datavolume *v1alpha1.DataVolume) error {
	log.Ctx(ctx).Info().Msgf("Updating status of datavolume <%s> to: <%s/%s/%s/%s>",
		datavolume.Name, datavolume.Status.Phase, datavolume.Status.Message,
		datavolume.Status.DatavolumePvcName, datavolume.Status.DatavolumePvName)

	patchBytes, err := updateStatusPatchBytes(&datavolume.Status)
	if err != nil {
		log.Error().Msgf("Failed to json.Marshal patch - update status of datavolume %s, err: %v",
			datavolume.Name, err)
		return err
	}

	patch := client.RawPatch(types.MergePatchType, patchBytes)
	err = c.client.Status().Patch(ctx, datavolume, patch)
	if err != nil {
		return err
	}
	return nil
}

func updateStatusPatchBytes(status *v1alpha1.DataVolumeStatus) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"status": map[string]interface{}{
		"phase":             status.Phase,
		"message":           status.Message,
		"datavolumePvcName": status.DatavolumePvcName,
		"datavolumePvName":  status.DatavolumePvName,
	}})
}
