/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package handlers

import (
	"time"

	_ "github.com/run-ai/runai/runai-cluster/cluster/agent/pkg/tests/helper"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	orgUnitApi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	v1alpha1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	schedv2alpha2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
	"k8s.io/utils/ptr"
)

func orgUnitDepartment(id, name string, resources ...orgUnitApi.ResourcesSync) orgUnitApi.DepartmentSync {
	return orgUnitApi.DepartmentSync{
		Id:        id,
		Name:      ptr.To(name),
		Resources: resources,
	}
}

func deletedOrgUnitDepartment(id, name string) orgUnitApi.DepartmentSync {
	department := orgUnitDepartment(id, name)
	department.DeletedAt = *orgUnitApi.NewNullableTime(ptr.To(time.Now()))
	return department
}

var _ = Describe("DepartmentsHandler conversion", func() {
	Describe("convertOrgUnitToKaiDepartment", func() {
		It("uses id-based q-<id> queue names and sets the department-id label", func() {
			department, err := convertOrgUnitToKaiDepartment(orgUnitDepartment("d-1", "dep-1",
				resourcesForNodePool(v1alpha1.DefaultNodePoolName, 4),
				resourcesForNodePool("gpu-pool", 4),
			))

			Expect(err).ToNot(HaveOccurred())
			Expect(department.Name).To(Equal("dep-1"))
			Expect(department.Labels).To(HaveKeyWithValue(schedv2alpha2.DepartmentIdLabel, "d-1"))
			Expect(department.Spec.Queues).To(HaveLen(2))
			Expect(department.Spec.Queues[0].Name).To(Equal("q-d-1"))
			Expect(department.Spec.Queues[1].Name).To(Equal("q-d-1-gpu-pool"))
		})
	})

	Describe("getClusterObjects", func() {
		It("adds a KAI department for add/update", func() {
			handler := &DepartmentsHandler{CommonHandler: &CommonHandler{}}

			toAddUpdate, toDelete, _, err := handler.getClusterObjects(
				[]orgUnitApi.DepartmentSync{orgUnitDepartment("d-1", "dep-1", resourcesForNodePool(v1alpha1.DefaultNodePoolName, 1))})

			Expect(err).ToNot(HaveOccurred())
			Expect(toDelete).To(BeEmpty())
			Expect(toAddUpdate).To(HaveLen(1))
			Expect(toAddUpdate[0]).To(BeAssignableToTypeOf(&kaiorg.Department{}))
			Expect(toAddUpdate[0].GetName()).To(Equal("dep-1"))
		})

		It("routes deleted departments to the delete map", func() {
			handler := &DepartmentsHandler{CommonHandler: &CommonHandler{}}

			toAddUpdate, toDelete, _, err := handler.getClusterObjects(
				[]orgUnitApi.DepartmentSync{deletedOrgUnitDepartment("d-1", "dep-1")})

			Expect(err).ToNot(HaveOccurred())
			Expect(toAddUpdate).To(BeEmpty())
			Expect(toDelete).To(HaveKey("d-1"))
		})
	})

	Describe("getClusterObjects delegation", func() {
		It("produces KAI Departments when the toggle is on", func() {
			handler := &DepartmentsHandler{CommonHandler: &CommonHandler{}}

			toAddUpdate, _, _, err := handler.getClusterObjects(
				[]orgUnitApi.DepartmentSync{orgUnitDepartment("d-1", "dep-1", resourcesForNodePool(v1alpha1.DefaultNodePoolName, 1))})

			Expect(err).ToNot(HaveOccurred())
			Expect(toAddUpdate).To(HaveLen(1))
			Expect(toAddUpdate[0]).To(BeAssignableToTypeOf(&kaiorg.Department{}))
		})
	})
})
