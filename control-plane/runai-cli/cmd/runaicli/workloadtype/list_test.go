/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloadtype_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/workloadtype"
)

var _ = Describe("Workload Type List Command", func() {
	Describe("Command Structure", func() {
		It("should have correct use string", func() {
			Expect(workloadtype.List.Use).To(Equal("list"))
		})

		It("should accept no arguments", func() {
			Expect(workloadtype.List.Args).NotTo(BeNil())
		})

		It("should have correct short description", func() {
			Expect(workloadtype.List.Short).To(Equal("list available workload types"))
		})
	})

	Describe("Flags", func() {
		It("should have external-only flag", func() {
			flag := workloadtype.List.Flags().Lookup("external-only")
			Expect(flag).NotTo(BeNil())
			Expect(flag.DefValue).To(Equal("false"))
		})

		It("should have columns flag", func() {
			flag := workloadtype.List.Flags().Lookup("columns")
			Expect(flag).NotTo(BeNil())
		})

		It("should have set-default-columns flag", func() {
			flag := workloadtype.List.Flags().Lookup("set-default-columns")
			Expect(flag).NotTo(BeNil())
		})

		It("should allow parsing external-only flag", func() {
			err := workloadtype.List.Flags().Parse([]string{"--external-only"})
			Expect(err).To(BeNil())
		})

		It("should allow running without external-only flag", func() {
			err := workloadtype.List.Flags().Parse([]string{})
			Expect(err).To(BeNil())
		})

		It("should allow parsing columns flag", func() {
			err := workloadtype.List.Flags().Parse([]string{"--columns", "name,category,priority"})
			Expect(err).To(BeNil())
		})
	})
})
