package jobs_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("JobWorkload", func() {
	Context("isStatusChangedFunc", func() {
		var (
			originalWorkload *jobs.JobWorkload
			modifiedWorkload *jobs.JobWorkload
		)
		BeforeEach(func() {
			originalWorkload = createJobWorkloadStatus()
			modifiedWorkload = createJobWorkloadStatus()
		})
		It("should return false when status is not changed at all", func() {
			modifiedWorkload.Status = originalWorkload.Status
			Expect(modifiedWorkload.IsStatusChanged(originalWorkload)).To(BeFalse())
		})
		It("should return true when status is changed", func() {
			completeTime := metav1.Now()
			modifiedWorkload.Status.WorkloadCommonStatus.CompletionTime = &completeTime
			Expect(modifiedWorkload.IsStatusChanged(originalWorkload)).To(BeTrue())
		})
		It("should return false when only last reconcile time have changed", func() {
			lastReconcileTime := metav1.Now()
			modifiedWorkload.Status.LastStatusChangeTime = &lastReconcileTime
			Expect(modifiedWorkload.IsStatusChanged(originalWorkload)).To(BeFalse())
		})
	})

	Context("IsUsingNodePools", func() {
		DescribeTable("with different node pool values",
			func(nodePoolValue string, expected bool) {
				spec := &jobs.JobSpec{
					RunaiSchedulerFields: field_groups.RunaiSchedulerFields{
						NodePools: field_types.NewOptionalStringField(nodePoolValue, ""),
					},
				}
				Expect(jobs.IsUsingNodePools(spec)).To(Equal(expected))
			},
			Entry("empty node pools", "", false),
			Entry("non-empty node pools", "some-node-pool", true),
		)

		It("should return false for nil spec", func() {
			Expect(jobs.IsUsingNodePools(nil)).To(Equal(false))
		})

		// Test confirms that the bug is fixed in IsUsingNodePools
		It("should handle empty node pools correctly", func() {
			// Create a spec with NodePools set to empty string
			spec := &jobs.JobSpec{
				RunaiSchedulerFields: field_groups.RunaiSchedulerFields{
					NodePools: field_types.NewOptionalStringField("", ""),
				},
			}

			// With the bug fixed, empty NodePools should return false
			actual := jobs.IsUsingNodePools(spec)
			Expect(actual).To(Equal(false), "IsUsingNodePools should return false for empty node pools")
		})
	})

	Context("IsElasticPytorchJob", func() {
		DescribeTable("with different replica configurations",
			func(minReplicas, maxReplicas *int, expected bool) {
				job := resource.CreateElasticPyTorchJob(minReplicas, maxReplicas)

				Expect(jobs.IsElasticPytorchJob(job)).To(Equal(expected))
			},
			Entry("not elastic (nil min and max)", nil, nil, false),
			Entry("elastic with min replicas only", ptr.To(1), nil, true),
			Entry("elastic with max replicas only", nil, ptr.To(3), true),
			Entry("elastic with both min and max replicas", ptr.To(1), ptr.To(3), true),
		)

		It("should return false for non-PytorchJob", func() {
			nonPytorchJob := &jobs.JobWorkload{
				Spec: jobs.JobSpec{
					DistributedFields: field_groups.DistributedFields{
						JobType: "non-pytorch",
					},
				},
			}
			Expect(jobs.IsElasticPytorchJob(nonPytorchJob)).To(Equal(false))
		})

		It("should return false for nil job", func() {
			Expect(jobs.IsElasticPytorchJob(nil)).To(Equal(false))
		})
	})
})

func createJobWorkloadStatus() *jobs.JobWorkload {
	return &jobs.JobWorkload{
		Status: jobs.JobStatus{
			WorkloadCommonStatus: common.WorkloadCommonStatus{
				Conditions: []common.WorkloadCondition{
					{
						Type:    common.Failed,
						Status:  v1.ConditionFalse,
						Message: "Failed to build workload",
						Reason:  common.WorkloadBuildFailedReason,
					},
					{
						Type:    common.Succeeded,
						Status:  v1.ConditionTrue,
						Message: "Workload is completed",
					},
					{
						Type:    common.Ready,
						Status:  v1.ConditionTrue,
						Message: "Workload is ready",
					},
				},
				SucceededPods: 1,
				RunningPods:   2,
				PendingPods:   0,
				FailedPods:    1,
				ExternalUrls:  []string{"http://external-url"},
			},
			WorkloadDebugStatus: common.WorkloadDebugStatus{
				ObservedGeneration: 1,

				OwnedResources: []common.ObservedResource{
					{
						Resource: common.Resource{
							Name: "pod1",
							Kind: metav1.GroupVersionKind{
								Kind: "Job",
							},
						},
						ObservedGeneration: 1,
						Uid:                "1",
					},
				},
				LastStatusChangeTime: nil,
			},
		},
	}
}
