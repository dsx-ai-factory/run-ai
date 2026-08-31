/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package deprecated

import (
	"fmt"
	"os"
	"testing"

	submitcmd "github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/submit"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth/fake"
	"k8s.io/utils/ptr"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/clients"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Deprecated aliases Test")
}

const (
	TestFilePath            = "deprecated"
	clusterDeprecatedMsg    = "Command \"clusters\" is deprecated, please use: `runai cluster list` instead"
	projectsDeprecatedMsg   = "Command \"projects\" is deprecated, please use: `runai project list` instead"
	jobsDeprecatedMsg       = "Command \"jobs\" is deprecated, please use: `runai workload list` instead"
	listDeprecatedMsg       = "Command \"list\" is deprecated, please use list in the relevant context: `runai project list`, `runai workload list`, `runai node list`, `runai cluster list`"
	listDeprecatedHelpMsg   = "[Deprecated] display resource list. By default displays the job list"
	DepMasterEnvironmentMsg = "Flag --master-environment has been deprecated, use --master-environment-variable instead"
	DepEnvironmentMsg       = "The --environment flag with key=value format for defining environment variables is deprecated and will be removed in a future release. Use the --environment-variable flag instead."
)

const (
	WorkloadName01  = "workload-01"
	WorkloadName02  = "workload-02"
	ProjectNameP301 = "p301"
)

