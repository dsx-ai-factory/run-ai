/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

var _ = Describe("SpecBuilder", func() {
	It("should build a Spec correctly", func() {
		labels := []openapi.Label{
			{Name: *openapi.NewNullableString(ptr.To("label1")), Value: *openapi.NewNullableString(ptr.To("value1"))},
		}

		masterLabels := []openapi.Label{
			{Name: *openapi.NewNullableString(ptr.To("master-label1")), Value: *openapi.NewNullableString(ptr.To("master-label1"))},
		}

		spec := displayer.NewSpecBuilder().
			Labels(lo.ToSlicePtr(labels)).
			MasterLabels(lo.ToSlicePtr(masterLabels)).
			RunaiCliCommand("runai submit --image=ubuntu").
			Build()

		Expect(spec.Labels).To(HaveLen(1))
		Expect(*spec.Labels[0].Name.Get()).To(Equal("label1"))
		Expect(spec.MasterLabels).To(HaveLen(1))
		Expect(*spec.MasterLabels[0].Name.Get()).To(Equal("master-label1"))
		Expect(spec.RunaiCliCommand).To(Equal("runai submit --image=ubuntu"))
	})

	It("should set entire Spec with Set method", func() {
		original := displayer.NewSpecBuilder().
			RunaiCliCommand("original command").
			Build()

		spec := displayer.NewSpecBuilder().
			Set(original).
			Build()

		Expect(spec.RunaiCliCommand).To(Equal("original command"))
	})
})
