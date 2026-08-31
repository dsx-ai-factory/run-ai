/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm/mocks"
	"github.com/stretchr/testify/mock"
)

func TestMigration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Migration Suite")
}

var _ = Describe("Migration", func() {
	Describe("getTenantsManagerTenants", func() {
		var (
			db      *sql.DB
			sqlMock sqlmock.Sqlmock
		)

		BeforeEach(func() {
			var err error
			db, sqlMock, err = sqlmock.New()
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				_ = db.Close()
			})
		})

		AfterEach(func() {
			Expect(sqlMock.ExpectationsWereMet()).To(Succeed())
		})

		Context("when the query succeeds", func() {
			It("returns tenants indexed by ID", func() {
				sqlMock.ExpectQuery("SELECT id AS tenant_id, name AS tenant_name FROM tenants").
					WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "tenant_name"}).
						AddRow(1, "tenant-one").
						AddRow(2, "tenant-two"))

				Expect(getTenantsManagerTenants(db)).To(Equal(map[int]string{
					1: "tenant-one",
					2: "tenant-two",
				}))
			})
		})

		Context("when the query fails", func() {
			It("returns an empty map", func() {
				sqlMock.ExpectQuery("SELECT id AS tenant_id, name AS tenant_name FROM tenants").
					WillReturnError(errors.New("query failed"))

				Expect(getTenantsManagerTenants(db)).To(BeEmpty())
			})
		})

		Context("when a row cannot be scanned", func() {
			It("skips the invalid row", func() {
				sqlMock.ExpectQuery("SELECT id AS tenant_id, name AS tenant_name FROM tenants").
					WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "tenant_name"}).
						AddRow("not-an-int", "invalid").
						AddRow(2, "tenant-two"))

				Expect(getTenantsManagerTenants(db)).To(Equal(map[int]string{2: "tenant-two"}))
			})
		})

		Context("when row iteration fails", func() {
			It("returns an empty map", func() {
				sqlMock.ExpectQuery("SELECT id AS tenant_id, name AS tenant_name FROM tenants").
					WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "tenant_name"}).
						AddRow(1, "tenant-one").
						RowError(0, errors.New("row failed")))

				Expect(getTenantsManagerTenants(db)).To(BeEmpty())
			})
		})
	})

	Describe("Migrate", func() {
		var (
			sourceDB    *sql.DB
			sourceMock  sqlmock.Sqlmock
			tenantsDB   *sql.DB
			tenantsMock sqlmock.Sqlmock
		)

		BeforeEach(func() {
			var err error
			sourceDB, sourceMock, err = sqlmock.New()
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				_ = sourceDB.Close()
			})

			tenantsDB, tenantsMock, err = sqlmock.New()
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func() {
				_ = tenantsDB.Close()
			})
		})

		AfterEach(func() {
			Expect(sourceMock.ExpectationsWereMet()).To(Succeed())
			Expect(tenantsMock.ExpectationsWereMet()).To(Succeed())
		})

		Context("when the legacy tables do not exist", func() {
			It("returns without migrating", func() {
				tenantsMock.ExpectQuery("SELECT id AS tenant_id, name AS tenant_name FROM tenants").
					WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "tenant_name"}))
				sourceMock.ExpectQuery("SELECT DISTINCT au.user_id AS user_id,").
					WillReturnError(&pgconn.PgError{Code: "42P01"})

				Expect(func() {
					Migrate(sourceDB, tenantsDB, mocks.NewDB(GinkgoT()))
				}).NotTo(Panic())
			})

			It("returns without migrating when 42P01 is wrapped (e.g. PrepareError)", func() {
				tenantsMock.ExpectQuery("SELECT id AS tenant_id, name AS tenant_name FROM tenants").
					WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "tenant_name"}))
				sourceMock.ExpectQuery("SELECT DISTINCT au.user_id AS user_id,").
					WillReturnError(fmt.Errorf("prepare failed: %w", &pgconn.PgError{Code: "42P01"}))

				Expect(func() {
					Migrate(sourceDB, tenantsDB, mocks.NewDB(GinkgoT()))
				}).NotTo(Panic())
			})
		})

		Context("when legacy role assignments exist", func() {
			It("creates access rules for valid assignments", func() {
				tenantsMock.ExpectQuery("SELECT id AS tenant_id, name AS tenant_name FROM tenants").
					WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "tenant_name"}).
						AddRow(1, "tenant-one"))

				columns := []string{
					"user_id", "client_id", "tenant_id", "entity_type", "created_by",
					"created_at", "updated_at", "all_clusters_permitted", "role",
					"cluster_uuid", "department_id", "project_id", "cluster_name",
					"department_name", "department_cluster_uuid", "project_name",
					"project_cluster_uuid", "tenant_name", "auth_realm",
				}
				rows := sqlmock.NewRows(columns).
					AddRow("admin", nil, 1, EntityTypeGroup, nil, nil, nil, true, "admin", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil).
					AddRow("department-admin", nil, 1, EntityTypeGroup, nil, nil, nil, false, "department_admin", "cluster-1", 10, nil, "cluster", "department", "cluster-1", nil, nil, nil, nil).
					AddRow("editor", nil, 1, EntityTypeGroup, nil, nil, nil, false, "editor", "cluster-1", nil, nil, "cluster", nil, nil, nil, nil, nil, nil).
					AddRow("ml-engineer", nil, 1, EntityTypeGroup, nil, nil, nil, false, "ml_engineer", "cluster-1", nil, nil, "cluster", nil, nil, nil, nil, nil, nil).
					AddRow("research-manager", nil, 1, EntityTypeGroup, nil, nil, nil, false, "research_manager", "cluster-1", nil, nil, "cluster", nil, nil, nil, nil, nil, nil).
					AddRow("researcher", nil, 1, EntityTypeGroup, nil, nil, nil, false, "researcher", "cluster-1", nil, 20, "cluster", nil, nil, "project", "cluster-1", nil, nil).
					AddRow("viewer", nil, 1, EntityTypeGroup, nil, nil, nil, false, "viewer", "cluster-1", nil, nil, "cluster", nil, nil, nil, nil, nil, nil).
					AddRow("no-role", nil, 1, EntityTypeGroup, nil, nil, nil, true, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil).
					AddRow("no-cluster", nil, 1, EntityTypeGroup, nil, nil, nil, false, "viewer", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil).
					AddRow("invalid-department", nil, 1, EntityTypeGroup, nil, nil, nil, false, "department_admin", "cluster-1", nil, nil, "cluster", nil, nil, nil, nil, nil, nil).
					AddRow("invalid-project", nil, 1, EntityTypeGroup, nil, nil, nil, false, "researcher", "cluster-1", nil, nil, "cluster", nil, nil, nil, nil, nil, nil)
				sourceMock.ExpectQuery("SELECT DISTINCT au.user_id AS user_id,").WillReturnRows(rows)

				targetDB := mocks.NewDB(GinkgoT())
				targetDB.On("Select", "id").Return(targetDB)
				targetDB.On("Where", "name = ?", mock.Anything).Return(targetDB)
				targetDB.On("First", mock.Anything).
					Run(func(args mock.Arguments) {
						args.Get(0).(*models.Role).ID = 42
					}).
					Return(nil)
				targetDB.On("Clauses", mock.Anything).Return(targetDB)
				targetDB.On("Create", mock.Anything).Return(int64(1), nil)

				Migrate(sourceDB, tenantsDB, targetDB)

				targetDB.AssertNumberOfCalls(GinkgoT(), "Create", 7)
			})
		})
	})
})
