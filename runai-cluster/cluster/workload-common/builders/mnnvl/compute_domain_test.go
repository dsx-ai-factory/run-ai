/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	"context"

	nvapi "github.com/NVIDIA/k8s-dra-driver-gpu/api/nvidia.com/resource/v1beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	workloadproject "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/project"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultNamespace = "runai-default"
	projectName      = "default"
	defaultName      = "test-workload"
)

var _ = Describe("ComputeDomain tests", func() {
	var (
		ctx    context.Context
		params ComputeDomainWorkloadParams
		scheme *runtime.Scheme
		owner  client.Object
	)

	BeforeEach(func() {
		// Setup common test variables
		ctx = context.Background()
		ctx = workloadproject.SetProjectInContext(context.Background(), projectName)

		// Initialize default parameters
		params = ComputeDomainWorkloadParams{
			WorkloadKind:      "Job",
			JobType:           "Distributed",
			WorkloadInitials:  "dw",
			WorkloadName:      defaultName,
			WorkloadNamespace: defaultNamespace,
			NumNodes:          2,
		}

		// Initialize scheme
		scheme = runtime.NewScheme()
		Expect(appsv1.AddToScheme(scheme)).To(Succeed())
		Expect(nvapi.AddToScheme(scheme)).To(Succeed())

		// Initialize owner
		owner = &appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Job",
				APIVersion: "batch/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      defaultName,
				Namespace: defaultNamespace,
				UID:       types.UID("test-uid"),
			},
		}
	})

	Context("BuildComputeDomain function", func() {
		It("should successfully build ComputeDomain with NVIDIA GPU vendor with num nodes set to pod number", func() {
			computeDomain, err := BuildComputeDomain(ctx, params, scheme, owner)

			Expect(err).To(BeNil())
			Expect(computeDomain).NotTo(BeNil())
			Expect(computeDomain.Kind).To(Equal(nvapi.ComputeDomainKind))
			Expect(computeDomain.APIVersion).To(Equal(nvapi.Version))
			Expect(computeDomain.Name).To(Equal("dw-test-workload"))
			Expect(computeDomain.Namespace).To(Equal(defaultNamespace))
			Expect(computeDomain.Spec.NumNodes).To(Equal(2))
		})

		It("should successfully build ComputeDomain with NVIDIA GPU vendor with num nodes set to 0", func() {
			params.ElasticEnabled = true
			computeDomain, err := BuildComputeDomain(ctx, params, scheme, owner)

			Expect(err).To(BeNil())
			Expect(computeDomain).NotTo(BeNil())
			Expect(computeDomain.Kind).To(Equal(nvapi.ComputeDomainKind))
			Expect(computeDomain.APIVersion).To(Equal(nvapi.Version))
			Expect(computeDomain.Name).To(Equal("dw-test-workload"))
			Expect(computeDomain.Namespace).To(Equal(defaultNamespace))
			Expect(computeDomain.Spec.NumNodes).To(Equal(0))
		})

		It("should return error for CPU-only workloads", func() {
			params.NumNodes = 0

			computeDomain, err := BuildComputeDomain(ctx, params, scheme, owner)

			Expect(err).To(Equal(ComputeDomainCpuOnlyErr))
			Expect(computeDomain).To(BeNil())
		})

		It("should return error when scheme is nil and owner is provided", func() {
			computeDomain, err := BuildComputeDomain(ctx, params, nil, owner)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(FailedToSetReferenceErrMsg))
			Expect(err.Error()).To(ContainSubstring("scheme is nil"))
			Expect(computeDomain).To(BeNil())
		})

		It("should build ComputeDomain without owner", func() {
			computeDomain, err := BuildComputeDomain(ctx, params, scheme, nil)

			Expect(err).To(BeNil())
			Expect(computeDomain).NotTo(BeNil())
			Expect(computeDomain.OwnerReferences).To(BeEmpty())
		})

		It("should set controller reference when owner and scheme are provided", func() {
			computeDomain, err := BuildComputeDomain(ctx, params, scheme, owner)

			Expect(err).To(BeNil())
			Expect(computeDomain).NotTo(BeNil())
			Expect(computeDomain.OwnerReferences).To(HaveLen(1))
			Expect(computeDomain.OwnerReferences[0].Name).To(Equal(defaultName))
			Expect(computeDomain.OwnerReferences[0].Kind).To(Equal("Deployment"))
		})
	})

	Context("buildObjectMeta function", func() {
		It("should build object metadata correctly", func() {
			objectMeta, err := buildObjectMeta(ctx, params)

			Expect(err).To(BeNil())
			Expect(objectMeta).NotTo(BeNil())
			Expect(objectMeta.Name).To(Equal("dw-test-workload"))
			Expect(objectMeta.Namespace).To(Equal(defaultNamespace))
			Expect(objectMeta.Labels).NotTo(BeEmpty())
		})

		It("should handle different workload initials", func() {
			params.WorkloadInitials = "dinfw"
			params.WorkloadName = "inference-job"

			objectMeta, err := buildObjectMeta(ctx, params)

			Expect(err).To(BeNil())
			Expect(objectMeta.Name).To(Equal("dinfw-inference-job"))
		})
	})

	Context("buildSpec function", func() {
		It("should build spec correctly for multi-node workload with static compute domain", func() {
			spec, err := buildSpec(params)

			Expect(spec).NotTo(BeNil())
			Expect(err).To(BeNil())
			Expect(spec.NumNodes).To(Equal(2))
			Expect(spec.Channel).NotTo(BeNil())
			Expect(spec.Channel.ResourceClaimTemplate).NotTo(BeNil())
		})

		It("should build spec correctly for multi-node workload with elastic compute domain", func() {
			params.ElasticEnabled = true
			spec, err := buildSpec(params)

			Expect(spec).NotTo(BeNil())
			Expect(err).To(BeNil())
			Expect(spec.NumNodes).To(Equal(0))
			Expect(spec.Channel).NotTo(BeNil())
			Expect(spec.Channel.ResourceClaimTemplate).NotTo(BeNil())
		})

		It("should return nil spec for CPU-only workload", func() {
			params.NumNodes = 0

			spec, err := buildSpec(params)

			Expect(spec).To(BeNil())
			Expect(err).ToNot(BeNil())
			Expect(err).To(Equal(ComputeDomainCpuOnlyErr))
		})

		It("should handle single node workload", func() {
			params.NumNodes = 1

			spec, err := buildSpec(params)

			Expect(spec).NotTo(BeNil())
			Expect(err).To(BeNil())
			Expect(spec.NumNodes).To(Equal(1))
		})

		It("should handle multiple node workload", func() {
			params.NumNodes = 10

			spec, err := buildSpec(params)

			Expect(spec).NotTo(BeNil())
			Expect(err).To(BeNil())
			Expect(spec.NumNodes).To(Equal(10))
		})
	})

	Context("PreviewComputeDomain function", func() {
		It("should create preview resource correctly and same as build", func() {
			previewComputeDomain := PreviewComputeDomain("dw", defaultName, defaultNamespace)

			Expect(previewComputeDomain).NotTo(BeNil())
			Expect(previewComputeDomain.Name).To(Equal("dw-test-workload"))
			Expect(previewComputeDomain.Namespace).To(Equal(defaultNamespace))

			// preview and build compute domain should be the same
			buildComputeDomain, err := BuildComputeDomain(ctx, params, scheme, owner)
			Expect(err).To(BeNil())
			Expect(buildComputeDomain).NotTo(BeNil())
			Expect(buildComputeDomain.Name).To(Equal(previewComputeDomain.Name))
			Expect(buildComputeDomain.Namespace).To(Equal(previewComputeDomain.Namespace))
		})

		It("should handle different workload names", func() {
			resource := PreviewComputeDomain("dinfw", "inference-model", "ml-namespace")

			Expect(resource.Name).To(Equal("dinfw-inference-model"))
			Expect(resource.Namespace).To(Equal("ml-namespace"))
		})
	})

	Context("ComposeComputeDomainName function", func() {
		It("should compose name correctly", func() {
			name := ComposeComputeDomainName("tr", "test-workload")

			Expect(name).To(Equal("tr-test-workload"))
		})

		It("should handle empty workload initials", func() {
			name := ComposeComputeDomainName("", "test-workload")

			Expect(name).To(Equal("-test-workload"))
		})

		It("should handle empty workload name", func() {
			name := ComposeComputeDomainName("tr", "")

			Expect(name).To(Equal("tr-"))
		})

		It("should handle special characters in names", func() {
			name := ComposeComputeDomainName("tr", "test-workload-123")

			Expect(name).To(Equal("tr-test-workload-123"))
		})
	})

	Context("Different workload types", func() {
		It("should handle training workloads", func() {
			params.WorkloadKind = "Job"
			params.JobType = "Distributed"
			params.WorkloadInitials = "dw"

			computeDomain, err := BuildComputeDomain(ctx, params, scheme, owner)

			Expect(err).To(BeNil())
			Expect(computeDomain).NotTo(BeNil())
		})

		It("should handle inference workloads", func() {
			params.WorkloadKind = "Deployment"
			params.JobType = "DistributedInference"
			params.WorkloadInitials = "dinfw"

			computeDomain, err := BuildComputeDomain(ctx, params, scheme, owner)

			Expect(err).To(BeNil())
			Expect(computeDomain).NotTo(BeNil())
		})
	})
})
