/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_groups"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api"
	dwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	inferencewv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	twv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	workloadUtils "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/utils/testing_utils/resource/workload"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("patch", func() {
	Context("PatchWorkloadPolicy", func() {
		var (
			unionWorkload *jobs.JobWorkload
		)
		BeforeEach(func() {
			master := jobs.JobSpec{
				ResourceAllocationFields: field_groups.ResourceAllocationFields{
					Gpu: field_types.NewOptionalNumberField(ptr.To[float64](1)),
				},
			}
			unionWorkload = workloadUtils.NewJobWorkloadBuilder().WithName("before").WithMasterSpec(&master).Build()

		})
		Context("training workload", func() {

			It("should patch spec", func() {
				training := twv2.NewFromJobWorkload(unionWorkload)
				err := PatchWorkloadPolicy(training, map[api.SubresourceComponent][]byte{
					api.SpecSubResource: []byte(`{"name":{"value":"after"}}`),
				})
				Expect(err).To(BeNil())
				Expect(training.Spec.Name.String()).To(Equal("after"))
			})
		})
		Context("interactive workload", func() {

			It("should patch spec", func() {
				interactive := iwv2.NewFromJobWorkload(unionWorkload)
				err := PatchWorkloadPolicy(interactive, map[api.SubresourceComponent][]byte{
					api.SpecSubResource: []byte(`{"name":{"value":"after"}}`),
				})
				Expect(err).To(BeNil())
				Expect(interactive.Spec.Name.String()).To(Equal("after"))
			})
		})
		Context("distributed workload", func() {

			It("should patch only master spec", func() {
				distributed := dwv2.NewFromJobWorkload(unionWorkload)
				err := PatchWorkloadPolicy(distributed, map[api.SubresourceComponent][]byte{
					api.MasterSpecSubResource: []byte(`{"gpu":{"value":"4"}}`),
				})
				Expect(err).To(BeNil())
				Expect(distributed.Spec.Name.String()).To(Equal("before"))
				Expect(distributed.MasterSpec.Gpu.String()).To(Equal("4"))
			})
			It("should patch only worker spec", func() {
				distributed := dwv2.NewFromJobWorkload(unionWorkload)
				err := PatchWorkloadPolicy(distributed, map[api.SubresourceComponent][]byte{
					api.WorkerSpecSubResource: []byte(`{"name":{"value":"after"}}`),
				})
				Expect(err).To(BeNil())
				Expect(distributed.Spec.Name.String()).To(Equal("after"))
				Expect(distributed.MasterSpec.Gpu.String()).To(Equal("1"))
			})
			It("should patch both master and worker spec", func() {
				distributed := dwv2.NewFromJobWorkload(unionWorkload)
				err := PatchWorkloadPolicy(distributed, map[api.SubresourceComponent][]byte{
					api.MasterSpecSubResource: []byte(`{"gpu":{"value":"4"}}`),
					api.WorkerSpecSubResource: []byte(`{"name":{"value":"after"}}`),
				})
				Expect(err).To(BeNil())
				Expect(distributed.Spec.Name.String()).To(Equal("after"))
				Expect(distributed.MasterSpec.Gpu.String()).To(Equal("4"))
			})
		})
		Context("inference workload", func() {

			It("should patch common spec", func() {
				inference := inferencewv2.NewFromJobWorkload(unionWorkload)
				err := PatchWorkloadPolicy(inference, map[api.SubresourceComponent][]byte{
					api.SpecSubResource: []byte(`{"name":{"value":"after"}}`),
				})
				Expect(err).To(BeNil())
				Expect(inference.Spec.Name.String()).To(Equal("after"))
			})

			It("should patch inference spec", func() {
				inference := inferencewv2.NewFromJobWorkload(unionWorkload)
				Expect(inference.Spec.IsPrivateServiceUrl.String()).To(Equal("false"))

				err := PatchWorkloadPolicy(inference, map[api.SubresourceComponent][]byte{
					api.SpecSubResource: []byte(`{"IsPrivateServiceUrl":{"value":true}}`),
				})
				Expect(err).To(BeNil())
				Expect(inference.Spec.IsPrivateServiceUrl.String()).To(Equal("true"))
			})
		})

	})
	Context("patchDistributedWorkload", func() {
		var distributed *dwv2.DistributedWorkload
		BeforeEach(func() {
			distributed = &dwv2.DistributedWorkload{}
		})
		Context("master spec is nil", func() {
			It("should patch master spec", func() {

				err := patchDistributedWorkload(distributed, map[api.SubresourceComponent][]byte{
					api.MasterSpecSubResource: []byte(`{"gpu":{"value":"4"}}`),
				})
				Expect(err).To(BeNil())
				Expect(distributed.MasterSpec.Gpu.String()).To(Equal("4"))
			})
		})
		Context("worker spec is empty", func() {
			It("should patch worker spec", func() {

				err := patchDistributedWorkload(distributed, map[api.SubresourceComponent][]byte{
					api.WorkerSpecSubResource: []byte(`{"gpu":{"value":"4"}}`),
				})
				Expect(err).To(BeNil())
				Expect(distributed.Spec.Gpu.String()).To(Equal("4"))
			})
		})
	})

})
