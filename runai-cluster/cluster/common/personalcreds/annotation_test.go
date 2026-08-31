/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package personalcreds

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPersonalcreds(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Personalcreds Annotation Suite")
}

func newObj(annotations map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Annotations: annotations},
	}
}

var _ = Describe("Parse", func() {
	DescribeTable("rejects bad input",
		func(raw string) {
			_, err := Parse(raw)
			Expect(err).To(HaveOccurred())
		},
		Entry("empty string", ""),
		Entry("malformed yaml", "- sourceSecret: foo\n  - broken"),
		Entry("not a list", "sourceSecret: foo"),
		Entry("empty list", "[]"),
		Entry("entry with empty sourceSecret", "- sourceSecret: \"\"\n- sourceSecret: ok"),
		Entry("entry missing sourceSecret", "- otherField: foo"),
	)

	It("parses a single entry", func() {
		entries, err := Parse("- sourceSecret: my-hf-token-abc1234567\n")
		Expect(err).ToNot(HaveOccurred())
		Expect(entries).To(Equal([]Cred{{SourceSecret: "my-hf-token-abc1234567"}}))
	})

	It("parses multiple entries preserving order", func() {
		raw := "- sourceSecret: a\n- sourceSecret: b\n- sourceSecret: c\n"
		entries, err := Parse(raw)
		Expect(err).ToNot(HaveOccurred())
		Expect(entries).To(Equal([]Cred{{SourceSecret: "a"}, {SourceSecret: "b"}, {SourceSecret: "c"}}))
	})
})

var _ = Describe("Marshal", func() {
	It("sorts entries by SourceSecret", func() {
		out, err := Marshal([]Cred{{SourceSecret: "c"}, {SourceSecret: "a"}, {SourceSecret: "b"}})
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(Equal("- sourceSecret: a\n- sourceSecret: b\n- sourceSecret: c\n"))
	})

	It("rejects empty entries", func() {
		_, err := Marshal(nil)
		Expect(err).To(HaveOccurred())
	})

	It("rejects entry with empty SourceSecret", func() {
		_, err := Marshal([]Cred{{SourceSecret: ""}})
		Expect(err).To(HaveOccurred())
	})

	It("round-trips with Parse", func() {
		input := []Cred{{SourceSecret: "x"}, {SourceSecret: "y"}}
		raw, err := Marshal(input)
		Expect(err).ToNot(HaveOccurred())
		out, err := Parse(raw)
		Expect(err).ToNot(HaveOccurred())
		Expect(out).To(Equal(input))
	})
})

var _ = Describe("Has", func() {
	It("returns false when annotation is absent", func() {
		Expect(Has(newObj(nil))).To(BeFalse())
		Expect(Has(newObj(map[string]string{"other": "x"}))).To(BeFalse())
	})

	It("returns true when annotation is present", func() {
		Expect(Has(newObj(map[string]string{Annotation: "- sourceSecret: a"}))).To(BeTrue())
	})
})

var _ = Describe("Get", func() {
	It("returns nil when annotation is absent", func() {
		entries, err := Get(newObj(nil))
		Expect(err).ToNot(HaveOccurred())
		Expect(entries).To(BeNil())
	})

	It("returns parsed entries when present", func() {
		obj := newObj(map[string]string{Annotation: "- sourceSecret: a\n- sourceSecret: b\n"})
		entries, err := Get(obj)
		Expect(err).ToNot(HaveOccurred())
		Expect(entries).To(Equal([]Cred{{SourceSecret: "a"}, {SourceSecret: "b"}}))
	})

	It("returns an error for malformed annotation", func() {
		obj := newObj(map[string]string{Annotation: "not: valid: list"})
		_, err := Get(obj)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("Set", func() {
	It("writes a sorted annotation value", func() {
		obj := newObj(nil)
		Expect(Set(obj, []Cred{{SourceSecret: "z"}, {SourceSecret: "a"}})).To(Succeed())
		Expect(obj.GetAnnotations()[Annotation]).To(Equal("- sourceSecret: a\n- sourceSecret: z\n"))
	})

	It("removes the annotation when entries is empty", func() {
		obj := newObj(map[string]string{Annotation: "- sourceSecret: a\n", "other": "keep"})
		Expect(Set(obj, nil)).To(Succeed())
		Expect(obj.GetAnnotations()).ToNot(HaveKey(Annotation))
		Expect(obj.GetAnnotations()).To(HaveKeyWithValue("other", "keep"))
	})

	It("rejects entries with empty SourceSecret", func() {
		Expect(Set(newObj(nil), []Cred{{SourceSecret: ""}})).To(MatchError(ContainSubstring("sourceSecret is required")))
	})
})

var _ = Describe("Add", func() {
	It("creates the annotation when absent", func() {
		obj := newObj(nil)
		Expect(Add(obj, Cred{SourceSecret: "a"})).To(Succeed())
		Expect(Has(obj)).To(BeTrue())
	})

	It("appends to existing entries", func() {
		obj := newObj(map[string]string{Annotation: "- sourceSecret: a\n"})
		Expect(Add(obj, Cred{SourceSecret: "b"})).To(Succeed())
		entries, _ := Get(obj)
		Expect(entries).To(ConsistOf(Cred{SourceSecret: "a"}, Cred{SourceSecret: "b"}))
	})

	It("is idempotent for duplicate SourceSecret", func() {
		obj := newObj(map[string]string{Annotation: "- sourceSecret: a\n"})
		Expect(Add(obj, Cred{SourceSecret: "a"})).To(Succeed())
		entries, _ := Get(obj)
		Expect(entries).To(Equal([]Cred{{SourceSecret: "a"}}))
	})

	It("rejects empty SourceSecret", func() {
		Expect(Add(newObj(nil), Cred{SourceSecret: ""})).To(HaveOccurred())
	})
})

var _ = Describe("Remove", func() {
	It("returns false when annotation is absent", func() {
		removed, err := Remove(newObj(nil), "a")
		Expect(err).ToNot(HaveOccurred())
		Expect(removed).To(BeFalse())
	})

	It("returns false when source is not present", func() {
		obj := newObj(map[string]string{Annotation: "- sourceSecret: a\n"})
		removed, err := Remove(obj, "missing")
		Expect(err).ToNot(HaveOccurred())
		Expect(removed).To(BeFalse())
	})

	It("removes a matching entry", func() {
		obj := newObj(map[string]string{Annotation: "- sourceSecret: a\n- sourceSecret: b\n"})
		removed, err := Remove(obj, "a")
		Expect(err).ToNot(HaveOccurred())
		Expect(removed).To(BeTrue())
		entries, _ := Get(obj)
		Expect(entries).To(Equal([]Cred{{SourceSecret: "b"}}))
	})

	It("removes the annotation entirely when last entry is dropped", func() {
		obj := newObj(map[string]string{Annotation: "- sourceSecret: a\n"})
		_, err := Remove(obj, "a")
		Expect(err).ToNot(HaveOccurred())
		Expect(Has(obj)).To(BeFalse())
	})
})
