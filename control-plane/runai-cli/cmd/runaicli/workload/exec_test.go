/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload_test

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	clusterconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Exec Command", func() {
	Context("Command structure", func() {
		It("should have correct command name", func() {
			Expect(workload.Exec.Use).To(Equal("exec"))
		})

		It("should have a short description", func() {
			Expect(workload.Exec.Short).NotTo(BeEmpty())
			Expect(workload.Exec.Short).To(ContainSubstring("exec"))
		})

		It("should have a long description with examples", func() {
			Expect(workload.Exec.Long).NotTo(BeEmpty())
			Expect(workload.Exec.Long).To(ContainSubstring("runai workload exec"))
		})

		It("should require at least 1 argument", func() {
			err := workload.Exec.Args(workload.Exec, []string{"workload-name"})
			Expect(err).To(BeNil())

			err = workload.Exec.Args(workload.Exec, []string{"workload-name", "bash"})
			Expect(err).To(BeNil())

			err = workload.Exec.Args(workload.Exec, []string{})
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Flag registration", func() {
		It("should have project flag", func() {
			flag := workload.Exec.Flags().Lookup("project")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("p"))
		})

		It("should have pod flag", func() {
			flag := workload.Exec.Flags().Lookup("pod")
			Expect(flag).NotTo(BeNil())
		})

		It("should have container flag", func() {
			flag := workload.Exec.Flags().Lookup("container")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("c"))
		})

		It("should have stdin flag", func() {
			flag := workload.Exec.Flags().Lookup("stdin")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("i"))
		})

		It("should have tty flag", func() {
			flag := workload.Exec.Flags().Lookup("tty")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("t"))
		})

		It("should have pod-running-timeout flag", func() {
			flag := workload.Exec.Flags().Lookup("pod-running-timeout")
			Expect(flag).NotTo(BeNil())
		})

		It("should have wait-timeout flag", func() {
			flag := workload.Exec.Flags().Lookup("wait-timeout")
			Expect(flag).NotTo(BeNil())
		})
	})

	Context("Command help and documentation", func() {
		It("should have meaningful help text", func() {
			Expect(workload.Exec.Short).To(ContainSubstring("exec"))
			Expect(workload.Exec.Short).To(ContainSubstring("workload"))
		})

		It("should have comprehensive examples", func() {
			Expect(workload.Exec.Long).To(ContainSubstring("runai workload exec"))
			Expect(workload.Exec.Long).To(ContainSubstring("--pod"))
			Expect(workload.Exec.Long).To(ContainSubstring("--container"))
			Expect(workload.Exec.Long).To(ContainSubstring("--stdin"))
			Expect(workload.Exec.Long).To(ContainSubstring("--tty"))
			Expect(workload.Exec.Long).To(ContainSubstring("--wait-timeout"))
		})
	})
})

var _ = Describe("DetermineTypeAndFramework (exec reuse)", func() {
	Context("Workspace workloads", func() {
		It("should return Workspace type and Runai framework", func() {
			workloadInfo := &workload.WorkloadInfo{
				Type: api.Workspace.StringCapitalize(),
			}

			wlType, framework, err := workload.DetermineTypeAndFramework(workloadInfo)

			Expect(err).To(BeNil())
			Expect(wlType).To(Equal(api.Workspace))
			Expect(framework).To(Equal(api.Runai))
		})
	})

	Context("Inference workloads", func() {
		It("should return Inference type and Knative framework", func() {
			workloadInfo := &workload.WorkloadInfo{
				Type: api.Inference.StringCapitalize(),
			}

			wlType, framework, err := workload.DetermineTypeAndFramework(workloadInfo)

			Expect(err).To(BeNil())
			Expect(wlType).To(Equal(api.Inference))
			Expect(framework).To(Equal(api.Knative))
		})
	})

	Context("Training workloads", func() {
		It("should return Training type and Runai framework", func() {
			workloadInfo := &workload.WorkloadInfo{
				Type: api.Training.StringCapitalize(),
			}

			wlType, framework, err := workload.DetermineTypeAndFramework(workloadInfo)

			Expect(err).To(BeNil())
			Expect(wlType).To(Equal(api.Training))
			Expect(framework).To(Equal(api.Runai))
		})
	})

	Context("Distributed PyTorch workloads", func() {
		It("should return Distributed type and PyTorch framework", func() {
			workloadInfo := &workload.WorkloadInfo{
				Type:      api.Distributed.StringCapitalize(),
				Framework: clusterconstants.KindPyTorchJob,
			}

			wlType, framework, err := workload.DetermineTypeAndFramework(workloadInfo)

			Expect(err).To(BeNil())
			Expect(wlType).To(Equal(api.Distributed))
			Expect(framework).To(Equal(api.PyTorch))
		})
	})

	Context("Unknown workload type", func() {
		It("should return error for unknown workload type", func() {
			workloadInfo := &workload.WorkloadInfo{
				Type: "UnknownType",
			}

			wlType, framework, err := workload.DetermineTypeAndFramework(workloadInfo)

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("unknown workload type"))
			Expect(wlType).To(Equal(api.WorkloadType(-1)))
			Expect(framework).To(Equal(api.Framework(-1)))
		})
	})
})
