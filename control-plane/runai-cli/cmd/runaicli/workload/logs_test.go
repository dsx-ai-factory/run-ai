/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/cmd/runaicli/workload"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
	clusterconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
)

var _ = Describe("Logs Command", func() {
	Context("Command structure", func() {
		It("should have correct command name", func() {
			Expect(workload.Logs.Use).To(Equal("logs"))
		})

		It("should have a short description", func() {
			Expect(workload.Logs.Short).NotTo(BeEmpty())
			Expect(workload.Logs.Short).To(ContainSubstring("logs"))
		})

		It("should have a long description with examples", func() {
			Expect(workload.Logs.Long).NotTo(BeEmpty())
			Expect(workload.Logs.Long).To(ContainSubstring("runai workload logs"))
		})

		It("should require exactly 1 argument", func() {
			err := workload.Logs.Args(workload.Logs, []string{"workload-name"})
			Expect(err).To(BeNil())

			err = workload.Logs.Args(workload.Logs, []string{})
			Expect(err).NotTo(BeNil())

			err = workload.Logs.Args(workload.Logs, []string{"arg1", "arg2"})
			Expect(err).NotTo(BeNil())
		})
	})

	Context("Flag registration", func() {
		It("should have project flag", func() {
			flag := workload.Logs.Flags().Lookup("project")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("p"))
		})

		It("should have pod flag", func() {
			flag := workload.Logs.Flags().Lookup("pod")
			Expect(flag).NotTo(BeNil())
		})

		It("should have container flag", func() {
			flag := workload.Logs.Flags().Lookup("container")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("c"))
		})

		It("should have tail flag", func() {
			flag := workload.Logs.Flags().Lookup("tail")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("t"))
			Expect(flag.DefValue).To(Equal("-1"))
		})

		It("should have follow flag", func() {
			flag := workload.Logs.Flags().Lookup("follow")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("f"))
		})

		It("should have timestamps flag", func() {
			flag := workload.Logs.Flags().Lookup("timestamps")
			Expect(flag).NotTo(BeNil())
		})

		It("should have previous flag", func() {
			flag := workload.Logs.Flags().Lookup("previous")
			Expect(flag).NotTo(BeNil())
		})

		It("should have limit-bytes flag", func() {
			flag := workload.Logs.Flags().Lookup("limit-bytes")
			Expect(flag).NotTo(BeNil())
		})

		It("should have since flag", func() {
			flag := workload.Logs.Flags().Lookup("since")
			Expect(flag).NotTo(BeNil())
		})

		It("should have since-time flag", func() {
			flag := workload.Logs.Flags().Lookup("since-time")
			Expect(flag).NotTo(BeNil())
		})

		It("should have wait-timeout flag", func() {
			flag := workload.Logs.Flags().Lookup("wait-timeout")
			Expect(flag).NotTo(BeNil())
		})
	})

	Context("Command help and documentation", func() {
		It("should have meaningful help text", func() {
			Expect(workload.Logs.Short).To(ContainSubstring("logs"))
			Expect(workload.Logs.Short).To(ContainSubstring("workload"))
		})

		It("should have comprehensive examples", func() {
			Expect(workload.Logs.Long).To(ContainSubstring("runai workload logs"))
			Expect(workload.Logs.Long).To(ContainSubstring("--pod"))
			Expect(workload.Logs.Long).To(ContainSubstring("--container"))
			Expect(workload.Logs.Long).To(ContainSubstring("--tail"))
			Expect(workload.Logs.Long).To(ContainSubstring("--timestamps"))
			Expect(workload.Logs.Long).To(ContainSubstring("--follow"))
			Expect(workload.Logs.Long).To(ContainSubstring("--previous"))
			Expect(workload.Logs.Long).To(ContainSubstring("--limit-bytes"))
			Expect(workload.Logs.Long).To(ContainSubstring("--since"))
			Expect(workload.Logs.Long).To(ContainSubstring("--since-time"))
			Expect(workload.Logs.Long).To(ContainSubstring("--wait-timeout"))
		})
	})
})

