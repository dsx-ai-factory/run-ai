/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package upgrade

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("MsiInstaller", func() {
	Context("NewMsiInstaller", func() {
		It("should set msiExecBin and args correctly with quiet option", func() {
			msiInstaller := NewMsiInstaller("msiexec", true)
			Expect(msiInstaller.msiExecBin).To(Equal("msiexec"))
			Expect(msiInstaller.args).To(ContainElement("/i"))
			Expect(msiInstaller.args).To(ContainElement("/quiet"))
		})

		It("should set msiExecBin and args correctly without quiet option", func() {
			msiInstaller := NewMsiInstaller("msiexec", false)
			Expect(msiInstaller.msiExecBin).To(Equal("msiexec"))
			Expect(msiInstaller.args).To(ContainElement("/i"))
			Expect(msiInstaller.args).NotTo(ContainElement("/quiet"))
		})
	})

})

var _ = Describe("UnixScriptInstaller", func() {

	Context("NewUnixScriptInstaller", func() {
		It("should set quite correctly when quite is true", func() {
			unixScriptInstaller := NewUnixScriptInstaller(true)
			Expect(unixScriptInstaller.quite).To(BeTrue())
		})

		It("should set quite correctly when quite is false", func() {
			unixScriptInstaller := NewUnixScriptInstaller(false)
			Expect(unixScriptInstaller.quite).To(BeFalse())
		})
	})
	Context("Install", func() {
		var scriptPath string
		var execShell string

		BeforeEach(func() {
			script, err := os.CreateTemp("", "test")
			Expect(err).NotTo(HaveOccurred())
			defer script.Close()
			scriptPath = script.Name()
			execShell = "bash"
			_, err = script.WriteString("#!/bin/bash\necho 'test'")
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			err := os.Remove(scriptPath)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("quite is true", func() {
			It("should run the script with quite option", func() {
				unixScriptInstaller := NewUnixScriptInstaller(true)
				output, err := unixScriptInstaller.Install(context.Background(), scriptPath, execShell)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).To(Equal(""))
			})
		})
		Context("quite is false", func() {
			It("should run the script without quite option", func() {
				unixScriptInstaller := NewUnixScriptInstaller(false)
				output, err := unixScriptInstaller.Install(context.Background(), scriptPath, execShell)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).ToNot(Equal(""))
			})
		})
	})
})
