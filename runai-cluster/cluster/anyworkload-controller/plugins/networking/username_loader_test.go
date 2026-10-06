/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package networking

import (
	"encoding/base64"

	workloadconfig "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-common/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("CleanAndBase64Encode", func() {
	DescribeTable("cleans the email list before encoding",
		func(input, expected string) {
			Expect(CleanAndBase64Encode(input)).To(Equal(base64.StdEncoding.EncodeToString([]byte(expected))))
		},
		Entry("single email", "email@run.ai", "email@run.ai"),
		Entry("multiple emails", "email@run.ai,email2@run.ai,email3@run.ai", "email@run.ai\nemail2@run.ai\nemail3@run.ai"),
		Entry("extra commas", "email@run.ai,,email2@run.ai,email3@run.ai,", "email@run.ai\nemail2@run.ai\nemail3@run.ai"),
		Entry("extra spaces", "email@run.ai ,email2@run.ai , email3@run.ai", "email@run.ai\nemail2@run.ai\nemail3@run.ai"),
	)
})

var _ = Describe("UsernameLoaderSecurityContext", func() {
	It("sets an explicit non-root UID on vanilla Kubernetes", func() {
		GinkgoT().Setenv(workloadconfig.EnvVarOpenshift, "false")

		securityContext := UsernameLoaderSecurityContext()

		Expect(securityContext.RunAsNonRoot).To(HaveValue(BeTrue()))
		Expect(securityContext.RunAsUser).To(HaveValue(Equal(UsernameLoaderRunAsUser)))
	})

	It("lets the SCC assign an allowed UID on OpenShift", func() {
		GinkgoT().Setenv(workloadconfig.EnvVarOpenshift, "true")

		securityContext := UsernameLoaderSecurityContext()

		Expect(securityContext.RunAsNonRoot).To(HaveValue(BeTrue()))
		Expect(securityContext.RunAsUser).To(BeNil())
	})
})

var _ = Describe("UsernameLoaderResources", func() {
	BeforeEach(func() {
		GinkgoT().Setenv(workloadconfig.EnvVarUsernameLoaderResources, "")
	})

	It("keeps legacy defaults when unset", func() {
		container, err := BuildUsernameLoaderInitContainerWithError("user@example.com", "test")
		Expect(err).NotTo(HaveOccurred())
		Expect(container).NotTo(BeNil())
		expected := corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("50m"),
				corev1.ResourceMemory: resource.MustParse("50Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("100Mi"),
			},
		}
		Expect(container.Resources.Requests[corev1.ResourceCPU].Equal(expected.Requests[corev1.ResourceCPU])).To(BeTrue())
		Expect(container.Resources.Requests[corev1.ResourceMemory].Equal(expected.Requests[corev1.ResourceMemory])).To(BeTrue())
		Expect(container.Resources.Limits[corev1.ResourceCPU].Equal(expected.Limits[corev1.ResourceCPU])).To(BeTrue())
		Expect(container.Resources.Limits[corev1.ResourceMemory].Equal(expected.Limits[corev1.ResourceMemory])).To(BeTrue())
	})

	It("applies configured resources", func() {
		GinkgoT().Setenv(workloadconfig.EnvVarUsernameLoaderResources, `{"requests":{"cpu":"25m"}}`)
		container, err := BuildUsernameLoaderInitContainerWithError("user@example.com", "test")
		Expect(err).NotTo(HaveOccurred())
		Expect(container).NotTo(BeNil())
		Expect(container.Resources.Requests[corev1.ResourceCPU].Equal(resource.MustParse("25m"))).To(BeTrue())
	})

	It("explicit empty overrides legacy defaults", func() {
		GinkgoT().Setenv(workloadconfig.EnvVarUsernameLoaderResources, `{}`)
		container, err := BuildUsernameLoaderInitContainerWithError("user@example.com", "test")
		Expect(err).NotTo(HaveOccurred())
		Expect(container).NotTo(BeNil())
		Expect(container.Resources.Requests).To(BeEmpty())
		Expect(container.Resources.Limits).To(BeEmpty())
	})

	It("rejects malformed resources", func() {
		GinkgoT().Setenv(workloadconfig.EnvVarUsernameLoaderResources, `{"requests":{"cpu":"invalid"}}`)
		container, err := BuildUsernameLoaderInitContainerWithError("user@example.com", "test")
		Expect(err).To(HaveOccurred())
		Expect(container).To(BeNil())
	})
})
