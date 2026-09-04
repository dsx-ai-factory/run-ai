/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package application_test

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth/fake"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/factories"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

const ApplicationDescribeTestFilePath = "application-describe"

var _ = Describe("Application Describe Test", Ordered, func() {
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
		configPath = cli.GetConfigPath(configPath, ApplicationDescribeTestFilePath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)

		mockCtrl := gomock.NewController(GinkgoT())
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)
		clients.MockAIApplication(mockCtrl)
	})

	BeforeEach(func() {
		_, err = cli.RunCommand(fmt.Sprintf("config set --cp-url=%s", cli.ClusterOneDomain), envMap)
		if err != nil {
			Fail("failed to set cp-url, error: " + err.Error())
		}
		err = clients.ClusterSet(envMap)
		if err != nil {
			Fail("cluster set failed, error: " + err.Error())
		}
	})

	Context("Pre command validation", Ordered, func() {
		It("login validation", func() {
			err := clients.AuthValidation(configPath, cli.ConfigJson, envMap)
			if err != nil {
				Fail("authentication failed, error: " + err.Error())
			}
		})
	})

	Context("application describe validation", Ordered, func() {
		It("describe shows the application name and project in the GENERAL section", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application describe %s -p %s", clients.AppName01, clients.ProjectOneName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring("GENERAL"))
			Expect(output).To(ContainSubstring(clients.AppName01))
			Expect(output).To(ContainSubstring(clients.ProjectOneName))
		})

		It("describe of a detailed application shows status, message, chart URL, and resources", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application describe %s -p %s", clients.AppNameDetailed, clients.ProjectOneName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring("GENERAL"))
			Expect(output).To(ContainSubstring(clients.AppNameDetailed))
			Expect(output).To(ContainSubstring("deployed"))
			Expect(output).To(ContainSubstring("install completed"))
			Expect(output).To(ContainSubstring("CHART REFERENCE"))
			Expect(output).To(ContainSubstring("charts.example.com"))
			Expect(output).To(ContainSubstring("RESOURCES (1)"))
			Expect(output).To(ContainSubstring("grafana-server"))
		})

		It("describe with -o json produces valid JSON with correct fields", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application describe %s -p %s -o json", clients.AppName01, clients.ProjectOneName), envMap)
			Expect(err).To(BeNil())
			var decoded map[string]interface{}
			Expect(json.Unmarshal([]byte(output), &decoded)).To(Succeed())
			Expect(decoded["name"]).To(Equal(clients.AppName01))
			Expect(decoded["type"]).To(Equal("inference"))
		})

		It("describe with -o yaml produces valid YAML with correct fields", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application describe %s -p %s -o yaml", clients.AppName01, clients.ProjectOneName), envMap)
			Expect(err).To(BeNil())
			var decoded map[string]interface{}
			Expect(yaml.Unmarshal([]byte(output), &decoded)).To(Succeed())
			Expect(decoded["name"]).To(Equal(clients.AppName01))
			Expect(decoded["type"]).To(Equal("inference"))
		})

		It("describe of a missing application fails with a not-found error", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application describe %s -p %s", clients.AppNameMissing, clients.ProjectOneName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(
				fmt.Sprintf("application %q not found in project %q", clients.AppNameMissing, clients.ProjectOneName)))
		})

		It("describe without a project fails with a clear message", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application describe %s", clients.AppName01), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring("project"))
		})

		It("describe with an unknown project fails with a project error, not an app-not-found error", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application describe %s -p p300", clients.AppName01), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).NotTo(ContainSubstring("not found in project"))
		})

		It("describe without an application name fails", func() {
			_, err := cli.RunCommand("application describe", envMap)
			Expect(err).ToNot(BeNil())
		})

		It("describe with an unsupported output format fails with a clear error", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application describe %s -p %s -o xml", clients.AppName01, clients.ProjectOneName), envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring("unsupported output format"))
		})

		It("describe shows REQUESTED RESOURCES section with GPU and CPU memory values", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application describe %s -p %s", clients.AppNameWithResources, clients.ProjectOneName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring("REQUESTED RESOURCES"))
			Expect(output).To(ContainSubstring("GPU:"))
			Expect(output).To(ContainSubstring("2"))
			Expect(output).To(ContainSubstring("CPU Memory:"))
			Expect(output).To(ContainSubstring("4Gi"))
		})

		It("describe shows ALLOCATED RESOURCES section with GPU and GPU memory values", func() {
			output, err := cli.RunCommand(
				fmt.Sprintf("application describe %s -p %s", clients.AppNameWithResources, clients.ProjectOneName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring("ALLOCATED RESOURCES"))
			Expect(output).To(ContainSubstring("GPU:"))
			Expect(output).To(ContainSubstring("1.5"))
			Expect(output).To(ContainSubstring("GPU Memory:"))
			Expect(output).To(ContainSubstring("3Gi"))
		})

		It("application name completion returns all applications in the project", func() {
			output, err := cli.RunCommand(
				factories.ConvertToCompletionTestCmd(
					fmt.Sprintf("application describe -p %s", clients.ProjectOneName), ""), envMap)
			Expect(err).To(BeNil())
			factories.AssertCompletionOptions(output, []string{clients.AppName01, clients.AppName02})
			factories.AssertCompletionResponseStatus(output, false)
		})
	})
})