var _ = Describe("Deprecated aliases Test", Ordered, func() {
	var (
		configPath        string
		err               error
		envMap            map[string]string
		mockPaginator     *clients.WorkloadPaginationMocker
		workloadGenerator clients.WorkloadGenerator
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
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		mockPaginator = clients.NewWorkloadPaginationMocker(mockCtrl)

		workloadGenerator.
			SetTenant(1).
			SetType(api.Workspace).
			SetPods(0, ptr.To[int32](4), nil, nil).
			SetClusterID(cli.ClusterOneUuid)
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

	Context("List subcommand validation", func() {
		// list cluster
		It("list clusters", func() {
			output, err := cli.RunCommand("list clusters", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(clusterDeprecatedMsg))
			Expect(output).To(ContainSubstring(cli.ClusterOneName))
			Expect(output).To(ContainSubstring(cli.ClusterOneDomain))
		})

		// list projects
		It("list projects", func() {
			output, err := cli.RunCommand(fmt.Sprintf("cluster set %s", cli.ClusterOneName), envMap)
			Expect(err).To(BeNil())
			Expect(output).To(Equal(fmt.Sprintf("Cluster %s set successfully\n", cli.ClusterOneName)))

			output, err = cli.RunCommand("list projects", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(projectsDeprecatedMsg))
			Expect(output).To(ContainSubstring("p301"))
			Expect(output).To(ContainSubstring("301"))
		})

		// list workspaces
		It("list workloads all", func() {
			mockPaginator.Set(filter.NewBuilder().WithClusterID(cli.ClusterOneUuid),
				int32(101), int32(0), nil, nil,
				workloadGenerator.SetNameAndUUID(WorkloadName01).Build(),
				workloadGenerator.SetNameAndUUID(WorkloadName02).Build())
			output, err := cli.RunCommand("list workloads -A", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(jobsDeprecatedMsg))
			Expect(output).To(ContainSubstring(WorkloadName01))
			Expect(output).To(ContainSubstring(WorkloadName02))
		})

		It("list workloads all projects", func() {
			mockPaginator.Set(filter.NewBuilder().WithClusterID(cli.ClusterOneUuid),
				int32(101), int32(0), nil, nil,
				workloadGenerator.SetNameAndUUID(WorkloadName01).Build(),
				workloadGenerator.SetNameAndUUID(WorkloadName02).Build())
			output, err := cli.RunCommand("list workloads --all-projects", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(jobsDeprecatedMsg))
			Expect(output).To(ContainSubstring(WorkloadName01))
			Expect(output).To(ContainSubstring(WorkloadName02))
		})

		It("list workloads all projects", func() {
			mockPaginator.Set(filter.NewBuilder().WithClusterID(cli.ClusterOneUuid),
				int32(101), int32(0), nil, nil,
				workloadGenerator.SetNameAndUUID(WorkloadName01).Build(),
				workloadGenerator.SetNameAndUUID(WorkloadName02).Build())
			output, err := cli.RunCommand("list workloads --all-projects", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(jobsDeprecatedMsg))
			Expect(output).To(ContainSubstring(WorkloadName01))
			Expect(output).To(ContainSubstring(WorkloadName02))
		})

		It("list jobs all", func() {
			mockPaginator.Set(filter.NewBuilder().WithClusterID(cli.ClusterOneUuid),
				int32(101), int32(0), nil, nil,
				workloadGenerator.SetNameAndUUID(WorkloadName01).Build(),
				workloadGenerator.SetNameAndUUID(WorkloadName02).Build())
			output, err := cli.RunCommand("list jobs -A", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(jobsDeprecatedMsg))
			Expect(output).To(ContainSubstring(WorkloadName01))
			Expect(output).To(ContainSubstring(WorkloadName02))
		})

		It("list jobs all projects", func() {
			mockPaginator.Set(filter.NewBuilder().WithClusterID(cli.ClusterOneUuid),
				int32(101), int32(0), nil, nil,
				workloadGenerator.SetNameAndUUID(WorkloadName01).Build(),
				workloadGenerator.SetNameAndUUID(WorkloadName02).Build())
			output, err := cli.RunCommand("list jobs --all-projects", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(jobsDeprecatedMsg))
			Expect(output).To(ContainSubstring(WorkloadName01))
			Expect(output).To(ContainSubstring(WorkloadName02))
		})

		It("workload list project p301", func() {
			mockPaginator.Set(filter.NewBuilder().WithClusterID(cli.ClusterOneUuid).WithProject(ProjectNameP301),
				int32(101), int32(0), nil, nil,
				workloadGenerator.SetNameAndUUID(WorkloadName01).Build())
			output, err := cli.RunCommand("list workloads --project=p301", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(jobsDeprecatedMsg))
			Expect(output).To(ContainSubstring(WorkloadName01))
			Expect(output).ToNot(ContainSubstring(WorkloadName02))
		})

		It("workload list project (short) p301", func() {
			mockPaginator.Set(filter.NewBuilder().WithClusterID(cli.ClusterOneUuid).WithProject(ProjectNameP301),
				int32(101), int32(0), nil, nil,
				workloadGenerator.SetNameAndUUID(WorkloadName01).Build())
			output, err := cli.RunCommand("list workloads -p=p301", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(jobsDeprecatedMsg))
			Expect(output).To(ContainSubstring(WorkloadName01))
			Expect(output).ToNot(ContainSubstring(WorkloadName02))
		})

		It("list jobs project p301", func() {
			mockPaginator.Set(filter.NewBuilder().WithClusterID(cli.ClusterOneUuid).WithProject(ProjectNameP301),
				int32(101), int32(0), nil, nil,
				workloadGenerator.SetNameAndUUID(WorkloadName01).Build())
			output, err := cli.RunCommand("list jobs --project=p301", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(jobsDeprecatedMsg))
			Expect(output).To(ContainSubstring(WorkloadName01))
			Expect(output).ToNot(ContainSubstring(WorkloadName02))
		})

		It("list jobs project (short) p301", func() {
			mockPaginator.Set(filter.NewBuilder().WithClusterID(cli.ClusterOneUuid).WithProject(ProjectNameP301),
				int32(101), int32(0), nil, nil,
				workloadGenerator.SetNameAndUUID(WorkloadName01).Build())
			output, err := cli.RunCommand("list jobs -p=p301", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(HavePrefix(jobsDeprecatedMsg))
			Expect(output).To(ContainSubstring(WorkloadName01))
			Expect(output).ToNot(ContainSubstring(WorkloadName02))
		})

		Context("workload list - no subcommand", func() {
			It("all", func() {
				mockPaginator.Set(filter.NewBuilder().WithClusterID(cli.ClusterOneUuid),
					int32(101), int32(0), nil, nil,
					workloadGenerator.SetNameAndUUID(WorkloadName01).Build(),
					workloadGenerator.SetNameAndUUID(WorkloadName02).Build())
				output, err := cli.RunCommand("list -A", envMap)
				Expect(err).To(BeNil())
				Expect(output).To(HavePrefix(listDeprecatedMsg))
				Expect(output).To(ContainSubstring(clients.WorkloadName01))
				Expect(output).To(ContainSubstring(clients.WorkloadName02))
			})

			It("short project", func() {
				mockPaginator.Set(filter.NewBuilder().WithClusterID(cli.ClusterOneUuid).WithProject(ProjectNameP301),
					int32(101), int32(0), nil, nil,
					workloadGenerator.SetNameAndUUID(WorkloadName01).Build())
				output, err := cli.RunCommand("list -p=p301", envMap)
				Expect(err).To(BeNil())
				Expect(output).To(HavePrefix(listDeprecatedMsg))
				Expect(output).To(ContainSubstring(clients.WorkloadName01))
				Expect(output).ToNot(ContainSubstring(clients.WorkloadName02))
			})
		})

		It("workload list - without subcommand and without flags", func() {
			output, err := cli.RunCommand("list", envMap)
			Expect(err).NotTo(BeNil())
			Expect(output).To(HavePrefix(listDeprecatedMsg))
			Expect(output).To(ContainSubstring("Error: project is missing. value is empty."))
		})

		It("workload list - help", func() {
			output, err := cli.RunCommand("list --help", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(listDeprecatedHelpMsg))
			Expect(output).To(ContainSubstring("Usage:"))
		})

		// todo: add list nodes test
	})

	Context("Deprecated environment submit validation", Ordered, func() {
		BeforeEach(func() {
			submitcmd.SubmitSpec.Clear()
		})

		It("verify mpi submit with deprecated --master-environment flag shows warning", func() {
			clients.NewWorkloadSubmitMocker(gomock.NewController(GinkgoT()), &clients.Distributed{}, []string{"mpi-test"})
			commandStr := "mpi submit mpi-test -i ubuntu -p p301 --master-environment a=b"
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).To(BeNil())

			Expect(output).To(ContainSubstring(DepMasterEnvironmentMsg))
			Expect(output).To(ContainSubstring("Creating distributed mpi-test"))
		})

		It("verify training submit with --environment key=value shows deprecation warning", func() {
			clients.NewWorkloadSubmitMocker(gomock.NewController(GinkgoT()), &clients.Training{}, []string{"training-test"})
			commandStr := "training submit training-test -i ubuntu -p p301 --environment MY_VAR=my_value"
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).To(BeNil())

			Expect(output).To(ContainSubstring(DepEnvironmentMsg))
			Expect(output).To(ContainSubstring("Creating training training-test"))
		})

		It("verify training submit with --environment-variable does not show deprecation", func() {
			clients.NewWorkloadSubmitMocker(gomock.NewController(GinkgoT()), &clients.Training{}, []string{"training-test-2"})
			commandStr := "training submit training-test-2 -i ubuntu -p p301 --environment-variable MY_VAR=my_value"
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).To(BeNil())

			Expect(output).ToNot(ContainSubstring(DepEnvironmentMsg))
			Expect(output).To(ContainSubstring("Creating training training-test-2"))
		})

		It("verify training submit with -e flag does not show deprecation", func() {
			clients.NewWorkloadSubmitMocker(gomock.NewController(GinkgoT()), &clients.Training{}, []string{"training-test-3"})
			commandStr := "training submit training-test-3 -i ubuntu -p p301 -e MY_VAR=my_value"
			output, err := cli.RunCommand(commandStr, envMap)
			Expect(err).To(BeNil())

			Expect(output).ToNot(ContainSubstring(DepEnvironmentMsg))
			Expect(output).To(ContainSubstring("Creating training training-test-3"))
		})
	})
})
