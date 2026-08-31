/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/authorization/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/authorization/pkg/database/models"
	"github.com/run-ai/runai/runai-common-packages/runaidb/orm/mocks"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Migration utilities", func() {
	Describe("getSubjectType", func() {
		DescribeTable("maps entity types to subject types",
			func(entityType string, expected api.SubjectType) {
				Expect(getSubjectType(entityType)).To(Equal(expected))
			},
			Entry("application", EntityTypeApp, api.SUBJECTTYPE_APP),
			Entry("group", EntityTypeGroup, api.SUBJECTTYPE_GROUP),
			Entry("regular user", EntityTypeRegularUser, api.SUBJECTTYPE_USER),
			Entry("unknown defaults to user", "unknown", api.SUBJECTTYPE_USER),
		)
	})

	Describe("CreateAccessRule", func() {
		var db *mocks.DB

		BeforeEach(func() {
			db = mocks.NewDB(GinkgoT())
		})

		Context("when the role exists and the rule can be created", func() {
			It("creates an access rule with the expected fields", func() {
				db.On("Select", "id").Return(db)
				db.On("Where", "name = ?", "viewer").Return(db)
				db.On("First", mock.Anything).
					Run(func(args mock.Arguments) {
						args.Get(0).(*models.Role).ID = 42
					}).
					Return(nil)
				db.On("Clauses", mock.Anything).Return(db)
				db.On("Create", mock.Anything).
					Run(func(args mock.Arguments) {
						rule := args.Get(0).(*models.AccessRule)
						Expect(rule.RoleID).To(Equal(uint(42)))
						Expect(rule.SubjectId).To(Equal("subject"))
						Expect(rule.SubjectType).To(Equal(api.SUBJECTTYPEEXTENDED_GROUP))
						Expect(rule.TenantId).To(Equal(uint(7)))
						Expect(rule.ScopeId).To(Equal("scope"))
						Expect(rule.ScopeType).To(Equal(api.SCOPETYPE_CLUSTER))
						Expect(rule.ScopeName).To(Equal("scope-name"))
						Expect(rule.ClusterUuid).To(Equal("cluster"))
					}).
					Return(int64(1), nil)

				CreateAccessRule(
					db,
					7,
					"subject",
					EntityTypeGroup,
					"viewer",
					"scope",
					api.SCOPETYPE_CLUSTER,
					"scope-name",
					"cluster",
				)
			})
		})

		Context("when the role lookup fails", func() {
			It("panics with the lookup error", func() {
				db.On("Select", "id").Return(db)
				db.On("Where", "name = ?", "missing").Return(db)
				db.On("First", mock.Anything).Return(errors.New("lookup failed"))

				Expect(func() {
					CreateAccessRule(db, 1, "subject", EntityTypeRegularUser, "missing", "scope", api.SCOPETYPE_TENANT, "scope", "")
				}).To(PanicWith(MatchError("lookup failed")))
			})
		})

		Context("when creating the rule fails", func() {
			It("panics with the creation error", func() {
				db.On("Select", "id").Return(db)
				db.On("Where", "name = ?", "viewer").Return(db)
				db.On("First", mock.Anything).
					Run(func(args mock.Arguments) {
						args.Get(0).(*models.Role).ID = 42
					}).
					Return(nil)
				db.On("Clauses", mock.Anything).Return(db)
				db.On("Create", mock.Anything).Return(int64(0), errors.New("create failed"))

				Expect(func() {
					CreateAccessRule(db, 1, "subject", EntityTypeRegularUser, "viewer", "scope", api.SCOPETYPE_TENANT, "scope", "")
				}).To(PanicWith(MatchError("create failed")))
			})
		})
	})

	Describe("getTenantName", func() {
		Context("when the row contains a tenant name", func() {
			It("returns the row value", func() {
				tenantName := "row-tenant"
				Expect(getTenantName(RoleAssignmentRow{TenantName: &tenantName}, map[int]string{1: "fallback"})).To(Equal(tenantName))
			})
		})

		Context("when the row has no tenant name", func() {
			It("returns the tenant-manager value when available", func() {
				Expect(getTenantName(RoleAssignmentRow{TenantId: 1}, map[int]string{1: "fallback"})).To(Equal("fallback"))
			})

			It("returns the placeholder when no fallback exists", func() {
				Expect(getTenantName(RoleAssignmentRow{TenantId: 2}, map[int]string{1: "fallback"})).To(Equal("-"))
			})
		})
	})

	Describe("getAuthRealm", func() {
		Context("when the row contains an auth realm", func() {
			It("returns the row value", func() {
				authRealm := "row-realm"
				Expect(getAuthRealm(RoleAssignmentRow{AuthRealm: &authRealm}, map[int]string{1: "fallback"})).To(Equal(authRealm))
			})
		})

		Context("when the row has no auth realm", func() {
			It("returns the tenant-manager value when available", func() {
				Expect(getAuthRealm(RoleAssignmentRow{TenantId: 1}, map[int]string{1: "fallback"})).To(Equal("fallback"))
			})

			It("returns an empty string when no fallback exists", func() {
				Expect(getAuthRealm(RoleAssignmentRow{TenantId: 2}, map[int]string{1: "fallback"})).To(BeEmpty())
			})
		})
	})
})
