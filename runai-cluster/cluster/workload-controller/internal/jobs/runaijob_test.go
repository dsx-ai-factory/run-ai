package jobs

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clusterCommon "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	runaiJobV1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/common"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	core1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("RunaiJob status Adapter", func() {

	var (
		job       *runaiJobV1.RunaiJob
		workload  *jobs.JobWorkload
		jobStatus *RunaiJobStatus
	)

	Context("Job Finished Successfully", func() {
		BeforeEach(func() {
			workload = &jobs.JobWorkload{}
			workload.Status = jobs.JobStatus{}
			job = &runaiJobV1.RunaiJob{}

			job.Status.CompletionTime = &metav1.Time{Time: time.Now()}
			job.Status.StartTime = &metav1.Time{Time: time.Now().Add(-time.Hour)}
			job.Status.Conditions = []runaiJobV1.JobCondition{{Type: runaiJobV1.JobFailed, Status: "False"},
				{Type: runaiJobV1.JobComplete, Status: "True"}, {Type: runaiJobV1.JobSuspended, Status: "False"}}
			job.Status.Succeeded = 1
			job.Status.Failed = 0
			job.Status.Active = 0
			jobStatus = NewRunaiJobStatus(job)
		})
		It("Job is not Running", func() {
			Expect(jobStatus.IsRunning()).To(BeFalse())
		})
		It("Job is not Pending", func() {
			Expect(jobStatus.IsPending()).To(BeFalse())
		})
		It("Job is not Suspended", func() {
			Expect(jobStatus.IsSuspended()).To(BeFalse())
		})
		It("Job Finished", func() {
			Expect(jobStatus.IsJobFinished()).To(BeTrue())
		})
		It("Job Failed", func() {
			Expect(jobStatus.IsJobFailed()).To(BeFalse())
		})
		It("Job CompleteTime is Not Zero", func() {
			Expect(jobStatus.GetCompletionTime().IsZero()).To(BeFalse())
		})
		It("Job is not creating", func() {
			Expect(jobStatus.IsCreating()).To(BeFalse())
		})
		It("Number of successful pods is 1", func() {
			Expect(jobStatus.GetSucceededPods()).To(Equal(int32(1)))
			Expect(jobStatus.GetFailedPods()).To(Equal(int32(0)))
			Expect(jobStatus.GetPendingPods()).To(Equal(int32(0)))
			Expect(jobStatus.GetReadyPods()).To(Equal(int32(0)))
		})
	})
	Context("Job Failed", func() {
		BeforeEach(func() {
			workload = &jobs.JobWorkload{}
			workload.Status = jobs.JobStatus{}
			job = &runaiJobV1.RunaiJob{}
			job.Status.StartTime = &metav1.Time{Time: time.Now().Add(-time.Hour)}
			job.Status.Conditions = []runaiJobV1.JobCondition{{Type: runaiJobV1.JobFailed, Status: "True", Message: "Job Failed"}}

			job.Status.Succeeded = 0
			job.Status.Failed = 1
			job.Status.Active = 0
			jobStatus = NewRunaiJobStatus(job)
		})
		Context("Job Failed with zero time condition", func() {
			BeforeEach(func() {
				timeNow := metav1.Time{Time: time.Now()}
				job.Status.Conditions = []runaiJobV1.JobCondition{{Type: runaiJobV1.JobFailed, Status: "True", Message: "Job Failed", LastTransitionTime: timeNow}}
			})
			It("completion time is not zero", func() {
				Expect(jobStatus.GetCompletionTime().IsZero()).To(BeFalse())
			})
			It("Completion time should not change between iteration", func() {
				Expect(jobStatus.GetCompletionTime().Equal(jobStatus.GetCompletionTime())).To(BeTrue())
			})
		})
		It("completion time is not zero", func() {
			Expect(jobStatus.GetCompletionTime().IsZero()).To(BeFalse())
		})
		It("Completion time should not change between iteration", func() {
			Expect(jobStatus.GetCompletionTime().Equal(jobStatus.GetCompletionTime())).To(BeTrue())
		})
		It("Job is not Running", func() {
			Expect(jobStatus.IsRunning()).To(BeFalse())
		})
		It("Job is not Pending", func() {
			Expect(jobStatus.IsPending()).To(BeFalse())
		})
		It("Job is not Suspended", func() {
			Expect(jobStatus.IsSuspended()).To(BeFalse())
		})
		It("Job is not creating", func() {
			Expect(jobStatus.IsCreating()).To(BeFalse())
		})

		It("Job Finished", func() {
			Expect(jobStatus.IsJobFinished()).To(BeTrue())
		})
		It("Job Failed", func() {
			Expect(jobStatus.IsJobFailed()).To(BeTrue())
			Expect(jobStatus.GetFailedInfo()).To(Equal("Job Failed"))
		})

	})
	Context("Job Pending with podgroup status", func() {
		BeforeEach(func() {
			workload = &jobs.JobWorkload{}
			workload.Status = jobs.JobStatus{}
			job = &runaiJobV1.RunaiJob{}
			job.Status.StartTime = &metav1.Time{Time: time.Now().Add(-time.Hour)}
			job.Status.Conditions = []runaiJobV1.JobCondition{}
			job.Status.Pending = 1
			jobStatus = NewRunaiJobStatus(job)
		})
		It("Job is not Suspended", func() {
			Expect(jobStatus.IsSuspended()).To(BeFalse())
		})

		It("Job is not Finished", func() {
			Expect(jobStatus.IsJobFinished()).To(BeFalse())
		})
		It("Job is not Failed", func() {
			Expect(jobStatus.IsJobFailed()).To(BeFalse())
		})
		It("Job Pending", func() {
			Expect(jobStatus.IsPending()).To(BeTrue())
			Expect(jobStatus.Status.Pending).To(Equal(int32(1)))
		})
		It("Job is not Running", func() {
			Expect(jobStatus.IsRunning()).To(BeFalse())
		})
	})
	Context("Job Pending without podgroup status", func() {
		BeforeEach(func() {
			workload = &jobs.JobWorkload{}
			workload.Status = jobs.JobStatus{}
			job = &runaiJobV1.RunaiJob{}
			job.Status.StartTime = &metav1.Time{Time: time.Now().Add(-time.Hour)}
			job.Status.Conditions = []runaiJobV1.JobCondition{}
			jobStatus = NewRunaiJobStatus(job)
		})
		It("Job is not Suspended", func() {
			Expect(jobStatus.IsSuspended()).To(BeFalse())
		})
		It("Job is creating", func() {
			Expect(jobStatus.IsCreating()).To(BeTrue())
			Expect(jobStatus.GetPendingPods()).To(Equal(int32(0)))
		})
		It("Job is not Finished", func() {
			Expect(jobStatus.IsJobFinished()).To(BeFalse())
		})
		It("Job is not Failed", func() {
			Expect(jobStatus.IsJobFailed()).To(BeFalse())
		})
		It("Job is not Running", func() {
			Expect(jobStatus.IsRunning()).To(BeFalse())
		})
	})
	Context("Non valid annotation", func() {
		BeforeEach(func() {
			workload = &jobs.JobWorkload{}
			workload.Status = jobs.JobStatus{}
			job = &runaiJobV1.RunaiJob{}
			job.Status.StartTime = &metav1.Time{Time: time.Now().Add(-time.Hour)}
			job.Status.Conditions = []runaiJobV1.JobCondition{}
			job.Status.Active = 1
			job.Annotations = map[string]string{common.RunaiCurrentAllocatedGpuMemoryAnnotation: "223ddddd1200"}
			jobStatus = NewRunaiJobStatus(job)
		})
		It("Number of pennding pods should be zero", func() {
			Expect(jobStatus.GetPendingPods()).To(Equal(int32(0)))
		})
		It("Number of running pods should be zero", func() {
			Expect(jobStatus.GetReadyPods()).To(Equal(int32(0)))
		})
		It("Number of allocated gpus memory should be zero", func() {
			Expect(jobStatus.GetAllocatedGPUMemory()).To(Equal(int64(0)))
		})
		It("Number of allocated gpus should be empty string", func() {
			Expect(jobStatus.GetAllocatedGPU()).To(Equal(""))
		})
	})
	Context("Job Running", func() {
		BeforeEach(func() {
			workload = &jobs.JobWorkload{}
			workload.Status = jobs.JobStatus{}
			job = &runaiJobV1.RunaiJob{}
			job.Status.StartTime = &metav1.Time{Time: time.Now().Add(-time.Hour)}
			job.Status.Conditions = []runaiJobV1.JobCondition{}
			job.Status.Active = 1
			job.Status.Ready = 1
			job.Annotations = map[string]string{common.RunaiCurrentAllocatedGpuMemoryAnnotation: "1200",
				common.RunaiCurrentAllocatedGpusAnnotation: "1"}
			jobStatus = NewRunaiJobStatus(job)
		})
		It("Job is not Finished", func() {
			Expect(jobStatus.IsJobFinished()).To(BeFalse())
		})
		It("Job is not Suspended", func() {
			Expect(jobStatus.IsSuspended()).To(BeFalse())
		})
		It("Job is not Failed", func() {
			Expect(jobStatus.IsJobFailed()).To(BeFalse())
		})
		It("Job is not Pending", func() {
			Expect(jobStatus.IsPending()).To(BeFalse())
		})
		It("Job is Running", func() {
			Expect(jobStatus.IsRunning()).To(BeTrue())
		})
		It("Job have 1 running pod", func() {
			Expect(jobStatus.GetReadyPods()).To(Equal(int32(1)))
		})
		It("Job should have 1 gpu", func() {
			Expect(jobStatus.GetAllocatedGPU()).To(Equal("1"))
		})
		It("Job should have 1200 gpu memory", func() {
			Expect(jobStatus.GetAllocatedGPUMemory()).To(Equal(int64(1200)))
		})
	})
	Context("Job Suspended", func() {
		BeforeEach(func() {
			workload = &jobs.JobWorkload{}
			workload.Status = jobs.JobStatus{}
			job = &runaiJobV1.RunaiJob{}
			job.Status.StartTime = &metav1.Time{Time: time.Now().Add(-time.Hour)}
			job.Status.Conditions = []runaiJobV1.JobCondition{{Type: runaiJobV1.JobSuspended, Status: "True", Message: "Suspended"}}
			job.Status.Active = 0
			job.Status.Pending = 0
			jobStatus = NewRunaiJobStatus(job)
		})
		It("Job is Suspended", func() {
			Expect(jobStatus.IsSuspended()).To(BeTrue())
			Expect(jobStatus.GetSuspendInfo()).To(Equal("Suspended"))
		})
		It("Job is not Creating", func() {
			Expect(jobStatus.IsCreating()).To(BeFalse())
		})
		It("Job is not pending", func() {
			Expect(jobStatus.IsPending()).To(BeFalse())
		})
		It("Job is not Finished", func() {
			Expect(jobStatus.IsJobFinished()).To(BeFalse())
		})
		It("Job is not Failed", func() {
			Expect(jobStatus.IsJobFailed()).To(BeFalse())
		})
		It("Job is not Running", func() {
			Expect(jobStatus.IsRunning()).To(BeFalse())
		})
		It("Job have 0 ready pod", func() {
			Expect(jobStatus.GetReadyPods()).To(Equal(int32(0)))
		})
	})
	Context("Job is Initializing", func() {
		BeforeEach(func() {
			workload = &jobs.JobWorkload{}
			workload.Status = jobs.JobStatus{}
			job = &runaiJobV1.RunaiJob{}
			job.Status.StartTime = &metav1.Time{Time: time.Now().Add(-time.Hour)}
			job.Status.Conditions = []runaiJobV1.JobCondition{{Type: runaiJobV1.JobInitializing, Status: "True", Message: "Initializing"}}
			job.Status.Active = 0
			job.Status.Initialized = 1
			job.Status.Running = 1
			job.Status.Pending = 0
			jobStatus = NewRunaiJobStatus(job)
		})
		It("Job is not Suspended", func() {
			Expect(jobStatus.IsSuspended()).To(BeFalse())
		})
		It("Job is not pending", func() {
			Expect(jobStatus.IsPending()).To(BeFalse())
		})
		It("Job is not Finished", func() {
			Expect(jobStatus.IsJobFinished()).To(BeFalse())
		})
		It("Job is not Failed", func() {
			Expect(jobStatus.IsJobFailed()).To(BeFalse())
		})
		It("Job is not Running", func() {
			Expect(jobStatus.IsRunning()).To(BeFalse())
		})
		It("Job have 0 ready pod", func() {
			Expect(jobStatus.GetReadyPods()).To(Equal(int32(0)))
		})
		It("Job is Initializing", func() {
			Expect(jobStatus.IsInitializing()).To(BeTrue())
		})
	})

	Context("Job is nil", func() {
		BeforeEach(func() {
			jobStatus = NewRunaiJobStatus(nil)
		})
		It("Job is not Suspended", func() {
			Expect(jobStatus.IsSuspended()).To(BeFalse())
		})
		It("Job is not Pending", func() {
			Expect(jobStatus.IsPending()).To(BeFalse())
		})
		It("Job is not Finished", func() {
			Expect(jobStatus.IsJobFinished()).To(BeFalse())
		})
		It("Job is not Failed", func() {
			Expect(jobStatus.IsJobFailed()).To(BeFalse())
		})
		It("Job is not Running", func() {
			Expect(jobStatus.IsRunning()).To(BeFalse())
		})
		It("Job have 0 ready pod", func() {
			Expect(jobStatus.GetReadyPods()).To(Equal(int32(0)))
		})
		It("Job is not Initializing", func() {
			Expect(jobStatus.IsInitializing()).To(BeFalse())
		})
		It("Job is Creating", func() {
			Expect(jobStatus.IsCreating()).To(BeTrue())
		})
		It("Job is not Terminating", func() {
			Expect(jobStatus.IsTerminating()).To(BeFalse())
		})
	})

	Context("edge cases", func() {
		It("Completed job with pending pods is not Pending", func() {
			job = &runaiJobV1.RunaiJob{}
			job.Status.CompletionTime = &metav1.Time{Time: time.Now()}
			job.Status.Pending = 1
			jobStatus = NewRunaiJobStatus(job)
			Expect(jobStatus.IsPending()).To(BeFalse())
		})
		It("Completed job with ready pods is not Running", func() {
			job = &runaiJobV1.RunaiJob{}
			job.Status.CompletionTime = &metav1.Time{Time: time.Now()}
			job.Status.Ready = 1
			jobStatus = NewRunaiJobStatus(job)
			Expect(jobStatus.IsRunning()).To(BeFalse())
		})
		It("Completed job with initialized pods is not Initializing", func() {
			job = &runaiJobV1.RunaiJob{}
			job.Status.CompletionTime = &metav1.Time{Time: time.Now()}
			job.Status.Initialized = 1
			job.Status.Ready = 0
			jobStatus = NewRunaiJobStatus(job)
			Expect(jobStatus.IsInitializing()).To(BeFalse())
		})
	})
})

