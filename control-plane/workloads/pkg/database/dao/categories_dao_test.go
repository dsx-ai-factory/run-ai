/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	categoriesListSelect   = "SELECT workloads_categories.id,workloads_categories.name,workloads_categories.description,workloads_categories.created_at,workloads_categories.created_by FROM workloads_categories WHERE deleted_at IS NULL AND ((scope_type = 'system' OR tenant_id = ?)) AND workloads_categories.deleted_at IS NULL"
	categoriesByNameSelect = "SELECT workloads_categories.id,workloads_categories.name,workloads_categories.description,workloads_categories.created_at,workloads_categories.created_by FROM workloads_categories WHERE deleted_at IS NULL AND name = ? AND ((tenant_id = ? OR scope_type = 'system')) AND workloads_categories.deleted_at IS NULL ORDER BY CASE WHEN scope_type = 'system' THEN 1 ELSE 0 END ASC,workloads_categories.id LIMIT ?"
	categoriesByIdSelect   = "SELECT workloads_categories.id,workloads_categories.name,workloads_categories.description,workloads_categories.created_at,workloads_categories.created_by FROM workloads_categories WHERE deleted_at IS NULL AND id = ? AND ((scope_type = 'system' OR tenant_id = ?)) AND workloads_categories.deleted_at IS NULL ORDER BY workloads_categories.id LIMIT ?"
)

type categoryRow struct {
	ID   string
	Name string
}

var _ = Describe("DbListCategories", func() {
	It("returns the tenant and system categories", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(categoriesListSelect).Returns([]categoryRow{{ID: "cat-1", Name: "Training"}})

		categories, apiErr := DbListCategories(testCtx(1), 1)

		Expect(apiErr).To(BeNil())
		Expect(categories).To(HaveLen(1))
		Expect(categories[0].Name).To(Equal("Training"))
	})

	It("returns an empty list when there are no categories", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(categoriesListSelect).Returns([]categoryRow{})

		categories, apiErr := DbListCategories(testCtx(1), 1)

		Expect(apiErr).To(BeNil())
		Expect(categories).To(BeEmpty())
	})

	It("returns a 503 when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(categoriesListSelect).ReturnsError(errors.New("boom"))

		categories, apiErr := DbListCategories(testCtx(1), 1)

		Expect(categories).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(Equal(int32(503)))
	})
})

var _ = Describe("DbGetCategoryByName", func() {
	It("returns the category", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(categoriesByNameSelect).Returns([]categoryRow{{ID: "cat-1", Name: "Training"}})

		category, apiErr := DbGetCategoryByName(testCtx(1), 1, "Training")

		Expect(apiErr).To(BeNil())
		Expect(category).ToNot(BeNil())
		Expect(category.Name).To(Equal("Training"))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(categoriesByNameSelect).ReturnsError(errors.New("boom"))

		category, apiErr := DbGetCategoryByName(testCtx(1), 1, "Training")

		Expect(category).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbGetCategoryById", func() {
	It("returns the category", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(categoriesByIdSelect).Returns([]categoryRow{{ID: "cat-1", Name: "Training"}})

		category, apiErr := DbGetCategoryById(testCtx(1), 1, "cat-1")

		Expect(apiErr).To(BeNil())
		Expect(category).ToNot(BeNil())
		Expect(category.Id).To(Equal("cat-1"))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(categoriesByIdSelect).ReturnsError(errors.New("boom"))

		category, apiErr := DbGetCategoryById(testCtx(1), 1, "cat-1")

		Expect(category).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})
