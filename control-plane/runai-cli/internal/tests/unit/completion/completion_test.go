/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package completion

import (
	"fmt"
	"os"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth/fake"

	"k8s.io/utils/ptr"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/clients"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/tests/unit/cli/factories"
)

const (
	TestFilePath              = "completion"
	FlagCompletionCmd         = "exec -p"
	ArgCompletionCmdProject   = "project set"
	ArgCompletionCmdWorkspace = "workspace describe"
	ArgCompletionCmdWorkload  = "workload describe --type training"

	PrioritiesCmdComplete = "workspace submit --priority"

	WorkloadName01 = "workload-01"
	WorkloadName02 = "workload-02"
)

var _ = Describe("Completion Tests", Ordered, func() {
	var (
		configPath string
		err        error
		envMap     map[string]string

		mockCtrl *gomock.Controller
	)

	BeforeAll(func() {
		configPath, err = cli.GetHomePath()
		if err != nil {
			Fail("failed to get pwd")
		}
		configPath = cli.GetConfigPath(configPath, TestFilePath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)

		mockCtrl = gomock.NewController(GinkgoT())
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)

		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)
		clients.MockWorkloadPriorities(mockCtrl)
	})

	Context("completion error", Ordered, func() {
		It("mandatory - login validation", func() {
			err := clients.AuthValidation(configPath, cli.ConfigJson, envMap)
			if err != nil {
				Fail("authentication failed, error: " + err.Error())
			}
		})

		It("arg", func() {
			output, _ := cli.RunCommand(factories.ConvertToCompletionTestCmd(ArgCompletionCmdProject, ""), envMap)
			factories.AssertCompletionResponseStatus(output, true)
		})

		It("flag", func() {
			output, _ := cli.RunCommand(factories.ConvertToCompletionTestCmd(FlagCompletionCmd, ""), envMap)
			factories.AssertCompletionResponseStatus(output, true)
		})
	})

	Context("completion success", Ordered, func() {
		It("mandatory - cluster set", func() {
			err := clients.ClusterSet(envMap)
			if err != nil {
				Fail("authentication failed, error: " + err.Error())
			}
		})

		Context("flag", Ordered, func() {
			It("no prefix", func() {
				output, err := cli.RunCommand(factories.ConvertToCompletionTestCmd(FlagCompletionCmd, ""), envMap)
				Expect(err).To(BeNil())
				factories.AssertCompletionOptions(output, clients.ProjectNames)
				factories.AssertCompletionResponseStatus(output, false)
			})

			It("given prefix exists", func() {
				prefix := clients.ProjectNames[0][:1]
				output, err := cli.RunCommand(factories.ConvertToCompletionTestCmd(FlagCompletionCmd, prefix), envMap)
				Expect(err).To(BeNil())
				factories.AssertCompletionOptions(output, clients.ProjectNames)
				factories.AssertCompletionResponseStatus(output, false)
			})

			It("given prefix doesn't exist", func() {
				output, err := cli.RunCommand(factories.ConvertToCompletionTestCmd(FlagCompletionCmd, "bla"), envMap)
				Expect(err).To(BeNil())
				factories.AssertCompletionOptions(output, nil)
				factories.AssertCompletionResponseStatus(output, false)
			})

			It("Priority flag", func() {
				output, err := cli.RunCommand(factories.ConvertToCompletionTestCmd(PrioritiesCmdComplete, ""), envMap)
				Expect(err).To(BeNil())
				factories.AssertCompletionOptions(output, clients.ClusterPrioritiesNames)
				factories.AssertCompletionResponseStatus(output, false)
			})

			It("Priority flag - prefix does not exist", func() {
				output, err := cli.RunCommand(factories.ConvertToCompletionTestCmd(PrioritiesCmdComplete, "bla"), envMap)
				Expect(err).To(BeNil())
				factories.AssertCompletionOptions(output, nil)
				factories.AssertCompletionResponseStatus(output, false)
			})

			It("Priority flag - prefix does exist", func() {
				output, err := cli.RunCommand(factories.ConvertToCompletionTestCmd(PrioritiesCmdComplete, "lo"), envMap)
				Expect(err).To(BeNil())
				factories.AssertCompletionOptions(output, []string{"low"})
				factories.AssertCompletionResponseStatus(output, false)
			})
		})

		DescribeTableSubtree("arg", Ordered,
			// mockInitFn - for workload cmd tests, should set the correct project in ctx and init the mock client with the correct filters
			func(cmdStr string, options []string, mockInitFn func(*gomock.Controller, map[string]string)) {
				It("completes when has no arg", func() {
					mockInitFn(mockCtrl, envMap)
					output, err := cli.RunCommand(factories.ConvertToCompletionTestCmd(cmdStr, ""), envMap)
					Expect(err).To(BeNil())
					factories.AssertCompletionOptions(output, options)
					factories.AssertCompletionResponseStatus(output, false)
				})

				It("don't complete when has arg", func() {
					mockInitFn(mockCtrl, envMap)
					output, err := cli.RunCommand(factories.ConvertToCompletionTestCmd(fmt.Sprintf("%s arg ", cmdStr), ""), envMap)
					Expect(err).To(BeNil())
					factories.AssertCompletionOptions(output, nil)
					factories.AssertCompletionResponseStatus(output, false)
				})

				It("given prefix exists", func() {
					mockInitFn(mockCtrl, envMap)
					prefix := options[0][:1]
					output, err := cli.RunCommand(factories.ConvertToCompletionTestCmd(cmdStr, prefix), envMap)
					Expect(err).To(BeNil())
					factories.AssertCompletionOptions(output, options)
					factories.AssertCompletionResponseStatus(output, false)
				})

				It("given prefix doesn't exist", func() {
					mockInitFn(mockCtrl, envMap)
					output, err := cli.RunCommand(factories.ConvertToCompletionTestCmd(cmdStr, "bla"), envMap)
					Expect(err).To(BeNil())
					factories.AssertCompletionOptions(output, nil)
					factories.AssertCompletionResponseStatus(output, false)
				})
			},
			Entry("project", ArgCompletionCmdProject, clients.ProjectNames, func(_ *gomock.Controller, _ map[string]string) {}),
			Entry("workload", ArgCompletionCmdWorkload, []string{clients.TrainingWorkloadName}, initTrainingMock),
			Entry("workspace", ArgCompletionCmdWorkspace, []string{clients.WorkspaceWorkloadName}, initWorkspaceMock),
		)

		DescribeTable("list options override - '-p' flag", Ordered,
			func(cmdStr string, options []string, mockInitFn func(*gomock.Controller, map[string]string)) {
				mockInitFn(mockCtrl, envMap)
				output, _ := cli.RunCommand(factories.ConvertToCompletionTestCmd(fmt.Sprintf("%s -p %s", cmdStr, "projectblabla"), ""), envMap)
				// non-existing project from flag was picked instead of project from context and raises error from list command
				factories.AssertCompletionResponseStatus(output, true)
			},
			Entry("workload", ArgCompletionCmdWorkload, []string{clients.TrainingWorkloadName}, initTrainingMock),
			Entry("workspace", ArgCompletionCmdWorkspace, []string{clients.WorkspaceWorkloadName}, initWorkspaceMock),
		)
	})
})

