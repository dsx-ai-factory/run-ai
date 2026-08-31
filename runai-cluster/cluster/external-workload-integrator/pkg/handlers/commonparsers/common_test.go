/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package commonparsers

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiorg "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/external-workload-integrator/test"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	v2 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	cpuLimit     int64 = 800
	cpuRequest   int64 = 400
	memLimit     int64 = 8000
	memRequest   int64 = 4000
	gpuLimit           = 5.5
	gpuRequest         = 2.5
	gpuMemLimit  int64 = 20
	someMigLimit int64 = 2

	gpuFractionAnnotationValue    = 0.4
	gpuFractionAnnotationValueStr = "0.4"
	gpuMemoryAnnotationValue      = 5000
)

func TestCommonPArsers(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Info().Msgf("Started common parsers tests")
	RegisterFailHandler(Fail)
	test.MockEnv(t)
	RunSpecs(t, "common parsers Test Suite")
}

var _ = Describe("TestCommonPArsers", Ordered, func() {
	var (
		externalWorkload   *v1alpha1.ExternalWorkload
		requestedResources = expectedRequestedResources(1)
		podTemplate        *corev1.PodTemplateSpec

		emptyGroupKind = schema.GroupKind{}

		ctx    context.Context
		cancel context.CancelFunc
	)

	BeforeAll(func() {
		ctx, cancel = context.WithCancel(context.Background())
		DeferCleanup(func() {
			cancel()
		})
	})

	BeforeEach(func() {
		externalWorkload = &v1alpha1.ExternalWorkload{}
		podTemplate = &corev1.PodTemplateSpec{}
	})

	Describe("test MultiplyRequestedResourcesPerNumberOfRequestedPods", func() {
		var (
			testExternalWorkload *v1alpha1.ExternalWorkload
		)

		BeforeEach(func() {
			externalWorkload.Spec.RequestedResources = requestedResources
			testExternalWorkload = externalWorkload.DeepCopy()
		})
		It("0 replicas", func() {
			MultiplyRequestedResourcesPerNumberOfRequestedPods(0, testExternalWorkload)
			Expect(testExternalWorkload.Spec.RequestedResources.Limits).To(HaveLen(0))
			Expect(testExternalWorkload.Spec.RequestedResources.Requests).To(HaveLen(0))
		})
		It("1 replicas", func() {
			MultiplyRequestedResourcesPerNumberOfRequestedPods(1, testExternalWorkload)
			test.ValidateRequestedResources(Default, 1, testExternalWorkload.Spec.RequestedResources, requestedResources)
		})
		It(">1 replicas", func() {
			for i := 2; i < 12; i++ {
				testExternalWorkload = externalWorkload.DeepCopy()
				Eventually(func(g Gomega) {
					MultiplyRequestedResourcesPerNumberOfRequestedPods(int32(i), testExternalWorkload)
					test.ValidateRequestedResources(g, 1, testExternalWorkload.Spec.RequestedResources,
						expectedRequestedResources(int64(i)))
				}, time.Millisecond*1).Should(Succeed(), fmt.Sprintf("Number of replicas: <%d>", i))
			}
		})
	})

	Describe("test fillContainersInfo", func() {
		var (
			testImageName  = "some-image"
			testImageName2 = "some-image2"
		)

		BeforeEach(func() {
			podTemplate.Spec.Containers = []corev1.Container{}
			podTemplate.Annotations = map[string]string{}
			externalWorkload.Spec.ImageName = ""
		})
		Describe("test fillContainersInfo - image", func() {
			It("0 containers - empty image", func() {
				externalWorkload.Spec.ImageName = testImageName2
				fillContainersInfo(ctx, podTemplate, externalWorkload)
				Expect(externalWorkload.Spec.ImageName).To(Equal(""))
			})
			It("1 container - 1 image", func() {
				podTemplate.Spec.Containers = []corev1.Container{{Image: testImageName}}
				fillContainersInfo(ctx, podTemplate, externalWorkload)
				Expect(externalWorkload.Spec.ImageName).To(Equal(testImageName))
			})
			It("2 containers - same image", func() {
				podTemplate.Spec.Containers = []corev1.Container{{Image: testImageName}, {Image: testImageName}}
				fillContainersInfo(ctx, podTemplate, externalWorkload)
				Expect(externalWorkload.Spec.ImageName).To(Equal(testImageName))
			})
			It("3 containers - same image", func() {
				podTemplate.Spec.Containers = []corev1.Container{{Image: testImageName2}, {Image: testImageName2},
					{Image: testImageName2}}
				fillContainersInfo(ctx, podTemplate, externalWorkload)
				Expect(externalWorkload.Spec.ImageName).To(Equal(testImageName2))
			})
			It("2 containers - different images - output should be empty", func() {
				externalWorkload.Spec.ImageName = testImageName2
				podTemplate.Spec.Containers = []corev1.Container{{Image: testImageName}, {Image: testImageName2}}
				fillContainersInfo(ctx, podTemplate, externalWorkload)
				Expect(externalWorkload.Spec.ImageName).To(Equal(""))
			})
		})
		Describe("test fillContainersInfo - requested resources", func() {
			var (
				expectedResourcesAnnotationsOnly = corev1.ResourceRequirements{
					Limits: map[corev1.ResourceName]resource.Quantity{
						constants.GpuResourceName:       *resource.NewMilliQuantity(int64(1000*gpuFractionAnnotationValue), resource.DecimalSI),
						constants.GpuMemoryResourceName: *resource.NewQuantity(int64(gpuMemoryAnnotationValue)*test.GpuMemoryLimitConversion, resource.DecimalSI),
					},
					Requests: map[corev1.ResourceName]resource.Quantity{
						constants.GpuResourceName:       *resource.NewMilliQuantity(int64(1000*gpuFractionAnnotationValue), resource.DecimalSI),
						constants.GpuMemoryResourceName: *resource.NewQuantity(int64(gpuMemoryAnnotationValue)*test.GpuMemoryLimitConversion, resource.DecimalSI),
					},
				}
			)

			It("0 containers - empty requested resources", func() {
				externalWorkload.Spec.RequestedResources = expectedRequestedResources(1)
				fillContainersInfo(ctx, podTemplate, externalWorkload)
				Expect(externalWorkload.Spec.RequestedResources.Limits).To(HaveLen(0))
				Expect(externalWorkload.Spec.RequestedResources.Requests).To(HaveLen(0))
			})
			It("0 containers - but has annotation resources", func() {
				externalWorkload.Spec.RequestedResources = expectedRequestedResources(1)
				podTemplate.Annotations[test.GpuFractionAnnotation] = gpuFractionAnnotationValueStr
				podTemplate.Annotations[test.GpuMemoryAnnotation] = strconv.Itoa(gpuMemoryAnnotationValue)
				fillContainersInfo(ctx, podTemplate, externalWorkload)
				Expect(externalWorkload.Spec.RequestedResources.Limits).To(HaveLen(2))
				Expect(externalWorkload.Spec.RequestedResources.Requests).To(HaveLen(2))
				test.ValidateRequestedResources(Default, 1, externalWorkload.Spec.RequestedResources,
					expectedResourcesAnnotationsOnly)
			})
			It("2 containers without resources - but has annotation resources", func() {
				podTemplate.Spec.Containers = []corev1.Container{{Name: "container1"}, {Name: "container2"}}

				externalWorkload.Spec.RequestedResources = expectedRequestedResources(1)
				podTemplate.Annotations[test.GpuFractionAnnotation] = gpuFractionAnnotationValueStr
				podTemplate.Annotations[test.GpuMemoryAnnotation] = strconv.Itoa(gpuMemoryAnnotationValue)
				fillContainersInfo(ctx, podTemplate, externalWorkload)
				Expect(externalWorkload.Spec.RequestedResources.Limits).To(HaveLen(2))
				Expect(externalWorkload.Spec.RequestedResources.Requests).To(HaveLen(2))
				test.ValidateRequestedResources(Default, 1, externalWorkload.Spec.RequestedResources,
					expectedResourcesAnnotationsOnly)
			})
			It("1 container resources with extra resources", func() {
				podTemplate.Spec.Containers = []corev1.Container{{Resources: corev1.ResourceRequirements{
					Limits: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceEphemeralStorage: *resource.NewQuantity(2, resource.DecimalSI),
						corev1.ResourceCPU:              *resource.NewQuantity(2, resource.DecimalSI),
						corev1.ResourceMemory:           *resource.NewQuantity(30, resource.DecimalSI),
						constants.AmdGpuResourceName:    *resource.NewQuantity(2, resource.DecimalSI),
					},
					Requests: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceStorage:       *resource.NewQuantity(3, resource.DecimalSI),
						corev1.ResourceCPU:           *resource.NewQuantity(1, resource.DecimalSI),
						corev1.ResourceMemory:        *resource.NewQuantity(10, resource.DecimalSI),
						constants.AmdGpuResourceName: *resource.NewQuantity(1, resource.DecimalSI),
					},
				}}}

				externalWorkload.Spec.RequestedResources = expectedRequestedResources(1)
				fillContainersInfo(ctx, podTemplate, externalWorkload)
				Expect(externalWorkload.Spec.RequestedResources.Limits).To(HaveLen(4))
				Expect(externalWorkload.Spec.RequestedResources.Requests).To(HaveLen(4))

				expected := corev1.ResourceRequirements{
					Limits: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceEphemeralStorage: *resource.NewQuantity(2, resource.DecimalSI),
						corev1.ResourceCPU:              *resource.NewQuantity(2, resource.DecimalSI),
						corev1.ResourceMemory:           *resource.NewQuantity(30, resource.DecimalSI),
						constants.AmdGpuResourceName:    *resource.NewQuantity(2, resource.DecimalSI),
					},
					Requests: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceStorage:       *resource.NewQuantity(3, resource.DecimalSI),
						corev1.ResourceCPU:           *resource.NewQuantity(1, resource.DecimalSI),
						corev1.ResourceMemory:        *resource.NewQuantity(10, resource.DecimalSI),
						constants.AmdGpuResourceName: *resource.NewQuantity(1, resource.DecimalSI),
					},
				}

				test.ValidateRequestedResources(Default, 1, externalWorkload.Spec.RequestedResources,
					expected)
			})
			It("2 containers with different resources + annotation resources", func() {
				podTemplate.Spec.Containers = []corev1.Container{
					{
						Resources: corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceEphemeralStorage: *resource.NewQuantity(2, resource.DecimalSI),
								corev1.ResourceCPU:              *resource.NewQuantity(2, resource.DecimalSI),
								corev1.ResourceMemory:           *resource.NewQuantity(30, resource.DecimalSI),
								constants.AmdGpuResourceName:    *resource.NewQuantity(2, resource.DecimalSI),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceStorage:       *resource.NewQuantity(3, resource.DecimalSI),
								corev1.ResourceCPU:           *resource.NewQuantity(1, resource.DecimalSI),
								corev1.ResourceMemory:        *resource.NewQuantity(10, resource.DecimalSI),
								constants.AmdGpuResourceName: *resource.NewQuantity(1, resource.DecimalSI),
							},
						},
					},
					{
						Resources: corev1.ResourceRequirements{
							Limits: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceEphemeralStorage: *resource.NewQuantity(10, resource.DecimalSI),
								corev1.ResourceCPU:              *resource.NewQuantity(4, resource.DecimalSI),
								corev1.ResourceMemory:           *resource.NewQuantity(50, resource.DecimalSI),
								constants.AmdGpuResourceName:    *resource.NewQuantity(4, resource.DecimalSI),
							},
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceStorage:       *resource.NewQuantity(5, resource.DecimalSI),
								corev1.ResourceCPU:           *resource.NewQuantity(2, resource.DecimalSI),
								corev1.ResourceMemory:        *resource.NewQuantity(30, resource.DecimalSI),
								constants.AmdGpuResourceName: *resource.NewQuantity(3, resource.DecimalSI),
							},
						},
					},
				}
				podTemplate.Annotations["nvidia.com/mig-1g.5gb"] = "1"

				externalWorkload.Spec.RequestedResources = expectedRequestedResources(1)
				fillContainersInfo(ctx, podTemplate, externalWorkload)
				Expect(externalWorkload.Spec.RequestedResources.Limits).To(HaveLen(5))
				Expect(externalWorkload.Spec.RequestedResources.Requests).To(HaveLen(5))

				expected := corev1.ResourceRequirements{
					Limits: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceEphemeralStorage: *resource.NewQuantity(12, resource.DecimalSI),
						corev1.ResourceCPU:              *resource.NewQuantity(6, resource.DecimalSI),
						corev1.ResourceMemory:           *resource.NewQuantity(80, resource.DecimalSI),
						constants.AmdGpuResourceName:    *resource.NewQuantity(6, resource.DecimalSI),
						"nvidia.com/mig-1g.5gb":         *resource.NewQuantity(1, resource.DecimalSI),
					},
					Requests: map[corev1.ResourceName]resource.Quantity{
						corev1.ResourceStorage:       *resource.NewQuantity(8, resource.DecimalSI),
						corev1.ResourceCPU:           *resource.NewQuantity(3, resource.DecimalSI),
						corev1.ResourceMemory:        *resource.NewQuantity(40, resource.DecimalSI),
						constants.AmdGpuResourceName: *resource.NewQuantity(4, resource.DecimalSI),
						"nvidia.com/mig-1g.5gb":      *resource.NewQuantity(1, resource.DecimalSI),
					},
				}

				test.ValidateRequestedResources(Default, 1, externalWorkload.Spec.RequestedResources,
					expected)

				// same, but using FillExternalWorkloadDataFromPodTemplateSpec instead of fillContainersInfo
				externalWorkload.Spec.RequestedResources = expectedRequestedResources(1)
				externalWorkload.Spec.RequestedNodePools = []string{v1alpha1.DefaultNodePoolName}
				FillExternalWorkloadDataFromPodTemplateSpec(ctx, metav1.ObjectMeta{}, emptyGroupKind,
					podTemplate, nil, nil, nil, externalWorkload)
				Expect(externalWorkload.Spec.RequestedResources.Limits).To(HaveLen(5))
				Expect(externalWorkload.Spec.RequestedResources.Requests).To(HaveLen(5))
				test.ValidateRequestedResources(Default, 1, externalWorkload.Spec.RequestedResources,
					expected)
			})
		})
	})

	Describe("test FillExternalWorkloadDataFromPodTemplateSpec", func() {
		It("test FillExternalWorkloadDataFromPodTemplateSpec - priority", func() {
			podTemplate.Spec.PriorityClassName = "system-node-critical"

			externalWorkload.Spec.PriorityClassName = ""
			externalWorkload.Spec.RequestedNodePools = []string{v1alpha1.DefaultNodePoolName}
			FillExternalWorkloadDataFromPodTemplateSpec(ctx, metav1.ObjectMeta{}, emptyGroupKind,
				podTemplate, nil, nil, nil, externalWorkload)
			Expect(externalWorkload.Spec.PriorityClassName).ToNot(BeEmpty())
			Expect(externalWorkload.Spec.PriorityClassName).To(Equal("system-node-critical"))

			podTemplate.Spec.PriorityClassName = ""
			FillExternalWorkloadDataFromPodTemplateSpec(ctx, metav1.ObjectMeta{}, emptyGroupKind,
				podTemplate, nil, nil, nil, externalWorkload)
			Expect(externalWorkload.Spec.PriorityClassName).To(BeEmpty())
		})

		Describe("test FillExternalWorkloadDataFromPodTemplateSpec - requested node pools", func() {
			// the code is using functions from a common package, that is also used by pod-group-assigner.
			// the full logic is tested there, so it's enough to only have sanity here
			var (
				cachedReader            client.WithWatch
				testObjectMeta          metav1.ObjectMeta
				namespaceName           = "test-namespace"
				projectName             = "test-project"
				projectDefaultNodePools = []string{"nodepool-a", v1alpha1.DefaultNodePoolName}
			)
			BeforeAll(func() {
				project := &v2.Project{
					ObjectMeta: metav1.ObjectMeta{
						Name: projectName,
					},
					Spec: v2.ProjectSpec{
						DefaultNodePools: projectDefaultNodePools,
					},
				}
				// kai.resources twin of the project so the handler resolves the
				// same DefaultNodePools; the entry point branches on the compile-time flag.
				kaiProject := &kaiorg.Project{
					ObjectMeta: metav1.ObjectMeta{
						Name: projectName,
					},
					Spec: kaiorg.ProjectSpec{
						DefaultNodePools: projectDefaultNodePools,
					},
				}
				namespace := &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: namespaceName,
						Labels: map[string]string{
							constants.RunaiQueueLabel: projectName,
						},
					},
				}

				scheme := scheme_init.Scheme()
				Expect(kaiorg.AddToScheme(scheme)).To(Succeed())
				cachedReader = fake.NewClientBuilder().WithScheme(scheme).WithObjects(
					test.TestNodePool, test.TestKaiNodePool, project, kaiProject, namespace).Build()
			})

			BeforeEach(func() {
				testObjectMeta = metav1.ObjectMeta{
					Labels: map[string]string{},
				}
			})

			It("sanity - request a node pool via label", func() {
				// have affinity but with nil RequiredDuringSchedulingIgnoredDuringExecution
				podTemplate.Spec.Affinity = &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: nil,
					},
				}
				externalWorkload.Spec.RequestedNodePools = []string{}
				testObjectMeta.Labels[v1alpha1.NodePoolNameLabel] = test.TestNodePool.Name
				FillExternalWorkloadDataFromPodTemplateSpec(ctx, testObjectMeta, emptyGroupKind,
					podTemplate, nil, cachedReader, nil, externalWorkload)

				Expect(externalWorkload.Spec.RequestedNodePools).To(HaveLen(1))
				Expect(externalWorkload.Spec.RequestedNodePools).To(HaveExactElements(test.TestNodePool.Name))
			})

			It("sanity - not requesting anything - default fallback of project's DefaultNodePools", func() {
				// have affinity but with nil NodeAffinity
				podTemplate.Spec.Affinity = &corev1.Affinity{
					NodeAffinity: nil,
				}
				externalWorkload.Spec.RequestedNodePools = []string{}
				testObjectMeta.Namespace = namespaceName
				FillExternalWorkloadDataFromPodTemplateSpec(ctx, testObjectMeta, emptyGroupKind,
					podTemplate, nil, cachedReader, nil, externalWorkload)
				Expect(externalWorkload.Spec.RequestedNodePools).To(HaveLen(2))
				Expect(externalWorkload.Spec.RequestedNodePools).To(
					Equal(projectDefaultNodePools))

				externalWorkload.Spec.RequestedNodePools = []string{}
				externalWorkload.Spec.ProjectName = projectName
				FillExternalWorkloadDataFromPodTemplateSpec(ctx, testObjectMeta, emptyGroupKind,
					podTemplate, nil, cachedReader, nil, externalWorkload)
				Expect(externalWorkload.Spec.RequestedNodePools).To(HaveLen(2))
				Expect(externalWorkload.Spec.RequestedNodePools).To(
					Equal(projectDefaultNodePools))
			})

			It("sanity - not requesting anything - default fallback", func() {
				// have affinity but with nil NodeAffinity
				podTemplate.Spec.Affinity = &corev1.Affinity{
					NodeAffinity: nil,
				}
				externalWorkload.Spec.RequestedNodePools = []string{}
				externalWorkload.Namespace = "some-other-namespaceee"
				FillExternalWorkloadDataFromPodTemplateSpec(ctx, testObjectMeta, emptyGroupKind,
					podTemplate, nil, cachedReader, nil, externalWorkload)

				Expect(externalWorkload.Spec.RequestedNodePools).To(HaveLen(1))
				Expect(externalWorkload.Spec.RequestedNodePools).To(HaveExactElements(v1alpha1.DefaultNodePoolName))
			})

			It("sanity - request a node pool via affinity", func() {
				podTemplate.Spec.Affinity = &corev1.Affinity{
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{
									MatchExpressions: []corev1.NodeSelectorRequirement{
										{
											Key:      test.TestNodePool.Spec.LabelKey,
											Operator: corev1.NodeSelectorOpIn,
											Values:   []string{test.TestNodePool.Spec.LabelValue},
										},
									},
								},
							},
						},
					},
				}
				externalWorkload.Spec.RequestedNodePools = []string{}
				FillExternalWorkloadDataFromPodTemplateSpec(ctx, testObjectMeta, emptyGroupKind,
					podTemplate, nil, cachedReader, nil, externalWorkload)

				Expect(externalWorkload.Spec.RequestedNodePools).To(HaveLen(1))
				Expect(externalWorkload.Spec.RequestedNodePools).To(HaveExactElements(test.TestNodePool.Name))
			})
		})

		Describe("test FillExternalWorkloadDataFromPodTemplateSpec - node affinity", func() {
			affinityMatch := corev1.NodeSelectorRequirement{
				Key:      "gpu-type",
				Operator: corev1.NodeSelectorOpIn,
				Values:   []string{"a100", "h100"},
			}
			buildPodTemplateWithRequiredAffinity := func(terms []corev1.NodeSelectorTerm) *corev1.PodTemplateSpec {
				return &corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Affinity: &corev1.Affinity{
							NodeAffinity: &corev1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
									NodeSelectorTerms: terms,
								},
							},
						},
					},
				}
			}

			findEntry := func(specName string) *v1alpha1.NodeAffinityEntry {
				for i := range externalWorkload.Spec.NodeAffinities {
					if externalWorkload.Spec.NodeAffinities[i].SpecName == specName {
						return &externalWorkload.Spec.NodeAffinities[i]
					}
				}
				return nil
			}

			// Pre-populating RequestedNodePools makes FillRequestedNodePools exit early
			// so these tests don't need a cachedReader (which FillRequestedNodePools requires).
			presetNodePools := func() {
				externalWorkload.Spec.RequestedNodePools = []string{v1alpha1.DefaultNodePoolName}
			}

			It("copies required nodeAffinity from pod template onto the spec", func() {
				presetNodePools()
				podTemplate = buildPodTemplateWithRequiredAffinity([]corev1.NodeSelectorTerm{
					{MatchExpressions: []corev1.NodeSelectorRequirement{affinityMatch}},
				})
				FillExternalWorkloadDataFromPodTemplateSpec(ctx, metav1.ObjectMeta{}, emptyGroupKind,
					podTemplate, nil, nil, nil, externalWorkload)

				entry := findEntry(v1alpha1.DefaultNodeAffinitySpecName)
				Expect(entry).ToNot(BeNil())
				Expect(entry.NodeAffinity).ToNot(BeNil())
				Expect(entry.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution).ToNot(BeNil())
				Expect(entry.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms).To(HaveLen(1))
				Expect(entry.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions).To(
					ConsistOf(affinityMatch))
			})

			It("does not add an entry when the pod template has no affinity", func() {
				presetNodePools()
				FillExternalWorkloadDataFromPodTemplateSpec(ctx, metav1.ObjectMeta{}, emptyGroupKind,
					podTemplate, nil, nil, nil, externalWorkload)
				Expect(findEntry(v1alpha1.DefaultNodeAffinitySpecName)).To(BeNil())
			})

			It("removes an existing entry when the pod template has no affinity", func() {
				presetNodePools()
				externalWorkload.Spec.NodeAffinities = []v1alpha1.NodeAffinityEntry{{
					SpecName: v1alpha1.DefaultNodeAffinitySpecName,
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{
								{MatchExpressions: []corev1.NodeSelectorRequirement{affinityMatch}},
							},
						},
					},
				}}
				FillExternalWorkloadDataFromPodTemplateSpec(ctx, metav1.ObjectMeta{}, emptyGroupKind,
					podTemplate, nil, nil, nil, externalWorkload)
				Expect(findEntry(v1alpha1.DefaultNodeAffinitySpecName)).To(BeNil())
			})

			It("preserves entries for other specNames when updating one", func() {
				presetNodePools()
				externalWorkload.Spec.NodeAffinities = []v1alpha1.NodeAffinityEntry{{
					SpecName: "master",
					NodeAffinity: &corev1.NodeAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
							NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{
								{Key: "role", Operator: corev1.NodeSelectorOpIn, Values: []string{"control"}},
							}}},
						},
					},
				}}
				podTemplate = buildPodTemplateWithRequiredAffinity([]corev1.NodeSelectorTerm{
					{MatchExpressions: []corev1.NodeSelectorRequirement{affinityMatch}},
				})
				FillExternalWorkloadDataFromPodTemplateSpec(ctx, metav1.ObjectMeta{}, emptyGroupKind,
					podTemplate, ptr.To("worker"), nil, nil, externalWorkload)
				Expect(externalWorkload.Spec.NodeAffinities).To(HaveLen(2))
				Expect(findEntry("master")).ToNot(BeNil())
				Expect(findEntry("worker")).ToNot(BeNil())
			})

			It("IfEmpty variant does not overwrite an existing entry for the same specName", func() {
				presetNodePools()
				existing := &corev1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{{MatchExpressions: []corev1.NodeSelectorRequirement{
							{Key: "existing", Operator: corev1.NodeSelectorOpExists},
						}}},
					},
				}
				externalWorkload.Spec.NodeAffinities = []v1alpha1.NodeAffinityEntry{{
					SpecName:     v1alpha1.DefaultNodeAffinitySpecName,
					NodeAffinity: existing,
				}}
				podTemplate = buildPodTemplateWithRequiredAffinity([]corev1.NodeSelectorTerm{
					{MatchExpressions: []corev1.NodeSelectorRequirement{affinityMatch}},
				})
				FillExternalWorkloadDataFromPodTemplateSpecIfEmpty(ctx, metav1.ObjectMeta{}, emptyGroupKind, podTemplate, nil, nil, nil, externalWorkload)
				Expect(findEntry(v1alpha1.DefaultNodeAffinitySpecName).NodeAffinity).To(Equal(existing))
			})

			It("IfEmpty variant adds an entry when none exists for that specName", func() {
				presetNodePools()
				podTemplate = buildPodTemplateWithRequiredAffinity([]corev1.NodeSelectorTerm{
					{MatchExpressions: []corev1.NodeSelectorRequirement{affinityMatch}},
				})
				FillExternalWorkloadDataFromPodTemplateSpecIfEmpty(ctx, metav1.ObjectMeta{}, emptyGroupKind, podTemplate, nil, nil, nil, externalWorkload)
				entry := findEntry(v1alpha1.DefaultNodeAffinitySpecName)
				Expect(entry).ToNot(BeNil())
				Expect(entry.NodeAffinity).ToNot(BeNil())
				Expect(entry.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms).To(HaveLen(1))
			})
		})
	})

	Describe("test CalculatePriorityClassNameAndPreemptibility", func() {
		var (
			cachedReader client.WithWatch
		)

		BeforeAll(func() {
			// Create a ConfigMap with workload type configurations
			workloadConfigMap := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      commonconstants.WorkloadsConfigMapName,
					Namespace: commonconstants.WorkloadsConfigMapNamespace,
				},
				Data: map[string]string{
					commonconstants.WorkloadsConfigMapTypesKey: `[
						{
							"TypeName": "Deployment",
							"CategoryName": "deployment",
							"PriorityName": "deployment-priority",
							"Preemptibility": "preemptible"
						},
						{
							"TypeName": "Deployment.apps",
							"CategoryName": "deployment",
							"PriorityName": "deployment-priority",
							"Preemptibility": "preemptible"
						},
						{
							"TypeName": "Job",
							"CategoryName": "job",
							"PriorityName": "job-priority",
							"Preemptibility": "non-preemptible"
						}
					]`,
				},
			}

			scheme := scheme_init.Scheme()
			cachedReader = fake.NewClientBuilder().WithScheme(scheme).WithObjects(workloadConfigMap).Build()
		})

		It("should return priority from labels and preemptibility from label when present", func() {
			topOwnerLabels := map[string]string{
				commonconstants.LabelPriorityClassName:  "high-priority",
				commonconstants.LabelPreemptibilityName: "non-preemptible",
			}
			priority, preemptibility := CalculatePriorityClassNameAndPreemptibility(ctx, cachedReader, topOwnerLabels, schema.GroupKind{Kind: "Deployment"}, "")
			Expect(priority).To(Equal("high-priority"))
			Expect(preemptibility).To(Equal("non-preemptible"))
		})

		It("should return priority from labels and default preemptibility when present", func() {
			topOwnerLabels := map[string]string{
				commonconstants.LabelPriorityClassName: "high-priority",
			}
			priority, preemptibility := CalculatePriorityClassNameAndPreemptibility(ctx, cachedReader, topOwnerLabels, schema.GroupKind{Kind: "Deployment"}, "")
			Expect(priority).To(Equal("high-priority"))
			Expect(preemptibility).To(Equal("preemptible"))
		})

		It("should return default priority and preemptibility from label when present", func() {
			topOwnerLabels := map[string]string{
				commonconstants.LabelPreemptibilityName: "non-preemptible",
			}
			priority, preemptibility := CalculatePriorityClassNameAndPreemptibility(ctx, cachedReader, topOwnerLabels, schema.GroupKind{Kind: "Deployment"}, "")
			Expect(priority).To(Equal("deployment-priority"))
			Expect(preemptibility).To(Equal("non-preemptible"))
		})

		It("should return priority from pod spec and preemptibility (default) when no label priority", func() {
			topOwnerLabels := map[string]string{}
			priority, preemptibility := CalculatePriorityClassNameAndPreemptibility(ctx, cachedReader, topOwnerLabels, schema.GroupKind{Kind: "Deployment"}, "medium-priority")
			Expect(priority).To(Equal("medium-priority"))
			Expect(preemptibility).To(Equal("preemptible"))
		})

		It("should return default priority and preemptibility from workload config when no label or pod spec priority", func() {
			topOwnerLabels := map[string]string{}
			priority, preemptibility := CalculatePriorityClassNameAndPreemptibility(ctx, cachedReader, topOwnerLabels, schema.GroupKind{Kind: "Deployment"}, "")
			Expect(priority).To(Equal("deployment-priority"))
			Expect(preemptibility).To(Equal("preemptible"))
		})

		It("should return default priority and preemptibility for Job type", func() {
			topOwnerLabels := map[string]string{}
			priority, preemptibility := CalculatePriorityClassNameAndPreemptibility(ctx, cachedReader, topOwnerLabels, schema.GroupKind{Kind: "Job"}, "")
			Expect(priority).To(Equal("job-priority"))
			Expect(preemptibility).To(Equal("non-preemptible"))
		})

		It("should return empty priority and preemptibility when kind is empty", func() {
			topOwnerLabels := map[string]string{}
			priority, preemptibility := CalculatePriorityClassNameAndPreemptibility(ctx, cachedReader, topOwnerLabels, schema.GroupKind{}, "")
			Expect(priority).To(BeEmpty())
			Expect(preemptibility).To(BeEmpty())
		})

		It("should return empty priority and preemptibility when configmap is not found", func() {
			// Create a new client without the ConfigMap
			scheme := scheme_init.Scheme()
			emptyReader := fake.NewClientBuilder().WithScheme(scheme).Build()

			topOwnerLabels := map[string]string{}
			priority, preemptibility := CalculatePriorityClassNameAndPreemptibility(ctx, emptyReader, topOwnerLabels, schema.GroupKind{Kind: "Deployment"}, "")
			Expect(priority).To(BeEmpty())
			Expect(preemptibility).To(BeEmpty())
		})
	})
})

