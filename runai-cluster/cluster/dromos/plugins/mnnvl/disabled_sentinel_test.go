/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvl

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/pkg/plugin/framework"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("MNNVL disabled sentinel", func() {
	var fw *framework.Framework

	BeforeEach(func() {
		GinkgoT().Setenv("ENABLE_MNNVL_SUPPORT", "false")

		build := NewBuilder()
		p, err := build(context.Background(), framework.PluginDeps{})
		Expect(err).NotTo(HaveOccurred())
		Expect(p.Name()).To(Equal(PluginName))

		fw, err = framework.NewFrameworkFromPluginBuilders(context.Background(), &framework.Config{},
			[]framework.PluginBuilder{build}, framework.PluginDeps{})
		Expect(err).NotTo(HaveOccurred())
	})

	It("rejects a create with an explicit mnnvl ask while disabled, with a Forbidden error on the mnnvl path", func() {
		runner := newTestRunner("wl", "ns", newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
		_, errs := fw.RunValidators(context.Background(), nil, runner)
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Type).To(Equal(field.ErrorTypeForbidden))
		Expect(errs[0].Field).To(Equal(mnnvlInstructionPath().String()))
	})

	It("allows updating a running mnnvl workload when the instruction is unchanged", func() {
		existing := newTestRunner("wl", "ns", newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
		desired := newTestRunner("wl", "ns", newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
		_, errs := fw.RunValidators(context.Background(), existing, desired)
		Expect(errs).To(BeEmpty())
	})

	It("rejects changing the mnnvl instruction on an update while disabled", func() {
		existing := newTestRunner("wl", "ns", newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel))
		desired := newTestRunner("wl", "ns", newMnnvlConfig(dromosv1alpha1.MnnvlModeRequired, testGPULabel, "extra-label"))
		_, errs := fw.RunValidators(context.Background(), existing, desired)
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Type).To(Equal(field.ErrorTypeForbidden))
	})

	It("does not reject None or a runner with no mnnvl instruction", func() {
		none := newTestRunner("wl", "ns", newMnnvlConfig(dromosv1alpha1.MnnvlModeNone, testGPULabel))
		_, errs := fw.RunValidators(context.Background(), nil, none)
		Expect(errs).To(BeEmpty())

		bare := newTestRunner("wl", "ns", nil)
		_, errs = fw.RunValidators(context.Background(), nil, bare)
		Expect(errs).To(BeEmpty())
	})
})
