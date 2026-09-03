/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package generator_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/secret-generator/pkg/generator"
)

var _ = Describe("SecretSpec JSON serialization", func() {
	It("should correctly round-trip a SecretSpec with all fields through marshal and unmarshal", func() {
		spec := generator.SecretSpec{
			Name:               "test-secret",
			LegacySecretName:   "old-secret",
			ExistingSecretName: "",
			Keys: []generator.SecretKeySpec{
				{
					Name:          "password",
					LegacyKeyName: "PASSWORD",
					Length:        24,
					Type:          "alphanumeric",
				},
				{
					Name:       "username",
					Type:       "static",
					ChartValue: stringPtr("admin"),
				},
			},
		}

		data, err := json.Marshal(spec)
		Expect(err).NotTo(HaveOccurred())

		var unmarshaled generator.SecretSpec
		Expect(json.Unmarshal(data, &unmarshaled)).To(Succeed())

		Expect(unmarshaled.Name).To(Equal("test-secret"), "Name should survive round-trip")

		Expect(unmarshaled.LegacySecretName).To(Equal("old-secret"), "LegacySecretName should survive round-trip")

		Expect(unmarshaled.Keys).To(HaveLen(2), "both keys should survive round-trip")
		Expect(unmarshaled.Keys[0].Type).To(Equal("alphanumeric"), "key type should be preserved")

		Expect(unmarshaled.Keys[1].Type).To(Equal("static"), "static type should be preserved")
		Expect(unmarshaled.Keys[1].ChartValue).NotTo(BeNil(), "chartValue pointer should not be nil after unmarshal")
		Expect(*unmarshaled.Keys[1].ChartValue).To(Equal("admin"), "chartValue should be preserved")
	})

	It("should correctly round-trip chartValue in SecretKeySpec", func() {
		chartVal := "my-chart-password"
		spec := generator.SecretSpec{
			Name: "test-secret",
			Keys: []generator.SecretKeySpec{
				{
					Name:       "password",
					Type:       "password",
					ChartValue: &chartVal,
				},
			},
		}

		data, err := json.Marshal(spec)
		Expect(err).NotTo(HaveOccurred())

		var unmarshaled generator.SecretSpec
		Expect(json.Unmarshal(data, &unmarshaled)).To(Succeed())
		Expect(unmarshaled.Keys).To(HaveLen(1))
		Expect(unmarshaled.Keys[0].ChartValue).NotTo(BeNil())
		Expect(*unmarshaled.Keys[0].ChartValue).To(Equal("my-chart-password"))
	})
})

var _ = Describe("GeneratorConfig JSON deserialization", func() {
	It("should correctly parse a multi-secret configuration with various fields from JSON", func() {
		configJSON := `{
			"secrets": [
				{
					"name": "runai-backend-gen-postgresql",
					"legacySecretName": "runai-backend-postgresql",
					"keys": [
						{"name": "password", "type": "alphanumeric", "length": 24},
						{"name": "postgres-password", "type": "alphanumeric", "length": 24}
					]
				},
				{
					"name": "runai-backend-gen-cache",
					"keys": [
						{"name": "cache-password", "type": "alphanumeric", "length": 24}
					]
				}
			]
		}`

		var config generator.GeneratorConfig
		Expect(json.Unmarshal([]byte(configJSON), &config)).To(Succeed())

		Expect(config.Secrets).To(HaveLen(2), "should parse both secret specifications")

		Expect(config.Secrets[0].Name).To(Equal("runai-backend-gen-postgresql"),
			"first secret name should match")
		Expect(config.Secrets[0].Keys).To(HaveLen(2),
			"first secret should have two key specifications")

		Expect(config.Secrets[1].LegacySecretName).To(BeEmpty(),
			"second secret should have no legacy secret name when omitted from JSON")
	})
})