var _ = Describe("storage refs extraction", func() {
	pvcVol := func(name, claim string) corev1.Volume {
		return corev1.Volume{Name: name, VolumeSource: corev1.VolumeSource{
			PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: claim}}}
	}
	secretVol := func(name, secret string) corev1.Volume {
		return corev1.Volume{Name: name, VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{SecretName: secret}}}
	}
	configMapVol := func(name, cm string) corev1.Volume {
		return corev1.Volume{Name: name, VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: cm}}}}
	}
	template := func(volumes []corev1.Volume, imagePullSecrets ...string) *corev1.PodTemplateSpec {
		var ips []corev1.LocalObjectReference
		for _, s := range imagePullSecrets {
			ips = append(ips, corev1.LocalObjectReference{Name: s})
		}
		return &corev1.PodTemplateSpec{Spec: corev1.PodSpec{Volumes: volumes, ImagePullSecrets: ips}}
	}
	pvc := func(name string) v1alpha1.StorageRef {
		return v1alpha1.StorageRef{Kind: v1alpha1.StorageRefKindPVC, Name: name}
	}
	secret := func(name string) v1alpha1.StorageRef {
		return v1alpha1.StorageRef{Kind: v1alpha1.StorageRefKindSecret, Name: name}
	}
	configMap := func(name string) v1alpha1.StorageRef {
		return v1alpha1.StorageRef{Kind: v1alpha1.StorageRefKindConfigMap, Name: name}
	}

	DescribeTable("extractStorageRefs",
		func(podTemplate *corev1.PodTemplateSpec, expected []v1alpha1.StorageRef) {
			Expect(extractStorageRefs(podTemplate)).To(Equal(expected))
		},
		Entry("nil template yields no refs", (*corev1.PodTemplateSpec)(nil), []v1alpha1.StorageRef(nil)),
		Entry("no volumes yields no refs", template(nil), []v1alpha1.StorageRef(nil)),
		Entry("pvc volume yields a PVC ref",
			template([]corev1.Volume{pvcVol("v", "data")}), []v1alpha1.StorageRef{pvc("data")}),
		Entry("secret volume yields a Secret ref",
			template([]corev1.Volume{secretVol("v", "creds")}), []v1alpha1.StorageRef{secret("creds")}),
		Entry("configMap volume yields a ConfigMap ref",
			template([]corev1.Volume{configMapVol("v", "cfg")}), []v1alpha1.StorageRef{configMap("cfg")}),
		Entry("imagePullSecrets yield Secret refs",
			template(nil, "regcred"), []v1alpha1.StorageRef{secret("regcred")}),
		Entry("empty claim name is skipped",
			template([]corev1.Volume{pvcVol("v", "")}), []v1alpha1.StorageRef(nil)),
		Entry("duplicates across volumes and imagePullSecrets are deduped",
			template([]corev1.Volume{secretVol("a", "creds"), secretVol("b", "creds")}, "creds"),
			[]v1alpha1.StorageRef{secret("creds")}),
		Entry("mixed kinds preserve volume order then imagePullSecrets",
			template([]corev1.Volume{pvcVol("a", "data"), configMapVol("b", "cfg")}, "regcred"),
			[]v1alpha1.StorageRef{pvc("data"), configMap("cfg"), secret("regcred")}),
	)

	resolvedPvc := func(name, uid string) v1alpha1.StorageRef {
		return v1alpha1.StorageRef{Kind: v1alpha1.StorageRefKindPVC, Name: name, UID: uid}
	}
	resolvedSecret := func(name, uid string) v1alpha1.StorageRef {
		return v1alpha1.StorageRef{Kind: v1alpha1.StorageRefKindSecret, Name: name, UID: uid}
	}

	It("UpsertStorageRefsFromPodTemplate unions across templates, resolves UIDs, and dedups", func() {
		reader := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).WithObjects(
			&corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "head-data", Namespace: "ns", UID: "uid-head"}},
			&corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "worker-data", Namespace: "ns", UID: "uid-worker"}},
			&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "regcred", Namespace: "ns", UID: "uid-regcred"}},
		).Build()

		ew := &v1alpha1.ExternalWorkload{ObjectMeta: metav1.ObjectMeta{Namespace: "ns"}}
		UpsertStorageRefsFromPodTemplate(context.Background(), template([]corev1.Volume{pvcVol("a", "head-data")}, "regcred"), reader, ew)
		UpsertStorageRefsFromPodTemplate(context.Background(), template([]corev1.Volume{pvcVol("a", "worker-data")}, "regcred"), reader, ew)
		Expect(ew.Spec.StorageRefs).To(Equal([]v1alpha1.StorageRef{
			resolvedPvc("head-data", "uid-head"), resolvedSecret("regcred", "uid-regcred"), resolvedPvc("worker-data", "uid-worker"),
		}))
	})

	It("UpsertStorageRefsFromPodTemplate drops refs whose resource is not found", func() {
		reader := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).WithObjects(
			&corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "exists", Namespace: "ns", UID: "uid-exists"}},
		).Build()

		ew := &v1alpha1.ExternalWorkload{ObjectMeta: metav1.ObjectMeta{Namespace: "ns"}}
		UpsertStorageRefsFromPodTemplate(context.Background(),
			template([]corev1.Volume{pvcVol("a", "exists"), pvcVol("b", "missing")}), reader, ew)
		Expect(ew.Spec.StorageRefs).To(Equal([]v1alpha1.StorageRef{resolvedPvc("exists", "uid-exists")}))
	})

	It("SetStorageRefsFromPodTemplate drops stale refs then unions the remaining templates", func() {
		reader := fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).WithObjects(
			&corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "head-data", Namespace: "ns", UID: "uid-head"}},
			&corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "worker-data", Namespace: "ns", UID: "uid-worker"}},
		).Build()

		ew := &v1alpha1.ExternalWorkload{
			ObjectMeta: metav1.ObjectMeta{Namespace: "ns"},
			Spec:       v1alpha1.ExternalWorkloadSpec{StorageRefs: []v1alpha1.StorageRef{resolvedPvc("gone", "uid-gone")}},
		}
		SetStorageRefsFromPodTemplate(context.Background(), template([]corev1.Volume{pvcVol("a", "head-data")}), reader, ew)
		UpsertStorageRefsFromPodTemplate(context.Background(), template([]corev1.Volume{pvcVol("a", "worker-data")}), reader, ew)
		Expect(ew.Spec.StorageRefs).To(Equal([]v1alpha1.StorageRef{
			resolvedPvc("head-data", "uid-head"), resolvedPvc("worker-data", "uid-worker"),
		}))
	})
})

