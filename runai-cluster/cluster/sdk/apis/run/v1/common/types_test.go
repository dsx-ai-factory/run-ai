/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("common", func() {

	It("loads a third-party image from the chart environment", func(ctx context.Context) {
		GinkgoT().Setenv("IMG_THIRD_PARTY_TEST_REGISTRY", "registry.example")
		GinkgoT().Setenv("IMG_THIRD_PARTY_TEST_REPOSITORY", "namespace/image")
		GinkgoT().Setenv("IMG_THIRD_PARTY_TEST_TAG", "test")
		image := ThirdPartyImage("TEST")
		Expect(*image.Registry).To(Equal("registry.example"))
		Expect(*image.Name).To(Equal("namespace/image"))
		Expect(*image.Tag).To(Equal("test"))
	})

	It("Image with explicit repo", func(ctx context.Context) {
		globalRegistry := ""
		image := &Image{
			Registry: ptr.To[string]("explicit.com"),
		}
		image.SetDefaultsWhereNeeded(&Image{
			Registry: ptr.To(globalRegistry),
			Name:     ptr.To("test_name"),
		})
		Expect(*image.Registry).To(Equal("explicit.com"))
		Expect(*image.Name).To(Equal("test_name"))
		Expect(*image.Tag).To(Equal(VersionTestValue))
		Expect(image.PullPolicy).To(Equal(v1.PullAlways))
	})

	It("Image with global registry", func(ctx context.Context) {
		globalRegistry := "private.com"
		image := &Image{}
		image.SetDefaultsWhereNeeded(&Image{
			Registry: ptr.To(globalRegistry),
			Name:     ptr.To("agent"),
		})
		Expect(*image.Registry).To(Equal("private.com"))
		Expect(*image.Name).To(Equal("agent"))
		Expect(*image.Tag).To(Equal(VersionTestValue))
		Expect(image.PullPolicy).To(Equal(v1.PullAlways))
	})

	It("Image override only name", func(ctx context.Context) {
		image := &Image{}
		image.SetDefaultsWhereNeeded(&Image{
			Name: ptr.To("agent"),
		})
		Expect(*image.Registry).To(Equal(RegistryTestValue))
		Expect(*image.Name).To(Equal("agent"))
		Expect(*image.Tag).To(Equal(VersionTestValue))
		Expect(image.PullPolicy).To(Equal(v1.PullAlways))
	})

	It("Override all image values", func(ctx context.Context) {
		image := &Image{}
		image.SetDefaultsWhereNeeded(&Image{
			Registry:   ptr.To("registry.k8s.io/git-sync"),
			Name:       ptr.To("git-sync"),
			Tag:        ptr.To("v4.4.0"),
			PullPolicy: v1.PullIfNotPresent,
		})
		Expect(*image.Registry).To(Equal("registry.k8s.io/git-sync"))
		Expect(*image.Name).To(Equal("git-sync"))
		Expect(*image.Tag).To(Equal("v4.4.0"))
		Expect(image.PullPolicy).To(Equal(v1.PullIfNotPresent))
	})

	Context("digest", func() {
		const digest = "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

		newImage := func(d *string) *Image {
			return &Image{
				Registry: ptr.To("registry.example"),
				Name:     ptr.To("agent"),
				Tag:      ptr.To("3.0.20"),
				Digest:   d,
			}
		}

		It("is never defaulted, since it is only valid for the image it was computed from", func() {
			image := &Image{}
			image.SetDefaultsWhereNeeded(&Image{Name: ptr.To("agent")})
			Expect(image.Digest).To(BeNil())
		})

		It("keeps a digest set on the image through defaulting", func() {
			image := &Image{Digest: ptr.To(digest)}
			image.SetDefaultsWhereNeeded(&Image{Name: ptr.To("agent")})
			Expect(image.Digest).To(Equal(ptr.To(digest)))
		})

		DescribeTable("renders the pullable reference",
			func(d *string, expected string) {
				Expect(newImage(d).Reference()).To(Equal(expected))
			},
			Entry("by tag alone when no digest is set", nil, "registry.example/agent:3.0.20"),
			Entry("treating an empty digest as unset", ptr.To(""), "registry.example/agent:3.0.20"),
			Entry("with the digest after the tag", ptr.To(digest), "registry.example/agent:3.0.20@"+digest),
		)

		It("omits the colon before the digest when the tag is explicitly empty", func() {
			image := &Image{
				Registry: ptr.To("registry.example"),
				Name:     ptr.To("agent"),
				Tag:      ptr.To(""),
				Digest:   ptr.To(digest),
			}
			Expect(image.Reference()).To(Equal("registry.example/agent@" + digest))
		})

		It("renders the tag with the digest for APIs that carry no digest field", func() {
			Expect(newImage(nil).tag()).To(Equal("3.0.20"))
			Expect(newImage(ptr.To(digest)).tag()).To(Equal("3.0.20@" + digest))
		})
	})
})
