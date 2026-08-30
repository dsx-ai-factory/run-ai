package checker

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-application-packages/convutils"
	"github.com/run-ai/runai/runai-application-packages/testutils"
	"github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	cache_testing "github.com/run-ai/runai/runai-common-packages/cache/v2/testing"

	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	workloadsService "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

var defaultCheckerClusterConfig = &ClustersCacheConfig{ExpirationSec: 600, CleanupSec: 10}

var _ = Describe("Cluster capabilities checker", func() {
	clusterId := "cluster-id"
	differentClusterId := "different-cluster-id"
	wrongVersionClusterId := "wrong-version-cluster-id"
	emptyVersionClusterId := "empty-version-cluster-id"
	var clusterMock cluster.SimpleMockClient

	BeforeEach(func() {
		// Before each test
		clusterMock = cluster.NewSimpleMockClient().
			WithCluster(clusterId, buildDisplayedCluster(testutils.NewRandomValues(), "2.13.0", nil)).
			WithCluster(differentClusterId, buildDisplayedCluster(testutils.NewRandomValues(), "2.17.0", nil)).
			WithCluster(wrongVersionClusterId, buildDisplayedCluster(testutils.NewRandomValues(), "wrong-version", nil)).
			WithCluster(emptyVersionClusterId, buildDisplayedCluster(testutils.NewRandomValues(), "", nil))
		cluster.SetMock(clusterMock)
		cache_testing.EnableCache()
		InitClustersCache(defaultCheckerClusterConfig)
	})

	Context("cluster capability checker functionality", func() {

		Context("Should not be capable", func() {
			It("should not be capable due to incompatible version", func() {
				checker := NewClusterCapabilitiesChecker(clusterId)
				isCapable, _ := checker.ValidateClusterCapability(context.Background(), MinVersionGitSyncOneParameter)
				Expect(isCapable).To(BeFalse())
			})

			It("should not be capable due to wrong version", func() {
				checker := NewClusterCapabilitiesChecker(wrongVersionClusterId)
				isCapable, _ := checker.ValidateClusterCapability(context.Background(), MinVersionGitSyncOneParameter)
				Expect(isCapable).To(BeFalse())
			})

			It("should return ErrClusterVersionEmpty when version is empty", func() {
				checker := NewClusterCapabilitiesChecker(emptyVersionClusterId)
				_, err := checker.ValidateClusterCapability(context.Background(), MinVersionGitSyncOneParameter)
				Expect(err).To(HaveOccurred())
				Expect(errors.Is(err, ErrClusterVersionEmpty)).To(BeTrue())
			})
		})

		It("should be capable for compatible version", func() {
			checker := NewClusterCapabilitiesChecker(differentClusterId)
			isCapable, _ := checker.ValidateClusterCapability(context.Background(), MinVersionGitSyncOneParameter)
			Expect(isCapable).To(BeTrue())
		})

		It("enforces the ServiceSpec labels capability version boundary", func() {
			Expect(MinVersionForSupporting(MinVersionWorkloadV2ServiceSpecLabels)).To(Equal("3.0.0"))

			checker, err := NewClusterCapabilitiesCheckerForVersion("2.99.9")
			Expect(err).NotTo(HaveOccurred())
			isCapable, err := checker.ValidateClusterCapability(context.Background(), MinVersionWorkloadV2ServiceSpecLabels)
			Expect(err).NotTo(HaveOccurred())
			Expect(isCapable).To(BeFalse())

			checker, err = NewClusterCapabilitiesCheckerForVersion("3.0.0")
			Expect(err).NotTo(HaveOccurred())
			isCapable, err = checker.ValidateClusterCapability(context.Background(), MinVersionWorkloadV2ServiceSpecLabels)
			Expect(err).NotTo(HaveOccurred())
			Expect(isCapable).To(BeTrue())
		})

		It("enforces the vLLM metrics capability version boundary", func() {
			Expect(MinVersionForSupporting(MinVersionVLLMMetrics)).To(Equal("3.0.0"))

			checker, err := NewClusterCapabilitiesCheckerForVersion("2.99.9")
			Expect(err).NotTo(HaveOccurred())
			isCapable, err := checker.ValidateClusterCapability(context.Background(), MinVersionVLLMMetrics)
			Expect(err).NotTo(HaveOccurred())
			Expect(isCapable).To(BeFalse())

			checker, err = NewClusterCapabilitiesCheckerForVersion("3.0.0")
			Expect(err).NotTo(HaveOccurred())
			isCapable, err = checker.ValidateClusterCapability(context.Background(), MinVersionVLLMMetrics)
			Expect(err).NotTo(HaveOccurred())
			Expect(isCapable).To(BeTrue())
		})

		It("should not call the cluster client if the cluster is in the cache", func() {
			err := cache.ClearStore[clusterInfo]()
			Expect(err).To(BeNil())

			checker := NewClusterCapabilitiesChecker(clusterId)
			_, _ = checker.ValidateClusterCapability(context.Background(), MinVersionGitSyncOneParameter)
			_, _ = checker.ValidateClusterCapability(context.Background(), MinVersionGitSyncOneParameter)
			Expect(*clusterMock.GetClusterByIdCallCount).To(Equal(1))
		})
	})

	Context("cluster capability checker ", func() {
		It("workload ownership protection", func() {
			cluster.SetMock(cluster.NewSimpleMockClient().
				WithCluster("no value", buildDisplayedCluster(testutils.NewRandomValues(), "2.13.0", nil)).
				WithCluster("protection off", buildDisplayedCluster(testutils.NewRandomValues(), "2.17.0", convutils.Ref(false))).
				WithCluster("protection on", buildDisplayedCluster(testutils.NewRandomValues(), "2.19.0", convutils.Ref(true))))
			checker := NewClusterCapabilitiesChecker("no value")
			protection, err := checker.ClusterConfiguration(context.Background())
			Expect(err).To(BeNil())
			Expect(protection.WorkloadOwnershipProtection).To(BeFalse())

			checker = NewClusterCapabilitiesChecker("protection off")
			protection, err = checker.ClusterConfiguration(context.Background())
			Expect(err).To(BeNil())
			Expect(protection.WorkloadOwnershipProtection).To(BeFalse())

			checker = NewClusterCapabilitiesChecker("protection on")
			protection, err = checker.ClusterConfiguration(context.Background())
			Expect(err).To(BeNil())
			Expect(protection.WorkloadOwnershipProtection).To(BeTrue())
		})

		It("distributed frameworks status", func() {
			cluster.SetMock(cluster.NewSimpleMockClient().
				WithCluster("id", buildDisplayedCluster(testutils.NewRandomValues(), "2.20.0", nil)))

			checker := NewClusterCapabilitiesChecker("id")
			configuration, err := checker.ClusterConfiguration(context.Background())
			Expect(err).To(BeNil())

			frameworksStatus := configuration.DistributedFrameworksStatus
			Expect(len(DistributedFrameworksDependencies)).To(Equal(5))
			Expect(len(frameworksStatus)).To(Equal(len(DistributedFrameworksDependencies)))

			mpiStatus, ok := frameworksStatus[workloadsService.DISTRIBUTEDFRAMEWORK_MPI]
			Expect(ok).To(BeTrue())
			Expect(mpiStatus.Available).To(BeTrue())
			Expect(mpiStatus.Suspendable).To(BeTrue())

			pytorchStatus, ok := frameworksStatus[workloadsService.DISTRIBUTEDFRAMEWORK_PY_TORCH]
			Expect(ok).To(BeTrue())
			Expect(pytorchStatus.Available).To(BeTrue())
			Expect(pytorchStatus.Suspendable).To(BeFalse())

			tensorflowStatus, ok := frameworksStatus[workloadsService.DISTRIBUTEDFRAMEWORK_TF]
			Expect(ok).To(BeTrue())
			Expect(tensorflowStatus.Available).To(BeFalse())
			Expect(tensorflowStatus.Suspendable).To(BeFalse())

			// not in config, should be false by default
			xgboostStatus, ok := frameworksStatus[workloadsService.DISTRIBUTEDFRAMEWORK_XG_BOOST]
			Expect(ok).To(BeTrue())
			Expect(xgboostStatus.Available).To(BeFalse())
			Expect(xgboostStatus.Suspendable).To(BeFalse())

			// not in config, should be false by default
			jaxStatus, ok := frameworksStatus[workloadsService.DISTRIBUTEDFRAMEWORK_JAX]
			Expect(ok).To(BeTrue())
			Expect(jaxStatus.Available).To(BeFalse())
			Expect(jaxStatus.Suspendable).To(BeFalse())
		})
	})

	Context("knative availability and scale-to-zero capability", func() {
		It("should handle missing knative sub components", func() {
			cluster.SetMock(cluster.NewSimpleMockClient().
				WithCluster("no-sub-components", buildDisplayedClusterWithKnative(testutils.NewRandomValues(), "2.20.0", true, nil)))

			checker := NewClusterCapabilitiesChecker("no-sub-components")
			configuration, err := checker.ClusterConfiguration(context.Background())
			Expect(err).To(BeNil())
			Expect(configuration.KnativeAvailable).To(BeTrue())
			Expect(configuration.KnativeScaleToZeroAvailable).To(BeTrue())
		})

		It("should handle missing scale-to-zero component", func() {
			cluster.SetMock(cluster.NewSimpleMockClient().
				WithCluster("no-scale-zero", buildDisplayedClusterWithKnative(testutils.NewRandomValues(), "2.20.0", true, &map[string]openapi.ClusterDependencyStatus{})))

			checker := NewClusterCapabilitiesChecker("no-scale-zero")
			configuration, err := checker.ClusterConfiguration(context.Background())
			Expect(err).To(BeNil())
			Expect(configuration.KnativeAvailable).To(BeTrue())
			Expect(configuration.KnativeScaleToZeroAvailable).To(BeTrue())
		})

		It("should handle scale-to-zero enabled", func() {
			cluster.SetMock(cluster.NewSimpleMockClient().
				WithCluster("scale-zero-enabled", buildDisplayedClusterWithKnative(testutils.NewRandomValues(), "2.20.0", true, &map[string]openapi.ClusterDependencyStatus{
					ScaleToZeroDependency: {
						Available: true,
					},
				})))

			checker := NewClusterCapabilitiesChecker("scale-zero-enabled")
			configuration, err := checker.ClusterConfiguration(context.Background())
			Expect(err).To(BeNil())
			Expect(configuration.KnativeAvailable).To(BeTrue())
			Expect(configuration.KnativeScaleToZeroAvailable).To(BeTrue())
		})

		It("should handle scale-to-zero disabled", func() {
			cluster.SetMock(cluster.NewSimpleMockClient().
				WithCluster("scale-zero-disabled", buildDisplayedClusterWithKnative(testutils.NewRandomValues(), "2.20.0", true, &map[string]openapi.ClusterDependencyStatus{
					ScaleToZeroDependency: {
						Available: false,
					},
				})))

			checker := NewClusterCapabilitiesChecker("scale-zero-disabled")
			configuration, err := checker.ClusterConfiguration(context.Background())
			Expect(err).To(BeNil())
			Expect(configuration.KnativeAvailable).To(BeTrue())
			Expect(configuration.KnativeScaleToZeroAvailable).To(BeFalse())
		})
	})
})