func expectedRequestedResources(numOfReplicas int64) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:              *resource.NewQuantity(numOfReplicas*cpuLimit, resource.DecimalSI),
			corev1.ResourceMemory:           *resource.NewQuantity(numOfReplicas*memLimit, resource.DecimalSI),
			constants.GpuResourceName:       *resource.NewMilliQuantity(numOfReplicas*int64(1000*gpuLimit), resource.DecimalSI),
			constants.GpuMemoryResourceName: *resource.NewQuantity(numOfReplicas*gpuMemLimit*test.GpuMemoryLimitConversion, resource.DecimalSI),
			"nvidia.com/mig-some-mig":       *resource.NewQuantity(numOfReplicas*someMigLimit, resource.DecimalSI),
		},
		Requests: map[corev1.ResourceName]resource.Quantity{
			corev1.ResourceCPU:              *resource.NewQuantity(numOfReplicas*cpuRequest, resource.DecimalSI),
			corev1.ResourceMemory:           *resource.NewQuantity(numOfReplicas*memRequest, resource.DecimalSI),
			constants.GpuResourceName:       *resource.NewQuantity(numOfReplicas*int64(1000*gpuRequest), resource.DecimalSI),
			constants.GpuMemoryResourceName: *resource.NewQuantity(numOfReplicas*gpuMemLimit*test.GpuMemoryLimitConversion, resource.DecimalSI),
			"nvidia.com/mig-some-mig":       *resource.NewQuantity(numOfReplicas*someMigLimit, resource.DecimalSI),
		},
	}
}
