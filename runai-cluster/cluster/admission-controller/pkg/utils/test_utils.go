/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/admission-controller/pkg/types"
)

func CreatePodController(namespaceName string) *types.PodControllerCommon {
	projectName := "very-special-project"

	podController := &types.PodControllerCommon{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespaceName,
			Labels:    map[string]string{},
		},
		Spec: types.PodControllerSpec{
			Template: types.PodControllerTemplateSpec{
				Spec: corev1.PodSpec{
					SchedulerName: types.RunaiScheduler,
					Containers: []corev1.Container{
						{
							Resources: corev1.ResourceRequirements{},
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"project": projectName},
				},
			},
		},
	}
	podController.Kind = "Deployment"
	return podController
}
