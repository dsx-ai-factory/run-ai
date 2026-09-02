/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	orgUnitsServiceDao "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-units-migrator/database/models"
)

var _ = Describe("Node types tests", Ordered, func() {
	It("Should get all node types", func() {
		tenants := database.GetTenantsToMigrate([]string{}, []string{})
		for _, tenant := range tenants {
			nodeTypes, err := database.GetTenantNodeTypes(tenant.Id)
			Expect(err).To(BeNil())
			Expect(len(nodeTypes)).To(Equal(database.TestNodeTypesCountPerTenant))
		}
	})
})

func validateNodeTypesMigration(tenant models.Tenant) {
	nodeTypes, err := orgUnitsServiceDao.GetNodeTypes(context.Background(), fmt.Sprintf("%d", tenant.Id), uuid.MustParse(database.MapTenantIdToClusterId[tenant.Id]), nil)
	Expect(err).To(BeNil())
	Expect(nodeTypes).To(Not(BeNil()))
	Expect(len(*nodeTypes)).To(Equal(database.TestNodeTypesCountPerTenant))

	for _, nodeType := range *nodeTypes {
		Expect(nodeType.ID).To(BeNumerically(">=", database.TestNodeTypesCountPerTenant*tenant.Id))
		Expect(nodeType.ID).To(BeNumerically("<=", database.TestNodeTypesCountPerTenant*(tenant.Id+1)))
		Expect(nodeType.Name).To(Equal(fmt.Sprintf("node-type-%d", nodeType.ID)))
		Expect(nodeType.TenantId).To(Equal(fmt.Sprintf("%d", tenant.Id)))
		Expect(nodeType.ClusterId.String()).To(Equal(database.MapTenantIdToClusterId[tenant.Id]))
		Expect(nodeType.CreatedAt).To(Not(Equal(time.Time{})))
		Expect(nodeType.UpdatedAt).To(Not(Equal(time.Time{})))
		Expect(nodeType.DeletedAt.Time).To(Equal(time.Time{}))
	}
}
