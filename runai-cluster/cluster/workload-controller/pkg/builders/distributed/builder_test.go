/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributed

import (
	"context"
	"errors"
	"fmt"
	mnnvlcommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders/mnnvl"
	"reflect"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	mnnvlapi "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/mnnvl"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	nodepoolbuilders "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/builders/nodepool"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/internal/mnnvlutils"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/testing"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/factory"
	"k8s.io/utils/ptr"

	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/topology"
	workloadproject "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/internal/autogenerate/mocks"
	workloadCommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	fakeClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	core1 "k8s.io/api/core/v1"

	dwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	builder "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common"

	"strings"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/pytorch"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultImage     = "bla"
	testWorkLoadName = "test-workload"
	testNamespace    = "runai-test-namespace"
	nodePoolName
)

var _ = Describe("Distributed Job builders", func() {
	var (
		computeBuilder    *ComputeBuilder
		ctx               context.Context
		baseWorkload      *jobs.JobWorkload
		masterWorkload    *jobs.JobWorkload
		k8client          client.Client
		exitingPodsSpec   map[workloadCommon.ReplicaType]*core1.PodSpec
		expectedAffinity  *core1.Affinity
		internalBuilderFn func(fromObject interface{}, toResource interface{},
			k8sClient client.Client, options factory.BuilderOptions,
			existingPodsSpec map[workloadCommon.ReplicaType]*core1.PodSpec,
			podLabels map[string]string, podAnnotations map[string]string) (factory.Builder, error)
	)
	BeforeEach(func() {
		ctx = workloadproject.SetProjectInContext(context.Background(), strings.TrimPrefix(testNamespace, "runai-"))
		project := resource.CreateProject(testNamespace)
		nodepool := resource.CreateNodePool(nodePoolName, "gpu", "2323")
		nodepoolKai := resource.CreateKaiNodePool(nodePoolName, "gpu", "2323")
		scheme := runtime.NewScheme()
		register.InitSchema(scheme)
		k8client = fakeClient.InitFakeClientWithObjectNoStatus(scheme, project, nodepool, nodepoolKai)
	})

	BeforeEach(func() {
		baseWorkload = workload.NewJobWorkloadBuilder().
			WithAnnotations(nil).
			WithDistribution().
			WithImage(defaultImage).
			AsPyTorch().
			WithPorts([]string{"123", "2333"}).
			WithNodePools(nodePoolName).
			Build()

		masterWorkload = workload.NewJobWorkloadBuilder().
			WithName(testWorkLoadName).
			WithNamespace(testNamespace).
			AsPyTorch().
			WithImage(defaultImage).
			WithNodePools(nodePoolName).
			WithMasterSpec(&baseWorkload.Spec).
			WithDistribution().
			WithWorkers(2).
			Build()

	})
	Context("Compute builder", func() {
		BeforeEach(func() {
			internalBuilderFn = pytorch.NewPyTorchBuilderWithExistingNodeAffinities
			computeBuilder = NewComputeBuilder(constants.KindPyTorchJob, internalBuilderFn, k8client, pytorch.NewPytorchJobFn, nil, nil, true)
			expectedAffinity = &core1.Affinity{}
			err := nodepoolbuilders.CreateNodeAffinityForNodePools(ctx, k8client, masterWorkload.Spec.NodePools.Deref(), masterWorkload.Spec.NodePool.Deref(), &expectedAffinity)
			Expect(err).To(BeNil())

		})
		Context("Preview", func() {
			It("Should return the correct preview", func() {
				previews, err := computeBuilder.GetPreview(ctx, masterWorkload)
				Expect(err).To(BeNil())
				Expect(previews).ToNot(BeNil())
				Expect(previews).To(HaveLen(1))
				previewJob := previews[0]
				Expect(previewJob).ToNot(BeNil())
				Expect(previewJob.Name).To(Equal(testWorkLoadName))
				Expect(previewJob.Namespace).To(Equal(testNamespace))
				Expect(previewJob.Kind.Kind).To(Equal(constants.KindPyTorchJob))
			})
			It("should return the same name as the real build", func() {
				previews, err := computeBuilder.GetPreview(ctx, masterWorkload)
				Expect(err).To(BeNil())
				Expect(previews).ToNot(BeNil())
				Expect(previews).To(HaveLen(1))

				previewJob := previews[0]
				computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{})
				Expect(err).To(BeNil())
				Expect(computeResources).ToNot(BeNil())

				realJob := computeResources.Job
				Expect(realJob).ToNot(BeNil())
				Expect(previewJob.Name).To(Equal(realJob.GetName()))
				Expect(previewJob.Namespace).To(Equal(realJob.GetNamespace()))
			})

			Context("when enabled MNNVL", func() {
				BeforeEach(func() {
					ctx = enableMnnvl(ctx)
				})

				It("should return ComputeDomain when enabled", func() {
					previews, err := computeBuilder.GetPreview(ctx, masterWorkload)
					Expect(err).To(BeNil())
					Expect(previews).ToNot(BeNil())
					Expect(previews).To(HaveLen(2))

					cd := previews[1]
					Expect(cd.Kind.Kind).To(Equal(nvapi.ComputeDomainKind))
					Expect(cd.Name).To(Equal(fmt.Sprintf("dw-%s", testWorkLoadName)))
					Expect(cd.Namespace).To(Equal(masterWorkload.Namespace))
				})
			})
		})
		Context("Build", func() {
			Context("Labels and annotations", func() {
				var (
					defaultPodLabels      map[string]string
					defaultPodAnnotations map[string]string
				)

				It("should have custom labels", func() {
					podLabels := map[string]string{
						"label": "value",
					}

					computeBuilder = NewComputeBuilder(constants.KindPyTorchJob, internalBuilderFn, k8client,
						pytorch.NewPytorchJobFn, podLabels, defaultPodAnnotations, true)
					computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{})
					Expect(err).ToNot(HaveOccurred())
					ptj := computeResources.Job.(*kubeflowv1.PyTorchJob)
					Expect(ptj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.GetLabels()).ToNot(BeNil())
					Expect(ptj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.GetLabels()["label"]).To(Equal("value"))
					Expect(ptj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.GetLabels()).ToNot(BeNil())
					Expect(ptj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.GetLabels()["label"]).To(Equal("value"))
				})

				It("should have custom annotations", func() {
					podAnnotations := map[string]string{
						"annotation": "value",
					}

					computeBuilder = NewComputeBuilder(constants.KindPyTorchJob, internalBuilderFn, k8client,
						pytorch.NewPytorchJobFn, defaultPodLabels, podAnnotations, true)
					computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{})
					Expect(err).ToNot(HaveOccurred())
					ptj := computeResources.Job.(*kubeflowv1.PyTorchJob)
					Expect(ptj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.GetAnnotations()).ToNot(BeNil())
					Expect(ptj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.GetAnnotations()["annotation"]).To(Equal("value"))
					Expect(ptj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.GetAnnotations()).ToNot(BeNil())
					Expect(ptj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.GetAnnotations()["annotation"]).To(Equal("value"))
				})

				It("should project segmented topology annotations to the correct PyTorchJob levels", func() {
					masterWorkload.Annotations = map[string]string{
						topology.KAITopologyAnnotationKey:          "network-topology",
						topology.KAITopologyAnnotationPreferredKey: "topology.kubernetes.io/rack",
					}
					masterWorkload.Spec.Annotations = field_types.NewDictionaryField(map[string]string{
						topology.KAITopologyAnnotationKey:                 "network-topology",
						topology.KAITopologyAnnotationPreferredKey:        "topology.kubernetes.io/rack",
						topology.KAISegmentSizeAnnotationKey:              "2",
						topology.KAISegmentTopologyPreferredAnnotationKey: "topology.kubernetes.io/rack",
					})

					computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{})
					Expect(err).ToNot(HaveOccurred())
					ptj := computeResources.Job.(*kubeflowv1.PyTorchJob)
					rootAnnotations := ptj.GetAnnotations()
					workerAnnotations := ptj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.GetAnnotations()
					masterAnnotations := ptj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.GetAnnotations()
					Expect(rootAnnotations).To(HaveKeyWithValue(topology.KAITopologyAnnotationKey, "network-topology"))
					Expect(rootAnnotations).To(HaveKeyWithValue(topology.KAITopologyAnnotationPreferredKey, "topology.kubernetes.io/rack"))
					Expect(rootAnnotations).NotTo(HaveKey(topology.KAISegmentSizeAnnotationKey))
					Expect(rootAnnotations).NotTo(HaveKey(topology.KAISegmentTopologyPreferredAnnotationKey))

					Expect(workerAnnotations).To(HaveKeyWithValue(topology.KAITopologyAnnotationKey, "network-topology"))
					Expect(workerAnnotations).To(HaveKeyWithValue(topology.KAISegmentSizeAnnotationKey, "2"))
					Expect(workerAnnotations).To(HaveKeyWithValue(topology.KAISegmentTopologyPreferredAnnotationKey, "topology.kubernetes.io/rack"))
					Expect(workerAnnotations).NotTo(HaveKey(topology.KAITopologyAnnotationPreferredKey))

					Expect(masterAnnotations).To(HaveKeyWithValue(topology.KAITopologyAnnotationKey, "network-topology"))
					Expect(masterAnnotations).NotTo(HaveKey(topology.KAISegmentSizeAnnotationKey))
					Expect(masterAnnotations).NotTo(HaveKey(topology.KAISegmentTopologyPreferredAnnotationKey))
					Expect(masterAnnotations).NotTo(HaveKey(topology.KAITopologyAnnotationPreferredKey))
				})
			})
			It("Should return the correct job", func() {
				computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{})
				Expect(err).To(BeNil())
				Expect(computeResources).ToNot(BeNil())

				job := computeResources.Job
				Expect(job).ToNot(BeNil())
				Expect(job.GetName()).To(Equal(testWorkLoadName))
				Expect(job.GetNamespace()).To(Equal(testNamespace))
				Expect(job.GetObjectKind().GroupVersionKind().Kind).To(Equal(constants.KindPyTorchJob))
			})
			It("Should return the correct job with owner", func() {
				owner := dwv2.NewFromJobWorkload(masterWorkload)
				computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{
					Owner: owner,
				})
				Expect(err).To(BeNil())
				Expect(computeResources).ToNot(BeNil())

				job := computeResources.Job
				Expect(job).ToNot(BeNil())
				Expect(job.GetName()).To(Equal(testWorkLoadName))
				Expect(job.GetNamespace()).To(Equal(testNamespace))
				Expect(job.GetOwnerReferences()).To(HaveLen(1))
				Expect(job.GetOwnerReferences()[0].Name).To(Equal(owner.GetName()))
			})
			It("should return err when owner is not valid", func() {
				_, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{
					Owner: masterWorkload,
				})
				Expect(err).ToNot(BeNil())
			})
			It("should build with node affinity from nodepool", func() {
				expectedAffinity := &core1.Affinity{}
				err := nodepoolbuilders.CreateNodeAffinityForNodePools(ctx, k8client, masterWorkload.Spec.NodePools.Deref(), masterWorkload.Spec.NodePool.Deref(), &expectedAffinity)
				Expect(err).To(BeNil())
				computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{})
				Expect(err).To(BeNil())
				Expect(computeResources).ToNot(BeNil())

				jobObj := computeResources.Job
				Expect(jobObj).ToNot(BeNil())
				pytorchJob, ok := jobObj.(*kubeflowv1.PyTorchJob)
				Expect(ok).To(BeTrue())
				Expect(reflect.DeepEqual(pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.Spec.Affinity.NodeAffinity, expectedAffinity.NodeAffinity)).To(BeTrue())
				Expect(reflect.DeepEqual(pytorchJob.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.Spec.Affinity.NodeAffinity, expectedAffinity.NodeAffinity)).To(BeTrue())

			})
			Context("worker has existing node affinity", func() {
				BeforeEach(func() {
					exitingPodsSpec = make(map[workloadCommon.ReplicaType]*core1.PodSpec)
					exitingPodsSpec[workloadCommon.Worker] = &core1.PodSpec{
						Affinity: &core1.Affinity{
							NodeAffinity: &core1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &core1.NodeSelector{
									NodeSelectorTerms: []core1.NodeSelectorTerm{
										{
											MatchExpressions: []core1.NodeSelectorRequirement{
												{
													Key:      "other-key",
													Operator: core1.NodeSelectorOpIn,
													Values:   []string{"test-value"},
												},
											},
										},
									},
								},
							},
						},
					}

				})
				It("should build with existing node affinity for workload", func() {
					computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{ExitingPodsSpec: exitingPodsSpec})
					Expect(err).To(BeNil())
					Expect(computeResources).ToNot(BeNil())

					job := computeResources.Job
					jobObj, ok := job.(*kubeflowv1.PyTorchJob)
					Expect(ok).To(BeTrue())

					Expect(reflect.DeepEqual(jobObj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].
						Template.Spec.Affinity.NodeAffinity, expectedAffinity.NodeAffinity)).To(BeTrue())
					Expect(reflect.DeepEqual(jobObj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].
						Template.Spec.Affinity.NodeAffinity, exitingPodsSpec[workloadCommon.Worker].Affinity.NodeAffinity)).To(BeTrue())
				})
			})
			Context("master has existing node affinity", func() {
				BeforeEach(func() {
					exitingPodsSpec = make(map[workloadCommon.ReplicaType]*core1.PodSpec)
					exitingPodsSpec[workloadCommon.Master] = &core1.PodSpec{
						Affinity: &core1.Affinity{
							NodeAffinity: &core1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &core1.NodeSelector{
									NodeSelectorTerms: []core1.NodeSelectorTerm{
										{
											MatchExpressions: []core1.NodeSelectorRequirement{
												{
													Key:      "other-key",
													Operator: core1.NodeSelectorOpIn,
													Values:   []string{"test-value"},
												},
											},
										},
									},
								},
							},
						},
					}

				})
				It("should build with existing node affinity for workload", func() {
					computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{ExitingPodsSpec: exitingPodsSpec})
					Expect(err).To(BeNil())
					Expect(computeResources).ToNot(BeNil())

					job := computeResources.Job
					Expect(err).To(BeNil())
					jobObj, ok := job.(*kubeflowv1.PyTorchJob)
					Expect(ok).To(BeTrue())

					Expect(reflect.DeepEqual(jobObj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].
						Template.Spec.Affinity.NodeAffinity, exitingPodsSpec[workloadCommon.Master].Affinity.NodeAffinity)).To(BeTrue())
					Expect(reflect.DeepEqual(jobObj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].
						Template.Spec.Affinity.NodeAffinity, expectedAffinity.NodeAffinity)).To(BeTrue())
				})
			})
			Context("master job has old image", func() {
				BeforeEach(func() {
					exitingPodsSpec = make(map[workloadCommon.ReplicaType]*core1.PodSpec)
					exitingPodsSpec[workloadCommon.Master] = &core1.PodSpec{
						Containers: []core1.Container{
							{
								Image: "old-image",
							},
						},
					}
				})
				It("should build with existing image for workload", func() {
					computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{ExitingPodsSpec: exitingPodsSpec})
					Expect(err).To(BeNil())
					Expect(computeResources).ToNot(BeNil())

					job := computeResources.Job
					jobObj, ok := job.(*kubeflowv1.PyTorchJob)
					Expect(ok).To(BeTrue())

					Expect(jobObj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].
						Template.Spec.Containers[0].Image).To(Equal("old-image"))
				})
				It("should use the new image for the worker", func() {
					computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{ExitingPodsSpec: exitingPodsSpec})
					Expect(err).To(BeNil())
					Expect(computeResources).ToNot(BeNil())

					job := computeResources.Job
					jobObj, ok := job.(*kubeflowv1.PyTorchJob)
					Expect(ok).To(BeTrue())

					Expect(jobObj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].
						Template.Spec.Containers[0].Image).To(Equal(defaultImage))
				})
			})
			Context("worker job has old image", func() {
				BeforeEach(func() {
					exitingPodsSpec = make(map[workloadCommon.ReplicaType]*core1.PodSpec)
					exitingPodsSpec[workloadCommon.Worker] = &core1.PodSpec{
						Containers: []core1.Container{
							{
								Image: "old-image",
							},
						},
					}
				})
				It("should build with existing image for worker", func() {
					computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{ExitingPodsSpec: exitingPodsSpec})
					Expect(err).To(BeNil())
					Expect(computeResources).ToNot(BeNil())

					job := computeResources.Job
					jobObj, ok := job.(*kubeflowv1.PyTorchJob)
					Expect(ok).To(BeTrue())

					Expect(jobObj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].
						Template.Spec.Containers[0].Image).To(Equal("old-image"))
				})
				It("should use the new image for the master", func() {
					computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{ExitingPodsSpec: exitingPodsSpec})
					Expect(err).To(BeNil())
					Expect(computeResources).ToNot(BeNil())

					job := computeResources.Job
					jobObj, ok := job.(*kubeflowv1.PyTorchJob)
					Expect(ok).To(BeTrue())

					Expect(jobObj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].
						Template.Spec.Containers[0].Image).To(Equal(defaultImage))
				})
			})
			Context("master and worker has old image", func() {
				BeforeEach(func() {
					exitingPodsSpec = make(map[workloadCommon.ReplicaType]*core1.PodSpec)
					exitingPodsSpec[workloadCommon.Master] = &core1.PodSpec{
						Containers: []core1.Container{
							{
								Image: "old-image",
							},
						},
					}
					exitingPodsSpec[workloadCommon.Worker] = &core1.PodSpec{
						Containers: []core1.Container{
							{
								Image: "old-image",
							},
						},
					}
				})
				It("should build with existing image for workload", func() {
					computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{ExitingPodsSpec: exitingPodsSpec})
					Expect(err).To(BeNil())
					Expect(computeResources).ToNot(BeNil())

					job := computeResources.Job
					jobObj, ok := job.(*kubeflowv1.PyTorchJob)
					Expect(ok).To(BeTrue())

					Expect(jobObj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].
						Template.Spec.Containers[0].Image).To(Equal("old-image"))
					Expect(jobObj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].
						Template.Spec.Containers[0].Image).To(Equal("old-image"))
				})
			})
			Context("master and worker has existing node affinity", func() {
				BeforeEach(func() {
					exitingPodsSpec = make(map[workloadCommon.ReplicaType]*core1.PodSpec)
					exitingPodsSpec[workloadCommon.Master] = &core1.PodSpec{
						Affinity: &core1.Affinity{
							NodeAffinity: &core1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &core1.NodeSelector{
									NodeSelectorTerms: []core1.NodeSelectorTerm{
										{
											MatchExpressions: []core1.NodeSelectorRequirement{
												{
													Key:      "other-key",
													Operator: core1.NodeSelectorOpIn,
													Values:   []string{"test-value"},
												},
											},
										},
									},
								},
							},
						},
					}
					exitingPodsSpec[workloadCommon.Worker] = &core1.PodSpec{
						Affinity: &core1.Affinity{
							NodeAffinity: &core1.NodeAffinity{
								RequiredDuringSchedulingIgnoredDuringExecution: &core1.NodeSelector{
									NodeSelectorTerms: []core1.NodeSelectorTerm{
										{
											MatchExpressions: []core1.NodeSelectorRequirement{
												{
													Key:      "other-key",
													Operator: core1.NodeSelectorOpIn,
													Values:   []string{"test-value"},
												},
											},
										},
									},
								},
							},
						},
					}

				})
				It("should build with existing node affinity for workload", func() {
					computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{ExitingPodsSpec: exitingPodsSpec})
					Expect(err).To(BeNil())
					Expect(computeResources).ToNot(BeNil())

					job := computeResources.Job
					jobObj, ok := job.(*kubeflowv1.PyTorchJob)
					Expect(ok).To(BeTrue())

					Expect(reflect.DeepEqual(jobObj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].
						Template.Spec.Affinity.NodeAffinity, exitingPodsSpec[workloadCommon.Master].Affinity.NodeAffinity)).To(BeTrue())
					Expect(reflect.DeepEqual(jobObj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].
						Template.Spec.Affinity.NodeAffinity, exitingPodsSpec[workloadCommon.Worker].Affinity.NodeAffinity)).To(BeTrue())
				})
			})

			Context("when enabled MNNVL", func() {
				BeforeEach(func() {
					ctx = enableMnnvl(ctx)

				})

				It("should return nil if error occurred", func() {
					computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{})
					Expect(err).NotTo(BeNil())
					Expect(errors.Is(err, mnnvlcommon.ComputeDomainCpuOnlyErr)).To(BeTrue())

					Expect(computeResources).To(BeNil())
				})

				Context("when requested GPU", func() {
					BeforeEach(func() {
						masterWorkload.Spec.ResourceAllocationFields = field_groups.ResourceAllocationFields{
							GpuDevices: field_types.NewOptionalIntField(ptr.To(1)),
						}
					})

					It("should build ComputeDomain", func() {
						computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{})
						Expect(err).To(BeNil())
						Expect(computeResources).ToNot(BeNil())

						cd := computeResources.ComputeDomain
						Expect(cd).ToNot(BeNil())
					})
					It("should built ComputeDomain should be match the preview", func() {
						previews, err := computeBuilder.GetPreview(ctx, masterWorkload)
						Expect(err).To(BeNil())
						Expect(previews).ToNot(BeNil())
						Expect(previews).To(HaveLen(2))
						previewCd := previews[1]

						computeResources, err := computeBuilder.Build(ctx, masterWorkload, &builder.ComputeBuilderOptions{})
						Expect(err).To(BeNil())
						Expect(computeResources).ToNot(BeNil())

						actualCd := computeResources.ComputeDomain
						Expect(actualCd).ToNot(BeNil())

						Expect(previewCd.Name).To(Equal(actualCd.Name))
						Expect(previewCd.Kind.Kind).To(Equal(actualCd.Kind))
						Expect(previewCd.Namespace).To(Equal(actualCd.Namespace))
					})
				})
			})
		})
	})
	Context("DistributedJobsBuilders", func() {
		var (
			distributedFactory *ComputeBuilderFactory
			mockComputeBuilder mocks.ComputeBuilder
		)
		BeforeEach(func() {

			mockComputeBuilder = mocks.ComputeBuilder{}
			builders := map[string]builder.ComputeBuilder{
				"mock": &mockComputeBuilder,
			}

			distributedFactory = NewComputeBuilderFactory(builders)

		})
		Context("Build", func() {
			It("Should return the correct job", func() {
				computeResources := &builder.ComputeResources{
					Job: pytorch.NewPytorchJobFn(),
				}
				baseWorkload.Spec.JobType = "mock"
				mockComputeBuilder.On("Build", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(computeResources, nil)
				computeResources, err := distributedFactory.Build(ctx, baseWorkload, &builder.ComputeBuilderOptions{})
				Expect(err).To(BeNil())
				Expect(computeResources).ToNot(BeNil())

				job := computeResources.Job.(*kubeflowv1.PyTorchJob)
				Expect(job).ToNot(BeNil())
				Expect(job).To(Equal(job))
			})
			It("Should return err when no builder found", func() {
				job, err := distributedFactory.Build(ctx, baseWorkload, &builder.ComputeBuilderOptions{})
				Expect(err).ToNot(BeNil())
				Expect(job).To(BeNil())
			})

		})
	})

})

func enableMnnvl(ctx context.Context) context.Context {
	mnnvlEnabledStatus := mnnvlutils.Status{
		WorkerMode: mnnvlapi.ModeRequired,
		MasterMode: mnnvlapi.ModePreferred,
	}
	return testing.AddMnnvlToContextWrapper(ctx, mnnvlEnabledStatus)
}
