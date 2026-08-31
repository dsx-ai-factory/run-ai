/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/common"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var _ = Describe("HTTPRoute helpers", func() {
	Context("ParseClusterDomain", func() {
		It("should remove port if specified", func() {
			url := "https://test.com:30080"
			parsedURL := ParseClusterDomain(url)
			Expect(parsedURL.Host).NotTo(ContainSubstring(":30080"))
		})

		It("should return hostname if port is not specified", func() {
			urlNoScheme := "test.com"
			url := "https://" + urlNoScheme

			parsedURL := ParseClusterDomain(url)
			Expect(parsedURL.Host).To(Equal(urlNoScheme))
		})
		It("should return hostname if scheme is not specified", func() {
			urlNoScheme := "test.com"

			parsedURL := ParseClusterDomain(urlNoScheme)
			Expect(parsedURL.Host).To(Equal(urlNoScheme))
		})
	})

	Context("GetIngressPath", func() {
		It("should return default path if cluster domain is empty", func() {
			result := GetIngressPath("", "api")
			Expect(result).To(Equal("/api"))
		})

		It("should return parsed path without trailing slash", func() {
			result := GetIngressPath("https://test.com/app/", "api")
			Expect(result).To(Equal("/app"))
		})

		It("should return default path without trailing slash if prefix has trailing slash", func() {
			result := GetIngressPath("", "api/")
			Expect(result).To(Equal("/api"))
		})

		It("should return parsed path as-is if no trailing slash exists", func() {
			result := GetIngressPath("https://test.com/app", "api")
			Expect(result).To(Equal("/app"))
		})

		It("should return default path if no path in parsed domain", func() {
			result := GetIngressPath("https://test.com", "api")
			Expect(result).To(Equal("/api"))
		})

		It("should return default path for root-path URL", func() {
			result := GetIngressPath("https://test.com/", "api")
			Expect(result).To(Equal("/api"))
		})
	})

	Context("BuildParentRef", func() {
		It("should return empty ParentReference when gatewayAPI is nil", func() {
			ref := BuildParentRef(nil, nil)
			Expect(ref).To(Equal(gwv1.ParentReference{}))
		})

		It("should return empty ParentReference when Name is empty", func() {
			gw := &global.GatewayAPI{
				Namespace: "gateway-ns",
			}
			ref := BuildParentRef(gw, nil)
			Expect(ref).To(Equal(gwv1.ParentReference{}))
		})

		It("should build ParentReference with name and namespace", func() {
			gw := &global.GatewayAPI{
				Name:      "my-gateway",
				Namespace: "gateway-ns",
			}
			ref := BuildParentRef(gw, nil)
			Expect(ref.Name).To(Equal(gwv1.ObjectName("my-gateway")))
			Expect(ref.Namespace).NotTo(BeNil())
			Expect(*ref.Namespace).To(Equal(gwv1.Namespace("gateway-ns")))
		})

		It("should build ParentReference without namespace when empty", func() {
			gw := &global.GatewayAPI{
				Name: "my-gateway",
			}
			ref := BuildParentRef(gw, nil)
			Expect(ref.Name).To(Equal(gwv1.ObjectName("my-gateway")))
			Expect(ref.Namespace).To(BeNil())
		})

		It("should use override name/namespace when provided", func() {
			gw := &global.GatewayAPI{
				Name:      "global-gateway",
				Namespace: "global-ns",
			}
			override := &common.GatewayRef{
				Name:      "override-gateway",
				Namespace: "override-ns",
			}
			ref := BuildParentRef(gw, override)
			Expect(ref.Name).To(Equal(gwv1.ObjectName("override-gateway")))
			Expect(ref.Namespace).NotTo(BeNil())
			Expect(*ref.Namespace).To(Equal(gwv1.Namespace("override-ns")))
		})

		It("should fall back to global when override name is empty", func() {
			gw := &global.GatewayAPI{
				Name:      "global-gateway",
				Namespace: "global-ns",
			}
			override := &common.GatewayRef{
				Name: "",
			}
			ref := BuildParentRef(gw, override)
			Expect(ref.Name).To(Equal(gwv1.ObjectName("global-gateway")))
			Expect(ref.Namespace).NotTo(BeNil())
			Expect(*ref.Namespace).To(Equal(gwv1.Namespace("global-ns")))
		})

		It("should fall back to global when override is nil", func() {
			gw := &global.GatewayAPI{
				Name:      "global-gateway",
				Namespace: "global-ns",
			}
			ref := BuildParentRef(gw, nil)
			Expect(ref.Name).To(Equal(gwv1.ObjectName("global-gateway")))
			Expect(ref.Namespace).NotTo(BeNil())
			Expect(*ref.Namespace).To(Equal(gwv1.Namespace("global-ns")))
		})
	})

	Context("HTTPRouteHostname", func() {
		It("should parse hostname from URL with scheme", func() {
			hostname := HTTPRouteHostname("https://test.example.com")
			Expect(hostname).To(Equal(gwv1.Hostname("test.example.com")))
		})

		It("should parse hostname from URL without scheme", func() {
			hostname := HTTPRouteHostname("test.example.com")
			Expect(hostname).To(Equal(gwv1.Hostname("test.example.com")))
		})

		It("should return empty hostname for empty URL", func() {
			hostname := HTTPRouteHostname("")
			Expect(hostname).To(Equal(gwv1.Hostname("")))
		})
	})
})
