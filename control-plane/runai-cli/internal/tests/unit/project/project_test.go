/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	"fmt"
	"os"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/project"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth/fake"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	TestFilePath = "project"

	ExpectedErrorFirstProjectListCmd           = "Error: Cluster URL is missing. parse \"\": empty url\nSet it by running 'runai cluster set <CLUSTER_NAME>'"
	ExpectedProjectSetSuccessfully             = "Project p301 set successfully as the default project\n"
	ExpectedErrorProjectListDoesNotExistSuffix = "project name: p111, not found\n"
	ExpectedOutputCreatingProject              = "Creating project p301 ..."
	ExpectedErrorInvalidArgument               = "invalid argument "
	ExpectedErrorInvalidGpuValues              = "values: gpu-limit should be higher or equal to gpu-deserved"
)

func TestProject(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI project Tests")
}

var _ = Describe("Project Test", Ordered, func() {
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
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)
	})

	BeforeEach(func() {
		_, err = cli.RunCommand(fmt.Sprintf("config set --cp-url=%s", cli.ClusterOneDomain), envMap)
		if err != nil {
			Fail("failed to set cp-url, error: " + err.Error())
		}
	})

	Context("Project command validation", Ordered, func() {
		It("project list - need to set cluster", func() {
			output, err := cli.RunCommand("project list", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(ExpectedErrorFirstProjectListCmd))
		})

		Context("project list - cluster set", func() {
			BeforeEach(func() {
				err = clients.ClusterSet(envMap)
				if err != nil {
					Fail("cluster set failed, error: " + err.Error())
				}
			})
			It("project list - print table", func() {
				output, err := cli.RunCommand("project list", envMap)
				Expect(err).To(BeNil())
				Expect(output).To(ContainSubstring("p301"))
			})

			It("project set - p301", func() {
				output, err := cli.RunCommand("project set p301", envMap)
				Expect(err).To(BeNil())
				Expect(output).To(Equal(ExpectedProjectSetSuccessfully))
			})

			It("project set none exist project - error", func() {
				output, err := cli.RunCommand("project set p111", envMap)
				Expect(err).ToNot(BeNil())
				Expect(output).To(HaveSuffix(ExpectedErrorProjectListDoesNotExistSuffix))
			})

			It("project set - p301 - quiet mode", func() {
				output, err := cli.RunCommand("project set p301 -q", envMap)
				Expect(err).To(BeNil())
				Expect(output).To(BeEmpty())
			})

			It("project set none exist project - error - quiet mode", func() {
				output, err := cli.RunCommand("project set p111 -q", envMap)
				Expect(err).ToNot(BeNil())
				Expect(output).To(HaveSuffix(ExpectedErrorProjectListDoesNotExistSuffix))
			})

			It("project create p301 with defaults", func() {
				output, err := cli.RunCommand(fmt.Sprintf("project create %s --resources nodepool=default", clients.ProjectOneName), envMap)
				Expect(err).To(BeNil())
				Expect(output).To(HavePrefix(ExpectedOutputCreatingProject))
			})

			It("project create p301 with gpu data", func() {
				output, err := cli.RunCommand(fmt.Sprintf("project create %s --resources nodepool=default,gpu-limit=2,gpu-deserved=1,priority=Low", clients.ProjectOneName), envMap)
				Expect(err).To(BeNil())
				Expect(output).To(HavePrefix(ExpectedOutputCreatingProject))
			})

			It("project create p301 with several nodepools", func() {
				output, err := cli.RunCommand(fmt.Sprintf("project create %s "+
					"--resources nodepool=%s,gpu-limit=2,gpu-deserved=1,priority=Low "+
					"--resources nodepool=%s,gpu-limit=2,gpu-deserved=1,priority=High", clients.ProjectOneName, clients.DefaultNodepool, clients.AnotherPoolName), envMap)
				Expect(err).To(BeNil())
				Expect(output).To(HavePrefix(ExpectedOutputCreatingProject))
			})

			It("project create p301 with gpu data", func() {
				output, err := cli.RunCommand(fmt.Sprintf("project create %s --resources nodepool=default,gpu-limit=2,gpu-deserved=1,priority=Low", clients.ProjectOneName), envMap)
				Expect(err).To(BeNil())
				Expect(output).To(HavePrefix(ExpectedOutputCreatingProject))
			})

			It("project create p301 with duplicated nodepools", func() {
				output, err := cli.RunCommand(fmt.Sprintf("project create %s "+
					"--resources nodepool=%s,gpu-limit=2,gpu-deserved=1,priority=Low "+
					"--resources nodepool=%s,gpu-limit=2,gpu-deserved=1,priority=High", clients.ProjectOneName, clients.AnotherPoolName, clients.AnotherPoolName), envMap)
				Expect(err).To(HaveOccurred())
				Expect(output).To(HavePrefix(fmt.Sprintf("Error: %s", project.FailedToCreateProject)))
			})

			It("project create p301 with invalid GPU arguments", func() {
				_, err := cli.RunCommand(fmt.Sprintf("project create %s --resources nodepool=default,gpu-limit=2,gpu-deserved=4", clients.ProjectOneName), envMap)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(HavePrefix(ExpectedErrorInvalidArgument))
				Expect(err.Error()).To(HaveSuffix(ExpectedErrorInvalidGpuValues))
			})

			It("project create p301 with department", func() {
				output, err := cli.RunCommand(fmt.Sprintf("project create %s --resources nodepool=default,gpu-limit=12,gpu-deserved=4 --department %s", clients.ProjectOneName, clients.AnotherDepartment), envMap)
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(HavePrefix(ExpectedOutputCreatingProject))
			})
		})

	})
})
