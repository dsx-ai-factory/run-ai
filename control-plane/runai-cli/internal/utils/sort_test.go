/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MultiColumnSort", func() {
	Describe("Single Column Sorting", func() {
		Context("when sorting ascending", func() {
			It("should sort by single column in ascending order", func() {
				data := [][]string{
					{"charlie", "30"},
					{"alice", "25"},
					{"bob", "35"},
				}

				expected := [][]string{
					{"alice", "25"},
					{"bob", "35"},
					{"charlie", "30"},
				}

				MultiColumnSort(data, []SortColumn{
					{Index: 0, Order: Ascending},
				})

				Expect(data).To(Equal(expected))
			})
		})

		Context("when sorting descending", func() {
			It("should sort by single column in descending order", func() {
				data := [][]string{
					{"alice", "25"},
					{"bob", "35"},
					{"charlie", "30"},
				}

				expected := [][]string{
					{"charlie", "30"},
					{"bob", "35"},
					{"alice", "25"},
				}

				MultiColumnSort(data, []SortColumn{
					{Index: 0, Order: Descending},
				})

				Expect(data).To(Equal(expected))
			})
		})
	})

	Describe("Multi-Column Sorting", func() {
		Context("when both columns are ascending", func() {
			It("should sort by first column, then by second column", func() {
				data := [][]string{
					{"active", "john", "30"},
					{"inactive", "alice", "25"},
					{"active", "bob", "35"},
					{"inactive", "charlie", "28"},
					{"active", "alice", "32"},
				}

				expected := [][]string{
					{"active", "alice", "32"},
					{"active", "bob", "35"},
					{"active", "john", "30"},
					{"inactive", "alice", "25"},
					{"inactive", "charlie", "28"},
				}

				MultiColumnSort(data, []SortColumn{
					{Index: 0, Order: Ascending}, // status
					{Index: 1, Order: Ascending}, // name
				})

				Expect(data).To(Equal(expected))
			})
		})

		Context("when columns have mixed sort orders", func() {
			It("should respect individual column sort orders", func() {
				data := [][]string{
					{"active", "john", "30"},
					{"inactive", "alice", "25"},
					{"active", "bob", "35"},
					{"inactive", "charlie", "28"},
					{"active", "alice", "32"},
				}

				expected := [][]string{
					{"inactive", "charlie", "28"},
					{"inactive", "alice", "25"},
					{"active", "john", "30"},
					{"active", "bob", "35"},
					{"active", "alice", "32"},
				}

				MultiColumnSort(data, []SortColumn{
					{Index: 0, Order: Descending}, // status (inactive first)
					{Index: 1, Order: Descending}, // name (reverse alphabetical)
				})

				Expect(data).To(Equal(expected))
			})
		})

		Context("when sorting by three columns", func() {
			It("should sort by all three columns in priority order", func() {
				data := [][]string{
					{"dept1", "active", "30"},
					{"dept2", "active", "25"},
					{"dept1", "active", "25"},
					{"dept1", "inactive", "30"},
					{"dept2", "inactive", "25"},
				}

				expected := [][]string{
					{"dept1", "active", "25"},
					{"dept1", "active", "30"},
					{"dept1", "inactive", "30"},
					{"dept2", "active", "25"},
					{"dept2", "inactive", "25"},
				}

				MultiColumnSort(data, []SortColumn{
					{Index: 0, Order: Ascending}, // department
					{Index: 1, Order: Ascending}, // status
					{Index: 2, Order: Ascending}, // age
				})

				Expect(data).To(Equal(expected))
			})
		})
	})

	Describe("Edge Cases", func() {
		Context("when slice is empty", func() {
			It("should handle empty slice without error", func() {
				data := [][]string{}
				expected := [][]string{}

				MultiColumnSort(data, []SortColumn{
					{Index: 0, Order: Ascending},
				})

				Expect(data).To(Equal(expected))
			})
		})

		Context("when slice has single row", func() {
			It("should handle single row without changes", func() {
				data := [][]string{
					{"alice", "25"},
				}
				expected := [][]string{
					{"alice", "25"},
				}

				MultiColumnSort(data, []SortColumn{
					{Index: 0, Order: Ascending},
				})

				Expect(data).To(Equal(expected))
			})
		})

		Context("when column index is out of bounds", func() {
			It("should handle gracefully and sort by valid columns", func() {
				data := [][]string{
					{"alice", "25"},
					{"bob", "30"},
					{"charlie"},
				}

				Expect(func() {
					MultiColumnSort(data, []SortColumn{
						{Index: 0, Order: Ascending}, // valid column
						{Index: 5, Order: Ascending}, // out of bounds
					})
				}).NotTo(Panic())

				// Should be sorted by first column at least
				Expect(data[0][0]).To(Equal("alice"))
				Expect(data[1][0]).To(Equal("bob"))
			})
		})

		Context("when no sort columns are specified", func() {
			It("should leave data unchanged", func() {
				data := [][]string{
					{"charlie", "30"},
					{"alice", "25"},
					{"bob", "35"},
				}

				original := make([][]string, len(data))
				copy(original, data)

				MultiColumnSort(data, []SortColumn{})

				Expect(data).To(Equal(original))
			})
		})

		Context("when all values are identical", func() {
			It("should handle identical values correctly", func() {
				data := [][]string{
					{"active", "alice"},
					{"active", "alice"},
					{"active", "alice"},
				}

				expected := [][]string{
					{"active", "alice"},
					{"active", "alice"},
					{"active", "alice"},
				}

				MultiColumnSort(data, []SortColumn{
					{Index: 0, Order: Ascending},
					{Index: 1, Order: Ascending},
				})

				Expect(data).To(Equal(expected))
			})
		})

		Context("when sorting numeric strings", func() {
			It("should sort lexicographically, not numerically", func() {
				data := [][]string{
					{"100", "alice"},
					{"20", "bob"},
					{"5", "charlie"},
				}

				// Note: This sorts lexicographically, not numerically
				expected := [][]string{
					{"100", "alice"},
					{"20", "bob"},
					{"5", "charlie"},
				}

				MultiColumnSort(data, []SortColumn{
					{Index: 0, Order: Ascending},
				})

				Expect(data).To(Equal(expected))
			})
		})
	})

	Describe("Sorting Behavior Validation", func() {
		Context("when first column values are equal", func() {
			It("should use second column as tiebreaker", func() {
				data := [][]string{
					{"active", "zebra"},
					{"active", "apple"},
					{"active", "banana"},
				}

				expected := [][]string{
					{"active", "apple"},
					{"active", "banana"},
					{"active", "zebra"},
				}

				MultiColumnSort(data, []SortColumn{
					{Index: 0, Order: Ascending},
					{Index: 1, Order: Ascending},
				})

				Expect(data).To(Equal(expected))
			})
		})

		Context("when testing column priority", func() {
			It("should prioritize first column over second column", func() {
				data := [][]string{
					{"z", "a"},
					{"a", "z"},
				}

				expected := [][]string{
					{"a", "z"},
					{"z", "a"},
				}

				MultiColumnSort(data, []SortColumn{
					{Index: 0, Order: Ascending}, // This should take priority
					{Index: 1, Order: Ascending},
				})

				Expect(data).To(Equal(expected))
			})
		})
	})
})
