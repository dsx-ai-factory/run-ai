/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datavolume

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c *DatavolumeReconciler) HandleDeletion(ctx context.Context, object client.Object) (client.Object, error) {
	datavolume, ok := object.(*v1alpha1.DataVolume)
	if !ok {
		log.Ctx(ctx).Error().Msgf("Expected DataVolume but got: %T", object)
		return object, nil
	}

	var err error
	if innerErr := c.cleanupDatavolumePvcs(ctx, datavolume); innerErr != nil {
		err = multierror.Append(err, innerErr)
	}
	if innerErr := c.cleanupDatavolumePvs(ctx, datavolume); innerErr != nil {
		err = multierror.Append(err, innerErr)
	}

	return datavolume, err
}

func (c *DatavolumeReconciler) cleanupDatavolumePvcs(ctx context.Context, datavolume *v1alpha1.DataVolume) error {
	pvcs, err := c.pvcHandler.ListPvcsOfDatavolume(ctx, datavolume)
	if err != nil {
		return err
	}

	var result error
	for i := range pvcs {
		pvc := &pvcs[i]
		if datavolume.IsOwner(pvc) != -1 {
			if innerErr := c.Delete(ctx, pvc); innerErr != nil && !errors.IsNotFound(innerErr) {
				result = multierror.Append(result, innerErr)
			}
			continue
		}

		if innerErr := c.removeDatavolumeLabelsFromObject(ctx, pvc, datavolume); innerErr != nil {
			result = multierror.Append(result, innerErr)
		}
	}
	return result
}

func (c *DatavolumeReconciler) cleanupDatavolumePvs(ctx context.Context, datavolume *v1alpha1.DataVolume) error {
	pvs, err := c.pvHandler.ListPvsOfDatavolume(ctx, datavolume)
	if err != nil {
		return err
	}

	var result error
	for i := range pvs {
		pv := &pvs[i]
		if datavolume.IsOwner(pv) != -1 {
			if innerErr := c.Delete(ctx, pv); innerErr != nil && !errors.IsNotFound(innerErr) {
				result = multierror.Append(result, innerErr)
			}
			continue
		}

		if innerErr := c.removeDatavolumeLabelsFromObject(ctx, pv, datavolume); innerErr != nil {
			result = multierror.Append(result, innerErr)
		}
	}
	return result
}

func (c *DatavolumeReconciler) removeDatavolumeLabelsFromObject(ctx context.Context, object client.Object, datavolume *v1alpha1.DataVolume) error {
	objectCopy, ok := object.DeepCopyObject().(client.Object)
	if !ok {
		log.Ctx(ctx).Error().Msgf("Expected client.Object but got: %T", object.DeepCopyObject())
		return nil
	}

	patch := client.MergeFrom(objectCopy)
	if !removeDatavolumeLabels(object, datavolume) {
		return nil
	}

	return c.client.Patch(ctx, object, patch)
}

func removeDatavolumeLabels(object client.Object, datavolume *v1alpha1.DataVolume) bool {
	labels := object.GetLabels()
	if labels == nil {
		return false
	}

	changed := false
	updatedLabels := make(map[string]string, len(labels))
	for key, value := range labels {
		updatedLabels[key] = value
	}

	if datavolumeId, ok := labels[v1alpha1.DataVolumeIdLabel]; ok && datavolumeId == datavolume.Spec.DatavolumeId {
		delete(updatedLabels, v1alpha1.DataVolumeIdLabel)
		changed = true
	}
	if datavolumeName, ok := labels[v1alpha1.DataVolumeNameLabel]; ok && datavolumeName == datavolume.Name {
		delete(updatedLabels, v1alpha1.DataVolumeNameLabel)
		changed = true
	}
	if changed {
		delete(updatedLabels, v1alpha1.DataVolumePvForNamespaceLabel)
	}

	if changed {
		object.SetLabels(updatedLabels)
	}
	return changed
}
