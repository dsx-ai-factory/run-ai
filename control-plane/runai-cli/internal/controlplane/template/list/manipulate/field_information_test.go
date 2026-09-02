/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	wcgenerated "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

var _ = Describe("Field Extract Functions", func() {
	var (
		template    *wcgenerated.SupersetTemplate
		templateRow *TemplateAlignedRow
	)

	BeforeEach(func() {
		template = &wcgenerated.SupersetTemplate{}
		templateRow = &TemplateAlignedRow{}
	})

	Context("extractImageData", func() {
		It("should extract image when spec and image exist", func() {
			image := "pytorch:latest"
			template.Spec = &wcgenerated.SupersetSpec{}
			template.Spec.SetImage(image)
			extractImageData(template, templateRow)
			Expect(templateRow.Image).To(Equal("pytorch:latest"))
		})

		It("should set empty string when spec is nil", func() {
			template.Spec = nil
			extractImageData(template, templateRow)
			Expect(templateRow.Image).To(Equal(""))
		})

		It("should set empty string when image is not set", func() {
			template.Spec = &wcgenerated.SupersetSpec{}
			extractImageData(template, templateRow)
			Expect(templateRow.Image).To(Equal(""))
		})

		It("should set empty string when image is empty", func() {
			template.Spec = &wcgenerated.SupersetSpec{}
			template.Spec.SetImage("")
			extractImageData(template, templateRow)
			Expect(templateRow.Image).To(Equal(""))
		})
	})

	Context("extractGpuComputeRequestData", func() {
		It("should extract GPU request when all fields exist", func() {
			gpuRequest := int32(4)
			template.Spec = &wcgenerated.SupersetSpec{}
			compute := wcgenerated.WorkspaceSpecSpecAllOfCompute{}
			compute.SetGpuDevicesRequest(gpuRequest)
			template.Spec.SetCompute(compute)
			extractGpuComputeRequestData(template, templateRow)
			Expect(templateRow.GpuComputeRequest).To(Equal("4"))
		})

		It("should set empty string when spec is nil", func() {
			template.Spec = nil
			extractGpuComputeRequestData(template, templateRow)
			Expect(templateRow.GpuComputeRequest).To(Equal(""))
		})

		It("should set empty string when compute is not set", func() {
			template.Spec = &wcgenerated.SupersetSpec{}
			extractGpuComputeRequestData(template, templateRow)
			Expect(templateRow.GpuComputeRequest).To(Equal(""))
		})

		It("should set empty string when GPU request is not set", func() {
			template.Spec = &wcgenerated.SupersetSpec{}
			compute := wcgenerated.WorkspaceSpecSpecAllOfCompute{}
			template.Spec.SetCompute(compute)
			extractGpuComputeRequestData(template, templateRow)
			Expect(templateRow.GpuComputeRequest).To(Equal(""))
		})

		It("should handle GPU request of 0", func() {
			gpuRequest := int32(0)
			template.Spec = &wcgenerated.SupersetSpec{}
			compute := wcgenerated.WorkspaceSpecSpecAllOfCompute{}
			compute.SetGpuDevicesRequest(gpuRequest)
			template.Spec.SetCompute(compute)
			extractGpuComputeRequestData(template, templateRow)
			Expect(templateRow.GpuComputeRequest).To(Equal("0"))
		})
	})

	Context("extractMandatoryFieldsMissingData", func() {
		It("should set nil when Runnable field is not set", func() {
			// template.Runnable is nil by default
			extractMandatoryFieldsMissingData(template, templateRow)
			Expect(templateRow.MandatoryFieldsMissing).To(BeNil())
		})

		It("should set true when IsRunnable is false", func() {
			runnable := wcgenerated.RunnableInfo{}
			runnable.SetIsRunnable(false)
			template.SetRunnable(runnable)
			extractMandatoryFieldsMissingData(template, templateRow)
			Expect(templateRow.MandatoryFieldsMissing).ToNot(BeNil())
			Expect(*templateRow.MandatoryFieldsMissing).To(BeTrue())
		})

		It("should set nil when IsRunnable is not set", func() {
			runnable := wcgenerated.RunnableInfo{}
			// IsRunnable is not set
			template.SetRunnable(runnable)
			extractMandatoryFieldsMissingData(template, templateRow)
			Expect(templateRow.MandatoryFieldsMissing).To(BeNil())
		})

		It("should set false when IsRunnable is true", func() {
			runnable := wcgenerated.RunnableInfo{}
			isRunnable := true
			runnable.SetIsRunnable(isRunnable)
			template.SetRunnable(runnable)
			extractMandatoryFieldsMissingData(template, templateRow)
			Expect(templateRow.MandatoryFieldsMissing).ToNot(BeNil())
			Expect(*templateRow.MandatoryFieldsMissing).To(BeFalse())
		})

		It("should set true when issue is set but IsRunnable is not set", func() {
			runnable := wcgenerated.RunnableInfo{}
			runnable.SetIssue("image is not provided")
			template.SetRunnable(runnable)
			extractMandatoryFieldsMissingData(template, templateRow)
			Expect(templateRow.MandatoryFieldsMissing).ToNot(BeNil())
			Expect(*templateRow.MandatoryFieldsMissing).To(BeTrue())
		})
	})
})

var _ = Describe("Field Format Functions", func() {
	var templateRow *TemplateAlignedRow

	BeforeEach(func() {
		templateRow = &TemplateAlignedRow{}
	})

	Context("formatMandatoryFieldsMissingValue", func() {
		It("should return 'Yes' when mandatory fields are missing", func() {
			mandatoryFieldsMissing := true
			templateRow.MandatoryFieldsMissing = &mandatoryFieldsMissing
			result := formatMandatoryFieldsMissingValue(templateRow)
			Expect(result).To(Equal("Yes"))
		})

		It("should return 'No' when mandatory fields are not missing", func() {
			mandatoryFieldsMissing := false
			templateRow.MandatoryFieldsMissing = &mandatoryFieldsMissing
			result := formatMandatoryFieldsMissingValue(templateRow)
			Expect(result).To(Equal("No"))
		})

		It("should return '-' when not set (nil)", func() {
			templateRow.MandatoryFieldsMissing = nil
			result := formatMandatoryFieldsMissingValue(templateRow)
			Expect(result).To(Equal("-"))
		})
	})
})
