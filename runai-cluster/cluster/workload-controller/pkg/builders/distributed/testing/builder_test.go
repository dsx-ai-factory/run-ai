/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"context"
	kubeflowv1 "github.com/kubeflow/training-operator/pkg/apis/kubeflow.org/v1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/factory"
	corev1 "k8s.io/api/core/v1"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	workloadproject "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/internal/autogenerate/mocks"
	dwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	builder "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/common"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/distributed"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/builders/distributed/pytorch"
	workloadCommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	fakeClient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const testWorkLoadName = "test-workload"
const testNamespace = "runai-test-namespace"

var _ = Describe("Distributed Job builders", func() {
	var (
		computeBuilder        *distributed.ComputeBuilder
		ctx                   context.Context
		baseWorkload          *jobs.JobWorkload
		k8client              client.Client
		defaultPodLabels      map[string]string
		defaultPodAnnotations map[string]string
		internalBuilderFn     func(fromObject interface{}, toResource interface{},
			k8sClient client.Client, options factory.BuilderOptions,
			existingPodsSpec map[workloadCommon.ReplicaType]*corev1.PodSpec,
			podLabels map[string]string, podAnnotations map[string]string) (factory.Builder, error)
	)
	BeforeEach(func() {
		ctx = workloadproject.SetProjectInContext(ctx, strings.TrimPrefix(testNamespace, "runai-"))
		scheme := runtime.NewScheme()
		register.InitSchema(scheme)
		k8client = fakeClient.InitFakeClientWithObjectNoStatus(scheme)
		defaultPodLabels = map[string]string{}
		defaultPodAnnotations = map[string]string{}
	})
	BeforeEach(func() {
		baseWorkload = workload.NewJobWorkloadBuilder().WithName(testWorkLoadName).
			WithLabels(nil).WithAnnotations(nil).
			WithDistribution().WithNamespace(testNamespace).WithNoMaster().WithWorkers(3).
			AsPyTorch().WithPorts([]string{"123", "2333"}).Build()
	})
	Context("Job builder", func() {
		BeforeEach(func() {

			internalBuilderFn = pytorch.NewPyTorchBuilderWithExistingNodeAffinities
			computeBuilder = distributed.NewComputeBuilder(constants.KindPyTorchJob, internalBuilderFn, k8client,
				pytorch.NewPytorchJobFn, defaultPodLabels, defaultPodAnnotations, true)

		})
		Context("Preview", func() {
			It("Should return the correct preview", func() {
				previews, err := computeBuilder.GetPreview(ctx, baseWorkload)
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
				previews, err := computeBuilder.GetPreview(ctx, baseWorkload)
				Expect(err).To(BeNil())
				Expect(previews).ToNot(BeNil())
				Expect(previews).To(HaveLen(1))
				previewJob := previews[0]

				computeResources, err := computeBuilder.Build(ctx, baseWorkload, &builder.ComputeBuilderOptions{})
				Expect(err).To(BeNil())
				Expect(computeResources).ToNot(BeNil())

				realJob := computeResources.Job
				Expect(realJob).ToNot(BeNil())
				Expect(previewJob.Name).To(Equal(realJob.GetName()))
				Expect(previewJob.Namespace).To(Equal(realJob.GetNamespace()))
			})
		})
		Context("Build", func() {
			Context("Labels and annotations", func() {
				It("should have custom labels", func() {
					podLabels := map[string]string{
						"label": "value",
					}

					computeBuilder = distributed.NewComputeBuilder(constants.KindPyTorchJob, internalBuilderFn, k8client,
						pytorch.NewPytorchJobFn, podLabels, defaultPodAnnotations, true)
					computeResources, err := computeBuilder.Build(ctx, baseWorkload, &builder.ComputeBuilderOptions{})
					Expect(err).ToNot(HaveOccurred())
					Expect(computeResources).ToNot(BeNil())

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

					computeBuilder = distributed.NewComputeBuilder(constants.KindPyTorchJob, internalBuilderFn, k8client,
						pytorch.NewPytorchJobFn, defaultPodLabels, podAnnotations, true)
					computeResources, err := computeBuilder.Build(ctx, baseWorkload, &builder.ComputeBuilderOptions{})
					Expect(err).ToNot(HaveOccurred())
					Expect(computeResources).ToNot(BeNil())

					ptj := computeResources.Job.(*kubeflowv1.PyTorchJob)
					Expect(ptj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.GetAnnotations()).ToNot(BeNil())
					Expect(ptj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeMaster].Template.GetAnnotations()["annotation"]).To(Equal("value"))
					Expect(ptj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.GetAnnotations()).ToNot(BeNil())
					Expect(ptj.Spec.PyTorchReplicaSpecs[kubeflowv1.PyTorchJobReplicaTypeWorker].Template.GetAnnotations()["annotation"]).To(Equal("value"))
				})
			})
			It("Should return the correct job", func() {
				computeResources, err := computeBuilder.Build(ctx, baseWorkload, &builder.ComputeBuilderOptions{})
				Expect(err).To(BeNil())
				Expect(computeResources).ToNot(BeNil())

				job := computeResources.Job
				Expect(job).ToNot(BeNil())
				Expect(job.GetName()).To(Equal(testWorkLoadName))
				Expect(job.GetNamespace()).To(Equal(testNamespace))
				Expect(job.GetObjectKind().GroupVersionKind().Kind).To(Equal(constants.KindPyTorchJob))
			})
			It("Should return the correct job with owner", func() {
				owner := dwv2.NewFromJobWorkload(baseWorkload)
				computeResources, err := computeBuilder.Build(ctx, baseWorkload, &builder.ComputeBuilderOptions{
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
				_, err := computeBuilder.Build(ctx, baseWorkload, &builder.ComputeBuilderOptions{
					Owner: baseWorkload,
				})
				Expect(err).ToNot(BeNil())
			})

		})
	})
	Context("DistributedJobsBuilders", func() {
		var (
			distributedFactory *distributed.ComputeBuilderFactory
			mockComputeBuilder mocks.ComputeBuilder
		)
		BeforeEach(func() {

			mockComputeBuilder = mocks.ComputeBuilder{}
			builders := map[string]builder.ComputeBuilder{
				"mock": &mockComputeBuilder,
			}

			distributedFactory = distributed.NewComputeBuilderFactory(builders)

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

				job := computeResources.Job
				Expect(job).ToNot(BeNil())
				Expect(job).To(Equal(job))
			})
			It("Should return err when no builder found", func() {
				computeResources, err := distributedFactory.Build(ctx, baseWorkload, &builder.ComputeBuilderOptions{})
				Expect(err).ToNot(BeNil())
				Expect(computeResources).To(BeNil())
			})

		})
	})

})
