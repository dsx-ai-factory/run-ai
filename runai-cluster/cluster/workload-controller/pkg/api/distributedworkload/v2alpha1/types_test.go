package v2alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	dp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	apiVersion = "run.ai/v2alpha1"
	name       = "test"
)

var _ = Describe("DistributedWorkload", func() {
	var (
		dw          *v2alpha1.DistributedWorkload
		jobWorkload *jobs.JobWorkload
	)

	BeforeEach(func() {

		dw = &v2alpha1.DistributedWorkload{
			TypeMeta: metav1.TypeMeta{
				Kind:       constants.KindDistributedWorkload,
				APIVersion: apiVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Status: v2alpha1.DistributedWorkloadStatus{

				WorkloadDebugStatus: common.WorkloadDebugStatus{
					ObservedGeneration:    0,
					FirstAcknowledgedTime: nil,
					OwnedResources:        nil,
					LastStatusChangeTime:  nil,
				},
			},
		}
		jobWorkload = &jobs.JobWorkload{}
	})

	Describe("CopyTo", func() {
		Context("when jobWorkload is nil", func() {
			It("does not panic", func() {
				Expect(func() { dw.CopyTo(nil) }).ShouldNot(Panic())
			})
		})

		Context("when jobWorkload is not nil", func() {
			It("copies fields correctly", func() {
				dw.CopyTo(jobWorkload)
				Expect(jobWorkload.TypeMeta).To(Equal(dw.TypeMeta))
				Expect(jobWorkload.ObjectMeta).To(Equal(dw.ObjectMeta))
				Expect(jobWorkload.Status.WorkloadStatus).To(Equal(dw.Status.WorkloadStatus))
				Expect(jobWorkload.Status.WorkloadDebugStatus).To(Equal(dw.Status.WorkloadDebugStatus))
			})

			Context("when MasterSpec is not nil", func() {
				BeforeEach(func() {
					dw.MasterSpec = &dp.DistributedMasterSpec{}
				})

				It("copies MasterSpec", func() {
					dw.CopyTo(jobWorkload)
					Expect(jobWorkload.MasterSpec).ShouldNot(BeNil())
				})
			})

			Context("when MasterSpec is nil", func() {
				BeforeEach(func() {
					dw.MasterSpec = nil
				})

				It("does not create MasterSpec", func() {
					dw.CopyTo(jobWorkload)
					Expect(jobWorkload.MasterSpec).Should(BeNil())
				})
			})
		})
	})
})
