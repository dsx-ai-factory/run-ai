/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	resource "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	jobsetv1alpha2 "sigs.k8s.io/jobset/api/jobset/v1alpha2"
)

const nvidiaGPUResourceName = "nvidia.com/gpu"

type JobSetBuilder struct {
	name      string
	namespace string
	replicas  int32
	withGPU   bool
	gpuCount  int64
}

func NewJobSetBuilder(name, namespace string) *JobSetBuilder {
	return &JobSetBuilder{
		name:      name,
		namespace: namespace,
		replicas:  1,
		withGPU:   false,
		gpuCount:  1,
	}
}

func (b *JobSetBuilder) WithDefault() *JobSetBuilder {
	b.replicas = 2
	b.withGPU = true
	b.gpuCount = 1
	return b
}

func (b *JobSetBuilder) WithReplicas(replicas int32) *JobSetBuilder {
	b.replicas = replicas
	return b
}

func (b *JobSetBuilder) WithGPU(count int64) *JobSetBuilder {
	b.withGPU = true
	b.gpuCount = count
	return b
}

func (b *JobSetBuilder) Build() *jobsetv1alpha2.JobSet {
	jobSet := &jobsetv1alpha2.JobSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "jobset.x-k8s.io/v1alpha2",
			Kind:       "JobSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      b.name,
			Namespace: b.namespace,
		},
		Spec: jobsetv1alpha2.JobSetSpec{
			ReplicatedJobs: []jobsetv1alpha2.ReplicatedJob{
				{
					Name:     "worker",
					Replicas: b.replicas,
					Template: batchv1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							Template: corev1.PodTemplateSpec{
								Spec: corev1.PodSpec{
									Containers: []corev1.Container{
										{
											Name:  "main",
											Image: "test:latest",
										},
									},
									RestartPolicy: corev1.RestartPolicyNever,
								},
							},
						},
					},
				},
			},
		},
	}

	if b.withGPU {
		jobSet.Spec.ReplicatedJobs[0].Template.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{
			Limits: corev1.ResourceList{
				corev1.ResourceName(nvidiaGPUResourceName): resource.MustParse(fmt.Sprintf("%d", b.gpuCount)),
			},
		}
	}

	return jobSet
}
