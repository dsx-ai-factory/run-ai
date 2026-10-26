/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package toleration

import (
	"testing"

	"k8s.io/utils/ptr"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	full         = "name=NAME,operator=Exists,key=KEY,value=VALUE,effect=PreferNoSchedule,seconds=200"
	minimalValid = "operator=Exists,key=KEY,effect=PreferNoSchedule"
	valid        = "name=NAME,operator=Equal,key=KEY,effect=NoExecute,seconds=200,value=VALUE"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "New Toleration Validation Tests")
}

var _ = Describe("Parser Tests", func() {
	Context("Sanity validation", func() {
		It("everything", func() {
			ap := NewParser(full)
			err := ap.Parse()
			Expect(err).To(BeNil())
			Expect(*ap.Get(Name)).To(Equal("NAME"))
			Expect(*ap.Get(Operator)).To(Equal("Exists"))
			Expect(*ap.Get(Key)).To(Equal("KEY"))
			Expect(*ap.Get(Value)).To(Equal("VALUE"))
			Expect(*ap.Get(Effect)).To(Equal("PreferNoSchedule"))
			Expect(*ap.Get(Seconds)).To(Equal("200"))
		})
		It("minimal", func() {
			ap := NewParser(minimalValid)
			err := ap.Parse()
			Expect(err).To(BeNil())
			Expect(*ap.Get(Operator)).To(Equal("Exists"))
			Expect(*ap.Get(Key)).To(Equal("KEY"))
			Expect(*ap.Get(Effect)).To(Equal("PreferNoSchedule"))
		})
	})

	Context("create", func() {
		It("minimal", func() {
			obj, err := New(minimalValid)
			Expect(err).To(BeNil())
			Expect(obj.GetKey()).To(Equal("KEY"))
			Expect(obj.GetOperator()).To(Equal(openapi.TOLERATIONOPERATOR_EXISTS))
			Expect(obj.HasName()).To(Equal(false))
			Expect(obj.HasValue()).To(Equal(false))
			Expect(obj.GetEffect()).To(Equal(openapi.TOLERATIONEFFECT_PREFER_NO_SCHEDULE))
			Expect(obj.HasSeconds()).To(Equal(false))
		})

		It("valid", func() {
			obj, err := New(valid)
			Expect(err).To(BeNil())
			Expect(obj.GetName()).To(Equal("NAME"))
			Expect(obj.GetOperator()).To(Equal(openapi.TOLERATIONOPERATOR_EQUAL))
			Expect(obj.GetKey()).To(Equal("KEY"))
			Expect(obj.GetValue()).To(Equal("VALUE"))
			Expect(obj.GetEffect()).To(Equal(openapi.TOLERATIONEFFECT_NO_EXECUTE))
			Expect(obj.GetSeconds()).To(Equal(int32(200)))
		})
	})

	Context("validate", func() {
		var (
			t *openapi.Toleration
		)

		BeforeEach(func() {
			t = &openapi.Toleration{
				Operator: *openapi.NewNullableTolerationOperator(openapi.TOLERATIONOPERATOR_EXISTS.Ptr()),
				Key:      *openapi.NewNullableString(ptr.To("KEY")),
				Effect:   *openapi.NewNullableTolerationEffect(openapi.TOLERATIONEFFECT_NO_EXECUTE.Ptr()),
			}
		})

		It("valid", func() {
			err := validate(t)
			Expect(err).NotTo(HaveOccurred())
		})

		It("missing Operator", func() {
			t.Operator.Unset()

			err := validate(t)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("operator must be provided"))
		})

		It("missing key", func() {
			t.Key.Unset()

			err := validate(t)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("key must be provided"))
		})

		It("missing effect", func() {
			t.Effect.Unset()

			err := validate(t)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("effect must be provided"))
		})

		It("Equal without value", func() {
			t.Operator.Set(openapi.TOLERATIONOPERATOR_EQUAL.Ptr())

			err := validate(t)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("value must be provided"))
		})

		It("Exists without value", func() {
			t.Value.Set(ptr.To("VALUE"))

			err := validate(t)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("value is not allowed"))
		})

		It("NoExecute with seconds", func() {
			t.Effect.Set(openapi.TOLERATIONEFFECT_PREFER_NO_SCHEDULE.Ptr())
			t.Seconds.Set(ptr.To(int32(100)))

			err := validate(t)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("seconds is not allowed"))
		})
	})
})
