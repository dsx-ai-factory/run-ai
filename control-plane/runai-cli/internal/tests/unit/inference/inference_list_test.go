/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference_test

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/list/manipulate"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth/fake"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/cli/clients"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/tests/unit/inference"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

const InferenceListTestFilePath = "inference-list"

var _ = Describe("Inference List Test", Ordered, func() {
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
		configPath = cli.GetConfigPath(configPath, InferenceListTestFilePath)
		envMap = cli.CreateEnvMap(configPath)
		_ = os.RemoveAll(configPath)

		mockCtrl := gomock.NewController(GinkgoT())
		auth.GetManager().Init(fake.NewValidTokenBackend, nil)
		clients.MockAuthentication(mockCtrl)
		clients.MockClusters(mockCtrl)
		clients.MockProject(mockCtrl)
		clients.MockWorkloadByList(mockCtrl, inference.FilterList)
	})

	BeforeEach(func() {
		ResetRunaiCommand()
		_, err = cli.RunCommand(fmt.Sprintf("config set --cp-url=%s", cli.ClusterOneDomain), envMap)
		if err != nil {
			Fail("failed to set cp-url, error: " + err.Error())
		}
		err = clients.ClusterSet(envMap)
		if err != nil {
			Fail("cluster set failed, error: " + err.Error())
		}
	})

	Context("Pre workload command validation", Ordered, func() {
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

	Context("Inference list validation", Ordered, func() {
		It("inference list all", func() {
			output, err := cli.RunCommand("inference list -A", envMap)
			Expect(err).To(BeNil())
			Expect(output).To(ContainSubstring(clients.WorkloadName01))
			Expect(output).To(ContainSubstring("0/0-1"))
			Expect(output).To(ContainSubstring(clients.WorkloadName02))
			Expect(output).To(ContainSubstring("1/1-3"))
			Expect(output).To(ContainSubstring(clients.WorkloadName03))
			Expect(output).To(ContainSubstring("-"))
			Expect(output).To(ContainSubstring(inference.Name04))
			Expect(output).To(ContainSubstring("2/2"))
		})

		It("inference list all as a json", func() {
			output, err := cli.RunCommand("inference list -A --json", envMap)
			Expect(err).To(BeNil())
			var data manipulate.Data
			err = json.Unmarshal([]byte(output), &data)
			Expect(err).To(BeNil())
			Expect(data.NextToken).To(BeNil())
			briefList := data.Workloads
			Expect(briefList).To(HaveLen(4))

			workload1 := briefList[0]
			Expect(workload1.Name).To(Equal(clients.WorkloadName01))
			Expect(workload1.RunningPods).To(Equal(int32(0)))
			Expect(workload1.RequestedPods.Min.Get()).To(Equal(newNullableInt32(0).Get()))
			Expect(workload1.RequestedPods.Max.Get()).To(Equal(newNullableInt32(1).Get()))
			Expect(workload1.RequestedPods.Number.IsSet()).To(Equal(false))

			workload2 := briefList[1]
			Expect(workload2.Name).To(Equal(clients.WorkloadName02))
			Expect(workload2.RunningPods).To(Equal(int32(1)))
			Expect(workload2.RequestedPods.Min.Get()).To(Equal(newNullableInt32(1).Get()))
			Expect(workload2.RequestedPods.Max.Get()).To(Equal(newNullableInt32(3).Get()))
			Expect(workload2.RequestedPods.Number.IsSet()).To(Equal(false))

			workload3 := briefList[2]
			Expect(workload3.Name).To(Equal(clients.WorkloadName03))
			Expect(workload3.RunningPods).To(Equal(int32(0)))
			Expect(workload3.RequestedPods).To(BeNil())

			workload4 := briefList[3]
			Expect(workload4.Name).To(Equal(inference.Name04))
			Expect(workload4.RunningPods).To(Equal(int32(2)))
			Expect(workload4.RequestedPods.Number.Get()).To(Equal(newNullableInt32(2).Get()))
		})

		It("inference list all and project flags", func() {
			_, err := cli.RunCommand("inference list -A --project=p301", envMap)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("[all project] are set none of the others can be;"))
		})

		It("inference list project p301", func() {
			output, err := cli.RunCommand("inference list --project=p301 --json", envMap)
			Expect(err).To(BeNil())

			var data manipulate.Data
			Expect(json.Unmarshal([]byte(output), &data)).To(Succeed())
			count := 0
			for _, workload := range data.Workloads {
				if workload.ProjectName == clients.ProjectNameP301 {
					count++
				}
			}
			Expect(count).To(Equal(3))
		})

		It("inference list project p300", func() {
			output, err := cli.RunCommand("inference list --project=p300", envMap)

			// Now inference list validates project existence like training does
			Expect(err).ToNot(BeNil())
			Expect(output).To(HaveSuffix(ExpectedInferenceListWithProjectThatDoesNotExist))
		})

		It("inference list empty - should show empty table", func() {
			// Test that when there are no workloads, an empty table is displayed
			output, err := cli.RunCommand("inference list --project=p302", envMap)
			Expect(err).To(BeNil())
			// Should contain table headers
			Expect(output).To(ContainSubstring("Workload"))
			Expect(output).To(ContainSubstring("Status"))
			// Should not contain "No inference workloads found" message
			Expect(output).ToNot(ContainSubstring("No inference workloads found"))
		})

		It("inference list empty as json - should return empty array", func() {
			// Test that when there are no workloads, an empty JSON array is returned
			output, err := cli.RunCommand("inference list --project=p302 --json", envMap)
			Expect(err).To(BeNil())

			var data manipulate.Data
			err = json.Unmarshal([]byte(output), &data)
			Expect(err).To(BeNil())
			Expect(data.Workloads).To(HaveLen(0))
			Expect(data.NextToken).To(BeNil())
		})

		Describe("Inference list project flags", func() {
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

			It("Inference list project with default project", func() {
				output, err := cli.RunCommand("inference list", envMap)
				Expect(err).To(BeNil())
				Expect(output).To(ContainSubstring(clients.WorkloadName01))
				Expect(output).To(ContainSubstring(clients.WorkloadName02))
				Expect(output).ToNot(ContainSubstring(clients.WorkloadName03))
				Expect(output).To(ContainSubstring(inference.Name04))
			})

			It("Inference list all projects with default project set", func() {
				output, err := cli.RunCommand("inference list --all", envMap)
				Expect(err).To(BeNil())
				Expect(output).To(ContainSubstring(clients.WorkloadName01))
				Expect(output).To(ContainSubstring(clients.WorkloadName02))
				Expect(output).To(ContainSubstring(clients.WorkloadName03))
				Expect(output).To(ContainSubstring(inference.Name04))
			})
		})

		Describe("Inference list framework", func() {
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

			It("Inference with framework - no such flag", func() {
				_, err := cli.RunCommand("inference list --framework runai", envMap)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("unknown flag: --framework"))
			})
		})
	})

})

func newNullableInt32(val int) *openapi.NullableInt32 {
	return openapi.NewNullableInt32(ptr.To(int32(val)))
}
