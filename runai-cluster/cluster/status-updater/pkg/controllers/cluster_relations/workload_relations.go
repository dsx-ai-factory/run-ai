/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_relations

import (
	"context"

	"github.com/run-ai/runai/runai-cluster/cluster/status-updater/pkg/controllers/metadata_calculator/workload"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"github.com/run-ai/runai/runai-cluster/cluster/status-updater/pkg/controllers/metadata"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetWorkloadData(ctx context.Context, ownerRef *metadata.TopOwnerReference, namespace string,
	kubeClient client.Client, dynamicClient dynamic.Interface, discoveryClient discovery.DiscoveryInterface) (
	*metadata.WorkloadObjects, error) {
	ownerObject, groupVersionResource, err := getOwnerFromTopOwnerReference(ctx, ownerRef, namespace, dynamicClient,
		discoveryClient)
	if err != nil {
		return nil, err
	}

	workloadPodGroups, err := getAllPodGroupsOfOwnerObj(ctx, ownerObject, kubeClient)
	if err != nil {
		return nil, err
	}

	var workloadPods []v1.Pod
	var latestChiefPod *v1.Pod
	for _, item := range workloadPodGroups.Items {
		workloadPg := item
		pgPods, err := GetAllPodsOfPodGroup(ctx, &workloadPg, kubeClient)
		if err != nil {
			return nil, err
		}

		workloadPods = append(workloadPods, pgPods.Items...)
		latestChiefPod = workload.CalcLatestChiefPod(pgPods, workloadPg, latestChiefPod)
	}

	return &metadata.WorkloadObjects{
		GroupVersionResource: groupVersionResource,
		Namespace:            namespace,
		WorkloadObj:          ownerObject,
		PodGroups:            workloadPodGroups.Items,
		Pods:                 workloadPods,
		LatestChiefPod:       latestChiefPod,
	}, nil
}
