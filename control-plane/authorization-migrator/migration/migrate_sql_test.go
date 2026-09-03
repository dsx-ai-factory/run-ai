/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const legacyRoleAssignmentsQueryPrefix = "SELECT DISTINCT au.user_id AS user_id,"

var _ = Describe("Migrate SQL", func() {
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
		DeferCleanup(func() { _ = sourceDB.Close() })

		tenantsDB, tenantsMock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = tenantsDB.Close() })

		_ = dbtesting.SetSQLMock(GinkgoT())
	})

	AfterEach(func() {
		Expect(sourceMock.ExpectationsWereMet()).To(Succeed())
		Expect(tenantsMock.ExpectationsWereMet()).To(Succeed())
	})

	It("skips migration when legacy tables are missing (42P01)", func() {
		tenantsMock.ExpectQuery("SELECT id AS tenant_id, name AS tenant_name FROM tenants").
			WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "tenant_name"}))
		sourceMock.ExpectQuery(legacyRoleAssignmentsQueryPrefix).
			WillReturnError(&pgconn.PgError{Code: "42P01"})

		Expect(func() {
			Migrate(sourceDB, tenantsDB, runaidb.DB())
		}).NotTo(Panic())
	})

	It("skips migration when 42P01 is wrapped (e.g. PrepareError)", func() {
		tenantsMock.ExpectQuery("SELECT id AS tenant_id, name AS tenant_name FROM tenants").
			WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "tenant_name"}))
		sourceMock.ExpectQuery(legacyRoleAssignmentsQueryPrefix).
			WillReturnError(fmt.Errorf("prepare failed: %w", &pgconn.PgError{Code: "42P01"}))

		Expect(func() {
			Migrate(sourceDB, tenantsDB, runaidb.DB())
		}).NotTo(Panic())
	})

	It("panics on unexpected source query errors", func() {
		tenantsMock.ExpectQuery("SELECT id AS tenant_id, name AS tenant_name FROM tenants").
			WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "tenant_name"}))
		sourceMock.ExpectQuery(legacyRoleAssignmentsQueryPrefix).
			WillReturnError(errors.New("connection refused"))

		Expect(func() {
			Migrate(sourceDB, tenantsDB, runaidb.DB())
		}).To(Panic())
	})

	It("completes with no legacy rows", func() {
		tenantsMock.ExpectQuery("SELECT id AS tenant_id, name AS tenant_name FROM tenants").
			WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "tenant_name"}).
				AddRow(1, "tenant-one"))
		sourceMock.ExpectQuery(legacyRoleAssignmentsQueryPrefix).
			WillReturnRows(sqlmock.NewRows([]string{
				"user_id", "client_id", "tenant_id", "entity_type", "created_by",
				"created_at", "updated_at", "all_clusters_permitted", "role",
				"cluster_uuid", "department_id", "project_id", "cluster_name",
				"department_name", "department_cluster_uuid", "project_name",
				"project_cluster_uuid", "tenant_name", "auth_realm",
			}))

		Expect(func() {
			Migrate(sourceDB, tenantsDB, runaidb.DB())
		}).NotTo(Panic())
	})
})

var _ = Describe("getTenantsManagerTenants SQL", func() {
	It("returns tenants indexed by ID", func() {
		db, mock, err := sqlmock.New()
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = db.Close() })

		mock.ExpectQuery("SELECT id AS tenant_id, name AS tenant_name FROM tenants").
			WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "tenant_name"}).
				AddRow(1, "tenant-one").
				AddRow(2, "tenant-two"))

		Expect(getTenantsManagerTenants(db)).To(Equal(map[int]string{
			1: "tenant-one",
			2: "tenant-two",
		}))
		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})

	It("returns an empty map when the query fails", func() {
		db, mock, err := sqlmock.New()
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() { _ = db.Close() })

		mock.ExpectQuery("SELECT id AS tenant_id, name AS tenant_name FROM tenants").
			WillReturnError(errors.New("query failed"))

		Expect(getTenantsManagerTenants(db)).To(BeEmpty())
		Expect(mock.ExpectationsWereMet()).To(Succeed())
	})
})
