/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_crud

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	migrationsInfoSelect = "SELECT * FROM crud_migrations WHERE crud_migrations.tenant_id = ? AND crud_migrations.tenant_id = ?"
	migrationsInfoUpdate = "UPDATE crud_migrations SET assets_to_flat_templates=? WHERE tenant_id = ?"
)

type migrationsInfoRow struct {
	TenantId              int64
	AssetsToFlatTemplates int32
}

var _ = Describe("GetMigrationsInfo", func() {
	It("returns the stored migrations info", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(migrationsInfoSelect).Returns([]migrationsInfoRow{{TenantId: 7, AssetsToFlatTemplates: 5}})

		info, apiErr := GetMigrationsInfo(testCtx(7))

		Expect(apiErr).To(BeNil())
		Expect(info).ToNot(BeNil())
		Expect(info.TenantId).To(Equal(int64(7)))
		Expect(info.AssetsToFlatTemplates).To(Equal(int32(5)))
	})

	It("returns a zeroed record when there are no rows", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(migrationsInfoSelect).Returns([]migrationsInfoRow{})

		info, apiErr := GetMigrationsInfo(testCtx(7))

		Expect(apiErr).To(BeNil())
		Expect(info).ToNot(BeNil())
		Expect(info.TenantId).To(Equal(int64(7)))
		Expect(info.AssetsToFlatTemplates).To(Equal(int32(0)))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(migrationsInfoSelect).ReturnsError(errors.New("boom"))

		info, apiErr := GetMigrationsInfo(testCtx(7))

		Expect(info).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("SetAssetsToFlatMigration", func() {
	It("persists the new migration value", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(migrationsInfoSelect).Returns([]migrationsInfoRow{{TenantId: 7, AssetsToFlatTemplates: 1}})
		m.ExpectTransaction().Begin()
		m.ExpectExec(migrationsInfoUpdate, int64(9), int64(7)).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := SetAssetsToFlatMigration(testCtx(7), 9)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when loading the current info fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(migrationsInfoSelect).ReturnsError(errors.New("boom"))

		apiErr := SetAssetsToFlatMigration(testCtx(7), 9)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when saving fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(migrationsInfoSelect).Returns([]migrationsInfoRow{{TenantId: 7, AssetsToFlatTemplates: 1}})
		m.ExpectTransaction().Begin()
		m.ExpectExec(migrationsInfoUpdate).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := SetAssetsToFlatMigration(testCtx(7), 9)

		Expect(apiErr).ToNot(BeNil())
	})
})
