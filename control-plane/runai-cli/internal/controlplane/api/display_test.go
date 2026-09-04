/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	"github.com/charmbracelet/bubbles/table"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MultiWorkloadsErrorTable", func() {
	var (
		workloads  []MinimalWorkloadInfo
		errorTable *MultiWorkloadsErrorTable
	)

	BeforeEach(func() {
		workloads = []MinimalWorkloadInfo{
			{Name: "w1", Project: "p1", Kind: "k1", ID: "id1"},
			{Name: "w2", Project: "p2", Kind: "k2", ID: "id2"},
		}
		errorTable = NewMultiWorkloadsErrorTable(MultiWorkloadsError{Workloads: workloads})
	})

	Describe("NewMultiWorkloadsErrorTable", func() {
		It("should create a new MultiWorkloadsErrorTable with the given workloads", func() {
			Expect(errorTable.Workloads).To(Equal(workloads))
		})
	})

	Describe("Subject", func() {
		It("should return 'Multiple Workloads'", func() {
			Expect(errorTable.Subject()).To(Equal("Multiple Workloads"))
		})
	})

	Describe("Columns", func() {
		It("should return the correct columns", func() {
			expectedColumns := []table.Column{
				{Title: "Name", Width: 20},
				{Title: "Project", Width: 20},
				{Title: "Type", Width: 20},
				{Title: "ID", Width: 20},
			}
			Expect(errorTable.Columns()).To(Equal(expectedColumns))
		})
	})

	Describe("Data", func() {
		It("should return the correct data", func() {
			expectedData := [][]string{
				{"w1", "p1", "k1", "id1"},
				{"w2", "p2", "k2", "id2"},
			}
			Expect(errorTable.Data()).To(Equal(expectedData))
		})
	})
})
