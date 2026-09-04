/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mycredential

import (
	"context"
	"os"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	createDescriptionTestPath = "mycredential_create_description"
	acceptsOneArgReceived     = "accepts 1 arg(s), received"
)

var envMap map[string]string

func createCmdArgs(descriptionValue string) []string {
	return []string{
		"my-credential", "create", "secret1",
		"--type", "genericSecret",
		"--item", "key=k,value=v",
		"--description", descriptionValue,
	}
}

var _ = Describe("create --description flag", Ordered, func() {
	BeforeAll(func() {
		homePath, err := cli.GetHomePath()
		Expect(err).ToNot(HaveOccurred())
		configPath := cli.GetConfigPath(homePath, createDescriptionTestPath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)
	})

	It("user can set single word description without \"\"", func() {
		args := createCmdArgs("singleword")
		output, err := cli.RunCommandStringSplit(context.Background(), args, envMap)
		Expect(output).ToNot(ContainSubstring(acceptsOneArgReceived), "single-word --description without quotes must not produce 'accepts 1 arg(s), received'")
		_ = err
	})

	It("user can set single word description with \"\"", func() {
		args := createCmdArgs("singleword")
		output, err := cli.RunCommandStringSplit(context.Background(), args, envMap)
		Expect(output).ToNot(ContainSubstring(acceptsOneArgReceived), "single-word --description (with or without quotes) must not produce 'accepts 1 arg(s), received'")
		_ = err
	})

	It("multiple words description without \"\" returns error", func() {
		cmd := "my-credential create secret1 --type genericSecret --item key=k,value=v --description hello world"
		_, err := cli.RunCommand(cmd, envMap)
		Expect(err).ToNot(BeNil())
		Expect(err.Error()).To(ContainSubstring(acceptsOneArgReceived))
		Expect(err.Error()).To(ContainSubstring("received 2"))
	})

	It("user can set multiple words description with \"\"", func() {
		args := createCmdArgs("hello world")
		output, err := cli.RunCommandStringSplit(context.Background(), args, envMap)
		Expect(output).ToNot(ContainSubstring(acceptsOneArgReceived), "multi-word --description in one token (as with quoted string) must not produce 'accepts 1 arg(s), received'")
		_ = err
	})
})
