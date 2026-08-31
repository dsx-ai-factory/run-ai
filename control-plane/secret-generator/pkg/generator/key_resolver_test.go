/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package generator_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/secret-generator/pkg/generator"
)

var _ = Describe("ResolveKeyData", func() {
	It("resolves chart value over copyFrom", func() {
		chartVal := "my-chart-password"
		keys := []generator.SecretKeySpec{
			{Name: "AUTH_TOKEN", Type: "alphanumeric", Length: 32},
			{Name: "pubsub-password", CopyFrom: "AUTH_TOKEN", ChartValue: &chartVal},
		}
		data, err := generator.ResolveKeyData(keys, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(data).To(HaveKey("AUTH_TOKEN"))
		Expect(data).To(HaveKey("pubsub-password"))
		Expect(string(data["pubsub-password"])).To(Equal("my-chart-password"))
	})

	It("resolves copyFrom when source is in initialData", func() {
		keys := []generator.SecretKeySpec{
			{Name: "alias", CopyFrom: "AUTH_TOKEN"},
		}
		initial := map[string][]byte{"AUTH_TOKEN": []byte("token123")}
		data, err := generator.ResolveKeyData(keys, initial)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data["alias"])).To(Equal("token123"))
	})

	It("resolves static then generate for independent keys", func() {
		keys := []generator.SecretKeySpec{
			{Name: "user", Type: "static", ChartValue: stringPtr("admin")},
			{Name: "password", Type: "password", Length: 16},
		}
		data, err := generator.ResolveKeyData(keys, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data["user"])).To(Equal("admin"))
		Expect(data["password"]).To(HaveLen(16))
	})

	It("rejects self-referencing CopyFrom", func() {
		keys := []generator.SecretKeySpec{
			{Name: "x", CopyFrom: "x"},
		}
		_, err := generator.ResolveKeyData(keys, nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no resolver in chain could resolve key x"))
	})

	It("rejects unresolved CopyFrom (missing source)", func() {
		keys := []generator.SecretKeySpec{
			{Name: "alias", CopyFrom: "MISSING"},
		}
		_, err := generator.ResolveKeyData(keys, nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no resolver in chain could resolve key alias"))
	})

	It("rejects externalOnly key when no explicit value is provided", func() {
		keys := []generator.SecretKeySpec{
			{Name: "password", Type: "password", Length: 24, ExternalOnly: true},
		}
		_, err := generator.ResolveKeyData(keys, nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("externalOnly"))
		Expect(err.Error()).To(ContainSubstring("password"))
	})

	It("resolves externalOnly key when chartValue is provided", func() {
		chartVal := "explicit-password"
		keys := []generator.SecretKeySpec{
			{Name: "password", Type: "password", Length: 24, ExternalOnly: true, ChartValue: &chartVal},
		}
		data, err := generator.ResolveKeyData(keys, nil)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data["password"])).To(Equal("explicit-password"))
	})

	It("resolves externalOnly key from legacy secret data", func() {
		keys := []generator.SecretKeySpec{
			{Name: "password", Type: "password", Length: 24, ExternalOnly: true},
		}
		legacy := map[string][]byte{"password": []byte("legacy-pw")}
		data, err := generator.ResolveKeyData(keys, legacy)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data["password"])).To(Equal("legacy-pw"))
	})
})
