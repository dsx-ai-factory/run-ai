/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package root

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Root Tests")
}

const (
	TestFilePath = "root"

	ExpectedUnknownFlag = "Error: unknown flag: %s\n"
	NotValidFlag        = "--not-valid-flag"
	HelpCommandSuffix   = "Use \"runai [command] --help\" for more information about a command.\n"
)

var _ = Describe("Version Test", Ordered, func() {
	var (
		configPath string
		err        error
		envMap     map[string]string
	)

	BeforeAll(func() {
		configPath, err = cli.GetHomePath()
		if err != nil {
			Fail("failed to get pwd")
		}
		configPath = cli.GetConfigPath(configPath, TestFilePath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)
	})

	Context("Root command validation", Ordered, func() {
		It("With invalid flag", func() {
			output, err := cli.RunCommand(NotValidFlag, envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(Equal(fmt.Sprintf(ExpectedUnknownFlag, NotValidFlag)))
		})

		It("Root nothing else - help with error", func() {
			output, err := cli.RunCommand("", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(HelpCommandSuffix))
		})

		It("root with help - help with no error", func() {
			output, err := cli.RunCommand("-h", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HaveSuffix(HelpCommandSuffix))
		})
	})
})
