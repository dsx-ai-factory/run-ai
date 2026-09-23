/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"context"
	"fmt"
	"maps"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const runaiScheduler = "runai-scheduler"

type DeploymentSpec struct {
	Name      string
	Namespace string
	Image     string
	Replicas  int32
	Labels    map[string]string
}

func K8sDeployment(ctx context.Context, kube k8sclient.Client, spec DeploymentSpec) (*appsv1.Deployment, error) {
	deployment := deploymentObject(spec)
	if err := kube.Create(ctx, deployment); err != nil {
		return nil, fmt.Errorf("creating deployment %s/%s: %w", spec.Namespace, spec.Name, err)
	}
	return deployment, nil
}

func deploymentObject(spec DeploymentSpec) *appsv1.Deployment {
	podLabels := maps.Clone(spec.Labels)
	if podLabels == nil {
		podLabels = map[string]string{}
	}
	podLabels["app"] = spec.Name
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: spec.Name, Namespace: spec.Namespace, Labels: maps.Clone(spec.Labels)},
		Spec: appsv1.DeploymentSpec{
			Replicas: &spec.Replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": spec.Name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: podLabels},
				Spec: corev1.PodSpec{
					SchedulerName: runaiScheduler,
					Containers:    []corev1.Container{{Name: "main", Image: spec.Image}},
				},
			},
		},
	}
}
