/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package inference_endpoint_controller

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"
)

func TestInferenceEndpointController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Inference endpoint configuration")
}

var _ = Describe("Defaults", func() {
	DescribeTable("keeps defaults compatible with partial resource overrides", func(resources Resources) {
		config := &InferenceEndpointController{Resources: &resources}
		requirements := v1.ResourceRequirements(resources)
		original := requirements.DeepCopy()
		config.SetDefaultsWhereNeeded("registry.example/runai", nil)
		for key, request := range config.Resources.Requests {
			limit := config.Resources.Limits[key]
			Expect(request.Cmp(limit)).To(BeNumerically("<=", 0))
		}
		for key, value := range original.Requests {
			got := config.Resources.Requests[key]
			Expect(got.Cmp(value)).To(BeZero())
		}
		for key, value := range original.Limits {
			got := config.Resources.Limits[key]
			Expect(got.Cmp(value)).To(BeZero())
		}
	}, Entry("CPU request above the default limit", Resources{Requests: v1.ResourceList{v1.ResourceCPU: resource.MustParse("1")}}),
		Entry("memory limit below the default request", Resources{Limits: v1.ResourceList{v1.ResourceMemory: resource.MustParse("64Mi")}}))

	It("enables the controller and inherits global replicas and image registry", func() {
		config := &InferenceEndpointController{}
		config.SetDefaultsWhereNeeded("registry.example/runai", ptr.To[int32](3))
		Expect(*config.Enabled).To(BeTrue())
		Expect(*config.Replicas).To(Equal(int32(3)))
		Expect(*config.Image.Registry).To(Equal("registry.example/runai"))
		Expect(*config.Image.Name).To(Equal("inference-endpoint-controller"))
		Expect(config.Resources.Requests).NotTo(BeEmpty())
	})

	It("preserves explicit disablement and image and replica overrides", func() {
		config := &InferenceEndpointController{
			Enabled: ptr.To(false), Replicas: ptr.To[int32](2),
			Image: &common.Image{Registry: ptr.To("mirror.example"), Tag: ptr.To("custom")},
		}
		config.SetDefaultsWhereNeeded("registry.example/runai", ptr.To[int32](3))
		Expect(*config.Enabled).To(BeFalse())
		Expect(*config.Replicas).To(Equal(int32(2)))
		Expect(*config.Image.Registry).To(Equal("mirror.example"))
		Expect(*config.Image.Tag).To(Equal("custom"))
	})
})
