/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package version_control

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/authorization-migrator/config"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm/mocks"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestVersionControl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Version Control Suite")
}

var _ = Describe("Migration version control", func() {
	Describe("Init", func() {
		var (
			rawDB   *sql.DB
			sqlMock sqlmock.Sqlmock
			gormDB  *gorm.DB
		)

		BeforeEach(func() {
			var err error
			rawDB, sqlMock, err = sqlmock.New()
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				_ = rawDB.Close()
			})

			gormDB, err = gorm.Open(
				postgres.New(postgres.Config{Conn: rawDB}),
				&gorm.Config{DisableAutomaticPing: true},
			)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(sqlMock.ExpectationsWereMet()).To(Succeed())
		})

		Context("when the migration control table does not exist", func() {
			It("creates the table and stores the configured version", func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(
					"SELECT count(*) FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name = $1 AND table_type = $2",
				)).
					WithArgs("migration_controls", "BASE TABLE").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
				sqlMock.ExpectExec(regexp.QuoteMeta(
					`CREATE TABLE "migration_controls" ("version" text,PRIMARY KEY ("version"))`,
				)).
					WillReturnResult(sqlmock.NewResult(0, 0))

				db := mocks.NewDB(GinkgoT())
				db.On("Actual").Return(gormDB)

				Init(&config.ConfigData{MigrationVersion: "v2"}, db)

				Expect(mc.Version).To(Equal("v2"))
			})
		})

		Context("when auto-migration fails", func() {
			It("still stores the configured version", func() {
				sqlMock.ExpectQuery(regexp.QuoteMeta(
					"SELECT count(*) FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name = $1 AND table_type = $2",
				)).
					WithArgs("migration_controls", "BASE TABLE").
					WillReturnError(errors.New("migration failed"))

				db := mocks.NewDB(GinkgoT())
				db.On("Actual").Return(gormDB)

				Init(&config.ConfigData{MigrationVersion: "v3"}, db)

				Expect(mc.Version).To(Equal("v3"))
			})
		})
	})

	Describe("IsMigrationDone", func() {
		BeforeEach(func() {
			mc = &MigrationControl{Version: "v1"}
		})

		Context("when the migration record exists", func() {
			It("returns true", func() {
				db := mocks.NewDB(GinkgoT())
				db.On("First", mc).Return(nil)

				Expect(IsMigrationDone(db)).To(BeTrue())
			})
		})

		Context("when reading the migration record fails", func() {
			It("returns false", func() {
				db := mocks.NewDB(GinkgoT())
				db.On("First", mc).Return(errors.New("not found"))

				Expect(IsMigrationDone(db)).To(BeFalse())
			})
		})
	})

	Describe("MarkMigrationDone", func() {
		BeforeEach(func() {
			mc = &MigrationControl{Version: "v1"}
		})

		Context("when the record can be created", func() {
			It("stores the migration version", func() {
				db := mocks.NewDB(GinkgoT())
				db.On("Create", mc).Return(int64(1), nil)

				MarkMigrationDone(db)
			})
		})

		Context("when creating the record fails", func() {
			It("does not panic", func() {
				db := mocks.NewDB(GinkgoT())
				db.On("Create", mock.Anything).Return(int64(0), errors.New("create failed"))

				Expect(func() {
					MarkMigrationDone(db)
				}).NotTo(Panic())
			})
		})
	})
})
