/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AssetsObjectInspector bool

func (ai AssetsObjectInspector) DeepCopySpec(from, to client.Object) {
	switch to.GetObjectKind().GroupVersionKind().Kind {
	case constants.KindPvc:
		fromPvc, ok := from.(*v1.PersistentVolumeClaim)
		if !ok {
			return
		}
		toPvc, ok := to.(*v1.PersistentVolumeClaim)
		if !ok {
			return
		}

		// in case spec of the PVC was updated by the volume name, we should preserve it
		if toPvc.Spec.VolumeName != "" {
			fromPvc.Spec.VolumeName = toPvc.Spec.VolumeName
		}

		fromPvc.Spec.DeepCopyInto(&toPvc.Spec)

	case constants.KindSecret:
		fromSecret, ok := from.(*v1.Secret)
		if !ok {
			return
		}
		toSecret, ok := to.(*v1.Secret)
		if !ok {
			return
		}

		if fromSecret.Data != nil {
			if toSecret.Data == nil {
				toSecret.Data = make(map[string][]byte)
			}

			for k, v := range fromSecret.Data {
				toSecret.Data[k] = v
			}
		}
	}

	// annotations
	annotations := from.GetAnnotations()

	if from.GetAnnotations() == nil {
		annotations = make(map[string]string)
	}

	for key, val := range from.GetAnnotations() {
		annotations[key] = val
	}

	to.SetAnnotations(annotations)

	// labels
	labels := from.GetLabels()

	if from.GetLabels() == nil {
		labels = make(map[string]string)
	}

	for key, val := range from.GetLabels() {
		labels[key] = val
	}

	to.SetLabels(labels)

}

func (ai AssetsObjectInspector) IsUpToDate(left, right client.Object) bool {
	if left == nil || right == nil {
		return false
	}

	rightAnnotations := right.GetAnnotations()

	rightCounterStr, ok := rightAnnotations[UpdateCountAnnotation]
	if !ok || rightCounterStr == "" {
		return false
	}

	rightAnnotationsCounter, err := strconv.Atoi(rightCounterStr)
	if err != nil {
		log.Warn().Msgf("Had an error while converting update counter to int. object annotation: %s , err: %s", rightCounterStr, err.Error())
		return false
	}

	leftAnnotations := left.GetAnnotations()

	leftCounterStr, ok := leftAnnotations[UpdateCountAnnotation]
	if !ok || leftCounterStr == "" {
		return false
	}

	leftAnnotationsCounter, err := strconv.Atoi(leftCounterStr)
	if err != nil {
		log.Warn().Msgf("Had an error while converting update counter to int. object annotation: %s , err: %s", leftCounterStr, err.Error())
		return false
	}

	if leftAnnotationsCounter <= rightAnnotationsCounter {
		// if the update counter is the same or higher than the cluster in-object counter,
		// it means that the update from the backend is newer than the revision of the cluster object
		return true
	}

	return false
}
