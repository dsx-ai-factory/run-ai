/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package create_test

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/mycredential/create"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("mycredential create parsing", func() {
	Describe("ParseGenericSecretItems", func() {
		It("parses one key-value pair", func() {
			pairs, err := create.ParseGenericSecretItems([]string{"key=API_URL,value=https://api.example.com"})
			Expect(err).ToNot(HaveOccurred())
			Expect(pairs).To(HaveLen(1))
			Expect(pairs[0].Key).To(Equal("API_URL"))
			Expect(pairs[0].Value).To(Equal("https://api.example.com"))
		})

		It("parses multiple key-value pairs", func() {
			pairs, err := create.ParseGenericSecretItems([]string{
				"key=API_URL,value=https://api.example.com",
				"key=DB_PASS,value=secret",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(pairs).To(HaveLen(2))
			Expect(pairs[0].Key).To(Equal("API_URL"))
			Expect(pairs[0].Value).To(Equal("https://api.example.com"))
			Expect(pairs[1].Key).To(Equal("DB_PASS"))
			Expect(pairs[1].Value).To(Equal("secret"))
		})

		It("rejects item with only value (missing key)", func() {
			_, err := create.ParseGenericSecretItems([]string{"value=only"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("genericSecret --item must contain key and value"))
		})

		It("rejects item with only key (missing value)", func() {
			_, err := create.ParseGenericSecretItems([]string{"key=API_URL"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("genericSecret --item must contain key and value"))
		})

		It("rejects item with empty key", func() {
			_, err := create.ParseGenericSecretItems([]string{"key=,value=secret"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("genericSecret --item must contain key and value"))
		})

		It("rejects item with empty value", func() {
			_, err := create.ParseGenericSecretItems([]string{"key=API_URL,value="})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("genericSecret --item must contain key and value"))
		})

		It("rejects when no items provided", func() {
			_, err := create.ParseGenericSecretItems(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("at least one"))
		})

		It("rejects item without equals sign in segment", func() {
			_, err := create.ParseGenericSecretItems([]string{"keyvalue"})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ParseDockerRegistryItems", func() {
		It("parses user, password, and url", func() {
			user, pass, url, err := create.ParseDockerRegistryItems([]string{
				"user=myuser,password=secret,url=https://index.docker.io/v1/",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(user).To(Equal("myuser"))
			Expect(pass).To(Equal("secret"))
			Expect(url).To(Equal("https://index.docker.io/v1/"))
		})

		It("rejects when password is missing", func() {
			_, _, _, err := create.ParseDockerRegistryItems([]string{"user=u,url=https://index.docker.io/v1/"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("dockerRegistry --item must contain user, password, and url"))
		})

		It("rejects when user is missing", func() {
			_, _, _, err := create.ParseDockerRegistryItems([]string{"password=p,url=https://index.docker.io/v1/"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("dockerRegistry --item must contain user, password, and url"))
		})

		It("rejects when url is missing", func() {
			_, _, _, err := create.ParseDockerRegistryItems([]string{"user=u,password=p"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("dockerRegistry --item must contain user, password, and url"))
		})

		It("rejects when no items provided", func() {
			_, _, _, err := create.ParseDockerRegistryItems(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("dockerRegistry --item must contain user, password, and url"))
		})

		It("rejects unknown or extra key", func() {
			_, _, _, err := create.ParseDockerRegistryItems([]string{"user=u,password=p,url=u,extra=bad"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("dockerRegistry --item must contain user, password, and url"))
		})
	})

	Describe("ParseNgcApiKeyItems", func() {
		It("parses ngc-api-key", func() {
			key, err := create.ParseNgcApiKeyItems([]string{"ngc-api-key=YOUR_NGC_API_KEY"})
			Expect(err).ToNot(HaveOccurred())
			Expect(key).To(Equal("YOUR_NGC_API_KEY"))
		})

		It("rejects when no items provided", func() {
			_, err := create.ParseNgcApiKeyItems(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ngcApiKey --item must contain ngc-api-key"))
		})

		It("rejects empty ngc-api-key value", func() {
			_, err := create.ParseNgcApiKeyItems([]string{"ngc-api-key="})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ngc-api-key"))
		})

		It("rejects wrong key name (api-key instead of ngc-api-key)", func() {
			_, err := create.ParseNgcApiKeyItems([]string{"api-key=MY_KEY"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ngcApiKey --item must contain ngc-api-key"))
		})
	})
})
