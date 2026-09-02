/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common_test

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	fakeClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"k8s.io/apimachinery/pkg/runtime"

	kubeflowcommon "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	workloadproject "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	dwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

var _ = Describe("DistributedBuilderWorker", func() {
	var (
		builder    *common.DistributedBuilder
		workerSpec *kubeflowcommon.ReplicaSpec
		ctx        context.Context
	)

	BeforeEach(func() {
		builder = &common.DistributedBuilder{}
		builder.MergedWorkload = &dwv2.DistributedWorkload{}
		builder.ObjectNaming = &metav1.ObjectMeta{
			Name:      "test",
			Namespace: "runai" + projectName,
		}
		workerSpec = &kubeflowcommon.ReplicaSpec{}
		ctx = workloadproject.SetProjectInContext(context.Background(), projectName)
	})

	Context("BuildWorkerReplicaSpec", func() {
		When("worker spec has readinessProbe", func() {
			BeforeEach(func() {
				builder.MergedWorkload.Spec.ReadinessProbe = &flds.ProbeField{
					Values: &flds.ProbeMembers{
						HttpGetPort: *flds.NewOptionalIntField(pointer.Int(8080)),
					},
				}
			})

			It("sets readinessProbe", func() {
				err := builder.BuildWorkerReplicaSpec(ctx, workerSpec, 0)
				Expect(err).NotTo(HaveOccurred())
				Expect(workerSpec.Template.Spec.Containers[0].ReadinessProbe).NotTo(BeNil())
				Expect(workerSpec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet).NotTo(BeNil())
				Expect(workerSpec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Port.IntVal).To(Equal(int32(8080)))
			})

			DescribeTable("default restart policy",
				func(jobType string, expectedPolicy kubeflowcommon.RestartPolicy) {
					builder.MergedWorkload.Spec.JobType = jobType

					err := builder.BuildWorkerReplicaSpec(ctx, workerSpec, 0)
					Expect(err).NotTo(HaveOccurred())
					Expect(workerSpec.RestartPolicy).To(Equal(expectedPolicy))
				},
				// empty policy == framework default in training/mpi-operator
				Entry("TF", constants.KindTFJob, kubeflowcommon.RestartPolicy("")),
				Entry("MPI", constants.KindMpiJob, kubeflowcommon.RestartPolicy("")),
				Entry("PyTorch", constants.KindPyTorchJob, kubeflowcommon.RestartPolicy("")),
				Entry("XGBoost", constants.KindXGBoostJob, kubeflowcommon.RestartPolicy("")),
			)
		})

		Context("Nodepools annotation", func() {
			const (
				nodepoolsList = "pool-a pool-b"
				nodepool      = "pool-c"
			)

			BeforeEach(func() {
				scheme := runtime.NewScheme()
				register.InitSchema(scheme)
				nodepoolA := resource.CreateNodePool("pool-a", "pool", "a")
				nodepoolB := resource.CreateNodePool("pool-b", "pool", "b")
				nodepoolC := resource.CreateNodePool("pool-c", "pool", "c")
				nodepoolAKai := resource.CreateKaiNodePool("pool-a", "pool", "a")
				nodepoolBKai := resource.CreateKaiNodePool("pool-b", "pool", "b")
				nodepoolCKai := resource.CreateKaiNodePool("pool-c", "pool", "c")

				builder.K8sClient = fakeClient.InitFakeClientWithObjectNoStatus(scheme, nodepoolA, nodepoolB, nodepoolC, nodepoolAKai, nodepoolBKai, nodepoolCKai)
				builder.MergedWorkload.Spec = v2alpha1.DistributedSpec{}
			})

			Context("When has no nodepools", func() {
				It("when has no nodepools", func() {
					Expect(builder.BuildWorkerReplicaSpec(ctx, workerSpec, 0)).To(Succeed())

					_, exist := workerSpec.Template.Annotations[commonconstants.AnnotationNodepools]
					Expect(exist).To(BeFalse())
				})
			})

			Context("When has nodepools", func() {
				It("when has Nodepools", func() {
					builder.MergedWorkload.Spec.NodePools = flds.NewOptionalStringField(nodepoolsList, "")
					Expect(builder.BuildWorkerReplicaSpec(ctx, workerSpec, 0)).To(Succeed())

					value, exist := workerSpec.Template.Annotations[commonconstants.AnnotationNodepools]
					Expect(exist).To(BeTrue())
					Expect(value).To(Equal(nodepoolsList))
				})

				It("when has Nodepool", func() {
					builder.MergedWorkload.Spec.NodePool = flds.NewOptionalStringField(nodepool, "")
					Expect(builder.BuildWorkerReplicaSpec(ctx, workerSpec, 0)).To(Succeed())

					value, exist := workerSpec.Template.Annotations[commonconstants.AnnotationNodepools]
					Expect(exist).To(BeTrue())
					Expect(value).To(Equal(nodepool))
				})

				It("when has Nodepools and Nodepool - uses Nodepools", func() {
					builder.MergedWorkload.Spec.NodePools = flds.NewOptionalStringField(nodepoolsList, "")
					builder.MergedWorkload.Spec.NodePool = flds.NewOptionalStringField(nodepool, "")
					Expect(builder.BuildWorkerReplicaSpec(ctx, workerSpec, 0)).To(Succeed())

					value, exist := workerSpec.Template.Annotations[commonconstants.AnnotationNodepools]
					Expect(exist).To(BeTrue())
					Expect(value).To(Equal(nodepoolsList))
				})
			})
		})
	})
})
