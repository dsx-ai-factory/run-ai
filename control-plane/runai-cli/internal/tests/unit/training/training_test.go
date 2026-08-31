/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package training

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/list/manipulate"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api/filter"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth/fake"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/clients"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

const (
	trainingTestFilePath = "training"
)

var _ = Describe("Training Test", Ordered, func() {
	var (
		configPath string
		err        error
		envMap     map[string]string

		mockPaginator     *clients.WorkloadPaginationMocker
		workloadGenerator clients.WorkloadGenerator
	)

	BeforeAll(func() {
		configPath, err = cli.GetHomePath()
		if err != nil {
			Fail("failed to get pwd")
		}
		configPath = cli.GetConfigPath(configPath, trainingTestFilePath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)

		mockCtrl := gomock.NewController(GinkgoT())
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)

		mockPaginator = clients.NewWorkloadPaginationMocker(mockCtrl)
		workloadGenerator.
			SetTenant(1).
			SetType(api.Training).
			SetPods(0, ptr.To[int32](5), nil, nil).
			SetClusterID(cli.ClusterOneUuid).
			SetPhase(openapi.PHASE_RUNNING)

		mockPaginator.SetAny(filter.NewBuilder().WithClusterID(cli.ClusterOneUuid).WithWorkloadType("Training||Distributed"),
			int32(101), int32(0), nil, nil,
			workloadGenerator.SetNameAndUUID(clients.WorkloadName02).SetPods(1, ptr.To[int32](5), ptr.To[int32](5), nil).Build(),
			workloadGenerator.SetNameAndUUID(clients.WorkloadName03).SetPods(1, nil, nil, ptr.To[int32](5)).Build())

		mockPaginator.SetAny(filter.NewBuilder().WithClusterID(cli.ClusterOneUuid).WithWorkloadType("Training||Distributed").WithProject(clients.ProjectNameP301),
			int32(101), int32(0), nil, nil,
			workloadGenerator.SetNameAndUUID(clients.WorkloadName03).Build())

		mockPaginator.SetAny(filter.NewBuilder().
			WithClusterID(cli.ClusterOneUuid).
			WithWorkloadType("Training||Distributed").
			WithWorkloadFramework(api.Runai).
			WithProject(clients.ProjectNameP302),
			int32(101), int32(0), nil, nil,
			workloadGenerator.SetNameAndUUID(clients.WorkloadName02).Build())
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

	Context("Training list validation", Ordered, func() {
		It("Training list all", func() {
			output, err := cli.RunCommand("training list -A", envMap)
			Expect(err).To(BeNil())

			Expect(output).ToNot(ContainSubstring(clients.WorkloadName01))

			// with parallelism
			Expect(output).To(ContainSubstring(clients.WorkloadName02))
			Expect(output).To(ContainSubstring("1/5"))

			// with completions
			Expect(output).To(ContainSubstring(clients.WorkloadName03))
			Expect(output).To(ContainSubstring("1/-"))
		})

		It("Training list all - JSON output", func() {
			output, err := cli.RunCommand("training list -A --json", envMap)
			Expect(err).To(BeNil())

			var data manipulate.Data
			err = json.Unmarshal([]byte(output), &data)
			Expect(err).To(BeNil())
			Expect(data.Workloads).To(HaveLen(2))
			Expect(data.NextToken).To(BeNil())

			// with parallelism
			workload1 := data.Workloads[0]
			Expect(workload1.Name).To(Equal(clients.WorkloadName02))
			Expect(workload1.RunningPods).To(Equal(int32(1)))
			Expect(workload1.RequestedPods.Parallelism.Get()).To(Equal(newNullableInt32(5).Get()))

			// with completions
			workload2 := data.Workloads[1]
			Expect(workload2.Name).To(Equal(clients.WorkloadName03))
			Expect(workload2.RunningPods).To(Equal(int32(1)))
			Expect(workload2.RequestedPods.Completions.Get()).To(Equal(newNullableInt32(5).Get()))
			Expect(workload2.RequestedPods.Parallelism.IsSet()).To(Equal(false))

		})

		It("Training list all", func() {
			output, err := cli.RunCommand("training list -A", envMap)
			Expect(err).To(BeNil())
			Expect(output).ToNot(ContainSubstring(clients.WorkloadName01))
			Expect(output).To(ContainSubstring("1/5"))
			Expect(output).To(ContainSubstring(clients.WorkloadName02))
		})

		It("Training list all and project flags", func() {
			_, err := cli.RunCommand("training list -A --project=p301", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("[all project] are set none of the others can be;"))
		})

		It("Training list project p301", func() {
			output, err := cli.RunCommand("training list --project=p301", envMap)
			Expect(err).To(BeNil())
			Expect(output).ToNot(ContainSubstring(clients.WorkloadName01))
			Expect(output).ToNot(ContainSubstring(clients.WorkloadName02))
		})

		It("Training list project p300", func() {
			output, err := cli.RunCommand("training list --project=p300", envMap)
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedTrainingListWithProjectThatDoesNotExist))
		})

		Describe("Training list framework", func() {
			BeforeAll(func() {
				err := clients.ProjectSet(clients.ProjectTwoName, envMap)
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

			It("Training with framework", func() {
				output, err := cli.RunCommand("training list --framework runai", envMap)
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(ContainSubstring(clients.WorkloadName02))
			})
		})
	})
})

func newNullableInt32(val int) *openapi.NullableInt32 {
	return openapi.NewNullableInt32(ptr.To(int32(val)))
}
