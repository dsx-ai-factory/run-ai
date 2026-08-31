/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package logs

import (
	"os"
	"testing"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/clients"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI logs Tests")
}

const (
	TestFilePath                            = "logs"
	ExpectedInvalidProject                  = "Set a default project by running 'runai project set <project-name>' or use flag -p <project-name>"
	ExpectedInvalidWorkloadTypeError        = "workload type is invalid must be one of: training, workspace\n"
	ExpectedInvalidSinceTimeFormatError     = "--sinceTime must be in RFC3339 format (e.g. 2025-05-25T14:30:00Z).\n"
	ExpectedContainCheckClusterAvailability = "please check cluster availability or change the default timeout duration with the"
	ExpectedSuffixContentDeadlineExceeded   = "context deadline exceeded\n"
)

var _ = Describe("Logs Test", Ordered, func() {
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
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)

		clients.MockWorkload(mockCtrl, []string{})
		clients.MockClusterStatusTimedOut(mockCtrl)
	})

	Context("Pre logs command validation", Ordered, func() {
		It("login validation", func() {
			err := clients.AuthValidation(configPath, cli.ConfigJson, envMap)
			if err != nil {
				Fail("authentication failed, error: " + err.Error())
			}
		})

		It("cluster set", func() {
			err := clients.ClusterSet(envMap)
			if err != nil {
				Fail("authentication failed, error: " + err.Error())
			}
		})
	})

	Context("logs command validation with invalid options", Ordered, func() {
		It("logs with invalid project", func() {
			output, err := cli.RunCommand("logs --type=workspace workspace-01", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(ExpectedInvalidProject))
		})
		It("logs with invalid workload type", func() {
			output, err := cli.RunCommand("logs workspace-01 --project test --type=invalid", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedInvalidWorkloadTypeError))
		})
		It("logs with invalid --since-time format", func() {
			output, err := cli.RunCommand("logs workspace-01 --type=workspace --since-time=invalid-time -p test", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedInvalidSinceTimeFormatError))
		})
		It("logs with cluster status timeout", func() {
			output, err := cli.RunCommand("workspace logs workspace-01 -p test", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(ExpectedContainCheckClusterAvailability))
			Expect(output).To(HaveSuffix(ExpectedSuffixContentDeadlineExceeded))
		})
	})

})
