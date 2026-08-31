/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CommandLine tests", func() {
	Context("Basic functionality", func() {
		It("should parse simple command line arguments", func() {
			cmd := "command arg1 arg2 arg3"
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command", "arg1", "arg2", "arg3"}))
		})

		It("should handle empty input", func() {
			cmd := ""
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(BeNil())
		})

		It("should handle whitespace-only input", func() {
			cmd := "   \t  "
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(BeNil())
		})
	})

	Context("Quoted strings", func() {
		It("should handle double-quoted strings", func() {
			cmd := `command "arg with spaces" arg3`
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command", "arg with spaces", "arg3"}))
		})

		It("should handle single-quoted strings", func() {
			cmd := `command 'arg with spaces' arg3`
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command", "arg with spaces", "arg3"}))
		})

		It("should detect unclosed quotes and return error", func() {
			cmd := `command "unclosed quote`
			args, err := ParseCommandLine(cmd)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unclosed quote"))
			Expect(args).To(Equal([]string{}))
		})

		It("should handle empty quoted strings", func() {
			cmd := `command "" arg3`
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command", "", "arg3"}))
		})

		It("should handle multiple quoted strings", func() {
			cmd := `command "first quoted" 'second quoted' arg3`
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command", "first quoted", "second quoted", "arg3"}))
		})
	})

	Context("Escape sequences", func() {
		It("should handle escaped spaces", func() {
			cmd := `command arg\ with\ escaped\ spaces arg2`
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command", "arg with escaped spaces", "arg2"}))
		})

		It("should handle escaped quotes within quoted strings", func() {
			cmd := `command "string with \"escaped\" quotes" arg2`
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command", `string with "escaped" quotes`, "arg2"}))
		})

		It("should handle escaped quotes outside quoted strings", func() {
			cmd := `command arg1 \"arg2\"`
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command", "arg1", `"arg2"`}))
		})

		It("should handle escaped backslashes", func() {
			cmd := `command arg1 \\arg2`
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"command", "arg1", "arg2"}))
		})
	})

	Context("Complex examples", func() {
		It("should parse docker run command", func() {
			cmd := `docker run -d --name test -p 8080:80 -v /host:/container nginx:latest`
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{
				"docker", "run", "-d", "--name", "test", "-p", "8080:80",
				"-v", "/host:/container", "nginx:latest",
			}))
		})

		It("should parse complex kubectl command", func() {
			cmd := `kubectl exec -it pod-name -- /bin/bash -c "echo 'hello world'"`
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{
				"kubectl", "exec", "-it", "pod-name", "--", "/bin/bash",
				"-c", "echo 'hello world'",
			}))
		})
	})

	Context("Edge cases", func() {
		It("should handle command with only quotes", func() {
			cmd := `""`
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{""}))
		})

		It("should handle command with tabs", func() {
			cmd := "cmd\targ1\targ2"
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"cmd", "arg1", "arg2"}))
		})

		It("should handle command with mixed whitespace", func() {
			cmd := "cmd   arg1\t\t  arg2"
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"cmd", "arg1", "arg2"}))
		})

		It("should handle command with special character", func() {
			// note this is a special character not hyphen
			stringWithDash := "this is dash –"
			cmd := fmt.Sprintf("runai training standard submit testdash2 -i nginx -- echo '%s'", stringWithDash)
			args, err := ParseCommandLine(cmd)
			Expect(err).NotTo(HaveOccurred())
			Expect(args).To(Equal([]string{"runai", "training", "standard", "submit", "testdash2", "-i", "nginx", "--", "echo", stringWithDash}))
			Expect(args[9]).To(Equal(stringWithDash))
		})
	})
})
