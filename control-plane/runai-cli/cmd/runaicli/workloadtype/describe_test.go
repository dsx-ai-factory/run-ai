/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloadtype_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/workloadtype"
)

var _ = Describe("Workload Type Describe Command", func() {
	Describe("Command Structure", func() {
		It("should have correct use string", func() {
			Expect(workloadtype.Describe.Use).To(Equal("describe WORKLOAD_TYPE_NAME"))
		})

		It("should require exactly one argument", func() {
			Expect(workloadtype.Describe.Args).NotTo(BeNil())
		})

		It("should have correct short description", func() {
			Expect(workloadtype.Describe.Short).To(Equal("Describe a workload type"))
		})
	})

	Describe("Flags", func() {
		It("should have output flag with table as default", func() {
			flag := workloadtype.Describe.Flags().Lookup("output")
			Expect(flag).NotTo(BeNil())
			Expect(flag.DefValue).To(Equal("table"))
			Expect(flag.Shorthand).To(Equal("o"))
		})

		It("should allow parsing output flag with json", func() {
			err := workloadtype.Describe.Flags().Parse([]string{"--output", "json"})
			Expect(err).To(BeNil())
		})

		It("should allow parsing output flag with yaml", func() {
			err := workloadtype.Describe.Flags().Parse([]string{"--output", "yaml"})
			Expect(err).To(BeNil())
		})

		It("should allow parsing output shorthand flag", func() {
			err := workloadtype.Describe.Flags().Parse([]string{"-o", "json"})
			Expect(err).To(BeNil())
		})
	})
})
