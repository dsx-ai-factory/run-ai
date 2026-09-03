/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package version_control

import (
	"errors"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type migrationControlRow struct {
	Version string `gorm:"column:version"`
}

var _ = Describe("version_control SQL", func() {
	BeforeEach(func() {
		mc = &MigrationControl{Version: "v1"}
	})

	Describe("IsMigrationDone SQL", func() {
		It("returns true when the version row exists", func() {
			mocker := dbtesting.SetSQLMock(GinkgoT())

			mocker.ExpectQuery(
				`SELECT * FROM migration_controls WHERE migration_controls.version = ? ORDER BY migration_controls.version LIMIT ?`,
				"v1", int64(1),
			).Returns([]migrationControlRow{{Version: "v1"}})

			Expect(IsMigrationDone(runaidb.DB())).To(BeTrue())
		})

		It("returns false when the version row is missing", func() {
			mocker := dbtesting.SetSQLMock(GinkgoT())

			mocker.ExpectQuery(
				`SELECT * FROM migration_controls WHERE migration_controls.version = ? ORDER BY migration_controls.version LIMIT ?`,
				"v1", int64(1),
			).Returns([]migrationControlRow{})

			Expect(IsMigrationDone(runaidb.DB())).To(BeFalse())
		})

		It("returns false on DB error", func() {
			mocker := dbtesting.SetSQLMock(GinkgoT())

			mocker.ExpectQuery(
				`SELECT * FROM migration_controls WHERE migration_controls.version = ? ORDER BY migration_controls.version LIMIT ?`,
				"v1", int64(1),
			).ReturnsError(errors.New("db error"))

			Expect(IsMigrationDone(runaidb.DB())).To(BeFalse())
		})
	})

	Describe("MarkMigrationDone SQL", func() {
		It("inserts the version row", func() {
			mocker := dbtesting.SetSQLMock(GinkgoT())

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`INSERT INTO migration_controls (version) VALUES (?)`,
				"v1",
			).Returns(1)
			mocker.ExpectTransaction().Commit()

			Expect(func() {
				MarkMigrationDone(runaidb.DB())
			}).NotTo(Panic())
		})

		It("logs and does not panic on insert failure", func() {
			mocker := dbtesting.SetSQLMock(GinkgoT())

			mocker.ExpectTransaction().Begin()
			mocker.ExpectExec(
				`INSERT INTO migration_controls (version) VALUES (?)`,
				"v1",
			).ReturnsError(errors.New("insert failed"))
			mocker.ExpectTransaction().Rollback()

			Expect(func() {
				MarkMigrationDone(runaidb.DB())
			}).NotTo(Panic())
		})
	})
})
