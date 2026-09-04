/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package login

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Login Test")
}

const (
	TestFilePath = "login"
	AppName      = "app-lies"
	AppSecret    = "shhhhh"

	ExpectedOutputMissingAuthURL                = "Error: Authentication failed. Authentication URL is missing. parse \"\": empty url\nSet it by running 'runai config set --auth-url <auth_url>'\n"
	ExpectedOutputMissingAuthURLWithDeprecation = "Command \"application\" is deprecated, use 'runai login access-key' instead\nError: Authentication failed. Authentication URL is missing. parse \"\": empty url\nSet it by running 'runai config set --auth-url <auth_url>'\n"
)

var _ = Describe("Login Test", Ordered, func() {
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

		mockCtrl := gomock.NewController(GinkgoT())
		clients.MockAuthentication(mockCtrl)
		clients.MockSingleCluster(mockCtrl)
	})

	Context("Login command validation", Ordered, func() {
		It("login fail, empty config file - missing URL", func() {
			output, err := cli.RunCommand("login app", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(Equal(ExpectedOutputMissingAuthURLWithDeprecation))
		})

		It("generate config file - fail", func() {
			configFile := filepath.Join(configPath, cli.ConfigJson)
			output, err := cli.RunCommand(fmt.Sprintf("config generate --file=%s", configFile), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HavePrefix("Error: Config generation failed"))
			Expect(output).To(HaveSuffix(fmt.Sprintf(cli.ErrorConfigFileAlreadyExist, TestFilePath)))
		})

		It("login fail, config exist, but auth URL is missing", func() {
			output, err := cli.RunCommand("login app", envMap)
			Expect(err.Error()).ToNot(BeNil())
			Expect(output).To(Equal(ExpectedOutputMissingAuthURLWithDeprecation))
		})

		It("config set", func() {
			output, err := cli.RunCommand("config set --auth-url=http://localhost.runailabs.com/ --cp-url=http://localhost.runailabs.com/", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal("Config set successfully\n"))
			instance := config.GetInstance()
			test := instance.GetConfiguration()
			Expect(test.Cluster.Name).To(Equal(""))
			Expect(test.Cluster.Uuid).To(Equal(""))
			Expect(test.Cluster.Domain).To(Equal(""))
		})

		It("login via application authentication", func() {
			output, err := cli.RunCommand(fmt.Sprintf("login app --client-id=%s --secret=%s --interactive=disabled", AppName, AppSecret), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring("Authentication was successful"))
			Expect(output).To(ContainSubstring("Cluster name-cluster-1 was automatically defined successfully"))

			instance := config.GetInstance()
			test := instance.GetConfiguration()
			Expect(test.Credentials.Name).To(Equal(""))
			Expect(test.Credentials.Secret).To(Equal(""))
			Expect(test.Credentials.Type).To(Equal(""))
			Expect(test.Global.Update.Auto).To(BeFalse())
		})

		It("set application login credentials", func() {
			output, err := cli.RunCommand(fmt.Sprintf("login app --client-id=%s --secret=%s --interactive=disabled --save", AppName, AppSecret), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring("Authentication was successful"))

			instance := config.GetInstance()
			test := instance.GetConfiguration()
			Expect(test.Credentials.Name).To(Equal(AppName))
			Expect(test.Credentials.Secret).To(Equal(AppSecret))
			Expect(test.Credentials.Type).To(Equal("AccessKeyIDLogin"))
			Expect(test.Cluster.Name).To(Equal("name-cluster-1"))
			Expect(test.Cluster.Uuid).To(Equal("uuid-cluster-1"))
			Expect(test.Cluster.Domain).To(Equal("http://localhost.runailabs.com/"))
		})

		It("set application login credentials (using deprecated --name)", func() {
			output, err := cli.RunCommand(fmt.Sprintf("login app --name=%s --secret=%s --interactive=disabled --save", AppName, AppSecret), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring("Flag --name has been deprecated, use --client-id instead."))
			Expect(output).To(ContainSubstring("Authentication was successful"))

			instance := config.GetInstance()
			test := instance.GetConfiguration()
			Expect(test.Credentials.Name).To(Equal("app-lies"))
			Expect(test.Credentials.Secret).To(Equal("shhhhh"))
			Expect(test.Credentials.Type).To(Equal("AccessKeyIDLogin"))
			Expect(test.Global.Update.Auto).To(BeFalse())
		})

		It("reset config - should clean the cluster config", func() {
			output, err := cli.RunCommand("config set --cp-url=http://localhost.runailabs.com/", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal("Config set successfully\n"))
			instance := config.GetInstance()
			test := instance.GetConfiguration()
			Expect(test.Cluster.Name).To(Equal(""))
			Expect(test.Cluster.Uuid).To(Equal(""))
			Expect(test.Cluster.Domain).To(Equal(""))
		})
	})

	Context("Auto-update functionality", Ordered, func() {
		It("should show auto-update disabled message", func() {
			instance := config.GetInstance()
			test := instance.GetConfiguration()
			Expect(test.Global.Update.Auto).To(BeFalse())

			output, err := cli.RunCommand(fmt.Sprintf("login app --name=%s --secret=%s --interactive=disabled", AppName, AppSecret), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring("Authentication was successful"))
			Expect(output).To(ContainSubstring("Cluster name-cluster-1 was automatically defined successfully"))
		})
	})
})