func initTrainingMock(mockCtrl *gomock.Controller, envMap map[string]string) {
	Expect(clients.ProjectSet(clients.TrainingProjectName, envMap)).To(Succeed())
	workloadGenerator := clients.WorkloadGenerator{}
	workloadGenerator.
		SetTenant(1).
		SetType(api.Training).
		SetPods(0, ptr.To[int32](4), nil, nil).
		SetClusterID(cli.ClusterOneUuid)

	mockPaginator := clients.NewWorkloadPaginationMocker(mockCtrl)
	mockPaginator.SetAny(filter.NewBuilder().
		WithClusterID(cli.ClusterOneUuid).
		WithProject(clients.ProjectNameP302).
		WithWorkloadType("Training||Distributed"),
		int32(11), int32(0), nil, nil,
		workloadGenerator.SetNameAndUUID(WorkloadName01).Build(),
		workloadGenerator.SetNameAndUUID(WorkloadName02).Build())
}

func initWorkspaceMock(mockCtrl *gomock.Controller, envMap map[string]string) {
	Expect(clients.ProjectSet(clients.WorkspaceProjectName, envMap)).To(Succeed())
	workloadGenerator := clients.WorkloadGenerator{}
	workloadGenerator.
		SetTenant(1).
		SetType(api.Workspace).
		SetPods(0, ptr.To[int32](4), nil, nil).
		SetClusterID(cli.ClusterOneUuid)

	mockPaginator := clients.NewWorkloadPaginationMocker(mockCtrl)
	mockPaginator.SetAny(filter.NewBuilder().
		WithClusterID(cli.ClusterOneUuid).
		WithProject(clients.ProjectNameP301).
		WithWorkloadType(api.Workspace.QueryParam()).
		WithWorkloadFramework(api.Runai),
		int32(11), int32(0), nil, nil,
		workloadGenerator.SetNameAndUUID(WorkloadName01).Build(),
		workloadGenerator.SetNameAndUUID(WorkloadName02).Build())
}
