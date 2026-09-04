/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package framework

import (
	"context"

	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("DisabledValidator", func() {
	path := field.NewPath("spec", "instructions", "example")
	instructionOf := func(r *dromosv1alpha1.WorkloadRunner) any {
		if r.Spec.ManagedBy == "" {
			return nil
		}
		return r.Spec.ManagedBy
	}
	withA := &dromosv1alpha1.WorkloadRunner{Spec: dromosv1alpha1.WorkloadRunnerSpec{ManagedBy: "a"}}
	withB := &dromosv1alpha1.WorkloadRunner{Spec: dromosv1alpha1.WorkloadRunnerSpec{ManagedBy: "b"}}
	without := &dromosv1alpha1.WorkloadRunner{}

	validator := func() ValidatorPlugin {
		return DisabledValidator("example", instructionOf, path).(ValidatorPlugin)
	}

	It("is a validator-only plugin", func() {
		v := DisabledValidator("example", instructionOf, path)
		Expect(v.Name()).To(Equal("example"))

		_, isMutator := v.(ResourceMutatePlugin)
		Expect(isMutator).To(BeFalse(), "the sentinel must not implement ResourceMutatePlugin")
		_, isValidator := v.(ValidatorPlugin)
		Expect(isValidator).To(BeTrue())
	})

	It("rejects a create that carries the instruction, with a Forbidden error on the instruction path", func() {
		_, errs := validator().WebhookValidate(context.Background(), nil, withA)
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Type).To(Equal(field.ErrorTypeForbidden))
		Expect(errs[0].Field).To(Equal(path.String()))
	})

	It("rejects an update that adds the instruction", func() {
		_, errs := validator().WebhookValidate(context.Background(), without, withA)
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Type).To(Equal(field.ErrorTypeForbidden))
	})

	It("rejects an update that changes the instruction", func() {
		_, errs := validator().WebhookValidate(context.Background(), withA, withB)
		Expect(errs).To(HaveLen(1))
		Expect(errs[0].Type).To(Equal(field.ErrorTypeForbidden))
	})

	It("allows an update that leaves the instruction unchanged", func() {
		_, errs := validator().WebhookValidate(context.Background(), withA, withA)
		Expect(errs).To(BeEmpty())
	})

	It("allows an update that removes the instruction", func() {
		_, errs := validator().WebhookValidate(context.Background(), withA, without)
		Expect(errs).To(BeEmpty())
	})

	It("allows a create without the instruction", func() {
		_, errs := validator().WebhookValidate(context.Background(), nil, without)
		Expect(errs).To(BeEmpty())
	})
})
