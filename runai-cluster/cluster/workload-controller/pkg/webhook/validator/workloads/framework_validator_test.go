package workloads_test

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/controller"
	fake "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	workloadUtils "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/webhook/validator/workloads"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	Installed               = "Installed"
	NotInstalled            = "NotInstalled"
	InstalledSuspendable    = "Installed_Suspendable"
	InstalledNotSuspendable = "Installed_NotSuspendable"
)

var _ = Describe("WorkloadFrameworkValidator", func() {
	var (
		validator  *workloads.WorkloadFrameworkValidator
		fakeClient client.Client
		scheme     *runtime.Scheme
	)
	scheme = runtime.NewScheme()
	register.InitSchema(scheme)

	BeforeEach(func() {
		fakeClient = fake.InitFakeClientWithObjectNoStatus(scheme)
		enabledFrameworks := map[string]controller.DistributedJobFrameworksStatus{
			Installed:               {JobObject: nil},
			InstalledSuspendable:    {JobObject: nil, SuspendEnabled: true},
			InstalledNotSuspendable: {JobObject: nil, SuspendEnabled: false},
		}
		validator = workloads.NewWorkloadFrameworkValidator(fakeClient, enabledFrameworks)
	})

	Context("when the EnabledDistributedJobs map is nil", func() {
		It("initializes an empty map", func() {
			validator = workloads.NewWorkloadFrameworkValidator(fakeClient, nil)
			Expect(validator.EnabledDistributedJobs).ToNot(BeNil())
			Expect(validator.EnabledDistributedJobs).To(BeEmpty())
		})
	})

	Context("when the EnabledDistributedJobs map is not nil", func() {

		It("uses the provided map", func() {
			Expect(validator.EnabledDistributedJobs).To(HaveKey(Installed))
		})
	})

	Context("when validating a distributed workload", func() {
		Context("when enabled distributed jobs map is nil", func() {
			It("should not panic", func() {
				validator = workloads.NewWorkloadFrameworkValidator(fakeClient, nil)
				workload := workloadUtils.NewJobWorkloadBuilder().WithDistribution().WithJobType(NotInstalled).SetActive().Build()
				Expect(func() { validator.Validate(context.Background(), workload) }).ToNot(Panic())
			})
		})
		It("returns an error if the job is not available", func() {
			workload := workloadUtils.NewJobWorkloadBuilder().WithDistribution().WithJobType(NotInstalled).SetActive().Build()
			err := validator.Validate(context.Background(), workload)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("distributed job NotInstalled is not available"))
		})

		It("does not return an error if the job is available", func() {
			workload := workloadUtils.NewJobWorkloadBuilder().WithDistribution().WithJobType(Installed).SetActive().Build()
			err := validator.Validate(context.Background(), workload)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("job is suspended", func() {
			It("does not return an error if the job is available and suspendable", func() {
				workload := workloadUtils.NewJobWorkloadBuilder().WithDistribution().WithJobType(InstalledSuspendable).Build()
				err := validator.Validate(context.Background(), workload)
				Expect(err).NotTo(HaveOccurred())
			})

			It("return an error if the job is available and not suspendable", func() {
				workload := workloadUtils.NewJobWorkloadBuilder().WithDistribution().WithJobType(InstalledNotSuspendable).Build()
				err := validator.Validate(context.Background(), workload)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(fmt.Sprintf(workloads.NotSuspendable, InstalledNotSuspendable)))
			})
		})
	})
})
