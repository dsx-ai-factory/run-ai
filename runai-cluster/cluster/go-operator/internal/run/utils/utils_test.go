/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"context"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	scheme_init "github.com/run-ai/runai/runai-cluster/cluster/go-operator/pkg/scheme-init"
	runv1 "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("Resource Utility Functions", func() {
	Context("ResourceStructType()", func() {
		It("should return a struct pointer's type name", func() {
			pod := &v1.Pod{}

			structName := ResourceStructType(pod)

			Expect(structName).To(Equal("Pod"))
		})

		It("should return a struct's nil pointer type name", func() {
			pod := (*v1.Pod)(nil)

			structName := ResourceStructType(pod)

			Expect(structName).To(Equal("Pod"))
		})
	})

	Context("ListProjectsNamespacesNames()", func() {
		var (
			k8s         rclient.Client
			runaiConfig *runv1.RunaiConfig
		)

		BeforeEach(func() {
			k8s = fake.NewClientBuilder().WithScheme(scheme_init.Scheme()).Build()
			runaiConfig = &runv1.RunaiConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "runai",
					Namespace: "runai",
				},
			}
			runaiConfig.SetDefaultsWhereNeeded()
		})

		It("should return a list of runai project namespaces", func() {
			Expect(k8s.Create(context.Background(), &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-runai",
					Labels: map[string]string{
						"runai/queue": "test",
					},
				},
			}))
			ns, err := ListProjectsNamespacesNames(context.Background(), k8s)
			Expect(err).ToNot(HaveOccurred())
			Expect(ns).To(ContainElement("test-runai"))
		})

		It("should return an empty slice when no project namespaces exist", func() {
			ns, err := ListProjectsNamespacesNames(context.Background(), k8s)
			Expect(err).ToNot(HaveOccurred())
			Expect(ns).To(BeEmpty())
		})
	})

	Context("Test utils", func() {
		var runaiConfig *runv1.RunaiConfig

		BeforeEach(func() {
			runaiConfig = GetDefaultRunaiConfig()
			runaiConfig.Spec.Global.HTTPProxy.Enabled = true
			runaiConfig.Spec.Global.HTTPProxy.TLSCertSecret = &v1.SecretKeySelector{
				LocalObjectReference: v1.LocalObjectReference{Name: "awesome-tls-secret"},
				Key:                  "tls.crt",
			}
		})

		It("volumeMountsForHTTPProxy", func() {
			volumeMount := VolumeMountsForHTTPProxy(runaiConfig)
			Expect(volumeMount[0].Name).To(BeEquivalentTo(runaiConfig.Spec.Global.HTTPProxy.ConfigMapName))
		})

		It("volumesForHTTPProxy", func() {
			volume := VolumesForHTTPProxy(runaiConfig)
			Expect(volume[0].Name).To(BeEquivalentTo(runaiConfig.Spec.Global.HTTPProxy.ConfigMapName))
		})
	})

	Context("MapToArgFormat", func() {
		It("should join a map to multi-valued arg format", func() {
			m := map[string]string{
				"foo": "bar",
				"bar": "baz",
				"baz": "foo",
			}

			arg := MapToArgFormat(m)
			splitArgs := strings.Split(arg, ",")

			Expect(len(splitArgs)).To(BeEquivalentTo(len(m)))
			Expect(splitArgs).To(ContainElement("foo=bar"))
			Expect(splitArgs).To(ContainElement("bar=baz"))
			Expect(splitArgs).To(ContainElement("baz=foo"))
		})
	})

	Context("GetImageTag", func() {
		It("should return tag for simple image:tag reference", func() {
			tag, err := GetImageTag("kubeflow/training-operator:v1-3f15cb8")
			Expect(err).ToNot(HaveOccurred())
			Expect(tag).To(Equal("v1-3f15cb8"))
		})

		It("should return tag when image has registry with port (host:port/path/image:tag)", func() {
			image := "test-registry:8281/kubeflow/training-v1/training-operator:v1-3f15cb8"
			tag, err := GetImageTag(image)
			Expect(err).ToNot(HaveOccurred())
			Expect(tag).To(Equal("v1-3f15cb8"))
		})

		It("should return tag when tag contains colons", func() {
			tag, err := GetImageTag("repo/image:tag:with:colons")
			Expect(err).ToNot(HaveOccurred())
			Expect(tag).To(Equal("colons"))
		})

		It("should return error when image has registry with port but no tag", func() {
			_, err := GetImageTag("myregistry:5000/myimage")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cannot determine image"))
		})

		It("should return error when image has no tag", func() {
			_, err := GetImageTag("kubeflow/training-operator")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cannot determine image"))
		})

		It("should return error when image has no tag but colon", func() {
			_, err := GetImageTag("kubeflow/training-operator:")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("cannot determine image"))
		})

		It("should return error for empty image string", func() {
			_, err := GetImageTag("")
			Expect(err).To(HaveOccurred())
		})
	})
})
