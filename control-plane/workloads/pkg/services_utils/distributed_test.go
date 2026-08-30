package services_utils

import (
	"context"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/clients/orgunit"
	"github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	testing "github.com/run-ai/runai/control-plane/workloads/pkg/test"
	"github.com/run-ai/runai/control-plane/workloads/pkg/testing_crud"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	workloadsApi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
)

var _ = Describe("distributed utils testing", func() {
	const (
		MpiAvailable   = true
		MpiSuspendable = true

		PyTorchAvailable   = true
		PyTorchSuspendable = false

		jaxAvailable   = true
		jaxSuspendable = false

		TfAvailable   = false
		TfSuspendable = false

		XgboostAvailable   = false
		XgboostSuspendable = true
	)

	var ctx context.Context
	var randoms *testutils.RandomValues

	BeforeEach(func() {
		randoms = testutils.NewRandomValues()

		ctx = testing.SetContext(randoms.TenantIdStr, testing.TestUser, []string{})

		authMock := testing_crud.NewRandomsBasedAuthMockClient().With(randoms)
		authMock.AllowedProjectsToContext(&ctx)

		orgunit.SetMock(authMock.OrgUnitMock())

		frameworkDeps := testing.GetCustomFrameworkDependencies(MpiAvailable, MpiSuspendable, PyTorchAvailable, PyTorchSuspendable, jaxAvailable, jaxSuspendable, TfAvailable, TfSuspendable, XgboostAvailable, XgboostSuspendable)
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(randoms.ClusterId, testing.BuildDisplayedClusterWithCustomFrameworks(randoms.ClusterId, testing.TestDefaultClusterVersion, nil, frameworkDeps)))

		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")
	})

	Context("CheckDistributedFrameworkAvailable", func() {
		It("Available = true", func() {
			Expect(CheckDistributedFrameworkAvailable(ctx, randoms.ClusterId, string(api.DISTRIBUTEDFRAMEWORK_MPI))).To(BeNil())
		})
		It("Available = true (parsing)", func() {
			Expect(CheckDistributedFrameworkAvailable(ctx, randoms.ClusterId, string(api.DISTRIBUTEDFRAMEWORK_MPI)+"Job")).To(BeNil())
			Expect(CheckDistributedFrameworkAvailable(ctx, randoms.ClusterId, "mpi")).To(BeNil())
		})
		It("Available = false", func() {
			apiErr := CheckDistributedFrameworkAvailable(ctx, randoms.ClusterId, string(api.DISTRIBUTEDFRAMEWORK_TF))
			Expect(apiErr).NotTo(BeNil())
			Expect(int(apiErr.Code)).To(Equal(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal(fmt.Sprintf(FrameworkNotAvailableMsg, api.DISTRIBUTEDFRAMEWORK_TF)))
		})
	})

	Context("CheckDistributedFrameworkSuspendable", func() {
		It("Available = true, Suspendable = true", func() {
			Expect(CheckDistributedFrameworkSuspendable(ctx, randoms.ClusterId, string(api.DISTRIBUTEDFRAMEWORK_MPI))).To(BeNil())
		})
		It("Available = true, Suspendable = true (parsing)", func() {
			Expect(CheckDistributedFrameworkSuspendable(ctx, randoms.ClusterId, string(api.DISTRIBUTEDFRAMEWORK_MPI)+"Job")).To(BeNil())
			Expect(CheckDistributedFrameworkSuspendable(ctx, randoms.ClusterId, "mpi")).To(BeNil())
		})
		It("Available = false, Suspendable = true", func() {
			apiErr := CheckDistributedFrameworkSuspendable(ctx, randoms.ClusterId, string(api.DISTRIBUTEDFRAMEWORK_XG_BOOST))
			Expect(apiErr).NotTo(BeNil())
			Expect(int(apiErr.Code)).To(Equal(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal(fmt.Sprintf(FrameworkNotAvailableMsg, api.DISTRIBUTEDFRAMEWORK_XG_BOOST)))
		})
		It("Available = false, Suspendable = false", func() {
			apiErr := CheckDistributedFrameworkSuspendable(ctx, randoms.ClusterId, string(api.DISTRIBUTEDFRAMEWORK_TF))
			Expect(apiErr).NotTo(BeNil())
			Expect(int(apiErr.Code)).To(Equal(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal(fmt.Sprintf(FrameworkNotAvailableMsg, api.DISTRIBUTEDFRAMEWORK_TF)))
		})
		It("Available = true, Suspendable = false", func() {
			apiErr := CheckDistributedFrameworkSuspendable(ctx, randoms.ClusterId, string(api.DISTRIBUTEDFRAMEWORK_PY_TORCH))
			Expect(apiErr).NotTo(BeNil())
			Expect(int(apiErr.Code)).To(Equal(http.StatusBadRequest))
			Expect(apiErr.Message).To(Equal(fmt.Sprintf(FrameworkNotSuspendableMsg, api.DISTRIBUTEDFRAMEWORK_PY_TORCH)))
		})
	})

	Context("parseDistributedFramework", func() {
		DescribeTable("parses correctly",
			func(options []string, expectedParsed workloadsApi.DistributedFramework) {
				for _, option := range options {
					parsed, apiErr := parseDistributedFramework(ctx, option)
					Expect(apiErr).To(BeNil())
					Expect(parsed).To(Equal(expectedParsed))
				}
			},
			Entry("MPI", []string{"MPI", "MPIJob", "mpijob", "MpiJob"}, workloadsApi.DISTRIBUTEDFRAMEWORK_MPI),
			Entry("TF", []string{"TF", "TFJob", "tensorflow", "TensorFlow", "tensorflowJob"}, workloadsApi.DISTRIBUTEDFRAMEWORK_TF),
			Entry("PyTorch", []string{"PyTorch", "PyTorchJob", "Pytorch", "PytorchJob", "pytorchjob", "pytorchJob"}, workloadsApi.DISTRIBUTEDFRAMEWORK_PY_TORCH),
			Entry("XGBoost", []string{"XGBoost", "XGBoostJob", "xgboost", "xgboostjob"}, workloadsApi.DISTRIBUTEDFRAMEWORK_XG_BOOST),
		)

		DescribeTable("parse failure",
			func(expectedFailure string) {
				parsed, apiErr := parseDistributedFramework(ctx, expectedFailure)
				Expect(parsed).To(Equal(workloadsApi.DistributedFramework("")))
				Expect(int(apiErr.Code)).To(Equal(http.StatusInternalServerError))
				Expect(apiErr.Message).To(Equal(fmt.Sprintf(UnknownFrameworkMsg, expectedFailure)))
			},
			Entry("MPIJobs", "MPIJobs"),
			Entry("BlaBla", "BlaBla"),
		)
	})
})
