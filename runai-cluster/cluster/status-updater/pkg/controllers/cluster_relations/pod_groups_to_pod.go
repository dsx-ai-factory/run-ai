/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_relations

import (
	"context"
	"fmt"

	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	PodGroupAnnotationForPodName = "pod-group-name"
	PodGroupAnnotationForPodPath = ".metadata.annotations.pod-group-name"
)

func GetPodGroupFromPod(ctx context.Context, pod *v1.Pod, kubeClient client.Client) (*kaiv2alpha2.PodGroup, error) {
	podGroupName, foundPGAnnotation := pod.Annotations[PodGroupAnnotationForPodName]
	if !foundPGAnnotation {
		return nil, fmt.Errorf("annotation %s connection pod <%s/%s> to podgroup not found",
			PodGroupAnnotationForPodName, pod.Namespace, pod.Name)
	}

	podGroup := kaiv2alpha2.PodGroup{}
	err := kubeClient.Get(ctx, types.NamespacedName{Namespace: pod.Namespace, Name: podGroupName}, &podGroup)
	if err != nil {
		return nil, err
	}

	return &podGroup, nil
}

func GetPodGroupFromDeletedPodObject(ctx context.Context, deletedPodObject client.Object, kubeClient client.Client) (
	*kaiv2alpha2.PodGroup, error) {
	podGroupName, foundPGAnnotation := deletedPodObject.GetAnnotations()[PodGroupAnnotationForPodName]
	if !foundPGAnnotation {
		return nil, fmt.Errorf("annotation %s connection pod <%s/%s> to podgroup not found",
			PodGroupAnnotationForPodName, deletedPodObject.GetNamespace(), deletedPodObject.GetName())
	}

	podGroup := kaiv2alpha2.PodGroup{}
	err := kubeClient.Get(ctx,
		types.NamespacedName{Namespace: deletedPodObject.GetNamespace(), Name: podGroupName},
		&podGroup)
	if err != nil {
		return nil, err
	}

	return &podGroup, nil
}

func GetAllPodsOfPodGroup(ctx context.Context, podGroup *kaiv2alpha2.PodGroup, kubeClient client.Client) (
	v1.PodList, error) {
	podGroupPods := v1.PodList{}
	err := kubeClient.List(ctx, &podGroupPods, client.MatchingFields{PodGroupAnnotationForPodPath: podGroup.Name})
	if err != nil {
		return v1.PodList{}, err
	}
	return podGroupPods, nil
}
