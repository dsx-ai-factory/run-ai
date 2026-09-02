/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Config Test")
}

const (
	TestFilePath = "config"
)

var _ = Describe("Config Test", Ordered, func() {
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
	AfterAll(func() {
		_ = os.RemoveAll(configPath)
	})

	Context("Login command validation", Ordered, func() {
		It("generate config file - success", func() {
			newFile := "new_config.json"
			configFile := filepath.Join(configPath, newFile)
			output, err := cli.RunCommand(fmt.Sprintf("config generate --file=%s", configFile), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal("Config file was generated successfully\n"))
		})

		It("generate config file - fail", func() {
			configFile := filepath.Join(configPath, cli.ConfigJson)
			output, err := cli.RunCommand(fmt.Sprintf("config generate --file=%s", configFile), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HavePrefix("Error: Config generation failed"))
			Expect(output).To(HaveSuffix(fmt.Sprintf(cli.ErrorConfigFileAlreadyExist, TestFilePath)))
		})

		It("config set - successfully", func() {
			instance := config.GetInstance()
			test := instance.GetConfiguration()
			Expect(test.Api.ClusterStatusTimeoutDuration).To(Equal(""))

			output, err := cli.RunCommand("config set --status-timeout-duration 400ms", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal("Config set successfully\n"))

			test = instance.GetConfiguration()
			Expect(test.Api.ClusterStatusTimeoutDuration).To(Equal("400ms"))
		})

		It("config set - failed", func() {
			instance := config.GetInstance()
			test := instance.GetConfiguration()
			Expect(test.Api.ClusterStatusTimeoutDuration).To(Equal("400ms"))

			_, err := cli.RunCommand("config set --status-timeout-duration 5", envMap)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("time: missing unit in duration \"5\""))

			test = instance.GetConfiguration()
			Expect(test.Api.ClusterStatusTimeoutDuration).To(Equal("400ms"))
		})

		It("config set auto-update - enable successfully", func() {
			instance := config.GetInstance()
			test := instance.GetConfiguration()
			Expect(test.Global.Update.Auto).To(BeFalse()) // Default should be false

			output, err := cli.RunCommand("config set --auto-update enabled", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal("Config set successfully\n"))

			test = instance.GetConfiguration()
			Expect(test.Global.Update.Auto).To(BeTrue())
		})

		It("config set auto-update - disable successfully", func() {
			instance := config.GetInstance()
			test := instance.GetConfiguration()
			Expect(test.Global.Update.Auto).To(BeTrue()) // From previous test

			output, err := cli.RunCommand("config set --auto-update disabled", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal("Config set successfully\n"))

			test = instance.GetConfiguration()
			Expect(test.Global.Update.Auto).To(BeFalse())
		})

		It("config set auto-update - invalid value", func() {
			instance := config.GetInstance()
			test := instance.GetConfiguration()
			initialValue := test.Global.Update.Auto

			_, err := cli.RunCommand("config set --auto-update invalid", envMap)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("invalid argument \"invalid\" for \"--auto-update\" flag: enable flag can be one of this values: enabled|disabled"))

			test = instance.GetConfiguration()
			Expect(test.Global.Update.Auto).To(Equal(initialValue)) // Should not change on error
		})
	})

	Context("Config set credential clearing behavior", Ordered, func() {
		var authPath string

		BeforeEach(func() {
			configPath, err = cli.GetHomePath()
			Expect(err).To(BeNil())
			configPath = cli.GetConfigPath(configPath, TestFilePath)
			authPath = filepath.Join(configPath, "authentication.json")

			err = os.MkdirAll(filepath.Dir(authPath), 0755)
			Expect(err).To(BeNil())

			dummyAuth := `{"accessToken": "dummy-token", "expiration": "2024-12-31T23:59:59Z"}`
			err = os.WriteFile(authPath, []byte(dummyAuth), 0644)
			Expect(err).To(BeNil())
			_, err = os.Stat(authPath)
			Expect(err).To(BeNil(), "Authentication file should exist before test")

		})

		It("config set cp-url - should clear credentials when authentication file exists", func() {
			output, err := cli.RunCommand("config set --cp-url http://new-control-plane.example.com/", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal("Config set successfully\n"))

			// Verify authentication file was removed
			_, err = os.Stat(authPath)
			Expect(os.IsNotExist(err)).To(BeTrue(), "Authentication file should be deleted when setting cp-url")
		})

		It("config set auth-url - should clear credentials when authentication file exists", func() {
			output, err := cli.RunCommand("config set --auth-url http://new-auth.example.com/", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal("Config set successfully\n"))

			// Verify authentication file was removed
			_, err = os.Stat(authPath)
			Expect(os.IsNotExist(err)).To(BeTrue(), "Authentication file should be deleted when setting auth-url")
		})

		It("config set non-auth parameter - should NOT clear credentials", func() {
			output, err := cli.RunCommand("config set --status-timeout-duration 600ms", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal("Config set successfully\n"))

			_, err = os.Stat(authPath)
			Expect(err).To(BeNil(), "Authentication file should still exist when setting non-auth parameters")

			os.Remove(authPath)
		})

		It("config set both cp-url and other parameters - should clear credentials", func() {
			output, err := cli.RunCommand("config set --cp-url http://new-cp.example.com/ --status-timeout-duration 500ms", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal("Config set successfully\n"))

			_, err = os.Stat(authPath)
			Expect(os.IsNotExist(err)).To(BeTrue(), "Authentication file should be deleted when setting cp-url even with other parameters")
		})
	})

})
