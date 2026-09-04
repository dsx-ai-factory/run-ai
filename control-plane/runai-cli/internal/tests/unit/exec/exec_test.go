/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package exec_test

import (
	"os"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI exec backward Tests")
}

const (
	TestFilePath                     = "exec"
	ExpectedInvalidProject           = "Set a default project by running 'runai project set <project-name>' or use flag -p <project-name>"
	ExpectedInvalidWorkloadTypeError = "workload type is invalid must be one of: training, workspace\n"
)

var _ = Describe("Exec Test", Ordered, func() {
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
	})

	Context("Pre exec command validation", Ordered, func() {
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

	Context("exec command validation with invalid options", Ordered, func() {
		It("exec with invalid project", func() {
			output, err := cli.RunCommand("exec --type=workspace workspace-01 -- /bin/bash", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring(ExpectedInvalidProject))
		})
		It("exec with invalid workload type", func() {
			output, err := cli.RunCommand("exec workspace-01 --type=invalid -p test -- /bin/bash", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedInvalidWorkloadTypeError))
		})
	})

})
