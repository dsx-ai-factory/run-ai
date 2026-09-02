/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/resourcing"
	schedv2alpha2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	"k8s.io/utils/ptr"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateDepartmentExistsEventually(departmentName string, k8sClient client.Client) {
	var dep *schedv2alpha2.Department = nil
	EventuallyWithOffset(1, func() *schedv2alpha2.Department {
		dep = getLegacyDepartmentFromClient(departmentName, k8sClient)
		return dep
	}, DefaultTimeout).ShouldNot(BeNil())
}

func validateDepartmentDeletedEventually(departmentName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func() *schedv2alpha2.Department {
		return getLegacyDepartmentFromClient(departmentName, k8sClient)
	}, DefaultTimeout).Should(BeNil())
}

func validateExistingDepartmentMatchesBackendSpec(departmentName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		validateExistingDepartmentWithBackendSpec(g, departmentName, k8sClient)
	}, 0).Should(Succeed())
}

func validateExistingDepartmentDoesNotMatchBackendSpec(departmentName string, k8sClient client.Client) {
	EventuallyWithOffset(1, func(g Gomega) {
		validateExistingDepartmentWithBackendSpec(g, departmentName, k8sClient)
	}, 0).ShouldNot(Succeed())
}

func validateExistingDepartmentWithBackendSpec(g Gomega, departmentName string, k8sClient client.Client) {
	clusterDepartment := getLegacyDepartmentFromClient(departmentName, k8sClient)
	g.ExpectWithOffset(1, clusterDepartment).ToNot(BeNil())

	backendDepSpec := OrgUnitDepartmentsMap[departmentName]
	g.ExpectWithOffset(1, backendDepSpec).ToNot(BeNil(), "Backend department spec should exist")

	g.ExpectWithOffset(1, clusterDepartment.Name).To(Equal(common.GetStringVal(backendDepSpec.Name)), "department name")
	g.ExpectWithOffset(1, clusterDepartment.Labels[schedv2alpha2.DepartmentIdLabel]).To(Equal(backendDepSpec.Id), "department id")

	g.ExpectWithOffset(1, len(clusterDepartment.Spec.Queues)).To(Equal(len(backendDepSpec.Resources)), "Queues length")
	for _, queue := range clusterDepartment.Spec.Queues {
		validateQueue(g, queue.Name, queue.Nodepool, queue.Priority, ptr.Deref(queue.Resources, resourcing.QueueResources{}), backendDepSpec.Resources, *backendDepSpec.Name)
	}
}
