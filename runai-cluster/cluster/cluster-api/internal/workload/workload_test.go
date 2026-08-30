package workload

import (
	"fmt"

	distinfwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"

	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	infwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dp "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"

	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
)

var _ = Describe("Workload tests", func() {
	Context("getJobType", func() {
		var mockWorkload *dwv2.DistributedWorkload

		BeforeEach(func() {
			mockWorkload = &dwv2.DistributedWorkload{
				Spec: dp.DistributedSpec{},
			}
		})
		Context("Nil workload", func() {
			It("should return error for nil workload", func() {
				jobType, err := getJobType(nil, api.Distributed)
				Expect(err).To(HaveOccurred())
				Expect(jobType).To(Equal(api.Unknown))
				Expect(err.Error()).To(Equal(workloadIsNilErrorMsg))
			})
		})
		Context("Non distributed workload", func() {
			It("should return InferenceJob for Inference workload type", func() {
				jobType, err := getJobType(&iwv2.InteractiveWorkload{}, api.Inference)
				Expect(err).NotTo(HaveOccurred())
				Expect(jobType).To(Equal(api.InferenceJob))
			})

			It("should return RunaiJob for non-distributed workload type", func() {
				jobType, err := getJobType(&iwv2.InteractiveWorkload{}, api.Workspace)
				Expect(err).NotTo(HaveOccurred())
				Expect(jobType).To(Equal(api.RunaiJob))
			})

			It("should return RunaiJob for non-distributed workload", func() {
				jobType, err := getJobType(&twv2.TrainingWorkload{}, api.Training)
				Expect(err).NotTo(HaveOccurred())
				Expect(jobType).To(Equal(api.RunaiJob))
			})
		})

		It("should return error for non-distributed workload", func() {
			jobType, err := getJobType(&iwv2.InteractiveWorkload{}, api.Distributed)
			Expect(err).To(HaveOccurred())
			Expect(jobType).To(Equal(api.Unknown))
		})

		It("should return MPIJob for distributed workload with MPIJob type", func() {
			mockWorkload.Spec.JobType = constants.KindMpiJob
			jobType, err := getJobType(mockWorkload, api.Distributed)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobType).To(Equal(api.MPIJob))
		})

		It("should return PyTorchJob for distributed workload with PyTorchJob type", func() {
			mockWorkload.Spec.JobType = constants.KindPyTorchJob
			jobType, err := getJobType(mockWorkload, api.Distributed)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobType).To(Equal(api.PyTorchJob))
		})

		It("should return XGBoost for distributed workload with XGBoostJob type", func() {
			mockWorkload.Spec.JobType = constants.KindXGBoostJob
			jobType, err := getJobType(mockWorkload, api.Distributed)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobType).To(Equal(api.XGBoostJob))
		})

		It("should return TFJob for distributed workload with TFJob type", func() {
			mockWorkload.Spec.JobType = constants.KindTFJob
			jobType, err := getJobType(mockWorkload, api.Distributed)
			Expect(err).NotTo(HaveOccurred())
			Expect(jobType).To(Equal(api.TFJob))
		})

		It("should return Unknown and error for distributed workload with unknown job type", func() {
			mockWorkload.Spec.JobType = "unknown"
			jobType, err := getJobType(mockWorkload, api.Distributed)
			Expect(err).To(HaveOccurred())
			Expect(jobType).To(Equal(api.Unknown))
			Expect(err.Error()).To(Equal("invalid workload, unknown job type unknown"))
		})
	})

	Context("getWorkloadPhase", func() {
		const (
			expectedPhase = "wonderful"
		)

		It("should return error for nil workload", func() {
			phase, err := getWorkloadPhase(nil, api.Distributed)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(workloadIsNilErrorMsg))
			Expect(len(phase)).To(Equal(0))
		})

		Context("distributed", func() {
			It("success", func() {
				mockWorkload := &dwv2.DistributedWorkload{}
				mockWorkload.Status.Phase = expectedPhase

				phase, err := getWorkloadPhase(mockWorkload, api.Distributed)
				Expect(phase).To(Equal(expectedPhase))
				Expect(err).NotTo(HaveOccurred())
			})

			It("type mismatch", func() {
				mockWorkload := &iwv2.InteractiveWorkload{}

				phase, err := getWorkloadPhase(mockWorkload, api.Distributed)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf(invalidWorkloadErrorMsg, api.Distributed)))
				Expect(len(phase)).To(Equal(0))
			})
		})

		Context("workspace", func() {
			It("success", func() {
				mockWorkload := &iwv2.InteractiveWorkload{}
				mockWorkload.Status.Phase = expectedPhase

				phase, err := getWorkloadPhase(mockWorkload, api.Workspace)
				Expect(phase).To(Equal(expectedPhase))
				Expect(err).NotTo(HaveOccurred())
			})

			It("type mismatch", func() {
				mockWorkload := &twv2.TrainingWorkload{}

				phase, err := getWorkloadPhase(mockWorkload, api.Workspace)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf(invalidWorkloadErrorMsg, api.Workspace)))
				Expect(len(phase)).To(Equal(0))
			})
		})

		Context("inference", func() {
			It("success", func() {
				mockWorkload := &infwv2.InferenceWorkload{}
				mockWorkload.Status.Phase = expectedPhase

				phase, err := getWorkloadPhase(mockWorkload, api.Inference)
				Expect(phase).To(Equal(expectedPhase))
				Expect(err).NotTo(HaveOccurred())
			})

			It("type mismatch", func() {
				mockWorkload := &twv2.TrainingWorkload{}

				phase, err := getWorkloadPhase(mockWorkload, api.Inference)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf(invalidWorkloadErrorMsg, api.Inference)))
				Expect(len(phase)).To(Equal(0))
			})
		})

		Context("distributedinference", func() {
			It("success", func() {
				mockWorkload := &distinfwv2.DistributedInferenceWorkload{}
				mockWorkload.Status.Phase = expectedPhase

				phase, err := getWorkloadPhase(mockWorkload, api.DistributedInference)
				Expect(phase).To(Equal(expectedPhase))
				Expect(err).NotTo(HaveOccurred())
			})

			It("type mismatch", func() {
				mockWorkload := &twv2.TrainingWorkload{}

				phase, err := getWorkloadPhase(mockWorkload, api.DistributedInference)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf(invalidWorkloadErrorMsg, api.DistributedInference)))
				Expect(len(phase)).To(Equal(0))
			})
		})

		Context("training", func() {
			It("success", func() {
				mockWorkload := &twv2.TrainingWorkload{}
				mockWorkload.Status.Phase = expectedPhase

				phase, err := getWorkloadPhase(mockWorkload, api.Training)
				Expect(phase).To(Equal(expectedPhase))
				Expect(err).NotTo(HaveOccurred())
			})

			It("type mismatch", func() {
				mockWorkload := &iwv2.InteractiveWorkload{}

				phase, err := getWorkloadPhase(mockWorkload, api.Training)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf(invalidWorkloadErrorMsg, api.Training)))
				Expect(len(phase)).To(Equal(0))
			})
		})

		Context("external", func() {
			It("success", func() {
				mockWorkload := &v1alpha1.ExternalWorkload{}
				mockWorkload.Status.ObservedStatus.Phase = v1alpha1.ExWRunningPhase

				phase, err := getWorkloadPhase(mockWorkload, api.External)
				Expect(phase).To(Equal(string(v1alpha1.ExWRunningPhase)))
				Expect(err).NotTo(HaveOccurred())
			})

			It("empty phase returns Undefined", func() {
				mockWorkload := &v1alpha1.ExternalWorkload{}

				phase, err := getWorkloadPhase(mockWorkload, api.External)
				Expect(phase).To(Equal("Undefined"))
				Expect(err).NotTo(HaveOccurred())
			})

			It("type mismatch", func() {
				mockWorkload := &twv2.TrainingWorkload{}

				phase, err := getWorkloadPhase(mockWorkload, api.External)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf(invalidWorkloadErrorMsg, api.External)))
				Expect(len(phase)).To(Equal(0))
			})
		})
	})
})
