/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"context"
	"fmt"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/record"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clusterrelations "github.com/run-ai/runai/runai-cluster/cluster/status-updater/pkg/controllers/cluster_relations"
	"github.com/run-ai/runai/runai-cluster/cluster/status-updater/pkg/controllers/metadata"
	podmetacalc "github.com/run-ai/runai/runai-cluster/cluster/status-updater/pkg/controllers/metadata_calculator/pod"
	workloadcalc "github.com/run-ai/runai/runai-cluster/cluster/status-updater/pkg/controllers/metadata_calculator/workload"
	"github.com/run-ai/runai/runai-cluster/cluster/status-updater/pkg/controllers/patcher"
)

func updatePod(ctx context.Context, pod *v1.Pod, kubeClient client.Client) error {
	logger := log.FromContext(ctx)

	podMeta := metadata.PodMetadata{}
	podMeta.Status = podmetacalc.TranslatePodStatus(pod)

	err := podmetacalc.CalcPodGpuResources(ctx, pod, &podMeta, kubeClient)
	if err != nil {
		return fmt.Errorf("failed to calculate gpu resources for pod: <%v/%v> with error : %w",
			pod.Namespace, pod.Name, err)
	}

	if patcher.ShouldUpdatePodAnnotations(pod, podMeta) {
		err = patcher.UpdatePodAnnotations(ctx, pod, podMeta, kubeClient)
		if err == nil {
			logger.V(1).Info(fmt.Sprintf("Updated pod status and meta: <%v/%v> status: %s", pod.Namespace,
				pod.Name, podMeta.Status))
		}

		return err
	}

	return nil
}

func updateWorkloadData(ctx context.Context, ownerRef *metadata.TopOwnerReference, namespace string,
	kubeClient client.Client, dynamicClient dynamic.Interface, discoveryClient discovery.DiscoveryInterface,
	eventRecorder record.EventRecorder) error {

	workloadData, err := clusterrelations.GetWorkloadData(ctx, ownerRef, namespace, kubeClient, dynamicClient,
		discoveryClient)
	if err != nil {
		return err
	}

	workloadMetadata := metadata.CreateWorkloadMetadata()
	if err = workloadcalc.CalcWorkloadStatus(ctx, workloadData, kubeClient, workloadMetadata); err != nil {
		return err
	}
	workloadcalc.CalcPodsUsedNodesMetadata(workloadMetadata, workloadData.Pods)
	workloadcalc.SumOwnedPodGroupResources(workloadMetadata, workloadData.PodGroups)
	workloadcalc.SumPodGpuMemoryResources(workloadMetadata, workloadData.Pods)

	return nil
}