func buildDisplayedCluster(randoms *testutils.RandomValues, version string, protection *bool) openapi.DisplayedCluster {
	result := openapi.DisplayedCluster{
		Uuid: randoms.ClusterId,
	}
	result.Version.Set(&version)

	displayedStatus := openapi.NewClusterDisplayedStatus()

	optionalDependencies := map[string]openapi.ClusterDependencyStatus{
		MPIDependency: {
			Available: true,
			Components: &map[string]openapi.ClusterDependencyStatus{
				DistributedSuspendComponent: {
					Available: true,
				},
			},
		},
		PyTorchDependency: {
			Available: true,
			Components: &map[string]openapi.ClusterDependencyStatus{
				DistributedSuspendComponent: {
					Available: false,
				},
			},
		},
		TensorflowDependency: {
			Available: false,
			Components: &map[string]openapi.ClusterDependencyStatus{
				DistributedSuspendComponent: {
					Available: false,
				},
			},
		},
	}
	dependenciesStatus := openapi.NewClusterDependenciesStatusWithDefaults()
	dependenciesStatus.SetOptional(optionalDependencies)
	displayedStatus.SetDependencies(*dependenciesStatus)

	if protection != nil {
		clusterConfig := openapi.NewClusterReportedConfig()
		clusterConfig.SetWorkloadOwnershipProtection(*protection)
		displayedStatus.SetConfig(*clusterConfig)
	}

	result.SetStatus(*displayedStatus)

	return result
}

func buildDisplayedClusterWithKnative(randoms *testutils.RandomValues, version string, knativeAvailable bool, knativeSubComponents *map[string]openapi.ClusterDependencyStatus) openapi.DisplayedCluster {
	result := openapi.DisplayedCluster{
		Uuid: randoms.ClusterId,
	}
	result.Version.Set(&version)

	displayedStatus := openapi.NewClusterDisplayedStatus()

	optionalDependencies := map[string]openapi.ClusterDependencyStatus{
		KnativeDependency: {
			Available: knativeAvailable,
			Components: &map[string]openapi.ClusterDependencyStatus{
				KnativeDependency: {
					Available:  knativeAvailable,
					Components: knativeSubComponents,
				},
			},
		},
	}
	dependenciesStatus := openapi.NewClusterDependenciesStatusWithDefaults()
	dependenciesStatus.SetOptional(optionalDependencies)
	displayedStatus.SetDependencies(*dependenciesStatus)

	result.SetStatus(*displayedStatus)

	return result
}
