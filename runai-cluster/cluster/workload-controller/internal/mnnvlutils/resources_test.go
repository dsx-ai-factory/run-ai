/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvlutils

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/jobs"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("MNNVL Resource Utils", func() {
	Context("ValidateResources", func() {
		var (
			jw *jobs.JobWorkload
		)

		BeforeEach(func() {
			jw = &jobs.JobWorkload{}
		})
		Context("nvidia gpu cluster", func() {
			Context("Use nodepool", func() {
				It("only worker uses GPU", func() {
					jw.Spec = jobs.JobSpec{}
					jw.Spec.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))
					jw.Spec.NodePool = field_types.NewOptionalStringField("default-pool", "")
					Expect(ValidateResources(jw)).To(Succeed())
				})
				It("only master uses GPU", func() {
					jw.Spec = jobs.JobSpec{}
					jw.Spec.NodePools = field_types.NewOptionalStringField("default-pool", "")

					jw.MasterSpec = &jobs.JobSpec{}
					jw.MasterSpec.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))
					jw.MasterSpec.NodePools = field_types.NewOptionalStringField("default-pool", "")

					Expect(ValidateResources(jw)).To(Succeed())
				})
				It("both master and worker uses GPU", func() {
					jw.Spec = jobs.JobSpec{}
					jw.Spec.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))
					jw.Spec.NodePools = field_types.NewOptionalStringField("default-pool", "")

					jw.MasterSpec = &jobs.JobSpec{}
					jw.MasterSpec.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))
					jw.MasterSpec.NodePools = field_types.NewOptionalStringField("default-pool", "")

					Expect(ValidateResources(jw)).To(Succeed())
				})
				It("none of the master and worker uses GPU", func() {
					jw.Spec = jobs.JobSpec{}
					jw.Spec.NodePool = field_types.NewOptionalStringField("default-pool", "")

					jw.MasterSpec = &jobs.JobSpec{}
					jw.MasterSpec.NodePool = field_types.NewOptionalStringField("default-pool", "")

					Expect(ValidateResources(jw)).Error()
				})
			})
			Context("without nodepools or nodepool should failed", func() {
				It("only worker uses GPU", func() {
					jw.Spec = jobs.JobSpec{}
					jw.Spec.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))
					Expect(ValidateResources(jw)).Error()
				})
				It("only master uses GPU", func() {
					jw.Spec = jobs.JobSpec{}
					jw.MasterSpec = &jobs.JobSpec{}
					jw.MasterSpec.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))
					Expect(ValidateResources(jw)).Error()
				})
				It("both master and worker GPU", func() {
					jw.Spec = jobs.JobSpec{}
					jw.Spec.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))
					jw.MasterSpec = &jobs.JobSpec{}
					jw.MasterSpec.GpuDevices = field_types.NewOptionalIntField(ptr.To(1))
					Expect(ValidateResources(jw)).Error()
				})
			})
		})
	})
})
