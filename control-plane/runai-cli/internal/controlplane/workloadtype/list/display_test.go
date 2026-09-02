/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workloadtype/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workloadtype/list/manipulate"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

var _ = Describe("Workload Type List Display", func() {
	var (
		output        *list.Output
		workloadTypes []openapi.WorkloadTypeConfig
	)

	BeforeEach(func() {
		wt1 := newTestWorkloadType("interactive", "Interactive", "Medium", openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE)
		wt2 := newTestWorkloadType("train", "Train", "High", openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE)
		workloadTypes = []openapi.WorkloadTypeConfig{wt1, wt2}

		var err error
		output, err = list.ParseListResponse(workloadTypes)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("ParseListResponse", func() {
		It("should create output successfully", func() {
			Expect(output).ToNot(BeNil())
		})

		It("should handle empty workload types list", func() {
			emptyOutput, err := list.ParseListResponse([]openapi.WorkloadTypeConfig{})
			Expect(err).ToNot(HaveOccurred())
			Expect(emptyOutput).ToNot(BeNil())
		})
	})

	Context("SetRaw", func() {
		It("should set raw flag", func() {
			output.SetRaw(true)
			// No direct way to verify, but ensure no panic
		})
	})

	Context("SetColumns", func() {
		It("should set valid columns", func() {
			err := output.SetColumns([]string{"name", "category", "priority"})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error for invalid columns", func() {
			err := output.SetColumns([]string{"invalid"})
			Expect(err).To(HaveOccurred())
		})

		It("should handle empty columns list", func() {
			err := output.SetColumns([]string{})
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Subject", func() {
		It("should return correct subject", func() {
			subject := output.Subject()
			Expect(subject).To(Equal("workload-type"))
		})
	})

	Context("Columns", func() {
		It("should return table columns", func() {
			err := output.SetColumns([]string{"name", "category", "priority"})
			Expect(err).ToNot(HaveOccurred())

			columns := output.Columns()
			Expect(columns).To(HaveLen(3))
			Expect(columns[0].Title).To(Equal("Name"))
			Expect(columns[1].Title).To(Equal("Category"))
			Expect(columns[2].Title).To(Equal("Priority"))
		})

		It("should return columns when explicitly set to all columns including source", func() {
			err := output.SetColumns([]string{"name", "category", "priority", "preemptibility", "group", "source"})
			Expect(err).ToNot(HaveOccurred())

			columns := output.Columns()
			Expect(columns).To(HaveLen(6)) // All available columns
		})
	})

	Context("Data", func() {
		It("should return table data", func() {
			err := output.SetColumns([]string{"name", "category", "priority"})
			Expect(err).ToNot(HaveOccurred())

			data := output.Data()
			Expect(data).To(HaveLen(2))
			Expect(data[0][0]).To(Equal("interactive"))
			Expect(data[0][1]).To(Equal("Interactive"))
			Expect(data[0][2]).To(Equal("Medium"))

			Expect(data[1][0]).To(Equal("train"))
			Expect(data[1][1]).To(Equal("Train"))
			Expect(data[1][2]).To(Equal("High"))
		})

		It("should handle all available columns including source", func() {
			err := output.SetColumns([]string{"name", "category", "priority", "preemptibility", "group", "source"})
			Expect(err).ToNot(HaveOccurred())

			data := output.Data()
			Expect(data).To(HaveLen(2))
			Expect(data[0]).To(HaveLen(6)) // All 6 available columns
		})
	})

	Context("GenerateJson", func() {
		It("should generate JSON output", func() {
			err := output.SetColumns([]string{"name", "category"})
			Expect(err).ToNot(HaveOccurred())

			json, err := output.GenerateJson()
			Expect(err).ToNot(HaveOccurred())
			Expect(json).To(ContainSubstring("interactive"))
			Expect(json).To(ContainSubstring("train"))
		})

		It("should generate raw JSON output", func() {
			output.SetRaw(true)
			json, err := output.GenerateJson()
			Expect(err).ToNot(HaveOccurred())
			Expect(json).To(ContainSubstring("interactive"))
		})
	})

	Context("GenerateYaml", func() {
		It("should generate YAML output", func() {
			err := output.SetColumns([]string{"name", "category"})
			Expect(err).ToNot(HaveOccurred())

			yaml, err := output.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())
			Expect(yaml).To(ContainSubstring("interactive"))
			Expect(yaml).To(ContainSubstring("train"))
		})

		It("should generate raw YAML output", func() {
			output.SetRaw(true)
			yaml, err := output.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())
			Expect(yaml).To(ContainSubstring("interactive"))
		})
	})

	Context("GenerateTable", func() {
		It("should generate table output", func() {
			table, err := output.GenerateTable()
			Expect(err).ToNot(HaveOccurred())
			Expect(table).To(ContainSubstring("interactive"))
			Expect(table).To(ContainSubstring("train"))
		})
	})

	Context("IsEmpty", func() {
		It("should return false for non-empty output", func() {
			Expect(output.IsEmpty()).To(BeFalse())
		})

		It("should return true for empty output", func() {
			emptyOutput, err := list.ParseListResponse([]openapi.WorkloadTypeConfig{})
			Expect(err).ToNot(HaveOccurred())
			Expect(emptyOutput.IsEmpty()).To(BeTrue())
		})
	})

	Context("GetList", func() {
		It("should return brief workload types", func() {
			briefList := output.GetList()
			Expect(briefList).To(HaveLen(2))
			Expect(briefList[0].Name).To(Equal("interactive"))
			Expect(briefList[1].Name).To(Equal("train"))
		})
	})
})

var _ = Describe("Default Columns", func() {
	It("should have correct default columns from manipulate package", func() {
		defaults := manipulate.GetDefaultColumns()
		Expect(defaults).To(HaveLen(5))
		Expect(defaults).To(ContainElement(manipulate.ColumnKeyName))
		Expect(defaults).To(ContainElement(manipulate.ColumnKeyCategory))
		Expect(defaults).To(ContainElement(manipulate.ColumnKeyPriority))
		Expect(defaults).To(ContainElement(manipulate.ColumnKeyPreemptibility))
		Expect(defaults).To(ContainElement(manipulate.ColumnKeyGroup))
		// Note: source is available but not in default columns
	})
})

// Helper function to create test workload type configs
func newTestWorkloadType(name string, category string, priority string, preemptibility openapi.WorkloadTypePreemptibility) openapi.WorkloadTypeConfig {
	return openapi.WorkloadTypeConfig{
		Name:           name,
		CategoryName:   &category,
		PriorityName:   &priority,
		Preemptibility: preemptibility,
		Group:          "default",
		CategoryId:     "cat-id",
		PriorityId:     "pri-id",
	}
}
