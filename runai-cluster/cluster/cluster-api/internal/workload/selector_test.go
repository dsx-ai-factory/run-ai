/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload_test

import (
	"fmt"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/common"
	infw "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	"k8s.io/utils/ptr"

	kubeflowcommon "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/workload"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	latestReadyRevisionName = "workload-0"
)

var _ = Describe("Workload Selectors", func() {
	Describe("GetSelectorCreator", func() {
		It("should return FirstPodSelector for nil workload", func() {
			selector := workload.GetSelectorCreator(nil)
			Expect(selector).To(BeNil())
		})

		DescribeTable("should return correct selector for different job types",
			func(jobType api.JobType, expectedLabelKey, expectedLabelValue string) {
				w := &workload.Workload{
					JobType: jobType, Namespace: "test-namespace", Original: &infw.InferenceWorkload{
						Status: infw.InferenceWorkloadStatus{
							InferenceStatus: common.InferenceStatus{
								LatestReadyRevisionName: ptr.To(latestReadyRevisionName),
							},
						},
					},
				}
				selector := workload.GetSelectorCreator(w)
				kvSelector, ok := selector.(*workload.KeyValueSelector)
				Expect(ok).To(BeTrue())
				Expect(kvSelector.String()).To(Equal(fmt.Sprintf("KeyValueSelector: LabelKey=%s, LabelValue=%s", expectedLabelKey, expectedLabelValue)))
			},
			Entry("MPIJob", api.MPIJob, kubeflowcommon.JobRoleLabel, "launcher"),
			Entry("PyTorchJob", api.PyTorchJob, kubeflowcommon.JobRoleLabel, "master"),
			Entry("XGBoost", api.XGBoostJob, kubeflowcommon.JobRoleLabel, "master"),
			Entry("TFJob", api.TFJob, kubeflowcommon.JobRoleLabel, "master"),
			Entry("JAX", api.JAXJob, kubeflowcommon.ReplicaIndexLabel, "0"),
			Entry("InferenceJob", api.InferenceJob, "serving.knative.dev/revision", latestReadyRevisionName),
			Entry("DistributedInferenceJob", api.DistributedInferenceJob, "leaderworkerset.sigs.k8s.io/worker-index", "0"),
		)

		It("should return FirstPodSelector for unknown job type", func() {
			w := &workload.Workload{JobType: "UnknownJobType", Namespace: "test-namespace"}
			selector := workload.GetSelectorCreator(w)
			Expect(selector).To(BeAssignableToTypeOf(&workload.FirstPodSelector{}))
		})
	})

	Describe("FallBackSelectorCreator", func() {
		It("should return nil for nil workload", func() {
			selector := workload.FallBackSelectorCreator(nil)
			Expect(selector).To(BeNil())
		})

		DescribeTable("should return correct fallback selector string representation for different job types",
			func(jobType api.JobType, expectedSelectorString string) {
				w := &workload.Workload{JobType: jobType, Namespace: "test-namespace"}
				selector := workload.FallBackSelectorCreator(w)
				Expect(selector).NotTo(BeNil())
				Expect(selector.String()).To(Equal(expectedSelectorString))
			},
			Entry("MPIJob", api.MPIJob,
				fmt.Sprintf("IndexRolePodSelector: indexKey=%s, indexValue=%s, roleKey=%s, roleValue=%s",
					kubeflowcommon.ReplicaIndexLabel, "0", kubeflowcommon.JobRoleLabel, "worker")),
			Entry("PyTorchJob", api.PyTorchJob,
				fmt.Sprintf("IndexRolePodSelector: indexKey=%s, indexValue=%s, roleKey=%s, roleValue=%s",
					kubeflowcommon.ReplicaIndexLabel, "0", kubeflowcommon.ReplicaTypeLabel, "worker")),
			Entry("XGBoost", api.XGBoostJob,
				fmt.Sprintf("IndexRolePodSelector: indexKey=%s, indexValue=%s, roleKey=%s, roleValue=%s",
					kubeflowcommon.ReplicaIndexLabel, "0", kubeflowcommon.ReplicaTypeLabel, "worker")),
			Entry("TFJob", api.TFJob,
				fmt.Sprintf("IndexRolePodSelector: indexKey=%s, indexValue=%s, roleKey=%s, roleValue=%s",
					kubeflowcommon.ReplicaIndexLabel, "0", kubeflowcommon.ReplicaTypeLabel, "worker")),
			Entry("InferenceJob", api.InferenceJob,
				"FirstPodSelector"),
		)

		It("should return nil for non-distributed job type", func() {
			w := &workload.Workload{JobType: "UnknownJobType", Namespace: "test-namespace"}
			selector := workload.FallBackSelectorCreator(w)
			Expect(selector).To(BeNil())
		})
		It("should return nil for JAX job type", func() {
			w := &workload.Workload{JobType: api.JAXJob, Namespace: "test-namespace"}
			selector := workload.FallBackSelectorCreator(w)
			Expect(selector).To(BeNil())
		})
		It("should return nil for DistributedInferenceJob - leader pod must exist", func() {
			w := &workload.Workload{JobType: api.DistributedInferenceJob, Namespace: "test-namespace"}
			selector := workload.FallBackSelectorCreator(w)
			Expect(selector).To(BeNil())
		})
	})

	Describe("FirstPodSelector", func() {
		It("should create a valid selector", func() {
			selector := workload.NewFirstPodSelector(&workload.Workload{Namespace: "test-namespace", ClusterUUID: "test-uid"})
			listOption := selector.CreateSelector()
			Expect(listOption).NotTo(BeNil())

			labelSelector, ok := listOption.(*client.ListOptions)
			Expect(ok).To(BeTrue())
			Expect(labelSelector.LabelSelector).NotTo(BeNil())
			Expect(labelSelector.Limit).To(Equal(int64(1)))
			Expect(labelSelector.Namespace).To(Equal("test-namespace"))

			requirements, _ := labelSelector.LabelSelector.Requirements()
			Expect(requirements).To(HaveLen(1))
			Expect(requirements[0].Key()).To(Equal(constants.ClusterTopOwnerUidLabel))
			Expect(requirements[0].Values().List()).To(ContainElement("test-uid"))
		})

		It("should have correct string representation", func() {
			selector := &workload.FirstPodSelector{}
			Expect(selector.String()).To(Equal("FirstPodSelector"))
		})
	})

	Describe("KeyValueSelector", func() {
		It("should create a valid selector", func() {
			w := &workload.Workload{Namespace: "test-namespace", ClusterUUID: "test-uid"}
			selector := workload.NewKeyValueSelector("testKey", "testValue", w)
			listOption := selector.CreateSelector()
			Expect(listOption).NotTo(BeNil())

			labelSelector, ok := listOption.(*client.ListOptions)
			Expect(ok).To(BeTrue())
			Expect(labelSelector.LabelSelector).NotTo(BeNil())
			Expect(labelSelector.Limit).To(Equal(int64(1)))
			Expect(labelSelector.Namespace).To(Equal("test-namespace"))

			requirements, _ := labelSelector.LabelSelector.Requirements()
			Expect(requirements).To(HaveLen(2))
			for _, req := range requirements {
				Expect(req.Key()).To(SatisfyAny(Equal(constants.ClusterTopOwnerUidLabel), Equal("testKey")))
				Expect(req.Values().List()).To(SatisfyAny(ContainElement("testValue"), ContainElement("test-uid")))
			}
		})

		It("should have correct string representation", func() {
			w := &workload.Workload{}
			selector := workload.NewKeyValueSelector("testKey", "testValue", w)
			Expect(selector.String()).To(Equal("KeyValueSelector: LabelKey=testKey, LabelValue=testValue"))
		})
	})

	Describe("IndexRolePodSelector", func() {
		It("should create a valid selector", func() {
			w := &workload.Workload{Namespace: "test-namespace", ClusterUUID: "test-uid"}
			selector := workload.NewIndexRolePodSelector("indexKey", "indexValue", "roleKey", "roleValue", w)
			listOption := selector.CreateSelector()
			Expect(listOption).NotTo(BeNil())

			labelSelector, ok := listOption.(*client.ListOptions)
			Expect(ok).To(BeTrue())
			Expect(labelSelector.LabelSelector).NotTo(BeNil())
			Expect(labelSelector.Limit).To(Equal(int64(1)))
			Expect(labelSelector.Namespace).To(Equal("test-namespace"))

			requirements, _ := labelSelector.LabelSelector.Requirements()
			Expect(requirements).To(HaveLen(3))
			for _, req := range requirements {
				Expect(req.Key()).To(SatisfyAny(Equal("indexKey"), Equal("roleKey"),
					Equal(constants.ClusterTopOwnerUidLabel)))
				Expect(req.Values().List()).To(SatisfyAny(ContainElement("indexValue"),
					ContainElement("roleValue"), ContainElement("test-uid")))
			}
		})

		It("should have correct string representation", func() {
			w := &workload.Workload{}
			selector := workload.NewIndexRolePodSelector("indexKey", "indexValue", "roleKey", "roleValue", w)
			Expect(selector.String()).To(Equal("IndexRolePodSelector: indexKey=indexKey, indexValue=indexValue, roleKey=roleKey, roleValue=roleValue"))
		})
	})

	Describe("GetLatestReadyRevision", func() {
		It("should return the latest ready revision name when it exists", func() {
			w := &infw.InferenceWorkload{
				Status: infw.InferenceWorkloadStatus{
					InferenceStatus: common.InferenceStatus{
						LatestReadyRevisionName: ptr.To("revision-123"),
					},
				},
			}

			revision, err := workload.GetLatestReadyRevision(w)
			Expect(err).ToNot(HaveOccurred())
			Expect(revision).To(Equal("revision-123"))
		})

		It("should return error when workload is nil", func() {
			revision, err := workload.GetLatestReadyRevision(nil)
			Expect(err).To(HaveOccurred())
			Expect(revision).To(BeEmpty())
		})

		It("should return error when revision name is empty", func() {
			w := &infw.InferenceWorkload{
				Status: infw.InferenceWorkloadStatus{
					InferenceStatus: common.InferenceStatus{
						LatestReadyRevisionName: nil, // Represents empty/missing in this context for the test
					},
				},
			}

			revision, err := workload.GetLatestReadyRevision(w)
			Expect(err).To(HaveOccurred())
			Expect(revision).To(BeEmpty())

			wEmpty := &infw.InferenceWorkload{
				Status: infw.InferenceWorkloadStatus{
					InferenceStatus: common.InferenceStatus{
						LatestReadyRevisionName: ptr.To(""),
					},
				},
			}
			revisionEmpty, errEmpty := workload.GetLatestReadyRevision(wEmpty)
			Expect(errEmpty).To(HaveOccurred())
			Expect(revisionEmpty).To(BeEmpty())
		})
	})
})
