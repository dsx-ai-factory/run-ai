/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pod_mutators

import (
	"k8s.io/api/admission/v1beta1"
	v1 "k8s.io/api/core/v1"

	pod_identifier "github.com/run-ai/runai/runai-cluster/cluster/admission-controller/pkg/pod-identifier"
)

const (
	gkeClusterName  = "gke"
	azureMlLabelKey = "ml.azure.com/compute"
)

func getDestinationNamespace(pod *v1.Pod, request *v1beta1.AdmissionRequest) string {
	if pod.Namespace != "" {
		return pod.Namespace
	}
	return request.Namespace
}

func isRunaiPodJob(pod *v1.Pod) bool {
	return pod_identifier.IsRunaiPod(pod) && len(pod.OwnerReferences) == 0
}
