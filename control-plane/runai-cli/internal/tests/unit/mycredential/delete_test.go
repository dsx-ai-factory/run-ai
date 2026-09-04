/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mycredential

import (
	"context"
	"os"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth/fake"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	deleteTestPath     = "mycredential_delete"
	unsupportedType    = "unsupported credential type"
	requiredFlagNotSet = "not set"
	acceptsOneArg      = "accepts 1 arg(s), received 0"
)

var deleteEnvMap map[string]string

func deleteCmdArgs(name, typeVal string) []string {
	return []string{
		"my-credential", "delete", name,
		"--type", typeVal,
	}
}

var _ = Describe("my-credential delete", Ordered, func() {
	BeforeAll(func() {
		homePath, err := cli.GetHomePath()
		Expect(err).ToNot(HaveOccurred())
		configPath := cli.GetConfigPath(homePath, deleteTestPath)
		deleteEnvMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)

		mockCtrl := gomock.NewController(GinkgoT())
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		clients.MockAuthentication(mockCtrl)

		err = clients.AuthValidation(configPath, cli.ConfigJson, deleteEnvMap)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("validation", func() {
		It("returns error when --type flag is omitted", func() {
			args := []string{"my-credential", "delete", "some-name"}
			_, err := cli.RunCommandStringSplit(context.Background(), args, deleteEnvMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring(requiredFlagNotSet))
		})

		It("returns error when SECRET_NAME argument is omitted", func() {
			cmd := "my-credential delete --type genericSecret"
			_, err := cli.RunCommand(cmd, deleteEnvMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring(acceptsOneArg))
		})

		It("returns unsupported credential type when --type is invalid", func() {
			args := deleteCmdArgs("x", "invalidType")
			_, err := cli.RunCommandStringSplit(context.Background(), args, deleteEnvMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring(unsupportedType))
		})

		It("returns unsupported credential type when --type has wrong case", func() {
			args := deleteCmdArgs("x", "GenericSecret")
			_, err := cli.RunCommandStringSplit(context.Background(), args, deleteEnvMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring(unsupportedType))
		})

		It("returns error when SECRET_NAME is only whitespace", func() {
			args := deleteCmdArgs("   ", "genericSecret")
			_, err := cli.RunCommandStringSplit(context.Background(), args, deleteEnvMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(And(ContainSubstring("SECRET_NAME"), ContainSubstring("empty")))
		})
	})

	Describe("not found for each credential type", func() {
		const nonexistent = "no-such-credential"
		const noSuchHost = "no such host"

		It("displays not-found error for genericSecret when credential does not exist", func() {
			args := deleteCmdArgs(nonexistent, "genericSecret")
			_, err := cli.RunCommandStringSplit(context.Background(), args, deleteEnvMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Or(
				ContainSubstring("personal credential '"+nonexistent+"' of type genericSecret not found"),
				ContainSubstring(noSuchHost),
			))
		})

		It("displays not-found error for dockerRegistry when credential does not exist", func() {
			args := deleteCmdArgs(nonexistent, "dockerRegistry")
			_, err := cli.RunCommandStringSplit(context.Background(), args, deleteEnvMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Or(
				ContainSubstring("personal credential '"+nonexistent+"' of type dockerRegistry not found"),
				ContainSubstring(noSuchHost),
			))
		})

		It("displays not-found error for ngcApiKey when credential does not exist", func() {
			args := deleteCmdArgs(nonexistent, "ngcApiKey")
			_, err := cli.RunCommandStringSplit(context.Background(), args, deleteEnvMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Or(
				ContainSubstring("personal credential '"+nonexistent+"' of type ngcApiKey not found"),
				ContainSubstring(noSuchHost),
			))
		})
	})
})