var _ = Describe("determineTypeAndFramework", func() {
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

	Context("DistributedInference workloads", func() {
		It("should return DistributedInference type and LWS framework", func() {
			workloadInfo := &workload.WorkloadInfo{
				Type: api.DistributedInference.StringCapitalize(),
			}

			wlType, framework, err := workload.DetermineTypeAndFramework(workloadInfo)

			Expect(err).To(BeNil())
			Expect(wlType).To(Equal(api.DistributedInference))
			Expect(framework).To(Equal(api.LWS))
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

	Context("Distributed MPI workloads", func() {
		It("should return Distributed type and MPI framework", func() {
			workloadInfo := &workload.WorkloadInfo{
				Type:      api.Distributed.StringCapitalize(),
				Framework: clusterconstants.KindMpiJob,
			}

			wlType, framework, err := workload.DetermineTypeAndFramework(workloadInfo)

			Expect(err).To(BeNil())
			Expect(wlType).To(Equal(api.Distributed))
			Expect(framework).To(Equal(api.MPI))
		})
	})

	Context("Distributed TensorFlow workloads", func() {
		It("should return Distributed type and TF framework", func() {
			workloadInfo := &workload.WorkloadInfo{
				Type:      api.Distributed.StringCapitalize(),
				Framework: clusterconstants.KindTFJob,
			}

			wlType, framework, err := workload.DetermineTypeAndFramework(workloadInfo)

			Expect(err).To(BeNil())
			Expect(wlType).To(Equal(api.Distributed))
			Expect(framework).To(Equal(api.TF))
		})
	})

	Context("Distributed XGBoost workloads", func() {
		It("should return Distributed type and XGBoost framework", func() {
			workloadInfo := &workload.WorkloadInfo{
				Type:      api.Distributed.StringCapitalize(),
				Framework: clusterconstants.KindXGBoostJob,
			}

			wlType, framework, err := workload.DetermineTypeAndFramework(workloadInfo)

			Expect(err).To(BeNil())
			Expect(wlType).To(Equal(api.Distributed))
			Expect(framework).To(Equal(api.XGBoost))
		})
	})

	Context("Distributed JAX workloads", func() {
		It("should return Distributed type and Jax framework", func() {
			workloadInfo := &workload.WorkloadInfo{
				Type:      api.Distributed.StringCapitalize(),
				Framework: clusterconstants.KindJAXJob,
			}

			wlType, framework, err := workload.DetermineTypeAndFramework(workloadInfo)

			Expect(err).To(BeNil())
			Expect(wlType).To(Equal(api.Distributed))
			Expect(framework).To(Equal(api.Jax))
		})
	})

	Context("Distributed RunaiJob workloads", func() {
		It("should return Distributed type and Runai framework for RunaiJob", func() {
			workloadInfo := &workload.WorkloadInfo{
				Type:      api.Distributed.StringCapitalize(),
				Framework: clusterconstants.KindRunaiJob,
			}

			wlType, framework, err := workload.DetermineTypeAndFramework(workloadInfo)

			Expect(err).To(BeNil())
			Expect(wlType).To(Equal(api.Distributed))
			Expect(framework).To(Equal(api.Runai))
		})

		It("should return Distributed type and Runai framework for empty framework", func() {
			workloadInfo := &workload.WorkloadInfo{
				Type:      api.Distributed.StringCapitalize(),
				Framework: "",
			}

			wlType, framework, err := workload.DetermineTypeAndFramework(workloadInfo)

			Expect(err).To(BeNil())
			Expect(wlType).To(Equal(api.Distributed))
			Expect(framework).To(Equal(api.Runai))
		})
	})

	Context("Distributed unknown framework", func() {
		It("should return Distributed type and AnyDistributed framework for unknown framework", func() {
			workloadInfo := &workload.WorkloadInfo{
				Type:      api.Distributed.StringCapitalize(),
				Framework: "UnknownFramework",
			}

			wlType, framework, err := workload.DetermineTypeAndFramework(workloadInfo)

			Expect(err).To(BeNil())
			Expect(wlType).To(Equal(api.Distributed))
			Expect(framework).To(Equal(api.AnyDistributed))
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
