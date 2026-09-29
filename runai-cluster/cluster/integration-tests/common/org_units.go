/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v2"
	schedulingv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Org-unit fixtures for the suites. The services under test reconcile and read the
// kai.resources Project and Department, so those are the only org units the suites create.

// BuildProjectObject builds a project with a single queue on the default node pool.
// The project and department id labels keep the run.ai keys: they are the ones the
// consumers (e.g. the external-workload-integrator) read off the kai project.
func BuildProjectObject(projectName, departmentName, projectId, departmentId string) *kaiv1alpha1.Project {
	return &kaiv1alpha1.Project{
		ObjectMeta: metav1.ObjectMeta{
			Name: projectName,
			Labels: map[string]string{
				v2.ProjectIdLabel:                    projectId,
				schedulingv2alpha2.DepartmentIdLabel: departmentId,
			},
		},
		Spec: kaiv1alpha1.ProjectSpec{
			Parent: departmentName,
			Queues: []kaiv1alpha1.QueueConfig{
				{
					Name:      projectName,
					Nodepool:  v1alpha1.DefaultNodePoolName,
					Resources: BuildQueueResources(1, 1000, 10000),
				},
			},
		},
	}
}

// BuildDepartmentObject builds a department with a single queue on the default node pool.
func BuildDepartmentObject(departmentName, departmentId string) *kaiv1alpha1.Department {
	return &kaiv1alpha1.Department{
		ObjectMeta: metav1.ObjectMeta{
			Name: departmentName,
			Labels: map[string]string{
				schedulingv2alpha2.DepartmentIdLabel: departmentId,
			},
		},
		Spec: kaiv1alpha1.DepartmentSpec{
			Queues: []kaiv1alpha1.QueueConfig{
				{
					Name:      departmentName,
					Nodepool:  v1alpha1.DefaultNodePoolName,
					Resources: BuildQueueResources(1, 1000, 10000),
				},
			},
		},
	}
}

// BuildQueueResources builds the resources of a project or department queue.
func BuildQueueResources(gpuDeserved, cpuDeserved, memoryDeserved int64) *kaiv1alpha1.QueueResourcesConfig {
	return &kaiv1alpha1.QueueResourcesConfig{
		GPU: kaiv1alpha1.SystemResource{
			Deserved: float64(gpuDeserved),
			Limit:    -1,
		},
		CPU: kaiv1alpha1.SystemResource{
			Deserved: float64(cpuDeserved),
			Limit:    -1,
		},
		Memory: kaiv1alpha1.SystemResource{
			Deserved: float64(memoryDeserved),
			Limit:    -1,
		},
	}
}
