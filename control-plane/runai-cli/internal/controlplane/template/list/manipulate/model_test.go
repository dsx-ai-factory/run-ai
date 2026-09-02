/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	wcgenerated "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

var _ = Describe("newBriefFromTemplate", func() {
	It("should convert template with all fields", func() {
		template := wcgenerated.SupersetTemplate{
			Meta: wcgenerated.TemplateMetaFields{
				Name:         "my-template",
				ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
				WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
				CreatedAt:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			},
		}
		spec := wcgenerated.SupersetSpec{}
		spec.SetImage("pytorch:2.0")
		compute := wcgenerated.WorkspaceSpecSpecAllOfCompute{}
		compute.SetGpuDevicesRequest(8)
		spec.SetCompute(compute)
		template.Spec = &spec

		runnable := wcgenerated.RunnableInfo{}
		runnable.SetIsRunnable(true)
		template.SetRunnable(runnable)

		result := newBriefFromTemplate(template)
		Expect(result.Name).To(Equal("my-template"))
		Expect(result.ScopeType).To(Equal("project"))
		Expect(result.WorkloadType).To(Equal("Training"))
		Expect(result.Image).To(Equal("pytorch:2.0"))
		Expect(result.GpuDevicesRequest).To(Equal("8"))
		Expect(result.MandatoryFieldsMissing).ToNot(BeNil())
		Expect(*result.MandatoryFieldsMissing).To(BeFalse())
		Expect(result.CreatedAt).ToNot(BeEmpty())
	})

	It("should handle template without spec", func() {
		template := wcgenerated.SupersetTemplate{
			Meta: wcgenerated.TemplateMetaFields{
				Name:         "minimal-template",
				ScopeType:    wcgenerated.SCOPETYPE_CLUSTER,
				WorkloadType: wcgenerated.WORKLOADTYPE_WORKSPACE,
				CreatedAt:    time.Now(),
			},
		}
		template.Spec = nil

		result := newBriefFromTemplate(template)
		Expect(result.Name).To(Equal("minimal-template"))
		Expect(result.ScopeType).To(Equal("cluster"))
		Expect(result.WorkloadType).To(Equal("Workspace"))
		Expect(result.Image).To(Equal(""))
		Expect(result.GpuDevicesRequest).To(Equal(""))
		Expect(result.MandatoryFieldsMissing).To(BeNil())
	})

	It("should handle spec without image", func() {
		template := wcgenerated.SupersetTemplate{
			Meta: wcgenerated.TemplateMetaFields{
				Name:         "no-image-template",
				ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
				WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
				CreatedAt:    time.Now(),
			},
		}
		spec := wcgenerated.SupersetSpec{}
		// Don't set image
		template.Spec = &spec

		result := newBriefFromTemplate(template)
		Expect(result.Name).To(Equal("no-image-template"))
		Expect(result.Image).To(Equal(""))
	})

	It("should handle spec without GPU request", func() {
		template := wcgenerated.SupersetTemplate{
			Meta: wcgenerated.TemplateMetaFields{
				Name:         "no-gpu-template",
				ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
				WorkloadType: wcgenerated.WORKLOADTYPE_WORKSPACE,
				CreatedAt:    time.Now(),
			},
		}
		spec := wcgenerated.SupersetSpec{}
		spec.SetImage("python:3.9")
		// Don't set compute
		template.Spec = &spec

		result := newBriefFromTemplate(template)
		Expect(result.Name).To(Equal("no-gpu-template"))
		Expect(result.Image).To(Equal("python:3.9"))
		Expect(result.GpuDevicesRequest).To(Equal(""))
	})

	It("should handle spec with compute but no GPU request", func() {
		template := wcgenerated.SupersetTemplate{
			Meta: wcgenerated.TemplateMetaFields{
				Name:         "compute-no-gpu",
				ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
				WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
				CreatedAt:    time.Now(),
			},
		}
		spec := wcgenerated.SupersetSpec{}
		spec.SetImage("tensorflow:2.0")
		compute := wcgenerated.WorkspaceSpecSpecAllOfCompute{}
		// Don't set GPU request
		spec.SetCompute(compute)
		template.Spec = &spec

		result := newBriefFromTemplate(template)
		Expect(result.Name).To(Equal("compute-no-gpu"))
		Expect(result.Image).To(Equal("tensorflow:2.0"))
		Expect(result.GpuDevicesRequest).To(Equal(""))
	})

	It("should handle isRunnable=false", func() {
		template := wcgenerated.SupersetTemplate{
			Meta: wcgenerated.TemplateMetaFields{
				Name:         "not-runnable",
				ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
				WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
				CreatedAt:    time.Now(),
			},
		}
		spec := wcgenerated.SupersetSpec{}
		spec.SetImage("pytorch:latest")
		template.Spec = &spec

		runnable := wcgenerated.RunnableInfo{}
		runnable.SetIsRunnable(false)
		template.SetRunnable(runnable)

		result := newBriefFromTemplate(template)
		Expect(result.MandatoryFieldsMissing).ToNot(BeNil())
		Expect(*result.MandatoryFieldsMissing).To(BeTrue())
	})

	It("should handle issue without isRunnable", func() {
		template := wcgenerated.SupersetTemplate{
			Meta: wcgenerated.TemplateMetaFields{
				Name:         "has-issue",
				ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
				WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
				CreatedAt:    time.Now(),
			},
		}
		spec := wcgenerated.SupersetSpec{}
		template.Spec = &spec

		runnable := wcgenerated.RunnableInfo{}
		runnable.SetIssue("image is not provided")
		template.SetRunnable(runnable)

		result := newBriefFromTemplate(template)
		Expect(result.MandatoryFieldsMissing).ToNot(BeNil())
		Expect(*result.MandatoryFieldsMissing).To(BeTrue())
	})

	It("should handle no runnable field", func() {
		template := wcgenerated.SupersetTemplate{
			Meta: wcgenerated.TemplateMetaFields{
				Name:         "no-runnable",
				ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
				WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
				CreatedAt:    time.Now(),
			},
		}
		spec := wcgenerated.SupersetSpec{}
		spec.SetImage("image:tag")
		template.Spec = &spec

		result := newBriefFromTemplate(template)

		Expect(result.MandatoryFieldsMissing).To(BeNil())
	})

	It("should format GPU request correctly", func() {
		template := wcgenerated.SupersetTemplate{
			Meta: wcgenerated.TemplateMetaFields{
				Name:         "gpu-format-test",
				ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
				WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
				CreatedAt:    time.Now(),
			},
		}
		spec := wcgenerated.SupersetSpec{}
		compute := wcgenerated.WorkspaceSpecSpecAllOfCompute{}
		compute.SetGpuDevicesRequest(0)
		spec.SetCompute(compute)
		template.Spec = &spec
		result := newBriefFromTemplate(template)
		Expect(result.GpuDevicesRequest).To(Equal("0"))
	})

	It("should handle different scope types", func() {
		testCases := []struct {
			scopeType wcgenerated.ScopeType
			expected  string
		}{
			{wcgenerated.SCOPETYPE_PROJECT, "project"},
			{wcgenerated.SCOPETYPE_CLUSTER, "cluster"},
			{wcgenerated.SCOPETYPE_DEPARTMENT, "department"},
		}

		for _, tc := range testCases {
			template := wcgenerated.SupersetTemplate{
				Meta: wcgenerated.TemplateMetaFields{
					Name:         "scope-test",
					ScopeType:    tc.scopeType,
					WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
					CreatedAt:    time.Now(),
				},
			}

			result := newBriefFromTemplate(template)
			Expect(result.ScopeType).To(Equal(tc.expected), "Failed for scope type: %s", tc.scopeType)
		}
	})

	It("should handle different workload types", func() {
		testCases := []struct {
			workloadType wcgenerated.WorkloadType
			expected     string
		}{
			{wcgenerated.WORKLOADTYPE_TRAINING, "Training"},
			{wcgenerated.WORKLOADTYPE_WORKSPACE, "Workspace"},
			{wcgenerated.WORKLOADTYPE_INFERENCE, "Inference"},
		}

		for _, tc := range testCases {
			template := wcgenerated.SupersetTemplate{
				Meta: wcgenerated.TemplateMetaFields{
					Name:         "workload-type-test",
					ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
					WorkloadType: tc.workloadType,
					CreatedAt:    time.Now(),
				},
			}

			result := newBriefFromTemplate(template)
			Expect(result.WorkloadType).To(Equal(tc.expected), "Failed for workload type: %s", tc.workloadType)
		}
	})
})

var _ = Describe("BriefListFromResponse", func() {
	It("should convert templates list response", func() {
		templates := []wcgenerated.SupersetTemplate{
			{
				Meta: wcgenerated.TemplateMetaFields{
					Name:         "template-1",
					ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
					WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
					CreatedAt:    time.Now(),
				},
			},
			{
				Meta: wcgenerated.TemplateMetaFields{
					Name:         "template-2",
					ScopeType:    wcgenerated.SCOPETYPE_CLUSTER,
					WorkloadType: wcgenerated.WORKLOADTYPE_WORKSPACE,
					CreatedAt:    time.Now(),
				},
			},
		}

		result := BriefListFromResponse(templates)
		Expect(result).To(HaveLen(2))
		Expect(result[0].Name).To(Equal("template-1"))
		Expect(result[0].ScopeType).To(Equal("project"))
		Expect(result[1].Name).To(Equal("template-2"))
		Expect(result[1].ScopeType).To(Equal("cluster"))
	})

	It("should handle empty templates list", func() {
		var templates []wcgenerated.SupersetTemplate
		result := BriefListFromResponse(templates)
		Expect(result).To(HaveLen(0))
	})
})
