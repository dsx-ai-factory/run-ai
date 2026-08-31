/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package policy

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer"
	factory2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/shared/enforcer/factory"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/internal/autogenerate/mocks"
	inferencepolicy "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferencepolicy/v2alpha1"
	infwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	client2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/client"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/register"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("WorkloadPolicyApplier", func() {

	Context("ApplyPolicy", func() {
		var (
			workloadMergeCreator  mocks.JobMergeCreator
			baseWorkload          *iwv2.InteractiveWorkload
			workloadPolicyApplier *WorkloadPolicyApplier
			ctx                   context.Context
			subresourcePolicy     mocks.SubresourceApplier
		)
		BeforeEach(func() {
			workloadMergeCreator = mocks.JobMergeCreator{}
			subresourcePolicy = mocks.SubresourceApplier{}
			ctx = context.Background()
			job := workload.NewJobWorkloadBuilder().WithName("test-workload").WithGpuCount(1).SetActive().WithInteractive().WithCommand("sleep 1000").Build()
			baseWorkload = iwv2.NewFromJobWorkload(job)
			workloadPolicyApplier = NewPolicyApplier(&workloadMergeCreator, &subresourcePolicy)
		})
		It("should return error if failed create Merge stack", func() {
			workloadMergeCreator.On("CreateMergeStack", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("failed to create Merge stack"))

			err := workloadPolicyApplier.ApplyPolicy(ctx, baseWorkload)
			Expect(err).To(HaveOccurred())

		})
		It("should apply policy and set CPU default", func() {
			const cpuDefault = 10
			templateWorkload := (baseWorkload.DeepCopyObject()).(*iwv2.InteractiveWorkload)
			templateWorkload.Spec.Cpu = field_types.NewOptionalNumberField(ptr.To[float64](cpuDefault))
			specJson, _ := json.Marshal(templateWorkload.Spec)

			workloadMergeCreator.On("CreateMergeStack", mock.Anything, mock.Anything).
				Return(map[api.SubresourceComponent]*enforcer.MergeStack{api.SpecSubResource: nil}, nil)
			subresourcePolicy.On("Apply", mock.Anything, mock.Anything).Return(specJson, nil)
			err := workloadPolicyApplier.ApplyPolicy(ctx, baseWorkload)
			Expect(err).NotTo(HaveOccurred())
			Expect(baseWorkload.Spec.Cpu.String()).To(Equal(strconv.Itoa(cpuDefault)))

		})
	})

	Context("SimpleApplyPolicy", func() {
		scheme := runtime.NewScheme()
		register.InitSchema(scheme)

		var (
			baseWorkload          *infwv2.InferenceWorkload
			workloadPolicyApplier *WorkloadPolicyApplier
			ctx                   context.Context
			k8sClient             client.Client
		)

		BeforeEach(func() {
			ctx = context.Background()

			k8sClient = client2.InitFakeClientWithObjectNoStatus(scheme)
			inferenceMergeStackCreator := factory2.NewInferenceWorkloadMergeCreator(k8sClient)
			workloadPolicyApplier = NewSimplePolicyApplier(inferenceMergeStackCreator)
		})

		It("should apply inference policy and set default", func() {

			// build inference policy with inference default values
			inferenceP := inferencepolicy.InferencePolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-policy",
					Namespace: "test-namespace",
				},

				Spec: inferencepolicy.InferenceSpec{
					InferenceFields: field_groups.InferenceFields{
						MpsFields: field_groups.MpsFields{
							Mps: field_types.NewOptionalBoolField(ptr.To[bool](true)),
						},
						AutoScalingFields: field_groups.AutoScalingFields{
							MinScale: field_types.NewOptionalIntField(ptr.To[int](1)),
						},
						RoutingAndAuthorizationFields: field_groups.RoutingAndAuthorizationFields{
							IsPrivateServiceUrl: field_types.NewOptionalBoolField(ptr.To[bool](true)),
						},
					},
				},
			}

			// create inference policy (old policy)
			err := k8sClient.Create(ctx, &inferenceP)
			Expect(err).NotTo(HaveOccurred())

			// create inference workload
			job := workload.NewJobWorkloadBuilder().WithName("test-workload").WithNamespace("test-namespace").SetActive().WithInference().Build()
			baseWorkload = infwv2.NewFromJobWorkload(job)

			// apply policy
			err = workloadPolicyApplier.ApplyPolicy(ctx, baseWorkload)

			// check if policy applied
			Expect(err).NotTo(HaveOccurred())
			Expect(baseWorkload.Spec.InferenceFields.MpsFields.Mps.Value).To(Equal(ptr.To(true)))
			Expect(baseWorkload.Spec.InferenceFields.AutoScalingFields.MinScale.Value).To(Equal(ptr.To(1)))
			Expect(baseWorkload.Spec.InferenceFields.RoutingAndAuthorizationFields.IsPrivateServiceUrl.Value).To(Equal(ptr.To(true)))
		})
	})
})
