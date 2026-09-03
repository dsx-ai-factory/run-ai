/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package metadata

import (
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	usedNodesDefault = "-"
)

type TopOwnerReference struct {
	Name    string `yaml:"name"`
	UID     string `yaml:"uid"`
	Group   string `yaml:"group"`
	Version string `yaml:"version"`
	Kind    string `yaml:"kind"`
}

type WorkloadObjects struct {
	GroupVersionResource *schema.GroupVersionResource
	Namespace            string
	WorkloadObj          *unstructured.Unstructured
	PodGroups            []kaiv2alpha2.PodGroup
	Pods                 []v1.Pod
	LatestChiefPod       *v1.Pod
}

type WorkloadMetadata struct {
	Status string

	RunningPods int64
	PendingPods int64
	UsedNodes   string

	RequestedGPUs              float64
	RequestedGPUsMemory        int64
	CurrentRequestedGPUs       float64
	CurrentRequestedGPUsMemory int64
	AllocatedGpus              float64
	AllocatedGPUsMemory        int64
}

func CreateWorkloadMetadata() *WorkloadMetadata {
	return &WorkloadMetadata{
		UsedNodes: usedNodesDefault,
	}
}
