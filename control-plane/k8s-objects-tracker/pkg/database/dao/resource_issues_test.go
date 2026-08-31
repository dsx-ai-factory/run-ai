/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/k8s-objects-tracker/pkg/database/models"
)

var _ = Describe("ResourceIssues", func() {
	Describe("ListResourceIssues", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM resource_issues WHERE resource_issues.tenant_id = ?`).Returns([]models.ResourceIssue{})
			_, apiErr := ListResourceIssues(ctxWithTenant(1), &models.ResourceIssueFilters{TenantId: 1})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM resource_issues WHERE resource_issues.tenant_id = ?`).ReturnsError(errors.New("boom"))
			_, apiErr := ListResourceIssues(ctxWithTenant(1), &models.ResourceIssueFilters{TenantId: 1})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("ListResourceIssuesByIds", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM resource_issues WHERE resource_id IN (?,?)`).Returns([]models.ResourceIssue{})
			_, apiErr := ListResourceIssuesByIds(context.Background(), []string{"a", "b"})
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			m.ExpectQuery(`SELECT * FROM resource_issues WHERE resource_id IN (?,?)`).ReturnsError(errors.New("boom"))
			_, apiErr := ListResourceIssuesByIds(context.Background(), []string{"a", "b"})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("UpdateResourceIssues", func() {
		It("commits early on an empty list", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM resource_issues WHERE tenant_id = ? AND cluster_id = ? AND resource_id = ?`).Returns(2)
			tx.Commit()
			changes, apiErr := UpdateResourceIssues(ctxWithTenant(1), 1, "c1", "r1", []models.ResourceIssue{})
			Expect(apiErr).To(BeNil())
			Expect(changes.Removed).To(Equal(int64(2)))
			Expect(changes.Added).To(Equal(int64(0)))
		})

		It("deletes then creates", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM resource_issues WHERE tenant_id = ? AND cluster_id = ? AND resource_id = ?`).Returns(1)
			m.ExpectExec(`INSERT INTO resource_issues (id,tenant_id,cluster_id,scope_type,scope_id,resource_id,type,issue) VALUES (?,?,?,?,?,?,?,?)`).Returns(1)
			tx.Commit()
			changes, apiErr := UpdateResourceIssues(ctxWithTenant(1), 1, "c1", "r1", []models.ResourceIssue{
				{ResourceIssueFilters: models.ResourceIssueFilters{TenantId: 1, ResourceId: strPtr("r1")}, Issue: "oops"},
			})
			Expect(apiErr).To(BeNil())
			Expect(changes.Removed).To(Equal(int64(1)))
			Expect(changes.Added).To(Equal(int64(1)))
		})

		It("returns an error on create failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM resource_issues WHERE tenant_id = ? AND cluster_id = ? AND resource_id = ?`).Returns(1)
			m.ExpectExec(`INSERT INTO resource_issues (id,tenant_id,cluster_id,scope_type,scope_id,resource_id,type,issue) VALUES (?,?,?,?,?,?,?,?)`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			_, apiErr := UpdateResourceIssues(ctxWithTenant(1), 1, "c1", "r1", []models.ResourceIssue{
				{ResourceIssueFilters: models.ResourceIssueFilters{TenantId: 1, ResourceId: strPtr("r1")}, Issue: "oops"},
			})
			Expect(apiErr).ToNot(BeNil())
		})
	})

	Describe("DeleteProjectResourceIssues", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM resource_issues WHERE tenant_id = ? AND scope_type = ? AND scope_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteProjectResourceIssues(ctxWithTenant(1), 1, "p1")
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("DeleteClusterResourceIssues", func() {
		It("succeeds", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM resource_issues WHERE tenant_id = ? AND cluster_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteClusterResourceIssues(ctxWithTenant(1), 1, "c1")
			Expect(apiErr).To(BeNil())
		})
	})

	Describe("DeleteTenantResourceIssues", func() {
		It("(bug) targets the configmaps table", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM configmaps WHERE tenant_id = ?`).Returns(1)
			tx.Commit()
			apiErr := DeleteTenantResourceIssues(ctxWithTenant(1), 1)
			Expect(apiErr).To(BeNil())
		})

		It("returns an error on db failure", func() {
			m := setupDaoMock(GinkgoT())
			tx := m.ExpectTransaction()
			tx.Begin()
			m.ExpectExec(`DELETE FROM configmaps WHERE tenant_id = ?`).ReturnsError(errors.New("boom"))
			tx.Rollback()
			apiErr := DeleteTenantResourceIssues(ctxWithTenant(1), 1)
			Expect(apiErr).ToNot(BeNil())
		})
	})
})

func strPtr(s string) *string { return &s }
