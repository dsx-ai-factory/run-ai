/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clusterCommon "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
)

var _ = Describe("API consts tests", func() {
	Context("WorkloadType", func() {
		const (
			workspace            = "workspace"
			training             = "training"
			distributed          = "distributed"
			inference            = "inference"
			distributedInference = "distributedinference"
		)

		Describe("String method", func() {
			It("should return the correct string representation", func() {
				Expect(Workspace.String()).To(Equal(workspace))
				Expect(Training.String()).To(Equal(training))
				Expect(Distributed.String()).To(Equal(distributed))
				Expect(Inference.String()).To(Equal(inference))
				Expect(DistributedInference.String()).To(Equal(distributedInference))
			})
		})

		Describe("ToK8sResource method", func() {
			It("should return the correct Kubernetes resource kind", func() {
				Expect(Workspace.ToK8sResource()).To(Equal(clusterCommon.KindInteractiveWorkload))
				Expect(Training.ToK8sResource()).To(Equal(clusterCommon.KindTrainingWorkload))
				Expect(Distributed.ToK8sResource()).To(Equal(clusterCommon.KindDistributedWorkload))
				Expect(Inference.ToK8sResource()).To(Equal(clusterCommon.KindInferenceWorkload))
				Expect(DistributedInference.ToK8sResource()).To(Equal(clusterCommon.KindDistributedInferenceWorkload))
			})

			It("should return an empty string for an invalid WorkloadType", func() {
				invalidWorkload := WorkloadType("invalid")
				Expect(invalidWorkload.ToK8sResource()).To(Equal(""))
			})
		})

		Describe("ValidateWorkloadType function", func() {
			It("should return true for valid workload types", func() {
				Expect(ValidateWorkloadType(workspace)).To(BeTrue())
				Expect(ValidateWorkloadType(training)).To(BeTrue())
				Expect(ValidateWorkloadType(distributed)).To(BeTrue())
				Expect(ValidateWorkloadType(inference)).To(BeTrue())
			})

			It("should return false for invalid workload types", func() {
				Expect(ValidateWorkloadType("invalid")).To(BeFalse())
				Expect(ValidateWorkloadType("")).To(BeFalse())
			})
		})
	})

	Context("WorkloadFramework", func() {
		It("String", func() {
			Expect(RunaiFramework.String()).To(Equal(RunaiFrameworkName))
			Expect(MPIFramework.String()).To(Equal(MPIFrameworkName))
			Expect(PyTorchFramework.String()).To(Equal(PyTorchFrameworkName))
			Expect(TFFramework.String()).To(Equal(TensorFlowFrameworkName))
			Expect(XGBoostFramework.String()).To(Equal(XGBoostFrameworkName))
			Expect(JAXFramework.String()).To(Equal(JAXFrameworkName))
			Expect(AnyDistributedFramework.String()).To(Equal(AnyDistributedFrameworkName))
			Expect(KnativeFramework.String()).To(Equal(KnativeFrameworkName))
			Expect(LwsFramework.String()).To(Equal(LwsFrameworkName))
		})

		Describe("IsValidWorkloadFramework", func() {
			It("should return true for valid workload frameworks", func() {
				for _, framework := range []string{RunaiFrameworkName, MPIFrameworkName, PyTorchFrameworkName, TensorFlowFrameworkName, XGBoostFrameworkName, AnyDistributedFrameworkName, JAXFrameworkName, KnativeFrameworkName} {
					Expect(IsValidWorkloadFramework(framework)).To(BeTrue())
				}
			})

			It("should return false for invalid workload framework", func() {
				for _, framework := range []string{"blabla", ""} {
					Expect(IsValidWorkloadFramework(framework)).To(BeFalse())
				}
			})
		})

		DescribeTable("JobType",
			func(framework WorkloadFramework, expectedType JobType) {
				jobType, err := framework.JobType()
				Expect(jobType).To(Equal(expectedType))
				Expect(err).NotTo(HaveOccurred())
			},
			Entry("runai", RunaiFramework, RunaiJob),
			Entry("mpi", MPIFramework, MPIJob),
			Entry("pytorch", PyTorchFramework, PyTorchJob),
			Entry("tf", TFFramework, TFJob),
			Entry("xgboost", XGBoostFramework, XGBoostJob),
			Entry("jax", JAXFramework, JAXJob),
			Entry("knative", KnativeFramework, InferenceJob),
		)

		It("JobType for unsupported framework", func() {
			jobType, err := AnyDistributedFramework.JobType()
			Expect(jobType.String()).To(Equal(""))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unsupported workload framework any"))
		})
	})
})
