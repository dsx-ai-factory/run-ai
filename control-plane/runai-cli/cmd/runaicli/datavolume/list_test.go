/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datavolume_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/datavolume"
)

var _ = Describe("Datavolume List Command", func() {
	Describe("Command Structure", func() {
		It("should have correct aliases", func() {
			Expect(datavolume.List.Aliases).To(ContainElement("ls"))
		})

		It("should accept no arguments", func() {
			Expect(datavolume.List.Args).NotTo(BeNil())
		})

		It("should have correct use string", func() {
			Expect(datavolume.List.Use).To(Equal("list"))
		})
	})

	Describe("Flags", func() {
		It("should have project flag with shorthand", func() {
			flag := datavolume.List.Flags().Lookup("project")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("p"))
		})

		It("should have all flag with shorthand", func() {
			flag := datavolume.List.Flags().Lookup("all")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("A"))
		})

		It("should allow running without project flag", func() {
			err := datavolume.List.Flags().Parse([]string{"-A"})
			Expect(err).To(BeNil())
		})
	})
})
