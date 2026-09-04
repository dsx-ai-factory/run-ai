/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package spec

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SubmitSpec", func() {
	var (
		submitSpec *SubmitSpec
	)

	BeforeEach(func() {
		submitSpec = NewSubmitSpec()
		submitSpec.flags.Network.Ports = []string{"30000:3000"}
	})

	Context("MutateServiceType", func() {
		When("custom URL is provided without service type", func() {
			It("returns an error", func() {
				err := submitSpec.MutateServiceType([]string{}, "http://custom.url")
				Expect(err).To(MatchError("custom-url provided without a service-type"))
			})
		})

		When("no service type is provided", func() {
			It("returns nil", func() {
				err := submitSpec.MutateServiceType([]string{}, "")
				Expect(err).To(BeNil())
			})
		})

		When("nodeport service type is provided", func() {
			It("correctly mutates ports", func() {

				err := submitSpec.MutateServiceType([]string{"nodeport"}, "")
				Expect(err).To(BeNil())
				Expect(submitSpec.GeneralSpec.flags.Network.Ports).To(ConsistOf(
					"service-type=NodePort,container=3000,external=30000",
				))
			})
		})

		When("external-url service type is provided", func() {
			It("correctly mutates external URLs with custom URL", func() {
				submitSpec = NewSubmitSpec()
				serviceType := []string{"external-url,port=3000,custom-url=http://my.customurl.com/myapi"}
				err := submitSpec.MutateServiceType(serviceType, "")
				Expect(err).To(BeNil())
				Expect(submitSpec.GeneralSpec.flags.Network.ExternalUrls).To(ConsistOf(
					"container=3000,url=http://my.customurl.com/myapi",
				))
			})

			It("correctly mutates external URLs without custom URL", func() {
				submitSpec = NewSubmitSpec()
				serviceType := []string{"external-url,port=3001"}
				err := submitSpec.MutateServiceType(serviceType, "")
				Expect(err).To(BeNil())
				Expect(submitSpec.GeneralSpec.flags.Network.ExternalUrls).To(ConsistOf(
					"container=3001",
				))
			})
		})

		When("multiple service types are provided", func() {
			It("correctly mutates both ports and external URLs", func() {
				submitSpec.flags.Network.Ports = []string{
					"32000:3000",
				}
				serviceType := []string{
					"nodeport",
					"external-url",
				}
				err := submitSpec.MutateServiceType(serviceType, "http://my.customurl.com/myapi")
				Expect(err).To(BeNil())

				Expect(submitSpec.GeneralSpec.flags.Network.Ports).To(ConsistOf(
					"service-type=NodePort,container=3000,external=32000",
				))
				Expect(submitSpec.GeneralSpec.flags.Network.ExternalUrls).To(ConsistOf(
					"container=3000,url=http://my.customurl.com/myapi",
				))
			})
		})
	})

	Context("mutateExternalUrl", func() {
		It("correctly formats external URL without custom URL", func() {
			result, err := mutateExternalUrl(ExternalUrlArg{port: "3000"})
			Expect(err).To(BeNil())
			Expect(result).To(Equal("container=3000"))
		})

		It("correctly formats external URL with custom URL", func() {
			result, err := mutateExternalUrl(ExternalUrlArg{
				port:      "3000",
				customUrl: "http://custom.url",
			})
			Expect(err).To(BeNil())
			Expect(result).To(Equal("container=3000,url=http://custom.url"))
		})
	})

	Context("mutateNodePort", func() {
		It("correctly formats node port string", func() {
			result, err := mutateNodePort("30000:3000")
			Expect(err).To(BeNil())
			Expect(result).To(Equal("service-type=NodePort,container=3000,external=30000"))
		})

		It("returns error for invalid port format", func() {
			result, err := mutateNodePort("invalid:port")
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeEmpty())
		})
	})
})
