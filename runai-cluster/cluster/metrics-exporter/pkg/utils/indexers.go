/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"

	v1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PodGroupAnnotationForPodPath = ".metadata.annotations.pod-group-name"
	podGroupAnnotationForPodName = "pod-group-name"
)

func SetIndexersManager(mgr ctrl.Manager) error {
	// This index will allow us to query for all pods related to this pod group
	err := mgr.GetFieldIndexer().IndexField(
		context.Background(), &v1.Pod{}, PodGroupAnnotationForPodPath,
		podGroupNameIndexerFunc)

	return err
}

func podGroupNameIndexerFunc(rawObj client.Object) []string {
	pod, conversionSuccess := rawObj.(*v1.Pod)
	if !conversionSuccess {
		return nil
	}

	podGroupName, foundPGAnnotation := pod.Annotations[podGroupAnnotationForPodName]
	if !foundPGAnnotation {
		return nil
	}
	return []string{podGroupName}
}
