/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/integration-tests/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func BuildNamespaceObject(nsName string, projectName *string) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   nsName,
			Labels: map[string]string{},
		},
	}

	if projectName != nil {
		ns.Labels[constants.RunaiQueueLabel] = *projectName
	}

	return ns
}

// BuildProjectObject builds the project the suites run against.
func BuildProjectObject(projectName, departmentName, projectId, departmentId string) *kaiv1alpha1.Project {
	return common.BuildProjectObject(projectName, departmentName, projectId, departmentId)
}

func buildDepartmentObject(departmentName, departmentId string) *kaiv1alpha1.Department {
	return common.BuildDepartmentObject(departmentName, departmentId)
}

// buildDefaultNodePoolObject builds the default node pool the suites bootstrap against.
// UseKaiNodePools (temporary dev-period flag) selects which NodePool CRD it is created on —
// only one of the two exists once the migration completes, so the suites create the same
// CRD the services under test read, rather than seeding both.
func buildDefaultNodePoolObject() client.Object {
	if !clusterconstants.UseKaiNodePools {
		return buildDefaultLegacyNodePoolObject()
	}

	// overProvisioningRatio and placementStrategy were dropped from the kai.resources CRD;
	// placement strategy now lives under schedulingShardConfig and is left to its defaults.
	return &kaiv1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: v1alpha1.DefaultNodePoolName,
		},
		Spec: kaiv1alpha1.NodePoolSpec{
			LabelKey:   "",
			LabelValue: "",
		},
	}
}

// TODO: delete once clusterconstants.UseKaiNodePools=true
func buildDefaultLegacyNodePoolObject() client.Object {
	return &v1alpha1.NodePool{
		ObjectMeta: metav1.ObjectMeta{
			Name: v1alpha1.DefaultNodePoolName,
		},
		Spec: v1alpha1.NodePoolSpec{
			LabelKey:   "",
			LabelValue: "",
			PlacementStrategy: v1alpha1.PlacementStrategy{
				Cpu:       "",
				Gpu:       "",
				GpuDevice: "",
			},
		},
	}
}

// newEmptyNodePoolObject returns an empty node pool of whichever CRD the suites bootstrap
// against, for Get calls that only need the object's identity.
func newEmptyNodePoolObject() client.Object {
	if !clusterconstants.UseKaiNodePools {
		return &v1alpha1.NodePool{}
	}
	return &kaiv1alpha1.NodePool{}
}
