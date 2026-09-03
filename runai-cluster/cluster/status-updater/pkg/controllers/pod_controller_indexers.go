/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package controllers

import (
	clusterrelations "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/status-updater/pkg/controllers/cluster_relations"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func podGroupNameIndexerFunc(rawObj client.Object) []string {
	pod, conversionSuccess := rawObj.(*v1.Pod)
	if !conversionSuccess {
		return nil
	}

	podGroupName, foundPGAnnotation := pod.Annotations[clusterrelations.PodGroupAnnotationForPodName]
	if !foundPGAnnotation {
		return nil
	}
	return []string{podGroupName}
}

func topOwnerUidIndexerFunc(rawObj client.Object) []string {
	podGroup, conversionSuccess := rawObj.(*kaiv2alpha2.PodGroup)
	if !conversionSuccess {
		return nil
	}

	topOwnerRef, err := clusterrelations.GetPodGroupTopOwnerRef(podGroup)
	if err != nil {
		return nil
	}
	return []string{topOwnerRef.UID}
}
