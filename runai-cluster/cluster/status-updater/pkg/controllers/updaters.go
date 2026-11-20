/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/record"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	clusterrelations "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/cluster_relations"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/metadata"
	podmetacalc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/metadata_calculator/pod"
	workloadcalc "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/metadata_calculator/workload"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/patcher"
)

// gpuCalcError marks a failure to calculate the GPU resources of a pod. The pod status annotation is still
// written when it happens, so the caller can keep updating the workload instead of stopping at this error.
type gpuCalcError struct {
	err error
}

// Error returns the message of the wrapped GPU calculation error.
func (e *gpuCalcError) Error() string {
	return e.err.Error()
}

// Unwrap returns the wrapped GPU calculation error.
func (e *gpuCalcError) Unwrap() error {
	return e.err
}

// updatePod writes the calculated status and GPU allocation annotations of the pod. When the GPU calculation
// fails, it still writes the status annotation, leaves the GPU annotations untouched and returns a *gpuCalcError.
func updatePod(ctx context.Context, pod *v1.Pod, kubeClient client.Client) error {
	logger := log.FromContext(ctx)

	podMeta := metadata.PodMetadata{}
	podMeta.Status = podmetacalc.TranslatePodStatus(pod)

	err := podmetacalc.CalcPodGpuResources(ctx, pod, &podMeta, kubeClient)
	if err != nil {
		return updatePodStatusOnGpuCalcError(ctx, pod, podMeta.Status, err, kubeClient)
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

// updatePodStatusOnGpuCalcError writes only the status annotation of a pod whose GPU resources could not be
// calculated, and returns the calculation error wrapped in a *gpuCalcError. When the status patch also fails,
// it returns the joined errors without the *gpuCalcError marker, so the caller stops as on any other update error.
func updatePodStatusOnGpuCalcError(ctx context.Context, pod *v1.Pod, status string, calcErr error,
	kubeClient client.Client) error {
	gpuErr := &gpuCalcError{err: fmt.Errorf("failed to calculate gpu resources for pod: <%v/%v> with error : %w",
		pod.Namespace, pod.Name, calcErr)}
	if pod.Annotations[patcher.PodStatusAnnotationName] == status {
		return gpuErr
	}

	if err := patcher.UpdatePodStatusAnnotation(ctx, pod, status, kubeClient); err != nil {
		return errors.Join(gpuErr.err, err)
	}

	log.FromContext(ctx).V(1).Info(fmt.Sprintf("Updated pod status: <%v/%v> status: %s", pod.Namespace,
		pod.Name, status))
	return gpuErr
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
