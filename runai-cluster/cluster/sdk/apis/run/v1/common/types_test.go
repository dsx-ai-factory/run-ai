/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"context"
	"fmt"

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

	It("Image with deprecated repository set", func(ctx context.Context) {
		image := &Image{
			Repository: ptr.To("repo"),
		}
		image.SetDefaultsWhereNeeded(&Image{
			Name: ptr.To("name"),
		})

		Expect(*image.Registry).To(Equal(fmt.Sprintf("%s%s", "repo", DefaultEnv)))
	})

	It("Image with deprecated env set", func(ctx context.Context) {
		image := &Image{
			Env: ptr.To("env"),
		}
		image.SetDefaultsWhereNeeded(&Image{
			Name: ptr.To("name"),
		})

		Expect(*image.Registry).To(Equal(fmt.Sprintf("%s%s", DefaultRepositeory, "env")))
	})

	It("Image with deprecated env and repository set", func(ctx context.Context) {
		image := &Image{
			Env:        ptr.To("env"),
			Repository: ptr.To("repo"),
		}
		image.SetDefaultsWhereNeeded(&Image{
			Name: ptr.To("name"),
		})

		Expect(*image.Registry).To(Equal(fmt.Sprintf("%s%s", "repo", "env")))
	})

	It("Image with deprecated env and repository set and global registry", func(ctx context.Context) {
		image := &Image{
			Env:        ptr.To("env"),
			Repository: ptr.To("repo"),
		}
		image.SetDefaultsWhereNeeded(&Image{
			Registry: ptr.To("global"),
			Name:     ptr.To("name"),
		})

		Expect(*image.Registry).To(Equal(fmt.Sprintf("%s%s", "repo", "env")))
	})
})
