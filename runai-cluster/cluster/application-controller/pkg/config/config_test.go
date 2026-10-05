/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("parseClientCredentials", func() {
	It("maps the whisper secret keys onto the credential fields", func() {
		// Exactly the JSON the go-operator vault writes into whisper-application-controller-secret.
		data := []byte(`{"CLIENT_NAME":"cluster-agent","CLIENT_SECRET":"s3cr3t","CLUSTER_ID":"abc-123"}`)

		creds, err := parseClientCredentials(data)

		Expect(err).NotTo(HaveOccurred())
		Expect(creds.ClientName).To(Equal("cluster-agent"))
		Expect(creds.ClientSecret).To(Equal("s3cr3t"))
		Expect(creds.ClusterID).To(Equal("abc-123"))
	})

	It("returns an error on invalid json", func() {
		_, err := parseClientCredentials([]byte("not-json"))

		Expect(err).To(HaveOccurred())
	})

	It("leaves fields empty when keys are absent", func() {
		creds, err := parseClientCredentials([]byte(`{}`))

		Expect(err).NotTo(HaveOccurred())
		Expect(creds.ClientName).To(BeEmpty())
		Expect(creds.ClientSecret).To(BeEmpty())
		Expect(creds.ClusterID).To(BeEmpty())
	})
})

var _ = Describe("Application admission identities", func() {
	BeforeEach(func() {
		GinkgoT().Setenv("HELM_CACHE_HOME", "/tmp/helm/cache")
		GinkgoT().Setenv("HELM_CONFIG_HOME", "/tmp/helm/config")
	})

	It("defaults to the application controller and agent's full usernames", func() {
		GinkgoT().Setenv("ALLOWED_SERVICE_ACCOUNTS", "")
		Expect(newConfig().AllowedServiceAccounts).To(Equal([]string{
			"system:serviceaccount:runai:application-controller",
			"system:serviceaccount:runai:runai-agent",
		}))
	})

	It("accepts configured full usernames for a custom namespace", func() {
		GinkgoT().Setenv("ALLOWED_SERVICE_ACCOUNTS", "system:serviceaccount:custom:application-controller")
		Expect(newConfig().AllowedServiceAccounts).To(Equal([]string{"system:serviceaccount:custom:application-controller"}))
	})
})
