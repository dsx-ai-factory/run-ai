package validation

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kfv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/runtime-rules/test"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/controller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Validation Test Suite", func() {
	var (
		validator *WorkloadValidator
		iw, dw    *jobs.JobWorkload
	)

	BeforeEach(func() {
		validator = NewWorkloadValidator(NewDistributedValidator(getTestFrameworksStatus()))
		jobStartTime := &metav1.Time{Time: time.Now()}
		iw = &jobs.JobWorkload{}
		test.BuildInteractiveWorkload("namespace", "iw", jobStartTime).CopyTo(iw)
		dw = &jobs.JobWorkload{}
		test.BuildDistributedWorkload("namespace", "dw", jobStartTime).CopyTo(dw)
	})

	Context("Distributed", func() {
		Context("validation success", func() {
			It("not distributed", func() {
				Expect(validator.Validate(iw)).To(BeNil())
			})

			It("framework enabled and suspendable", func() {
				dw.Spec.JobType = constants.KindTFJob
				Expect(validator.Validate(dw)).To(BeNil())
			})
		})
		Context("validation failure", func() {
			It("framework enabled and not suspendable", func() {
				dw.Spec.JobType = constants.KindPyTorchJob
				err := validator.Validate(dw)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf("Cluster prerequisites were not met for %s framework. Contact your administrator for help", dw.Spec.JobType)))
			})

			It("framework not enabled", func() {
				dw.Spec.JobType = constants.KindMpiJob
				err := validator.Validate(dw)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf("Cluster prerequisites were not met for %s framework. Contact your administrator for help", dw.Spec.JobType)))
			})
		})

	})
})

func getTestFrameworksStatus() map[string]controller.DistributedJobFrameworksStatus {
	enabledFrameworks := make(map[string]controller.DistributedJobFrameworksStatus)
	enabledFrameworks[constants.KindTFJob] = controller.DistributedJobFrameworksStatus{JobObject: &kfv1.TFJob{}, SuspendEnabled: true}
	enabledFrameworks[constants.KindPyTorchJob] = controller.DistributedJobFrameworksStatus{JobObject: &kfv1.PyTorchJob{}, SuspendEnabled: false}

	return enabledFrameworks
}
