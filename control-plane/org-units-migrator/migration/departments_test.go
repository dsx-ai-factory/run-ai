/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/run-ai/runai/runai-common-packages/dao-utils/validators"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	orgUnitsServiceDao "github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/dao"
	models2 "github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/org-units-migrator/database"
	"github.com/run-ai/runai/control-plane/org-units-migrator/database/models"
)

var _ = Describe("Departments tests", Ordered, func() {
	It("Should get departments grouped", func() {
		tenants := database.GetTenantsToMigrate([]string{}, []string{})
		for _, tenant := range tenants {
			departments, err := database.GetTenantDepartmentsV1(tenant.Id)
			Expect(err).To(BeNil())
			Expect(len(departments)).To(Equal(database.TestDepartmentsCountPerTenant))
		}
	})
})

func validateDepartmentsMigration(tenant models.Tenant, setting map[string]bool) {
	departments, err := orgUnitsServiceDao.GetDepartments(context.Background(), fmt.Sprintf("%d", tenant.Id), []validators.ParameterizedFilter{}, "", "", 0, 0, nil, false, models2.SlimVerbosity)
	Expect(err).To(BeNil())
	Expect(departments).To(Not(BeNil()))
	Expect(len(*departments)).To(Equal(database.TestDepartmentsCountPerTenant))
	limit := float64(0.5)
	deserved := float64(0.5)
	oqw := deserved
	for _, department := range *departments {
		Expect(department.ID).To(BeNumerically(">=", 1*tenant.Id))
		Expect(department.ID).To(BeNumerically("<=", database.TestDepartmentsCountPerTenant*tenant.Id))
		Expect(department.Name).To(Equal(fmt.Sprintf("dep-%d", department.ID)))
		Expect(department.TenantId).To(Equal(fmt.Sprintf("%d", tenant.Id)))
		Expect(department.ClusterId.String()).To(Equal(database.MapTenantIdToClusterId[tenant.Id]))
		Expect(department.CreatedAt).To(Not(Equal(time.Time{})))
		Expect(department.DeletedAt.Time).To(Equal(time.Time{}))
		Expect(len(department.OrgUnitData.Resources)).To(Equal(database.TestNodepoolsCountPerTenant))
		for _, resource := range department.OrgUnitData.Resources {
			Expect(resource.CpuDeserved).To(Not(BeNil()))
			Expect(*resource.CpuDeserved).To(Equal(deserved))
			Expect(resource.CpuLimit).To(Not(BeNil()))
			Expect(*resource.CpuLimit).To(Equal(limit))
			Expect(resource.CpuOverQuotaWeight).To(Not(BeNil()))
			Expect(*resource.CpuOverQuotaWeight).To(Equal(oqw))
			Expect(resource.GpuDeserved).To(Not(BeNil()))
			Expect(*resource.GpuDeserved).To(Equal(deserved))
			Expect(resource.GpuLimit).To(Not(BeNil()))
			Expect(*resource.GpuLimit).To(Equal(limit))
			Expect(resource.GpuOverQuotaWeight).To(Not(BeNil()))
			Expect(*resource.GpuOverQuotaWeight).To(Equal(oqw))
			Expect(resource.MemoryDeserved).To(Not(BeNil()))
			Expect(*resource.MemoryDeserved).To(Equal(deserved))
			Expect(resource.MemoryLimit).To(Not(BeNil()))
			Expect(*resource.MemoryLimit).To(Equal(limit))
			Expect(resource.MemoryOverQuotaWeight).To(Not(BeNil()))
			Expect(*resource.MemoryOverQuotaWeight).To(Equal(oqw))

			Expect(resource.NodePoolName).To(Equal(fmt.Sprintf("np-%s", resource.NodePoolId)))
		}
	}
}
