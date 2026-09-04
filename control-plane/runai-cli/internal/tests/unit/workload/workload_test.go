/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/workload"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth/fake"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Workload Tests")
}

const (
	TestFilePath                                    = "workload"
	ExpectedWorkloadListWithProjectThatDoesNotExist = "project name: p300, not found\n"
)

var _ = Describe("Workload Test", Ordered, func() {
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

		clients.MockWorkloadByList(mockCtrl, workload.FilterList)
		clients.MockWorkloadBatchDelete(mockCtrl, workload.BatchList)
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

	Context("Workspace command list validation", Ordered, func() {
		It("workload list all", func() {
			output, err := cli.RunCommand("workload list -A", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(clients.WorkloadName01))
			Expect(output).To(ContainSubstring(clients.WorkloadName02))
		})

		It("Workload list all and project flags", func() {
			_, err := cli.RunCommand("workload list -A --project=p301", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("[all project] are set none of the others can be"))
		})

		It("workload list project p301", func() {
			output, err := cli.RunCommand("workload list --project=p301", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(clients.WorkloadName01))
			Expect(output).ToNot(ContainSubstring(clients.WorkloadName02))
		})

		It("workload list project p300", func() {
			output, err := cli.RunCommand("workload list --project=p300", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedWorkloadListWithProjectThatDoesNotExist))
		})

		Describe("Workload list project flags", func() {
			BeforeAll(func() {
				err := clients.ProjectSet(clients.ProjectOneName, envMap)
				if err != nil {
					Fail("project set failed, error: " + err.Error())
				}
			})

			AfterAll(func() {
				// workarounds for the fact that the project is not cleared after each test
				err := clients.ClusterSet(envMap)
				if err != nil {
					Fail("cluster set failed, error: " + err.Error())
				}
			})

			It("Training list project with default project", func() {
				output, err := cli.RunCommand("workload list", envMap)
				Expect(err).To(BeNil())
				Expect(output).To(ContainSubstring(clients.WorkloadName01))
				Expect(output).ToNot(ContainSubstring(clients.WorkloadName02))

			})

			It("Training list all projects with default project set", func() {
				output, err := cli.RunCommand("workload list --all", envMap)
				Expect(err).To(BeNil())
				Expect(output).To(ContainSubstring(clients.WorkloadName01))
				Expect(output).To(ContainSubstring(clients.WorkloadName02))
			})
		})
	})

	Context("Workspace command delete validation", Ordered, func() {
		It("workload delete workload-01 workload-02 - success", func() {
			output, err := cli.RunCommand("workload delete -y -p p301 workload-01 workload-02", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring("Successfully deleted all workloads"))
		})

		It("workload delete workload-01 workload-03 - failed on 03", func() {
			output, err := cli.RunCommand("workload delete -y -p p301 workload-01 workload-03", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring("At least one workload can't be deleted. You're either not authorized to delete them or they are not managed by Run:ai, delete them the way they were created."))
			Expect(output).To(ContainSubstring("You are not permitted"))
		})

		It("workload delete workload-04 - error", func() {
			output, err := cli.RunCommand("workload delete -y -p p301 workload-04", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring("Bulk delete failed to execute. big problem"))
		})

		It("workload delete list too long - error", func() {
			output, err := cli.RunCommand("workload delete -y -p p301 ws1 ws2 ws3 ws4 ws5 ws6 ws7 ws8 ws9 ws10 "+
				"ws11 ws12 ws13 ws14 ws15 ws16 ws17 ws18 ws19 ws20 ws21 ws22 ws23 ws24 ws25 ws26 ws27 ws28 ws29 ws30 "+
				"ws31 ws32 ws33 ws34 ws35 ws36 ws37 ws38 ws39 ws40 ws41 ws42 ws43 ws44 ws45 ws46 ws47 ws48 ws49 ws50 "+
				"ws51 ws52 ws53 ws54 ws55 ws56 ws57 ws58 ws59 ws60 ws61 ws62 ws63 ws64 ws65 ws66 ws67 ws68 ws69 ws70 "+
				"ws71 ws72 ws73 ws74 ws75 ws76 ws77 ws78 ws79 ws80 ws81 ws82 ws83 ws84 ws85 ws86 ws87 ws88 ws89 ws90 "+
				"ws91 ws92 ws93 ws94 ws95 ws96 ws97 ws98 ws99 ws100 ws101", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(ContainSubstring("too many workloads to delete. limit: 100, actual: 101"))
		})

	})
})
