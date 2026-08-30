package mutators

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	distributedpolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedpolicy/v2alpha1"
	dwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	iwv1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Mnnvl mutations", func() {
	const (
		labelTestValue       = "blabla"
		clusterMNNVLEnabled  = true
		clusterMNNVLDisabled = false
		nodePoolValue        = "test-nodepool"
	)

	var (
		dw *dwv2.DistributedWorkload
	)

	BeforeEach(func() {
		dw = &dwv2.DistributedWorkload{}
	})

	Context("non-distributed", func() {
		type testCase struct {
			workloadType     string
			featureEnabled   bool
			elasticCDEnabled bool
			getWorkload      func() client.Object
		}

		DescribeTable("should not add label to non-distributed workloads",
			func(tc testCase) {
				err := setMnnvlFeatureLabel(tc.getWorkload(), tc.featureEnabled, tc.elasticCDEnabled)

				Expect(err).NotTo(HaveOccurred())
				Expect(tc.getWorkload().GetLabels()).To(HaveLen(0))
			},
			Entry("Training with MNNVL enabled", testCase{
				workloadType:     "Training",
				featureEnabled:   clusterMNNVLEnabled,
				elasticCDEnabled: true,
				getWorkload:      func() client.Object { return &twv2.TrainingWorkload{} },
			}),
			Entry("Interactive with MNNVL enabled", testCase{
				workloadType:     "Interactive",
				featureEnabled:   clusterMNNVLEnabled,
				elasticCDEnabled: true,
				getWorkload:      func() client.Object { return &iwv1.InteractiveWorkload{} },
			}),
			Entry("Training with MNNVL disabled", testCase{
				workloadType:     "Training",
				featureEnabled:   clusterMNNVLDisabled,
				elasticCDEnabled: true,
				getWorkload:      func() client.Object { return &twv2.TrainingWorkload{} },
			}),
			Entry("Interactive with MNNVL disabled", testCase{
				workloadType:     "Interactive",
				featureEnabled:   clusterMNNVLDisabled,
				elasticCDEnabled: true,
				getWorkload:      func() client.Object { return &iwv1.InteractiveWorkload{} },
			}),
		)
	})

	It("skip if MNNVL meta label exist", func() {
		dw.ObjectMeta.Labels = map[string]string{
			mnnvl.FeatureLabelKey: labelTestValue,
		}

		err := setMnnvlFeatureLabel(dw, clusterMNNVLEnabled, true)

		Expect(err).NotTo(HaveOccurred())
		Expect(dw.ObjectMeta.Labels).To(HaveLen(1))
		Expect(dw.ObjectMeta.Labels[mnnvl.FeatureLabelKey]).To(Equal(labelTestValue))
	})

	Context("from spec", func() {
		BeforeEach(func() {
			dw.Spec = distributedpolicy.DistributedSpec{
				WorkloadFields: field_groups.WorkloadFields{
					ContainerFields: field_groups.ContainerFields{
						Labels: field_types.NewDictionaryField(make(map[string]string)),
					},
				},
			}
			dw.MasterSpec = &distributedpolicy.DistributedMasterSpec{
				WorkloadFields: field_groups.WorkloadFields{
					ContainerFields: field_groups.ContainerFields{
						Labels: field_types.NewDictionaryField(make(map[string]string)),
					},
				},
			}

			Expect(dw.ObjectMeta.Labels).To(HaveLen(0))
		})
		It("MNNVL label in spec", func() {
			dw.Spec.Labels.Put(mnnvl.FeatureLabelKey, labelTestValue)

			err := setMnnvlFeatureLabel(dw, clusterMNNVLEnabled, true)

			Expect(err).NotTo(HaveOccurred())
			Expect(dw.ObjectMeta.Labels).To(HaveLen(1))
			Expect(dw.ObjectMeta.Labels[mnnvl.FeatureLabelKey]).To(Equal(labelTestValue))
		})
		It("MNNVL label in master spec", func() {
			dw.MasterSpec.Labels.Put(mnnvl.FeatureLabelKey, labelTestValue)

			err := setMnnvlFeatureLabel(dw, clusterMNNVLEnabled, true)

			Expect(err).NotTo(HaveOccurred())
			Expect(dw.ObjectMeta.Labels).To(HaveLen(1))
			Expect(dw.ObjectMeta.Labels[mnnvl.FeatureLabelKey]).To(Equal(labelTestValue))
		})
		It("MNNVL label in spec & master spec - same value", func() {
			dw.Spec.Labels.Put(mnnvl.FeatureLabelKey, labelTestValue)
			dw.MasterSpec.Labels.Put(mnnvl.FeatureLabelKey, labelTestValue)

			err := setMnnvlFeatureLabel(dw, clusterMNNVLEnabled, true)

			Expect(err).NotTo(HaveOccurred())
			Expect(dw.ObjectMeta.Labels).To(HaveLen(1))
			Expect(dw.ObjectMeta.Labels[mnnvl.FeatureLabelKey]).To(Equal(labelTestValue))
		})
		It("MNNVL label in spec & master spec - different value", func() {
			dw.Spec.Labels.Put(mnnvl.FeatureLabelKey, labelTestValue)
			dw.MasterSpec.Labels.Put(mnnvl.FeatureLabelKey, "otherValue")

			err := setMnnvlFeatureLabel(dw, clusterMNNVLEnabled, true)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("providing different values"))
			Expect(dw.ObjectMeta.Labels).To(HaveLen(0))
		})

		It("MNNVL label in spec but feature disabled", func() {
			dw.Spec.Labels.Put(mnnvl.FeatureLabelKey, mnnvl.Preferred)

			err := setMnnvlFeatureLabel(dw, clusterMNNVLDisabled, true)

			Expect(err).NotTo(HaveOccurred())
			Expect(dw.ObjectMeta.Labels).To(HaveLen(1))
			Expect(dw.ObjectMeta.Labels[mnnvl.FeatureLabelKey]).To(Equal(mnnvl.Preferred))
		})

		It("No MNNVL label in spec and feature disabled", func() {
			// Don't set any MNNVL label in the spec

			err := setMnnvlFeatureLabel(dw, clusterMNNVLDisabled, false)

			Expect(err).NotTo(HaveOccurred())
			Expect(dw.ObjectMeta.Labels).To(HaveLen(1))
			Expect(dw.ObjectMeta.Labels[mnnvl.FeatureLabelKey]).To(Equal(mnnvl.None))
		})

		It("No MNNVL label in spec but feature enabled", func() {
			// Don't set any MNNVL label in the spec

			err := setMnnvlFeatureLabel(dw, clusterMNNVLEnabled, true)

			Expect(err).NotTo(HaveOccurred())
			Expect(dw.ObjectMeta.Labels).To(HaveLen(1))
			Expect(dw.ObjectMeta.Labels[mnnvl.FeatureLabelKey]).To(Equal(mnnvl.None))
		})
	})

	Context("default value", func() {
		type testCase struct {
			name             string
			setupWorkload    func(*dwv2.DistributedWorkload)
			featureEnabled   bool
			elasticCDEnabled bool
			expectedLabel    string
		}

		DescribeTable("should set the correct default value based on conditions",
			func(tc testCase) {
				tc.setupWorkload(dw)

				err := setMnnvlFeatureLabel(dw, tc.featureEnabled, tc.elasticCDEnabled)

				Expect(err).NotTo(HaveOccurred())
				Expect(dw.ObjectMeta.Labels).To(HaveLen(1))
				Expect(dw.ObjectMeta.Labels[mnnvl.FeatureLabelKey]).To(Equal(tc.expectedLabel))
			},
			Entry("cpu-only with feature enabled", testCase{
				name:           "cpu-only with feature enabled",
				setupWorkload:  func(dw *dwv2.DistributedWorkload) {},
				featureEnabled: clusterMNNVLEnabled,
				expectedLabel:  mnnvl.None,
			}),
			Entry("gpu requested with feature enabled and valid node pools", testCase{
				name: "gpu requested with feature enabled and valid node pools",
				setupWorkload: func(dw *dwv2.DistributedWorkload) {
					dw.Spec.ResourceAllocationFields = field_groups.ResourceAllocationFields{
						GpuDevices: field_types.NewOptionalIntField(ptr.To(1)),
					}
					// Set node pools for both spec and master spec
					dw.Spec.NodePools = field_types.NewOptionalStringField(nodePoolValue, "")
					if dw.MasterSpec != nil {
						dw.MasterSpec.NodePools = field_types.NewOptionalStringField(nodePoolValue, "")
					}
				},
				featureEnabled: clusterMNNVLEnabled,
				expectedLabel:  mnnvl.Preferred,
			}),
			Entry("gpu requested with feature disabled", testCase{
				name: "gpu requested with feature disabled",
				setupWorkload: func(dw *dwv2.DistributedWorkload) {
					dw.Spec.ResourceAllocationFields = field_groups.ResourceAllocationFields{
						GpuDevices: field_types.NewOptionalIntField(ptr.To(1)),
					}
					// Node pools don't matter when feature is disabled
					dw.Spec.NodePools = field_types.NewOptionalStringField(nodePoolValue, "")
				},
				featureEnabled: clusterMNNVLDisabled,
				expectedLabel:  mnnvl.None,
			}),
			Entry("gpu requested but empty node pools in spec", testCase{
				name: "gpu requested but empty node pools in spec",
				setupWorkload: func(dw *dwv2.DistributedWorkload) {
					dw.Spec.ResourceAllocationFields = field_groups.ResourceAllocationFields{
						GpuDevices: field_types.NewOptionalIntField(ptr.To(1)),
					}
					// Set empty node pools in spec
					dw.Spec.NodePools = field_types.NewOptionalStringField("", "")
					if dw.MasterSpec != nil {
						dw.MasterSpec.NodePools = field_types.NewOptionalStringField(nodePoolValue, "")
					}
				},
				featureEnabled: clusterMNNVLEnabled,
				expectedLabel:  mnnvl.None, // Should be None since node pools are empty in spec
			}),
			Entry("gpu requested but empty node pools in master spec", testCase{
				name: "gpu requested but empty node pools in master spec",
				setupWorkload: func(dw *dwv2.DistributedWorkload) {
					dw.Spec.ResourceAllocationFields = field_groups.ResourceAllocationFields{
						GpuDevices: field_types.NewOptionalIntField(ptr.To(1)),
					}
					// Set empty node pools in master spec
					dw.Spec.NodePools = field_types.NewOptionalStringField(nodePoolValue, "")
					if dw.MasterSpec == nil {
						dw.MasterSpec = &distributedpolicy.DistributedMasterSpec{}
					}
					dw.MasterSpec.NodePools = field_types.NewOptionalStringField("", "")
				},
				featureEnabled: clusterMNNVLEnabled,
				expectedLabel:  mnnvl.Preferred, // According to isValidForMNNVL implementation, OR condition is used for node pools
			}),
			Entry("gpu requested with no node pools set", testCase{
				name: "gpu requested with no node pools set",
				setupWorkload: func(dw *dwv2.DistributedWorkload) {
					dw.Spec.ResourceAllocationFields = field_groups.ResourceAllocationFields{
						GpuDevices: field_types.NewOptionalIntField(ptr.To(1)),
					}
					// Don't set any node pools
				},
				featureEnabled: clusterMNNVLEnabled,
				expectedLabel:  mnnvl.None, // Should be None since node pools are not set
			}),
		)

		It("elastic job not supported", func() {
			elasticJob := resource.CreateElasticPyTorchJob(ptr.To(1), ptr.To(3))
			elasticDw := dwv2.NewFromJobWorkload(elasticJob)

			// Add node pools to ensure they're not the limiting factor
			elasticDw.Spec.NodePools = field_types.NewOptionalStringField(nodePoolValue, "")
			if elasticDw.MasterSpec != nil {
				elasticDw.MasterSpec.NodePools = field_types.NewOptionalStringField(nodePoolValue, "")
			}

			err := setMnnvlFeatureLabel(elasticDw, clusterMNNVLEnabled, false)

			Expect(err).NotTo(HaveOccurred())
			Expect(elasticDw.ObjectMeta.Labels).To(HaveLen(1))
			Expect(elasticDw.ObjectMeta.Labels[mnnvl.FeatureLabelKey]).To(Equal(mnnvl.None))
		})

		It("elastic job supported", func() {
			elasticJob := resource.CreateElasticPyTorchJob(ptr.To(1), ptr.To(3))
			elasticDw := dwv2.NewFromJobWorkload(elasticJob)

			// Add node pools to ensure they're not the limiting factor
			elasticDw.Spec.NodePools = field_types.NewOptionalStringField(nodePoolValue, "")
			if elasticDw.MasterSpec != nil {
				elasticDw.MasterSpec.NodePools = field_types.NewOptionalStringField(nodePoolValue, "")
			}

			err := setMnnvlFeatureLabel(elasticDw, clusterMNNVLEnabled, true)

			Expect(err).NotTo(HaveOccurred())
			Expect(elasticDw.ObjectMeta.Labels).To(HaveLen(1))
			Expect(elasticDw.ObjectMeta.Labels[mnnvl.FeatureLabelKey]).To(Equal(mnnvl.Preferred))
		})
	})

	It("original meta labels are kept", func() {
		dw.ObjectMeta.Labels = map[string]string{
			"dummyKey": "dummyValue",
		}

		err := setMnnvlFeatureLabel(dw, clusterMNNVLEnabled, true)

		Expect(err).NotTo(HaveOccurred())
		Expect(dw.ObjectMeta.Labels).To(HaveLen(2))
	})
})
