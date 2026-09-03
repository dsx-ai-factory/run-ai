/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package application_test

import (
	"bytes"
	"testing"

	application "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/application"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestApplication(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Application Tests")
}

var _ = Describe("Application Command", func() {
	var actual *bytes.Buffer

	BeforeEach(func() {
		actual = new(bytes.Buffer)
		application.Cmd.SetOut(actual)
		application.Cmd.SetErr(actual)
	})

	Context("application command", func() {
		It("should print help", func() {
			application.Cmd.SetArgs([]string{"--help"})
			err := application.Cmd.Execute()
			Expect(err).To(BeNil())
			Expect(actual.String()).To(ContainSubstring("ai application management"))
		})
	})

	Context("list subcommand flags", func() {
		It("should show list help", func() {
			application.Cmd.SetArgs([]string{"list", "--help"})
			err := application.Cmd.Execute()
			Expect(err).To(BeNil())
			output := actual.String()
			Expect(output).To(ContainSubstring("List AI applications"))
			Expect(output).To(ContainSubstring("--all"))
			Expect(output).To(ContainSubstring("--project"))
			Expect(output).To(ContainSubstring("--columns"))
			Expect(output).To(ContainSubstring("--json"))
			Expect(output).To(ContainSubstring("--yaml"))
			Expect(output).To(ContainSubstring("--no-headers"))
		})

		It("should have --all and --project in the same mutual exclusion group", func() {
			listCmd, _, err := application.Cmd.Find([]string{"list"})
			Expect(err).To(BeNil())
			Expect(listCmd).NotTo(BeNil())
			Expect(listCmd.Flags().Lookup("all")).NotTo(BeNil())
			Expect(listCmd.Flags().Lookup("project")).NotTo(BeNil())
		})
	})

	Context("submit subcommand flags", func() {
		It("should show submit help", func() {
			application.Cmd.SetArgs([]string{"submit", "--help"})
			err := application.Cmd.Execute()
			Expect(err).To(BeNil())
			output := actual.String()
			Expect(output).To(ContainSubstring("Submit an AI application"))
			Expect(output).To(ContainSubstring("--chart-url"))
			Expect(output).To(ContainSubstring("--values"))
			Expect(output).To(ContainSubstring("--set"))
			Expect(output).To(ContainSubstring("--credential"))
			Expect(output).To(ContainSubstring("--project"))
		})

		It("should belong to the AI applications command group", func() {
			Expect(application.Cmd.GroupID).To(Equal(consts.GroupApplicationID))
		})
	})

	Context("describe subcommand flags", func() {
		It("should show describe help", func() {
			application.Cmd.SetArgs([]string{"describe", "--help"})
			err := application.Cmd.Execute()
			Expect(err).To(BeNil())
			output := actual.String()
			Expect(output).To(ContainSubstring("Describe an AI application"))
			Expect(output).To(ContainSubstring("--project"))
			Expect(output).To(ContainSubstring("--output"))
		})
	})

	Context("delete subcommand flags", func() {
		It("should show delete help", func() {
			application.Cmd.SetArgs([]string{"delete", "--help"})
			err := application.Cmd.Execute()
			Expect(err).To(BeNil())
			output := actual.String()
			Expect(output).To(ContainSubstring("Delete an AI application"))
			Expect(output).To(ContainSubstring("--project"))
			Expect(output).To(ContainSubstring("--force"))
			Expect(output).To(ContainSubstring("--yes"))
		})

		It("should register force and yes as prompt-skipping flags", func() {
			deleteCmd, _, err := application.Cmd.Find([]string{"delete"})
			Expect(err).To(BeNil())
			Expect(deleteCmd).NotTo(BeNil())
			Expect(deleteCmd.Flags().Lookup("force")).NotTo(BeNil())
			Expect(deleteCmd.Flags().Lookup("yes")).NotTo(BeNil())
			Expect(deleteCmd.Flags().ShorthandLookup("y")).NotTo(BeNil())
		})
	})
})
