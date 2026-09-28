/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package template_test

import (
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/template/describe"

	workloads_openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Helper to create a template with specific scope
func createTemplate(id, name string, scopeType workloads_openapi.ScopeType) workloads_openapi.SupersetTemplate {
	createdAt := time.Now()
	return workloads_openapi.SupersetTemplate{
		Meta: workloads_openapi.TemplateMetaFields{
			Id:           id,
			Name:         name,
			ScopeType:    scopeType,
			ScopeId:      "test-scope",
			WorkloadType: workloads_openapi.WORKLOADTYPE_TRAINING,
			CreatedAt:    createdAt,
			CreatedBy:    "test-user",
		},
	}
}

var _ = Describe("Template Scope Priority", func() {
	Context("when sorting templates by scope", func() {
		It("should prioritize project scope highest", func() {
			templates := []workloads_openapi.SupersetTemplate{
				createTemplate("cluster-id", "t1", workloads_openapi.SCOPETYPE_CLUSTER),
				createTemplate("project-id", "t1", workloads_openapi.SCOPETYPE_PROJECT),
				createTemplate("department-id", "t1", workloads_openapi.SCOPETYPE_DEPARTMENT),
			}

			sorted := describe.SortByScope(templates)
			Expect(sorted).NotTo(BeEmpty())
			Expect(sorted[0].Meta.GetId()).To(Equal("project-id"))
		})

		It("should select department scope when no project exists", func() {
			templates := []workloads_openapi.SupersetTemplate{
				createTemplate("cluster-id", "t1", workloads_openapi.SCOPETYPE_CLUSTER),
				createTemplate("department-id", "t1", workloads_openapi.SCOPETYPE_DEPARTMENT),
			}

			sorted := describe.SortByScope(templates)
			Expect(sorted).NotTo(BeEmpty())
			Expect(sorted[0].Meta.GetId()).To(Equal("department-id"))
		})

		It("should select cluster scope when no project or department exists", func() {
			templates := []workloads_openapi.SupersetTemplate{
				createTemplate("cluster-id", "t1", workloads_openapi.SCOPETYPE_CLUSTER),
				createTemplate("tenant-id", "t1", workloads_openapi.SCOPETYPE_TENANT),
			}

			sorted := describe.SortByScope(templates)
			Expect(sorted).NotTo(BeEmpty())
			Expect(sorted[0].Meta.GetId()).To(Equal("cluster-id"))
		})

		It("should select tenant scope over system", func() {
			templates := []workloads_openapi.SupersetTemplate{
				createTemplate("tenant-id", "t1", workloads_openapi.SCOPETYPE_TENANT),
				createTemplate("system-id", "t1", workloads_openapi.SCOPETYPE_SYSTEM),
			}

			sorted := describe.SortByScope(templates)
			Expect(sorted).NotTo(BeEmpty())
			Expect(sorted[0].Meta.GetId()).To(Equal("tenant-id"))
		})

		It("should select system scope when it's the only one", func() {
			templates := []workloads_openapi.SupersetTemplate{
				createTemplate("system-id", "t1", workloads_openapi.SCOPETYPE_SYSTEM),
			}

			sorted := describe.SortByScope(templates)
			Expect(sorted).NotTo(BeEmpty())
			Expect(sorted[0].Meta.GetId()).To(Equal("system-id"))
		})

		It("should select project when all scopes are present", func() {
			templates := []workloads_openapi.SupersetTemplate{
				createTemplate("system-id", "t1", workloads_openapi.SCOPETYPE_SYSTEM),
				createTemplate("tenant-id", "t1", workloads_openapi.SCOPETYPE_TENANT),
				createTemplate("cluster-id", "t1", workloads_openapi.SCOPETYPE_CLUSTER),
				createTemplate("department-id", "t1", workloads_openapi.SCOPETYPE_DEPARTMENT),
				createTemplate("project-id", "t1", workloads_openapi.SCOPETYPE_PROJECT),
			}

			sorted := describe.SortByScope(templates)
			Expect(sorted).NotTo(BeEmpty())
			Expect(sorted[0].Meta.GetId()).To(Equal("project-id"))
		})
	})
})

var _ = Describe("Template Scope Priority Order", func() {
	It("should maintain correct priority order across all scopes", func() {
		templates := []workloads_openapi.SupersetTemplate{
			createTemplate("system-id", "t1", workloads_openapi.SCOPETYPE_SYSTEM),
			createTemplate("tenant-id", "t1", workloads_openapi.SCOPETYPE_TENANT),
			createTemplate("cluster-id", "t1", workloads_openapi.SCOPETYPE_CLUSTER),
			createTemplate("department-id", "t1", workloads_openapi.SCOPETYPE_DEPARTMENT),
			createTemplate("project-id", "t1", workloads_openapi.SCOPETYPE_PROJECT),
		}

		sorted := describe.SortByScope(templates)

		expectedOrder := []string{
			"project-id",
			"department-id",
			"cluster-id",
			"tenant-id",
			"system-id",
		}

		Expect(sorted).To(HaveLen(len(expectedOrder)))

		for i, expectedId := range expectedOrder {
			Expect(sorted[i].Meta.GetId()).To(Equal(expectedId))
		}
	})
})

var _ = Describe("GetScopePriority", func() {
	It("should return correct priority values", func() {
		Expect(describe.GetScopePriority(workloads_openapi.SCOPETYPE_PROJECT)).To(Equal(1))
		Expect(describe.GetScopePriority(workloads_openapi.SCOPETYPE_DEPARTMENT)).To(Equal(2))
		Expect(describe.GetScopePriority(workloads_openapi.SCOPETYPE_CLUSTER)).To(Equal(3))
		Expect(describe.GetScopePriority(workloads_openapi.SCOPETYPE_TENANT)).To(Equal(4))
		Expect(describe.GetScopePriority(workloads_openapi.SCOPETYPE_SYSTEM)).To(Equal(5))
	})

	It("should return lower priority for project than department", func() {
		projectPriority := describe.GetScopePriority(workloads_openapi.SCOPETYPE_PROJECT)
		departmentPriority := describe.GetScopePriority(workloads_openapi.SCOPETYPE_DEPARTMENT)
		Expect(projectPriority).To(BeNumerically("<", departmentPriority))
	})

	It("should return default priority for unknown scope type", func() {
		// Create an invalid scope type
		var unknownScope workloads_openapi.ScopeType = "UNKNOWN"
		Expect(describe.GetScopePriority(unknownScope)).To(Equal(6))
	})
})

var _ = Describe("SortByScope Edge Cases", func() {
	It("should handle empty templates list", func() {
		templates := []workloads_openapi.SupersetTemplate{}
		sorted := describe.SortByScope(templates)
		Expect(sorted).To(BeEmpty())
	})

	It("should handle single template", func() {
		templates := []workloads_openapi.SupersetTemplate{
			createTemplate("single-id", "t1", workloads_openapi.SCOPETYPE_CLUSTER),
		}
		sorted := describe.SortByScope(templates)
		Expect(sorted).To(HaveLen(1))
		Expect(sorted[0].Meta.GetId()).To(Equal("single-id"))
	})

	It("should not modify the original slice", func() {
		templates := []workloads_openapi.SupersetTemplate{
			createTemplate("cluster-id", "t1", workloads_openapi.SCOPETYPE_CLUSTER),
			createTemplate("project-id", "t1", workloads_openapi.SCOPETYPE_PROJECT),
		}

		originalFirstId := templates[0].Meta.GetId()
		sorted := describe.SortByScope(templates)

		// Original should be unchanged
		Expect(templates[0].Meta.GetId()).To(Equal(originalFirstId))
		// Sorted should have project first
		Expect(sorted[0].Meta.GetId()).To(Equal("project-id"))
	})

	It("should handle multiple templates with same scope type", func() {
		templates := []workloads_openapi.SupersetTemplate{
			createTemplate("cluster-1", "t1", workloads_openapi.SCOPETYPE_CLUSTER),
			createTemplate("cluster-2", "t2", workloads_openapi.SCOPETYPE_CLUSTER),
			createTemplate("project-1", "t3", workloads_openapi.SCOPETYPE_PROJECT),
		}

		sorted := describe.SortByScope(templates)
		Expect(sorted).To(HaveLen(3))
		// Project should be first
		Expect(sorted[0].Meta.GetId()).To(Equal("project-1"))
		// The two cluster templates should follow (order doesn't matter between same priority)
		Expect(sorted[1].Meta.ScopeType).To(Equal(workloads_openapi.SCOPETYPE_CLUSTER))
		Expect(sorted[2].Meta.ScopeType).To(Equal(workloads_openapi.SCOPETYPE_CLUSTER))
	})
})