const (
	name      = "test"
	namespace = "default"
)

var _ = Describe("RunaiJobAdapter", func() {
	var (
		runaiJob *runaiJobV1.RunaiJob
		adapter  *RunaiJobAdapter
	)

	BeforeEach(func() {
		runaiJob = &runaiJobV1.RunaiJob{}
		adapter = NewRunaiJobAdapter(runaiJob)
	})

	It("returns correct job kind", func() {
		Expect(adapter.GetJobKind()).To(Equal(clusterCommon.KindRunaiJob))
	})

	It("returns correct spec", func() {
		Expect(adapter.GetSpec()).To(BeAssignableToTypeOf(&RunaiJobSpec{}))
	})

	It("returns correct status", func() {
		Expect(adapter.GetStatus()).To(BeAssignableToTypeOf(&RunaiJobStatus{}))
	})

	It("returns correct concurrent job", func() {
		Expect(adapter.GetConcurrentJob()).To(Equal(runaiJob))
	})
})

var _ = Describe("RunaiJobSpec", func() {
	var (
		runaiJob *runaiJobV1.RunaiJob
		spec     *RunaiJobSpec
	)

	BeforeEach(func() {
		runaiJob = resource.CreateDumbRunaiJob(name, namespace, "", "", runaiJobV1.JobSpec{
			Parallelism: ptr.To(int32(1)),
			Template: core1.PodTemplateSpec{

				Spec: core1.PodSpec{
					Volumes:        nil,
					InitContainers: nil,
					Containers: []core1.Container{
						{
							Name:  "test",
							Image: "test",
						},
					},

					Affinity: &core1.Affinity{
						NodeAffinity: &core1.NodeAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []core1.PreferredSchedulingTerm{
								{
									Weight: 1,
									Preference: core1.NodeSelectorTerm{
										MatchExpressions: []core1.NodeSelectorRequirement{},
									},
								},
							},
						},
					},
				},
			},
		})
		spec = &RunaiJobSpec{RunaiJob: runaiJob}
	})

	It("returns correct pod template spec", func() {
		Expect(spec.GetPodTemplateSpec()).To(Equal(&runaiJob.Spec.Template))
	})

	It("returns correct replicas", func() {
		Expect(spec.GetReplicas()).To(Equal(runaiJob.Spec.Parallelism))
	})

	It("returns correct replica type specs", func() {
		Expect(spec.GetReplicaTypeSpecs()).To(HaveKey(constants.Worker))
	})

	It("returns correct replica spec for worker", func() {
		Expect(spec.GetReplicaSpec(constants.Worker)).To(Equal(spec))
	})

	It("returns nil replica spec for non-worker", func() {
		Expect(spec.GetReplicaSpec(constants.Master)).To(BeNil())
	})
})
